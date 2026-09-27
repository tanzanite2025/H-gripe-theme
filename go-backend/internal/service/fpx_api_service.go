package service

import (
	"bytes"
	"context"
	"crypto/md5" // #nosec G501 -- 4PX's documented gateway signature is MD5.
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/secretbox"
	"commerce-platform/internal/repository"
)

const (
	FpxAPIMasterKeyEnv = "FPX_API_MASTER_KEY"
	fpxChannelMethod   = "ds.xms.logistics_product.getlist"
	fpxAPIVersion      = "1.0.0"
	fpxDefaultEndpoint = "https://open.4px.com/router/api/service"
	fpxTestEndpoint    = "https://open-test.4px.com/router/api/service"
	fpxRequestTimeout  = 20 * time.Second
	fpxResponseLimit   = 4 << 20
)

type FpxAPIService struct {
	configs  *repository.FpxAPIConfigRepository
	shipping *repository.ShippingRepository
	gateway  *FpxGatewayClient
	now      func() time.Time
}

func NewFpxAPIService(configs *repository.FpxAPIConfigRepository, shipping *repository.ShippingRepository) *FpxAPIService {
	return &FpxAPIService{
		configs:  configs,
		shipping: shipping,
		gateway:  NewFpxGatewayClient(),
		now:      time.Now,
	}
}

type FpxPingResult struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	LatencyMS int64  `json:"latency_ms"`
}

type FpxChannelSyncSummary struct {
	Scanned          int `json:"scanned"`
	Added            int `json:"added"`
	Updated          int `json:"updated"`
	PreservedEnabled int `json:"preserved_enabled"`
}

type FpxAPIConfigInput struct {
	Environment string `json:"environment"`
	Endpoint    string `json:"endpoint"`
	AppKey      string `json:"app_key"`
	AppSecret   string `json:"app_secret"`
	AccessToken string `json:"access_token"`
	Enabled     bool   `json:"enabled"`
}

type fpxGatewayCredentials struct {
	environment string
	endpoint    string
	appKey      string
	appSecret   string
	accessToken string
}

func (s *FpxAPIService) View(environment string) (*shipping.FpxAPIConfigView, error) {
	environment = normalizeFpxEnvironment(environment)
	if environment != "test" && environment != "production" {
		return nil, errors.New("4PX environment must be test or production")
	}
	if s == nil || s.configs == nil {
		return nil, errors.New("4PX configuration repository is not configured")
	}
	c, err := s.configs.FindByEnvironment(environment)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return &shipping.FpxAPIConfigView{
				Environment: environment,
				Endpoint:    fpxEndpointForEnvironment(environment),
			}, nil
		}
		return nil, err
	}
	return &shipping.FpxAPIConfigView{
		Environment:              c.Environment,
		Endpoint:                 c.Endpoint,
		AppKeyConfigured:         c.AppKeyEncrypted != "",
		AppSecretConfigured:      c.AppSecretEncrypted != "",
		AccessTokenConfigured:    c.AccessTokenEncrypted != "",
		Enabled:                  c.Enabled,
		LastSyncStatus:           c.LastSyncStatus,
		LastSyncedAt:             c.LastSyncedAt,
		LastError:                c.LastError,
		LastSyncScanned:          c.LastSyncScanned,
		LastSyncAdded:            c.LastSyncAdded,
		LastSyncUpdated:          c.LastSyncUpdated,
		LastSyncPreservedEnabled: c.LastSyncPreservedEnabled,
	}, nil
}

