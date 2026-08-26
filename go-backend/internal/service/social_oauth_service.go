package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"commerce-platform/internal/domain/social"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/secretbox"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

const (
	metaOAuthAuthorizationEndpoint = "https://www.facebook.com/%s/dialog/oauth"
	metaOAuthTokenEndpoint         = "https://graph.facebook.com/%s/oauth/access_token"
	metaGraphAPIEndpoint           = "https://graph.facebook.com/%s"
	xOAuthAuthorizationEndpoint    = "https://x.com/i/oauth2/authorize"
	xOAuthTokenEndpoint            = "https://api.x.com/2/oauth2/token"
	xAPIEndpoint                   = "https://api.x.com/2"
	googleOAuthAuthorizationURL    = "https://accounts.google.com/o/oauth2/v2/auth"
	googleOAuthTokenURL            = "https://oauth2.googleapis.com/token"
	googleUserInfoURL              = "https://openidconnect.googleapis.com/v1/userinfo"
	youtubeAPIEndpoint             = "https://www.googleapis.com/youtube/v3"
	redditOAuthAuthorizationURL    = "https://www.reddit.com/api/v1/authorize"
	redditOAuthTokenURL            = "https://www.reddit.com/api/v1/access_token"
	redditAPIEndpoint              = "https://oauth.reddit.com"
)

var (
	ErrSocialOAuthNotConfigured   = errors.New("social OAuth provider is not configured")
	ErrSocialOAuthProviderInvalid = errors.New("social OAuth provider is invalid")
	ErrSocialOAuthStateInvalid    = errors.New("social OAuth state is invalid or expired")
	ErrSocialOAuthExchange        = errors.New("social OAuth exchange failed")
	ErrSocialOAuthRemoteAPI       = errors.New("social OAuth provider API request failed")
)

type SocialOAuthStartInput struct {
	Provider   string `json:"provider"`
	ReturnPath string `json:"return_path"`
}

type SocialOAuthStartResult struct {
	Provider         string `json:"provider"`
	AuthorizationURL string `json:"authorization_url"`
}

type SocialOAuthConnectionView struct {
	Provider             string          `json:"provider"`
	Label                string          `json:"label"`
	Configured           bool            `json:"configured"`
	Connected            bool            `json:"connected"`
	Status               string          `json:"status"`
	ProviderAccountID    string          `json:"provider_account_id,omitempty"`
	ProviderAccountName  string          `json:"provider_account_name,omitempty"`
	ProviderAccountURL   string          `json:"provider_account_url,omitempty"`
	ProviderAccountEmail string          `json:"provider_account_email,omitempty"`
	GrantedScopes        []string        `json:"granted_scopes,omitempty"`
	ProviderResources    json.RawMessage `json:"provider_resources,omitempty"`
	TokenExpiresAt       *time.Time      `json:"token_expires_at,omitempty"`
	LastConnectedAt      *time.Time      `json:"last_connected_at,omitempty"`
	LastError            string          `json:"last_error,omitempty"`
}

type SocialOAuthListView struct {
	Connections []SocialOAuthConnectionView `json:"connections"`
}

type SocialOAuthCallbackResult struct {
	Provider   string
	ReturnPath string
	Status     string
	Message    string
}

type SocialOAuthService struct {
	connections *repository.SocialOAuthRepository
	config      config.SocialOAuthConfig
	httpClient  *http.Client
}

