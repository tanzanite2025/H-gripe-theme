package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

var ErrOpsHostingerSync = errors.New("operations Hostinger sync failed")

type OpsHostingerSyncService struct {
	vpsRepo          *repository.OpsVPSBindingRepository
	projectRepo      *repository.OpsProjectBindingRepository
	connectorService *OpsConnectorService
}

type hostingerVirtualMachine struct {
	ID           uint                 `json:"id"`
	Hostname     string               `json:"hostname"`
	State        string               `json:"state"`
	DataCenterID *uint                `json:"data_center_id"`
	Plan         hostingerNamedValue  `json:"plan"`
	IPv4         []hostingerIPAddress `json:"ipv4"`
	Template     *hostingerTemplate   `json:"template"`
}

type hostingerNamedValue struct {
	Value string
}

func (value *hostingerNamedValue) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '"' {
		return json.Unmarshal(data, &value.Value)
	}
	var object struct {
		Name  string `json:"name"`
		Title string `json:"title"`
		Slug  string `json:"slug"`
		ID    string `json:"id"`
	}
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	value.Value = hostingerFirstNonEmpty(object.Name, object.Title, object.Slug, object.ID)
	return nil
}

type hostingerIPAddress struct {
	Address string `json:"address"`
}

func (address *hostingerIPAddress) UnmarshalJSON(data []byte) error {
	var object struct {
		Address string `json:"address"`
		IPv4    string `json:"ipv4"`
		IP      string `json:"ip"`
	}
	if len(data) > 0 && data[0] == '"' {
		return json.Unmarshal(data, &address.Address)
	}
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	address.Address = hostingerFirstNonEmpty(object.Address, object.IPv4, object.IP)
	return nil
}

type hostingerTemplate struct {
	Name string `json:"name"`
}

type hostingerDockerProject struct {
	Name       string                     `json:"name"`
	Status     string                     `json:"status"`
	State      string                     `json:"state"`
	Path       string                     `json:"path"`
	Containers []hostingerDockerContainer `json:"containers"`
}

type hostingerDockerContainer struct {
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
	State  string `json:"state"`
	Health string `json:"health"`
}

func NewOpsHostingerSyncService(
	vpsRepo *repository.OpsVPSBindingRepository,
	projectRepo *repository.OpsProjectBindingRepository,
	connectorService *OpsConnectorService,
) *OpsHostingerSyncService {
	return &OpsHostingerSyncService{
		vpsRepo:          vpsRepo,
		projectRepo:      projectRepo,
		connectorService: connectorService,
	}
}

