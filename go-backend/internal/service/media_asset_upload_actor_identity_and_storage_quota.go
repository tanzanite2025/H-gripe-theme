package service

import (
	"fmt"

	"commerce-platform/internal/domain/media"
)

const DefaultMediaVisitorSessionStorageQuotaBytes int64 = 20 << 20

type MediaUploadActorQuotaOptions struct {
	VisitorSessionStorageQuotaBytes int64
}

func resolveMediaUploadActorIdentity(input MediaUploadInput) (media.MediaUploadActorIdentity, error) {
	identity := input.UploadActorIdentity
	if identity.ActorType == "" && identity.AccountID == 0 && identity.VisitorSessionHash == "" {
		identity = media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorLegacyAccount, AccountID: input.UploaderID}
	}
	if !identity.IsValidMediaUploadActorIdentity() || (input.UploaderID != 0 && input.UploaderID != identity.AccountID) {
		return media.MediaUploadActorIdentity{}, ErrMediaUploadIdentityRequired
	}
	return identity, nil
}

func (s *MediaService) storageQuotaBytesForMediaUploadActor(identity media.MediaUploadActorIdentity) int64 {
	if identity.ActorType == media.MediaUploadActorVisitor {
		return s.visitorSessionStorageQuotaBytes
	}
	return s.accountStorageQuotaBytes
}

func mediaStorageQuotaExceededErrorForUploadActor(identity media.MediaUploadActorIdentity) error {
	if identity.ActorType == media.MediaUploadActorVisitor {
		return ErrMediaVisitorSessionStorageQuotaExceeded
	}
	return ErrMediaAccountStorageQuotaExceeded
}

func (s *MediaService) ensureMediaUploadActorStorageQuota(identity media.MediaUploadActorIdentity, incomingSize int64) error {
	if incomingSize <= 0 {
		return ErrMediaUploadFileRequired
	}
	quotaBytes := s.storageQuotaBytesForMediaUploadActor(identity)
	if quotaBytes <= 0 {
		return mediaStorageQuotaExceededErrorForUploadActor(identity)
	}
	used, err := s.repo.AssetStorageUsageByUploadActor(identity)
	if err != nil {
		return fmt.Errorf("get media upload actor storage usage: %w", err)
	}
	if used >= quotaBytes || incomingSize > quotaBytes-used {
		return mediaStorageQuotaExceededErrorForUploadActor(identity)
	}
	return nil
}