func NewSocialOAuthService(
	connections *repository.SocialOAuthRepository,
	cfg config.SocialOAuthConfig,
) *SocialOAuthService {
	return &SocialOAuthService{
		connections: connections,
		config:      cfg,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (s *SocialOAuthService) ConfigureHTTPClient(client *http.Client) {
	if client != nil {
		s.httpClient = client
	}
}

func (s *SocialOAuthService) ListConnections() (*SocialOAuthListView, error) {
	stored, err := s.connections.ListConnections()
	if err != nil {
		return nil, err
	}
	byProvider := make(map[string]social.OAuthConnection, len(stored))
	for _, connection := range stored {
		byProvider[connection.Provider] = connection
	}

	view := &SocialOAuthListView{
		Connections: make([]SocialOAuthConnectionView, 0, len(social.SupportedProviders)),
	}
	for _, provider := range social.SupportedProviders {
		connection, exists := byProvider[provider]
		if !exists {
			view.Connections = append(view.Connections, s.emptyConnectionView(provider))
			continue
		}
		view.Connections = append(view.Connections, s.connectionView(&connection))
	}
	return view, nil
}

func (s *SocialOAuthService) Start(userID uint, input SocialOAuthStartInput) (*SocialOAuthStartResult, error) {
	provider := normalizeSocialProvider(input.Provider)
	if provider == "" {
		return nil, ErrSocialOAuthProviderInvalid
	}
	providerConfig := s.providerConfig(provider)
	if !socialOAuthProviderConfigured(providerConfig) ||
		strings.TrimSpace(s.config.TokenEncryptionKey) == "" {
		return nil, fmt.Errorf("%w: %s", ErrSocialOAuthNotConfigured, provider)
	}

	state, err := randomOAuthValue(32)
	if err != nil {
		return nil, fmt.Errorf("generate social OAuth state: %w", err)
	}
	verifier := ""
	if socialOAuthProviderUsesPKCE(provider) {
		verifier, err = randomOAuthValue(48)
		if err != nil {
			return nil, fmt.Errorf("generate social OAuth verifier: %w", err)
		}
	}

	encryptedVerifier := ""
	if verifier != "" {
		encryptedVerifier, err = secretbox.EncryptString(verifier, s.config.TokenEncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt social OAuth verifier: %w", err)
		}
	}

	ttl := time.Duration(s.config.StateTTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	returnPath := normalizeSocialOAuthReturnPath(input.ReturnPath, s.config.PostConnectURL)
	session := &social.OAuthSession{
		Provider:              provider,
		StateHash:             hashSocialOAuthState(state),
		CodeVerifierEncrypted: encryptedVerifier,
		RedirectURI:           strings.TrimSpace(providerConfig.RedirectURL),
		ReturnPath:            returnPath,
		InitiatedByUserID:     userID,
		ExpiresAt:             time.Now().UTC().Add(ttl),
		Status:                social.OAuthSessionStatusPending,
	}
	_ = s.connections.DeleteExpiredSessions(time.Now().UTC())
	if err := s.connections.CreateSession(session); err != nil {
		return nil, err
	}

	authorizationURL, err := s.buildAuthorizationURL(provider, state, verifier)
	if err != nil {
		return nil, err
	}
	return &SocialOAuthStartResult{
		Provider:         provider,
		AuthorizationURL: authorizationURL,
	}, nil
}

func (s *SocialOAuthService) Complete(
	ctx context.Context,
	stateValue string,
	codeValue string,
	providerError string,
	providerErrorDescription string,
) (*SocialOAuthCallbackResult, error) {
	result := &SocialOAuthCallbackResult{
		ReturnPath: normalizeSocialOAuthReturnPath("", s.config.PostConnectURL),
	}
	stateValue = strings.TrimSpace(stateValue)
	if stateValue == "" {
		return result, ErrSocialOAuthStateInvalid
	}

	session, err := s.connections.ConsumeSession(hashSocialOAuthState(stateValue), time.Now().UTC())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return result, ErrSocialOAuthStateInvalid
		}
		return result, err
	}
	result.Provider = session.Provider
	result.ReturnPath = session.ReturnPath

	fail := func(callbackErr error) (*SocialOAuthCallbackResult, error) {
		result.Status = social.OAuthStatusError
		result.Message = callbackErr.Error()
		_ = s.connections.MarkSessionError(session.ID, callbackErr.Error())
		_ = s.saveConnectionError(session.Provider, callbackErr.Error())
		return result, callbackErr
	}

	if providerError != "" {
		message := strings.TrimSpace(providerErrorDescription)
		if message == "" {
			message = strings.TrimSpace(providerError)
		}
		return fail(fmt.Errorf("%w: provider denied authorization: %s", ErrSocialOAuthExchange, message))
	}
	codeValue = strings.TrimSpace(codeValue)
	if codeValue == "" {
		return fail(fmt.Errorf("%w: authorization code is missing", ErrSocialOAuthExchange))
	}

	verifier := ""
	if session.CodeVerifierEncrypted != "" {
		verifier, err = secretbox.DecryptString(session.CodeVerifierEncrypted, s.config.TokenEncryptionKey)
		if err != nil {
			return fail(fmt.Errorf("%w: stored PKCE verifier cannot be decrypted", ErrSocialOAuthExchange))
		}
	}

	token, profile, err := s.exchangeAndFetchProfile(ctx, session.Provider, session.RedirectURI, codeValue, verifier)
	if err != nil {
		return fail(err)
	}
	if strings.TrimSpace(token.AccessToken) == "" {
		return fail(fmt.Errorf("%w: access token is missing", ErrSocialOAuthExchange))
	}

	connection, err := s.connections.FindConnection(session.Provider)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		connection = &social.OAuthConnection{
			Provider: session.Provider,
		}
	} else if err != nil {
		return result, err
	}

	accessBundle := socialAccessTokenBundle{
		AccessToken:    token.AccessToken,
		ResourceTokens: profile.ResourceTokens,
	}
	accessPlaintext, err := json.Marshal(accessBundle)
	if err != nil {
		return fail(fmt.Errorf("%w: encode access token bundle", ErrSocialOAuthExchange))
	}
	connection.AccessTokenEncrypted, err = secretbox.EncryptString(string(accessPlaintext), s.config.TokenEncryptionKey)
	if err != nil {
		return fail(fmt.Errorf("%w: encrypt access token", ErrSocialOAuthExchange))
	}

	refreshToken := strings.TrimSpace(token.RefreshToken)
	if refreshToken != "" {
		connection.RefreshTokenEncrypted, err = secretbox.EncryptString(refreshToken, s.config.TokenEncryptionKey)
		if err != nil {
			return fail(fmt.Errorf("%w: encrypt refresh token", ErrSocialOAuthExchange))
		}
	}

	now := time.Now().UTC()
	connection.Provider = session.Provider
	connection.Status = social.OAuthStatusConnected
	connection.ProviderAccountID = profile.AccountID
	connection.ProviderAccountName = profile.AccountName
	connection.ProviderAccountURL = profile.AccountURL
	connection.ProviderAccountEmail = profile.AccountEmail
	connection.GrantedScopes = socialFirstNonEmpty(strings.TrimSpace(token.Scope), s.providerScopes(session.Provider))
	connection.ProviderResources = profile.ResourcesJSON()
	connection.LastConnectedAt = &now
	connection.LastError = ""
	if token.ExpiresIn > 0 {
		expiresAt := now.Add(time.Duration(token.ExpiresIn) * time.Second)
		connection.TokenExpiresAt = &expiresAt
	} else {
		connection.TokenExpiresAt = nil
	}
	if err := s.connections.SaveConnection(connection); err != nil {
		return result, err
	}

	result.Status = social.OAuthStatusConnected
	result.Message = fmt.Sprintf("%s connected", socialProviderLabel(session.Provider))
	return result, nil
}