func (s *FpxAPIService) Save(input FpxAPIConfigInput) error {
	input.Environment = normalizeFpxEnvironment(input.Environment)
	if input.Environment != "test" && input.Environment != "production" {
		return errors.New("4PX environment must be test or production")
	}
	input.Endpoint = strings.TrimSpace(input.Endpoint)
	if input.Endpoint == "" {
		input.Endpoint = fpxEndpointForEnvironment(input.Environment)
	}
	if err := validateFpxEndpoint(input.Endpoint); err != nil {
		return err
	}
	if !sameFpxEndpoint(input.Endpoint, fpxEndpointForEnvironment(input.Environment)) {
		return fmt.Errorf("4PX %s environment must use its official gateway endpoint", input.Environment)
	}

	masterKey := strings.TrimSpace(os.Getenv(FpxAPIMasterKeyEnv))
	if masterKey == "" {
		return errors.New("FPX_API_MASTER_KEY is required")
	}
	c, err := s.configs.FindByEnvironment(input.Environment)
	if err != nil && !repository.IsRecordNotFound(err) {
		return err
	}
	if c == nil {
		c = &shipping.FpxAPIConfig{Environment: input.Environment}
	}
	c.Endpoint = input.Endpoint
	c.Enabled = input.Enabled

	if value := strings.TrimSpace(input.AppKey); value != "" {
		c.AppKeyEncrypted, err = secretbox.EncryptString(value, masterKey)
		if err != nil {
			return err
		}
	}
	if value := strings.TrimSpace(input.AppSecret); value != "" {
		c.AppSecretEncrypted, err = secretbox.EncryptString(value, masterKey)
		if err != nil {
			return err
		}
	}
	if value := strings.TrimSpace(input.AccessToken); value != "" {
		c.AccessTokenEncrypted, err = secretbox.EncryptString(value, masterKey)
		if err != nil {
			return err
		}
	}
	return s.configs.Save(c)
}

// Ping performs the same signed, read-only call used by channel sync. This
// makes a successful ping meaningful: credentials, signature, endpoint and
// gateway status have all been accepted by 4PX.
func (s *FpxAPIService) Ping(ctx context.Context, input FpxAPIConfigInput) (FpxPingResult, error) {
	credentials, err := s.resolveCredentials(input)
	if err != nil {
		return FpxPingResult{}, err
	}
	latency, err := s.gateway.Ping(ctx, credentials)
	if err != nil {
		return FpxPingResult{}, err
	}
	return FpxPingResult{
		OK:        true,
		Message:   "4PX gateway signature and status check passed",
		LatencyMS: latency.Milliseconds(),
	}, nil
}

func (s *FpxAPIService) SyncChannels(ctx context.Context, input FpxAPIConfigInput) (FpxChannelSyncSummary, error) {
	input.Environment = normalizeFpxEnvironment(input.Environment)
	credentials, err := s.resolveCredentials(input)
	if err != nil {
		return FpxChannelSyncSummary{}, err
	}
	result, err := s.gateway.FetchChannels(ctx, credentials)
	if err != nil {
		if s.configs != nil {
			_ = s.configs.RecordSyncFailure(credentials.environment, safeFpxError(err, credentials))
		}
		return FpxChannelSyncSummary{}, err
	}
	if s.shipping == nil {
		return FpxChannelSyncSummary{}, errors.New("4PX channel repository is not configured")
	}
	stats, err := s.shipping.UpsertFpxChannels(result.Channels)
	if err != nil {
		if s.configs != nil {
			_ = s.configs.RecordSyncFailure(credentials.environment, safeFpxError(err, credentials))
		}
		return FpxChannelSyncSummary{}, fmt.Errorf("save 4PX channels: %w", err)
	}
	// Report the number of raw official entries returned by 4PX, including
	// entries skipped because they do not contain a usable service code/name.
	stats.Scanned = result.Scanned
	if s.configs != nil {
		if err := s.configs.RecordSyncSuccess(credentials.environment, s.currentTime().UTC(), stats); err != nil {
			return FpxChannelSyncSummary{}, fmt.Errorf("record 4PX sync status: %w", err)
		}
	}
	return FpxChannelSyncSummary{
		Scanned:          stats.Scanned,
		Added:            stats.Added,
		Updated:          stats.Updated,
		PreservedEnabled: stats.PreservedEnabled,
	}, nil
}

