package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"tanzanite/internal/domain/media"
	settingdomain "tanzanite/internal/domain/setting"
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMediaEvidencePreservesOriginalAndDetectsStorageChanges(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &settingdomain.Setting{}))

	uploadRoot := t.TempDir()
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: uploadRoot,
		BaseURL:   "https://media.example.test",
	})
	require.NoError(t, err)

	settingService := NewSettingService(repository.NewSettingRepository(db), nil, 0)
	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "Northstar Wheels", Type: "string", Locale: "en", Group: "site"},
		{Key: "copyright_holder", Value: "Northstar Wheels LLC", Type: "string", Locale: "en", Group: "site"},
		{Key: "copyright_notice", Value: "Copyright 2026 Northstar Wheels LLC. All rights reserved.", Type: "string", Locale: "en", Group: "site"},
		{Key: "copyright_url", Value: "https://shop.northstar.example/policies/copyright", Type: "string", Locale: "en", Group: "site"},
	}))

	service := NewMediaService(repository.NewMediaRepository(db), storageService, settingService, "https://shop.northstar.example", 20<<30)
	original := []byte("original image bytes for copyright evidence")
	asset, err := service.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "wheelset.jpg", "image/jpeg", original),
		MediaType:  "image",
		UploaderID: 42,
	})
	require.NoError(t, err)

	expectedHash := sha256.Sum256(original)
	require.Equal(t, hex.EncodeToString(expectedHash[:]), asset.ContentSHA256)

	archiveBytes, archiveName, err := service.ExportCopyrightEvidence(context.Background(), asset.ID)
	require.NoError(t, err)
	require.Equal(t, "copyright-evidence-"+stringUint(asset.ID)+".zip", archiveName)

	archive, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	require.NoError(t, err)
	require.Len(t, archive.File, 4)

	originalEntry := zipEntry(t, archive, "original/wheelset.jpg")
	originalContents := zipEntryContents(t, originalEntry)
	require.Equal(t, original, originalContents)

	manifestEntry := zipEntry(t, archive, "manifest.json")
	var manifest CopyrightEvidenceManifest
	require.NoError(t, json.Unmarshal(zipEntryContents(t, manifestEntry), &manifest))
	require.Equal(t, "copyright-evidence/v1", manifest.Format)
	require.Equal(t, asset.ID, manifest.Asset.ID)
	require.Equal(t, asset.ContentSHA256, manifest.Verification.OriginalSHA256)
	require.True(t, manifest.Verification.IntegrityChecked)
	require.Equal(t, "Northstar Wheels", manifest.Copyright.SiteName)
	require.Equal(t, "shop.northstar.example", manifest.Copyright.SiteDomain)
	require.Equal(t, "Northstar Wheels LLC", manifest.Copyright.RightsHolder)
	require.Equal(t, "Copyright 2026 Northstar Wheels LLC. All rights reserved.", manifest.Copyright.CopyrightNotice)
	require.Equal(t, "https://shop.northstar.example/policies/copyright", manifest.Copyright.CopyrightPolicyURL)
	require.Equal(t, asset.ContentSHA256, manifest.Copyright.OriginalSHA256)
	require.Equal(t, asset.OriginalFilename, manifest.Copyright.OriginalFilename)
	require.False(t, manifest.Copyright.ServerReceivedAt.IsZero())

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "Changed Brand", Type: "string", Locale: "en", Group: "site"},
	}))

	archiveBytes, _, err = service.ExportCopyrightEvidence(context.Background(), asset.ID)
	require.NoError(t, err)
	archive, err = zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	require.NoError(t, err)
	manifestEntry = zipEntry(t, archive, "manifest.json")
	require.NoError(t, json.Unmarshal(zipEntryContents(t, manifestEntry), &manifest))
	require.Equal(t, "Northstar Wheels", manifest.Copyright.SiteName)

	require.NoError(t, os.WriteFile(
		filepath.Join(uploadRoot, filepath.FromSlash(asset.StorageKey)),
		[]byte("tampered bytes"),
		0o600,
	))
	_, _, err = service.ExportCopyrightEvidence(context.Background(), asset.ID)
	require.ErrorIs(t, err, ErrMediaEvidenceIntegrityMismatch)
}

func multipartFileHeader(t *testing.T, filename, mimeType string, contents []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write(contents)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, "/", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, request.ParseMultipartForm(int64(len(contents)+1024)))
	t.Cleanup(func() {
		if request.MultipartForm != nil {
			_ = request.MultipartForm.RemoveAll()
		}
	})

	file := request.MultipartForm.File["file"][0]
	file.Header.Set("Content-Type", mimeType)
	return file
}

func zipEntry(t *testing.T, archive *zip.Reader, name string) *zip.File {
	t.Helper()
	for _, entry := range archive.File {
		if entry.Name == name {
			return entry
		}
	}
	t.Fatalf("zip entry %q not found", name)
	return nil
}

func zipEntryContents(t *testing.T, entry *zip.File) []byte {
	t.Helper()
	reader, err := entry.Open()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = reader.Close()
	})

	var contents bytes.Buffer
	_, err = contents.ReadFrom(reader)
	require.NoError(t, err)
	return contents.Bytes()
}

func stringUint(value uint) string {
	if value == 0 {
		return "0"
	}

	var digits [20]byte
	index := len(digits)
	for value > 0 {
		index--
		digits[index] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[index:])
}
