package localbackup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"commerce-platform/internal/pkg/maintenancegate"
)

// This is a release contract and operator attestation, not an adapter that
// rebuilds or automatically verifies any external provider's data.
type ExternalProjectionReconciliationContract struct {
	Version                   int                                           `json:"version"`
	Scope                     string                                        `json:"scope"`
	SiteID                    string                                        `json:"site_id"`
	Environment               string                                        `json:"environment"`
	TaskID                    string                                        `json:"task_id"`
	PackageID                 string                                        `json:"package_id"`
	DatabaseOID               int64                                         `json:"database_oid"`
	DatabaseFingerprint       string                                        `json:"database_fingerprint"`
	MediaFingerprint          string                                        `json:"media_fingerprint"`
	ApplicationBoundarySHA256 string                                        `json:"application_boundary_sha256"`
	Configuration             *maintenancegate.ExternalProjectionBoundary   `json:"configuration"`
	Verification              string                                        `json:"verification"`
	Requirements              []ExternalProjectionReconciliationRequirement `json:"requirements"`
	Blockers                  []string                                      `json:"blockers"`
	ContractSHA256            string                                        `json:"contract_sha256"`
}

type ExternalProjectionReconciliationRequirement struct {
	Projection  string `json:"projection"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

type ExternalProjectionReconciliationAttestation struct {
	Projection        string `json:"projection"`
	Action            string `json:"action"`
	EvidenceReference string `json:"evidence_reference"`
	EvidenceSHA256    string `json:"evidence_sha256"`
}

type ExternalProjectionReconciliationEvidence struct {
	ContractSHA256 string                                        `json:"contract_sha256"`
	Attestations   []ExternalProjectionReconciliationAttestation `json:"attestations"`
}

func (manager *Manager) buildTaskBoundExternalProjectionReconciliationContract(task *Task, report *ReconciliationReport) *ExternalProjectionReconciliationContract {
	cfg := manager.Configuration
	contract := &ExternalProjectionReconciliationContract{
		Version: 1, Scope: "whole_application", SiteID: cfg.SiteID, Environment: cfg.Environment,
		TaskID:              task.ID,
		DatabaseFingerprint: report.DatabaseFingerprint, MediaFingerprint: report.MediaFingerprint,
		Verification: "operator_attestation_required_external_state_not_automatically_verified",
		Requirements: []ExternalProjectionReconciliationRequirement{}, Blockers: []string{},
	}
	if !taskHasSwitchedWholeApplicationRecoveryJournal(task) {
		contract.Blockers = append(contract.Blockers, "whole_application_paired_recovery_required_partial_release_unsupported")
	}
	if task.Journal != nil {
		contract.DatabaseOID = task.Journal.NewDatabaseOID
	}
	if task.Manifest != nil {
		contract.PackageID = task.Manifest.PackageID
	}
	data, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, "application-boundary.json"))
	var boundary maintenancegate.ApplicationBoundary
	if err != nil || decodeStrictJSON(data, &boundary) != nil || cfg.validateRuntimeBoundary() != nil {
		contract.Blockers = append(contract.Blockers, "registered_application_boundary_missing_invalid_or_mismatched")
	} else {
		contract.ApplicationBoundarySHA256 = maintenancegate.FingerprintExternalProjectionConfiguration(boundary)
		contract.Configuration = boundary.ExternalProjections
		if (boundary.Protocol != 2 && boundary.Protocol != 3) || boundary.ValidateRegisteredRecoveryProjectionProtocol() != nil || boundary.ExternalProjections == nil {
			contract.Blockers = append(contract.Blockers, "external_projection_configuration_unknown_reregister_with_current_business_build")
		}
	}
	if len(contract.Blockers) == 0 {
		searchAction := "sql_search_verified_without_external_index"
		searchDescription := "按恢复后 SQL 核对历史搜索、权限及不可见记录；业务配置已明确没有外部搜索索引。"
		if contract.Configuration.CustomerServiceSearchMode == "external_delete_only" {
			searchAction = "external_search_rebuilt_and_verified_against_restored_sql"
			searchDescription = "通过服务商独立流程按恢复后 SQL 重建并核对缺失、过期、已删记录及权限；现有 DELETE 接口不提供重建能力，无法完成时保持隔离。"
		}
		realtimeAction := "process_local_realtime_restart_and_http_baseline_verified"
		realtimeDescription := "核对全部实例已停止、重启将清除进程内事件状态，并独立核对 HTTP 会话/消息/未读快照与浏览器重新连接流程。"
		if contract.Configuration.CustomerServiceRealtimeMode == "redis_stream" {
			realtimeAction = "scoped_stream_and_dedupe_invalidated_with_http_baseline_verified"
			realtimeDescription = "证明 Stream 与按 hash tag 前缀生成的全部去重键仅属于本商城，再在停写窗口失效旧投影并核对 HTTP 基线；只清 Stream 或只改名称后缀不合格，归属不明继续隔离。"
		}
		contract.Requirements = []ExternalProjectionReconciliationRequirement{
			{Projection: "customer_service_search", Action: searchAction, Description: searchDescription},
			{Projection: "redis_cache", Action: "owned_cache_invalidated_with_jobs_leases_and_receipts_preserved", Description: "逐项核对本商城缓存归属并失效旧值，核对持久任务、锁、租约与幂等事实；禁止对共享 Redis 使用 FLUSHDB/FLUSHALL 或按未知前缀清理。"},
			{Projection: "customer_service_realtime", Action: realtimeAction, Description: realtimeDescription},
			{Projection: "media_delivery_and_browser", Action: "scoped_media_edge_cache_and_browser_refresh_verified", Description: "独立核对公开/私有附件权限、旧 CDN/浏览器内容、下载身份及恢复后 HTTP 查询；涉及 SQL 运维连接器的投影另列清单，不因本机环境未配置而跳过。"},
		}
	}
	// CapturedAt is intentionally excluded: an unchanged frozen snapshot has the
	// same contract, while task/database/media/runtime identity changes invalidate it.
	contract.ContractSHA256 = maintenancegate.FingerprintExternalProjectionConfiguration(contract)
	return contract
}

func taskHasSwitchedWholeApplicationRecoveryJournal(task *Task) bool {
	return task != nil && task.Kind == "upload" && task.Manifest != nil && task.Manifest.Scope == "whole_application" && task.Journal != nil && task.Journal.DatabaseSwitched && task.Journal.NewDatabaseOID > 0
}

func validateTaskBoundExternalProjectionReconciliationEvidence(contract *ExternalProjectionReconciliationContract, evidence *ExternalProjectionReconciliationEvidence) error {
	if contract == nil || contract.Version != 1 || contract.Scope != "whole_application" || len(contract.Blockers) != 0 || contract.Configuration == nil || contract.Configuration.Validate() != nil || len(contract.Requirements) != 4 {
		return errors.New("external projection configuration is unknown or blocked; register the business configuration and inspect reconciliation again")
	}
	if evidence == nil || evidence.ContractSHA256 != contract.ContractSHA256 || !isValidExternalProjectionEvidenceSHA256(evidence.ContractSHA256) {
		return errors.New("external projection evidence must bind the latest task, database, media and registered runtime contract")
	}
	if len(evidence.Attestations) != len(contract.Requirements) {
		return errors.New("all four external projection operator attestations are required")
	}
	seen := map[string]bool{}
	for _, attestation := range evidence.Attestations {
		matched := false
		for _, requirement := range contract.Requirements {
			if requirement.Projection == attestation.Projection && requirement.Action == attestation.Action {
				matched = true
			}
		}
		if !matched || seen[attestation.Projection] || len(strings.TrimSpace(attestation.EvidenceReference)) < 4 || len(attestation.EvidenceReference) > 2000 || !isValidExternalProjectionEvidenceSHA256(attestation.EvidenceSHA256) {
			return errors.New("projection action, unique identity, evidence reference and independent record SHA-256 must all be valid")
		}
		seen[attestation.Projection] = true
	}
	return nil
}

func isValidExternalProjectionEvidenceSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !(character >= '0' && character <= '9' || character >= 'a' && character <= 'f') {
			return false
		}
	}
	return value != strings.Repeat("0", 64)
}
