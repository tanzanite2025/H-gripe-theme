package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/secretbox"
	"commerce-platform/internal/repository"
)

// YanwenGatewayCredentialResolver is the smallest credential dependency needed
// by catalog, waybill, tracking, and Yanwen-only customs services.
type YanwenGatewayCredentialResolver interface {
	resolveYanwenGatewayCredentials(input YanwenAPIConfigInput) (yanwenGatewayCredentials, error)
}

// yanwenGatewayClientProvider lets Yanwen-only preflight services reuse the
// already configured client without knowing which credential resolver owns it.
// The interface is deliberately package-private so no other domain can depend
// on the Yanwen gateway implementation.
type yanwenGatewayClientProvider interface {
	yanwenGatewayClient() *YanwenGatewayClient
}

// YanwenGatewayConfigurationService owns only encrypted Yanwen gateway
// configuration, credential resolution, and the explicit gateway self-check.
// It never reads orders, shipping templates, tracking events, or other carrier
// repositories.
type YanwenGatewayConfigurationService struct {
	configs *repository.YanwenAPIConfigRepository
	gateway *YanwenGatewayClient
}

func NewYanwenGatewayConfigurationService(
	configs *repository.YanwenAPIConfigRepository,
	gateway *YanwenGatewayClient,
) *YanwenGatewayConfigurationService {
	if gateway == nil {
		gateway = NewYanwenGatewayClient()
	}
	return &YanwenGatewayConfigurationService{configs: configs, gateway: gateway}
}

func (s *YanwenGatewayConfigurationService) GetYanwenAPIConfigurationView(environment string) (*shipping.YanwenAPIConfigView, error) {
	environment = normalizeYanwenEnvironment(environment)
	if !isSupportedYanwenEnvironment(environment) {
		return nil, errors.New("Yanwen environment must be fat or production")
	}
	if s == nil || s.configs == nil {
		return nil, errors.New("Yanwen configuration repository is not configured")
	}
	config, err := s.configs.FindYanwenAPIConfigByEnvironment(environment)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return &shipping.YanwenAPIConfigView{
				Environment: environment,
				Endpoint:    yanwenEndpointForEnvironment(environment),
			}, nil
		}
		return nil, err
	}
	return &shipping.YanwenAPIConfigView{
		Environment:        config.Environment,
		Endpoint:           config.Endpoint,
		UserIDConfigured:   config.UserIDEncrypted != "",
		APITokenConfigured: config.APITokenEncrypted != "",
		Enabled:            config.Enabled,
	}, nil
}

func (s *YanwenGatewayConfigurationService) SaveYanwenAPIConfiguration(input YanwenAPIConfigInput) error {
	input.Environment = normalizeYanwenEnvironment(input.Environment)
	if !isSupportedYanwenEnvironment(input.Environment) {
		return errors.New("Yanwen environment must be fat or production")
	}
	input.Endpoint = strings.TrimSpace(input.Endpoint)
	if input.Endpoint == "" {
		input.Endpoint = yanwenEndpointForEnvironment(input.Environment)
	}
	if err := validateYanwenEndpoint(input.Endpoint); err != nil {
		return err
	}
	if !sameYanwenEndpoint(input.Endpoint, yanwenEndpointForEnvironment(input.Environment)) {
		return fmt.Errorf("Yanwen %s environment must use its official gateway endpoint", input.Environment)
	}
	if s == nil || s.configs == nil {
		return errors.New("Yanwen configuration repository is not configured")
	}
	masterKey := strings.TrimSpace(os.Getenv(YanwenAPIMasterKeyEnv))
	if masterKey == "" {
		return errors.New("YANWEN_API_MASTER_KEY is required")
	}
	config, err := s.configs.FindYanwenAPIConfigByEnvironment(input.Environment)
	if err != nil && !repository.IsRecordNotFound(err) {
		return err
	}
	if config == nil {
		config = &shipping.YanwenAPIConfig{Environment: input.Environment}
	}
	config.Endpoint = input.Endpoint
	config.Enabled = input.Enabled
	if value := strings.TrimSpace(input.UserID); value != "" {
		config.UserIDEncrypted, err = secretbox.EncryptString(value, masterKey)
		if err != nil {
			return err
		}
	}
	if value := strings.TrimSpace(input.APIToken); value != "" {
		config.APITokenEncrypted, err = secretbox.EncryptString(value, masterKey)
		if err != nil {
			return err
		}
	}
	return s.configs.SaveYanwenAPIConfig(config)
}

