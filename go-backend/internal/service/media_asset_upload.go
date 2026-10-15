package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
)

var trustedMediaUploadImageRule = upload.FileRule{
	MaxSize:             12 << 20,
	AllowedExtensions:   []string{".jpg", ".jpeg", ".png", ".webp", ".gif"},
	AllowedContentTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
	MaxWidth:            8000,
	MaxHeight:           8000,
	MaxPixels:           24_000_000,
}

type MediaUploadInput struct {
	// Internal service context, selected by authorized routes; never bound from JSON/form origin fields.
	UploadOrigin        media.MediaAssetUploadOrigin
	UploadPurpose       string
	File                *multipart.FileHeader
	MediaType           string
	Alt                 string
	Caption             string
	UploaderID          uint
	UploadActorIdentity media.MediaUploadActorIdentity
	Width               int
	Height              int
	// StoragePrefix scopes an upload to a feature-owned storage directory.
	// Empty keeps the legacy storage layout unchanged.
	StoragePrefix string
	// Internal-only prepared receipt. The session and exact private object key
	// must already exist before writing bytes; callers cannot select visibility.
	PreparedCustomerServiceUploadSessionID   string
	PreparedCustomerServiceOriginalObjectKey string
}

func (s *MediaService) UploadAsset(ctx context.Context, input MediaUploadInput) (*media.MediaAsset, error) {
	if input.File == nil {
		return nil, ErrMediaUploadFileRequired
	}

	mediaType, err := normalizeMediaType(input.MediaType)
	if err != nil {
		return nil, err
	}
	input.MediaType = mediaType
	input.UploadPurpose = strings.TrimSpace(input.UploadPurpose)
	if mediaType == "video" {
		input.UploadPurpose = ""
	}
	identity, err := resolveMediaUploadActorIdentity(input)
	if err != nil {
		return nil, err
	}
	input.UploaderID = identity.AccountID
	input.UploadActorIdentity = identity
	classification, err := resolveMediaAssetClassificationFromTrustedUploadOrigin(input, identity)
	if err != nil {
		return nil, err
	}
	if classification.originalPrefix != "" && input.PreparedCustomerServiceUploadSessionID == "" {
		if input.StoragePrefix != "" && input.StoragePrefix != classification.originalPrefix {
			return nil, ErrUnsupportedMediaLibraryClassification
		}
		input.StoragePrefix = classification.originalPrefix
	}
	if err := s.ensureMediaUploadActorStorageQuota(identity, input.File.Size); err != nil {
		return nil, err
	}

	trustedMimeType, err := trustedMediaUploadMimeType(input.File, mediaType)
	if err != nil {
		return nil, err
	}

	contentSHA256, err := contentSHA256FromMultipartFile(input.File)
	if err != nil {
		return nil, err
	}

	var url string
	if input.PreparedCustomerServiceUploadSessionID != "" {
		if !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(input.PreparedCustomerServiceOriginalObjectKey) {
			return nil, ErrMediaAssetForbidden
		}
		writer, ok := s.storage.(storage.PreallocatedPrivateObjectWriter)
		if !ok {
			return nil, ErrMediaStorageUnavailable
		}
		session, sessionErr := s.repo.FindCustomerServiceMediaUploadSession(ctx, input.PreparedCustomerServiceUploadSessionID)
		if sessionErr != nil || session.ObjectKey != input.PreparedCustomerServiceOriginalObjectKey || session.Status != "uploading" || session.UploadActorIdentity() != identity ||
			!session.ExpiresAt.After(time.Now().UTC()) || !session.CustomerServiceAttachmentUploadWriteDeadline().After(time.Now().UTC()) {
			return nil, repository.ErrCustomerServiceAttachmentBindingUnavailable
		}
		var cancelPreparedUpload context.CancelFunc
		ctx, cancelPreparedUpload = context.WithDeadline(ctx, session.CustomerServiceAttachmentUploadWriteDeadline())
		defer cancelPreparedUpload()
		source, openErr := input.File.Open()
		if openErr != nil {
			return nil, openErr
		}
		url, err = writer.WritePrivateObjectAtPreallocatedKey(ctx, source, session.ObjectKey)
		_ = source.Close()
	} else if prefix := strings.TrimSpace(input.StoragePrefix); prefix != "" {
		url, err = s.storage.UploadWithPrefix(ctx, input.File, prefix)
	} else {
		url, err = s.storage.Upload(ctx, input.File)
	}
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(url) == "" {
		return nil, ErrMediaAssetURLUnavailable
	}

	copyrightClaim, err := s.buildCopyrightClaim(input, contentSHA256)
	if err != nil {
		_ = s.storage.Delete(ctx, url)
		return nil, err
	}

	asset := &media.MediaAsset{
		Filename:                    path.Base(url),
		OriginalFilename:            input.File.Filename,
		URL:                         url,
		StorageKey:                  storageObjectKey(s.storage, url),
		MimeType:                    trustedMimeType,
		MediaType:                   mediaType,
		Size:                        input.File.Size,
		Width:                       input.Width,
		Height:                      input.Height,
		ContentSHA256:               contentSHA256,
		CopyrightClaimJSON:          copyrightClaim,
		Alt:                         strings.TrimSpace(input.Alt),
		Caption:                     strings.TrimSpace(input.Caption),
		UploaderID:                  input.UploaderID,
		UploaderActorType:           identity.ActorType,
		UploaderVisitorSessionHash:  identity.VisitorSessionHash,
		Status:                      "active",
		Visibility:                  "public",
		LibraryPartition:            classification.partition,
		UploadOrigin:                classification.origin,
		ClassificationMethod:        classification.method,
		CustomerServiceMediaPurpose: classification.customerServicePurpose,
		UploadPurpose:               strings.TrimSpace(input.UploadPurpose),
	}
	if input.PreparedCustomerServiceUploadSessionID != "" {
		asset.Visibility = "private"
	}

	if err := s.repo.CreateAssetWithinUploadActorStorageQuota(ctx, asset, s.storageQuotaBytesForMediaUploadActor(identity), input.PreparedCustomerServiceUploadSessionID); err != nil {
		_ = s.storage.Delete(ctx, url)
		if errors.Is(err, repository.ErrMediaUploadActorStorageQuotaExceeded) {
			return nil, mediaStorageQuotaExceededErrorForUploadActor(identity)
		}
		return nil, err
	}

	derivatives, err := s.generateAssetDerivatives(ctx, asset, input.File)
	if err != nil {
		// Derivative conversion is best-effort. Keep the original asset and its
		// database record intact when conversion times out or fails; a rebuild
		// worker can generate missing variants later without losing the upload.
		s.hydrateAssetAccessURL(asset)
		return asset, nil
	}
	if len(derivatives) > 0 {
		if err := s.repo.CreateAssetDerivatives(derivatives); err != nil {
			derivativeURLs := make([]string, 0, len(derivatives))
			for _, derivative := range derivatives {
				derivativeURLs = append(derivativeURLs, derivative.URL)
			}
			deleteUploadedMediaObjectsBestEffort(ctx, s.storage, derivativeURLs)
			// Preserve the source asset even when derivative metadata persistence
			// fails; cleanup of derivative objects is still attempted above.
			return nil, err
		}
		asset.Derivatives = derivatives
	}

	s.hydrateAssetAccessURL(asset)
	return asset, nil
}

func trustedMediaUploadMimeType(file *multipart.FileHeader, mediaType string) (string, error) {
	var rule upload.FileRule
	switch mediaType {
	case "image":
		rule = trustedMediaUploadImageRule
	case "video":
		rule = upload.ProductVideoRule
	default:
		return "", ErrUnsupportedMediaType
	}

	if err := upload.ValidateFile(file, rule); err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnsupportedMediaType, err)
	}

	mimeType, err := upload.DetectContentType(file)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnsupportedMediaType, err)
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(mimeType)), mediaType+"/") {
		return "", fmt.Errorf("%w: detected %s for %s upload", ErrUnsupportedMediaType, mimeType, mediaType)
	}
	return mimeType, nil
}
