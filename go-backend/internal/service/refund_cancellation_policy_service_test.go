package service

import (
	"testing"

	refundcancellationdomain "commerce-platform/internal/domain/refundcancellation"
	settingdomain "commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRefundCancellationPolicyServiceFallsBackAndPersistsLocaleContent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}))

	policyService := NewRefundCancellationPolicyService(repository.NewSettingRepository(db))
	defaultResult, err := policyService.GetPublic("fr")
	require.NoError(t, err)
	require.True(t, defaultResult.Fallback)
	require.Equal(t, "en", defaultResult.Locale)
	require.NotEmpty(t, defaultResult.Policy.Sections)

	policy := refundcancellationdomain.DefaultPolicy()
	policy.Title = "Politique de retour"
	policy.Sections = []refundcancellationdomain.Section{{
		ID:    "image-proof",
		Title: "Photo evidence",
		Body:  "Attach a clear photo when the package arrives damaged.",
		Image: &refundcancellationdomain.Image{
			URL:     "/uploads/policy/damaged-package.webp",
			Alt:     "Damaged package",
			Caption: "Example of visible package damage.",
		},
	}}
	_, err = policyService.Update("fr-FR", policy)
	require.NoError(t, err)

	savedResult, err := policyService.GetPublic("fr")
	require.NoError(t, err)
	require.False(t, savedResult.Fallback)
	require.Equal(t, "fr", savedResult.Locale)
	require.Equal(t, "Politique de retour", savedResult.Policy.Title)
	require.Equal(t, "/uploads/policy/damaged-package.webp", savedResult.Policy.Sections[0].Image.URL)
}

func TestRefundCancellationPolicyServiceDoesNotFallbackOnDatabaseError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}))
	require.NoError(t, db.Migrator().DropTable(&settingdomain.Setting{}))

	policyService := NewRefundCancellationPolicyService(repository.NewSettingRepository(db))
	result, err := policyService.GetPublic("fr")

	require.Error(t, err)
	require.Empty(t, result.Policy.Title)
	require.NotContains(t, err.Error(), "Refund & Cancellation Policy")
}
