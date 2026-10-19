package repository

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type mediaUploadActorQuotaTestObjectIdentityResolverWithLocalExistenceVerification struct {
	provider storage.StorageService
}

func (r *mediaUploadActorQuotaTestObjectIdentityResolverWithLocalExistenceVerification) ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference string) (string, error) {
	return storage.TrustedBusinessMediaReferenceObjectKey(reference, r.provider, "")
}

func (r *mediaUploadActorQuotaTestObjectIdentityResolverWithLocalExistenceVerification) ResolveRegisteredMediaObjectKeyForCleanupOwnership(reference string) (string, error) {
	return r.provider.ObjectKey(reference)
}

func (r *mediaUploadActorQuotaTestObjectIdentityResolverWithLocalExistenceVerification) ValidateMediaObjectExistsBeforeRegisteringOwnership(ctx context.Context, key string) error {
	return r.provider.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(ctx, key)
}

func createIsolatedPostgreSQLMediaUploadActorQuotaRepository(t *testing.T) *MediaRepository {
	t.Helper()
	host, username, password := os.Getenv("DB_HOST"), os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD")
	if host == "" || username == "" || password == "" {
		t.Skip("PostgreSQL integration environment is not configured")
	}
	port, databaseName := os.Getenv("DB_PORT"), os.Getenv("DB_NAME")
	if port == "" {
		port = "5432"
	}
	if databaseName == "" {
		databaseName = "postgres"
	}
	connectionURL := &url.URL{Scheme: "postgres", User: url.UserPassword(username, password), Host: net.JoinHostPort(host, port), Path: "/" + databaseName, RawQuery: "sslmode=disable&connect_timeout=5"}
	bootstrapDB, err := gorm.Open(postgres.Open(connectionURL.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	bootstrapSQL, err := bootstrapDB.DB()
	require.NoError(t, err)
	schemaName := "test_media_upload_actor_quota_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, bootstrapDB.Exec("CREATE SCHEMA "+schemaName).Error)
	t.Cleanup(func() {
		require.NoError(t, bootstrapDB.Exec("DROP SCHEMA "+schemaName+" CASCADE").Error)
		_ = bootstrapSQL.Close()
	})
	query := connectionURL.Query()
	query.Set("search_path", schemaName)
	connectionURL.RawQuery = query.Encode()
	db, err := gorm.Open(postgres.Open(connectionURL.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(8)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}))
	objectRoot := t.TempDir()
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: objectRoot, BaseURL: "https://media.example.test"})
	require.NoError(t, err)
	objectKeys := []string{strings.Repeat("a", 64) + ".png", strings.Repeat("b", 64) + ".png"}
	for index := 0; index < 8; index++ {
		objectKeys = append(objectKeys, fmt.Sprintf("%d.png", index))
	}
	for _, key := range objectKeys {
		require.NoError(t, os.WriteFile(filepath.Join(objectRoot, key), []byte("synthetic quota fixture"), 0o600))
	}
	repo := NewMediaRepository(db)
	repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(&mediaUploadActorQuotaTestObjectIdentityResolverWithLocalExistenceVerification{provider: provider})
	return repo
}

func TestPostgreSQLMediaUploadActorStorageQuotaSerializesConcurrentRegistrations(t *testing.T) {
	for _, actorType := range []media.MediaUploadActorType{media.MediaUploadActorVisitor, media.MediaUploadActorAdministrator} {
		t.Run(string(actorType), func(t *testing.T) {
			repo := createIsolatedPostgreSQLMediaUploadActorQuotaRepository(t)
			start := make(chan struct{})
			results := make(chan error, 8)
			var uploads sync.WaitGroup
			for index := 0; index < 8; index++ {
				uploads.Go(func() {
					asset := &media.MediaAsset{Filename: fmt.Sprintf("concurrent-%d.png", index), URL: fmt.Sprintf("https://media.example.test/uploads/%d.png", index), Size: 6, UploaderActorType: actorType}
					if actorType == media.MediaUploadActorVisitor {
						asset.UploaderVisitorSessionHash = strings.Repeat("a", 64)
					} else {
						asset.UploaderID = 42
					}
					<-start
					results <- repo.CreateAssetWithinUploadActorStorageQuota(t.Context(), asset, 10)
				})
			}
			close(start)
			uploads.Wait()
			close(results)
			successCount, quotaRejectionCount := 0, 0
			for err := range results {
				if err == nil {
					successCount++
				} else if errors.Is(err, ErrMediaUploadActorStorageQuotaExceeded) {
					quotaRejectionCount++
				} else {
					t.Fatalf("unexpected concurrent upload error: %v", err)
				}
			}
			require.Equal(t, 1, successCount)
			require.Equal(t, 7, quotaRejectionCount)
			var count int64
			require.NoError(t, repo.db.Model(&media.MediaAsset{}).Count(&count).Error)
			require.EqualValues(t, 1, count)
		})
	}
}

func TestPostgreSQLMediaUploadActorStorageQuotaKeepsDifferentVisitorSessionsIndependent(t *testing.T) {
	repo := createIsolatedPostgreSQLMediaUploadActorQuotaRepository(t)
	for _, visitorHash := range []string{strings.Repeat("a", 64), strings.Repeat("b", 64)} {
		asset := &media.MediaAsset{Filename: "guest.png", URL: "https://media.example.test/uploads/" + visitorHash + ".png", Size: 10,
			UploaderActorType: media.MediaUploadActorVisitor, UploaderVisitorSessionHash: visitorHash}
		require.NoError(t, repo.CreateAssetWithinUploadActorStorageQuota(t.Context(), asset, 10))
	}
}