func (s *OpsHostingerSyncService) SyncVPS(ctx context.Context, id uint) (*ops.HostingerVPSSyncResult, error) {
	if s == nil || s.vpsRepo == nil || s.connectorService == nil {
		return nil, errors.New("operations Hostinger sync service is not configured")
	}

	vps, err := s.vpsRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	checkedAt := time.Now().UTC()
	result := &ops.HostingerVPSSyncResult{
		VPSID:              vps.ID,
		Name:               vps.Name,
		ProviderResourceID: strings.TrimSpace(vps.ProviderResourceID),
		Hostname:           strings.TrimSpace(vps.Hostname),
		IPv4:               strings.TrimSpace(vps.IPv4),
		OperatingSystem:    strings.TrimSpace(vps.OperatingSystem),
		ObservedStatus:     ops.VPSObservedUnknown,
		ObservedSource:     "hostinger",
		LastObservedAt:     checkedAt,
	}

	if vps.Provider != ops.VPSProviderHostinger {
		return s.persistVPSFailure(result, "VPS 绑定的提供商不是 Hostinger")
	}
	if vps.ConnectorID == nil {
		return s.persistVPSFailure(result, "VPS 未绑定 Hostinger 连接器")
	}
	result.ConnectorID = *vps.ConnectorID
	connector, err := s.connectorService.Get(*vps.ConnectorID)
	if err != nil {
		return s.persistVPSFailure(result, "读取 Hostinger 连接器失败")
	}
	result.ConnectorName = connector.Name
	result.ObservedSource = "hostinger:" + connector.Name
	if connector.Provider != ops.ConnectorProviderHostinger {
		return s.persistVPSFailure(result, "VPS 绑定的连接器不是 Hostinger")
	}
	if connector.Environment != "" && connector.Environment != vps.Environment {
		return s.persistVPSFailure(result, "VPS 绑定的连接器环境与 VPS 环境不一致")
	}

	resourceID := strings.TrimSpace(vps.ProviderResourceID)
	if resourceID == "" {
		return s.persistVPSFailure(result, "VPS 未登记 Hostinger 资源 ID")
	}
	var remote hostingerVirtualMachine
	_, err = s.connectorService.HostingerRead(
		ctx,
		*vps.ConnectorID,
		"/api/vps/v1/virtual-machines/"+url.PathEscape(resourceID),
		nil,
		&remote,
	)
	if err != nil {
		return s.persistVPSFailure(result, "读取 Hostinger VPS 详情失败")
	}

	result.Hostname = hostingerFirstNonEmpty(remote.Hostname, vps.Hostname)
	result.IPv4 = hostingerFirstNonEmpty(firstHostingerIPv4(remote.IPv4), vps.IPv4)
	if remote.Template != nil {
		result.OperatingSystem = hostingerFirstNonEmpty(remote.Template.Name, vps.OperatingSystem)
	}
	result.RemoteState = strings.TrimSpace(remote.State)
	result.ObservedPlan = strings.TrimSpace(remote.Plan.Value)
	if remote.DataCenterID != nil {
		result.ObservedRegion = fmt.Sprintf("data_center:%d", *remote.DataCenterID)
	}
	result.ObservedStatus = hostingerVPSObservedStatus(remote.State)
	result.Message = "Hostinger VPS 观察状态已更新"
	if result.ObservedStatus == ops.VPSObservedHealthy {
		result.Message = "Hostinger VPS 正常运行"
	} else if result.ObservedStatus == ops.VPSObservedOffline {
		result.Message = "Hostinger VPS 当前不在线"
	} else if result.ObservedStatus == ops.VPSObservedDegraded {
		result.Message = "Hostinger VPS 当前处于非稳定状态"
	}
	if err := s.vpsRepo.UpdateObservedState(
		result.VPSID,
		result.ObservedStatus,
		result.RemoteState,
		result.ObservedSource,
		result.Hostname,
		result.IPv4,
		result.OperatingSystem,
		result.ObservedPlan,
		result.ObservedRegion,
		result.LastObservedAt,
		"",
	); err != nil {
		return result, err
	}
	return result, nil
}

