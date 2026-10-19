package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPostgreSQLCustomerServiceDanglingBindingDiagnosisSnapshotRetainsBindingDuringConcurrentMessageRecreation(t *testing.T) {
	repo := createIsolatedPostgreSQLMediaUploadActorQuotaRepository(t)
	// Deliberately incomplete legacy rows exist only inside this random schema.
	for _, statement := range []string{
		"CREATE TABLE tickets(id BIGINT PRIMARY KEY,deleted_at TIMESTAMPTZ)",
		"CREATE TABLE ticket_messages(id BIGINT PRIMARY KEY,ticket_id BIGINT)",
		"CREATE TABLE ticket_message_media_attachments(id BIGINT PRIMARY KEY,ticket_message_id BIGINT,media_asset_id BIGINT)",
		"INSERT INTO ticket_message_media_attachments VALUES(17,19,23)",
	} {
		require.NoError(t, repo.db.Exec(statement).Error)
	}
	query := mediaAssetReferenceQuery{AssetID: 23}
	require.NoError(t, repo.ReadMediaAssociationDiagnosisSnapshot(t.Context(), func(snapshot *MediaRepository, tx *gorm.DB) error {
		before, err := snapshot.customerServiceMessageMediaAssetBindingReferences(query)
		require.NoError(t, err)
		require.Len(t, before, 1)
		require.Equal(t, "ticket_message_attachment_binding", before[0].ResourceType)
		require.Equal(t, "missing_parent", before[0].UsageState)
		require.Zero(t, before[0].ParentResourceID)
		require.NoError(t, repo.db.WithContext(t.Context()).Transaction(func(writer *gorm.DB) error {
			if err := writer.Exec("INSERT INTO tickets VALUES(29,NULL)").Error; err != nil {
				return err
			}
			return writer.Exec("INSERT INTO ticket_messages VALUES(19,29)").Error
		}))
		after, err := snapshot.customerServiceMessageMediaAssetBindingReferences(query)
		require.NoError(t, err)
		require.Equal(t, before, after)
		return nil
	}))
	current, err := repo.customerServiceMessageMediaAssetBindingReferences(query)
	require.NoError(t, err)
	require.Len(t, current, 1)
	require.Equal(t, "ticket_message", current[0].ResourceType)
	require.Equal(t, "active", current[0].UsageState)
	require.EqualValues(t, 29, current[0].ParentResourceID)
	var bindingCount int64
	require.NoError(t, repo.db.Table("ticket_message_media_attachments").Count(&bindingCount).Error)
	require.EqualValues(t, 1, bindingCount)
}
