package admin

import "commerce-platform/internal/service"

type opsProjectBindingRequest struct {
	Name                    string              `json:"name" binding:"required"`
	VPSBindingID            uint                `json:"vps_binding_id" binding:"required"`
	ConnectorID             optionalUintRequest `json:"connector_id"`
	ProviderResourceID      string              `json:"provider_resource_id"`
	Environment             string              `json:"environment"`
	ComposeSource           string              `json:"compose_source"`
	ComposeProjectName      string              `json:"compose_project_name"`
	GatewayNetwork          string              `json:"gateway_network"`
	GatewayAlias            string              `json:"gateway_alias"`
	Services                string              `json:"services"`
	Networks                string              `json:"networks"`
	Volumes                 string              `json:"volumes"`
	CurrentImageTag         string              `json:"current_image_tag"`
	CurrentCommitSHA        string              `json:"current_commit_sha"`
	Status                  string              `json:"status"`
	Enabled                 *bool               `json:"enabled"`
	LastDeploymentAt        string              `json:"last_deployment_at"`
	BackupPolicy            string              `json:"backup_policy"`
	RestoreNotes            string              `json:"restore_notes"`
	QuickBuyRateLimitPolicy string              `json:"quick_buy_rate_limit_policy"`
	Notes                   string              `json:"notes"`
}

type opsProjectBindingStatusRequest struct {
	Enabled bool `json:"enabled"`
}

func (r opsProjectBindingRequest) toServiceInput() service.OpsProjectBindingInput {
	return service.OpsProjectBindingInput{
		Name:                    r.Name,
		VPSBindingID:            r.VPSBindingID,
		ConnectorID:             r.ConnectorID.Value,
		ConnectorIDSet:          r.ConnectorID.Set,
		ProviderResourceID:      r.ProviderResourceID,
		Environment:             r.Environment,
		ComposeSource:           r.ComposeSource,
		ComposeProjectName:      r.ComposeProjectName,
		GatewayNetwork:          r.GatewayNetwork,
		GatewayAlias:            r.GatewayAlias,
		Services:                r.Services,
		Networks:                r.Networks,
		Volumes:                 r.Volumes,
		CurrentImageTag:         r.CurrentImageTag,
		CurrentCommitSHA:        r.CurrentCommitSHA,
		Status:                  r.Status,
		Enabled:                 r.Enabled,
		LastDeploymentAt:        r.LastDeploymentAt,
		BackupPolicy:            r.BackupPolicy,
		RestoreNotes:            r.RestoreNotes,
		QuickBuyRateLimitPolicy: r.QuickBuyRateLimitPolicy,
		Notes:                   r.Notes,
	}
}