func (s *OpsHostingerSyncService) SyncProject(ctx context.Context, id uint) (*ops.HostingerProjectSyncResult, error) {
	if s == nil || s.vpsRepo == nil || s.projectRepo == nil || s.connectorService == nil {
		return nil, errors.New("operations Hostinger sync service is not configured")
	}

	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	vps, err := s.vpsRepo.FindByID(project.VPSBindingID)
	if err != nil {
		return nil, err
	}
	checkedAt := time.Now().UTC()
	result := &ops.HostingerProjectSyncResult{
		ProjectID:          project.ID,
		Name:               project.Name,
		VPSID:              project.VPSBindingID,
		VPSName:            project.VPSName,
		ComposeProjectName: strings.TrimSpace(project.ComposeProjectName),
		HealthStatus:       ops.ProjectHealthUnknown,
		ObservedSource:     "hostinger",
		LastCheckedAt:      checkedAt,
	}

	if vps.Provider != ops.VPSProviderHostinger {
		return s.persistProjectFailure(result, "项目绑定的 VPS 提供商不是 Hostinger")
	}
	connectorID := project.ConnectorID
	if connectorID == nil {
		connectorID = vps.ConnectorID
	}
	if connectorID == nil {
		return s.persistProjectFailure(result, "项目未绑定 Hostinger 连接器")
	}
	result.ConnectorID = *connectorID
	connector, err := s.connectorService.Get(*connectorID)
	if err != nil {
		return s.persistProjectFailure(result, "读取 Hostinger 连接器失败")
	}
	result.ConnectorName = connector.Name
	result.ObservedSource = "hostinger:" + connector.Name
	if connector.Provider != ops.ConnectorProviderHostinger {
		return s.persistProjectFailure(result, "项目绑定的连接器不是 Hostinger")
	}
	if connector.Environment != "" && connector.Environment != project.Environment {
		return s.persistProjectFailure(result, "项目绑定的连接器环境与项目环境不一致")
	}
	if strings.TrimSpace(vps.Environment) != "" && vps.Environment != project.Environment {
		return s.persistProjectFailure(result, "项目环境与绑定 VPS 环境不一致")
	}
	if strings.TrimSpace(vps.ProviderResourceID) == "" {
		return s.persistProjectFailure(result, "项目绑定的 VPS 未登记 Hostinger 资源 ID")
	}
	if result.ComposeProjectName == "" {
		return s.persistProjectFailure(result, "项目未登记 Compose 项目名")
	}

	vmPath := "/api/vps/v1/virtual-machines/" + url.PathEscape(strings.TrimSpace(vps.ProviderResourceID))
	var projects []hostingerDockerProject
	if _, err := s.connectorService.HostingerRead(ctx, *connectorID, vmPath+"/docker", nil, &projects); err != nil {
		return s.persistProjectFailure(result, "读取 Hostinger Docker 项目列表失败")
	}
	var listed *hostingerDockerProject
	for index := range projects {
		if strings.EqualFold(strings.TrimSpace(projects[index].Name), result.ComposeProjectName) {
			listed = &projects[index]
			break
		}
	}
	if listed == nil {
		result.HealthStatus = ops.ProjectHealthOffline
		result.RemoteState = "not_found"
		result.ObservedError = "Hostinger 未找到该 Compose 项目"
		result.Message = result.ObservedError
		if err := s.persistProjectResult(result); err != nil {
			return result, err
		}
		return result, fmt.Errorf("%w: %s", ErrOpsHostingerSync, result.ObservedError)
	}

	var detail hostingerDockerProject
	if _, err := s.connectorService.HostingerRead(
		ctx,
		*connectorID,
		vmPath+"/docker/"+url.PathEscape(result.ComposeProjectName),
		nil,
		&detail,
	); err != nil {
		return s.persistProjectFailure(result, "读取 Hostinger Docker 项目详情失败")
	}
	var containers []hostingerDockerContainer
	if _, err := s.connectorService.HostingerRead(
		ctx,
		*connectorID,
		vmPath+"/docker/"+url.PathEscape(result.ComposeProjectName)+"/containers",
		nil,
		&containers,
	); err != nil {
		return s.persistProjectFailure(result, "读取 Hostinger Docker 容器状态失败")
	}
	if len(containers) == 0 {
		containers = detail.Containers
	}
	if len(containers) == 0 {
		containers = listed.Containers
	}

	result.RemoteState = hostingerFirstNonEmpty(detail.State, detail.Status, listed.State, listed.Status)
	result.ContainerCount = len(containers)
	for _, container := range containers {
		if hostingerContainerState(container) == "running" {
			result.RunningContainerCount++
		}
		if hostingerContainerHealth(container) == "healthy" {
			result.HealthyContainerCount++
		}
	}
	result.HealthStatus = hostingerProjectHealth(result.RemoteState, containers)
	result.Message = "Hostinger Docker 项目观察状态已更新"
	if result.HealthStatus == ops.ProjectHealthHealthy {
		result.Message = "Hostinger Docker 项目运行正常"
	} else if result.HealthStatus == ops.ProjectHealthOffline {
		result.Message = "Hostinger Docker 项目当前未运行"
	} else if result.HealthStatus == ops.ProjectHealthDegraded {
		result.Message = "Hostinger Docker 项目存在未正常运行的容器"
	}
	if err := s.persistProjectResult(result); err != nil {
		return result, err
	}
	return result, nil
}

