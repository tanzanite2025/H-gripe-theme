package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// These transient identities are for accounting only. They grant no read URL,
// business ownership or deletion authority, and must not be exposed by the API.
type PhysicalCapacityLocation struct {
	ID        string
	localRoot string
	bucket    string
}

type PhysicalCapacityObjectAddress struct {
	LocationID string
	ObjectID   string
}

type PhysicalCapacityObjectMetadata struct {
	PhysicalCapacityObjectAddress
	Size int64
}

// Implementations enumerate current object metadata without opening bodies,
// creating directories, listing versions/multipart uploads or changing objects.
type ReadOnlyPhysicalCapacityMetadataProvider interface {
	PhysicalCapacityLocations() ([]PhysicalCapacityLocation, error)
	PhysicalCapacityObjectAddress(key string) (PhysicalCapacityObjectAddress, error)
	WalkPhysicalCapacityObjectMetadata(context.Context, PhysicalCapacityLocation, func(PhysicalCapacityObjectMetadata) error) error
}

func physicalCapacityIdentity(parts ...string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(parts, "\x00"))))
}

func canonicalPhysicalCapacityLocalRootWithoutCreatingDirectories(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("capacity storage root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	// Resolve existing ancestors too: a not-yet-created private directory can
	// reside below an alias, and an audit must never create that directory.
	resolved, err := filepath.EvalSymlinks(abs)
	if os.IsNotExist(err) {
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", err
		}
		resolved, err = canonicalPhysicalCapacityLocalRootWithoutCreatingDirectories(parent)
		resolved = filepath.Join(resolved, filepath.Base(abs))
	}
	if err != nil {
		return "", err
	}
	return filepath.Clean(resolved), nil
}

func localPhysicalCapacityPathIdentity(value string) string {
	if runtime.GOOS == "windows" {
		value = strings.ToLower(value)
	}
	return physicalCapacityIdentity("local", filepath.Clean(value))
}

func (s *localStorage) PhysicalCapacityLocations() ([]PhysicalCapacityLocation, error) {
	if s == nil || s.config == nil {
		return nil, fmt.Errorf("capacity storage is unavailable")
	}
	roots := []string{s.config.LocalPath, s.privateStorage().config.LocalPath}
	locations := make([]PhysicalCapacityLocation, 0, 2)
	for _, root := range roots {
		canonical, err := canonicalPhysicalCapacityLocalRootWithoutCreatingDirectories(root)
		if err != nil {
			return nil, err
		}
		locations = append(locations, PhysicalCapacityLocation{ID: localPhysicalCapacityPathIdentity(canonical), localRoot: canonical})
	}
	return locations, nil
}

func (s *localStorage) PhysicalCapacityObjectAddress(key string) (PhysicalCapacityObjectAddress, error) {
	canonical, ok := NormalizeObjectKey(key)
	if !ok || canonical != key {
		return PhysicalCapacityObjectAddress{}, fmt.Errorf("capacity requires an exact object key")
	}
	locations, err := s.PhysicalCapacityLocations()
	if err != nil {
		return PhysicalCapacityObjectAddress{}, err
	}
	location := locations[0]
	if IsPrivateObjectKey(key) {
		location = locations[1]
	}
	target, err := (&localStorage{config: &Config{LocalPath: location.localRoot}}).localPath(key)
	if err != nil {
		return PhysicalCapacityObjectAddress{}, err
	}
	return PhysicalCapacityObjectAddress{LocationID: location.ID, ObjectID: localPhysicalCapacityPathIdentity(target)}, nil
}

func (s *localStorage) WalkPhysicalCapacityObjectMetadata(ctx context.Context, location PhysicalCapacityLocation, visit func(PhysicalCapacityObjectMetadata) error) error {
	return filepath.WalkDir(location.localRoot, func(candidate string, entry fs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if os.IsNotExist(walkErr) && candidate == location.localRoot {
			return nil
		}
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("capacity scan refuses symbolic links inside storage")
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("capacity scan encountered a non-regular file")
		}
		return visit(PhysicalCapacityObjectMetadata{PhysicalCapacityObjectAddress: PhysicalCapacityObjectAddress{LocationID: location.ID, ObjectID: localPhysicalCapacityPathIdentity(candidate)}, Size: info.Size()})
	})
}

