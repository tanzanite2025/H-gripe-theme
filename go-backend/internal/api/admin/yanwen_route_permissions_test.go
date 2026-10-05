package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

// TestYanwenViewRoutesAreReachableWithViewPermission verifies every read-only
// Yanwen endpoint is registered behind the domain view permission. The
// handlers intentionally use nil services here; a non-403/non-404 response
// proves the permission middleware admitted the request before the handler's
// dependency error.
func TestYanwenViewRoutesAreReachableWithViewPermission(t *testing.T) {
	viewRoutes := []struct {
		name   string
		method string
		path   string
	}{
		{name: "published channels", method: http.MethodGet, path: "/logistics/yanwen/channels?environment=production"},
		{name: "published collection references", method: http.MethodGet, path: "/logistics/yanwen/collection"},
		{name: "overview", method: http.MethodGet, path: "/logistics/yanwen/overview?environment=production"},
		{name: "api config", method: http.MethodGet, path: "/logistics/yanwen/api/config?environment=fat"},
		{name: "official products", method: http.MethodGet, path: "/logistics/yanwen/api/products?environment=fat"},
		{name: "official countries", method: http.MethodGet, path: "/logistics/yanwen/api/countries?environment=fat"},
		{name: "official warehouses", method: http.MethodGet, path: "/logistics/yanwen/api/warehouses?environment=fat"},
		{name: "waybills", method: http.MethodGet, path: "/logistics/yanwen/waybills?environment=fat"},
		{name: "tracking", method: http.MethodGet, path: "/logistics/yanwen/tracking?nums=YE123"},
		{name: "tracking alerts", method: http.MethodGet, path: "/logistics/yanwen/tracking/alerts"},
	}

	for _, route := range viewRoutes {
		t.Run(route.name, func(t *testing.T) {
			status := performYanwenPermissionRequest(t, auth.RoleViewer, route.method, route.path)
			if status == http.StatusForbidden || status == http.StatusNotFound {
				t.Fatalf("Yanwen view route %s %s was blocked with status %d", route.method, route.path, status)
			}
		})
	}
}

// TestYanwenActionRoutesUseDedicatedPermissions verifies that a role with
// Yanwen view access alone cannot invoke collection management, catalog sync,
// gateway mutation, waybill actions, or official customs verification.
func TestYanwenActionRoutesUseDedicatedPermissions(t *testing.T) {
	actionRoutes := []struct {
		name string
		path string
	}{
		{name: "create collection channel", path: "/logistics/yanwen/channels"},
		{name: "update collection channel", path: "/logistics/yanwen/channels/1"},
		{name: "delete collection channel", path: "/logistics/yanwen/channels/1"},
		{name: "save api config", path: "/logistics/yanwen/api/config"},
		{name: "ping api", path: "/logistics/yanwen/api/ping"},
		{name: "sync products", path: "/logistics/yanwen/api/sync-products"},
		{name: "sync countries", path: "/logistics/yanwen/api/sync-countries"},
		{name: "sync warehouses", path: "/logistics/yanwen/api/sync-warehouses"},
		{name: "create waybill", path: "/logistics/yanwen/waybills"},
		{name: "batch create waybills", path: "/logistics/yanwen/waybills/batch"},
		{name: "batch sync waybills", path: "/logistics/yanwen/waybills/batch-sync"},
		{name: "batch labels", path: "/logistics/yanwen/waybills/batch-labels"},
		{name: "sync waybill", path: "/logistics/yanwen/waybills/1/sync"},
		{name: "single label", path: "/logistics/yanwen/waybills/1/label"},
		{name: "batch cancel waybills", path: "/logistics/yanwen/waybills/batch-cancel"},
		{name: "cancel waybill", path: "/logistics/yanwen/waybills/1/cancel"},
		{name: "Korea PCCC", path: "/logistics/yanwen/customs/korea/pccc"},
		{name: "United States address", path: "/logistics/yanwen/customs/united-states/address"},
	}

	methods := map[string]string{
		"create collection channel": http.MethodPost,
		"update collection channel": http.MethodPut,
		"delete collection channel": http.MethodDelete,
		"save api config":           http.MethodPut,
		"ping api":                  http.MethodPost,
		"sync products":             http.MethodPost,
		"sync countries":            http.MethodPost,
		"sync warehouses":           http.MethodPost,
		"create waybill":            http.MethodPost,
		"batch create waybills":     http.MethodPost,
		"batch sync waybills":       http.MethodPost,
		"batch labels":              http.MethodPost,
		"sync waybill":              http.MethodPost,
		"single label":              http.MethodPost,
		"batch cancel waybills":     http.MethodPost,
		"cancel waybill":            http.MethodPost,
		"Korea PCCC":                http.MethodPost,
		"United States address":     http.MethodPost,
	}

	for _, route := range actionRoutes {
		t.Run(route.name, func(t *testing.T) {
			status := performYanwenPermissionRequest(t, auth.RoleViewer, methods[route.name], route.path)
			if status != http.StatusForbidden {
				t.Fatalf("Yanwen action route %s %s admitted view-only role with status %d", methods[route.name], route.path, status)
			}
		})
	}
}

func performYanwenPermissionRequest(t *testing.T, role auth.Role, method, path string) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authenticated := router.Group("")
	authenticated.Use(func(c *gin.Context) {
		c.Set("user_role", string(role))
		c.Next()
	})

	registerLogisticsDomainRoutes(authenticated, &FpxHandler{}, NewYanwenCollectionHandler(nil))
	registerYanwenAPIConfigRoutes(authenticated, NewYanwenAPIConfigHandler(nil))
	registerYanwenCatalogRoutes(authenticated, NewYanwenCatalogHandler(nil))
	registerYanwenOverviewRoutes(authenticated, NewYanwenOverviewHandler(nil))
	registerYanwenWaybillRoutes(authenticated, NewYanwenWaybillHandler(nil))
	registerYanwenTrackingRoutes(authenticated, NewYanwenTrackingHandler(nil))
	registerYanwenTrackingAlertRoutes(authenticated, NewYanwenTrackingAlertHandler(nil))
	registerYanwenKoreaPersonalCustomsClearanceCodeRoutes(authenticated, NewYanwenKoreaPersonalCustomsClearanceCodeHandler(nil))
	registerYanwenUnitedStatesAddressVerificationRoutes(authenticated, NewYanwenUnitedStatesAddressVerificationHandler(nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, nil)
	router.ServeHTTP(recorder, request)
	return recorder.Code
}
