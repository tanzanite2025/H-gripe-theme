package repository

import (
	"errors"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"gorm.io/gorm"
)

// Public uploads made before private upload receipts are enabled have uploader
// identity, but no conversation receipt. Only their verified uploader can send
// them as attachments. This never makes them reusable card/library media.
func (r *TicketRepository) resolveVerifiedUploaderLegacyPublicConversationAttachment(reference string, actor media.MediaUploadActorIdentity) (*media.MediaAsset, error) {
	if actor.ActorType == "" || (actor.ActorType == media.MediaUploadActorVisitor && actor.VisitorSessionHash == "") || (actor.ActorType != media.MediaUploadActorVisitor && actor.AccountID == 0) {
		return nil, nil
	}
	repo := NewMediaRepository(r.db)
	if !repo.hasTable("media_assets") {
		return nil, nil
	}
	key, err := storage.TrustedBusinessMediaReferenceObjectKey(reference, nil, "")
	if err == nil && r.businessMediaReferenceResolver != nil {
		key, err = r.businessMediaReferenceResolver.ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference)
	}
	if err != nil {
		return nil, nil
	}
	asset, derivative, err := repo.FindUniquelyRegisteredMediaObjectByExactURLOrTrustedKey(reference, key)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if asset == nil || derivative != nil || asset.UploadOrigin != media.MediaUploadOriginCustomerServiceConversation || asset.DeletedAt.Valid || asset.Status != "active" || asset.Visibility != "public" || storage.IsPrivateObjectKey(asset.StorageKey) || storage.IsPrivateObjectKey(key) || dedicatedSiteIdentityOrVisualShowcaseNamespaceForObjectKey(asset.StorageKey) != "" || dedicatedSiteIdentityOrVisualShowcaseNamespaceForObjectKey(key) != "" {
		return nil, nil
	}
	validMediaType := asset.MediaType == "image" && (asset.MimeType == "image/jpeg" || asset.MimeType == "image/png" || asset.MimeType == "image/webp" || asset.MimeType == "image/gif")
	validMediaType = validMediaType || (asset.MediaType == "video" && actor.ActorType == media.MediaUploadActorAdministrator && (asset.MimeType == "video/mp4" || asset.MimeType == "video/webm" || asset.MimeType == "video/quicktime"))
	if !validMediaType {
		return nil, nil
	}
	uploader := media.MediaUploadActorIdentity{ActorType: asset.UploaderActorType, AccountID: asset.UploaderID, VisitorSessionHash: asset.UploaderVisitorSessionHash}
	if uploader != actor {
		return nil, nil
	}
	return asset, nil
}
