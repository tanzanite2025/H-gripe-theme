package v1

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/app"
	"commerce-platform/internal/domain/warranty"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/service"
)

func TestPrivateOrderAttachmentActualRoutesRequireAuthenticationAndHeaderBoundGuestClaimToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, warrantyService, _ := newWarrantyRouteFixture(t)
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(), BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	reference, err := provider.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(t.Context(), strings.NewReader("guest private evidence"), "warranty/historical.png")
	require.NoError(t, err)
	warrantyService.ConfigurePrivateOrderBusinessAttachmentUploads(service.NewPrivateOrderBusinessAttachmentUploadService(nil, provider, "https://shop.example.test"))
	claim := warranty.WarrantyClaim{OrderNumber: "PRIVATE-GUEST-READ", Email: "guest@example.test", Images: fmt.Sprintf("[%q]", reference)}
	require.NoError(t, db.Omit("OrderItem", "ServiceRecords").Create(&claim).Error)
	token, err := warrantyService.IssueWarrantyClaimAccessToken(&claim)
	require.NoError(t, err)
	router := gin.New()
	RegisterRoutes(router, &app.Dependencies{Services: app.Services{Warranty: warrantyService}}, &config.Config{JWT: config.JWTConfig{Secret: "test-email-secret"}})
	path := fmt.Sprintf("/api/v1/warranty/claims/%d/attachments/images/0", claim.ID)
	for _, fixture := range []struct {
		name, path, token string
		status            int
	}{
		{"missing identity", path, "", http.StatusUnauthorized},
		{"query token is not a file credential", path + "?access_token=" + token, "", http.StatusUnauthorized},
		{"invalid guest token", path, token + "invalid", http.StatusForbidden},
		{"authorized guest", path, token, http.StatusOK},
		{"out of range", fmt.Sprintf("/api/v1/warranty/claims/%d/attachments/images/1", claim.ID), token, http.StatusNotFound},
		{"customer shipment route", "/api/v1/warranty/orders/ORDER/images/0", "", http.StatusUnauthorized},
		{"customer after-sales route", "/api/v1/orders/ORDER/after-sales/attachments/opaque", "", http.StatusUnauthorized},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, fixture.path, nil)
			if fixture.token != "" {
				request.Header.Set("X-Warranty-Claim-Token", fixture.token)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			require.Equal(t, fixture.status, recorder.Code, recorder.Body.String())
			require.NotContains(t, recorder.Body.String(), reference)
			if recorder.Code == http.StatusOK {
				require.Equal(t, "guest private evidence", recorder.Body.String())
				require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
				require.NotContains(t, recorder.Header().Get("Content-Disposition"), token)
			}
		})
	}
}
