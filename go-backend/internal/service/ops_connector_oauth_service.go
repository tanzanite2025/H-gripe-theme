package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/pkg/secretbox"
	"commerce-platform/internal/repository"
)

const (
	opsConnectorOAuthCallbackPath = "/api/admin/ops/connectors/oauth/callback"
	cloudflareOAuthAuthorizeURL   = "https://dash.cloudflare.com/oauth2/auth"
	cloudflareOAuthTokenURL       = "https://dash.cloudflare.com/oauth2/token"
	hostingerOAuthIssuer          = "https://auth.hostinger.com"
	hostingerOAuthRegisterURL     = hostingerOAuthIssuer + "/api/external/v1/oauth-server/register"
	hostingerOAuthAuthorizeURL    = hostingerOAuthIssuer + "/api/external/v1/oauth-server/authorize"
	hostingerOAuthTokenURL        = hostingerOAuthIssuer + "/api/external/v1/oauth-server/token"
)

var (
	ErrOpsConnectorOAuthNotConfigured = errors.New("operations connector OAuth is not configured")
	ErrOpsConnectorOAuthStateInvalid  = errors.New("operations connector OAuth state is invalid or expired")
	ErrOpsConnectorOAuthExchange      = errors.New("operations connector OAuth exchange failed")
)

type OpsConnectorOAuthStartInput struct {
	Provider    string `json:"provider"`
	ConnectorID *uint  `json:"connector_id"`
	Environment string `json:"environment"`
	ReturnPath  string `json:"return_path"`
}

type OpsConnectorOAuthStartResult struct {
	AuthorizationURL string `json:"authorization_url"`
	Provider         string `json:"provider"`
	ConnectorID      uint   `json:"connector_id"`
	ConnectorName    string `json:"connector_name"`
}

type OpsConnectorOAuthCallbackResult struct {
	Provider          string                 `json:"provider"`
	ConnectorID       uint                   `json:"connector_id"`
	ConnectorName     string                 `json:"connector_name"`
	Status            string                 `json:"status"`
	Message           string                 `json:"message"`
	ReturnPath        string                 `json:"return_path"`
	BoundVPSCount     int                    `json:"bound_vps_count"`
	BoundProjectCount int                    `json:"bound_project_count"`
	BoundDomainCount  int                    `json:"bound_domain_count"`
	VPSCandidates     []OpsOAuthResourceItem `json:"vps_candidates,omitempty"`
	CloudflareZones   []OpsOAuthResourceItem `json:"cloudflare_zones,omitempty"`
	BindingWarnings   []string               `json:"binding_warnings,omitempty"`
}

type OpsOAuthResourceItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
}

type OpsConnectorOAuthService struct {
	sessionRepo        *repository.OpsConnectorOAuthRepository
	connectorRepo      *repository.OpsConnectorRepository
	connectorService   *OpsConnectorService
	vpsRepo            *repository.OpsVPSBindingRepository
	projectRepo        *repository.OpsProjectBindingRepository
	domainRepo         *repository.OpsDomainBindingRepository
	vpsSyncService     *OpsHostingerSyncService
	projectSyncService *OpsHostingerSyncService
	domainSyncService  *OpsDomainSyncService
	httpClient         *http.Client
	publicBaseURL      string
}

func NewOpsConnectorOAuthService(
	sessionRepo *repository.OpsConnectorOAuthRepository,
	connectorRepo *repository.OpsConnectorRepository,
	connectorService *OpsConnectorService,
	vpsRepo *repository.OpsVPSBindingRepository,
	projectRepo *repository.OpsProjectBindingRepository,
	domainRepo *repository.OpsDomainBindingRepository,
	vpsSyncService *OpsHostingerSyncService,
	projectSyncService *OpsHostingerSyncService,
	domainSyncService *OpsDomainSyncService,
	publicBaseURL string,
) *OpsConnectorOAuthService {
	return &OpsConnectorOAuthService{
		sessionRepo:        sessionRepo,
		connectorRepo:      connectorRepo,
		connectorService:   connectorService,
		vpsRepo:            vpsRepo,
		projectRepo:        projectRepo,
		domainRepo:         domainRepo,
		vpsSyncService:     vpsSyncService,
		projectSyncService: projectSyncService,
		domainSyncService:  domainSyncService,
		httpClient:         &http.Client{Timeout: 15 * time.Second},
		publicBaseURL:      strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"),
	}
}

