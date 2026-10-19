package service

import (
	"bytes"
	"context"
	"fmt"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
)

func mediaAssetDerivativeStoragePrefixAndCachePolicy(asset *media.MediaAsset, preset string, version int) (string, string, error) {
	isConversationAttachment := storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(assetStorageKey(asset))
	if asset.Visibility == "private" && isConversationAttachment {
		return fmt.Sprintf("%s/%d/%s/v%d", storage.CustomerServiceConversationAttachmentDerivativesPrefix, asset.ID, preset, normalizeMediaDerivativeVersion(version)),
			"private, no-store", nil
	}
	if asset.Visibility != "public" || storage.IsPrivateObjectKey(assetStorageKey(asset)) {
		return "", "", fmt.Errorf("%w: private media requires an explicit derivative storage policy", ErrMediaDerivativeGenerationFailed)
	}
	return derivativePrefixForAsset(asset.ID, preset, version), mediaDerivativeCacheControl, nil
}

func uploadMediaDerivativeObjectUsingAssetAccessPolicy(ctx context.Context, storageService storage.StorageService, asset *media.MediaAsset, reader *bytes.Reader, filename, preset string, version int) (string, error) {
	prefix, cacheControl, err := mediaAssetDerivativeStoragePrefixAndCachePolicy(asset, preset, version)
	if err != nil {
		return "", err
	}
	if asset.Visibility == "private" {
		uploader, ok := storageService.(storage.PrivateCacheControlledReaderUploader)
		if !ok {
			return "", ErrMediaStorageUnavailable
		}
		return uploader.UploadPrivateObjectFromReaderWithPrefixAndCacheControl(ctx, reader, filename, prefix, cacheControl)
	}
	// Public generations must use the precommitted batch receipt protocol.
	return "", ErrMediaStorageUnavailable
}