func (s *YanwenGatewayConfigurationService) PingYanwenGateway(ctx context.Context, input YanwenAPIConfigInput) (YanwenPingResult, error) {
	credentials, err := s.resolveYanwenGatewayCredentials(input)
	if err != nil {
		return YanwenPingResult{}, err
	}
	if s == nil || s.gateway == nil {
		return YanwenPingResult{}, errors.New("Yanwen gateway client is not configured")
	}
	latency, err := s.gateway.PingYanwenGateway(ctx, credentials)
	if err != nil {
		return YanwenPingResult{}, err
	}
	return YanwenPingResult{OK: true, Message: "燕文网关签名与 common.country.getlist 自检通过", LatencyMS: latency.Milliseconds()}, nil
}

func (s *YanwenGatewayConfigurationService) resolveYanwenGatewayCredentials(input YanwenAPIConfigInput) (yanwenGatewayCredentials, error) {
	input.Environment = normalizeYanwenEnvironment(input.Environment)
	if !isSupportedYanwenEnvironment(input.Environment) {
		return yanwenGatewayCredentials{}, errors.New("Yanwen environment must be fat or production")
	}
	var stored *shipping.YanwenAPIConfig
	if s != nil && s.configs != nil {
		config, err := s.configs.FindYanwenAPIConfigByEnvironment(input.Environment)
		if err == nil {
			stored = config
		} else if !repository.IsRecordNotFound(err) {
			return yanwenGatewayCredentials{}, err
		}
	}
	credentials := yanwenGatewayCredentials{
		environment: input.Environment,
		endpoint:    strings.TrimSpace(input.Endpoint),
		userID:      strings.TrimSpace(input.UserID),
		apiToken:    strings.TrimSpace(input.APIToken),
	}
	if stored != nil {
		if credentials.endpoint == "" {
			credentials.endpoint = strings.TrimSpace(stored.Endpoint)
		}
		var err error
		if credentials.userID == "" {
			credentials.userID, err = decryptYanwenSecret(stored.UserIDEncrypted)
			if err != nil {
				return yanwenGatewayCredentials{}, fmt.Errorf("decrypt Yanwen user_id: %w", err)
			}
		}
		if credentials.apiToken == "" {
			credentials.apiToken, err = decryptYanwenSecret(stored.APITokenEncrypted)
			if err != nil {
				return yanwenGatewayCredentials{}, fmt.Errorf("decrypt Yanwen apitoken: %w", err)
			}
		}
	}
	if credentials.endpoint == "" {
		credentials.endpoint = yanwenEndpointForEnvironment(input.Environment)
	}
	if err := validateYanwenEndpoint(credentials.endpoint); err != nil {
		return yanwenGatewayCredentials{}, err
	}
	if !sameYanwenEndpoint(credentials.endpoint, yanwenEndpointForEnvironment(input.Environment)) {
		return yanwenGatewayCredentials{}, fmt.Errorf("Yanwen %s environment must use its official gateway endpoint", input.Environment)
	}
	if credentials.userID == "" || credentials.apiToken == "" {
		return yanwenGatewayCredentials{}, errors.New("Yanwen user_id and apitoken are required")
	}
	return credentials, nil
}

func (s *YanwenGatewayConfigurationService) yanwenGatewayClient() *YanwenGatewayClient {
	if s == nil {
		return nil
	}
	return s.gateway
}

// resolveYanwenTrackingAuthorization reads only the production user ID used
// by the separate official tracking endpoint. The tracking client owns the
// request protocol; credential storage remains in this configuration service.
func (s *YanwenGatewayConfigurationService) resolveYanwenTrackingAuthorization() (string, error) {
	if s == nil || s.configs == nil {
		return "", errors.New("Yanwen API configuration repository is not configured")
	}
	config, err := s.configs.FindYanwenAPIConfigByEnvironment("production")
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return "", errors.New("Yanwen production tracking credentials are not configured")
		}
		return "", fmt.Errorf("load Yanwen production tracking credentials: %w", err)
	}
	authorization, err := decryptYanwenSecret(config.UserIDEncrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt Yanwen tracking authorization: %w", err)
	}
	if strings.TrimSpace(authorization) == "" {
		return "", errors.New("Yanwen production tracking authorization is not configured")
	}
	return strings.TrimSpace(authorization), nil
}
