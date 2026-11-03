package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	fitmenthttp "commerce-platform/internal/api/v1/fitmentcatalog"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestConfigurationFitmentCatalogWriteScopeStaysInsideConfigurationExportWithoutGrantingOtherModules(t *testing.T) {
	configuration := selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}}
	exported := selectedBusinessModuleTableSelectors(configuration.Modules)
	require.Len(t, exported, 79)
	allowed := configurationModuleMissingBusinessConfigurationWritableTableNames()
	require.Len(t, allowed, 53)
	require.Len(t, configurationModuleMissingFitmentSpokeAndTireCatalogWritableTableNames(), 12)
	seen := map[string]bool{}
	for _, table := range allowed {
		require.Contains(t, exported, table)
		require.False(t, seen[table], table)
		seen[table] = true
		require.True(t, configurationModuleHasMissingSiteContentWriteContract(configuration, table))
		for _, modules := range [][]string{{"products"}, {"media"}, {"orders"}, {"customers"}, {"customer_service"}, {"configuration", "products"}} {
			require.False(t, configurationModuleHasMissingSiteContentWriteContract(selectedBusinessModuleCompleteSnapshot{Modules: modules}, table))
		}
	}
	for table := range exported {
		require.Equal(t, seen[table], configurationModuleHasMissingSiteContentWriteContract(configuration, table), table)
	}
	for _, table := range []string{"spoke_histories", "quick_buy_sessions", "quick_buy_session_items", "products", "product_variants", "orders", "media_assets", "audit_logs"} {
		require.False(t, configurationModuleHasMissingSiteContentWriteContract(configuration, table), table)
	}
}

func TestConfigurationFitmentCatalogDomainValidationDoesNotNormalizeBackupRowsOrAcceptInvalidFields(t *testing.T) {
	for _, entry := range []struct{ table, validRow, invalidRow string }{
		{"fitment_frame_entries", `{"brand_name":"Test","model_name":"Frame","year_mode":"range","year_from":2020,"year_to":2024}`, `{"brand_name":"Test","model_name":"Frame","year_mode":"range","year_from":2024,"year_to":2020}`},
		{"fitment_fork_entries", `{"brand_name":"Test","model_name":"Fork","year_mode":"unknown"}`, `{"brand_name":"","model_name":"Fork","year_mode":"unknown"}`},
		{"fitment_hub_specifications", `{"spec_code":"TEST-FRONT","display_name":"Front","position":"front","axle_type":"thru_axle","axle_spacing_mm":100}`, `{"spec_code":"TEST-FRONT","display_name":"Front","position":"front","axle_type":"thru_axle","axle_spacing_mm":100,"wr_mm":30}`},
	} {
		t.Run(entry.table, func(t *testing.T) {
			row := json.RawMessage(entry.validRow)
			require.NoError(t, validateConfigurationModuleMissingFitmentCatalogDomainRows(map[string][]json.RawMessage{entry.table: {row}}))
			require.Equal(t, entry.validRow, string(row))
			require.Error(t, validateConfigurationModuleMissingFitmentCatalogDomainRows(map[string][]json.RawMessage{entry.table: {json.RawMessage(entry.invalidRow)}}))
			require.Error(t, validateConfigurationModuleMissingFitmentCatalogDomainRows(map[string][]json.RawMessage{entry.table: {json.RawMessage(`{"id":`)}}))
		})
	}
}

