package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderEvidenceAttachmentServiceUploadsAndScopesObject(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, false)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	pkg, err := NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)
	items, err := evidenceRepo.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	identity := findEvidenceItemByType(items, orderevidence.EvidenceItemTypeProductIdentity)
	require.NotZero(t, identity.ID)

	txManager := newOrderEvidenceAdminServiceForTest(db, evidenceRepo).txManager
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: t.TempDir(),
		BaseURL:   "http://evidence.test",
	})
	require.NoError(t, err)
	attachmentService := NewConfiguredOrderEvidenceAttachmentService(
		txManager,
		evidenceRepo,
		storageService,
	)

	fileHeader := newEvidencePNGFileHeader(t, "identity.png")
	attachment, err := attachmentService.Upload(
		context.Background(),
		snapshot.OrderID,
		identity.ID,
		fileHeader,
		7,
	)
	require.NoError(t, err)
	require.NotNil(t, attachment)
	assert.True(t, strings.HasPrefix(attachment.StorageKey, "order-evidence/"))
	assert.Contains(t, attachment.StorageKey, "/"+stringID(identity.ID)+"/")
	assert.Equal(t, int64(fileHeader.Size), attachment.SizeBytes)

	raw := mustReadEvidenceFileHeader(t, fileHeader)
	hash := sha256.Sum256(raw)
	assert.Equal(t, hex.EncodeToString(hash[:]), attachment.SHA256)

	openedAttachment, object, err := attachmentService.Open(
		context.Background(),
		snapshot.OrderID,
		identity.ID,
		attachment.ID,
	)
	require.NoError(t, err)
	require.NotNil(t, openedAttachment)
	require.NotNil(t, object)
	defer func() { _ = object.ReadCloser.Close() }()
	assert.Equal(t, attachment.ID, openedAttachment.ID)

	var opened bytes.Buffer
	_, err = opened.ReadFrom(object.ReadCloser)
	require.NoError(t, err)
	assert.Equal(t, raw, opened.Bytes())

	_, _, err = attachmentService.Open(
		context.Background(),
		snapshot.OrderID+1,
		identity.ID,
		attachment.ID,
	)
	require.ErrorIs(t, err, ErrOrderEvidenceAttachmentItemNotFound)
}

func TestOrderEvidenceAttachmentServiceAcceptsCarrierPODPDF(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, false)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	pkg, err := NewOrderEvidenceService().CreateInitialPackage(repository.TxRepositories{OrderEvidence: evidenceRepo}, snapshot)
	require.NoError(t, err)
	items, err := evidenceRepo.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	pod := findEvidenceItemByType(items, orderevidence.EvidenceItemTypeSignedPOD)
	require.NotZero(t, pod.ID)
	txManager := newOrderEvidenceAdminServiceForTest(db, evidenceRepo).txManager
	storageService, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "http://evidence.test"})
	require.NoError(t, err)
	attachmentService := NewConfiguredOrderEvidenceAttachmentService(txManager, evidenceRepo, storageService)

	attachment, err := attachmentService.Upload(context.Background(), snapshot.OrderID, pod.ID, newEvidencePDFFileHeader(t, "carrier-pod.pdf"), 7)
	require.NoError(t, err)
	require.Equal(t, "application/pdf", attachment.MimeType)
}

func TestOrderEvidenceAttachmentServiceDeletesReferenceRetainsObjectAndRejectsLockedPackage(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, false)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	pkg, err := NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)
	items, err := evidenceRepo.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	identity := findEvidenceItemByType(items, orderevidence.EvidenceItemTypeProductIdentity)
	require.NotZero(t, identity.ID)

	txManager := newOrderEvidenceAdminServiceForTest(db, evidenceRepo).txManager
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: t.TempDir(),
		BaseURL:   "http://evidence.test",
	})
	require.NoError(t, err)
	attachmentService := NewConfiguredOrderEvidenceAttachmentService(
		txManager,
		evidenceRepo,
		storageService,
	)

	fileHeader := newEvidencePNGFileHeader(t, "wrong-identity.png")
	raw := mustReadEvidenceFileHeader(t, fileHeader)
	attachment, err := attachmentService.Upload(
		context.Background(),
		snapshot.OrderID,
		identity.ID,
		fileHeader,
		7,
	)
	require.NoError(t, err)

	require.NoError(t, attachmentService.Delete(
		context.Background(),
		snapshot.OrderID,
		identity.ID,
		attachment.ID,
	))
	_, err = evidenceRepo.FindAttachmentByIDForOrder(snapshot.OrderID, identity.ID, attachment.ID)
	require.True(t, repository.IsRecordNotFound(err))

	opener, ok := storageService.(storage.ObjectOpener)
	require.True(t, ok)
	object, err := opener.Open(context.Background(), attachment.StorageKey)
	require.NoError(t, err)
	var retained bytes.Buffer
	_, err = retained.ReadFrom(object.ReadCloser)
	require.NoError(t, err)
	require.NoError(t, object.ReadCloser.Close())
	assert.Equal(t, raw, retained.Bytes())

	replacement, err := attachmentService.Upload(
		context.Background(),
		snapshot.OrderID,
		identity.ID,
		newEvidencePNGFileHeader(t, "replacement.png"),
		7,
	)
	require.NoError(t, err)

	require.NoError(t, db.Exec(
		"UPDATE order_evidence_packages SET status = ?, locked_at = ? WHERE id = ?",
		orderevidence.PackageStatusLocked,
		time.Now().UTC(),
		pkg.ID,
	).Error)
	err = attachmentService.Delete(
		context.Background(),
		snapshot.OrderID,
		identity.ID,
		replacement.ID,
	)
	require.ErrorIs(t, err, orderevidence.ErrOrderEvidencePackageLocked)
}

func newEvidencePNGFileHeader(t *testing.T, filename string) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	require.NoError(t, png.Encode(part, image.NewRGBA(image.Rect(0, 0, 4, 4))))
	require.NoError(t, writer.Close())

	request := httptest.NewRequest("POST", "/", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, request.ParseMultipartForm(int64(body.Len()+1024)))
	files := request.MultipartForm.File["file"]
	require.Len(t, files, 1)
	return files[0]
}

func newEvidencePDFFileHeader(t *testing.T, filename string) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write([]byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\n%%EOF\n"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest("POST", "/", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, request.ParseMultipartForm(int64(body.Len()+1024)))
	files := request.MultipartForm.File["file"]
	require.Len(t, files, 1)
	return files[0]
}

func mustReadEvidenceFileHeader(t *testing.T, file *multipart.FileHeader) []byte {
	t.Helper()
	src, err := file.Open()
	require.NoError(t, err)
	defer func() { _ = src.Close() }()

	var raw bytes.Buffer
	_, err = raw.ReadFrom(src)
	require.NoError(t, err)
	return raw.Bytes()
}

func stringID(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}