func (s *SocialOAuthService) Disconnect(providerValue string) error {
	provider := normalizeSocialProvider(providerValue)
	if provider == "" {
		return ErrSocialOAuthProviderInvalid
	}
	connection, err := s.connections.FindConnection(provider)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	connection.Status = social.OAuthStatusDisconnected
	connection.ProviderAccountID = ""
	connection.ProviderAccountName = ""
	connection.ProviderAccountURL = ""
	connection.ProviderAccountEmail = ""
	connection.AccessTokenEncrypted = ""
	connection.RefreshTokenEncrypted = ""
	connection.TokenExpiresAt = nil
	connection.GrantedScopes = ""
	connection.ProviderResources = []byte("{}")
	connection.LastConnectedAt = nil
	connection.LastError = ""
	return s.connections.SaveConnection(connection)
}

func (s *SocialOAuthService) buildAuthorizationURL(provider, state, verifier string) (string, error) {
	cfg := s.providerConfig(provider)
	query := url.Values{}
	query.Set("client_id", strings.TrimSpace(cfg.ClientID))
	query.Set("redirect_uri", strings.TrimSpace(cfg.RedirectURL))
	query.Set("response_type", "code")
	query.Set("scope", s.providerScopes(provider))
	query.Set("state", state)

	switch provider {
	case social.ProviderMeta:
		return fmt.Sprintf(metaOAuthAuthorizationEndpoint, s.metaGraphAPIVersion()) + "?" + query.Encode(), nil
	case social.ProviderX:
		query.Set("code_challenge", pkceChallenge(verifier))
		query.Set("code_challenge_method", "S256")
		return xOAuthAuthorizationEndpoint + "?" + query.Encode(), nil
	case social.ProviderYouTube:
		query.Set("access_type", "offline")
		query.Set("prompt", "consent")
		query.Set("include_granted_scopes", "true")
		query.Set("code_challenge", pkceChallenge(verifier))
		query.Set("code_challenge_method", "S256")
		return googleOAuthAuthorizationURL + "?" + query.Encode(), nil
	case social.ProviderReddit:
		query.Set("duration", "permanent")
		return redditOAuthAuthorizationURL + "?" + query.Encode(), nil
	default:
		return "", ErrSocialOAuthProviderInvalid
	}
}

