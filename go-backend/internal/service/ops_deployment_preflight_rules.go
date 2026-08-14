package service

import (
	"fmt"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

func buildDeploymentPreflightChecks(
	project *ops.ProjectBindingView,
	connectors []ops.Connector,
	domains []ops.DomainBinding,
	vps *ops.VPSBinding,
	vpsErr error,
) []ops.DeploymentPreflightCheck {
	vpsConnector, vpsConnectorErr := lookupConnectorSnapshot(connectors, vpsConnectorID(vps))
	projectConnectorID := effectiveProjectConnectorID(project, vps)
	projectConnector, projectConnectorErr := lookupConnectorSnapshot(connectors, projectConnectorID)
	projectConnectorSource := projectConnectorBindingSource(project, vps)

	return []ops.DeploymentPreflightCheck{
		checkComposeSource(project),
		checkProjectName(project),
		checkImage(project),
		checkCommit(project),
		checkImageCommitConsistency(project),
		checkServices(project),
		checkNetworks(project),
		checkVolumes(project),
		checkGateway(project),
		checkVPS(project, vps, vpsErr),
		checkVPSObserved(project, vps, vpsErr),
		checkVPSConnector(vps, vpsConnector, vpsConnectorErr),
		checkHostingerConnectorWithSource(project, projectConnector, projectConnectorErr, projectConnectorSource),
		checkHostingerProjectIdentity(project),
		checkGHCRConnector(connectors, project.Environment),
		checkDomains(project, domains),
		checkCloudflareConnectors(connectors, project, domains),
		checkRemoteHealth(project),
		checkDeploymentRecord(project),
		checkBackups(project),
	}
}

func countDeploymentPreflightChecks(checks []ops.DeploymentPreflightCheck) (blocking, warning, pass, info int) {
	for _, check := range checks {
		switch check.Status {
		case ops.DeploymentCheckBlock:
			blocking++
		case ops.DeploymentCheckWarning:
			warning++
		case ops.DeploymentCheckPass:
			pass++
		case ops.DeploymentCheckInfo:
			info++
		}
	}
	return blocking, warning, pass, info
}

func effectiveProjectConnectorID(project *ops.ProjectBindingView, vps *ops.VPSBinding) *uint {
	if project != nil && project.ConnectorID != nil && *project.ConnectorID != 0 {
		return project.ConnectorID
	}
	return vpsConnectorID(vps)
}

func projectConnectorBindingSource(project *ops.ProjectBindingView, vps *ops.VPSBinding) string {
	if project != nil && project.ConnectorID != nil && *project.ConnectorID != 0 {
		return "project"
	}
	if vps != nil && vps.ConnectorID != nil && *vps.ConnectorID != 0 {
		return "vps"
	}
	return ""
}

func lookupConnectorSnapshot(connectors []ops.Connector, id *uint) (*ops.Connector, error) {
	if id == nil || *id == 0 {
		return nil, repository.ErrRecordNotFound
	}
	for index := range connectors {
		if connectors[index].ID == *id {
			return &connectors[index], nil
		}
	}
	return nil, repository.ErrRecordNotFound
}

func vpsConnectorID(vps *ops.VPSBinding) *uint {
	if vps == nil {
		return nil
	}
	return vps.ConnectorID
}

func preflightReportSummary(blockingCount, warningCount int) string {
	switch {
	case blockingCount > 0:
		return fmt.Sprintf("存在 %d 项阻断问题，当前不具备只读报告定义的发布前条件。", blockingCount)
	case warningCount > 0:
		return fmt.Sprintf("未发现阻断问题，但仍有 %d 项需要人工确认。", warningCount)
	default:
		return "项目级发布前检查通过，当前登记边界和观察状态满足第一版 preflight 条件。"
	}
}