func (s *FpxAPIService) resolveCredentials(input FpxAPIConfigInput) (fpxGatewayCredentials, error) {
	input.Environment = normalizeFpxEnvironment(input.Environment)
	if input.Environment != "test" && input.Environment != "production" {
		return fpxGatewayCredentials{}, errors.New("4PX environment must be test or production")
	}

	var stored *shipping.FpxAPIConfig
	if s.configs != nil {
		config, err := s.configs.FindByEnvironment(input.Environment)
		if err == nil {
			stored = config
		} else if !repository.IsRecordNotFound(err) {
			return fpxGatewayCredentials{}, err
		}
	}

	credentials := fpxGatewayCredentials{
		environment: input.Environment,
		endpoint:    strings.TrimSpace(input.Endpoint),
		appKey:      strings.TrimSpace(input.AppKey),
		appSecret:   strings.TrimSpace(input.AppSecret),
		accessToken: strings.TrimSpace(input.AccessToken),
	}
	if stored != nil {
		if credentials.endpoint == "" {
			credentials.endpoint = strings.TrimSpace(stored.Endpoint)
		}
		var err error
		if credentials.appKey == "" {
			credentials.appKey, err = decryptFpxSecret(stored.AppKeyEncrypted)
			if err != nil {
				return fpxGatewayCredentials{}, fmt.Errorf("decrypt 4PX AppKey: %w", err)
			}
		}
		if credentials.appSecret == "" {
			credentials.appSecret, err = decryptFpxSecret(stored.AppSecretEncrypted)
			if err != nil {
				return fpxGatewayCredentials{}, fmt.Errorf("decrypt 4PX AppSecret: %w", err)
			}
		}
		if credentials.accessToken == "" {
			credentials.accessToken, err = decryptFpxSecret(stored.AccessTokenEncrypted)
			if err != nil {
				return fpxGatewayCredentials{}, fmt.Errorf("decrypt 4PX access token: %w", err)
			}
		}
	}
	if credentials.endpoint == "" {
		credentials.endpoint = fpxEndpointForEnvironment(input.Environment)
	}
	if err := validateFpxEndpoint(credentials.endpoint); err != nil {
		return fpxGatewayCredentials{}, err
	}
	if !sameFpxEndpoint(credentials.endpoint, fpxEndpointForEnvironment(input.Environment)) {
		return fpxGatewayCredentials{}, fmt.Errorf("4PX %s environment must use its official gateway endpoint", input.Environment)
	}
	if credentials.appKey == "" || credentials.appSecret == "" {
		return fpxGatewayCredentials{}, errors.New("4PX AppKey and AppSecret are required")
	}
	return credentials, nil
}

func decryptFpxSecret(encrypted string) (string, error) {
	if strings.TrimSpace(encrypted) == "" {
		return "", nil
	}
	masterKey := strings.TrimSpace(os.Getenv(FpxAPIMasterKeyEnv))
	if masterKey == "" {
		return "", errors.New("FPX_API_MASTER_KEY is required to read saved credentials")
	}
	return secretbox.DecryptString(encrypted, masterKey)
}

func signFpxRequest(publicParams map[string]string, body []byte, appSecret string) string {
	keys := make([]string, 0, len(publicParams))
	for key := range publicParams {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteString(publicParams[key])
	}
	builder.Write(body)
	builder.WriteString(appSecret)
	digest := md5.Sum([]byte(builder.String()))
	return hex.EncodeToString(digest[:])
}

type fpxChannelResponse struct {
	Result    json.RawMessage `json:"result"`
	Errors    json.RawMessage `json:"errors"`
	Data      json.RawMessage `json:"data"`
	Message   string          `json:"message"`
	Msg       string          `json:"msg"`
	Error     string          `json:"error"`
	ErrorCode string          `json:"error_code"`
	ErrorMsg  string          `json:"error_msg"`
}

func parseFpxChannelResponse(responseBody []byte) ([]shipping.FpxChannel, error) {
	channels, _, err := parseFpxChannelResponseWithCount(responseBody)
	return channels, err
}