func (s *SocialOAuthService) exchangeAndFetchProfile(
	ctx context.Context,
	provider string,
	redirectURI string,
	code string,
	verifier string,
) (*socialOAuthToken, *socialOAuthProfile, error) {
	switch provider {
	case social.ProviderMeta:
		return s.exchangeMeta(ctx, redirectURI, code)
	case social.ProviderX:
		return s.exchangeX(ctx, redirectURI, code, verifier)
	case social.ProviderYouTube:
		return s.exchangeYouTube(ctx, redirectURI, code, verifier)
	case social.ProviderReddit:
		return s.exchangeReddit(ctx, redirectURI, code)
	default:
		return nil, nil, ErrSocialOAuthProviderInvalid
	}
}

func (s *SocialOAuthService) exchangeMeta(ctx context.Context, redirectURI, code string) (*socialOAuthToken, *socialOAuthProfile, error) {
	cfg := s.config.Meta
	query := url.Values{}
	query.Set("client_id", strings.TrimSpace(cfg.ClientID))
	query.Set("client_secret", strings.TrimSpace(cfg.ClientSecret))
	query.Set("redirect_uri", redirectURI)
	query.Set("code", code)

	shortToken, err := s.getToken(ctx, fmt.Sprintf(metaOAuthTokenEndpoint, s.metaGraphAPIVersion()), query, nil, "")
	if err != nil {
		return nil, nil, fmt.Errorf("%w: Meta token exchange: %v", ErrSocialOAuthExchange, err)
	}

	longLivedQuery := url.Values{}
	longLivedQuery.Set("grant_type", "fb_exchange_token")
	longLivedQuery.Set("client_id", strings.TrimSpace(cfg.ClientID))
	longLivedQuery.Set("client_secret", strings.TrimSpace(cfg.ClientSecret))
	longLivedQuery.Set("fb_exchange_token", shortToken.AccessToken)
	longLivedToken, err := s.getToken(ctx, fmt.Sprintf(metaOAuthTokenEndpoint, s.metaGraphAPIVersion()), longLivedQuery, nil, "")
	if err != nil {
		return nil, nil, fmt.Errorf("%w: Meta long-lived token exchange: %v", ErrSocialOAuthExchange, err)
	}
	if longLivedToken.ExpiresIn == 0 {
		longLivedToken.ExpiresIn = shortToken.ExpiresIn
	}
	if longLivedToken.Scope == "" {
		longLivedToken.Scope = shortToken.Scope
	}

	profile, err := s.fetchMetaProfile(ctx, longLivedToken.AccessToken)
	if err != nil {
		return nil, nil, err
	}
	return longLivedToken, profile, nil
}

func (s *SocialOAuthService) exchangeX(ctx context.Context, redirectURI, code, verifier string) (*socialOAuthToken, *socialOAuthProfile, error) {
	cfg := s.config.X
	form := url.Values{}
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", strings.TrimSpace(cfg.ClientID))
	form.Set("redirect_uri", redirectURI)
	form.Set("code_verifier", verifier)
	token, err := s.postTokenForm(ctx, xOAuthTokenEndpoint, form, cfg.ClientID, cfg.ClientSecret, "social-oauth-x/1.0")
	if err != nil {
		return nil, nil, fmt.Errorf("%w: X token exchange: %v", ErrSocialOAuthExchange, err)
	}

	var payload struct {
		Data struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			Username        string `json:"username"`
			ProfileImageURL string `json:"profile_image_url"`
			URL             string `json:"url"`
		} `json:"data"`
	}
	if err := s.getJSON(ctx, xAPIEndpoint+"/users/me?user.fields=id,name,username,profile_image_url,url", token.AccessToken, "social-oauth-x/1.0", &payload); err != nil {
		return nil, nil, fmt.Errorf("%w: X account lookup: %v", ErrSocialOAuthRemoteAPI, err)
	}
	if strings.TrimSpace(payload.Data.ID) == "" {
		return nil, nil, fmt.Errorf("%w: X account id is missing", ErrSocialOAuthRemoteAPI)
	}
	accountURL := strings.TrimSpace(payload.Data.URL)
	if accountURL == "" && payload.Data.Username != "" {
		accountURL = "https://x.com/" + url.PathEscape(payload.Data.Username)
	}
	profile := &socialOAuthProfile{
		AccountID:   payload.Data.ID,
		AccountName: socialFirstNonEmpty(payload.Data.Name, payload.Data.Username),
		AccountURL:  accountURL,
		Resources: map[string]interface{}{
			"username":          payload.Data.Username,
			"profile_image_url": payload.Data.ProfileImageURL,
		},
	}
	return token, profile, nil
}

