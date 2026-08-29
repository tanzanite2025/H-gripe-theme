package app

import (
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

func trimOrigin(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}
