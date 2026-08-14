package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

type OpsDeploymentPreflightService struct {
	projectRepo   *repository.OpsProjectBindingRepository
	vpsRepo       *repository.OpsVPSBindingRepository
	connectorRepo *repository.OpsConnectorRepository
	domainRepo    *repository.OpsDomainBindingRepository
}

type deploymentPreflightEvidenceSnapshot struct {
	generatedAt time.Time
	connectors  []ops.Connector
	domains     []ops.DomainBinding
	vpsByID     map[uint]ops.VPSBinding
}

func NewOpsDeploymentPreflightService(
	projectRepo *repository.OpsProjectBindingRepository,
	vpsRepo *repository.OpsVPSBindingRepository,
	connectorRepo *repository.OpsConnectorRepository,
	domainRepo *repository.OpsDomainBindingRepository,
) *OpsDeploymentPreflightService {
	return &OpsDeploymentPreflightService{
		projectRepo:   projectRepo,
		vpsRepo:       vpsRepo,
		connectorRepo: connectorRepo,
		domainRepo:    domainRepo,
	}
}

func (s *OpsDeploymentPreflightService) EvaluateProject(projectID uint) (*ops.DeploymentPreflight, error) {
	if s == nil || s.projectRepo == nil || s.vpsRepo == nil || s.connectorRepo == nil || s.domainRepo == nil {
		return nil, errors.New("operations deployment preflight service is not configured")
	}

	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, err
	}

	evidence, err := s.loadPreflightEvidenceSnapshot(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	vps, vpsErr := lookupVPSSnapshot(evidence.vpsByID, project.VPSBindingID)
	return s.buildProjectPreflightReport(project, evidence, vps, vpsErr), nil
}