func (s *SocialOAuthService) exchangeYouTube(ctx context.Context, redirectURI, code, verifier string) (*socialOAuthToken, *socialOAuthProfile, error) {
	cfg := s.config.YouTube
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", strings.TrimSpace(cfg.ClientID))
	form.Set("client_secret", strings.TrimSpace(cfg.ClientSecret))
	form.Set("redirect_uri", redirectURI)
	form.Set("grant_type", "authorization_code")
	form.Set("code_verifier", verifier)
	token, err := s.postTokenForm(ctx, googleOAuthTokenURL, form, "", "", "social-oauth-youtube/1.0")
	if err != nil {
		return nil, nil, fmt.Errorf("%w: YouTube token exchange: %v", ErrSocialOAuthExchange, err)
	}

	var channels struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title      string `json:"title"`
				CustomURL  string `json:"customUrl"`
				Thumbnails map[string]struct {
					URL string `json:"url"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			ContentDetails struct {
				RelatedPlaylists struct {
					Uploads string `json:"uploads"`
				} `json:"relatedPlaylists"`
			} `json:"contentDetails"`
		} `json:"items"`
	}
	endpoint := youtubeAPIEndpoint + "/channels?part=snippet,contentDetails&mine=true"
	if err := s.getJSON(ctx, endpoint, token.AccessToken, "social-oauth-youtube/1.0", &channels); err != nil {
		return nil, nil, fmt.Errorf("%w: YouTube channel lookup: %v", ErrSocialOAuthRemoteAPI, err)
	}
	if len(channels.Items) == 0 || strings.TrimSpace(channels.Items[0].ID) == "" {
		return nil, nil, fmt.Errorf("%w: no YouTube channel is available for this account", ErrSocialOAuthRemoteAPI)
	}

	channel := channels.Items[0]
	profile := &socialOAuthProfile{
		AccountID:   channel.ID,
		AccountName: channel.Snippet.Title,
		AccountURL:  "https://www.youtube.com/channel/" + url.PathEscape(channel.ID),
		Resources: map[string]interface{}{
			"custom_url":       channel.Snippet.CustomURL,
			"uploads_playlist": channel.ContentDetails.RelatedPlaylists.Uploads,
			"thumbnails":       channel.Snippet.Thumbnails,
		},
	}
	var userInfo struct {
		Email string `json:"email"`
	}
	if err := s.getJSON(ctx, googleUserInfoURL, token.AccessToken, "social-oauth-youtube/1.0", &userInfo); err == nil {
		profile.AccountEmail = strings.TrimSpace(userInfo.Email)
	}
	return token, profile, nil
}

func (s *SocialOAuthService) exchangeReddit(ctx context.Context, redirectURI, code string) (*socialOAuthToken, *socialOAuthProfile, error) {
	cfg := s.config.Reddit
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	token, err := s.postTokenForm(ctx, redditOAuthTokenURL, form, cfg.ClientID, cfg.ClientSecret, "social-oauth-reddit/1.0")
	if err != nil {
		return nil, nil, fmt.Errorf("%w: Reddit token exchange: %v", ErrSocialOAuthExchange, err)
	}

	var payload struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		IconImg   string `json:"icon_img"`
		Subreddit struct {
			DisplayName string `json:"display_name"`
			Title       string `json:"title"`
		} `json:"subreddit"`
	}
	if err := s.getJSON(ctx, redditAPIEndpoint+"/api/v1/me", token.AccessToken, "social-oauth-reddit/1.0", &payload); err != nil {
		return nil, nil, fmt.Errorf("%w: Reddit account lookup: %v", ErrSocialOAuthRemoteAPI, err)
	}
	if strings.TrimSpace(payload.ID) == "" || strings.TrimSpace(payload.Name) == "" {
		return nil, nil, fmt.Errorf("%w: Reddit account identity is missing", ErrSocialOAuthRemoteAPI)
	}
	profile := &socialOAuthProfile{
		AccountID:   payload.ID,
		AccountName: payload.Name,
		AccountURL:  "https://www.reddit.com/user/" + url.PathEscape(payload.Name),
		Resources: map[string]interface{}{
			"icon_img":  payload.IconImg,
			"subreddit": payload.Subreddit,
		},
	}
	return token, profile, nil
}

func (s *SocialOAuthService) fetchMetaProfile(ctx context.Context, accessToken string) (*socialOAuthProfile, error) {
	var me struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Link string `json:"link"`
	}
	meEndpoint := fmt.Sprintf(metaGraphAPIEndpoint, s.metaGraphAPIVersion()) + "/me?fields=id,name,link"
	if err := s.getJSON(ctx, meEndpoint, accessToken, "social-oauth-meta/1.0", &me); err != nil {
		return nil, fmt.Errorf("%w: Meta account lookup: %v", ErrSocialOAuthRemoteAPI, err)
	}
	if strings.TrimSpace(me.ID) == "" {
		return nil, fmt.Errorf("%w: Meta account id is missing", ErrSocialOAuthRemoteAPI)
	}

	var pages struct {
		Data []struct {
			ID                       string `json:"id"`
			Name                     string `json:"name"`
			Link                     string `json:"link"`
			AccessToken              string `json:"access_token"`
			InstagramBusinessAccount *struct {
				ID                string `json:"id"`
				Username          string `json:"username"`
				Name              string `json:"name"`
				ProfilePictureURL string `json:"profile_picture_url"`
			} `json:"instagram_business_account"`
		} `json:"data"`
	}
	pagesEndpoint := fmt.Sprintf(metaGraphAPIEndpoint, s.metaGraphAPIVersion()) +
		"/me/accounts?fields=id,name,link,access_token,instagram_business_account{id,username,name,profile_picture_url}"
	if err := s.getJSON(ctx, pagesEndpoint, accessToken, "social-oauth-meta/1.0", &pages); err != nil {
		return nil, fmt.Errorf("%w: Facebook Page lookup: %v", ErrSocialOAuthRemoteAPI, err)
	}

	resourceTokens := make(map[string]string)
	pageResources := make([]map[string]interface{}, 0, len(pages.Data))
	instagramResources := make([]map[string]interface{}, 0)
	for _, page := range pages.Data {
		pageID := strings.TrimSpace(page.ID)
		if pageID == "" {
			continue
		}
		if page.AccessToken != "" {
			resourceTokens["page:"+pageID] = page.AccessToken
		}
		pageURL := socialFirstNonEmpty(page.Link, "https://www.facebook.com/"+url.PathEscape(pageID))
		pageResource := map[string]interface{}{
			"id":   pageID,
			"name": page.Name,
			"url":  pageURL,
		}
		if page.InstagramBusinessAccount != nil && page.InstagramBusinessAccount.ID != "" {
			instagram := page.InstagramBusinessAccount
			instagramURL := "https://www.instagram.com/"
			if instagram.Username != "" {
				instagramURL += url.PathEscape(instagram.Username) + "/"
			}
			pageResource["instagram_account_id"] = instagram.ID
			pageResource["instagram_username"] = instagram.Username
			pageResource["instagram_name"] = instagram.Name
			pageResource["instagram_url"] = instagramURL
			instagramResources = append(instagramResources, map[string]interface{}{
				"id":                  instagram.ID,
				"username":            instagram.Username,
				"name":                instagram.Name,
				"url":                 instagramURL,
				"profile_picture_url": instagram.ProfilePictureURL,
				"page_id":             pageID,
			})
		}
		pageResources = append(pageResources, pageResource)
	}

	profileURL := socialFirstNonEmpty(me.Link, "https://www.facebook.com/"+url.PathEscape(me.ID))
	return &socialOAuthProfile{
		AccountID:      me.ID,
		AccountName:    me.Name,
		AccountURL:     profileURL,
		ResourceTokens: resourceTokens,
		Resources: map[string]interface{}{
			"pages":              pageResources,
			"instagram_accounts": instagramResources,
		},
	}, nil
}

func (s *SocialOAuthService) getToken(
	ctx context.Context,
	endpoint string,
	query url.Values,
	headers map[string]string,
	userAgent string,
) (*socialOAuthToken, error) {
	requestURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	requestURL.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	applySocialOAuthHeaders(req, headers, userAgent)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return decodeSocialOAuthTokenResponse(resp)
}

func (s *SocialOAuthService) postTokenForm(
	ctx context.Context,
	endpoint string,
	form url.Values,
	basicUser string,
	basicPassword string,
	userAgent string,
) (*socialOAuthToken, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if basicUser != "" || basicPassword != "" {
		req.SetBasicAuth(basicUser, basicPassword)
	}
	applySocialOAuthHeaders(req, nil, userAgent)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return decodeSocialOAuthTokenResponse(resp)
}

func (s *SocialOAuthService) getJSON(
	ctx context.Context,
	endpoint string,
	accessToken string,
	userAgent string,
	target interface{},
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(accessToken) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	}
	applySocialOAuthHeaders(req, nil, userAgent)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("provider returned HTTP %d: %s", resp.StatusCode, socialOAuthResponseMessage(body))
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("invalid provider JSON response: %w", err)
	}
	return nil
}

func decodeSocialOAuthTokenResponse(resp *http.Response) (*socialOAuthToken, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	var payload struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		ExpiresIn        int    `json:"expires_in"`
		Scope            string `json:"scope"`
		TokenType        string `json:"token_type"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid provider token JSON response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || payload.Error != "" || payload.AccessToken == "" {
		message := socialFirstNonEmpty(payload.ErrorDescription, payload.Error, socialOAuthResponseMessage(body))
		return nil, fmt.Errorf("provider returned HTTP %d: %s", resp.StatusCode, message)
	}
	return &socialOAuthToken{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresIn:    payload.ExpiresIn,
		Scope:        payload.Scope,
		TokenType:    payload.TokenType,
	}, nil
}

func (s *SocialOAuthService) saveConnectionError(provider, message string) error {
	connection, err := s.connections.FindConnection(provider)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		connection = &social.OAuthConnection{Provider: provider}
	} else if err != nil {
		return err
	}
	connection.Status = social.OAuthStatusError
	connection.LastError = message
	return s.connections.SaveConnection(connection)
}