func (s *OpsConnectorOAuthService) Start(
	ctx context.Context,
	userID uint,
	input OpsConnectorOAuthStartInput,
) (*OpsConnectorOAuthStartResult, error) {
	if s == nil || s.sessionRepo == nil || s.connectorRepo == nil || s.connectorService == nil {
		return nil, errors.New("operations connector OAuth service is not configured")
	}
	if strings.TrimSpace(os.Getenv(OpsConnectorMasterKeyEnv)) == "" {
		return nil, fmt.Errorf("%w: %s is required", ErrOpsConnectorOAuthNotConfigured, OpsConnectorMasterKeyEnv)
	}
	provider := normalizeOAuthProvider(input.Provider)
	if provider == "" {
		return nil, fmt.Errorf("%w: provider must be Cloudflare or Hostinger", ErrOpsConnectorOAuthNotConfigured)
	}
	environment, err := normalizeOAuthConnectorEnvironment(input.Environment)
	if err != nil {
		return nil, err
	}
	connector, err := s.resolveConnector(provider, input.ConnectorID, environment)
	if err != nil {
		return nil, err
	}
	redirectURI, err := s.redirectURI()
	if err != nil {
		return nil, err
	}
	returnPath := normalizeOAuthReturnPath(input.ReturnPath)
	clientID := ""
	if provider == ops.ConnectorProviderHostinger {
		clientID, err = s.hostingerClientID(ctx, *connector, redirectURI)
		if err != nil {
			return nil, err
		}
	} else {
		clientID = strings.TrimSpace(os.Getenv("OPS_CLOUDFLARE_OAUTH_CLIENT_ID"))
		if clientID == "" {
			return nil, fmt.Errorf("%w: OPS_CLOUDFLARE_OAUTH_CLIENT_ID is required", ErrOpsConnectorOAuthNotConfigured)
		}
	}

	state, err := randomOAuthString(32)
	if err != nil {
		return nil, err
	}
	verifier, err := randomOAuthString(32)
	if err != nil {
		return nil, err
	}
	challenge := oauthPKCEChallenge(verifier)
	encryptedVerifier, err := secretbox.EncryptString(verifier, os.Getenv(OpsConnectorMasterKeyEnv))
	if err != nil {
		return nil, fmt.Errorf("%w: encrypt OAuth verifier: %v", ErrOpsConnectorOAuthNotConfigured, err)
	}
	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	session := &ops.ConnectorOAuthSession{
		Provider:              provider,
		ConnectorID:           &connector.ID,
		StateHash:             hashOAuthState(state),
		CodeVerifierEncrypted: encryptedVerifier,
		ClientID:              clientID,
		RedirectURI:           redirectURI,
		ReturnPath:            returnPath,
		CreatedByUserID:       userID,
		ExpiresAt:             expiresAt,
		Status:                ops.ConnectorOAuthStatusPending,
	}
	if err := s.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	authorizationURL, err := s.buildAuthorizationURL(provider, clientID, redirectURI, state, challenge, connector.Scopes)
	if err != nil {
		_ = s.sessionRepo.MarkError(session.ID, err.Error())
		return nil, err
	}
	return &OpsConnectorOAuthStartResult{
		AuthorizationURL: authorizationURL,
		Provider:         provider,
		ConnectorID:      connector.ID,
		ConnectorName:    connector.Name,
	}, nil
}

