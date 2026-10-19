package repository

import (
	"context"
	"errors"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
)

var ErrCustomerServiceAttachmentBindingUnavailable = errors.New("customer-service attachment is unavailable for binding")

type resolvedCustomerServiceMessageAttachmentRequest struct {
	assetID   uint
	sessionID string
}

type resolvedCustomerServiceMessageAttachmentPlan struct {
	requests []resolvedCustomerServiceMessageAttachmentRequest
	assetIDs []uint
}

func (r *MediaRepository) CreateCustomerServicePendingUploadSession(ctx context.Context, session *media.MediaUploadSession) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if session == nil {
		return ErrCustomerServiceAttachmentBindingUnavailable
	}
	key, valid := storage.NormalizeObjectKey(session.ObjectKey)
	if !valid || key != session.ObjectKey || !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(key) {
		return ErrCustomerServiceAttachmentBindingUnavailable
	}
	return r.WithinMediaAssetTransactionWithRequestContext(ctx, func(repo *MediaRepository, tx *gorm.DB) error {
		// A receipt precedes the first byte, so admission shares cleanup's
		// object lock without requiring an already-existing provider object.
		// The enclosing conversation transaction must commit before upload.
		if err := repo.lockMediaObjectKeysAndPrivateDerivativeNamespacesWithinTransaction([]string{key}, nil); err != nil {
			return err
		}
		return tx.Create(session).Error
	})
}

func (r *MediaRepository) FindCustomerServiceMediaUploadSession(ctx context.Context, id string) (*media.MediaUploadSession, error) {
	var session media.MediaUploadSession
	if err := r.db.WithContext(ctx).First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *MediaRepository) MarkCustomerServiceUploadSessionFailed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&media.MediaUploadSession{}).Where("id = ? AND status = ?", id, "uploading").Update("status", "failed").Error
}

func (r *MediaRepository) RegisterCustomerServiceUploadSessionAssetWithinTransaction(sessionID string, asset *media.MediaAsset) error {
	var session media.MediaUploadSession
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&session, "id = ?", sessionID).Error; err != nil {
		return err
	}
	identity := media.MediaUploadActorIdentity{ActorType: asset.UploaderActorType, AccountID: asset.UploaderID, VisitorSessionHash: asset.UploaderVisitorSessionHash}
	if session.Status != "uploading" || !session.ExpiresAt.After(time.Now().UTC()) || session.UploadActorIdentity() != identity ||
		session.ObjectKey != asset.StorageKey || asset.Visibility != "private" ||
		!storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(asset.StorageKey) {
		return ErrCustomerServiceAttachmentBindingUnavailable
	}
	result := r.db.Model(&media.MediaUploadSession{}).Where("id = ? AND status = ?", sessionID, "uploading").Updates(map[string]interface{}{"media_asset_id": asset.ID, "status": "ready"})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrCustomerServiceAttachmentBindingUnavailable
	}
	return nil
}

func (r *TicketRepository) FindCustomerServiceMessageMediaAssetBindings(messageID uint) ([]ticket.TicketMessageMediaAttachment, error) {
	var bindings []ticket.TicketMessageMediaAttachment
	err := r.db.Where("ticket_message_id = ?", messageID).Order("attachment_index ASC").Find(&bindings).Error
	return bindings, err
}

// Called inside the existing conversation/message/Outbox transaction. Asset
// rows are locked in ascending ID order, matching media deletion's row locks.
func (r *TicketRepository) BindCustomerServiceMessageMediaAssetsWithinTransaction(
	conversationID, messageID uint, identity media.MediaUploadActorIdentity, requests []ticket.CustomerServiceMessageMediaAttachmentBindingRequest,
) error {
	plan, err := r.resolveCustomerServiceMessageAttachmentRequestsWithinTransaction(identity, requests)
	if err != nil {
		return err
	}
	return r.bindResolvedCustomerServiceMessageAttachmentsWithinTransaction(conversationID, messageID, identity, plan)
}