func (s *OpsHostingerSyncService) persistVPSFailure(result *ops.HostingerVPSSyncResult, message string) (*ops.HostingerVPSSyncResult, error) {
	result.ObservedStatus = ops.VPSObservedUnknown
	result.ObservedError = message
	result.Message = message
	if err := s.vpsRepo.UpdateObservedState(
		result.VPSID,
		result.ObservedStatus,
		result.RemoteState,
		result.ObservedSource,
		result.Hostname,
		result.IPv4,
		result.OperatingSystem,
		result.ObservedPlan,
		result.ObservedRegion,
		result.LastObservedAt,
		result.ObservedError,
	); err != nil {
		return result, err
	}
	return result, fmt.Errorf("%w: %s", ErrOpsHostingerSync, message)
}

func (s *OpsHostingerSyncService) persistProjectFailure(result *ops.HostingerProjectSyncResult, message string) (*ops.HostingerProjectSyncResult, error) {
	result.HealthStatus = ops.ProjectHealthUnknown
	result.ObservedError = message
	result.Message = message
	if err := s.persistProjectResult(result); err != nil {
		return result, err
	}
	return result, fmt.Errorf("%w: %s", ErrOpsHostingerSync, message)
}

func (s *OpsHostingerSyncService) persistProjectResult(result *ops.HostingerProjectSyncResult) error {
	return s.projectRepo.UpdateObservedState(
		result.ProjectID,
		result.HealthStatus,
		result.RemoteState,
		result.ObservedSource,
		result.ContainerCount,
		result.RunningContainerCount,
		result.HealthyContainerCount,
		result.LastCheckedAt,
		result.ObservedError,
	)
}

func firstHostingerIPv4(addresses []hostingerIPAddress) string {
	for _, address := range addresses {
		value := strings.TrimSpace(address.Address)
		if value == "" {
			continue
		}
		if parsed := net.ParseIP(value); parsed != nil && parsed.To4() != nil {
			return parsed.To4().String()
		}
	}
	return ""
}

func hostingerVPSObservedStatus(state string) string {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "running":
		return ops.VPSObservedHealthy
	case "stopped", "suspended", "destroyed", "destroying":
		return ops.VPSObservedOffline
	case "starting", "stopping", "creating", "initial", "suspending", "unsuspending", "recreating", "restoring", "recovery", "stopping_recovery", "error":
		return ops.VPSObservedDegraded
	default:
		return ops.VPSObservedUnknown
	}
}

func hostingerProjectHealth(state string, containers []hostingerDockerContainer) string {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "stopped", "created":
		return ops.ProjectHealthOffline
	case "running":
		if len(containers) == 0 {
			return ops.ProjectHealthDegraded
		}
		for _, container := range containers {
			if hostingerContainerState(container) != "running" {
				return ops.ProjectHealthDegraded
			}
			if health := hostingerContainerHealth(container); health == "unhealthy" {
				return ops.ProjectHealthDegraded
			}
		}
		return ops.ProjectHealthHealthy
	case "mixed":
		return ops.ProjectHealthDegraded
	default:
		return ops.ProjectHealthUnknown
	}
}

func hostingerContainerState(container hostingerDockerContainer) string {
	return strings.ToLower(strings.TrimSpace(hostingerFirstNonEmpty(container.State, container.Status)))
}

func hostingerContainerHealth(container hostingerDockerContainer) string {
	return strings.ToLower(strings.TrimSpace(container.Health))
}

func hostingerFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
