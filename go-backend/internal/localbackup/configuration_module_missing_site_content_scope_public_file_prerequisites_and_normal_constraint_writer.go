package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/lib/pq"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
)

const configurationModuleSiteContentRestorePolicy = "restore_missing_business_configuration_including_shipping_reference_packaging_and_customs_preserve_existing_rows_and_registered_files_v6"

type ConfigurationModuleSiteContentRestoreEligibility struct {
	Policy         string   `json:"policy"`
	Eligible       bool     `json:"eligible"`
	WritableTables []string `json:"writable_tables"`
	Blockers       []string `json:"blockers"`
}

// This entry restores only the explicitly approved missing business configuration.
// Exporting other configuration tables never grants write permission. Same-ID
// differences, credentials, settings and current execution controls are refused.
func configurationModuleMissingBusinessConfigurationWritableTableNames() []string {
	tables := append(strings.Fields("categories posts post_categories galleries gallery_images gallery_product_links faq_pages faqs visual_showcase_items storefront_redirect_rules"), configurationModuleMissingFitmentSpokeAndTireCatalogWritableTableNames()...)
	tables = append(tables, configurationModuleMissingSelectionFlowAndQuestionnaireWritableTableNames()...)
	tables = append(tables, configurationModuleMissingMarketCurrencyAndTaxWritableTableNames()...)
	tables = append(tables, configurationModuleMissingEmailTemplateAndMemberLevelWritableTableNames()...)
	return append(tables, configurationModuleMissingShippingReferencePackagingAndCustomsWritableTableNames()...)
}

func configurationModuleHasMissingSiteContentWriteContract(snapshot selectedBusinessModuleCompleteSnapshot, table string) bool {
	if !reflect.DeepEqual(snapshot.Modules, []string{"configuration"}) {
		return false
	}
	for _, name := range configurationModuleMissingBusinessConfigurationWritableTableNames() {
		if name == table {
			return true
		}
	}
	return false
}

func validateConfigurationModuleMissingSiteContentInsertContract(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, table string, observation *moduleRecoveryDatabaseObservation) error {
	if !configurationModuleHasMissingSiteContentWriteContract(snapshot, table) {
		return errors.New("配置恢复首期未授权写入该表：" + table)
	}
	return validateBusinessModuleInsertOnlyTableContractWithUpdateOnlyGuardsPreserved(ctx, tx, table, observation)
}

// File restoration remains the media module's responsibility in this entry.
// Inspect real files, including company static assets, instead of silently
// accepting an absent file or installing every file from the larger config pack.
func (cfg Configuration) verifyConfigurationModuleReferencedFilesAlreadyMatchPackage(manifest *Manifest) error {
	for _, entry := range manifest.Files {
		if entry.Path == selectedBusinessModulesRowsFilename {
			continue
		}
		filename, err := cfg.resolveSelectedBusinessModuleConfinedFileSource(entry.Path)
		if err != nil {
			return errors.New("配置包引用文件无法定位，请先恢复图片与附件")
		}
		actual, err := fileIdentity(filename, entry.Path)
		if err != nil || actual != entry {
			return errors.New("配置包引用文件缺失或字节不同，请先恢复图片与附件并重新演练")
		}
	}
	return nil
}