func createConfigurationFitmentCatalogRestoreIntegrationFixture(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO product_brands(id,name,slug) VALUES(900060,'Preserved catalog product brand','configuration-catalog-brand');
INSERT INTO fitment_hub_specifications(id,spec_code,display_name,position,axle_type,axle_spacing_mm,is_enabled,wr_mm,wl_mm,pcdr_mm,pcdl_mm)
VALUES(900060,'RESTORE-FRONT','Restored front hub','front','thru_axle',100,true,30,35,45,45),(900061,'RESTORE-REAR','Restored rear hub','rear','thru_axle',142,true,20,35,45,45);
INSERT INTO spoke_rim_brands(id,code,name,product_brand_id) VALUES(900060,'configuration-rim-brand','Restored rim brand',900060);
INSERT INTO spoke_hub_brands(id,code,name) VALUES(900060,'configuration-hub-brand','Restored hub brand');
INSERT INTO spoke_rim_models(id,brand_id,code,name,erd_mm) VALUES(900060,900060,'configuration-rim-model','Restored rim model',600);
INSERT INTO spoke_hub_models(id,brand_id,code,name,fitment_hub_specification_id) VALUES(900060,900060,'configuration-hub-model','Restored hub model',900060);
INSERT INTO spoke_build_presets(id,code,name,rim_brand_id,rim_model_id,hub_brand_id,hub_model_id,wheel_position,spoke_count,crossing)
VALUES(900060,'configuration-preset','Restored spoke preset',900060,900060,900060,900060,'front',32,3);
INSERT INTO fitment_frame_entries(id,brand_name,model_name,year_mode,is_enabled) VALUES(900060,'Configuration fixture','Restored frame','unknown',true);
INSERT INTO fitment_fork_entries(id,brand_name,model_name,year_mode,is_enabled) VALUES(900060,'Configuration fixture','Restored fork','unknown',true);
INSERT INTO fitment_frame_hub_specifications(frame_entry_id,hub_specification_id) VALUES(900060,900061);
INSERT INTO fitment_fork_hub_specifications(fork_entry_id,hub_specification_id) VALUES(900060,900060);
INSERT INTO schwalbe_tire_specifications(article_no,model_name,etrto,source_url,source_checked_at)
VALUES('CONFIG-RESTORE','Restored tire candidate','40-622','https://synthetic.invalid/tire-catalog','2026-10-31');
INSERT INTO schwalbe_tire_hookless_compatibility(scope_key,article_no,model_name,status,source_basis,source_version,source_url,source_checked_at)
VALUES('article:CONFIG-RESTORE','CONFIG-RESTORE','Restored tire candidate','supported','Independent source fact','synthetic-v1','https://synthetic.invalid/hookless','2026-10-31');
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"wrong_fork_hub_position":      `UPDATE fitment_fork_hub_specifications SET hub_specification_id=900061 WHERE fork_entry_id=900060`,
		"wrong_frame_hub_position":     `UPDATE fitment_frame_hub_specifications SET hub_specification_id=900060 WHERE frame_entry_id=900060`,
		"disabled_associated_hub":      `UPDATE fitment_hub_specifications SET is_enabled=false WHERE id=900061`,
		"enabled_fork_without_hub":     `DELETE FROM fitment_fork_hub_specifications WHERE fork_entry_id=900060`,
		"wrong_preset_rim_brand":       `INSERT INTO spoke_rim_brands(id,code,name) VALUES(900062,'configuration-other-rim-brand','Other brand'); UPDATE spoke_build_presets SET rim_brand_id=900062 WHERE id=900060`,
		"wrong_preset_wheel_position":  `UPDATE spoke_build_presets SET wheel_position='rear' WHERE id=900060`,
		"invalid_catalog_domain_field": `UPDATE fitment_hub_specifications SET display_name='' WHERE id=900060`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func deleteConfigurationFitmentCatalogFixtureRowsAndAddPreservedTargetOnlyEntry(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
DELETE FROM spoke_build_presets WHERE id=900060;
DELETE FROM spoke_rim_models WHERE id=900060;DELETE FROM spoke_hub_models WHERE id=900060;
DELETE FROM spoke_rim_brands WHERE id IN (900060,900062);DELETE FROM spoke_hub_brands WHERE id=900060;
DELETE FROM fitment_frame_entries WHERE id=900060;DELETE FROM fitment_fork_entries WHERE id=900060;
DELETE FROM fitment_hub_specifications WHERE id IN (900060,900061);
DELETE FROM schwalbe_tire_hookless_compatibility WHERE scope_key='article:CONFIG-RESTORE';
DELETE FROM schwalbe_tire_specifications WHERE article_no='CONFIG-RESTORE';
INSERT INTO fitment_frame_entries(id,brand_name,model_name,year_mode,is_enabled) VALUES(900080,'Current extra brand','Preserved current frame','unknown',false);
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"missing_catalog_product_brand":     `DELETE FROM product_brands WHERE id=900060`,
		"catalog_natural_identity_conflict": `INSERT INTO spoke_rim_brands(id,code,name) VALUES(900080,'configuration-rim-brand','Current identity must survive')`,
		"changed_existing_catalog":          `INSERT INTO spoke_hub_brands(id,code,name) VALUES(900060,'configuration-hub-brand','Changed current brand')`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func verifyRestoredConfigurationFitmentCatalogThroughPublicBusinessHTTPAndIndependentTireFacts(t *testing.T, database *sql.DB) {
	t.Helper()
	gormDatabase, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: database}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	hubRepository := repository.NewFitmentHubSpecificationRepository(gormDatabase)
	frameAssociations := repository.NewFitmentFrameHubSpecificationRepository(gormDatabase)
	forkAssociations := repository.NewFitmentForkHubSpecificationRepository(gormDatabase)
	frameService := service.NewFrameFitmentEntryService(repository.NewFrameFitmentEntryRepository(gormDatabase), hubRepository, frameAssociations)
	forkService := service.NewForkFitmentEntryService(repository.NewForkFitmentEntryRepository(gormDatabase), hubRepository, forkAssociations)
	hubService := service.NewFitmentHubSpecificationService(hubRepository, frameAssociations)
	hubService.ConfigureForkHubSpecificationRepository(forkAssociations)
	router := gin.New()
	fitmenthttp.NewHandler(frameService, forkService, hubService).RegisterRoutes(router.Group("/fitment"))
	for _, entry := range []struct{ path, expected string }{
		{"/fitment/frame-entries/900060", "Restored rear hub"},
		{"/fitment/fork-entries/900060", "Restored front hub"},
		{"/fitment/hub-specifications/900060", "RESTORE-FRONT"},
		{"/fitment/frame-entries?search=Restored%20frame", "Restored frame"},
	} {
		t.Run("restored_public_catalog"+entry.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, entry.path, nil))
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			require.Contains(t, response.Body.String(), entry.expected)
		})
	}
	var brand, model, status string
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT rb.name,rm.name FROM spoke_build_presets p JOIN spoke_rim_brands rb ON rb.id=p.rim_brand_id JOIN spoke_rim_models rm ON rm.id=p.rim_model_id WHERE p.id=900060`).Scan(&brand, &model))
	require.Equal(t, "Restored rim brand", brand)
	require.Equal(t, "Restored rim model", model)
	spokeRepository := repository.NewSpokeRepository(gormDatabase, hubRepository)
	spokeRepository.ConfigureProductBrandRepository(repository.NewProductBrandRepository(gormDatabase))
	export, configured, err := spokeRepository.GetCatalogExport()
	require.NoError(t, err)
	require.True(t, configured)
	catalog, err := json.Marshal(export)
	require.NoError(t, err)
	require.Contains(t, string(catalog), "configuration-rim-model")
	require.Contains(t, string(catalog), "Restored front hub")
	require.Contains(t, string(catalog), "configuration-preset")
	tireCandidates, err := repository.NewProductRepository(gormDatabase).ListSchwalbeTireCatalog("CONFIG-RESTORE")
	require.NoError(t, err)
	require.Len(t, tireCandidates, 1)
	require.Equal(t, "Restored tire candidate", tireCandidates[0].ModelName)
	require.False(t, tireCandidates[0].ProductExists)
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT status FROM schwalbe_tire_hookless_compatibility WHERE article_no='CONFIG-RESTORE'`).Scan(&status))
	require.Equal(t, "supported", status)
	var currentExtraRows int
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT count(*) FROM fitment_frame_entries WHERE id=900080 AND model_name='Preserved current frame'`).Scan(&currentExtraRows))
	require.Equal(t, 1, currentExtraRows)
}
