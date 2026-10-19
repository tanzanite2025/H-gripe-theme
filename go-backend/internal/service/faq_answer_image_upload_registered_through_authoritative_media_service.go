package service

import (
	"context"
	"mime/multipart"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
)

var ErrFAQMediaReferenceInvalid = repository.ErrFAQMediaReferenceInvalid

// The authorized FAQ route supplies the actor; multipart identity and origin
// fields never participate in classification or storage quota ownership.
func (s *FAQService) UploadAnswerImage(ctx context.Context, file *multipart.FileHeader, actor media.MediaUploadActorIdentity, alt string) (string, error) {
	if actor.ActorType != media.MediaUploadActorAdministrator || !actor.IsValidMediaUploadActorIdentity() {
		return "", ErrMediaUploadIdentityRequired
	}
	if s.mediaService == nil {
		return "", ErrMediaStorageUnavailable
	}
	if err := upload.ValidateSpecFile(file, string(upload.SpecFAQAnswerImage)); err != nil {
		return "", err
	}
	width, height, err := upload.ReadImageDimensions(file)
	if err != nil {
		return "", err
	}
	asset, err := s.mediaService.UploadBusinessDraftMediaAsset(ctx, MediaUploadInput{
		UploadOrigin:  media.MediaUploadOriginAdministratorContent,
		UploadPurpose: string(upload.SpecFAQAnswerImage),
		File:          file, MediaType: "image", Alt: strings.TrimSpace(alt),
		UploadActorIdentity: actor, Width: width, Height: height,
	})
	if err != nil {
		return "", err
	}
	return asset.URL, nil
}