func (cfg Configuration) inspectConfigurationModuleMissingSiteContentRestoreEligibility(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, missing map[string][]json.RawMessage, manifest *Manifest, preview *SelectedBusinessModuleIsolatedRestorePreview) (*ConfigurationModuleSiteContentRestoreEligibility, error) {
	result := &ConfigurationModuleSiteContentRestoreEligibility{Policy: configurationModuleSiteContentRestorePolicy, Eligible: true, WritableTables: configurationModuleMissingBusinessConfigurationWritableTableNames(), Blockers: []string{}}
	record := func(reason string) {
		for _, existing := range result.Blockers {
			if existing == reason {
				return
			}
		}
		result.Eligible = false
		result.Blockers = append(result.Blockers, reason)
	}
	if !reflect.DeepEqual(snapshot.Modules, []string{"configuration"}) || manifest == nil || !reflect.DeepEqual(manifest.SelectedModules, snapshot.Modules) {
		record("需要单独商城业务配置包")
		return result, nil
	}
	if preview != nil {
		for _, table := range preview.Tables {
			if table.ProtectedNewRows > 0 {
				record("本期不支持补回该配置或媒体登记（" + table.Table + "）")
			}
			if table.ConflictingRows > 0 {
				record("原包与当前配置有同 ID 内容差异（" + table.Table + "）")
			}
		}
	}
	for table, rows := range missing {
		if len(rows) > 0 && !configurationModuleHasMissingSiteContentWriteContract(snapshot, table) {
			record("本期只补回网站内容、目录、选购问卷、市场/汇率/税率、邮件模板/会员等级及物流基础资料/包装/报关，其他配置或媒体登记缺失须另行恢复（" + table + "）")
		}
	}
	if err := cfg.verifyConfigurationModuleReferencedFilesAlreadyMatchPackage(manifest); err != nil {
		record(err.Error())
	}
	if err := validateConfigurationModuleMissingFitmentCatalogDomainRows(missing); err != nil {
		record(err.Error())
	}
	if err := validateConfigurationModuleMissingSelectionConfigurationKeyRows(missing); err != nil {
		record(err.Error())
	}
	if err := validateConfigurationModuleMissingMarketCurrencyAndTaxDomainRows(missing); err != nil {
		record(err.Error())
	}
	if err := validateConfigurationModuleMissingEmailTemplateAndMemberLevelDomainRows(missing); err != nil {
		record(err.Error())
	}
	if err := validateConfigurationModuleMissingShippingReferencePackagingAndCustomsDomainRows(missing); err != nil {
		record(err.Error())
	}
	keys, assetIDs := map[string]bool{}, map[string]bool{}
	for _, entry := range manifest.Files {
		if entry.Path == selectedBusinessModulesRowsFilename || strings.HasPrefix(entry.Path, companyStaticPackagePrefix) {
			continue
		}
		root, key, err := transactionalBusinessModuleRestoreMediaPath(entry.Path)
		if err != nil || root == "uploads_private" || storage.IsPrivateObjectKey(key) {
			record("网站内容恢复不能引入私有附件")
			continue
		}
		keys[key] = true
	}
	publicParents := map[string]bool{}
	parents, derivatives := map[string]productCatalogMediaObjectRegistration{}, map[string]productCatalogMediaObjectRegistration{}
	for _, table := range []string{"media_assets", "media_asset_derivatives"} {
		rows, err := readProductCatalogRows(ctx, tx, table, "true")
		if err != nil {
			return nil, err
		}
		if len(rows) > businessModuleSnapshotMaximumRows {
			return nil, errors.New("配置媒体登记检查超过容量限制")
		}
		for _, row := range rows {
			field := func(column string) string {
				value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
				return value
			}
			registration := productCatalogMediaObjectRegistration{ID: field("id"), ParentID: field("media_asset_id"), Key: field("storage_key"), URL: field("url"), DeletedAt: field("deleted_at")}
			if registration.Key == "" {
				_, registration.Key = cfg.resolveLocalMediaReference(registration.URL)
			}
			if table == "media_assets" {
				publicParents[registration.ID] = field("deleted_at") == "" && field("customer_service_derivative_write_token") == "" && field("customer_service_derivative_write_expires_at") == "" && media.CanMediaAssetBeReusedByPublicBusinessContent(&media.MediaAsset{StorageKey: registration.Key, URL: registration.URL, Status: field("status"), Visibility: field("visibility"), LibraryPartition: media.MediaAssetLibraryPartition(field("library_partition")), UploaderActorType: media.MediaUploadActorType(field("uploader_actor_type")), UploadOrigin: media.MediaAssetUploadOrigin(field("upload_origin")), CustomerServiceMediaPurpose: field("customer_service_media_purpose")})
				parents[registration.ID] = registration
			} else {
				derivatives[registration.ID] = registration
			}
			if keys[registration.Key] {
				if table == "media_assets" && !media.CanMediaAssetBeReusedByPublicBusinessContent(&media.MediaAsset{StorageKey: registration.Key, URL: registration.URL, Status: field("status"), Visibility: field("visibility"), LibraryPartition: media.MediaAssetLibraryPartition(field("library_partition")), UploaderActorType: media.MediaUploadActorType(field("uploader_actor_type")), UploadOrigin: media.MediaAssetUploadOrigin(field("upload_origin")), CustomerServiceMediaPurpose: field("customer_service_media_purpose")}) {
					record("配置图片不能复用客户、访客或会话所有的素材")
				}
				if table == "media_assets" {
					assetIDs[registration.ID] = true
				}
				if field("deleted_at") != "" || field("customer_service_derivative_write_token") != "" || field("customer_service_derivative_write_expires_at") != "" {
					record("配置图片仍有退役登记或派生写入租约")
				}
			}
		}
	}
	for _, derivative := range derivatives {
		if keys[derivative.Key] {
			assetIDs[derivative.ParentID] = true
			if !publicParents[derivative.ParentID] {
				record("配置派生图片缺少可公开复用的有效原件")
			}
		}
	}
	events, err := readProductCatalogRows(ctx, tx, "outbox_events", "event_type ILIKE '%cleanup%'")
	if err != nil {
		return nil, err
	}
	if len(events) > 20000 {
		return nil, errors.New("配置图片清理事件检查超过容量限制")
	}
	for _, row := range events {
		if err := cfg.validateProductCatalogRetainedCleanupReceipt(row, keys, assetIDs, parents, derivatives); err != nil {
			record("当前清理事件可能影响配置图片，须先核对原事件；恢复不会重放或重置事件")
		}
	}
	sort.Strings(result.Blockers)
	return result, nil
}