func (s *OpsConnectorOAuthService) Complete(
	ctx context.Context,
	state string,
	code string,
	providerError string,
	providerErrorDescription string,
) (*OpsConnectorOAuthCallbackResult, error) {
	if s == nil || s.sessionRepo == nil || s.connectorRepo == nil || s.connectorService == nil {
		return nil, errors.New("operations connector OAuth service is not configured")
	}
	state = strings.TrimSpace(state)
	if state == "" {
		return nil, ErrOpsConnectorOAuthStateInvalid
	}
	session, err := s.sessionRepo.Consume(hashOAuthState(state), time.Now().UTC())
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrOpsConnectorOAuthStateInvalid
		}
		return nil, err
	}
	result := &OpsConnectorOAuthCallbackResult{
		Provider:   session.Provider,
		ReturnPath: normalizeOAuthReturnPath(session.ReturnPath),
		Status:     "error",
	}
	fail := func(callbackErr error) (*OpsConnectorOAuthCallbackResult, error) {
		if callbackErr != nil {
			result.Message = callbackErr.Error()
			_ = s.sessionRepo.MarkError(session.ID, callbackErr.Error())
		}
		return result, callbackErr
	}
	if providerError != "" {
		message := strings.TrimSpace(providerError)
		if description := strings.TrimSpace(providerErrorDescription); description != "" {
			message += ": " + description
		}
		return fail(fmt.Errorf("%w: provider returned %s", ErrOpsConnectorOAuthExchange, message))
	}
	if strings.TrimSpace(code) == "" {
		return fail(fmt.Errorf("%w: authorization code is missing", ErrOpsConnectorOAuthExchange))
	}
	verifier, err := secretbox.DecryptString(session.CodeVerifierEncrypted, os.Getenv(OpsConnectorMasterKeyEnv))
	if err != nil {
		return fail(fmt.Errorf("%w: decrypt OAuth verifier", ErrOpsConnectorOAuthExchange))
	}
	token, err := s.exchangeCode(ctx, session.Provider, session.ClientID, session.RedirectURI, code, verifier)
	if err != nil {
		return fail(err)
	}
	connector, err := s.connectorRepo.FindByID(*session.ConnectorID)
	if err != nil {
		return fail(err)
	}
	result.ConnectorID = connector.ID
	result.ConnectorName = connector.Name
	credentials, err := s.connectorService.readCredentials(*connector)
	if err != nil && strings.TrimSpace(connector.CredentialsEncrypted) != "" {
		return fail(err)
	}
	if credentials == nil {
		credentials = map[string]string{}
	}
	for key, value := range token {
		if strings.TrimSpace(value) != "" {
			credentials[key] = value
		}
	}
	credentials["client_id"] = session.ClientID
	if err := s.storeOAuthCredentials(*connector, credentials); err != nil {
		return fail(err)
	}
	test, testErr := s.connectorService.Test(ctx, connector.ID)
	if testErr != nil || test == nil || !test.Success {
		message := "OAuth succeeded but connector test failed"
		if test != nil && test.Message != "" {
			message += ": " + test.Message
		}
		if testErr != nil {
			message += ": " + testErr.Error()
		}
		return fail(errors.New(message))
	}

	if session.Provider == ops.ConnectorProviderHostinger {
		if err := s.bindHostinger(ctx, connector.ID, connector.Environment, result); err != nil {
			return fail(err)
		}
	} else {
		if err := s.bindCloudflare(ctx, connector.ID, connector.Environment, result); err != nil {
			return fail(err)
		}
	}
	finalizeOAuthBindingResult(connector.Name, result)
	return result, nil
}

func finalizeOAuthBindingResult(connectorName string, result *OpsConnectorOAuthCallbackResult) {
	if result == nil {
		return
	}
	result.Status = "connected"
	if len(result.BindingWarnings) > 0 {
		result.Status = "connected_with_warnings"
	}
	result.Message = fmt.Sprintf(
		"%s connected; bound %d VPS, %d projects, and %d domains",
		connectorName,
		result.BoundVPSCount,
		result.BoundProjectCount,
		result.BoundDomainCount,
	)
	if len(result.BindingWarnings) > 0 {
		previewCount := len(result.BindingWarnings)
		if previewCount > 3 {
			previewCount = 3
		}
		result.Message += fmt.Sprintf("; %d binding or sync warnings require review: %s",
			len(result.BindingWarnings),
			strings.Join(result.BindingWarnings[:previewCount], "; "),
		)
		if len(result.BindingWarnings) > previewCount {
			result.Message += fmt.Sprintf("; plus %d more", len(result.BindingWarnings)-previewCount)
		}
	}
}

