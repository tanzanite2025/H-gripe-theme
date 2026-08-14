package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/pkg/secretbox"
	"commerce-platform/internal/repository"
)

const OpsConnectorMasterKeyEnv = "OPS_CONNECTOR_MASTER_KEY"

const (
	cloudflareAPIBaseURL = "https://api.cloudflare.com/client/v4"
	hostingerAPIBaseURL  = "https://developers.hostinger.com"
)

var (
	ErrInvalidOpsConnector = errors.New("invalid operations connector")
)

type OpsConnectorService struct {
	repo       *repository.OpsConnectorRepository
	httpClient *http.Client
}

type OpsHostingerUpdateResult struct {
	StatusCode  int    `json:"status_code"`
	OperationID string `json:"operation_id,omitempty"`
	Message     string `json:"message"`
}

type OpsConnectorInput struct {
	Name          string            `json:"name"`
	Provider      string            `json:"provider"`
	Environment   string            `json:"environment"`
	Endpoint      string            `json:"endpoint"`
	AuthType      string            `json:"auth_type"`
	CredentialRef string            `json:"credential_ref"`
	Credentials   map[string]string `json:"credentials"`
	Scopes        string            `json:"scopes"`
	Status        string            `json:"status"`
	Enabled       *bool             `json:"enabled"`
	Notes         string            `json:"notes"`
}

func NewOpsConnectorService(repo *repository.OpsConnectorRepository) *OpsConnectorService {
	return &OpsConnectorService{
		repo:       repo,
		httpClient: &http.Client{Timeout: 12 * time.Second},
	}
}

func (s *OpsConnectorService) List() ([]ops.ConnectorView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations connector service is not configured")
	}
	records, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	views := make([]ops.ConnectorView, 0, len(records))
	for _, record := range records {
		views = append(views, connectorView(record))
	}
	return views, nil
}

func (s *OpsConnectorService) Get(id uint) (*ops.ConnectorView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations connector service is not configured")
	}
	record, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	view := connectorView(*record)
	return &view, nil
}

func (s *OpsConnectorService) Create(input OpsConnectorInput) (*ops.ConnectorView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations connector service is not configured")
	}
	record, err := s.normalizeInput(input, nil)
	if err != nil {
		return nil, err
	}
	if existing, err := s.repo.FindByName(record.Name); err == nil && existing != nil {
		return nil, fmt.Errorf("%w: connector name already exists", ErrInvalidOpsConnector)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.Create(&record); err != nil {
		if repository.IsDuplicatedKey(err) {
			return nil, fmt.Errorf("%w: connector name already exists", ErrInvalidOpsConnector)
		}
		return nil, err
	}
	view := connectorView(record)
	return &view, nil
}

func (s *OpsConnectorService) Update(id uint, input OpsConnectorInput) (*ops.ConnectorView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations connector service is not configured")
	}
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	record, err := s.normalizeInput(input, existing)
	if err != nil {
		return nil, err
	}
	record.ID = id
	if other, err := s.repo.FindByName(record.Name); err == nil && other != nil && other.ID != id {
		return nil, fmt.Errorf("%w: connector name already exists", ErrInvalidOpsConnector)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.Update(&record); err != nil {
		if repository.IsDuplicatedKey(err) {
			return nil, fmt.Errorf("%w: connector name already exists", ErrInvalidOpsConnector)
		}
		return nil, err
	}
	view := connectorView(record)
	return &view, nil
}

