package urlmanagement

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createCurrentURLStateHTTPTestServices(t *testing.T) (*gorm.DB, *service.StorefrontRouteCatalogService, *service.StorefrontURLIssueService, seodomain.StorefrontRouteCatalogEntry) {
	t.Helper()
	t.Setenv("STOREFRONT_ROUTE_MANIFEST_PATH", "")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&seodomain.StorefrontRouteCatalogEntry{}, &seodomain.StorefrontRouteCheckResult{}, &urlmanagementdomain.StorefrontURLIssue{}, &urlmanagementdomain.StorefrontURLIssueEvent{}))
	storefront := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/storefront-route-manifest.json" {
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprint(writer, `{"version":"current","routes":[{"key":"current","path":"/current","canonical_path":"/current","sitemap_locales":["en"],"is_checkable":true,"is_searchable":true,"is_indexable":true}]}`)
			return
		}
		writer.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(writer, `<html><head><link rel="canonical" href="http://%s/current"></head><body>Recovered</body></html>`, request.Host)
	}))
	t.Cleanup(storefront.Close)
	catalogRepository := repository.NewStorefrontRouteCatalogRepository(db)
	issues := service.NewStorefrontURLIssueService(repository.NewStorefrontURLIssueRepository(db), catalogRepository, nil)
	catalog := service.NewStorefrontRouteCatalogService(catalogRepository, nil, nil, storefront.URL, storefront.URL)
	catalog.ConfigureIssueReconciler(issues)
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey: "manifest:current:en", Path: "/current", CanonicalPath: "/current", Locale: "en", SourceType: "static", EntryStatus: "active", IsCheckable: true, LastSeenAt: time.Now().UTC().Add(-time.Hour),
	}
	require.NoError(t, db.Create(&entry).Error)
	previousCheck := seodomain.StorefrontRouteCheckResult{RouteEntryID: entry.ID, Status: "not_found", HTTPStatus: 404}
	require.NoError(t, catalogRepository.SaveCheck(&previousCheck))
	_, err = repository.NewStorefrontURLIssueRepository(db).RecordDetection(entry.ID, "not_found", "high", &previousCheck.ID, time.Now().UTC())
	require.NoError(t, err)
	return db, catalog, issues, entry
}

func TestRecoveredURLIssueRecheckAndVerificationReturnSuccessAfterDeletingProblemRecord(t *testing.T) {
	for _, action := range []string{"recheck", "verify"} {
		t.Run(action, func(t *testing.T) {
			db, catalog, issues, _ := createCurrentURLStateHTTPTestServices(t)
			var issue urlmanagementdomain.StorefrontURLIssue
			require.NoError(t, db.First(&issue).Error)
			if action == "verify" {
				require.NoError(t, db.Model(&issue).Update("state", "resolved").Error)
			}
			handler := NewIssuesHandler(issues, catalog)
			router := gin.New()
			router.POST("/:id/recheck", handler.Recheck)
			router.POST("/:id/verify", handler.Verify)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, fmt.Sprintf("/%d/%s", issue.ID, action), nil))
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			var body struct {
				Data struct {
					Cleared bool                                    `json:"cleared"`
					Issue   *urlmanagementdomain.StorefrontURLIssue `json:"issue"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
			require.True(t, body.Data.Cleared)
			require.Nil(t, body.Data.Issue)
			for _, model := range []interface{}{&urlmanagementdomain.StorefrontURLIssue{}, &urlmanagementdomain.StorefrontURLIssueEvent{}} {
				var count int64
				require.NoError(t, db.Model(model).Count(&count).Error)
				require.Zero(t, count)
			}
		})
	}
}

func TestURLCheckHTTPTaskSynchronizesCurrentManifestDeletesRemovedURLsAndReplacesOldErrorWithHealthyState(t *testing.T) {
	db, catalog, _, _ := createCurrentURLStateHTTPTestServices(t)
	removedEntry := seodomain.StorefrontRouteCatalogEntry{RouteKey: "manifest:removed:en", Path: "/removed", Locale: "en", SourceType: "static", EntryStatus: "stale", LastSeenAt: time.Now().UTC().Add(-time.Hour)}
	require.NoError(t, db.Create(&removedEntry).Error)
	router := gin.New()
	router.POST("/check", NewRoutesHandler(catalog).Check)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/check?locale=en&include_aliases=true", nil))
	require.Equal(t, http.StatusAccepted, response.Code, response.Body.String())
	var body struct {
		Data service.StorefrontRouteCatalogCheckTask `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Eventually(t, func() bool {
		task, err := catalog.GetCheckTask(body.Data.ID)
		return err == nil && task.Status == service.StorefrontRouteCatalogCheckTaskCompleted
	}, 5*time.Second, 10*time.Millisecond)
	stats, err := catalog.StatsForLocale("en")
	require.NoError(t, err)
	require.EqualValues(t, 1, stats.Total)
	require.EqualValues(t, 1, stats.OK)
	require.Zero(t, stats.NeedsAttention)
	require.Zero(t, stats.Unchecked)
	var checks []seodomain.StorefrontRouteCheckResult
	require.NoError(t, db.Find(&checks).Error)
	require.Len(t, checks, 1)
	require.Equal(t, "ok", checks[0].Status)
	var issueCount int64
	require.NoError(t, db.Model(&urlmanagementdomain.StorefrontURLIssue{}).Count(&issueCount).Error)
	require.Zero(t, issueCount)
	var removedRoute seodomain.StorefrontRouteCatalogEntry
	require.ErrorIs(t, db.First(&removedRoute, removedEntry.ID).Error, gorm.ErrRecordNotFound)
}