// One INSERT per table also handles same-table parent/translation references
// with normal statement-level FK enforcement. Existing rows, all other tables
// and unrelated sequences are verified before the caller can commit.
func restoreConfigurationModuleMissingBusinessConfigurationRowsWithNormalConstraintsAndVerifiedPreservation(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, missing map[string][]json.RawMessage) (int, error) {
	for table, rows := range missing {
		if len(rows) > 0 {
			if err := validateConfigurationModuleMissingSiteContentInsertContract(ctx, tx, snapshot, table, observation); err != nil {
				return 0, err
			}
		}
	}
	if err := validateConfigurationModuleMissingFitmentCatalogDomainRows(missing); err != nil {
		return 0, err
	}
	var role string
	if err := tx.QueryRowContext(ctx, `SHOW session_replication_role`).Scan(&role); err != nil {
		return 0, err
	}
	if err := validateBusinessModuleNormalForeignKeyEnforcementBeforeWriting(ctx, tx, snapshot); err != nil {
		return 0, err
	}
	if role != "origin" {
		return 0, errors.New("网站内容恢复要求正常数据库约束")
	}
	sequences, err := readBusinessModuleRestoreSequenceStates(ctx, tx)
	if err != nil {
		return 0, err
	}
	before := map[string]BusinessModuleRestoreTablePreservation{}
	for table := range observation.Tables {
		fingerprint, err := fingerprintBusinessModuleRetainedTargetTable(ctx, tx, table, observation.Shapes[table], missing[table])
		if err != nil {
			return 0, err
		}
		before[table] = fingerprint
	}
	restored := 0
	for _, table := range configurationModuleMissingBusinessConfigurationWritableTableNames() {
		if len(missing[table]) == 0 {
			continue
		}
		columns := []string{}
		for column := range observation.Columns[table] {
			columns = append(columns, pq.QuoteIdentifier(column))
		}
		sort.Strings(columns)
		list := strings.Join(columns, ",")
		rows, err := json.Marshal(missing[table])
		if err != nil {
			return 0, err
		}
		result, err := tx.ExecContext(ctx, `INSERT INTO public.`+pq.QuoteIdentifier(table)+` (`+list+`) SELECT `+list+` FROM jsonb_populate_recordset(NULL::public.`+pq.QuoteIdentifier(table)+`,$1::jsonb)`, string(rows)) // #nosec G202 -- Table and column identifiers use pq.QuoteIdentifier from the validated schema; complete row values use bound JSON parameters.
		if err != nil {
			return 0, fmt.Errorf("网站内容恢复违反数据库身份或业务约束：%s", table)
		}
		count, err := result.RowsAffected()
		if err != nil || count != int64(len(missing[table])) {
			return 0, errors.New("网站内容写入数量与备份不一致")
		}
		restored += int(count)
	}
	if _, err := tx.ExecContext(ctx, `SET CONSTRAINTS ALL IMMEDIATE`); err != nil {
		return 0, err
	}
	if err := validateProductCatalogAffectedForeignKeys(ctx, tx, productCatalogModuleSnapshot{Tables: snapshot.Tables}); err != nil {
		return 0, err
	}
	if err := validateConfigurationModuleRestoredFitmentCatalogBusinessRelationships(ctx, tx, missing); err != nil {
		return 0, err
	}
	if err := validateConfigurationModuleRestoredSelectionFlowsAndQuestionnairesAgainstPublishedBusinessRules(ctx, tx, missing); err != nil {
		return 0, err
	}
	if err := validateConfigurationModuleRestoredMarketCurrencyAndTaxRelationships(ctx, tx, snapshot, missing); err != nil {
		return 0, err
	}
	if err := validateConfigurationModuleRestoredEmailTemplateHistoriesAndMemberLevelRanges(ctx, tx, missing); err != nil {
		return 0, err
	}
	if err := validateConfigurationModuleRestoredShippingPackagingProductVariantRelationships(ctx, tx, missing); err != nil {
		return 0, err
	}
	for table, contents := range snapshot.Tables {
		for _, row := range contents.Rows {
			var same bool
			if err := tx.QueryRowContext(ctx, `SELECT to_jsonb(record)=$1::jsonb FROM public.`+pq.QuoteIdentifier(table)+` record WHERE `+businessModuleCompleteRowTypedPrimaryKeyPredicate(table, observation.PrimaryKeys[table]), string(row)).Scan(&same); err != nil || !same {
				return 0, fmt.Errorf("配置完整行与原包不一致：%s", table)
			}
		}
	}
	for table, expected := range before {
		actual, err := fingerprintBusinessModuleRetainedTargetTable(ctx, tx, table, observation.Shapes[table], missing[table])
		if err != nil {
			return 0, err
		}
		if actual != expected {
			return 0, fmt.Errorf("配置恢复改变了当前保留数据：%s", table)
		}
	}
	if err := advanceBusinessModuleRestoreOwnedSequencesTransactionallyAndVerifyPreservation(ctx, tx, sequences, missing); err != nil {
		return 0, err
	}
	return restored, nil
}
