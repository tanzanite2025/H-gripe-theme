package service

import (
	"strings"

	"commerce-platform/internal/domain/ops"
)

// The production boundary mirrors deployment/verify-vps-release-boundary.sh and
// compose.prod.yml. Project bindings store resolved network names and resolved
// volume names, while the deployment script keeps Compose logical keys local.
type deploymentComposeBaseline struct {
	Source         string
	Services       []string
	Networks       []string
	VolumeKeys     []string
	GatewayNetwork string
	GatewayAlias   string
}

var productionComposeBaseline = deploymentComposeBaseline{
	Source:         "compose.prod.yml",
	Services:       []string{"db", "redis", "migrate", "edge-config", "api", "storefront", "admin", "web"},
	Networks:       []string{"db", "cache", "app", "shared-edge"},
	VolumeKeys:     []string{"postgres_data", "redis_data", "uploads"},
	GatewayNetwork: "shared-edge",
	GatewayAlias:   "theme-web",
}

func composeBaselineForEnvironment(environment string) (deploymentComposeBaseline, bool) {
	value := strings.TrimSpace(environment)
	if value == "" || value == ops.ProjectEnvironmentProduction {
		return productionComposeBaseline, true
	}
	return deploymentComposeBaseline{}, false
}

func boundaryBaselineUnavailable(key, label, environment string) ops.DeploymentPreflightCheck {
	return warningCheck(
		key,
		"boundary",
		label,
		"当前环境未配置 Compose 边界基线，暂不具备可比对的边界证据。",
		"环境："+firstNonEmptyString(environment, "未设置")+"；生产环境基线来源："+productionComposeBaseline.Source,
	)
}

func composeVolumeCandidates(projectName, volumeKey string) []string {
	prefix := strings.TrimSpace(projectName)
	resolved := strings.ReplaceAll(volumeKey, "_", "-")
	if prefix == "" {
		return []string{volumeKey}
	}
	return []string{volumeKey, prefix + "-" + resolved}
}
