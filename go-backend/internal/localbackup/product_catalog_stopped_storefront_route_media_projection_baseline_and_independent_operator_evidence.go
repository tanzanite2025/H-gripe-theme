package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"commerce-platform/internal/domain/seo"
	"commerce-platform/internal/pkg/maintenancegate"
)

type ProductCatalogStoppedStorefrontRouteExpectation struct {
	ProductID    string `json:"product_id"`
	Slug         string `json:"slug"`
	Locale       string `json:"locale"`
	Status       string `json:"status"`
	SoftDeleted  bool   `json:"soft_deleted"`
	PagePath     string `json:"page_path"`
	PublicActive bool   `json:"public_active"`
}

// This snapshot supplies SQL/filesystem expectations for independent checks.
// It never sends HTTP requests, purges a provider or opens business service.
type ProductCatalogStoppedStorefrontProjectionInspection struct {
	Version                          int                                               `json:"version"`
	Scope                            string                                            `json:"scope"`
	SiteID                           string                                            `json:"site_id"`
	Environment                      string                                            `json:"environment"`
	TaskID                           string                                            `json:"task_id"`
	WindowID                         string                                            `json:"window_id"`
	PackageID                        string                                            `json:"package_id"`
	SwitchJournalSHA256              string                                            `json:"switch_journal_sha256"`
	PostSwitchContractSHA256         string                                            `json:"post_switch_contract_sha256"`
	DetailCacheProgressSHA256        string                                            `json:"detail_cache_progress_sha256"`
	DatabaseOID                      int64                                             `json:"database_oid"`
	DatabaseFingerprint              string                                            `json:"database_fingerprint"`
	CompleteMediaSHA256              string                                            `json:"complete_media_sha256"`
	ApplicationBoundarySHA256        string                                            `json:"application_boundary_sha256"`
	ConfigurationSHA256              string                                            `json:"configuration_sha256"`
	MediaDeliveryConfigurationSHA256 string                                            `json:"media_delivery_configuration_sha256"`
	PublicOrigins                    []string                                          `json:"public_origins"`
	OriginalRoutes                   []ProductCatalogStoppedStorefrontRouteExpectation `json:"original_routes"`
	FinalRoutes                      []ProductCatalogStoppedStorefrontRouteExpectation `json:"final_routes"`
	PackageImageFiles                []FileIdentity                                    `json:"package_image_files"`
	Requirements                     []ExternalProjectionReconciliationRequirement     `json:"requirements"`
	Verification                     string                                            `json:"verification"`
	ApplicationRemainsStopped        bool                                              `json:"application_remains_stopped"`
	OnlineOverwriteAvailable         bool                                              `json:"online_overwrite_available"`
	ContractSHA256                   string                                            `json:"contract_sha256"`
	CapturedAt                       time.Time                                         `json:"captured_at"`
}

type ProductCatalogStoppedProjectionEvidenceRequest struct {
	Confirmation         string                                        `json:"confirmation"`
	ContractSHA256       string                                        `json:"contract_sha256"`
	ExpectedRecordSHA256 string                                        `json:"expected_record_sha256"`
	Attestations         []ExternalProjectionReconciliationAttestation `json:"attestations"`
}

// Operator assertions and hashes of records retained on the operator's computer,
// not automatically verified external state or authority to resume service.
type ProductCatalogStoppedProjectionOperatorEvidenceRecord struct {
	Version                   int                                           `json:"version"`
	Scope                     string                                        `json:"scope"`
	TaskID                    string                                        `json:"task_id"`
	WindowID                  string                                        `json:"window_id"`
	ContractSHA256            string                                        `json:"contract_sha256"`
	DetailCacheProgressSHA256 string                                        `json:"detail_cache_progress_sha256"`
	Operator                  string                                        `json:"operator"`
	RecordedAt                time.Time                                     `json:"recorded_at"`
	Verification              string                                        `json:"verification"`
	Attestations              []ExternalProjectionReconciliationAttestation `json:"attestations"`
	SHA256                    string                                        `json:"sha256"`
}

