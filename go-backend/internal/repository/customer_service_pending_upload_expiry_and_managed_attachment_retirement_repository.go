package repository

import (
	"context"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
)

func (r *MediaRepository) FindExpiredUnboundCustomerServiceUploadSessions(ctx context.Context, now time.Time, limit int) ([]media.MediaUploadSession, error) {
	var sessions []media.MediaUploadSession
	if limit <= 0 || !r.hasTable("media_upload_sessions") {
		return sessions, nil
	}
	if limit > 1000 {
		limit = 1000
	}
	err := r.db.WithContext(ctx).Where("business_context_type = ? AND status IN ? AND expires_at <= ? AND created_at <= ?",
		media.MediaUploadSessionCustomerServiceContext, []string{"uploading", "ready", "failed"}, now, now.Add(-media.CustomerServiceAttachmentUploadMaximumWriteDuration)).
		Order("expires_at ASC, id ASC").Limit(limit).Find(&sessions).Error
	return sessions, err
}

func (r *MediaRepository) FindCustomerServiceUploadSessionForUpdate(ctx context.Context, id string) (*media.MediaUploadSession, error) {
	var session media.MediaUploadSession
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *MediaRepository) FindCustomerServiceAttachmentAssetUnscopedForUpdate(id uint) (*media.MediaAsset, error) {
	var asset media.MediaAsset
	query := r.db.Unscoped().Clauses(clause.Locking{Strength: "UPDATE"})
	if r.hasAssetDerivativesTable() {
		query = query.Preload("Derivatives", func(db *gorm.DB) *gorm.DB { return db.Unscoped() })
	}
	if err := query.First(&asset, id).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *MediaRepository) CountCustomerServiceMessageBindingsForUploadSession(id string) (int64, error) {
	var count int64
	if !r.hasTable("ticket_message_media_attachments") {
		return 0, nil
	}
	err := r.db.Model(&ticket.TicketMessageMediaAttachment{}).Where("upload_session_id = ?", id).Count(&count).Error
	return count, err
}

func (r *MediaRepository) DeleteRetiredCustomerServiceUploadSession(id string) error {
	return r.db.Delete(&media.MediaUploadSession{}, "id = ?", id).Error
}

func (r *TicketRepository) FindCustomerServiceConversationManagedUploadSessions(ticketID uint) ([]media.MediaUploadSession, error) {
	var sessions []media.MediaUploadSession
	if !r.db.Migrator().HasTable("media_upload_sessions") {
		return sessions, nil
	}
	err := r.db.Where("business_context_type = ? AND business_context_id = ?", media.MediaUploadSessionCustomerServiceContext, ticketID).
		Order("id ASC").Find(&sessions).Error
	return sessions, err
}

func (r *TicketRepository) FindCustomerServiceConversationBoundMediaAssetIDs(ticketID uint) ([]uint, error) {
	var ids []uint
	if !r.db.Migrator().HasTable("ticket_message_media_attachments") {
		return ids, nil
	}
	err := r.db.Table("ticket_message_media_attachments AS bindings").Distinct("bindings.media_asset_id").
		Joins("JOIN ticket_messages AS messages ON messages.id=bindings.ticket_message_id").Where("messages.ticket_id = ?", ticketID).
		Pluck("bindings.media_asset_id", &ids).Error
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, err
}

// The parent and all related asset rows must already be locked. Remove
// relations first to satisfy RESTRICT keys, then recheck surviving references.
func (r *TicketRepository) DeleteCustomerServiceConversationManagedAttachmentsAndMessages(ticketID uint) error {
	if r.db.Migrator().HasTable("ticket_message_media_attachments") {
		messageIDs := r.db.Model(&ticket.TicketMessage{}).Select("id").Where("ticket_id = ?", ticketID)
		if err := r.db.Where("ticket_message_id IN (?)", messageIDs).Delete(&ticket.TicketMessageMediaAttachment{}).Error; err != nil {
			return err
		}
	}
	if r.db.Migrator().HasTable("media_upload_sessions") {
		if err := r.db.Where("business_context_type = ? AND business_context_id = ?", media.MediaUploadSessionCustomerServiceContext, ticketID).
			Delete(&media.MediaUploadSession{}).Error; err != nil {
			return err
		}
	}
	return r.db.Where("ticket_id = ?", ticketID).Delete(&ticket.TicketMessage{}).Error
}

func (r *MediaRepository) HasCustomerServiceUploadOrRegisteredMediaAssetAtStorageKey(key string) (bool, error) {
	var count int64
	if err := r.db.Unscoped().Model(&media.MediaAsset{}).Where("storage_key = ?", key).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	if r.hasTable("media_upload_sessions") {
		if err := r.db.Model(&media.MediaUploadSession{}).Where("object_key = ?", key).Count(&count).Error; err != nil {
			return false, err
		}
	}
	return count > 0, nil
}

func (r *MediaRepository) ReleaseCustomerServiceDerivativeWriteLease(ctx context.Context, assetID uint, token string) error {
	return r.db.WithContext(ctx).Model(&media.MediaAsset{}).Where("id = ? AND customer_service_derivative_write_token = ?", assetID, token).
		Updates(map[string]interface{}{"customer_service_derivative_write_token": "", "customer_service_derivative_write_expires_at": nil}).Error
}

func (r *MediaRepository) WithinMediaAssetTransactionWithRequestContext(ctx context.Context, operation func(*MediaRepository, *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return operation(r.WithTx(tx), tx) })
}

func (r *TicketRepository) HasCustomerServiceConversationAttachmentWritesInProgress(ticketID uint, now time.Time) (bool, error) {
	var count int64
	if r.db.Migrator().HasTable("media_upload_sessions") {
		if err := r.db.Model(&media.MediaUploadSession{}).Where("business_context_type = ? AND business_context_id = ? AND created_at > ?",
			media.MediaUploadSessionCustomerServiceContext, ticketID, now.Add(-media.CustomerServiceAttachmentUploadMaximumWriteDuration)).Count(&count).Error; err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	assetIDs, err := r.FindCustomerServiceConversationBoundMediaAssetIDs(ticketID)
	if err != nil {
		return false, err
	}
	sessions, err := r.FindCustomerServiceConversationManagedUploadSessions(ticketID)
	if err != nil {
		return false, err
	}
	for _, session := range sessions {
		if session.MediaAssetID != nil {
			assetIDs = append(assetIDs, *session.MediaAssetID)
		}
	}
	if len(assetIDs) == 0 {
		return false, nil
	}
	err = r.db.Unscoped().Model(&media.MediaAsset{}).Where("id IN ? AND customer_service_derivative_write_expires_at > ?", assetIDs, now).Count(&count).Error
	return count > 0, err
}