func (s *OpsConnectorOAuthService) resolveConnector(provider string, connectorID *uint, environment string) (*ops.Connector, error) {
	if connectorID != nil && *connectorID > 0 {
		connector, err := s.connectorRepo.FindByID(*connectorID)
		if err != nil {
			return nil, err
		}
		if connector.Provider != provider {
			return nil, fmt.Errorf("%w: connector provider mismatch", ErrOpsConnectorOAuthNotConfigured)
		}
		if connector.Environment != environment {
			return nil, fmt.Errorf("%w: connector environment mismatch", ErrOpsConnectorOAuthNotConfigured)
		}
		return connector, nil
	}
	connector, err := s.connectorRepo.FindByProviderEnvironment(provider, environment)
	if err == nil {
		return connector, nil
	}
	if !repository.IsRecordNotFound(err) {
		return nil, err
	}
	name := "Cloudflare " + oauthConnectorEnvironmentLabel(environment)
	endpoint := cloudflareAPIBaseURL + "/zones?per_page=1"
	scopes := cloudflareOAuthScopes()
	authType := ops.ConnectorAuthBearer
	if provider == ops.ConnectorProviderHostinger {
		name = "Hostinger " + oauthConnectorEnvironmentLabel(environment)
		endpoint = hostingerAPIBaseURL + "/api/vps/v1/virtual-machines"
		scopes = "vps:read,project:read"
		authType = ops.ConnectorAuthBearer
	}
	created, err := s.connectorService.Create(OpsConnectorInput{
		Name:        name,
		Provider:    provider,
		Environment: environment,
		Endpoint:    endpoint,
		AuthType:    authType,
		Scopes:      scopes,
		Status:      ops.ConnectorStatusPending,
		Enabled:     opsOAuthBoolPtr(true),
		Notes:       "Created by provider OAuth connection.",
	})
	if err != nil {
		return nil, err
	}
	return s.connectorRepo.FindByID(created.ID)
}

func normalizeOAuthConnectorEnvironment(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ops.ConnectorEnvironmentProduction, nil
	}
	switch value {
	case ops.ConnectorEnvironmentProduction,
		ops.ConnectorEnvironmentStaging,
		ops.ConnectorEnvironmentTest,
		ops.ConnectorEnvironmentLocal:
		return value, nil
	default:
		return "", fmt.Errorf("%w: environment is invalid", ErrOpsConnectorOAuthNotConfigured)
	}
}

func oauthConnectorEnvironmentLabel(environment string) string {
	switch environment {
	case ops.ConnectorEnvironmentStaging:
		return "Staging"
	case ops.ConnectorEnvironmentTest:
		return "Test"
	case ops.ConnectorEnvironmentLocal:
		return "Local"
	default:
		return "Production"
	}
}

