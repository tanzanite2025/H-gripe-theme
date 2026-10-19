package service

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"testing"

	mediadomain "commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMediaUploadStoresSniffedMimeTypeInsteadOfClientHeader(t *testing.T) {
	service, storageService := newMediaAssetUploadTestService(t)

	payload := mediaUploadTestPNG(t)
	asset, err := service.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "wheel.jpg", "image/jpeg", payload),
		MediaType:  "image",
		UploaderID: 42,
	})

	require.NoError(t, err)
	require.Equal(t, "image/png", asset.MimeType)
	require.Equal(t, 1, storageService.uploadCalls)
}

func TestMediaUploadRejectsDisguisedMarkupBeforeStorageWrite(t *testing.T) {
	cases := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "svg",
			payload: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" onload="alert(1)"><script>alert(1)</script></svg>`),
		},
		{
			name:    "html",
			payload: []byte(`<!doctype html><html><body><script>alert(1)</script></body></html>`),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			service, storageService := newMediaAssetUploadTestService(t)

			asset, err := service.UploadAsset(context.Background(), MediaUploadInput{
				File:       multipartFileHeader(t, "payload.jpg", "image/jpeg", tt.payload),
				MediaType:  "image",
				UploaderID: 42,
			})

			require.Nil(t, asset)
			require.ErrorIs(t, err, ErrUnsupportedMediaType)
			require.Zero(t, storageService.uploadCalls)
		})
	}
}

func newMediaAssetUploadTestService(t *testing.T) (*MediaService, *recordingMediaUploadStorage) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(4)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(&mediadomain.MediaAsset{}, &outbox.Event{}))
	storageService := &recordingMediaUploadStorage{}
	service := NewMediaService(repository.NewMediaRepository(db), storageService, nil, "", 20<<30)
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	return service, storageService
}

func mediaUploadTestPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, color.RGBA{R: 10, G: uint8(80 + y), B: uint8(120 + x), A: 255})
		}
	}

	var buffer bytes.Buffer
	require.NoError(t, png.Encode(&buffer, img))
	return buffer.Bytes()
}

type recordingMediaUploadStorage struct {
	uploadCalls                  int
	objects                      map[string][]byte
	permanentWriteResponseError  error
	permanentWriteReturnWrongKey bool
	onPermanentWrite             func(string)
}

func (s *recordingMediaUploadStorage) WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if !storage.IsPreallocatedPermanentMediaOriginalObjectKey(key) {
		return "", errors.New("invalid permanent media original key")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	payload, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	s.uploadCalls++
	if s.objects == nil {
		s.objects = make(map[string][]byte)
	}
	s.objects[key] = payload
	if s.onPermanentWrite != nil {
		s.onPermanentWrite(key)
	}
	if s.permanentWriteResponseError != nil {
		return "", s.permanentWriteResponseError
	}
	if s.permanentWriteReturnWrongKey {
		return s.GetURL("foreign-library-key.png"), nil
	}
	return s.GetURL(key), nil
}

func (s *recordingMediaUploadStorage) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
	return s.UploadWithPrefix(ctx, file, "")
}

func (s *recordingMediaUploadStorage) UploadWithPrefix(ctx context.Context, file *multipart.FileHeader, prefix string) (string, error) {
	source, err := file.Open()
	if err != nil {
		return "", err
	}
	defer source.Close()
	return s.UploadFromReaderWithPrefix(ctx, source, file.Filename, prefix)
}

func (s *recordingMediaUploadStorage) UploadFromReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	return s.UploadFromReaderWithPrefix(ctx, reader, filename, "")
}

func (s *recordingMediaUploadStorage) UploadFromReaderWithPrefix(_ context.Context, reader io.Reader, filename string, prefix string) (string, error) {
	payload, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	s.uploadCalls++
	if s.objects == nil {
		s.objects = make(map[string][]byte)
	}
	key := path.Join(prefix, path.Base(filename))
	s.objects[key] = payload
	return s.GetURL(key), nil
}

func (s *recordingMediaUploadStorage) Delete(_ context.Context, reference string) error {
	key, _ := s.ObjectKey(reference)
	delete(s.objects, key)
	return nil
}

func (s *recordingMediaUploadStorage) Open(_ context.Context, key string) (*storage.StoredObject, error) {
	payload, exists := s.objects[key]
	if !exists {
		return nil, os.ErrNotExist
	}
	return &storage.StoredObject{ReadCloser: io.NopCloser(bytes.NewReader(payload)), Size: int64(len(payload))}, nil
}

func (s *recordingMediaUploadStorage) GetURL(filename string) string {
	return "https://media.example.test/uploads/" + strings.TrimLeft(path.Clean(filename), "/")
}

func (s *recordingMediaUploadStorage) ObjectKey(reference string) (string, error) {
	const marker = "/uploads/"
	if index := strings.Index(reference, marker); index >= 0 {
		return strings.TrimLeft(reference[index+len(marker):], "/"), nil
	}
	return strings.TrimLeft(path.Clean(reference), "/"), nil
}

func (s *recordingMediaUploadStorage) CopyObject(context.Context, string, string) error {
	return nil
}
