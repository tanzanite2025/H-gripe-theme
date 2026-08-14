package admin

import "commerce-platform/internal/service"

type opsConnectorRequest struct {
	Name          string            `json:"name" binding:"required"`
	Provider      string            `json:"provider" binding:"required"`
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

type opsConnectorStatusRequest struct {
	Enabled bool `json:"enabled"`
}

func (r opsConnectorRequest) toServiceInput() service.OpsConnectorInput {
	return service.OpsConnectorInput{
		Name:          r.Name,
		Provider:      r.Provider,
		Environment:   r.Environment,
		Endpoint:      r.Endpoint,
		AuthType:      r.AuthType,
		CredentialRef: r.CredentialRef,
		Credentials:   r.Credentials,
		Scopes:        r.Scopes,
		Status:        r.Status,
		Enabled:       r.Enabled,
		Notes:         r.Notes,
	}
}