func stoppedProductStorefrontIndependentProjectionRequirements() []ExternalProjectionReconciliationRequirement {
	return []ExternalProjectionReconciliationRequirement{
		{Projection: "product_storefront_http_routes", Action: "verify_final_id_slug_locale_search_visibility_and_stock_against_switched_sql", Description: "使用独立且不启动业务写入/任务的受控只读环境，对照原/最终 SQL 商品清单核对 ID、旧/新 slug、locale、搜索、不可见/已删商品、价格和当前库存；维护页、503 或仅能访问页面不能算通过。现有路由索引/站点地图也须与最终商品核对，保留非商品条目。"},
		{Projection: "product_public_private_media_delivery", Action: "verify_public_image_bytes_and_private_attachment_authorization_against_complete_media", Description: "按包内真实文件大小/SHA-256 核对公开图片与静态图片，核对原/新地址、私有附件匿名拒绝及授权读取身份；不能拿本地文件存在或数据库摘要代替 HTTP 权限/字节核对。保留其他模块媒体。"},
		{Projection: "product_html_edge_and_browser_caches", Action: "verify_origin_html_no_store_scoped_edge_configuration_and_fresh_browser", Description: "独立核对实际 Nuxt 构建、商品/列表 HTML no-store、HTML 清理配置、CDN/反向代理规则及目的地、旧缓存和全新浏览器结果；媒体地址登记摘要不代表这些配置已登记。仅按已证明归属的当前商城处理边缘缓存，不清其他站点，不把当前仓库默认规则视为实际部署。"},
	}
}

func (manager *Manager) InspectStoppedProductStorefrontProjectionBaseline(ctx context.Context, id string) (*ProductCatalogStoppedStorefrontProjectionInspection, error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("another maintenance operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	return manager.inspectStoppedProductStorefrontProjectionBaselineWithHeldLease(ctx, task)
}

func (manager *Manager) inspectStoppedProductStorefrontProjectionBaselineWithHeldLease(ctx context.Context, task *Task) (*ProductCatalogStoppedStorefrontProjectionInspection, error) {
	if task.State != "product_module_switched_still_isolated" || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) || task.ProductCatalogStoppedDetailCacheInvalidation == nil || task.ProductCatalogStoppedDetailCacheInvalidation.Phase != "verified_absent" {
		return nil, errors.New("an unexpired switched product result with verified detail-cache absence is required")
	}
	before, err := manager.inspectStoppedProductCatalogPostSwitchWithHeldLease(ctx, task)
	if err != nil {
		return nil, err
	}
	if err := manager.verifyStoppedProductProjectionDetailCacheStillAbsent(ctx, task, before); err != nil {
		return nil, err
	}
	boundaryBytes, err := readCustomerServicePairedRecoveryBoundedRegularFile(filepath.Join(manager.Configuration.MaintenanceDirectory, "application-boundary.json"), 1<<20)
	var boundary maintenancegate.ApplicationBoundary
	if err != nil || decodeStrictJSON(boundaryBytes, &boundary) != nil || boundary.ExternalProjections == nil || boundary.ValidateRegisteredRecoveryProjectionProtocol() != nil {
		return nil, errors.New("registered product projection configuration cannot be read")
	}
	originalRoutes, err := manager.readBoundStoppedProductStorefrontRoutes(ctx, task.Journal.RollbackDatabase, task.ProductCatalogPreviewBinding.Target)
	if err != nil {
		return nil, err
	}
	finalRoutes, err := manager.readBoundStoppedProductStorefrontRoutes(ctx, manager.Configuration.DatabaseName, task.ProductCatalogPreviewBinding.Candidate)
	if err != nil {
		return nil, err
	}
	images := []FileIdentity{}
	for _, file := range task.Manifest.Files {
		if strings.HasPrefix(file.Path, "media/") || strings.HasPrefix(file.Path, "product-static/") {
			images = append(images, file)
		}
	}
	if len(images) > 10000 {
		return nil, errors.New("product image projection exceeds the supported 10000-file inspection bound")
	}
	// Recheck full SQL/media and cache after temporarily opening maintenance-only
	// connections. The stable independent inspection contract must be unchanged.
	after, err := manager.inspectStoppedProductCatalogPostSwitchWithHeldLease(ctx, task)
	if err != nil || after.ContractSHA256 != before.ContractSHA256 {
		return nil, errors.New("product switch changed during storefront projection inspection")
	}
	if err := manager.verifyStoppedProductProjectionDetailCacheStillAbsent(ctx, task, after); err != nil {
		return nil, err
	}
	report := &ProductCatalogStoppedStorefrontProjectionInspection{Version: 1, Scope: productCatalogBackupScope, SiteID: manager.Configuration.SiteID, Environment: manager.Configuration.Environment, TaskID: task.ID, WindowID: after.WindowID, PackageID: task.Manifest.PackageID, SwitchJournalSHA256: after.SwitchJournalSHA256, PostSwitchContractSHA256: after.ContractSHA256, DetailCacheProgressSHA256: task.ProductCatalogStoppedDetailCacheInvalidation.SHA256, DatabaseOID: after.DatabaseOID, DatabaseFingerprint: after.DatabaseFingerprint, CompleteMediaSHA256: after.CompleteMediaSHA256, ApplicationBoundarySHA256: after.ApplicationBoundarySHA256, ConfigurationSHA256: after.ConfigurationSHA256, MediaDeliveryConfigurationSHA256: boundary.ExternalProjections.MediaDeliveryConfigurationSHA256, PublicOrigins: append([]string{}, manager.Configuration.PublicOrigins...), OriginalRoutes: originalRoutes, FinalRoutes: finalRoutes, PackageImageFiles: images, Requirements: stoppedProductStorefrontIndependentProjectionRequirements(), Verification: "sql_media_and_cache_baseline_verified_independent_http_edge_browser_operator_records_required", ApplicationRemainsStopped: true}
	encoded, err := json.Marshal(report)
	if err != nil || len(encoded) > 4<<20 {
		return nil, errors.New("product storefront projection exceeds its supported 4 MiB report bound")
	}
	report.ContractSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(report)
	if err != nil {
		return nil, err
	}
	report.CapturedAt = time.Now().UTC()
	return report, nil
}

