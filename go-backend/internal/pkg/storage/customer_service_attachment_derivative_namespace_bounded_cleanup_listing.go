package storage

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Returns one bounded page. Callers remove those keys before requesting the
// next page; no whole-bucket scan or second object inventory is required.
type CustomerServiceAttachmentDerivativeNamespaceCleanupLister interface {
	ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(ctx context.Context, assetID uint) ([]string, error)
}

func customerServiceAttachmentDerivativeCleanupPrefix(assetID uint) (string, error) {
	if assetID == 0 {
		return "", fmt.Errorf("customer-service derivative cleanup requires an asset ID")
	}
	return fmt.Sprintf("%s/%d/", CustomerServiceConversationAttachmentDerivativesPrefix, assetID), nil
}

func (s *localStorage) ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(ctx context.Context, assetID uint) ([]string, error) {
	prefix, err := customerServiceAttachmentDerivativeCleanupPrefix(assetID)
	if err != nil {
		return nil, err
	}
	private := s.privateStorage()
	directory, err := private.localPath(strings.TrimSuffix(prefix, "/"))
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0)
	err = filepath.WalkDir(directory, func(candidate string, entry fs.DirEntry, walkErr error) error {
		if os.IsNotExist(walkErr) {
			return nil
		}
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("private derivative cleanup refuses symbolic links")
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(directory, candidate)
		if err != nil {
			return err
		}
		keys = append(keys, prefix+filepath.ToSlash(relative))
		if len(keys) == 1000 {
			return filepath.SkipAll
		}
		return nil
	})
	return keys, err
}

func (s *s3StorageImpl) ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(ctx context.Context, assetID uint) ([]string, error) {
	prefix, err := customerServiceAttachmentDerivativeCleanupPrefix(assetID)
	if err != nil {
		return nil, err
	}
	return s.ListObjects(ctx, prefix, 1000)
}

func (s *ossStorageImpl) ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(ctx context.Context, assetID uint) ([]string, error) {
	prefix, err := customerServiceAttachmentDerivativeCleanupPrefix(assetID)
	if err != nil {
		return nil, err
	}
	bucket := s.bucketForKey(prefix)
	if bucket == nil {
		return nil, fmt.Errorf("private storage bucket is not configured")
	}
	result, err := bucket.ListObjects(oss.Prefix(prefix), oss.MaxKeys(1000), oss.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(result.Objects))
	for _, object := range result.Objects {
		keys = append(keys, object.Key)
	}
	return keys, nil
}
