package localbackup

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/api/publicuploadresponse"
	publicticket "commerce-platform/internal/api/v1/ticket"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/securecookie"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func verifyRestoredMediaModulePublicImagesAndPrivateAttachmentAuthorization(t *testing.T, cfg Configuration, database *sql.DB) {
	t.Helper()
	orm, err := gorm.Open(postgres.New(postgres.Config{Conn: database}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	store, err := storage.NewReadOnlyLocalRecoveryObjectStorageWithoutDirectoryCreation(storage.Config{Type: storage.StorageTypeLocal, LocalPath: cfg.StorageRoots["uploads"], PrivateLocalPath: cfg.StorageRoots["uploads_private"], BaseURL: "https://shop.example.test/uploads", PrivateBaseURL: "https://shop.example.test/uploads"})
	require.NoError(t, err)
	mediaService := service.NewMediaService(repository.NewMediaRepository(orm), store, nil, "https://shop.example.test/uploads", 0)
	users := repository.NewUserRepository(orm)
	tickets := service.NewTicketService(repository.NewTicketRepository(orm), users)
	handler := publicticket.NewHandler(tickets, publicticket.Options{MediaService: mediaService, VisitorSecret: cfg.RecoverySecrets["JWT_SECRET"]})
	authentication := service.NewAuthService(users, config.JWTConfig{Secret: cfg.RecoverySecrets["JWT_SECRET"]})
	router := gin.New()
	router.GET("/uploads/*key", publicuploadresponse.NewRegisteredLocalUploadHTTPHandler(service.NewPublicUploadAccessService(mediaService, nil), publicuploadresponse.NewRestrictedLocalUploadFileServer(cfg.StorageRoots["uploads"]), nil))
	router.GET("/api/v1/customer-service/conversations/:conversation_id/messages/:message_id/attachments/:attachment_index/file", middleware.OptionalAuthMiddleware(authentication), handler.ServePublicCustomerServiceMessageAttachmentFile)
	for _, scenario := range []string{"public_original", "public_derivative", "anonymous_private_path", "anonymous_attachment", "owner_original", "owner_derivative", "other_account", "tampered_cookie"} {
		t.Run("restored_media_http_"+scenario, func(t *testing.T) {
			target := "/api/v1/customer-service/conversations/preserved-media-chat/messages/900001/attachments/0/file"
			status := http.StatusOK
			switch scenario {
			case "public_original":
				target = "/uploads/products/media-original.png"
			case "public_derivative":
				target = "/uploads/media-derivatives/900001/thumbnail/v1/public.png"
			case "anonymous_private_path":
				target = "/uploads/customer-service/attachments/originals/media-private.png"
				status = http.StatusNotFound
			case "anonymous_attachment", "other_account", "tampered_cookie":
				status = http.StatusNotFound
			case "owner_derivative":
				target += "?preset=thumbnail"
			}
			request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
			if scenario == "owner_original" || scenario == "owner_derivative" || scenario == "other_account" || scenario == "tampered_cookie" {
				id := uint(900001)
				if scenario == "other_account" {
					id = 900002
				}
				token := signSyntheticStoppedProductMediaBusinessAccessCookie(t, id, "user", "access", cfg.RecoverySecrets["JWT_SECRET"], false)
				if scenario == "tampered_cookie" {
					token += "x"
				}
				request.AddCookie(&http.Cookie{Name: securecookie.StorefrontAuthTokenCookie, Value: token})
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, status, response.Code, response.Body.String())
			if status == http.StatusOK {
				require.Equal(t, "synthetic-media-file!", response.Body.String())
			}
			if scenario != "public_original" && scenario != "public_derivative" {
				require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
			}
		})
	}
}