func (manager *Manager) verifyStoppedProductProjectionDetailCacheStillAbsent(ctx context.Context, task *Task, inspection *ProductCatalogStoppedPostSwitchInspection) error {
	if err := validateStoppedProductDetailCacheInvalidationBinding(task, inspection); err != nil {
		return err
	}
	client, ownership, err := manager.connectAndVerifyRuntimeOwnedProductViewRedis(ctx)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(client.Close)
	if err := manager.verifyStoppedProductDetailCacheRedisBoundary(ctx, task, client, ownership); err != nil {
		return err
	}
	keys, err := inspectExactStoppedProductDetailCacheAndReadLeaseKeys(ctx, client)
	if err != nil || len(keys) != 0 {
		return errors.New("product detail cache or read lease reappeared; recheck the stopped runtime before recording projection evidence")
	}
	return nil
}

func (manager *Manager) readBoundStoppedProductStorefrontRoutes(ctx context.Context, name string, expected productCatalogRecoveryDatabaseBinding) ([]ProductCatalogStoppedStorefrontRouteExpectation, error) {
	routes := []ProductCatalogStoppedStorefrontRouteExpectation{}
	err := manager.withBoundStoppedProductReadOnlyTransaction(ctx, name, expected, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `SELECT id::text,COALESCE(slug,''),COALESCE(locale,''),COALESCE(status,''),deleted_at IS NOT NULL FROM public.products ORDER BY id LIMIT 10001`)
		if err != nil {
			return err
		}
		defer func(closeResource func() error) { _ = closeResource() }(rows.Close)
		for rows.Next() {
			var route ProductCatalogStoppedStorefrontRouteExpectation
			if err := rows.Scan(&route.ProductID, &route.Slug, &route.Locale, &route.Status, &route.SoftDeleted); err != nil {
				return err
			}
			route.PagePath = seo.BuildProductRoute(route.Locale, route.Slug).Path
			route.PublicActive = route.Status == "active" && !route.SoftDeleted && route.PagePath != ""
			routes = append(routes, route)
			if len(routes) > 10000 {
				return errors.New("product route inspection exceeds 10000 rows per database; no truncated baseline accepted")
			}
		}
		return rows.Err()
	})
	return routes, err
}