func (s *OpsConnectorOAuthService) redirectURI() (string, error) {
	raw := strings.TrimRight(strings.TrimSpace(os.Getenv("OPS_CONNECTOR_OAUTH_REDIRECT_URL")), "/")
	if raw == "" {
		raw = s.publicBaseURL
		if raw == "" {
			return "", fmt.Errorf("%w: OPS_CONNECTOR_OAUTH_REDIRECT_URL or server base URL is required", ErrOpsConnectorOAuthNotConfigured)
		}
		raw += opsConnectorOAuthCallbackPath
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil {
		return "", fmt.Errorf("%w: OAuth redirect URL is invalid", ErrOpsConnectorOAuthNotConfigured)
	}
	if parsed.Scheme != "https" && parsed.Hostname() != "localhost" && parsed.Hostname() != "127.0.0.1" {
		return "", fmt.Errorf("%w: OAuth redirect URL must use HTTPS outside local development", ErrOpsConnectorOAuthNotConfigured)
	}
	return parsed.String(), nil
}

func (s *OpsConnectorOAuthService) buildAuthorizationURL(
	provider string,
	clientID string,
	redirectURI string,
	state string,
	challenge string,
	connectorScopes string,
) (string, error) {
	query := url.Values{}
	query.Set("client_id", clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("response_type", "code")
	query.Set("state", state)
	query.Set("code_challenge", challenge)
	query.Set("code_challenge_method", "S256")
	if provider == ops.ConnectorProviderCloudflare {
		if scopes := cloudflareOAuthScopesFromConnector(connectorScopes); scopes != "" {
			query.Set("scope", scopes)
		}
		return cloudflareOAuthAuthorizeURL + "?" + query.Encode(), nil
	}
	return hostingerOAuthAuthorizeURL + "?" + query.Encode(), nil
}

func (s *OpsConnectorOAuthService) hostingerClientID(ctx context.Context, connector ops.Connector, redirectURI string) (string, error) {
	credentials, err := s.connectorService.readCredentials(connector)
	if err == nil {
		if value := strings.TrimSpace(credentials["client_id"]); value != "" {
			return value, nil
		}
	}
	body, err := json.Marshal(map[string]interface{}{
		"client_name":   "tanzanite-ops",
		"redirect_uris": []string{redirectURI},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hostingerOAuthRegisterURL, strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: Hostinger client registration failed: %v", ErrOpsConnectorOAuthExchange, err)
	}
	defer response.Body.Close()
	var payload struct {
		ClientID string `json:"client_id"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("%w: invalid Hostinger client registration response", ErrOpsConnectorOAuthExchange)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || strings.TrimSpace(payload.ClientID) == "" {
		return "", fmt.Errorf("%w: Hostinger client registration returned HTTP %d", ErrOpsConnectorOAuthExchange, response.StatusCode)
	}
	return strings.TrimSpace(payload.ClientID), nil
}

func (s *OpsConnectorOAuthService) exchangeCode(
	ctx context.Context,
	provider string,
	clientID string,
	redirectURI string,
	code string,
	verifier string,
) (map[string]string, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("code_verifier", verifier)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", clientID)
	endpoint := hostingerOAuthTokenURL
	if provider == ops.ConnectorProviderCloudflare {
		endpoint = cloudflareOAuthTokenURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: create token request", ErrOpsConnectorOAuthExchange)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if provider == ops.ConnectorProviderCloudflare {
		if secret := strings.TrimSpace(os.Getenv("OPS_CLOUDFLARE_OAUTH_CLIENT_SECRET")); secret != "" {
			req.SetBasicAuth(clientID, secret)
		}
	}
	return s.readOAuthTokenResponse(req)
}

func (s *OpsConnectorOAuthService) readOAuthTokenResponse(req *http.Request) (map[string]string, error) {
	response, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpsConnectorOAuthExchange, err)
	}
	defer response.Body.Close()
	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
		Error        string `json:"error"`
		Description  string `json:"error_description"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: invalid token response", ErrOpsConnectorOAuthExchange)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || strings.TrimSpace(payload.AccessToken) == "" {
		message := strings.TrimSpace(payload.Description)
		if message == "" {
			message = strings.TrimSpace(payload.Error)
		}
		if message == "" {
			message = fmt.Sprintf("token endpoint returned HTTP %d", response.StatusCode)
		}
		return nil, fmt.Errorf("%w: %s", ErrOpsConnectorOAuthExchange, message)
	}
	credentials := map[string]string{
		"token":        payload.AccessToken,
		"access_token": payload.AccessToken,
		"token_type":   payload.TokenType,
	}
	if payload.RefreshToken != "" {
		credentials["refresh_token"] = payload.RefreshToken
	}
	if payload.Scope != "" {
		credentials["scope"] = payload.Scope
	}
	if payload.ExpiresIn > 0 {
		credentials["expires_at"] = time.Now().UTC().Add(time.Duration(opsOAuthMaxInt(payload.ExpiresIn-60, 1)) * time.Second).Format(time.RFC3339)
	}
	return credentials, nil
}

func refreshConnectorOAuthToken(ctx context.Context, provider string, credentials map[string]string) (map[string]string, error) {
	clientID := strings.TrimSpace(credentials["client_id"])
	refreshToken := strings.TrimSpace(credentials["refresh_token"])
	if clientID == "" || refreshToken == "" {
		return nil, errors.New("OAuth refresh credentials are incomplete")
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", clientID)
	endpoint := hostingerOAuthTokenURL
	if provider == ops.ConnectorProviderCloudflare {
		endpoint = cloudflareOAuthTokenURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if provider == ops.ConnectorProviderCloudflare {
		if secret := strings.TrimSpace(os.Getenv("OPS_CLOUDFLARE_OAUTH_CLIENT_SECRET")); secret != "" {
			req.SetBasicAuth(clientID, secret)
		}
	}
	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || payload.AccessToken == "" {
		return nil, fmt.Errorf("OAuth refresh returned HTTP %d", response.StatusCode)
	}
	updated := cloneStringMap(credentials)
	updated["token"] = payload.AccessToken
	updated["access_token"] = payload.AccessToken
	if payload.RefreshToken != "" {
		updated["refresh_token"] = payload.RefreshToken
	}
	if payload.TokenType != "" {
		updated["token_type"] = payload.TokenType
	}
	if payload.ExpiresIn > 0 {
		updated["expires_at"] = time.Now().UTC().Add(time.Duration(opsOAuthMaxInt(payload.ExpiresIn-60, 1)) * time.Second).Format(time.RFC3339)
	}
	return updated, nil
}

func (s *OpsConnectorOAuthService) storeOAuthCredentials(connector ops.Connector, credentials map[string]string) error {
	if connector.Provider == ops.ConnectorProviderCloudflare {
		if strings.TrimSpace(credentials["client_id"]) == "" {
			credentials["client_id"] = strings.TrimSpace(os.Getenv("OPS_CLOUDFLARE_OAUTH_CLIENT_ID"))
		}
	}
	plaintext, err := json.Marshal(credentials)
	if err != nil {
		return err
	}
	encrypted, err := secretbox.EncryptString(string(plaintext), os.Getenv(OpsConnectorMasterKeyEnv))
	if err != nil {
		return fmt.Errorf("encrypt connector OAuth credentials: %w", err)
	}
	scopes := credentials["scope"]
	if scopes == "" {
		scopes = connector.Scopes
	}
	endpoint := connector.Endpoint
	if endpoint == "" {
		endpoint = defaultConnectorEndpoint(connector.Provider)
		if connector.Provider == ops.ConnectorProviderHostinger {
			endpoint = hostingerAPIBaseURL + "/api/vps/v1/virtual-machines"
		}
	} else if connector.Provider == ops.ConnectorProviderCloudflare &&
		(endpoint == cloudflareAPIBaseURL+"/user" ||
			endpoint == cloudflareAPIBaseURL+"/user/tokens/verify") {
		// Older OAuth-created connectors used endpoints that require user
		// permissions; zone discovery is the actual read capability needed here.
		endpoint = defaultConnectorEndpoint(connector.Provider)
	}
	return s.connectorRepo.UpdateOAuthCredentials(
		connector.ID,
		ops.ConnectorAuthBearer,
		endpoint,
		encrypted,
		encodeCredentialFields(credentials),
		scopes,
		ops.ConnectorStatusPending,
	)
}

func (s *OpsConnectorOAuthService) bindHostinger(
	ctx context.Context,
	connectorID uint,
	environment string,
	result *OpsConnectorOAuthCallbackResult,
) error {
	if s.vpsRepo == nil || s.projectRepo == nil {
		return errors.New("Hostinger binding repositories are not configured")
	}
	var remoteVPS []hostingerVirtualMachine
	if _, err := s.connectorService.HostingerRead(ctx, connectorID, "/api/vps/v1/virtual-machines", nil, &remoteVPS); err != nil {
		return fmt.Errorf("discover Hostinger VPS: %w", err)
	}
	existingVPS, err := s.vpsRepo.ListByEnvironment(environment)
	if err != nil {
		return err
	}
	boundByRemoteID := map[string]ops.VPSBinding{}
	for _, remote := range remoteVPS {
		remoteID := strconv.FormatUint(uint64(remote.ID), 10)
		match := findVPSBinding(existingVPS, remoteID)
		if match == nil && len(remoteVPS) == 1 {
			match = firstEnvironmentVPS(existingVPS, environment)
		}
		if match == nil {
			result.VPSCandidates = append(result.VPSCandidates, OpsOAuthResourceItem{
				ID:     remoteID,
				Name:   remote.Hostname,
				Status: remote.State,
			})
			result.BindingWarnings = append(result.BindingWarnings,
				fmt.Sprintf("Hostinger VPS %s was discovered but not auto-bound; select its resource ID in the VPS ledger", hostingerFirstNonEmpty(remote.Hostname, remoteID)),
			)
			continue
		}
		match.ConnectorID = &connectorID
		match.ProviderResourceID = remoteID
		match.Hostname = hostingerFirstNonEmpty(remote.Hostname, match.Hostname)
		match.IPv4 = hostingerFirstNonEmpty(firstHostingerIPv4(remote.IPv4), match.IPv4)
		if remote.Template != nil {
			match.OperatingSystem = hostingerFirstNonEmpty(remote.Template.Name, match.OperatingSystem)
		}
		match.Status = ops.VPSStatusActive
		if err := s.vpsRepo.Update(match); err != nil {
			return err
		}
		observedAt := time.Now().UTC()
		if err := s.vpsRepo.UpdateObservedState(
			match.ID,
			hostingerVPSObservedStatus(remote.State),
			remote.State,
			"hostinger:oauth",
			match.Hostname,
			match.IPv4,
			match.OperatingSystem,
			remote.Plan.Value,
			"",
			observedAt,
			"",
		); err != nil {
			return err
		}
		boundByRemoteID[remoteID] = *match
		result.BoundVPSCount++
	}

	projects, err := s.projectRepo.ListByEnvironment(environment)
	if err != nil {
		return err
	}
	boundProjectIDs := make([]uint, 0)
	for _, project := range projects {
		if project.Environment != environment || project.VPSBindingID == 0 {
			continue
		}
		var selected *ops.VPSBinding
		for _, vps := range boundByRemoteID {
			if vps.ID == project.VPSBindingID {
				selected = &vps
				break
			}
		}
		if selected == nil || strings.TrimSpace(project.ComposeProjectName) == "" {
			result.BindingWarnings = append(result.BindingWarnings,
				fmt.Sprintf("production project %s was not auto-bound because its VPS or Compose project name is incomplete", project.Name),
			)
			continue
		}
		vmPath := "/api/vps/v1/virtual-machines/" + url.PathEscape(selected.ProviderResourceID)
		var remoteProjects []hostingerDockerProject
		if _, err := s.connectorService.HostingerRead(ctx, connectorID, vmPath+"/docker", nil, &remoteProjects); err != nil {
			result.BindingWarnings = append(result.BindingWarnings,
				fmt.Sprintf("discover Docker projects for %s failed: %v", project.Name, err),
			)
			continue
		}
		projectMatched := false
		for _, remoteProject := range remoteProjects {
			if !strings.EqualFold(strings.TrimSpace(remoteProject.Name), strings.TrimSpace(project.ComposeProjectName)) {
				continue
			}
			projectMatched = true
			project.ConnectorID = &connectorID
			project.ProviderResourceID = remoteProject.Name
			project.Status = ops.ProjectStatusActive
			if err := s.projectRepo.Update(&project.ProjectBinding); err != nil {
				result.BindingWarnings = append(result.BindingWarnings,
					fmt.Sprintf("bind Hostinger project %s failed: %v", project.Name, err),
				)
				break
			}
			boundProjectIDs = append(boundProjectIDs, project.ID)
			result.BoundProjectCount++
			break
		}
		if !projectMatched {
			result.BindingWarnings = append(result.BindingWarnings,
				fmt.Sprintf("Hostinger Docker project %s was not found on VPS %s", project.ComposeProjectName, selected.Name),
			)
		}
	}

	if s.vpsSyncService != nil {
		for _, vps := range boundByRemoteID {
			if _, err := s.vpsSyncService.SyncVPS(ctx, vps.ID); err != nil {
				result.BindingWarnings = append(result.BindingWarnings,
					fmt.Sprintf("sync Hostinger VPS %s failed: %v", vps.Name, err),
				)
			}
		}
	}
	if s.projectSyncService != nil {
		for _, projectID := range boundProjectIDs {
			if _, err := s.projectSyncService.SyncProject(ctx, projectID); err != nil {
				result.BindingWarnings = append(result.BindingWarnings,
					fmt.Sprintf("sync Hostinger project %d failed: %v", projectID, err),
				)
			}
		}
	}
	return nil
}

func (s *OpsConnectorOAuthService) bindCloudflare(
	ctx context.Context,
	connectorID uint,
	environment string,
	result *OpsConnectorOAuthCallbackResult,
) error {
	if s.domainRepo == nil {
		return errors.New("Cloudflare binding repository is not configured")
	}
	var zones []cloudflareOAuthZone
	query := url.Values{"page": []string{"1"}, "per_page": []string{"100"}}
	if _, err := s.connectorService.CloudflareRead(ctx, connectorID, "/zones", query, &zones); err != nil {
		return fmt.Errorf("discover Cloudflare zones: %w", err)
	}
	for _, zone := range zones {
		result.CloudflareZones = append(result.CloudflareZones, OpsOAuthResourceItem{
			ID:     zone.ID,
			Name:   zone.Name,
			Status: zone.Status,
		})
	}
	domains, err := s.domainRepo.ListByEnvironment(environment)
	if err != nil {
		return err
	}
	boundDomainIDs := make([]uint, 0)
	for _, domain := range domains {
		if domain.Provider != ops.DomainProviderCloudflare || domain.Environment != environment {
			continue
		}
		zone := matchingCloudflareZone(domain, zones)
		if zone == nil {
			result.BindingWarnings = append(result.BindingWarnings,
				fmt.Sprintf("Cloudflare zone for domain %s was not found", domain.Domain),
			)
			continue
		}
		domain.ConnectorID = &connectorID
		domain.Zone = zone.Name
		domain.Status = ops.DomainStatusActive
		if err := s.domainRepo.Update(&domain); err != nil {
			result.BindingWarnings = append(result.BindingWarnings,
				fmt.Sprintf("bind Cloudflare domain %s failed: %v", domain.Domain, err),
			)
			continue
		}
		boundDomainIDs = append(boundDomainIDs, domain.ID)
		result.BoundDomainCount++
	}
	if s.domainSyncService != nil {
		for _, id := range boundDomainIDs {
			if _, err := s.domainSyncService.Sync(ctx, id); err != nil {
				result.BindingWarnings = append(result.BindingWarnings,
					fmt.Sprintf("sync Cloudflare domain %d failed: %v", id, err),
				)
			}
		}
	}
	return nil
}

type cloudflareOAuthZone struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Account struct {
		ID string `json:"id"`
	} `json:"account"`
}

func matchingCloudflareZone(domain ops.DomainBinding, zones []cloudflareOAuthZone) *cloudflareOAuthZone {
	domainName := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain.Domain)), ".")
	preferred := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain.Zone)), ".")
	var selected *cloudflareOAuthZone
	for index := range zones {
		zone := &zones[index]
		zoneName := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(zone.Name)), ".")
		if preferred != "" && preferred == zoneName {
			return zone
		}
		if domainName == zoneName || strings.HasSuffix(domainName, "."+zoneName) {
			if selected == nil || len(zoneName) > len(selected.Name) {
				selected = zone
			}
		}
	}
	return selected
}

