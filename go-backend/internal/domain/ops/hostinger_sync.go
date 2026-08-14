package ops

import "time"

type HostingerVPSSyncResult struct {
	VPSID              uint      `json:"vps_id"`
	Name               string    `json:"name"`
	ConnectorID        uint      `json:"connector_id"`
	ConnectorName      string    `json:"connector_name"`
	ProviderResourceID string    `json:"provider_resource_id"`
	Hostname           string    `json:"hostname,omitempty"`
	IPv4               string    `json:"ipv4,omitempty"`
	OperatingSystem    string    `json:"operating_system,omitempty"`
	RemoteState        string    `json:"remote_state,omitempty"`
	ObservedPlan       string    `json:"observed_plan,omitempty"`
	ObservedRegion     string    `json:"observed_region,omitempty"`
	ObservedStatus     string    `json:"observed_status"`
	ObservedSource     string    `json:"observed_source"`
	LastObservedAt     time.Time `json:"last_observed_at"`
	ObservedError      string    `json:"observed_error,omitempty"`
	Message            string    `json:"message"`
}

type HostingerProjectSyncResult struct {
	ProjectID             uint      `json:"project_id"`
	Name                  string    `json:"name"`
	VPSID                 uint      `json:"vps_id"`
	VPSName               string    `json:"vps_name"`
	ConnectorID           uint      `json:"connector_id"`
	ConnectorName         string    `json:"connector_name"`
	ComposeProjectName    string    `json:"compose_project_name"`
	RemoteState           string    `json:"remote_state,omitempty"`
	HealthStatus          string    `json:"health_status"`
	ContainerCount        int       `json:"container_count"`
	RunningContainerCount int       `json:"running_container_count"`
	HealthyContainerCount int       `json:"healthy_container_count"`
	ObservedSource        string    `json:"observed_source"`
	LastCheckedAt         time.Time `json:"last_checked_at"`
	ObservedError         string    `json:"observed_error,omitempty"`
	Message               string    `json:"message"`
}