func validateStoppedProductProjectionOperatorAttestations(report *ProductCatalogStoppedStorefrontProjectionInspection, request ProductCatalogStoppedProjectionEvidenceRequest) error {
	if report == nil || report.Scope != productCatalogBackupScope || report.Version != 1 || !report.ApplicationRemainsStopped || report.OnlineOverwriteAvailable || request.ContractSHA256 != report.ContractSHA256 || !isValidExternalProjectionEvidenceSHA256(request.ContractSHA256) || len(report.Requirements) != 3 || len(request.Attestations) != 3 {
		return errors.New("all three independent operator records must bind the latest stopped product projection snapshot")
	}
	seen := map[string]bool{}
	for _, attestation := range request.Attestations {
		matched := false
		for _, requirement := range report.Requirements {
			if attestation.Projection == requirement.Projection && attestation.Action == requirement.Action {
				matched = true
			}
		}
		if !matched || seen[attestation.Projection] || attestation.EvidenceReference != strings.TrimSpace(attestation.EvidenceReference) || len(attestation.EvidenceReference) < 4 || len(attestation.EvidenceReference) > 2000 || !isValidExternalProjectionEvidenceSHA256(attestation.EvidenceSHA256) {
			return errors.New("product projection action, unique identity, independent reference and record SHA-256 are required")
		}
		seen[attestation.Projection] = true
	}
	return nil
}

func (manager *Manager) SaveStoppedProductProjectionIndependentOperatorEvidence(ctx context.Context, id, actor string, request ProductCatalogStoppedProjectionEvidenceRequest) (*ProductCatalogStoppedProjectionOperatorEvidenceRecord, error) {
	if request.Confirmation != "RECORD STOPPED PRODUCT PROJECTIONS "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id || strings.TrimSpace(actor) == "" {
		return nil, errors.New("explicit stopped product projection record confirmation is required")
	}
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("another maintenance operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if !taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) {
		return nil, errors.New("a stopped product switch is required before projection evidence")
	}
	previous := task.ProductCatalogStoppedProjectionEvidence
	if previous == nil {
		if request.ExpectedRecordSHA256 != "" {
			return nil, errors.New("product projection record changed; refresh before saving")
		}
	} else {
		snapshotCopy := *previous
		snapshotCopy.SHA256 = ""
		digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(snapshotCopy)
		if err != nil || digest != previous.SHA256 || request.ExpectedRecordSHA256 != previous.SHA256 || previous.Version != 1 || previous.Scope != productCatalogBackupScope || previous.TaskID != task.ID || previous.WindowID != task.Journal.ProductCatalogStoppedSwitch.WindowID {
			return nil, errors.New("product projection record identity, digest or expected revision changed")
		}
	}
	report, err := manager.inspectStoppedProductStorefrontProjectionBaselineWithHeldLease(ctx, task)
	if err != nil {
		return nil, err
	}
	if err := validateStoppedProductProjectionOperatorAttestations(report, request); err != nil {
		return nil, err
	}
	record := &ProductCatalogStoppedProjectionOperatorEvidenceRecord{Version: 1, Scope: productCatalogBackupScope, TaskID: task.ID, WindowID: report.WindowID, ContractSHA256: report.ContractSHA256, DetailCacheProgressSHA256: report.DetailCacheProgressSHA256, Operator: actor, RecordedAt: time.Now().UTC(), Verification: "operator_attestation_only_external_results_not_automatically_verified", Attestations: append([]ExternalProjectionReconciliationAttestation{}, request.Attestations...)}
	record.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(record)
	if err != nil {
		return nil, err
	}
	task.ProductCatalogStoppedProjectionEvidence = record
	if err := manager.saveTask(task); err != nil {
		return nil, err
	}
	return record, nil
}