func parseFpxGatewayResponse(responseBody []byte) (fpxChannelResponse, error) {
	var response fpxChannelResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return fpxChannelResponse{}, fmt.Errorf("decode 4PX response: %w", err)
	}
	if fpxRawHasError(response.Errors) {
		return fpxChannelResponse{}, fmt.Errorf("4PX API error: %s", fpxResponseError(response))
	}
	if len(response.Result) == 0 || !fpxResultOK(response.Result) {
		message := fpxResponseError(response)
		if message == "" {
			message = fpxResponseMessage(response)
		}
		if message == "" {
			message = "4PX API returned an unsuccessful result"
		}
		return fpxChannelResponse{}, errors.New(message)
	}
	return response, nil
}

func parseFpxChannelResponseWithCount(responseBody []byte) ([]shipping.FpxChannel, int, error) {
	response, err := parseFpxGatewayResponse(responseBody)
	if err != nil {
		return nil, 0, err
	}

	records, scanned := fpxChannelRecordsWithCount(response.Data)
	channels := make([]shipping.FpxChannel, 0, len(records))
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		code := firstFpxString(record, "logistics_product_code", "service_code", "logisticsProductCode", "code")
		name := firstFpxString(record, "logistics_product_name_cn", "service_name_cn", "display_name", "logistics_product_name", "logistics_product_name_en", "service_name")
		if code == "" || name == "" {
			continue
		}
		key := strings.ToLower(code)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		channels = append(channels, shipping.FpxChannel{ServiceCode: code, DisplayName: name})
	}
	if len(response.Data) == 0 || string(response.Data) == "null" {
		return nil, 0, errors.New("4PX response is missing data")
	}
	return channels, scanned, nil
}

func fpxChannelRecordsWithCount(raw json.RawMessage) ([]map[string]json.RawMessage, int) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, 0
	}
	if raw[0] == '[' {
		var entries []json.RawMessage
		if json.Unmarshal(raw, &entries) != nil {
			return nil, 0
		}
		records := make([]map[string]json.RawMessage, 0, len(entries))
		for _, entry := range entries {
			var record map[string]json.RawMessage
			if json.Unmarshal(entry, &record) == nil && record != nil {
				records = append(records, record)
			}
		}
		return records, len(entries)
	}
	if raw[0] != '{' {
		return nil, 0
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return nil, 0
	}
	if fpxRecordLooksLikeChannel(object) {
		return []map[string]json.RawMessage{object}, 1
	}
	for _, key := range []string{"data", "list", "records", "items", "result", "logistics_product_list", "logistics_products"} {
		if nested, ok := object[key]; ok {
			if records, scanned := fpxChannelRecordsWithCount(nested); len(records) > 0 || scanned > 0 {
				return records, scanned
			}
		}
	}
	return nil, 0
}

func fpxRecordLooksLikeChannel(record map[string]json.RawMessage) bool {
	for _, key := range []string{"logistics_product_code", "service_code", "logisticsProductCode", "code"} {
		if _, ok := record[key]; ok {
			return true
		}
	}
	return false
}

func firstFpxString(record map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		raw, ok := record[key]
		if !ok {
			continue
		}
		var value string
		if json.Unmarshal(raw, &value) == nil && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func fpxRawHasError(raw json.RawMessage) bool {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return false
	}
	if bytes.Equal(raw, []byte("false")) || bytes.Equal(raw, []byte("0")) || bytes.Equal(raw, []byte(`"0"`)) ||
		bytes.Equal(raw, []byte("[]")) || bytes.Equal(raw, []byte("{}")) || bytes.Equal(raw, []byte(`""`)) {
		return false
	}
	return true
}

func firstFpxError(raw json.RawMessage) string {
	if !fpxRawHasError(raw) {
		return ""
	}
	var message string
	if json.Unmarshal(raw, &message) == nil && strings.TrimSpace(message) != "" {
		return strings.TrimSpace(message)
	}
	var messages []map[string]json.RawMessage
	if json.Unmarshal(raw, &messages) == nil && len(messages) > 0 {
		for _, item := range messages {
			code := firstFpxString(item, "error_code", "errorCode", "code", "error_code")
			message := firstFpxString(item, "error_msg", "errorMsg", "message", "msg", "error")
			if result := formatFpxGatewayError(code, message); result != "" {
				return result
			}
		}
	}
	var errorObject map[string]json.RawMessage
	if json.Unmarshal(raw, &errorObject) == nil && len(errorObject) > 0 {
		code := firstFpxString(errorObject, "error_code", "errorCode", "code")
		message := firstFpxString(errorObject, "error_msg", "errorMsg", "message", "msg", "error")
		if result := formatFpxGatewayError(code, message); result != "" {
			return result
		}
	}
	var simpleMessages []string
	if json.Unmarshal(raw, &simpleMessages) == nil && len(simpleMessages) > 0 && strings.TrimSpace(simpleMessages[0]) != "" {
		return strings.TrimSpace(simpleMessages[0])
	}
	return summarizeFpxResponse(raw)
}

