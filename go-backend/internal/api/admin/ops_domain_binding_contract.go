package admin

import (
	"bytes"
	"encoding/json"

	"commerce-platform/internal/service"
)

type optionalUintRequest struct {
	Set   bool
	Value *uint
}

func (r *optionalUintRequest) UnmarshalJSON(data []byte) error {
	r.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		r.Value = nil
		return nil
	}

	var value uint
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value == 0 {
		r.Value = nil
		return nil
	}
	r.Value = &value
	return nil
}

type opsDomainBindingRequest struct {
	Domain           string              `json:"domain" binding:"required"`
	ConnectorID      optionalUintRequest `json:"connector_id"`
	ProjectBindingID optionalUintRequest `json:"project_binding_id"`
	Role             string              `json:"role" binding:"required"`
	Environment      string              `json:"environment" binding:"required"`
	Provider         string              `json:"provider" binding:"required"`
	Zone             string              `json:"zone"`
	Target           string              `json:"target"`
	ProxyMode        string              `json:"proxy_mode"`
	TLSMode          string              `json:"tls_mode"`
	RedirectTarget   string              `json:"redirect_target"`
	Status           string              `json:"status"`
	Enabled          *bool               `json:"enabled"`
	Notes            string              `json:"notes"`
}

type opsDomainBindingStatusRequest struct {
	Enabled bool `json:"enabled"`
}

func (r opsDomainBindingRequest) toServiceInput() service.OpsDomainBindingInput {
	return service.OpsDomainBindingInput{
		Domain:              r.Domain,
		ConnectorID:         r.ConnectorID.Value,
		ConnectorIDSet:      r.ConnectorID.Set,
		ProjectBindingID:    r.ProjectBindingID.Value,
		ProjectBindingIDSet: r.ProjectBindingID.Set,
		Role:                r.Role,
		Environment:         r.Environment,
		Provider:            r.Provider,
		Zone:                r.Zone,
		Target:              r.Target,
		ProxyMode:           r.ProxyMode,
		TLSMode:             r.TLSMode,
		RedirectTarget:      r.RedirectTarget,
		Status:              r.Status,
		Enabled:             r.Enabled,
		Notes:               r.Notes,
	}
}
