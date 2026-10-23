package localbackup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func verifyStoppedProductProjectionActualSQLBaselineOperatorRecordsAndStaleCacheRefusal(t *testing.T, manager *Manager, taskID string, redisOptions *redis.Options, invokeHTTP func(string, any, int) *httptest.ResponseRecorder) {
	t.Helper()
	ctx := context.Background()
	readBaseline := func() ProductCatalogStoppedStorefrontProjectionInspection {
		var result struct {
			Report ProductCatalogStoppedStorefrontProjectionInspection
		}
		require.NoError(t, json.Unmarshal(invokeHTTP("product-catalog-stopped-projection-inspection", struct{}{}, http.StatusOK).Body.Bytes(), &result))
		return result.Report
	}
	before, err := manager.GetTask(taskID)
	require.NoError(t, err)
	first, second := readBaseline(), readBaseline()
	require.Equal(t, first.ContractSHA256, second.ContractSHA256)
	require.NotEmpty(t, first.OriginalRoutes)
	require.NotEmpty(t, first.FinalRoutes)
	require.Len(t, first.PackageImageFiles, 3)
	require.Len(t, first.Requirements, 3)
	require.Equal(t, before.Journal.NewDatabaseOID, first.DatabaseOID)
	require.Equal(t, before.ProductCatalogStoppedDetailCacheInvalidation.SHA256, first.DetailCacheProgressSHA256)
	require.False(t, first.OnlineOverwriteAvailable)
	require.True(t, first.ApplicationRemainsStopped)
	afterRead, err := manager.GetTask(taskID)
	require.NoError(t, err)
	require.Equal(t, before, afterRead, "SQL projection inspection must not save approval or change the task")
	request := ProductCatalogStoppedProjectionEvidenceRequest{Confirmation: "RECORD STOPPED PRODUCT PROJECTIONS " + manager.Configuration.SiteID + "/" + manager.Configuration.Environment + "/" + taskID, ContractSHA256: first.ContractSHA256}
	for _, requirement := range first.Requirements {
		request.Attestations = append(request.Attestations, ExternalProjectionReconciliationAttestation{Projection: requirement.Projection, Action: requirement.Action, EvidenceReference: "synthetic-only-independent-record-" + requirement.Projection, EvidenceSHA256: productCatalogBytesSHA256([]byte("synthetic-only-record-" + requirement.Projection))})
	}
	// These are synthetic operator assertions, not actual storefront/CDN checks.
	incomplete := request
	incomplete.Attestations = incomplete.Attestations[:2]
	invokeHTTP("product-catalog-stopped-projection-evidence", incomplete, http.StatusConflict)
	wrongContract := request
	wrongContract.ContractSHA256 = strings.Repeat("b", 64)
	invokeHTTP("product-catalog-stopped-projection-evidence", wrongContract, http.StatusConflict)
	var saved struct {
		Record ProductCatalogStoppedProjectionOperatorEvidenceRecord
	}
	require.NoError(t, json.Unmarshal(invokeHTTP("product-catalog-stopped-projection-evidence", request, http.StatusOK).Body.Bytes(), &saved))
	require.Equal(t, "operator_attestation_only_external_results_not_automatically_verified", saved.Record.Verification)
	require.Equal(t, first.ContractSHA256, saved.Record.ContractSHA256)
	require.Len(t, saved.Record.Attestations, 3)
	stored, err := manager.GetTask(taskID)
	require.NoError(t, err)
	require.Equal(t, "product_module_switched_still_isolated", stored.State)
	require.Equal(t, before.Journal, stored.Journal)
	require.Equal(t, saved.Record.SHA256, stored.ProductCatalogStoppedProjectionEvidence.SHA256)
	require.Error(t, manager.ReleaseUnmodifiedMaintenance(taskID))
	require.Error(t, manager.CompleteReconciledRecovery(taskID, "RESUME "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+taskID, ReconciliationEvidence{}))
	invokeHTTP("product-catalog-stopped-projection-evidence", request, http.StatusConflict)
	// Refilled details block both inspection and evidence without deleting them.
	client := redis.NewClient(redisOptions)
	cacheKey := "product:7654321"
	require.NoError(t, client.Set(ctx, cacheKey, "synthetic-unexpected-reader", 0).Err())
	require.NoError(t, client.Close())
	invokeHTTP("product-catalog-stopped-projection-inspection", struct{}{}, http.StatusConflict)
	request.ExpectedRecordSHA256 = saved.Record.SHA256
	invokeHTTP("product-catalog-stopped-projection-evidence", request, http.StatusConflict)
	client = redis.NewClient(redisOptions)
	require.Equal(t, "synthetic-unexpected-reader", client.Get(ctx, cacheKey).Val())
	require.NoError(t, client.Del(ctx, cacheKey).Err())
	require.NoError(t, client.Close())
	// A fresh cache observation changes the contract. Old manual evidence stays
	// historical and cannot be relabelled as matching the new snapshot.
	invokeHTTP("product-catalog-stopped-detail-cache-invalidation", ProductCatalogStoppedSwitchConfirmationRequest{Confirmation: "INVALIDATE STOPPED PRODUCT CACHE " + manager.Configuration.SiteID + "/" + manager.Configuration.Environment + "/" + taskID}, http.StatusOK)
	refreshed := waitForIndependentRecoveryTask(t, manager, taskID)
	require.Equal(t, "product_module_switched_still_isolated", refreshed.State, refreshed.Error)
	third := readBaseline()
	require.NotEqual(t, first.ContractSHA256, third.ContractSHA256)
	require.NotEqual(t, first.DetailCacheProgressSHA256, third.DetailCacheProgressSHA256)
	invokeHTTP("product-catalog-stopped-projection-evidence", request, http.StatusConflict)
	request.ContractSHA256 = third.ContractSHA256
	invokeHTTP("product-catalog-stopped-projection-evidence", request, http.StatusOK)
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	defer control.Close()
	var allowed int
	require.NoError(t, control.QueryRowContext(ctx, `SELECT count(*) FROM pg_database WHERE oid=ANY($1) AND datallowconn`, pq.Array([]int64{before.Journal.OldDatabaseOID, before.Journal.NewDatabaseOID})).Scan(&allowed))
	require.Zero(t, allowed, "both databases must remain sealed after reading and saving evidence")
}