func findVPSBinding(records []ops.VPSBinding, resourceID string) *ops.VPSBinding {
	for index := range records {
		if strings.TrimSpace(records[index].ProviderResourceID) == resourceID {
			return &records[index]
		}
	}
	return nil
}

func firstEnvironmentVPS(records []ops.VPSBinding, environment string) *ops.VPSBinding {
	for index := range records {
		if records[index].Environment == environment && records[index].Enabled {
			return &records[index]
		}
	}
	return nil
}

func normalizeOAuthProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ops.ConnectorProviderCloudflare:
		return ops.ConnectorProviderCloudflare
	case ops.ConnectorProviderHostinger:
		return ops.ConnectorProviderHostinger
	default:
		return ""
	}
}

func normalizeOAuthReturnPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return "/ops/connectors"
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return "/ops/connectors"
	}
	return value
}

func cloudflareOAuthScopes() string {
	return strings.TrimSpace(os.Getenv("OPS_CLOUDFLARE_OAUTH_SCOPES"))
}

func cloudflareOAuthScopesFromConnector(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, ":") {
		return cloudflareOAuthScopes()
	}
	return strings.Join(strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t'
	}), " ")
}

func randomOAuthString(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashOAuthState(value string) string {
	hash := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", hash[:])
}

func oauthPKCEChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func cloneStringMap(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func opsOAuthBoolPtr(value bool) *bool {
	return &value
}

func opsOAuthMaxInt(value, minimum int) int {
	if value < minimum {
		return minimum
	}
	return value
}