func (s *OpsConnectorService) SetEnabled(id uint, enabled bool) (*ops.ConnectorView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations connector service is not configured")
	}
	status := ops.ConnectorStatusDisabled
	if enabled {
		status = ops.ConnectorStatusPending
	}
	if err := s.repo.UpdateEnabled(id, enabled, status); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *OpsConnectorService) Test(ctx context.Context, id uint) (*ops.ConnectorTestResult, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations connector service is not configured")
	}
	record, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	checkedAt := time.Now().UTC()
	result := &ops.ConnectorTestResult{
		ConnectorID:          id,
		CheckedAt:            checkedAt,
		CredentialConfigured: connectorCredentialConfigured(*record),
	}

	if !record.Enabled {
		result.Message = "connector is disabled"
		return s.saveTestResult(result, record, false)
	}
	if record.AuthType == ops.ConnectorAuthManual {
		result.Message = "manual connector does not support an automated test"
		return s.saveTestResult(result, record, false)
	}

	credentials, err := s.readCredentials(*record)
	if err != nil {
		result.Message = err.Error()
		return s.saveTestResult(result, record, false)
	}
	if requiresCredentials(record.AuthType) && len(credentials) == 0 {
		result.Message = "connector credentials are not configured"
		return s.saveTestResult(result, record, false)
	}

	endpoint := strings.TrimSpace(record.Endpoint)
	if endpoint == "" {
		endpoint = defaultConnectorEndpoint(record.Provider)
	}
	if endpoint == "" {
		result.Message = "connector test endpoint is required"
		return s.saveTestResult(result, record, false)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		result.Message = fmt.Sprintf("invalid connector endpoint: %v", err)
		return s.saveTestResult(result, record, false)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "tanzanite-ops-connector/1.0")
	applyConnectorAuth(req, record.AuthType, credentials)

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	if err := ensureConnectorEndpointSafe(ctx, req.URL, record.Environment); err != nil {
		result.Message = err.Error()
		return s.saveTestResult(result, record, false)
	}
	clientCopy := *client
	clientCopy.CheckRedirect = func(redirectRequest *http.Request, _ []*http.Request) error {
		return ensureConnectorEndpointSafe(redirectRequest.Context(), redirectRequest.URL, record.Environment)
	}
	response, err := clientCopy.Do(req)
	if err != nil {
		result.Message = fmt.Sprintf("connector request failed: %v", err)
		return s.saveTestResult(result, record, false)
	}
	defer response.Body.Close()
	result.StatusCode = response.StatusCode

	success := response.StatusCode >= 200 && response.StatusCode < 300
	if record.Provider == ops.ConnectorProviderCloudflare {
		var payload struct {
			Success bool `json:"success"`
		}
		if err := json.NewDecoder(response.Body).Decode(&payload); err == nil {
			success = success && payload.Success
		}
	}
	if success {
		result.Success = true
		result.Message = "connector test succeeded"
		return s.saveTestResult(result, record, true)
	}

	result.Message = fmt.Sprintf("remote API returned HTTP %d", response.StatusCode)
	return s.saveTestResult(result, record, false)
}

func (s *OpsConnectorService) CloudflareRead(ctx context.Context, id uint, path string, query url.Values, target interface{}) (int, error) {
	if s == nil || s.repo == nil {
		return 0, errors.New("operations connector service is not configured")
	}
	if !strings.HasPrefix(path, "/zones") {
		return 0, fmt.Errorf("%w: Cloudflare read path is not allowed", ErrInvalidOpsConnector)
	}

	record, err := s.repo.FindByID(id)
	if err != nil {
		return 0, err
	}
	if record.Provider != ops.ConnectorProviderCloudflare {
		return 0, fmt.Errorf("%w: connector is not a Cloudflare connector", ErrInvalidOpsConnector)
	}
	if !record.Enabled {
		return 0, errors.New("Cloudflare connector is disabled")
	}

	credentials, err := s.readCredentials(*record)
	if err != nil {
		return 0, err
	}
	if len(credentials) == 0 {
		return 0, errors.New("Cloudflare connector credentials are not configured")
	}

	endpoint, err := url.Parse(cloudflareAPIBaseURL + path)
	if err != nil {
		return 0, fmt.Errorf("build Cloudflare request: %w", err)
	}
	endpoint.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("build Cloudflare request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "tanzanite-ops-cloudflare-read/1.0")
	applyConnectorAuth(req, record.AuthType, credentials)

	if err := ensureConnectorEndpointSafe(ctx, req.URL, record.Environment); err != nil {
		return 0, err
	}
	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	clientCopy := *client
	clientCopy.CheckRedirect = func(redirectRequest *http.Request, _ []*http.Request) error {
		return ensureConnectorEndpointSafe(redirectRequest.Context(), redirectRequest.URL, record.Environment)
	}

	response, err := clientCopy.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Cloudflare request failed: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return response.StatusCode, fmt.Errorf("read Cloudflare response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return response.StatusCode, fmt.Errorf("Cloudflare API returned HTTP %d", response.StatusCode)
	}

	var envelope struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return response.StatusCode, errors.New("Cloudflare response was not valid JSON")
	}
	if !envelope.Success {
		return response.StatusCode, errors.New("Cloudflare API reported an unsuccessful response")
	}
	if target != nil && len(envelope.Result) > 0 && string(envelope.Result) != "null" {
		if err := json.Unmarshal(envelope.Result, target); err != nil {
			return response.StatusCode, fmt.Errorf("decode Cloudflare response: %w", err)
		}
	}
	return response.StatusCode, nil
}