func (r *TicketRepository) resolveCustomerServiceMessageAttachmentRequestsWithinTransaction(identity media.MediaUploadActorIdentity, requests []ticket.CustomerServiceMessageMediaAttachmentBindingRequest) (*resolvedCustomerServiceMessageAttachmentPlan, error) {
	if len(requests) == 0 || len(requests) > 4 {
		return nil, ErrCustomerServiceAttachmentBindingUnavailable
	}
	if identity.ActorType == media.MediaUploadActorVisitor && len(requests) > 1 {
		return nil, ErrCustomerServiceAttachmentBindingUnavailable
	}
	resolved := make([]resolvedCustomerServiceMessageAttachmentRequest, len(requests))
	assetIDs := make([]uint, 0, len(requests))
	seen := make(map[uint]bool, len(requests))
	for index, request := range requests {
		if (request.MediaAssetID == 0) == (request.UploadSessionID == "") {
			return nil, ErrCustomerServiceAttachmentBindingUnavailable
		}
		resolved[index] = resolvedCustomerServiceMessageAttachmentRequest{assetID: request.MediaAssetID, sessionID: request.UploadSessionID}
		if request.UploadSessionID != "" {
			var session media.MediaUploadSession
			if err := r.db.First(&session, "id = ?", request.UploadSessionID).Error; err != nil {
				return nil, ErrCustomerServiceAttachmentBindingUnavailable
			}
			if session.MediaAssetID == nil {
				return nil, ErrCustomerServiceAttachmentBindingUnavailable
			}
			resolved[index].assetID = *session.MediaAssetID
		}
		if seen[resolved[index].assetID] {
			return nil, ErrCustomerServiceAttachmentBindingUnavailable
		}
		seen[resolved[index].assetID] = true
		assetIDs = append(assetIDs, resolved[index].assetID)
	}
	sort.Slice(assetIDs, func(i, j int) bool { return assetIDs[i] < assetIDs[j] })
	return &resolvedCustomerServiceMessageAttachmentPlan{requests: resolved, assetIDs: assetIDs}, nil
}

func (r *TicketRepository) bindResolvedCustomerServiceMessageAttachmentsWithinTransaction(conversationID, messageID uint, identity media.MediaUploadActorIdentity, plan *resolvedCustomerServiceMessageAttachmentPlan) error {
	assetIDs, resolved := plan.assetIDs, plan.requests
	assets := make(map[uint]media.MediaAsset, len(assetIDs))
	for _, assetID := range assetIDs {
		var asset media.MediaAsset
		if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&asset, assetID).Error; err != nil {
			return ErrCustomerServiceAttachmentBindingUnavailable
		}
		if asset.Status != "active" || (asset.MediaType != "image" && asset.MediaType != "video") {
			return ErrCustomerServiceAttachmentBindingUnavailable
		}
		switch asset.MimeType {
		case "image/jpeg", "image/png", "image/webp", "image/gif":
			if asset.MediaType != "image" {
				return ErrCustomerServiceAttachmentBindingUnavailable
			}
		case "video/mp4", "video/webm", "video/quicktime":
			if asset.MediaType != "video" || identity.ActorType != media.MediaUploadActorAdministrator {
				return ErrCustomerServiceAttachmentBindingUnavailable
			}
		default:
			return ErrCustomerServiceAttachmentBindingUnavailable
		}
		assets[assetID] = asset
	}
	for index, request := range resolved {
		asset := assets[request.assetID]
		binding := ticket.TicketMessageMediaAttachment{TicketMessageID: messageID, MediaAssetID: asset.ID, AttachmentIndex: index}
		if request.sessionID == "" {
			if asset.Visibility != "public" || storage.IsPrivateObjectKey(asset.StorageKey) {
				return ErrCustomerServiceAttachmentBindingUnavailable
			}
		} else {
			var session media.MediaUploadSession
			if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&session, "id = ?", request.sessionID).Error; err != nil {
				return ErrCustomerServiceAttachmentBindingUnavailable
			}
			assetIdentity := media.MediaUploadActorIdentity{ActorType: asset.UploaderActorType, AccountID: asset.UploaderID, VisitorSessionHash: asset.UploaderVisitorSessionHash}
			if session.BusinessContextType != media.MediaUploadSessionCustomerServiceContext || session.BusinessContextID != conversationID ||
				session.Status != "ready" || !session.ExpiresAt.After(time.Now().UTC()) ||
				session.UploadActorIdentity() != identity || session.UploadActorIdentity() != assetIdentity ||
				session.MediaAssetID == nil || *session.MediaAssetID != asset.ID || session.ObjectKey != asset.StorageKey ||
				asset.Visibility != "private" || !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(asset.StorageKey) {
				return ErrCustomerServiceAttachmentBindingUnavailable
			}
			binding.UploadSessionID = &request.sessionID
			if err := r.db.Model(&media.MediaUploadSession{}).Where("id = ?", session.ID).Update("status", "bound").Error; err != nil {
				return err
			}
		}
		if err := r.db.Create(&binding).Error; err != nil {
			return err
		}
	}
	return nil
}