func TestStoppedProductProjectionOperatorEvidenceRejectsWrongScopeActionsMissingRecordsAndInvalidDigests(t *testing.T) {
	report := &ProductCatalogStoppedStorefrontProjectionInspection{Version: 1, Scope: productCatalogBackupScope, ApplicationRemainsStopped: true, ContractSHA256: strings.Repeat("a", 64), Requirements: stoppedProductStorefrontIndependentProjectionRequirements()}
	valid := ProductCatalogStoppedProjectionEvidenceRequest{ContractSHA256: report.ContractSHA256}
	for _, requirement := range report.Requirements {
		valid.Attestations = append(valid.Attestations, ExternalProjectionReconciliationAttestation{Projection: requirement.Projection, Action: requirement.Action, EvidenceReference: "synthetic-independent-record", EvidenceSHA256: strings.Repeat("b", 64)})
	}
	require.NoError(t, validateStoppedProductProjectionOperatorAttestations(report, valid))
	for _, scenario := range []string{"missing", "duplicate", "whole_storefront_action", "bad_digest", "empty_reference", "stale_contract", "claims_release", "wrong_scope"} {
		t.Run(scenario, func(t *testing.T) {
			copy := valid
			copy.Attestations = append([]ExternalProjectionReconciliationAttestation{}, valid.Attestations...)
			reportCopy := *report
			switch scenario {
			case "missing":
				copy.Attestations = copy.Attestations[:2]
			case "duplicate":
				copy.Attestations[1] = copy.Attestations[0]
			case "whole_storefront_action":
				copy.Attestations[0].Action = "sql_search_verified_without_external_index"
			case "bad_digest":
				copy.Attestations[0].EvidenceSHA256 = strings.Repeat("0", 64)
			case "empty_reference":
				copy.Attestations[0].EvidenceReference = "  "
			case "stale_contract":
				copy.ContractSHA256 = strings.Repeat("c", 64)
			case "claims_release":
				reportCopy.OnlineOverwriteAvailable = true
			case "wrong_scope":
				reportCopy.Scope = "whole_application"
			}
			require.Error(t, validateStoppedProductProjectionOperatorAttestations(&reportCopy, copy))
		})
	}
}