func (s *OpsConnectorService) HostingerRead(ctx context.Context, id uint, path string, query url.Values, target interface{}) (int, error) {
	if s == nil || s.repo == nil {
		return 0, errors.New("operations connector service is not configured")
	}
	const virtualMachinesPath = "/api/vps/v1/virtual-machines"
	if path != virtualMachinesPath && !strings.HasPrefix(path, virtualMachinesPath+"/") {
		return 0, fmt.Errorf("%w: Hostinger read path is not allowed", ErrInvalidOpsConnector)
	}
	if strings.Contains(path, "..") {
		return 0, fmt.Errorf("%w: Hostinger read path is not allowed", ErrInvalidOpsConnector)
	}

	record, err := s.repo.FindByID(id)
	if err != nil {
		return 0, err
	}
	if record.Provider != ops.ConnectorProviderHostinger {
		return 0, fmt.Errorf("%w: connector is not a Hostinger connector", ErrInvalidOpsConnector)
	}
	if !record.Enabled {
		return 0, errors.New("Hostinger connector is disabled")
	}

	credentials, err := s.readCredentials(*record)
	if err != nil {
		return 0, err
	}
	if len(credentials) == 0 {
		return 0, errors.New("Hostinger connector credentials are not configured")
	}

	endpoint, err := url.Parse(hostingerAPIBaseURL + path)
	if err != nil {
		return 0, fmt.Errorf("build Hostinger request: %w", err)
	}
	endpoint.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("build Hostinger request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "tanzanite-ops-hostinger-read/1.0")
	applyConnectorAuth(req, record.AuthType, credentials)

	if err := ensureConnectorEndpointSafe(ctx, req.URL, record.Environment); err != nil {
		return 0, err
	}
	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	clientCopy := *client
	clientCopy.CheckRedirect = func(redirectRequest *http.Request, _ []*http.Request) error {
		return ensureConnectorEndpointSafe(redirectRequest.Context(), redirectRequest.URL, record.Environment)
	}

	response, err := clientCopy.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Hostinger request failed: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return response.StatusCode, fmt.Errorf("read Hostinger response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return response.StatusCode, fmt.Errorf("Hostinger API returned HTTP %d", response.StatusCode)
	}
	if target == nil || len(body) == 0 {
		return response.StatusCode, nil
	}

	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if len(body) > 0 && body[0] == '{' && json.Unmarshal(body, &envelope) == nil && len(envelope.Data) > 0 {
		body = envelope.Data
	}
	if err := json.Unmarshal(body, target); err != nil {
		return response.StatusCode, fmt.Errorf("decode Hostinger response: %w", err)
	}
	return response.StatusCode, nil
}

func (s *OpsConnectorService) HostingerUpdateProject(
	ctx context.Context,
	id uint,
	virtualMachineID string,
	projectName string,
	idempotencyKey string,
) (*OpsHostingerUpdateResult, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations connector service is not configured")
	}
	virtualMachineID = strings.TrimSpace(virtualMachineID)
	projectName = strings.TrimSpace(projectName)
	if virtualMachineID == "" || projectName == "" {
		return nil, fmt.Errorf("%w: Hostinger virtual machine and project names are required", ErrInvalidOpsConnector)
	}
	if strings.Contains(virtualMachineID, "..") || strings.Contains(projectName, "..") {
		return nil, fmt.Errorf("%w: Hostinger project path is not allowed", ErrInvalidOpsConnector)
	}

	record, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if record.Provider != ops.ConnectorProviderHostinger {
		return nil, fmt.Errorf("%w: connector is not a Hostinger connector", ErrInvalidOpsConnector)
	}
	if !record.Enabled {
		return nil, errors.New("Hostinger connector is disabled")
	}
	credentials, err := s.readCredentials(*record)
	if err != nil {
		return nil, err
	}
	if len(credentials) == 0 {
		return nil, errors.New("Hostinger connector credentials are not configured")
	}

	path := "/api/vps/v1/virtual-machines/" + url.PathEscape(virtualMachineID) +
		"/docker/" + url.PathEscape(projectName) + "/update"
	endpoint, err := url.Parse(hostingerAPIBaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("build Hostinger update request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(nil))
	if err != nil {
		return nil, fmt.Errorf("build Hostinger update request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "tanzanite-ops-hostinger-update/1.0")
	if key := strings.TrimSpace(idempotencyKey); key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	applyConnectorAuth(req, record.AuthType, credentials)
	if err := ensureConnectorEndpointSafe(ctx, req.URL, record.Environment); err != nil {
		return nil, err
	}

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	clientCopy := *client
	clientCopy.CheckRedirect = func(redirectRequest *http.Request, _ []*http.Request) error {
		return ensureConnectorEndpointSafe(redirectRequest.Context(), redirectRequest.URL, record.Environment)
	}
	response, err := clientCopy.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Hostinger update request failed: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read Hostinger update response: %w", err)
	}
	result := &OpsHostingerUpdateResult{
		StatusCode: response.StatusCode,
		Message:    "Hostinger Docker project update accepted",
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Hostinger update API returned HTTP %d", response.StatusCode)
	}
	var envelope struct {
		Data struct {
			ID          string `json:"id"`
			OperationID string `json:"operation_id"`
			TaskID      string `json:"task_id"`
		} `json:"data"`
		ID          string `json:"id"`
		OperationID string `json:"operation_id"`
		TaskID      string `json:"task_id"`
	}
	if len(body) > 0 && json.Unmarshal(body, &envelope) == nil {
		result.OperationID = firstNonEmptyWorkflowValue(
			envelope.Data.OperationID,
			envelope.Data.TaskID,
			envelope.Data.ID,
			envelope.OperationID,
			envelope.TaskID,
			envelope.ID,
		)
	}
	return result, nil
}

func (s *OpsConnectorService) CloudflareWrite(
	ctx context.Context,
	id uint,
	method string,
	path string,
	body []byte,
	target interface{},
) (int, error) {
	if s == nil || s.repo == nil {
		return 0, errors.New("operations connector service is not configured")
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch && method != http.MethodDelete {
		return 0, fmt.Errorf("%w: Cloudflare write method is not allowed", ErrInvalidOpsConnector)
	}
	if !isAllowedCloudflareWritePath(path) {
		return 0, fmt.Errorf("%w: Cloudflare write path is not allowed", ErrInvalidOpsConnector)
	}
	record, err := s.repo.FindByID(id)
	if err != nil {
		return 0, err
	}
	if record.Provider != ops.ConnectorProviderCloudflare {
		return 0, fmt.Errorf("%w: connector is not a Cloudflare connector", ErrInvalidOpsConnector)
	}
	if !record.Enabled {
		return 0, errors.New("Cloudflare connector is disabled")
	}
	credentials, err := s.readCredentials(*record)
	if err != nil {
		return 0, err
	}
	if len(credentials) == 0 {
		return 0, errors.New("Cloudflare connector credentials are not configured")
	}
	endpoint, err := url.Parse(cloudflareAPIBaseURL + path)
	if err != nil {
		return 0, fmt.Errorf("build Cloudflare write request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("build Cloudflare write request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "tanzanite-ops-cloudflare-write/1.0")
	applyConnectorAuth(req, record.AuthType, credentials)
	if err := ensureConnectorEndpointSafe(ctx, req.URL, record.Environment); err != nil {
		return 0, err
	}
	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	clientCopy := *client
	clientCopy.CheckRedirect = func(redirectRequest *http.Request, _ []*http.Request) error {
		return ensureConnectorEndpointSafe(redirectRequest.Context(), redirectRequest.URL, record.Environment)
	}
	response, err := clientCopy.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Cloudflare write request failed: %w", err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return response.StatusCode, fmt.Errorf("read Cloudflare write response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return response.StatusCode, fmt.Errorf("Cloudflare API returned HTTP %d", response.StatusCode)
	}
	var envelope struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil || !envelope.Success {
		return response.StatusCode, errors.New("Cloudflare API reported an unsuccessful response")
	}
	if target != nil && len(envelope.Result) > 0 && string(envelope.Result) != "null" {
		if err := json.Unmarshal(envelope.Result, target); err != nil {
			return response.StatusCode, fmt.Errorf("decode Cloudflare write response: %w", err)
		}
	}
	return response.StatusCode, nil
}

func (s *OpsConnectorService) saveTestResult(result *ops.ConnectorTestResult, record *ops.Connector, success bool) (*ops.ConnectorTestResult, error) {
	testStatus := ops.ConnectorTestFailed
	connectorStatus := ops.ConnectorStatusError
	if !record.Enabled {
		connectorStatus = ops.ConnectorStatusDisabled
	}
	if success {
		testStatus = ops.ConnectorTestSuccess
		connectorStatus = ops.ConnectorStatusActive
	}
	lastError := ""
	if !success {
		lastError = result.Message
	}
	if err := s.repo.UpdateTestState(record.ID, testStatus, &result.CheckedAt, connectorStatus, lastError); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *OpsConnectorService) normalizeInput(input OpsConnectorInput, existing *ops.Connector) (ops.Connector, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 120 {
		return ops.Connector{}, fmt.Errorf("%w: name is required and must be at most 120 characters", ErrInvalidOpsConnector)
	}

	provider := normalizeConnectorEnum(input.Provider, map[string]struct{}{
		ops.ConnectorProviderCloudflare: {},
		ops.ConnectorProviderHostinger:  {},
		ops.ConnectorProviderGitHub:     {},
		ops.ConnectorProviderGHCR:       {},
		ops.ConnectorProviderOther:      {},
	})
	if provider == "" {
		return ops.Connector{}, fmt.Errorf("%w: provider is invalid", ErrInvalidOpsConnector)
	}

	environment := normalizeConnectorEnum(input.Environment, map[string]struct{}{
		ops.ConnectorEnvironmentProduction: {},
		ops.ConnectorEnvironmentStaging:    {},
		ops.ConnectorEnvironmentTest:       {},
		ops.ConnectorEnvironmentLocal:      {},
	})
	if environment == "" {
		environment = ops.ConnectorEnvironmentProduction
	}

	authType := normalizeConnectorEnum(input.AuthType, map[string]struct{}{
		ops.ConnectorAuthNone:        {},
		ops.ConnectorAuthAPIToken:    {},
		ops.ConnectorAuthAPIKey:      {},
		ops.ConnectorAuthBearer:      {},
		ops.ConnectorAuthBasic:       {},
		ops.ConnectorAuthEnvironment: {},
		ops.ConnectorAuthManual:      {},
	})
	if authType == "" {
		authType = ops.ConnectorAuthAPIToken
	}

	enabled := true
	status := ops.ConnectorStatusPending
	if existing != nil {
		enabled = existing.Enabled
		status = existing.Status
	}
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	if input.Status != "" {
		status = normalizeConnectorEnum(input.Status, map[string]struct{}{
			ops.ConnectorStatusActive:   {},
			ops.ConnectorStatusPending:  {},
			ops.ConnectorStatusDisabled: {},
			ops.ConnectorStatusError:    {},
		})
		if status == "" {
			return ops.Connector{}, fmt.Errorf("%w: status is invalid", ErrInvalidOpsConnector)
		}
	}
	if !enabled {
		status = ops.ConnectorStatusDisabled
	}
	if enabled && status == ops.ConnectorStatusDisabled {
		status = ops.ConnectorStatusPending
	}

	endpoint := strings.TrimSpace(input.Endpoint)
	if err := validateConnectorEndpoint(endpoint, environment); err != nil {
		return ops.Connector{}, err
	}

	record := ops.Connector{
		Name:          name,
		Provider:      provider,
		Environment:   environment,
		Endpoint:      endpoint,
		AuthType:      authType,
		CredentialRef: strings.TrimSpace(input.CredentialRef),
		Scopes:        strings.TrimSpace(input.Scopes),
		Status:        status,
		Enabled:       enabled,
		Notes:         strings.TrimSpace(input.Notes),
	}
	if existing != nil {
		record.ID = existing.ID
		record.CredentialsEncrypted = existing.CredentialsEncrypted
		record.CredentialFields = existing.CredentialFields
		record.LastTestStatus = existing.LastTestStatus
		record.LastTestedAt = existing.LastTestedAt
		record.LastError = existing.LastError
	}

	if len(input.Credentials) > 0 {
		credentials, err := s.mergeCredentials(record, input.Credentials)
		if err != nil {
			return ops.Connector{}, err
		}
		if len(credentials) > 0 {
			masterKey := strings.TrimSpace(os.Getenv(OpsConnectorMasterKeyEnv))
			if masterKey == "" {
				return ops.Connector{}, fmt.Errorf("%w: %s is required to save connector credentials", ErrInvalidOpsConnector, OpsConnectorMasterKeyEnv)
			}
			plaintext, err := json.Marshal(credentials)
			if err != nil {
				return ops.Connector{}, fmt.Errorf("%w: encode connector credentials: %v", ErrInvalidOpsConnector, err)
			}
			encrypted, err := secretbox.EncryptString(string(plaintext), masterKey)
			if err != nil {
				return ops.Connector{}, fmt.Errorf("%w: encrypt connector credentials: %v", ErrInvalidOpsConnector, err)
			}
			record.CredentialsEncrypted = encrypted
			record.CredentialFields = encodeCredentialFields(credentials)
		}
	}
	return record, nil
}

func (s *OpsConnectorService) mergeCredentials(record ops.Connector, incoming map[string]string) (map[string]string, error) {
	credentials := map[string]string{}
	if strings.TrimSpace(record.CredentialsEncrypted) != "" {
		masterKey := strings.TrimSpace(os.Getenv(OpsConnectorMasterKeyEnv))
		if masterKey == "" {
			return nil, fmt.Errorf("%w: %s is required to update connector credentials", ErrInvalidOpsConnector, OpsConnectorMasterKeyEnv)
		}
		plaintext, err := secretbox.DecryptString(record.CredentialsEncrypted, masterKey)
		if err != nil {
			return nil, fmt.Errorf("%w: decrypt existing connector credentials: %v", ErrInvalidOpsConnector, err)
		}
		if err := json.Unmarshal([]byte(plaintext), &credentials); err != nil {
			return nil, fmt.Errorf("%w: decode existing connector credentials: %v", ErrInvalidOpsConnector, err)
		}
	}
	for key, value := range incoming {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		credentials[key] = value
	}
	return credentials, nil
}

func (s *OpsConnectorService) readCredentials(record ops.Connector) (map[string]string, error) {
	if record.AuthType == ops.ConnectorAuthEnvironment {
		ref := strings.TrimSpace(record.CredentialRef)
		if ref == "" {
			return nil, errors.New("credential environment variable is required")
		}
		value := strings.TrimSpace(os.Getenv(ref))
		if value == "" {
			return nil, fmt.Errorf("credential environment variable %s is empty", ref)
		}
		return map[string]string{"token": value}, nil
	}
	if strings.TrimSpace(record.CredentialsEncrypted) == "" {
		return map[string]string{}, nil
	}
	masterKey := strings.TrimSpace(os.Getenv(OpsConnectorMasterKeyEnv))
	if masterKey == "" {
		return nil, fmt.Errorf("%s is required to read connector credentials", OpsConnectorMasterKeyEnv)
	}
	plaintext, err := secretbox.DecryptString(record.CredentialsEncrypted, masterKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt connector credentials: %w", err)
	}
	credentials := map[string]string{}
	if err := json.Unmarshal([]byte(plaintext), &credentials); err != nil {
		return nil, fmt.Errorf("decode connector credentials: %w", err)
	}
	return credentials, nil
}

func connectorView(record ops.Connector) ops.ConnectorView {
	return ops.ConnectorView{
		ID:                   record.ID,
		Name:                 record.Name,
		Provider:             record.Provider,
		Environment:          record.Environment,
		Endpoint:             record.Endpoint,
		AuthType:             record.AuthType,
		CredentialRef:        record.CredentialRef,
		CredentialConfigured: connectorCredentialConfigured(record),
		CredentialFields:     decodeCredentialFields(record.CredentialFields),
		Scopes:               record.Scopes,
		Status:               record.Status,
		Enabled:              record.Enabled,
		LastTestStatus:       record.LastTestStatus,
		LastTestedAt:         record.LastTestedAt,
		LastError:            record.LastError,
		Notes:                record.Notes,
		CreatedAt:            record.CreatedAt,
		UpdatedAt:            record.UpdatedAt,
	}
}

func normalizeConnectorEnum(value string, allowed map[string]struct{}) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if _, ok := allowed[normalized]; !ok {
		return ""
	}
	return normalized
}

func encodeCredentialFields(credentials map[string]string) string {
	fields := make([]string, 0, len(credentials))
	for key, value := range credentials {
		if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
			fields = append(fields, key)
		}
	}
	sort.Strings(fields)
	encoded, _ := json.Marshal(fields)
	return string(encoded)
}

func decodeCredentialFields(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	var fields []string
	if err := json.Unmarshal([]byte(value), &fields); err != nil {
		return []string{}
	}
	sort.Strings(fields)
	return fields
}

func requiresCredentials(authType string) bool {
	return authType != ops.ConnectorAuthNone && authType != ops.ConnectorAuthManual
}

func connectorCredentialConfigured(record ops.Connector) bool {
	if strings.TrimSpace(record.CredentialsEncrypted) != "" {
		return true
	}
	if record.AuthType != ops.ConnectorAuthEnvironment {
		return false
	}
	ref := strings.TrimSpace(record.CredentialRef)
	return ref != "" && strings.TrimSpace(os.Getenv(ref)) != ""
}

func validateConnectorEndpoint(endpoint, environment string) error {
	if endpoint == "" {
		return nil
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("%w: endpoint must be a valid URL", ErrInvalidOpsConnector)
	}
	if parsed.Scheme != "https" && !(environment == ops.ConnectorEnvironmentLocal && parsed.Scheme == "http") {
		return fmt.Errorf("%w: endpoint must use HTTPS outside local environment", ErrInvalidOpsConnector)
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "" || host == "localhost" || host == "localhost.localdomain" {
		return fmt.Errorf("%w: endpoint must not target localhost", ErrInvalidOpsConnector)
	}
	if isUnsafeConnectorIP(net.ParseIP(host)) {
		return fmt.Errorf("%w: endpoint must not target a private or loopback address", ErrInvalidOpsConnector)
	}
	return nil
}

func ensureConnectorEndpointSafe(ctx context.Context, endpoint *url.URL, environment string) error {
	if endpoint == nil {
		return fmt.Errorf("%w: endpoint is required", ErrInvalidOpsConnector)
	}
	if err := validateConnectorEndpoint(endpoint.String(), environment); err != nil {
		return err
	}

	host := endpoint.Hostname()
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("%w: endpoint host could not be resolved", ErrInvalidOpsConnector)
	}
	for _, ip := range ips {
		if isUnsafeConnectorIP(ip) {
			return fmt.Errorf("%w: endpoint resolves to a private or loopback address", ErrInvalidOpsConnector)
		}
	}
	return nil
}

func isUnsafeConnectorIP(ip net.IP) bool {
	return ip != nil &&
		(ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified())
}

func defaultConnectorEndpoint(provider string) string {
	switch provider {
	case ops.ConnectorProviderCloudflare:
		return "https://api.cloudflare.com/client/v4/user/tokens/verify"
	case ops.ConnectorProviderGitHub, ops.ConnectorProviderGHCR:
		return "https://api.github.com/user"
	default:
		return ""
	}
}

func isAllowedCloudflareWritePath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" || !strings.HasPrefix(path, "/zones/") || strings.Contains(path, "..") || strings.Contains(path, "?") {
		return false
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) < 3 || segments[0] != "zones" || segments[1] == "" {
		return false
	}
	switch segments[2] {
	case "dns_records":
		return len(segments) == 3 || len(segments) == 4
	case "purge_cache":
		return len(segments) == 3
	case "settings":
		return len(segments) == 4 && segments[3] != ""
	default:
		return false
	}
}

func applyConnectorAuth(req *http.Request, authType string, credentials map[string]string) {
	switch authType {
	case ops.ConnectorAuthAPIToken, ops.ConnectorAuthBearer, ops.ConnectorAuthEnvironment:
		token := firstCredential(credentials, "token", "api_token", "access_token", "api_key")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case ops.ConnectorAuthAPIKey:
		key := firstCredential(credentials, "api_key", "token", "api_token")
		if key != "" {
			req.Header.Set("X-API-Key", key)
		}
	case ops.ConnectorAuthBasic:
		req.SetBasicAuth(credentials["username"], credentials["password"])
	}
}

func firstCredential(credentials map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(credentials[key]); value != "" {
			return value
		}
	}
	return ""
}
