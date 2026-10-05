package service

import (
	"commerce-platform/internal/pkg/secretbox"
	"errors"
	"net/url"
	"os"
	"strings"
)

// Shared Yanwen environment, endpoint, and encrypted-secret helpers. The
// configuration service owns credential resolution.

func decryptYanwenSecret(encrypted string) (string, error) {
	if strings.TrimSpace(encrypted) == "" {
		return "", nil
	}
	masterKey := strings.TrimSpace(os.Getenv(YanwenAPIMasterKeyEnv))
	if masterKey == "" {
		return "", errors.New("YANWEN_API_MASTER_KEY is required to read saved credentials")
	}
	return secretbox.DecryptString(encrypted, masterKey)
}

func normalizeYanwenEnvironment(environment string) string {
	environment = strings.ToLower(strings.TrimSpace(environment))
	if environment == "test" {
		return "fat"
	}
	return environment
}

func isSupportedYanwenEnvironment(environment string) bool {
	return environment == "fat" || environment == "production"
}

func yanwenEndpointForEnvironment(environment string) string {
	if normalizeYanwenEnvironment(environment) == "fat" {
		return yanwenFATEndpoint
	}
	return yanwenDefaultEndpoint
}

func validateYanwenEndpoint(endpoint string) error {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Fragment != "" {
		return errors.New("Yanwen endpoint must be a valid HTTPS gateway URL")
	}
	if !strings.EqualFold(parsed.Hostname(), "open.yw56.com.cn") && !strings.EqualFold(parsed.Hostname(), "open-fat.yw56.com.cn") {
		return errors.New("Yanwen endpoint must use an official gateway host")
	}
	if parsed.Port() != "" && parsed.Port() != "443" {
		return errors.New("Yanwen endpoint must use the HTTPS gateway port")
	}
	if parsed.Path != "/api/order" || parsed.RawQuery != "" {
		return errors.New("Yanwen endpoint must target /api/order")
	}
	return nil
}

func sameYanwenEndpoint(left, right string) bool {
	leftURL, leftErr := url.Parse(strings.TrimSpace(left))
	rightURL, rightErr := url.Parse(strings.TrimSpace(right))
	return leftErr == nil && rightErr == nil &&
		strings.EqualFold(leftURL.Scheme, rightURL.Scheme) &&
		strings.EqualFold(leftURL.Host, rightURL.Host) &&
		leftURL.Path == rightURL.Path
}
