package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"

	"gorm.io/gorm"
)

func (s *HomeVisualTileService) uploadVisualShowcaseDraftImageWithPrecommittedExactKeyReceipt(ctx context.Context, tileSetKey, locale string, file *multipart.FileHeader, width, height int) (*HomeVisualTileImageUpload, error) {
	writer, available := s.storage.(storage.PreallocatedVisualShowcaseDraftObjectWriter)
	if !available || s.outbox == nil || s.repo == nil {
		return nil, ErrObjectStorageCleanupUnavailable
	}
	key, err := storage.AllocateVisualShowcaseDraftUploadObjectKey(homeVisualTileStoragePrefix(tileSetKey, locale), file.Filename)
	if err != nil {
		return nil, err
	}
	id, _ := storage.VisualShowcaseDraftUploadReceiptIDFromObjectKey(key)
	expectedURL := s.storage.GetURL(key)
	resolved, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, HomeVisualTileStoragePrefix, expectedURL)
	if err != nil || resolved != key {
		return nil, ErrHomeVisualTileImageInvalid
	}
	digest, err := contentSHA256FromMultipartFile(file)
	if err != nil {
		return nil, err
	}
	receipt, err := newObjectStorageCleanupEvent(objectCleanupResourceHomeVisualTile, id, outbox.AggregateTypeHomeVisualTileSet, tileSetKey+":"+locale, []string{key})
	if err != nil {
		return nil, err
	}
	receipt.EventKey = homevisualtile.VisualShowcaseDraftUploadReceiptEventKeyPrefix + id
	var payload homevisualtile.VisualShowcaseDraftUploadCleanupReceiptPayload
	// The original resource and object list remain understood by the existing
	// namespace/owner/history-aware cleanup handler.
	if err := json.Unmarshal(receipt.Payload, &payload.ObjectStorageCleanupPayload); err != nil {
		return nil, err
	}
	payload.Size, payload.ContentSHA256 = file.Size, digest
	err = s.outbox.RunVisualShowcaseDraftUploadWithPrecommittedExactKeyCleanupReceipt(ctx, receipt, &payload, func(tx *gorm.DB) error {
		reader, err := file.Open()
		if err != nil {
			return err
		}
		defer func() { _ = reader.Close() }()
		url, err := writer.WriteVisualShowcaseDraftObjectAtPreallocatedKey(tx.Statement.Context, reader, key)
		if err != nil {
			return err
		}
		returnedKey, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, HomeVisualTileStoragePrefix, url)
		if err != nil || returnedKey != key {
			return ErrHomeVisualTileImageInvalid
		}
		return s.verifyPreallocatedVisualShowcaseDraftObjectSizeAndDigest(tx.Statement.Context, key, file.Size, digest)
	})
	if err != nil {
		return nil, err
	}
	return &HomeVisualTileImageUpload{ImageURL: expectedURL, ThumbnailURL: expectedURL, StorageKey: key, Width: width, Height: height, DraftExpiresAt: payload.DraftExpiresAt}, nil
}

func (s *HomeVisualTileService) verifyPreallocatedVisualShowcaseDraftObjectSizeAndDigest(ctx context.Context, key string, size int64, expectedDigest string) error {
	opener, available := s.storage.(storage.ObjectOpener)
	if !available {
		return ErrHomeVisualTileStorageUnavailable
	}
	object, err := opener.Open(ctx, key)
	if err != nil {
		return err
	}
	if object == nil || object.ReadCloser == nil {
		return ErrHomeVisualTileImageInvalid
	}
	digest := sha256.New()
	copied, readErr := io.Copy(digest, io.LimitReader(object.ReadCloser, size+1))
	closeErr := object.ReadCloser.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return err
	}
	if copied != size || fmt.Sprintf("%x", digest.Sum(nil)) != expectedDigest {
		return fmt.Errorf("%w: uploaded image integrity mismatch", ErrHomeVisualTileImageInvalid)
	}
	return ctx.Err()
}
