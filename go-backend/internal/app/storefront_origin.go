package app

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"commerce-platform/internal/pkg/config"
)

func resolveStorefrontOrigins(cfg *config.Config) (string, string) {
	publicOrigin := trimOrigin(os.Getenv("STOREFRONT_BASE_URL"))
	internalOrigin := trimOrigin(os.Getenv("STOREFRONT_INTERNAL_ORIGIN"))
	releaseMode := cfg != nil && strings.EqualFold(strings.TrimSpace(cfg.Server.Mode), "release")

	if publicOrigin == "" {
		if releaseMode {
			if cfg != nil {
				publicOrigin = trimOrigin(cfg.Server.BaseURL)
			}
		} else {
			publicOrigin = defaultDevStorefrontOrigin
		}
	}

	if internalOrigin == "" && !releaseMode {
		internalOrigin = publicOrigin
	}

	return publicOrigin, internalOrigin
}

func resolveSiteQualityTargetOrigin(publicOrigin string) string {
	targetOrigin := trimOrigin(os.Getenv("SITE_QUALITY_TARGET_ORIGIN"))
	if targetOrigin == "" {
		return publicOrigin
	}
	return targetOrigin
}

func validateStorefrontInternalOrigin(cfg *config.Config, internalOrigin string) error {
	trimmedOrigin := trimOrigin(internalOrigin)
	if trimmedOrigin == "" {
		if cfg != nil && strings.EqualFold(strings.TrimSpace(cfg.Server.Mode), "release") {
			return fmt.Errorf("STOREFRONT_INTERNAL_ORIGIN is required in release mode")
		}
		return nil
	}

	parsedOrigin, err := url.Parse(trimmedOrigin)
	if err != nil || parsedOrigin.Scheme == "" || parsedOrigin.Host == "" ||
		!strings.EqualFold(parsedOrigin.Scheme, "http") &&
			!strings.EqualFold(parsedOrigin.Scheme, "https") {
		return fmt.Errorf(
			"STOREFRONT_INTERNAL_ORIGIN %q must be an absolute http or https origin",
			internalOrigin,
		)
	}

	if cfg == nil {
		return nil
	}

	apiOrigin := trimOrigin(cfg.Server.BaseURL)
	if apiOrigin != "" && sameURLAuthority(apiOrigin, internalOrigin) {
		return fmt.Errorf(
			"STOREFRONT_INTERNAL_ORIGIN %q points to the API origin; configure the storefront origin instead",
			internalOrigin,
		)
	}

	if apiPort := strings.TrimPrefix(strings.TrimSpace(cfg.Server.Port), ":"); apiPort != "" &&
		parsedOrigin.Port() == apiPort {
		return fmt.Errorf(
			"STOREFRONT_INTERNAL_ORIGIN %q uses the API port %s; configure the storefront origin instead",
			internalOrigin,
			apiPort,
		)
	}

	return nil
}

func sameURLAuthority(left string, right string) bool {
	leftURL, leftErr := url.Parse(trimOrigin(left))
	rightURL, rightErr := url.Parse(trimOrigin(right))
	if leftErr != nil || rightErr != nil {
		return strings.EqualFold(trimOrigin(left), trimOrigin(right))
	}
	return strings.EqualFold(leftURL.Scheme, rightURL.Scheme) &&
		strings.EqualFold(leftURL.Host, rightURL.Host)
}

func trimOrigin(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}
