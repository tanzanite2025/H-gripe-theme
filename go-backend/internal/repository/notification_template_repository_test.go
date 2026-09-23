package repository

import (
	"testing"

	"commerce-platform/internal/domain/notification"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestNotificationTemplateRepositorySaveWithVersionIsAtomic(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&notification.EmailTemplate{}, &notification.EmailTemplateVersion{}))

	repo := NewNotificationTemplateRepository(db)
	template := &notification.EmailTemplate{
		Code:              "order_confirmation",
		Locale:            "en",
		Category:          notification.TemplateCategoryOrder,
		Name:              "Order confirmation",
		SubjectTemplate:   "Order {{order_number}}",
		BodyText:          "Thanks for order {{order_number}}",
		AllowedVariables:  datatypes.JSON(`["order_number"]`),
		RequiredVariables: datatypes.JSON(`["order_number"]`),
		IsEnabled:         true,
		Version:           1,
	}
	version := &notification.EmailTemplateVersion{Version: 1, ChangeReason: "initial seed"}
	require.NoError(t, repo.SaveWithVersion(template, version))
	assert.NotZero(t, template.ID)
	assert.Equal(t, template.ID, version.TemplateID)

	stored, err := repo.FindEnabledByCodeLocale("order_confirmation", "en")
	require.NoError(t, err)
	assert.Equal(t, template.ID, stored.ID)
	versions, err := repo.ListVersions(template.ID)
	require.NoError(t, err)
	require.Len(t, versions, 1)
	assert.Equal(t, "initial seed", versions[0].ChangeReason)
}

func TestNotificationTemplateRepositorySaveWithVersionRejectsInvalidTemplate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&notification.EmailTemplate{}, &notification.EmailTemplateVersion{}))

	err = NewNotificationTemplateRepository(db).SaveWithVersion(
		&notification.EmailTemplate{Code: "order_confirmation"},
		&notification.EmailTemplateVersion{Version: 1},
	)
	assert.ErrorIs(t, err, notification.ErrTemplateLocaleRequired)
}

func TestNotificationTemplateRepositorySaveWithVersionUsesOptimisticVersioning(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&notification.EmailTemplate{}, &notification.EmailTemplateVersion{}))

	repo := NewNotificationTemplateRepository(db)
	template := &notification.EmailTemplate{
		Code:            "order_confirmation",
		Locale:          "en",
		Category:        notification.TemplateCategoryOrder,
		Name:            "Order confirmation",
		SubjectTemplate: "Order",
		BodyText:        "Thanks",
		Version:         1,
	}
	require.NoError(t, repo.SaveWithVersion(template, &notification.EmailTemplateVersion{Version: 1}))

	template.Version = 3
	err = repo.SaveWithVersion(template, &notification.EmailTemplateVersion{Version: 3})
	assert.ErrorIs(t, err, notification.ErrTemplateVersionConflict)
}
