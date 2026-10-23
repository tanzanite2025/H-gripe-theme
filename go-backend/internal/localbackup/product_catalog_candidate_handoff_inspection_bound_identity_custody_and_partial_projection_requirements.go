package localbackup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
)

// A read-only inspection of a still-isolated candidate. It is never an ownership
// transfer, a switched-database reconciliation record, or permission to release.
type ProductCatalogCandidateHandoffInspection struct {
	Version                      int                                           `json:"version"`
	CapturedAt                   time.Time                                     `json:"captured_at"`
	Scope                        string                                        `json:"scope"`
	SiteID                       string                                        `json:"site_id"`
	Environment                  string                                        `json:"environment"`
	TaskID                       string                                        `json:"task_id"`
	PackageID                    string                                        `json:"package_id"`
	ApplicationVersion           string                                        `json:"application_version"`
	MigrationVersion             int64                                         `json:"migration_version"`
	ExpiresAt                    time.Time                                     `json:"expires_at"`
	CiphertextSHA256             string                                        `json:"ciphertext_sha256"`
	PackageBindingSHA256         string                                        `json:"package_binding_sha256"`
	PreviewBindingSHA256         string                                        `json:"preview_binding_sha256"`
	TargetDatabaseOID            int64                                         `json:"target_database_oid"`
	CandidateDatabaseOID         int64                                         `json:"candidate_database_oid"`
	TargetDatabaseFingerprint    string                                        `json:"target_database_fingerprint"`
	CandidateDatabaseFingerprint string                                        `json:"candidate_database_fingerprint"`
	TargetCompleteMediaSHA256    string                                        `json:"target_complete_media_sha256"`
	CandidateCompleteMediaSHA256 string                                        `json:"candidate_complete_media_sha256"`
	ApplicationBoundarySHA256    string                                        `json:"application_boundary_sha256"`
	ProjectionConfiguration      *maintenancegate.ExternalProjectionBoundary   `json:"projection_configuration,omitempty"`
	Observation                  string                                        `json:"observation"`
	CandidateCustodian           string                                        `json:"candidate_custodian"`
	CandidateDatabaseOwner       string                                        `json:"candidate_database_owner"`
	CustodyTransferAvailable     bool                                          `json:"custody_transfer_available"`
	OnlineOverwriteAvailable     bool                                          `json:"online_overwrite_available"`
	StoppedRevalidationRequired  bool                                          `json:"stopped_revalidation_required"`
	UnselectedDataPreserved      bool                                          `json:"unselected_data_preserved"`
	PreservedTables              int                                           `json:"preserved_tables"`
	InventoryPreservation        *ProductCatalogInventoryPreservationResult    `json:"inventory_preservation,omitempty"`
	CustodyRules                 []string                                      `json:"custody_rules"`
	PreservedDomains             []string                                      `json:"preserved_domains"`
	ProjectionRequirements       []ExternalProjectionReconciliationRequirement `json:"projection_requirements"`
	InvalidationConditions       []string                                      `json:"invalidation_conditions"`
	Blockers                     []string                                      `json:"blockers"`
	ContractSHA256               string                                        `json:"contract_sha256"`
}

func validateFreshProductCatalogCandidateTaskForHandoffInspection(task *Task) error {
	if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
		return errors.New("stopped product handoff preparation requires its own continuous-window validation")
	}
	if task != nil && task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" {
		return errors.New("unfinished stopped product view drain invalidates the old preview")
	}
	if task == nil || !safeTaskID.MatchString(task.ID) || task.Kind != "upload" || task.State != "product_module_isolated_verified" || task.Manifest == nil || task.Manifest.Scope != productCatalogBackupScope || task.Journal != nil || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) || task.ProductCatalogRecovery == nil || task.ProductCatalogRecovery.OnlineOverwriteAvailable || !task.ProductCatalogRecovery.PreservedTargetOnlyRows || task.ProductCatalogRecovery.InventoryPreservation == nil || !task.ProductCatalogRecovery.InventoryPreservation.ExistingInventoryAndIdentityFactsVerified {
		return errors.New("a fresh isolated product candidate with verified preservation is required for handoff inspection")
	}
	binding := task.ProductCatalogPreviewBinding
	if binding == nil || binding.Version != 1 || binding.TargetMedia == nil || binding.CandidateMedia == nil || binding.RequiredImageCleanupProtection == nil || binding.Candidate.OID <= 0 || binding.Target.OID <= 0 || binding.Target.OID == binding.Candidate.OID || binding.Candidate.Owner != "local_backup_isolated_owner" || task.StageDatabase != "local_backup_stage_"+task.ID[:16] || task.StageFingerprint != binding.Candidate.Fingerprint || task.CiphertextSHA256 == "" || task.Manifest.PackageID == "" {
		return errors.New("product candidate handoff requires the original bound database, media, package and isolated owner identities")
	}
	return nil
}