func (s *SocialOAuthService) connectionView(connection *social.OAuthConnection) SocialOAuthConnectionView {
	resources := json.RawMessage(connection.ProviderResources)
	if len(resources) == 0 || !json.Valid(resources) {
		resources = json.RawMessage("{}")
	}
	return SocialOAuthConnectionView{
		Provider:             connection.Provider,
		Label:                socialProviderLabel(connection.Provider),
		Configured:           socialOAuthProviderConfigured(s.providerConfig(connection.Provider)),
		Connected:            connection.Status == social.OAuthStatusConnected && connection.AccessTokenEncrypted != "",
		Status:               socialFirstNonEmpty(connection.Status, social.OAuthStatusDisconnected),
		ProviderAccountID:    connection.ProviderAccountID,
		ProviderAccountName:  connection.ProviderAccountName,
		ProviderAccountURL:   connection.ProviderAccountURL,
		ProviderAccountEmail: connection.ProviderAccountEmail,
		GrantedScopes:        strings.Fields(connection.GrantedScopes),
		ProviderResources:    resources,
		TokenExpiresAt:       connection.TokenExpiresAt,
		LastConnectedAt:      connection.LastConnectedAt,
		LastError:            connection.LastError,
	}
}

func (s *SocialOAuthService) emptyConnectionView(provider string) SocialOAuthConnectionView {
	return SocialOAuthConnectionView{
		Provider:          provider,
		Label:             socialProviderLabel(provider),
		Configured:        socialOAuthProviderConfigured(s.providerConfig(provider)),
		Connected:         false,
		Status:            social.OAuthStatusDisconnected,
		ProviderResources: json.RawMessage("{}"),
	}
}

