package service

import (
	"testing"

	"commerce-platform/internal/domain/notification"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newTransactionalNotificationTemplateServiceForTest(t *testing.T) *TransactionalNotificationTemplateService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&notification.EmailTemplate{}, &notification.EmailTemplateVersion{}))
	t.Cleanup(func() {
		sqlDB, dbErr := db.DB()
		if dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	return NewTransactionalNotificationTemplateService(repository.NewNotificationTemplateRepository(db))
}

func orderConfirmationTemplateInput(locale string, version int) SaveTransactionalNotificationTemplateInput {
	return SaveTransactionalNotificationTemplateInput{
		Code:              NotificationTemplateOrderConfirmation,
		Locale:            locale,
		Category:          "order",
		Name:              "Order confirmation",
		SubjectTemplate:   "Order {{order_number}}",
		BodyText:          "Thanks for order {{order_number}}",
		AllowedVariables:  []string{"order_number", "order_amount", "paid_at", "customer_name", "currency", "items"},
		RequiredVariables: []string{"order_number", "order_amount", "paid_at"},
		IsEnabled:         true,
		Version:           version,
	}
}

func TestTransactionalNotificationTemplateServiceSaveAndResolveUsesLocaleFallback(t *testing.T) {
	service := newTransactionalNotificationTemplateServiceForTest(t)
	_, err := service.Save(orderConfirmationTemplateInput("en", 1))
	require.NoError(t, err)

	resolved, err := service.Resolve(NotificationTemplateOrderConfirmation, "fr")
	require.NoError(t, err)
	assert.Equal(t, "en", resolved.Locale)
	assert.Equal(t, "Order {{order_number}}", resolved.SubjectTemplate)
}

func TestTransactionalNotificationTemplateServiceRejectsDefinitionDrift(t *testing.T) {
	service := newTransactionalNotificationTemplateServiceForTest(t)
	input := orderConfirmationTemplateInput("en", 1)
	input.RequiredVariables = []string{"order_number"}
	_, err := service.Save(input)
	assert.ErrorIs(t, err, ErrNotificationTemplateDefinitionMismatch)
}

func TestTransactionalNotificationTemplateServiceRequiresNextVersion(t *testing.T) {
	service := newTransactionalNotificationTemplateServiceForTest(t)
	input := orderConfirmationTemplateInput("en", 1)
	_, err := service.Save(input)
	require.NoError(t, err)

	input.Version = 3
	_, err = service.Save(input)
	assert.ErrorIs(t, err, notification.ErrTemplateVersionConflict)
}

func TestTransactionalNotificationTemplateServiceNewTemplateMustStartAtVersionOne(t *testing.T) {
	service := newTransactionalNotificationTemplateServiceForTest(t)
	input := orderConfirmationTemplateInput("en", 2)
	_, err := service.Save(input)
	assert.ErrorIs(t, err, notification.ErrTemplateVersionConflict)
}

func TestTransactionalNotificationTemplateServiceRollbackCreatesNewVersion(t *testing.T) {
	service := newTransactionalNotificationTemplateServiceForTest(t)
	first := orderConfirmationTemplateInput("en", 1)
	first.SubjectTemplate = "Original {{order_number}}"
	_, err := service.Save(first)
	require.NoError(t, err)
	second := first
	second.Version = 2
	second.SubjectTemplate = "Changed {{order_number}}"
	_, err = service.Save(second)
	require.NoError(t, err)

	restored, err := service.Rollback(1, 1, nil, "restore original wording")
	require.NoError(t, err)
	assert.Equal(t, 3, restored.Version)
	assert.Equal(t, "Original {{order_number}}", restored.SubjectTemplate)

	versions, err := service.Versions(1)
	require.NoError(t, err)
	require.Len(t, versions, 3)
	assert.Equal(t, 3, versions[0].Version)
	assert.Equal(t, "restore original wording", versions[0].ChangeReason)
	assert.Equal(t, "Changed {{order_number}}", versions[1].SubjectTemplate)
	assert.Equal(t, "Original {{order_number}}", versions[2].SubjectTemplate)
}

func TestTransactionalNotificationTemplateServiceRollbackRejectsCurrentOrFutureVersion(t *testing.T) {
	service := newTransactionalNotificationTemplateServiceForTest(t)
	_, err := service.Save(orderConfirmationTemplateInput("en", 1))
	require.NoError(t, err)
	_, err = service.Rollback(1, 1, nil, "")
	assert.ErrorIs(t, err, notification.ErrTemplateVersionConflict)
}