// This validation is also used under the existing stopped rehearsal lease.
// All operations here read; it does not open application connections, rename,
// change owners, consume events, or contact Redis or an external provider.
func (manager *Manager) validateProductCatalogCandidateStillMatchesIsolatedPreview(ctx context.Context, task *Task) error {
	cfg, binding := manager.Configuration, task.ProductCatalogPreviewBinding
	configurationDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(cfg)
	if err != nil || binding == nil || configurationDigest != binding.ConfigurationSHA256 {
		return errors.New("product recovery configuration changed after preview")
	}
	if err := manager.verifyPreparedRecoveryFiles(task); err != nil {
		return err
	}
	if err := manager.validateProductCatalogBundledStaticImagesAlreadyMatchTarget(ctx, task); err != nil {
		return err
	}
	if err := validateProductCatalogCandidateCompleteMediaPreservation(binding.TargetMedia, binding.CandidateMedia, task.Manifest); err != nil {
		return err
	}
	current, err := manager.captureProductCatalogRecoveryPreviewBaseline(ctx)
	if err != nil {
		return err
	}
	if current.Target != binding.Target || !reflect.DeepEqual(current.TargetMedia, binding.TargetMedia) {
		return errors.New("product target changed after preview; rebuild the candidate under maintenance")
	}
	actualCandidate, err := cfg.captureProductCatalogRecoveryDatabaseBinding(ctx, task.StageDatabase)
	if err != nil || *actualCandidate != binding.Candidate {
		return errors.New("product candidate database identity, data or schema changed after preview")
	}
	protection, err := cfg.inspectProductCatalogRequiredImageCleanupProtectionInDatabase(ctx, task.StageDatabase)
	if err != nil {
		return err
	}
	if binding.RequiredImageCleanupProtection == nil || !reflect.DeepEqual(protection, binding.RequiredImageCleanupProtection) {
		return errors.New("product required image cleanup protection is missing or changed; rebuild the candidate")
	}
	actualMedia, err := manager.productCatalogCandidateMediaConfiguration(task).inventoryModuleRecoveryCompleteMedia(ctx)
	if err != nil || !reflect.DeepEqual(actualMedia, binding.CandidateMedia) {
		return errors.New("product candidate complete media changed after preview")
	}
	return nil
}

func (manager *Manager) InspectProductCatalogCandidateHandoff(ctx context.Context, id string) (*ProductCatalogCandidateHandoffInspection, error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if err := validateFreshProductCatalogCandidateTaskForHandoffInspection(task); err != nil {
		return nil, err
	}
	cfg := manager.Configuration
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	if fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory); err != nil || fenced {
		return nil, errors.New("product handoff inspection requires an unchanged candidate outside an active recovery window")
	}
	boundaryData, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, "application-boundary.json"))
	var boundary maintenancegate.ApplicationBoundary
	if err != nil || decodeStrictJSON(boundaryData, &boundary) != nil {
		return nil, errors.New("registered product runtime boundary is missing or invalid")
	}
	if err := manager.validateProductCatalogCandidateStillMatchesIsolatedPreview(ctx, task); err != nil {
		return nil, err
	}
	// A live check is observational. Detect changes during the longer candidate
	// scan, and require the stopped validation again for any future handoff.
	current, err := manager.captureProductCatalogRecoveryPreviewBaseline(ctx)
	if err != nil {
		return nil, err
	}
	if current.Target != task.ProductCatalogPreviewBinding.Target || !reflect.DeepEqual(current.TargetMedia, task.ProductCatalogPreviewBinding.TargetMedia) {
		return nil, errors.New("product target changed during handoff inspection; rebuild the candidate")
	}
	latestBoundary, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, "application-boundary.json"))
	if err != nil || string(latestBoundary) != string(boundaryData) {
		return nil, errors.New("registered runtime boundary changed during product handoff inspection")
	}
	if err := validateFreshProductCatalogCandidateTaskForHandoffInspection(task); err != nil {
		return nil, err
	}
	return manager.buildProductCatalogCandidateHandoffInspection(task, boundary)
}