func configuredRemotePhysicalCapacityLocations(config *Config, kind StorageType) ([]PhysicalCapacityLocation, error) {
	if config == nil {
		return nil, fmt.Errorf("capacity storage is unavailable")
	}
	endpoint := strings.TrimRight(strings.TrimSpace(config.Endpoint), "/")
	if endpoint == "" {
		if kind == StorageTypeS3 {
			endpoint = "aws-s3"
		} else {
			endpoint = "https://oss-" + config.Region + ".aliyuncs.com"
		}
	} else {
		if !strings.Contains(endpoint, "://") {
			endpoint = "https://" + endpoint
		}
		parsed, err := url.Parse(endpoint)
		if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("capacity endpoint identity is invalid")
		}
		parsed.Scheme = strings.ToLower(parsed.Scheme)
		parsed.Host = strings.ToLower(parsed.Host)
		endpoint = strings.TrimRight(parsed.String(), "/")
	}
	if strings.TrimSpace(config.Bucket) == "" {
		return nil, fmt.Errorf("capacity bucket is required")
	}
	locations := make([]PhysicalCapacityLocation, 0, 2)
	for _, bucket := range []string{config.Bucket, config.PrivateBucket} {
		bucket = strings.TrimSpace(bucket)
		if bucket == "" {
			continue
		}
		locations = append(locations, PhysicalCapacityLocation{ID: physicalCapacityIdentity(string(kind), endpoint, bucket), bucket: bucket})
	}
	return locations, nil
}

func remotePhysicalCapacityObjectAddress(config *Config, kind StorageType, key string) (PhysicalCapacityObjectAddress, error) {
	canonical, ok := NormalizeObjectKey(key)
	if !ok || canonical != key {
		return PhysicalCapacityObjectAddress{}, fmt.Errorf("capacity requires an exact object key")
	}
	locations, err := configuredRemotePhysicalCapacityLocations(config, kind)
	if err != nil {
		return PhysicalCapacityObjectAddress{}, err
	}
	location := locations[0]
	if IsPrivateObjectKey(key) {
		if len(locations) < 2 || locations[0].ID == locations[1].ID {
			return PhysicalCapacityObjectAddress{}, fmt.Errorf("private capacity location is unavailable")
		}
		location = locations[1]
	}
	return PhysicalCapacityObjectAddress{LocationID: location.ID, ObjectID: physicalCapacityIdentity(location.ID, key)}, nil
}

func (s *s3StorageImpl) PhysicalCapacityLocations() ([]PhysicalCapacityLocation, error) {
	return configuredRemotePhysicalCapacityLocations(s.config, StorageTypeS3)
}
func (s *ossStorageImpl) PhysicalCapacityLocations() ([]PhysicalCapacityLocation, error) {
	return configuredRemotePhysicalCapacityLocations(s.config, StorageTypeOSS)
}
func (s *s3StorageImpl) PhysicalCapacityObjectAddress(key string) (PhysicalCapacityObjectAddress, error) {
	return remotePhysicalCapacityObjectAddress(s.config, StorageTypeS3, key)
}
func (s *ossStorageImpl) PhysicalCapacityObjectAddress(key string) (PhysicalCapacityObjectAddress, error) {
	return remotePhysicalCapacityObjectAddress(s.config, StorageTypeOSS, key)
}

func (s *s3StorageImpl) WalkPhysicalCapacityObjectMetadata(ctx context.Context, location PhysicalCapacityLocation, visit func(PhysicalCapacityObjectMetadata) error) error {
	var continuation *string
	seenCursors := make(map[string]struct{})
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		page, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(location.bucket), MaxKeys: aws.Int32(1000), ContinuationToken: continuation})
		if err != nil {
			return err
		}
		for _, object := range page.Contents {
			if object.Key == nil || object.Size == nil {
				return fmt.Errorf("capacity metadata is incomplete")
			}
			if err := visit(PhysicalCapacityObjectMetadata{PhysicalCapacityObjectAddress: PhysicalCapacityObjectAddress{LocationID: location.ID, ObjectID: physicalCapacityIdentity(location.ID, *object.Key)}, Size: *object.Size}); err != nil {
				return err
			}
		}
		if !aws.ToBool(page.IsTruncated) {
			return nil
		}
		next := aws.ToString(page.NextContinuationToken)
		if _, duplicate := seenCursors[next]; next == "" || duplicate {
			return fmt.Errorf("capacity pagination did not advance")
		}
		seenCursors[next] = struct{}{}
		continuation = aws.String(next)
	}
}

func (s *ossStorageImpl) WalkPhysicalCapacityObjectMetadata(ctx context.Context, location PhysicalCapacityLocation, visit func(PhysicalCapacityObjectMetadata) error) error {
	bucket, err := s.client.Bucket(location.bucket)
	if err != nil {
		return err
	}
	marker := ""
	seenCursors := make(map[string]struct{})
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		page, err := bucket.ListObjects(oss.MaxKeys(1000), oss.Marker(marker), oss.WithContext(ctx))
		if err != nil {
			return err
		}
		for _, object := range page.Objects {
			if err := visit(PhysicalCapacityObjectMetadata{PhysicalCapacityObjectAddress: PhysicalCapacityObjectAddress{LocationID: location.ID, ObjectID: physicalCapacityIdentity(location.ID, object.Key)}, Size: object.Size}); err != nil {
				return err
			}
		}
		if !page.IsTruncated {
			return nil
		}
		if _, duplicate := seenCursors[page.NextMarker]; page.NextMarker == "" || duplicate {
			return fmt.Errorf("capacity pagination did not advance")
		}
		seenCursors[page.NextMarker] = struct{}{}
		marker = page.NextMarker
	}
}