func (s *SocialOAuthService) providerConfig(provider string) config.SocialOAuthProviderConfig {
	switch provider {
	case social.ProviderMeta:
		return s.config.Meta
	case social.ProviderX:
		return s.config.X
	case social.ProviderYouTube:
		return s.config.YouTube
	case social.ProviderReddit:
		return s.config.Reddit
	default:
		return config.SocialOAuthProviderConfig{}
	}
}

func (s *SocialOAuthService) providerScopes(provider string) string {
	cfg := s.providerConfig(provider)
	if value := normalizeOAuthScopes(cfg.Scopes); value != "" {
		return value
	}
	switch provider {
	case social.ProviderMeta:
		return "pages_show_list pages_read_engagement pages_manage_posts instagram_basic instagram_content_publish"
	case social.ProviderX:
		return "users.read tweet.read tweet.write offline.access"
	case social.ProviderYouTube:
		return "openid email profile https://www.googleapis.com/auth/youtube.readonly https://www.googleapis.com/auth/youtube.upload"
	case social.ProviderReddit:
		return "identity read submit"
	default:
		return ""
	}
}

func (s *SocialOAuthService) metaGraphAPIVersion() string {
	version := strings.TrimSpace(s.config.MetaGraphAPIVersion)
	if version == "" {
		return "v23.0"
	}
	return strings.TrimPrefix(version, "/")
}

type socialOAuthToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	Scope        string
	TokenType    string
}

type socialOAuthProfile struct {
	AccountID      string
	AccountName    string
	AccountURL     string
	AccountEmail   string
	ResourceTokens map[string]string
	Resources      map[string]interface{}
}

func (p *socialOAuthProfile) ResourcesJSON() []byte {
	if p == nil || p.Resources == nil {
		return []byte("{}")
	}
	data, err := json.Marshal(p.Resources)
	if err != nil {
		return []byte("{}")
	}
	return data
}

type socialAccessTokenBundle struct {
	AccessToken    string            `json:"access_token"`
	ResourceTokens map[string]string `json:"resource_tokens,omitempty"`
}

func socialOAuthProviderConfigured(cfg config.SocialOAuthProviderConfig) bool {
	return strings.TrimSpace(cfg.ClientID) != "" &&
		strings.TrimSpace(cfg.ClientSecret) != "" &&
		strings.TrimSpace(cfg.RedirectURL) != ""
}

func socialOAuthProviderUsesPKCE(provider string) bool {
	return provider == social.ProviderX || provider == social.ProviderYouTube
}

func normalizeSocialProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case social.ProviderMeta:
		return social.ProviderMeta
	case social.ProviderX:
		return social.ProviderX
	case social.ProviderYouTube:
		return social.ProviderYouTube
	case social.ProviderReddit:
		return social.ProviderReddit
	default:
		return ""
	}
}

func socialProviderLabel(provider string) string {
	switch provider {
	case social.ProviderMeta:
		return "Facebook / Instagram"
	case social.ProviderX:
		return "X"
	case social.ProviderYouTube:
		return "YouTube"
	case social.ProviderReddit:
		return "Reddit"
	default:
		return provider
	}
}

func normalizeOAuthScopes(value string) string {
	value = strings.NewReplacer(",", " ", "\n", " ", "\r", " ", "\t", " ").Replace(value)
	return strings.Join(strings.Fields(value), " ")
}

func normalizeSocialOAuthReturnPath(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		parsed, err := url.Parse(strings.TrimSpace(fallback))
		if err == nil && parsed.Path != "" {
			value = parsed.Path
			if parsed.RawQuery != "" {
				value += "?" + parsed.RawQuery
			}
		}
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(parsed.Path, "//") {
		return "/social"
	}
	if parsed.Path == "" {
		return "/social"
	}
	if parsed.RawQuery == "" {
		return parsed.Path
	}
	return parsed.Path + "?" + parsed.RawQuery
}

func applySocialOAuthHeaders(req *http.Request, headers map[string]string, userAgent string) {
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(userAgent) != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

func socialOAuthResponseMessage(body []byte) string {
	var payload struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		Message          string `json:"message"`
	}
	if json.Unmarshal(body, &payload) == nil {
		if message := socialFirstNonEmpty(payload.ErrorDescription, payload.Message, payload.Error); message != "" {
			return message
		}
	}
	message := strings.TrimSpace(string(body))
	if len(message) > 300 {
		return message[:300]
	}
	if message == "" {
		return "empty provider response"
	}
	return message
}

func randomOAuthValue(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashSocialOAuthState(value string) string {
	hash := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", hash[:])
}

func pkceChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func socialFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