func (manager *Manager) buildProductCatalogCandidateHandoffInspection(task *Task, boundary maintenancegate.ApplicationBoundary) (*ProductCatalogCandidateHandoffInspection, error) {
	if err := validateFreshProductCatalogCandidateTaskForHandoffInspection(task); err != nil {
		return nil, err
	}
	binding, cfg := task.ProductCatalogPreviewBinding, manager.Configuration
	packageDigest, err := productCatalogPairedRehearsalPackageBinding(task)
	if err != nil {
		return nil, err
	}
	previewDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(binding)
	if err != nil {
		return nil, err
	}
	report := &ProductCatalogCandidateHandoffInspection{
		Version: 1, Scope: productCatalogBackupScope, SiteID: cfg.SiteID, Environment: cfg.Environment, TaskID: task.ID, PackageID: task.Manifest.PackageID,
		ApplicationVersion: cfg.ApplicationVersion, MigrationVersion: task.Manifest.MigrationVersion, ExpiresAt: task.ExpiresAt, CiphertextSHA256: task.CiphertextSHA256,
		PackageBindingSHA256: packageDigest, PreviewBindingSHA256: previewDigest,
		TargetDatabaseOID: binding.Target.OID, CandidateDatabaseOID: binding.Candidate.OID,
		TargetDatabaseFingerprint: binding.Target.Fingerprint, CandidateDatabaseFingerprint: binding.Candidate.Fingerprint,
		TargetCompleteMediaSHA256: maintenancegate.FingerprintExternalProjectionConfiguration(binding.TargetMedia), CandidateCompleteMediaSHA256: maintenancegate.FingerprintExternalProjectionConfiguration(binding.CandidateMedia),
		ApplicationBoundarySHA256: maintenancegate.FingerprintExternalProjectionConfiguration(boundary), ProjectionConfiguration: boundary.ExternalProjections,
		Observation: "isolated_preview_matches_at_inspection_stopped_revalidation_still_required", CandidateCustodian: "maintenance_executor", CandidateDatabaseOwner: binding.Candidate.Owner,
		StoppedRevalidationRequired: true, UnselectedDataPreserved: true, PreservedTables: task.ProductCatalogRecovery.PreservedTables, InventoryPreservation: task.ProductCatalogRecovery.InventoryPreservation,
		CustodyRules: []string{
			"隔离库及完整附件仍由维护器保管，本次检查不改变数据库所有者、连接权限或任务状态。",
			"正式接收前须停写并重新核对任务、包、当前库、候选库及三媒体根；旧检查不能批准交接。",
			"切换后旧库和附件由维护器保留，须绑定切换后的实际身份核对；核对失败保持隔离并成对回退。",
			"候选过期或拒绝须按原任务、库 OID、所有者和媒体归属清理；身份不符保留现场，不能按名称删除。",
		},
		PreservedDomains:       []string{"current_variant_stock_and_existing_product_views", "orders_payments_refunds_and_shipping", "customers_customer_service_messages_and_read_positions", "current_outbox_jobs_delivery_receipts_and_global_idempotencies", "shared_media_registrations_and_unselected_media_bytes"},
		InvalidationConditions: []string{"task_package_scope_or_expiration_changed", "target_or_candidate_database_oid_owner_data_sequence_or_schema_changed", "target_or_candidate_complete_media_or_bundled_static_bytes_changed", "shared_registrations_cleanup_receipts_or_guard_versions_changed", "business_build_recovery_configuration_or_registered_projection_configuration_changed", "candidate_disposed_or_any_recovery_journal_created", "future_switch_requires_new_post_switch_projection_evidence_not_this_report"},
		Blockers:               []string{"production_product_candidate_ownership_handoff_not_implemented", "product_post_switch_external_projection_reconciliation_not_executed", "production_product_restore_eligibility_and_final_acceptance_not_implemented"},
	}
	viewAction, viewDescription := "product_view_count_configuration_unknown_keep_all_redis_state", "商品浏览数持久化配置未知，禁止把 Redis 计数当缓存删除。"
	if (boundary.Protocol != 2 && boundary.Protocol != 3) || boundary.ValidateRegisteredRecoveryProjectionProtocol() != nil || boundary.ExternalProjections == nil || boundary.ExternalProjections.ProductCatalog == nil {
		report.Blockers = append(report.Blockers, "product_projection_configuration_unknown_reregister_current_business_build")
	} else if boundary.ExternalProjections.ProductCatalog.ViewCountMode == "redis_delta_buffer" {
		viewAction = "preserve_active_and_processing_view_deltas_and_prove_no_replay_before_release"
		if boundary.ExternalProjections.ProductCatalog.Version == 2 {
			viewDescription = "已登记固定批次及 SQL 事务幂等收据；仍须证明 Redis 归属、停止全部读写和定时任务，核对 active/processing/批次身份及 SQL 收据，再排空增量。落库会改变当前 SQL，须据此重建商品候选。旧 processing 无身份时人工核对，禁止猜测补批次号、清计数或重放。当前报告没有访问 Redis，不能批准交接。"
			report.Blockers = append(report.Blockers, "buffered_product_view_delta_stopped_handoff_and_target_verification_not_implemented")
		} else {
			viewDescription = "停写前独立核对 active 和 processing 两份浏览增量及 SQL 落库边界；商品恢复保留当前 SQL 浏览数，禁止删计数、重复落库或重放。旧缓冲器无恢复幂等收据，本检查不证明在途落库已经安全。"
			report.Blockers = append(report.Blockers, "buffered_product_view_delta_drain_and_inflight_idempotency_contract_not_implemented")
		}
	} else {
		viewAction = "preserve_current_sql_product_views_without_redis_delta_rebuild"
		viewDescription = "业务明确直接写入 SQL 浏览数；已有商品继续保留当前值，核对商品身份，不从商品包重置或累计。"
	}
	report.ProjectionRequirements = productCatalogPartialProjectionReconciliationRequirements(viewAction, viewDescription)
	// The timestamp is informational. Stable recorded identities give the same
	// plan digest; none of these fields are accepted by whole-storefront release.
	report.ContractSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(report)
	if err != nil {
		return nil, err
	}
	report.CapturedAt = time.Now().UTC()
	return report, nil
}

