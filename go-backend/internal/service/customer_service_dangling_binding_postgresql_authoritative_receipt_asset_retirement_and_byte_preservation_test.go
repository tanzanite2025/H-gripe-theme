package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
)

func TestPostgreSQLCustomerServiceDanglingBindingAndBoundReceiptBlockAssetRetirementUntilBothOwnersAreRemoved(t *testing.T) {
	env, _, asset := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	// Simulate a historical missing message by relaxing this FK only in the
	// fixture's newly created isolated schema, never on a business database.
	require.NoError(t, env.db.Exec("ALTER TABLE ticket_message_media_attachments DROP CONSTRAINT ticket_message_media_attachments_ticket_message_id_fkey").Error)
	var session media.MediaUploadSession
	require.NoError(t, env.db.Where("media_asset_id = ?", asset.ID).First(&session).Error)
	require.NoError(t, env.db.Model(&session).Update("status", "bound").Error)
	binding := ticket.TicketMessageMediaAttachment{TicketMessageID: 876543, MediaAssetID: asset.ID, UploadSessionID: &session.ID}
	require.NoError(t, env.db.Create(&binding).Error)
	storedPath := filepath.Join(env.privateRoot, filepath.FromSlash(asset.StorageKey))
	beforeBytes, err := os.ReadFile(storedPath)
	require.NoError(t, err)
	assertOwnersPreventRetirement := func(t *testing.T, expectedReferences int) {
		t.Helper()
		references, err := env.mediaService.repo.FindAssetReferences(asset)
		require.NoError(t, err)
		require.Len(t, references, expectedReferences)
		require.ErrorIs(t, env.mediaService.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
		afterBytes, err := os.ReadFile(storedPath)
		require.NoError(t, err)
		require.Equal(t, beforeBytes, afterBytes)
		var persisted media.MediaAsset
		require.NoError(t, env.db.First(&persisted, asset.ID).Error)
		require.Equal(t, "active", persisted.Status)
	}
	assertOwnersPreventRetirement(t, 2)
	require.NoError(t, env.db.Delete(&binding).Error)
	assertOwnersPreventRetirement(t, 1)
	// The receipt independently owns bytes even after the broken binding is removed.
	require.NoError(t, env.db.Delete(&session).Error)
	references, err := env.mediaService.repo.FindAssetReferences(asset)
	require.NoError(t, err)
	require.Empty(t, references)
}

func TestPostgreSQLCustomerServiceDanglingReferencesDistinguishDeletedConversationAndRestoredParent(t *testing.T) {
	env, _, asset := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	message := ticket.TicketMessage{TicketID: env.conversation.ID, Content: "Synthetic retained message", MessageType: "text", AttachmentReferenceFormat: ticket.TicketMessageMediaAssetBindings}
	require.NoError(t, env.db.Create(&message).Error)
	require.NoError(t, env.db.Create(&ticket.TicketMessageMediaAttachment{TicketMessageID: message.ID, MediaAssetID: asset.ID}).Error)
	require.NoError(t, env.db.Model(&ticket.Ticket{}).Where("id = ?", env.conversation.ID).Update("deleted_at", time.Now()).Error)
	references, err := env.mediaService.repo.FindAssetReferences(asset)
	require.NoError(t, err)
	require.Len(t, references, 2)
	for _, reference := range references {
		require.Equal(t, "retained_deleted_parent", reference.UsageState)
	}
	report, err := env.mediaService.DiagnoseMediaReferenceAssociationsWithPermissionScopedPresentation(t.Context(), asset.ID, customerServiceAttachmentReferenceDiagnosisScope())
	require.NoError(t, err)
	require.Len(t, report.Entries, 2)
	for _, entry := range report.Entries {
		require.Equal(t, "deleted_history_must_remain_retained", entry.Diagnosis)
	}
	require.NoError(t, env.db.Unscoped().Model(&ticket.Ticket{}).Where("id = ?", env.conversation.ID).Update("deleted_at", nil).Error)
	references, err = env.mediaService.repo.FindAssetReferences(asset)
	require.NoError(t, err)
	require.Len(t, references, 2)
	for _, reference := range references {
		if reference.ResourceType == "ticket_message" {
			require.Equal(t, "active", reference.UsageState)
		} else {
			require.Equal(t, "pending_upload", reference.UsageState)
		}
	}
}