func (s *OpsDeploymentPreflightService) EvaluateOverview() (*ops.DeploymentPreflightOverview, error) {
	if s == nil || s.projectRepo == nil || s.vpsRepo == nil || s.connectorRepo == nil || s.domainRepo == nil {
		return nil, errors.New("operations deployment preflight service is not configured")
	}

	projects, err := s.projectRepo.List()
	if err != nil {
		return nil, fmt.Errorf("load operations projects for deployment preflight overview: %w", err)
	}
	evidence, err := s.loadPreflightEvidenceSnapshot(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	overview := &ops.DeploymentPreflightOverview{
		Environment:  "all",
		GeneratedAt:  evidence.generatedAt,
		ProjectCount: len(projects),
		Projects:     make([]ops.DeploymentPreflightSummary, 0, len(projects)),
	}
	categoryTotals := make(map[string]ops.DeploymentPreflightGroup)
	for index := range projects {
		vps, vpsErr := lookupVPSSnapshot(evidence.vpsByID, projects[index].VPSBindingID)
		report := s.buildProjectPreflightReport(&projects[index], evidence, vps, vpsErr)
		summary := summarizeDeploymentPreflight(report)
		overview.Projects = append(overview.Projects, summary)
		mergePreflightGroups(categoryTotals, report.Categories)
		switch report.StatusLevel {
		case ops.DeploymentStatusReady:
			overview.ReadyCount++
		case ops.DeploymentStatusReview:
			overview.ReviewCount++
		default:
			overview.BlockedCount++
		}
		if report.WarningCount > 0 {
			overview.WarningCount++
		}
	}
	overview.Categories = sortedPreflightGroups(categoryTotals)
	return overview, nil
}

func (s *OpsDeploymentPreflightService) loadPreflightEvidenceSnapshot(generatedAt time.Time) (*deploymentPreflightEvidenceSnapshot, error) {
	connectors, err := s.connectorRepo.List()
	if err != nil {
		return nil, fmt.Errorf("load operations connectors for deployment preflight: %w", err)
	}
	domains, err := s.domainRepo.List()
	if err != nil {
		return nil, fmt.Errorf("load operations domains for deployment preflight: %w", err)
	}
	vpsRecords, err := s.vpsRepo.List()
	if err != nil {
		return nil, fmt.Errorf("load operations VPS bindings for deployment preflight: %w", err)
	}
	vpsByID := make(map[uint]ops.VPSBinding, len(vpsRecords))
	for _, vps := range vpsRecords {
		vpsByID[vps.ID] = vps
	}
	return &deploymentPreflightEvidenceSnapshot{
		generatedAt: generatedAt,
		connectors:  connectors,
		domains:     domains,
		vpsByID:     vpsByID,
	}, nil
}

func lookupVPSSnapshot(vpsByID map[uint]ops.VPSBinding, id uint) (*ops.VPSBinding, error) {
	vps, ok := vpsByID[id]
	if !ok || id == 0 {
		return nil, repository.ErrRecordNotFound
	}
	return &vps, nil
}

func (s *OpsDeploymentPreflightService) buildProjectPreflightReport(
	project *ops.ProjectBindingView,
	evidence *deploymentPreflightEvidenceSnapshot,
	vps *ops.VPSBinding,
	vpsErr error,
) *ops.DeploymentPreflight {
	report := &ops.DeploymentPreflight{
		ProjectID:   project.ID,
		Project:     project.Name,
		Environment: project.Environment,
		GeneratedAt: evidence.generatedAt,
		Checks:      make([]ops.DeploymentPreflightCheck, 0, 20),
	}

	report.Checks = buildDeploymentPreflightChecks(project, evidence.connectors, evidence.domains, vps, vpsErr)
	report.BlockingCount, report.WarningCount, report.PassCount, report.InfoCount = countDeploymentPreflightChecks(report.Checks)

	report.StatusLevel = preflightStatusLevel(report.BlockingCount, report.WarningCount)
	report.Ready = report.StatusLevel == ops.DeploymentStatusReady
	report.Summary = preflightReportSummary(report.BlockingCount, report.WarningCount)
	report.NextActions = preflightNextActions(report.Checks)
	report.Categories = summarizePreflightGroups(report.Checks)

	return report
}

func checkComposeSource(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	source := strings.TrimSpace(project.ComposeSource)
	if source == "" {
		return blockedCheck("compose_source", "compose", "Compose 来源", "未登记 Compose 来源", "需要明确部署使用的 Compose 文件或来源。")
	}
	baseline, hasBaseline := composeBaselineForEnvironment(project.Environment)
	if hasBaseline && source != baseline.Source {
		return warningCheck("compose_source", "compose", "Compose 来源", "生产项目当前不是标准 compose.prod.yml 来源。", source)
	}
	return passCheck("compose_source", "compose", "Compose 来源", source, "已登记 Compose 来源。")
}

func checkProjectName(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	name := strings.TrimSpace(project.Name)
	composeName := strings.TrimSpace(project.ComposeProjectName)
	if name == "" || composeName == "" {
		return blockedCheck("project_name", "compose", "项目名", "项目绑定名或 Compose 项目名缺失。", "项目名和 Compose 项目名必须同时明确。")
	}
	if name != composeName {
		return blockedCheck("project_name", "compose", "项目名", "项目绑定名与 Compose 项目名不一致。", fmt.Sprintf("绑定：%s；Compose：%s", name, composeName))
	}
	return passCheck("project_name", "compose", "项目名", composeName, "项目绑定名与 Compose 项目名一致。")
}

func checkImage(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	tag := strings.TrimSpace(project.CurrentImageTag)
	if tag == "" {
		return blockedCheck("image_tag", "version", "镜像标签", "未登记当前镜像标签。", "需要记录 master、完整 SHA 标签或 digest。")
	}
	if !isImmutableImageTag(tag) {
		return warningCheck("image_tag", "version", "镜像标签", "当前镜像标签不是不可变 SHA 标签或 digest。", tag)
	}
	return passCheck("image_tag", "version", "镜像标签", tag, "已登记不可变镜像标签或 digest。")
}

func checkCommit(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	commit := strings.TrimSpace(project.CurrentCommitSHA)
	if commit == "" {
		if strings.TrimSpace(project.CurrentImageTag) == "master" {
			return warningCheck("commit_sha", "version", "Commit SHA", "当前使用 master 流动标签，尚未固化 Commit SHA。", "生产发布建议记录完整 40 位 Commit SHA。")
		}
		if isDigestImageTag(project.CurrentImageTag) {
			return warningCheck("commit_sha", "version", "Commit SHA", "当前使用镜像 digest，但尚未关联源码 Commit SHA。", "建议记录触发该 digest 的完整 40 位 Commit SHA。")
		}
		return blockedCheck("commit_sha", "version", "Commit SHA", "未登记当前 Commit SHA。", "无法将镜像标签与源码版本建立证据关联。")
	}
	if len(commit) != 40 || !isHexString(commit) {
		return blockedCheck("commit_sha", "version", "Commit SHA", "Commit SHA 格式无效。", commit)
	}
	return passCheck("commit_sha", "version", "Commit SHA", commit, "已登记完整 Commit SHA。")
}

func checkImageCommitConsistency(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	tag := strings.TrimSpace(project.CurrentImageTag)
	commit := strings.TrimSpace(project.CurrentCommitSHA)
	if strings.HasPrefix(tag, "sha-") && commit != "" && len(commit) == 40 && isHexString(commit) {
		if strings.EqualFold(strings.TrimPrefix(tag, "sha-"), commit) {
			return passCheck("image_commit", "version", "镜像 / Commit 一致性", "镜像 SHA 标签与 Commit SHA 一致。", tag)
		}
		return blockedCheck("image_commit", "version", "镜像 / Commit 一致性", "镜像 SHA 标签与 Commit SHA 不一致。", fmt.Sprintf("镜像：%s；Commit：%s", tag, commit))
	}
	if tag == "master" && commit != "" {
		return warningCheck("image_commit", "version", "镜像 / Commit 一致性", "镜像使用 master 流动标签，Commit SHA 只能作为辅助证据。", commit)
	}
	if isDigestImageTag(tag) {
		return infoCheck("image_commit", "version", "镜像 / Commit 一致性", "镜像使用 digest，无法从标签直接校验 Commit。", tag)
	}
	return infoCheck("image_commit", "version", "镜像 / Commit 一致性", "镜像标签与 Commit 暂无可校验关系。", firstNonEmptyString(tag, commit, "未登记版本证据"))
}

func checkServices(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	baseline, ok := composeBaselineForEnvironment(project.Environment)
	if !ok {
		return boundaryBaselineUnavailable("services", "服务边界", project.Environment)
	}
	actual := splitList(project.Services)
	return compareBoundary("services", "boundary", "服务边界", actual, baseline.Services, "服务")
}

func checkNetworks(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	baseline, ok := composeBaselineForEnvironment(project.Environment)
	if !ok {
		return boundaryBaselineUnavailable("networks", "网络边界", project.Environment)
	}
	actual := splitList(project.Networks)
	return compareBoundary("networks", "boundary", "网络边界", actual, baseline.Networks, "网络")
}

func checkVolumes(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	baseline, ok := composeBaselineForEnvironment(project.Environment)
	if !ok {
		return boundaryBaselineUnavailable("volumes", "卷边界", project.Environment)
	}
	actual := splitList(project.Volumes)
	required := make([][]string, 0, len(baseline.VolumeKeys))
	for _, key := range baseline.VolumeKeys {
		required = append(required, composeVolumeCandidates(project.ComposeProjectName, key))
	}
	return compareBoundaryAlternatives("volumes", "boundary", "卷边界", actual, required, baseline.VolumeKeys, "卷")
}

func checkGateway(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	baseline, ok := composeBaselineForEnvironment(project.Environment)
	if !ok {
		return boundaryBaselineUnavailable("gateway_boundary", "网关边界", project.Environment)
	}
	network := strings.TrimSpace(project.GatewayNetwork)
	alias := strings.TrimSpace(project.GatewayAlias)
	if network == "" || alias == "" {
		return blockedCheck("gateway_boundary", "boundary", "网关边界", "共享网关网络或别名缺失。", "需要同时登记 shared-edge 和 theme-web 这类边界信息。")
	}
	if network != baseline.GatewayNetwork || alias != baseline.GatewayAlias {
		return warningCheck("gateway_boundary", "boundary", "网关边界", "当前网关边界与生产基线不同，请人工确认。", fmt.Sprintf("网络：%s；别名：%s", network, alias))
	}
	return passCheck("gateway_boundary", "boundary", "网关边界", fmt.Sprintf("%s / %s", network, alias), "已登记共享网关网络和边界别名。")
}

func checkVPS(project *ops.ProjectBindingView, vps *ops.VPSBinding, err error) ops.DeploymentPreflightCheck {
	if err != nil || vps == nil {
		return blockedCheck("vps", "infra", "VPS", "绑定 VPS 不存在或无法读取。", fmt.Sprintf("VPS binding ID：%d", project.VPSBindingID))
	}
	if !vps.Enabled || vps.Status == ops.VPSStatusDisabled {
		return blockedCheck("vps", "infra", "VPS", "绑定 VPS 已停用。", vps.Name)
	}
	if vps.Provider != ops.VPSProviderHostinger {
		return warningCheck("vps", "infra", "VPS", "绑定 VPS 不是 Hostinger 资源，请人工确认连接器和发布边界。", fmt.Sprintf("%s / %s", vps.Provider, vps.Name))
	}
	if strings.TrimSpace(vps.ProviderResourceID) == "" {
		return blockedCheck("vps", "infra", "VPS", "未登记 Hostinger VPS 资源 ID。", vps.Name)
	}
	return passCheck("vps", "infra", "VPS", fmt.Sprintf("%s · %s", vps.Name, vps.ProviderResourceID), "VPS 资源绑定和声明式状态可读取。")
}

func checkVPSObserved(project *ops.ProjectBindingView, vps *ops.VPSBinding, err error) ops.DeploymentPreflightCheck {
	if err != nil || vps == nil {
		return blockedCheck("vps_observed", "infra", "VPS 观察状态", "绑定 VPS 不存在，无法检查观察状态。", fmt.Sprintf("VPS binding ID：%d", project.VPSBindingID))
	}
	if vps.LastObservedAt == nil || vps.ObservedStatus == ops.VPSObservedUnknown {
		return warningCheck("vps_observed", "infra", "VPS 观察状态", "VPS 尚未完成远端只读同步。", "需要 Hostinger VPS 同步结果作为发布前证据。")
	}
	if vps.ObservedStatus == ops.VPSObservedOffline || vps.ObservedStatus == ops.VPSObservedDegraded {
		return blockedCheck("vps_observed", "infra", "VPS 观察状态", "VPS 远端观察状态不可用于发布。", fmt.Sprintf("%s · %s", vps.ObservedStatus, vps.ObservedState))
	}
	mismatches := make([]string, 0)
	if strings.TrimSpace(vps.Hostname) != "" && strings.TrimSpace(vps.ObservedHostname) != "" && !strings.EqualFold(strings.TrimSpace(vps.Hostname), strings.TrimSpace(vps.ObservedHostname)) {
		mismatches = append(mismatches, fmt.Sprintf("hostname %s != %s", vps.Hostname, vps.ObservedHostname))
	}
	if strings.TrimSpace(vps.IPv4) != "" && strings.TrimSpace(vps.ObservedIPv4) != "" && strings.TrimSpace(vps.IPv4) != strings.TrimSpace(vps.ObservedIPv4) {
		mismatches = append(mismatches, fmt.Sprintf("ipv4 %s != %s", vps.IPv4, vps.ObservedIPv4))
	}
	if len(mismatches) > 0 {
		return blockedCheck("vps_observed", "infra", "VPS 观察状态", "VPS 期望身份与观察身份不一致。", strings.Join(mismatches, "；"))
	}
	return passCheck("vps_observed", "infra", "VPS 观察状态", vps.ObservedStatus, fmt.Sprintf("%s · %s · %s", firstNonEmptyString(vps.ObservedSource, "未设置来源"), firstNonEmptyString(vps.ObservedState, "未设置远端状态"), vps.LastObservedAt.Format(time.RFC3339)))
}

func checkVPSConnector(vps *ops.VPSBinding, connector *ops.Connector, err error) ops.DeploymentPreflightCheck {
	if vps == nil {
		return blockedCheck("vps_connector", "connector", "VPS 连接器", "无法读取 VPS 绑定，不能确认 VPS 连接器。", "先修正 VPS 绑定后重新生成报告。")
	}
	if vps.ConnectorID == nil || *vps.ConnectorID == 0 || err != nil || connector == nil {
		return blockedCheck("vps_connector", "connector", "VPS 连接器", "VPS 未绑定可读取的 Hostinger 连接器。", fmt.Sprintf("VPS：%s", vps.Name))
	}
	if connector.Provider != ops.ConnectorProviderHostinger {
		return blockedCheck("vps_connector", "connector", "VPS 连接器", "VPS 绑定的连接器不是 Hostinger。", connector.Name)
	}
	if connector.Environment != "" && vps.Environment != "" && connector.Environment != vps.Environment {
		return blockedCheck("vps_connector", "connector", "VPS 连接器", "VPS 连接器环境与 VPS 环境不一致。", fmt.Sprintf("VPS：%s；连接器：%s", vps.Environment, connector.Environment))
	}
	if !connector.Enabled || connector.Status != ops.ConnectorStatusActive {
		return blockedCheck("vps_connector", "connector", "VPS 连接器", "VPS 的 Hostinger 连接器未启用或尚未通过测试。", fmt.Sprintf("%s · %s", connector.Name, connector.Status))
	}
	if !connectorCredentialConfigured(*connector) {
		return blockedCheck("vps_connector", "connector", "VPS 连接器", "VPS 的 Hostinger 连接器没有已配置凭据。", connector.Name)
	}
	if !hasReadScope(connector.Scopes) {
		return warningCheck("vps_connector", "connector", "VPS 连接器", "VPS 连接器未明确登记 VPS/project 读取作用域。", connector.Scopes)
	}
	return passCheck("vps_connector", "connector", "VPS 连接器", connector.Name, "VPS 资源同步连接器已启用、已测试并具备读取凭据。")
}

func checkHostingerConnector(project *ops.ProjectBindingView, connector *ops.Connector, err error) ops.DeploymentPreflightCheck {
	return checkHostingerConnectorWithSource(project, connector, err, "project")
}

func checkHostingerConnectorWithSource(
	project *ops.ProjectBindingView,
	connector *ops.Connector,
	err error,
	source string,
) ops.DeploymentPreflightCheck {
	if err != nil || connector == nil {
		return blockedCheck("hostinger_connector", "connector", "Hostinger 连接器", "项目没有可读取的 Hostinger 连接器。", "需要登记项目显式连接器，或确保 VPS 绑定提供可继承的 Hostinger 只读连接器。")
	}
	if connector.Provider != ops.ConnectorProviderHostinger {
		return blockedCheck("hostinger_connector", "connector", "Hostinger 连接器", "项目生效连接器不是 Hostinger。", connector.Name)
	}
	if connector.Environment != "" && project.Environment != "" && connector.Environment != project.Environment {
		return blockedCheck("hostinger_connector", "connector", "Hostinger 连接器", "项目生效连接器环境与项目环境不一致。", fmt.Sprintf("项目：%s；连接器：%s", project.Environment, connector.Environment))
	}
	if !connector.Enabled || connector.Status != ops.ConnectorStatusActive {
		return blockedCheck("hostinger_connector", "connector", "Hostinger 连接器", "项目生效的 Hostinger 连接器未启用或尚未通过测试。", fmt.Sprintf("%s · %s", connector.Name, connector.Status))
	}
	if !connectorCredentialConfigured(*connector) {
		return blockedCheck("hostinger_connector", "connector", "Hostinger 连接器", "项目生效的 Hostinger 连接器没有已配置凭据。", connector.Name)
	}
	if !hasReadScope(connector.Scopes) {
		return warningCheck("hostinger_connector", "connector", "Hostinger 连接器", "项目生效连接器未明确登记 VPS/project 读取作用域。", connector.Scopes)
	}
	return passCheck("hostinger_connector", "connector", "Hostinger 连接器", connector.Name, hostingerConnectorEvidenceDetail(connector.Name, source))
}

func hostingerConnectorEvidenceDetail(name, source string) string {
	switch source {
	case "vps":
		return fmt.Sprintf("项目未显式绑定连接器，沿用 VPS 绑定的 Hostinger 只读连接器：%s。", name)
	case "project":
		return fmt.Sprintf("项目显式绑定的 Hostinger 连接器已启用、已测试并具备读取凭据：%s。", name)
	default:
		return fmt.Sprintf("Hostinger 只读连接器已启用、已测试并具备读取凭据：%s。", name)
	}
}

func checkHostingerProjectIdentity(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	if strings.TrimSpace(project.ProviderResourceID) == "" {
		return warningCheck("hostinger_project_identity", "infra", "Hostinger 项目标识", "项目未登记 Hostinger 项目 ID。", "当前 Hostinger 只读同步可通过 Compose 项目名定位，但项目 ID 缺失会降低证据质量。")
	}
	return passCheck("hostinger_project_identity", "infra", "Hostinger 项目标识", project.ProviderResourceID, "已登记 Hostinger 项目 ID。")
}

func checkGHCRConnector(connectors []ops.Connector, environment string) ops.DeploymentPreflightCheck {
	candidates := deploymentConnectors(connectors, environment, ops.ConnectorProviderGHCR, ops.ConnectorProviderGitHub)
	if len(candidates) == 0 {
		return warningCheck("ghcr_connector", "connector", "镜像连接器", "未登记 GHCR/GitHub 只读连接器。", "当前报告只能校验本地台账中的镜像标签，不能读取 GHCR 包状态。")
	}
	best := bestConfiguredConnector(candidates)
	if best == nil {
		return warningCheck("ghcr_connector", "connector", "镜像连接器", "GHCR/GitHub 连接器尚未启用、测试或配置凭据。", connectorSummary(candidates))
	}
	if !hasPackageReadScope(best.Scopes) {
		return warningCheck("ghcr_connector", "connector", "镜像连接器", "镜像连接器未明确登记 packages:read 或 repo:read 作用域。", best.Scopes)
	}
	return passCheck("ghcr_connector", "connector", "镜像连接器", best.Name, "已登记可用于镜像只读检查的连接器。")
}

func checkDomains(project *ops.ProjectBindingView, domains []ops.DomainBinding) ops.DeploymentPreflightCheck {
	relevant, unboundMatches := projectDeploymentDomains(domains, project)
	if len(relevant) == 0 {
		if len(unboundMatches) > 0 {
			return warningCheck(
				"domains",
				"domain",
				"域名边界",
				"存在目标匹配的域名，但尚未显式绑定当前项目。",
				"未计入项目证据："+domainSummary(unboundMatches),
			)
		}
		if project.Environment == ops.ProjectEnvironmentProduction {
			return blockedCheck("domains", "domain", "域名边界", "生产项目没有显式绑定的公网域名。", "需要至少将 canonical、alias 或 admin 域名绑定到当前项目。")
		}
		return warningCheck("domains", "domain", "域名边界", "当前项目没有显式绑定的公网域名。", project.Environment)
	}

	matchedTargets := 0
	warnings := make([]string, 0)
	blocks := make([]string, 0)
	if len(unboundMatches) > 0 {
		warnings = append(warnings, fmt.Sprintf("%d 个目标匹配域名尚未显式绑定项目，未计入检查", len(unboundMatches)))
	}
	for _, domain := range relevant {
		if targetMatchesProject(domain.Target, project) {
			matchedTargets++
		}
		if domain.Status != ops.DomainStatusActive {
			warnings = append(warnings, fmt.Sprintf("%s 期望状态 %s", domain.Domain, domain.Status))
		}
		switch domain.ObservedStatus {
		case ops.DomainObservedMatched:
		case ops.DomainObservedUnknown:
			warnings = append(warnings, fmt.Sprintf("%s 未同步", domain.Domain))
		case ops.DomainObservedDrifted, ops.DomainObservedError:
			blocks = append(blocks, fmt.Sprintf("%s 实际状态 %s", domain.Domain, domain.ObservedStatus))
		}
	}
	if matchedTargets == 0 {
		blocks = append(blocks, "没有域名目标指向当前项目网关别名")
	}
	if len(blocks) > 0 {
		return blockedCheck("domains", "domain", "域名边界", "域名状态存在阻断问题。", strings.Join(append(blocks, warnings...), "；"))
	}
	if len(warnings) > 0 {
		return warningCheck("domains", "domain", "域名边界", "域名边界需要人工确认。", strings.Join(warnings, "；"))
	}
	return passCheck("domains", "domain", "域名边界", fmt.Sprintf("%d 个域名已匹配", len(relevant)), domainSummary(relevant))
}

func checkCloudflareConnectors(
	connectors []ops.Connector,
	project *ops.ProjectBindingView,
	domains []ops.DomainBinding,
) ops.DeploymentPreflightCheck {
	cloudflareDomains := make([]ops.DomainBinding, 0)
	projectDomains, _ := projectDeploymentDomains(domains, project)
	for _, domain := range projectDomains {
		if domain.Provider == ops.DomainProviderCloudflare {
			cloudflareDomains = append(cloudflareDomains, domain)
		}
	}
	if len(cloudflareDomains) == 0 {
		return infoCheck("cloudflare_connector", "connector", "Cloudflare 连接器", "当前项目没有显式绑定的 Cloudflare 公网域名。", project.Environment)
	}

	connectorByID := map[uint]ops.Connector{}
	for _, connector := range connectors {
		connectorByID[connector.ID] = connector
	}
	warnings := make([]string, 0)
	for _, domain := range cloudflareDomains {
		if domain.ConnectorID == nil || *domain.ConnectorID == 0 {
			warnings = append(warnings, domain.Domain+" 未绑定连接器")
			continue
		}
		connector, ok := connectorByID[*domain.ConnectorID]
		if !ok {
			warnings = append(warnings, domain.Domain+" 的连接器不存在")
			continue
		}
		if connector.Provider != ops.ConnectorProviderCloudflare {
			warnings = append(warnings, domain.Domain+" 的连接器不是 Cloudflare")
			continue
		}
		if !connector.Enabled || connector.Status != ops.ConnectorStatusActive || !connectorCredentialConfigured(connector) {
			warnings = append(warnings, fmt.Sprintf("%s 连接器未就绪：%s", domain.Domain, connector.Name))
			continue
		}
		if !hasCloudflareReadScope(connector.Scopes) {
			warnings = append(warnings, fmt.Sprintf("%s 连接器缺少 zones/dns 读取作用域", connector.Name))
		}
	}
	if len(warnings) > 0 {
		return warningCheck("cloudflare_connector", "connector", "Cloudflare 连接器", "Cloudflare 只读同步证据不完整。", strings.Join(warnings, "；"))
	}
	return passCheck("cloudflare_connector", "connector", "Cloudflare 连接器", "Cloudflare 只读连接器已覆盖公网域名。", domainSummary(cloudflareDomains))
}

func checkRemoteHealth(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	if project.LastCheckedAt == nil || project.HealthStatus == ops.ProjectHealthUnknown {
		return blockedCheck("remote_health", "runtime", "远端观察健康", "尚未获得远端项目健康观察结果。", "请先完成 Hostinger 项目只读同步，再重新生成报告。")
	}
	if project.HealthStatus == ops.ProjectHealthDegraded || project.HealthStatus == ops.ProjectHealthOffline {
		return blockedCheck("remote_health", "runtime", "远端观察健康", "远端项目当前不是健康状态。", fmt.Sprintf("%s · 容器 %d / 运行 %d / 健康 %d", project.HealthStatus, project.ObservedContainerCount, project.ObservedRunningCount, project.ObservedHealthyCount))
	}
	if project.ObservedContainerCount == 0 || project.ObservedHealthyCount != project.ObservedContainerCount {
		return warningCheck("remote_health", "runtime", "远端观察健康", "远端状态标记为健康，但容器计数证据不完整。", fmt.Sprintf("容器 %d / 运行 %d / 健康 %d", project.ObservedContainerCount, project.ObservedRunningCount, project.ObservedHealthyCount))
	}
	return passCheck("remote_health", "runtime", "远端观察健康", "healthy", fmt.Sprintf("容器 %d / 运行 %d / 健康 %d", project.ObservedContainerCount, project.ObservedRunningCount, project.ObservedHealthyCount))
}

func checkDeploymentRecord(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	if project.LastDeploymentAt == nil {
		return warningCheck("deployment_record", "evidence", "部署记录", "未登记最近一次部署时间。", "报告无法关联最近一次发布证据。")
	}
	return passCheck("deployment_record", "evidence", "部署记录", project.LastDeploymentAt.Format(time.RFC3339), "已登记最近一次部署时间。")
}

func checkBackups(project *ops.ProjectBindingView) ops.DeploymentPreflightCheck {
	policy := strings.TrimSpace(project.BackupPolicy)
	restore := strings.TrimSpace(project.RestoreNotes)
	if policy == "" {
		return blockedCheck("backups", "evidence", "备份记录", "未登记备份策略。", "至少需要记录 PostgreSQL、uploads、异地保存和恢复安排。")
	}
	if restore == "" {
		return warningCheck("backups", "evidence", "备份记录", "已登记备份策略，但没有恢复记录。", policy)
	}
	if containsPendingMarker(restore) {
		return warningCheck("backups", "evidence", "备份记录", "恢复记录仍标记为待补录或未完成。", restore)
	}
	return passCheck("backups", "evidence", "备份记录", "已登记备份策略和恢复记录。", fmt.Sprintf("策略：%s；恢复：%s", policy, restore))
}

func compareBoundary(key, category, label string, actual, required []string, noun string) ops.DeploymentPreflightCheck {
	actualSet := make(map[string]struct{}, len(actual))
	for _, item := range actual {
		actualSet[item] = struct{}{}
	}
	requiredSet := make(map[string]struct{}, len(required))
	for _, item := range required {
		requiredSet[item] = struct{}{}
	}

	missing := make([]string, 0)
	for _, item := range required {
		if _, ok := actualSet[item]; !ok {
			missing = append(missing, item)
		}
	}
	extra := make([]string, 0)
	for _, item := range actual {
		if _, ok := requiredSet[item]; !ok {
			extra = append(extra, item)
		}
	}
	if len(missing) > 0 || len(extra) > 0 {
		sort.Strings(missing)
		sort.Strings(extra)
		detail := fmt.Sprintf("已登记：%s", strings.Join(actual, ", "))
		if len(missing) > 0 {
			detail += fmt.Sprintf("；缺少%s：%s", noun, strings.Join(missing, ", "))
		}
		if len(extra) > 0 {
			detail += fmt.Sprintf("；多出%s：%s", noun, strings.Join(extra, ", "))
		}
		return blockedCheck(key, category, label, fmt.Sprintf("%s边界与生产基线不一致。", noun), detail)
	}
	return passCheck(key, category, label, strings.Join(actual, ", "), fmt.Sprintf("%s边界与生产基线一致。", noun))
}

func compareBoundaryAlternatives(
	key, category, label string,
	actual []string,
	required [][]string,
	requiredLabels []string,
	noun string,
) ops.DeploymentPreflightCheck {
	actualSet := make(map[string]struct{}, len(actual))
	for _, item := range actual {
		actualSet[item] = struct{}{}
	}
	missing := make([]string, 0)
	for index, candidates := range required {
		matched := false
		for _, candidate := range candidates {
			if _, ok := actualSet[candidate]; ok {
				matched = true
				break
			}
		}
		if !matched {
			if index < len(requiredLabels) {
				missing = append(missing, requiredLabels[index])
			}
		}
	}
	known := make(map[string]struct{})
	for _, candidates := range required {
		for _, candidate := range candidates {
			known[candidate] = struct{}{}
		}
	}
	extra := make([]string, 0)
	for _, item := range actual {
		if _, ok := known[item]; !ok {
			extra = append(extra, item)
		}
	}
	if len(missing) > 0 || len(extra) > 0 {
		sort.Strings(missing)
		sort.Strings(extra)
		detail := fmt.Sprintf("已登记：%s", strings.Join(actual, ", "))
		if len(missing) > 0 {
			detail += fmt.Sprintf("；缺少%s：%s", noun, strings.Join(missing, ", "))
		}
		if len(extra) > 0 {
			detail += fmt.Sprintf("；多出%s：%s", noun, strings.Join(extra, ", "))
		}
		return blockedCheck(key, category, label, fmt.Sprintf("%s边界与生产基线不一致。", noun), detail)
	}
	return passCheck(key, category, label, strings.Join(actual, ", "), fmt.Sprintf("%s边界与生产基线一致。", noun))
}

func splitList(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	values := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values
}

func hasReadScope(raw string) bool {
	value := strings.ToLower(strings.ReplaceAll(raw, " ", ""))
	return strings.Contains(value, "vps:read") && strings.Contains(value, "project:read")
}

func hasPackageReadScope(raw string) bool {
	value := strings.ToLower(strings.ReplaceAll(raw, " ", ""))
	return strings.Contains(value, "packages:read") || strings.Contains(value, "repo:read") || strings.Contains(value, "read:packages")
}

func hasCloudflareReadScope(raw string) bool {
	value := strings.ToLower(strings.ReplaceAll(raw, " ", ""))
	return strings.Contains(value, "zones:read") && strings.Contains(value, "dns:read")
}

func deploymentConnectors(connectors []ops.Connector, environment string, providers ...string) []ops.Connector {
	providerSet := make(map[string]struct{}, len(providers))
	for _, provider := range providers {
		providerSet[provider] = struct{}{}
	}
	matches := make([]ops.Connector, 0)
	for _, connector := range connectors {
		if connector.Environment != environment {
			continue
		}
		if _, ok := providerSet[connector.Provider]; ok {
			matches = append(matches, connector)
		}
	}
	return matches
}

func bestConfiguredConnector(connectors []ops.Connector) *ops.Connector {
	for index := range connectors {
		if connectors[index].Enabled && connectors[index].Status == ops.ConnectorStatusActive && connectorCredentialConfigured(connectors[index]) {
			return &connectors[index]
		}
	}
	return nil
}

func connectorSummary(connectors []ops.Connector) string {
	parts := make([]string, 0, len(connectors))
	for _, connector := range connectors {
		parts = append(parts, fmt.Sprintf("%s/%s/%t", connector.Name, connector.Status, connector.Enabled))
	}
	return strings.Join(parts, "；")
}

func deploymentDomains(domains []ops.DomainBinding, environment string) []ops.DomainBinding {
	matches := make([]ops.DomainBinding, 0)
	for _, domain := range domains {
		if !domain.Enabled || domain.Environment != environment {
			continue
		}
		switch domain.Role {
		case ops.DomainRoleCanonical, ops.DomainRoleAlias, ops.DomainRoleAdmin, ops.DomainRoleRedirect:
			matches = append(matches, domain)
		}
	}
	return matches
}

func projectDeploymentDomains(
	domains []ops.DomainBinding,
	project *ops.ProjectBindingView,
) ([]ops.DomainBinding, []ops.DomainBinding) {
	if project == nil {
		return nil, nil
	}
	bound := make([]ops.DomainBinding, 0)
	unboundMatches := make([]ops.DomainBinding, 0)
	for _, domain := range deploymentDomains(domains, project.Environment) {
		if domain.ProjectBindingID != nil && *domain.ProjectBindingID == project.ID {
			bound = append(bound, domain)
			continue
		}
		if domain.ProjectBindingID == nil && targetMatchesProject(domain.Target, project) {
			unboundMatches = append(unboundMatches, domain)
		}
	}
	return bound, unboundMatches
}

func targetMatchesProject(target string, project *ops.ProjectBindingView) bool {
	value := strings.TrimSpace(strings.ToLower(target))
	if value == "" {
		return false
	}
	alias := strings.TrimSpace(strings.ToLower(project.GatewayAlias))
	if alias != "" && (value == alias || strings.HasPrefix(value, alias+":")) {
		return true
	}
	projectName := strings.TrimSpace(strings.ToLower(project.ComposeProjectName))
	return projectName != "" && (value == projectName || strings.HasPrefix(value, projectName+":"))
}

func domainSummary(domains []ops.DomainBinding) string {
	parts := make([]string, 0, len(domains))
	for _, domain := range domains {
		parts = append(parts, fmt.Sprintf("%s/%s/%s", domain.Domain, domain.Role, domain.ObservedStatus))
	}
	return strings.Join(parts, "；")
}

func isImmutableImageTag(tag string) bool {
	value := strings.TrimSpace(tag)
	return strings.HasPrefix(value, "sha-") || isDigestImageTag(value)
}

func isDigestImageTag(tag string) bool {
	value := strings.TrimSpace(tag)
	return strings.Contains(value, "@sha256:") || strings.HasPrefix(value, "sha256:")
}

func containsPendingMarker(raw string) bool {
	value := strings.ToLower(raw)
	for _, marker := range []string{"待补录", "待完成", "未完成", "未记录", "pending", "todo"} {
		if strings.Contains(value, marker) {
			return true
		}
	}
	return false
}

func isHexString(value string) bool {
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}

func passCheck(key, category, label, message, detail string) ops.DeploymentPreflightCheck {
	return ops.DeploymentPreflightCheck{Key: key, Category: category, Label: label, Status: ops.DeploymentCheckPass, Message: message, Detail: detail}
}

func infoCheck(key, category, label, message, detail string) ops.DeploymentPreflightCheck {
	return ops.DeploymentPreflightCheck{Key: key, Category: category, Label: label, Status: ops.DeploymentCheckInfo, Message: message, Detail: detail}
}

func warningCheck(key, category, label, message, detail string) ops.DeploymentPreflightCheck {
	return ops.DeploymentPreflightCheck{Key: key, Category: category, Label: label, Status: ops.DeploymentCheckWarning, Message: message, Detail: detail}
}

func blockedCheck(key, category, label, message, detail string) ops.DeploymentPreflightCheck {
	return ops.DeploymentPreflightCheck{Key: key, Category: category, Label: label, Status: ops.DeploymentCheckBlock, Message: message, Detail: detail}
}
