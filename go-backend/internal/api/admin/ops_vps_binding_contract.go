package admin

import "commerce-platform/internal/service"

type opsVPSBindingRequest struct {
	Name               string              `json:"name" binding:"required"`
	Provider           string              `json:"provider" binding:"required"`
	Environment        string              `json:"environment"`
	ConnectorID        optionalUintRequest `json:"connector_id"`
	ProviderResourceID string              `json:"provider_resource_id"`
	Hostname           string              `json:"hostname"`
	IPv4               string              `json:"ipv4"`
	Region             string              `json:"region"`
	OperatingSystem    string              `json:"operating_system"`
	Status             string              `json:"status"`
	Enabled            *bool               `json:"enabled"`
	Notes              string              `json:"notes"`
}

type opsVPSBindingStatusRequest struct {
	Enabled bool `json:"enabled"`
}

func (r opsVPSBindingRequest) toServiceInput() service.OpsVPSBindingInput {
	return service.OpsVPSBindingInput{
		Name:               r.Name,
		Provider:           r.Provider,
		Environment:        r.Environment,
		ConnectorID:        r.ConnectorID.Value,
		ConnectorIDSet:     r.ConnectorID.Set,
		ProviderResourceID: r.ProviderResourceID,
		Hostname:           r.Hostname,
		IPv4:               r.IPv4,
		Region:             r.Region,
		OperatingSystem:    r.OperatingSystem,
		Status:             r.Status,
		Enabled:            r.Enabled,
		Notes:              r.Notes,
	}
}