func fpxResponseMessage(response fpxChannelResponse) string {
	for _, message := range []string{response.Message, response.Msg, response.ErrorMsg, response.Error} {
		if strings.TrimSpace(message) != "" {
			return strings.TrimSpace(message)
		}
	}
	return ""
}

func fpxResponseError(response fpxChannelResponse) string {
	if message := firstFpxError(response.Errors); message != "" {
		return message
	}
	if message := formatFpxGatewayError(response.ErrorCode, response.ErrorMsg); message != "" {
		return message
	}
	return strings.TrimSpace(response.Error)
}

func formatFpxGatewayError(code string, message string) string {
	code = strings.TrimSpace(code)
	message = strings.TrimSpace(message)
	switch {
	case code != "" && message != "":
		return code + ": " + message
	case code != "":
		return code
	default:
		return message
	}
}

func fpxResultOK(raw json.RawMessage) bool {
	var boolean bool
	if json.Unmarshal(raw, &boolean) == nil {
		return boolean
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return strings.EqualFold(strings.TrimSpace(text), "true") || strings.TrimSpace(text) == "1"
	}
	var number int
	if json.Unmarshal(raw, &number) == nil {
		return number == 1
	}
	return false
}

func summarizeFpxResponse(body []byte) string {
	message := strings.TrimSpace(string(body))
	if len(message) > 500 {
		message = message[:500]
	}
	if message == "" {
		return "empty response"
	}
	return message
}

func safeFpxError(err error, credentials fpxGatewayCredentials) string {
	if err == nil {
		return ""
	}
	return summarizeFpxResponse([]byte(redactFpxCredentials(err.Error(), credentials)))
}

func redactFpxCredentials(message string, credentials fpxGatewayCredentials) string {
	for _, secret := range []string{credentials.appKey, credentials.appSecret, credentials.accessToken} {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[REDACTED]")
		}
	}
	return message
}

func normalizeFpxEnvironment(environment string) string {
	return strings.ToLower(strings.TrimSpace(environment))
}

func fpxEndpointForEnvironment(environment string) string {
	if normalizeFpxEnvironment(environment) == "test" {
		return fpxTestEndpoint
	}
	return fpxDefaultEndpoint
}

func validateFpxEndpoint(endpoint string) error {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Fragment != "" {
		return errors.New("4PX endpoint must be a valid HTTPS gateway URL")
	}
	if !strings.EqualFold(parsed.Hostname(), "open.4px.com") && !strings.EqualFold(parsed.Hostname(), "open-test.4px.com") {
		return errors.New("4PX endpoint must use an official 4PX gateway host")
	}
	if parsed.Port() != "" && parsed.Port() != "443" {
		return errors.New("4PX endpoint must use the HTTPS gateway port")
	}
	if parsed.Path != "/router/api/service" || parsed.RawQuery != "" {
		return errors.New("4PX endpoint must target /router/api/service")
	}
	return nil
}

func sameFpxEndpoint(left string, right string) bool {
	leftURL, leftErr := url.Parse(strings.TrimSpace(left))
	rightURL, rightErr := url.Parse(strings.TrimSpace(right))
	return leftErr == nil && rightErr == nil &&
		strings.EqualFold(leftURL.Scheme, rightURL.Scheme) &&
		strings.EqualFold(leftURL.Host, rightURL.Host) &&
		leftURL.Path == rightURL.Path
}

func (s *FpxAPIService) currentTime() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}
