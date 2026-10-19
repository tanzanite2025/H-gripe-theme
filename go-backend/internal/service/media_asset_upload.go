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
	// Permanent originals use a unique preallocated subdirectory within this scope.
	StoragePrefix string
	// Internal-only prepared receipt. The session and exact private object key
	// must already exist before writing bytes; callers cannot select visibility.
	PreparedCustomerServiceUploadSessionID   string
	PreparedCustomerServiceOriginalObjectKey string
}

func (s *MediaService) UploadAsset(ctx context.Context, input MediaUploadInput) (*media.MediaAsset, error) {
	if s == nil || s.repo == nil || s.storage == nil {
		return nil, ErrMediaStorageUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if input.File == nil {
		return nil, ErrMediaUploadFileRequired
	}

	mediaType, err := normalizeMediaType(input.MediaType)
	if err != nil {
		return nil, err
	}
	input.MediaType = mediaType
	if prefix := strings.TrimSpace(input.StoragePrefix); prefix == storage.BusinessDraftMediaObjectPrefix || strings.HasPrefix(prefix, storage.BusinessDraftMediaObjectPrefix+"/") {
		return nil, ErrMediaAssetForbidden
	}
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

	if input.PreparedCustomerServiceUploadSessionID == "" {
		return s.uploadPermanentMediaOriginalWithPrecommittedReceiptAndAtomicAssetRegistration(ctx, input, classification, trustedMimeType, contentSHA256)
	}

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
	url, err := writer.WritePrivateObjectAtPreallocatedKey(ctx, source, session.ObjectKey)
	_ = source.Close()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(url) == "" {
		return nil, ErrMediaAssetURLUnavailable
	}

	copyrightClaim, err := s.buildCopyrightClaim(input, contentSHA256)
	if err != nil {
		// Prepared private originals already have a durable upload receipt;
		// their owning expiry lifecycle handles failed registration.
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
		Visibility:                  "private",
		LibraryPartition:            classification.partition,
		UploadOrigin:                classification.origin,
		ClassificationMethod:        classification.method,
		CustomerServiceMediaPurpose: classification.customerServicePurpose,
		UploadPurpose:               strings.TrimSpace(input.UploadPurpose),
	}

	if err := s.repo.CreateAssetWithinUploadActorStorageQuota(ctx, asset, s.storageQuotaBytesForMediaUploadActor(identity), input.PreparedCustomerServiceUploadSessionID); err != nil {
		if errors.Is(err, repository.ErrMediaUploadActorStorageQuotaExceeded) {
			err = mediaStorageQuotaExceededErrorForUploadActor(identity)
		}
		return nil, err
	}

	return s.finishRegisteredMediaAssetUploadWithBestEffortDerivatives(ctx, input, asset)
}

func (s *MediaService) finishRegisteredMediaAssetUploadWithBestEffortDerivatives(ctx context.Context, input MediaUploadInput, asset *media.MediaAsset) (*media.MediaAsset, error) {
	derivatives, err := s.generateAssetDerivatives(ctx, asset, input.File)
	if err != nil {
		if errors.Is(err, ErrObjectStorageCleanupUnavailable) || errors.Is(err, ErrPublicMediaDerivativeRegistrationExchangeFailed) {
			return nil, err
		}
		// Derivative conversion is best-effort. Keep the original asset and its
		// database record intact when conversion times out or fails; a rebuild
		// worker can generate missing variants later without losing the upload.
		s.hydrateAssetAccessURL(asset)
		return asset, nil
	}
	if len(derivatives) > 0 {
		if asset.Visibility != "private" {
			asset.Derivatives = derivatives
			s.hydrateAssetAccessURL(asset)
			return asset, nil
		}
		if err := s.repo.CreateAssetDerivatives(derivatives); err != nil {
			derivativeURLs := make([]string, 0, len(derivatives))
			for _, derivative := range derivatives {
				derivativeURLs = append(derivativeURLs, derivative.URL)
			}
			// Preserve the source asset even when derivative metadata persistence
			// fails; persist compensation before attempting object cleanup.
			return nil, errors.Join(err, s.persistFailedMediaUploadCompensationBeforeAttemptingExactObjectCleanup(ctx, asset, derivativeURLs))
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