func productCatalogPartialProjectionReconciliationRequirements(viewAction, viewDescription string) []ExternalProjectionReconciliationRequirement {
	return []ExternalProjectionReconciliationRequirement{
		{Projection: "product_sql_search_and_visibility", Action: "verify_final_sql_product_search_translations_visibility_and_stock", Description: "切换后按最终 SQL 核对 ID/slug 查询、类目、翻译、搜索、可见性和当前库存；整站或旧候选报告不可代用。"},
		{Projection: "product_detail_cache_and_read_leases", Action: "prove_redis_ownership_then_invalidate_old_and_final_id_slug_cache_with_readers_stopped", Description: "先证明 Redis 归属，再按当前与最终 ID/slug/locale 的并集失效商品详情缓存，核对读租约和进程缓存；禁止按 product:* 清理，以免删掉浏览增量，也不允许 FLUSHDB/FLUSHALL。"},
		{Projection: "product_view_count_deltas", Action: viewAction, Description: viewDescription},
		{Projection: "preserved_events_jobs_and_customer_service", Action: "preserve_current_events_idempotencies_jobs_and_customer_service_streams_without_replay", Description: "保留当前已处理/待处理事件、任务、租约、幂等和通知凭据；商品恢复不清客服 Stream/去重键，不重放订单、退款、物流或注册推荐奖励。"},
		{Projection: "storefront_http_media_and_edge_cache", Action: "verify_final_product_http_and_scoped_media_edge_browser_refresh", Description: "切换后独立核对商城商品页、旧/新地址、媒体字节与公开/私有权限、静态资源和本商城边缘/浏览器缓存；本检查不访问 CDN 或浏览器，不代表已核对。"},
	}
}
