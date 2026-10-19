package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"strings"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type orderEvidenceCompensationRegressionStorage struct {
	storage.StorageService
	uploadError             error
	uploadResultOverride    string
	lastUploadReference     string
	deleteError             error
	deleteCalls             int
	beforeDelete            func(context.Context)
	preallocatedKeys        []string
	beforePreallocatedWrite func(context.Context, string)
	afterPreallocatedWrite  func(context.Context, string)
	preallocatedResponse    func(string) string
	preallocatedReader      func(io.Reader) io.Reader
	skipPreallocatedWrite   bool
}

func (s *orderEvidenceCompensationRegressionStorage) UploadWithPrefixPrivate(ctx context.Context, file *multipart.FileHeader, prefix string) (string, error) {
	reference, err := s.StorageService.(storage.PrivateObjectUploader).UploadWithPrefixPrivate(ctx, file, prefix)
	if err != nil {
		return reference, err
	}
	s.lastUploadReference = reference
	if s.uploadResultOverride != "" {
		reference = s.uploadResultOverride
	}
	return reference, s.uploadError
}

func (s *orderEvidenceCompensationRegressionStorage) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func (s *orderEvidenceCompensationRegressionStorage) Delete(ctx context.Context, key string) error {
	s.deleteCalls++
	if s.beforeDelete != nil {
		s.beforeDelete(ctx)
	}
	if s.deleteError != nil {
		return s.deleteError
	}
	return s.StorageService.Delete(ctx, key)
}

type orderEvidenceCompensationRegressionEnvironment struct {
	db       *gorm.DB
	service  *OrderEvidenceAttachmentService
	provider *orderEvidenceCompensationRegressionStorage
	pkg      *orderevidence.OrderEvidencePackage
	item     orderevidence.OrderEvidenceItem
}

func configureOrderEvidenceCompensationRegressionEnvironment(t *testing.T, db *gorm.DB) orderEvidenceCompensationRegressionEnvironment {
	t.Helper()
	snapshot := servicePlanSnapshot(t, db, false)
	repo := repository.NewOrderEvidenceRepository(db)
	pkg, err := NewOrderEvidenceService().CreateInitialPackage(repository.TxRepositories{OrderEvidence: repo}, snapshot)
	require.NoError(t, err)
	items, err := repo.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(), BaseURL: "https://evidence.example.test", PrivateBaseURL: "https://private-evidence.example.test"})
	require.NoError(t, err)
	controlled := &orderEvidenceCompensationRegressionStorage{StorageService: provider}
	service := NewConfiguredOrderEvidenceAttachmentService(newOrderEvidenceAdminServiceForTest(db, repo).txManager, repo, controlled)
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	return orderEvidenceCompensationRegressionEnvironment{db: db, service: service, provider: controlled, pkg: pkg, item: findEvidenceItemByType(items, orderevidence.EvidenceItemTypeProductIdentity)}
}

func writeOrderEvidenceCompensationCandidateAndReferenceInput(t *testing.T, env orderEvidenceCompensationRegressionEnvironment) (string, OrderEvidenceAttachmentReferenceInput) {
	t.Helper()
	file := newEvidencePNGFileHeader(t, "proof.png")
	prefix := fmt.Sprintf("order-evidence/%d/%d/owned-upload-%s", env.item.OrderID, env.item.ID, uuid.NewString())
	reference, err := env.provider.StorageService.(storage.PrivateObjectUploader).UploadWithPrefixPrivate(t.Context(), file, prefix)
	require.NoError(t, err)
	key, err := env.provider.ObjectKey(reference)
	require.NoError(t, err)
	digest, size, err := hashMultipartFile(file)
	require.NoError(t, err)
	return prefix, OrderEvidenceAttachmentReferenceInput{EvidenceItemID: env.item.ID, StorageKey: key, OriginalFilename: file.Filename, MimeType: "image/png", SHA256: digest, SizeBytes: size, UploadedBy: 7}
}

func assertOrderEvidenceRegressionObjectExists(t *testing.T, env orderEvidenceCompensationRegressionEnvironment, key string, exists bool) {
	t.Helper()
	object, err := env.provider.Open(t.Context(), key)
	if exists {
		require.NoError(t, err)
		require.NoError(t, object.ReadCloser.Close())
	} else {
		require.Error(t, err)
	}
}

func TestLegacyOrderEvidenceCompensationPersistsPartialProviderFailureAndRetriesOriginalEvent(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		t.Run(fmt.Sprintf("canceled=%v", canceled), func(t *testing.T) {
			env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
			env.provider.uploadError = errors.New("provider returned a key with failure")
			env.provider.deleteError = errors.New("provider delete unavailable")
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if canceled {
				cancel()
			}
			_, err := env.service.uploadLegacyOrderEvidenceAttachmentWithReturnedIdentityForRegression(ctx, env.item.OrderID, env.item.ID, newEvidencePNGFileHeader(t, "proof.png"), 7)
			require.ErrorIs(t, err, env.provider.uploadError)
			var events []outbox.Event
			require.NoError(t, env.db.Find(&events).Error)
			require.Len(t, events, 1)
			key, keyErr := env.provider.ObjectKey(env.provider.lastUploadReference)
			require.NoError(t, keyErr)
			assertOrderEvidenceRegressionObjectExists(t, env, key, true)
			env.provider.deleteError = nil
			handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
			handler.ConfigureOrderEvidenceAttachmentService(env.service)
			require.NoError(t, handler.Handle(t.Context(), events[0]))
			require.NoError(t, handler.Handle(t.Context(), events[0]))
			assertOrderEvidenceRegressionObjectExists(t, env, key, false)
			var count int64
			require.NoError(t, env.db.Model(&outbox.Event{}).Count(&count).Error)
			require.EqualValues(t, 1, count)
		})
	}
}

func TestLegacyOrderEvidenceCompensationNeverDeletesWithoutDurableEventOrExactAttemptIdentity(t *testing.T) {
	for _, scenario := range []string{"outbox failure", "foreign provider", "other item", "legacy directory"} {
		t.Run(scenario, func(t *testing.T) {
			env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
			env.provider.uploadError = errors.New("partial upload")
			switch scenario {
			case "outbox failure":
				require.NoError(t, env.db.Migrator().DropTable(&outbox.Event{}))
			case "foreign provider":
				env.provider.uploadResultOverride = "https://untrusted.example/order-evidence/1/proof.png"
			case "other item":
				env.provider.uploadResultOverride = "order-evidence/1/999/proof.png"
			case "legacy directory":
				env.provider.uploadResultOverride = "order-evidence/1/1/old-proof.png"
			}
			_, err := env.service.uploadLegacyOrderEvidenceAttachmentWithReturnedIdentityForRegression(t.Context(), env.item.OrderID, env.item.ID, newEvidencePNGFileHeader(t, "proof.png"), 7)
			require.Error(t, err)
			require.Zero(t, env.provider.deleteCalls)
			key, err := env.provider.ObjectKey(env.provider.lastUploadReference)
			require.NoError(t, err)
			assertOrderEvidenceRegressionObjectExists(t, env, key, true)
		})
	}
}

func TestLegacyOrderEvidenceRegistrationRollbackUsesDurableCompensation(t *testing.T) {
	env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
	callback := "reject_evidence_registration_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "order_evidence_attachments" {
			tx.AddError(errors.New("registration rolled back"))
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
	_, err := env.service.uploadLegacyOrderEvidenceAttachmentWithReturnedIdentityForRegression(t.Context(), env.item.OrderID, env.item.ID, newEvidencePNGFileHeader(t, "proof.png"), 7)
	require.ErrorContains(t, err, "registration rolled back")
	var count int64
	require.NoError(t, env.db.Unscoped().Model(&orderevidence.OrderEvidenceAttachment{}).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, env.db.Model(&outbox.Event{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	key, err := env.provider.ObjectKey(env.provider.lastUploadReference)
	require.NoError(t, err)
	assertOrderEvidenceRegressionObjectExists(t, env, key, false)
}

func TestOrderEvidenceRegistrationVerifiesExactItemAndStoredBytes(t *testing.T) {
	for _, scenario := range []string{"valid", "other item", "other order", "missing bytes", "wrong hash", "wrong size", "wrong MIME", "unconfigured provider"} {
		t.Run(scenario, func(t *testing.T) {
			env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
			_, input := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
			switch scenario {
			case "other item":
				input.StorageKey = strings.Replace(input.StorageKey, fmt.Sprintf("/%d/owned-upload", env.item.ID), "/999/owned-upload", 1)
			case "other order":
				input.StorageKey = strings.Replace(input.StorageKey, "order-evidence/1/", "order-evidence/999/", 1)
			case "missing bytes":
				require.NoError(t, env.provider.StorageService.Delete(t.Context(), input.StorageKey))
			case "wrong hash":
				input.SHA256 = strings.Repeat("a", 64)
			case "wrong size":
				input.SizeBytes++
			case "wrong MIME":
				input.MimeType = "image/jpeg"
			case "unconfigured provider":
				env.service.storage = nil
			}
			attachment, err := env.service.RegisterReference(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, input)
			if scenario == "valid" {
				require.NoError(t, err)
				require.NotNil(t, attachment)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestOrderEvidenceDetachedHistoryRetainsBytesAndInvalidatesLastAttachmentReadiness(t *testing.T) {
	env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
	prefix, input := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
	attachment, err := env.service.RegisterReference(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, input)
	require.NoError(t, err)
	_, err = NewOrderEvidenceService().UpdateItem(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, OrderEvidenceItemUpdateInput{ItemID: env.item.ID, Status: orderevidence.EvidenceItemStatusComplete, DataJSON: []byte(`{"serial":"proof"}`), CapturedBy: 7})
	require.NoError(t, err)
	require.NoError(t, env.service.Delete(t.Context(), env.item.OrderID, env.item.ID, attachment.ID))
	_, _, err = env.service.Open(t.Context(), env.item.OrderID, env.item.ID, attachment.ID)
	require.ErrorIs(t, err, ErrOrderEvidenceAttachmentItemNotFound)
	var retained orderevidence.OrderEvidenceAttachment
	require.NoError(t, env.db.Unscoped().First(&retained, attachment.ID).Error)
	require.True(t, retained.DetachedAt.Valid)
	item, err := env.service.evidenceRepo.FindItemByID(env.item.ID)
	require.NoError(t, err)
	require.Equal(t, orderevidence.EvidenceItemStatusDraft, item.Status)
	err = env.service.cleanupFailedOrderEvidenceUploadWithImmutableHistoryRetention(t.Context(), prefix, []string{input.StorageKey})
	require.ErrorIs(t, err, ErrMediaAssetInUse)
	require.Zero(t, env.provider.deleteCalls)
	assertOrderEvidenceRegressionObjectExists(t, env, input.StorageKey, true)
}

func TestOrderEvidenceSnapshotExactReferenceRetentionScansAcrossPagesWithoutGrantingPublicAccess(t *testing.T) {
	for _, table := range []string{"order_evidence_snapshots", "order_evidence_submission_snapshots", "order_evidence_export_snapshots"} {
		t.Run(table, func(t *testing.T) {
			env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
			prefix, input := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
			require.NoError(t, env.db.Exec("DROP TABLE IF EXISTS "+table).Error)
			require.NoError(t, env.db.Exec("CREATE TABLE "+table+" (id INTEGER PRIMARY KEY, snapshot_data TEXT NOT NULL)").Error)
			for id := 1; id <= 257; id++ {
				require.NoError(t, env.db.Table(table).Create(map[string]any{"id": id, "snapshot_data": `{"key":"unrelated"}`}).Error)
			}
			value, err := json.Marshal(map[string]any{"nested": []any{map[string]string{"storage_key": input.StorageKey}}})
			require.NoError(t, err)
			require.NoError(t, env.db.Table(table).Where("id = ?", 257).Update("snapshot_data", string(value)).Error)
			err = env.service.cleanupFailedOrderEvidenceUploadWithImmutableHistoryRetention(t.Context(), prefix, []string{input.StorageKey})
			require.ErrorIs(t, err, ErrMediaAssetInUse)
			allowed, err := NewPublicUploadAccessService(nil, nil).CanServePublicUpload(t.Context(), input.StorageKey)
			require.NoError(t, err)
			require.False(t, allowed)
			require.NoError(t, env.db.Table(table).Where("id = ?", 257).Update("snapshot_data", `{"key":"https://untrusted.test/`+input.StorageKey+`"}`).Error)
			require.NoError(t, env.service.cleanupFailedOrderEvidenceUploadWithImmutableHistoryRetention(t.Context(), prefix, []string{input.StorageKey}))
			assertOrderEvidenceRegressionObjectExists(t, env, input.StorageKey, false)
		})
	}
}

func TestOrderEvidenceSharedOriginalAndDerivativeCleanupCannotEnterDedicatedNamespace(t *testing.T) {
	for _, role := range []string{"original", "derivative"} {
		t.Run(role, func(t *testing.T) {
			env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
			_, input := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
			asset := &media.MediaAsset{ID: 42, StorageKey: input.StorageKey, URL: env.provider.GetURL(input.StorageKey)}
			if role == "derivative" {
				asset.StorageKey = "media/original.png"
				asset.URL = env.provider.GetURL(asset.StorageKey)
				asset.Derivatives = []media.MediaAssetDerivative{{StorageKey: input.StorageKey, URL: env.provider.GetURL(input.StorageKey)}}
			}
			err := (&MediaService{storage: env.provider}).validateSharedMediaCleanupDoesNotOwnDedicatedPublicObjects(asset)
			require.ErrorIs(t, err, ErrMediaAssetInUse)
			require.NoError(t, env.db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}))
			sharedRepo := repository.NewMediaRepository(env.db)
			sharedRepo.ConfigurePublicBusinessMediaReferenceIdentityResolver(&MediaService{storage: env.provider})
			require.ErrorIs(t, sharedRepo.CreateAsset(asset), repository.ErrMediaObjectRegistrationDedicatedOwnerConflict)
			require.Equal(t, "order-evidence", storage.DedicatedPublicObjectNamespaceForStorageKey(input.StorageKey))
			require.Error(t, validateFailedSharedMediaOriginalCleanupOwnership(input.StorageKey))
			require.Error(t, validateOrderEvidenceFailedUploadAttemptNamespace(path.Dir(path.Dir(input.StorageKey))))
		})
	}
}

func TestOrderEvidenceRevisionCopiesOnlyActiveAttachmentsAndRetainsOriginalItemObjectOwnership(t *testing.T) {
	env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
	firstPrefix, first := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
	secondPrefix, second := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
	attachment, err := env.service.RegisterReference(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, first)
	require.NoError(t, err)
	_, err = env.service.RegisterReference(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, second)
	require.NoError(t, err)
	require.NoError(t, env.service.Delete(t.Context(), env.item.OrderID, env.item.ID, attachment.ID))
	require.NoError(t, env.db.Exec("UPDATE order_evidence_packages SET status='locked' WHERE id=?", env.pkg.ID).Error)
	revision, err := env.service.evidenceRepo.CreateRevision(env.pkg.ID, 7)
	require.NoError(t, err)
	loaded, err := env.service.evidenceRepo.FindPackageByID(revision.ID)
	require.NoError(t, err)
	var copied []orderevidence.OrderEvidenceAttachment
	for _, item := range loaded.Items {
		copied = append(copied, item.Attachments...)
	}
	require.Len(t, copied, 1)
	require.Equal(t, second.StorageKey, copied[0].StorageKey)
	require.NotEqual(t, env.item.ID, copied[0].EvidenceItemID)
	_, object, err := env.service.Open(t.Context(), env.item.OrderID, copied[0].EvidenceItemID, copied[0].ID)
	require.NoError(t, err)
	require.NoError(t, object.ReadCloser.Close())
	for _, candidate := range []struct{ prefix, key string }{{firstPrefix, first.StorageKey}, {secondPrefix, second.StorageKey}} {
		err = env.service.cleanupFailedOrderEvidenceUploadWithImmutableHistoryRetention(t.Context(), candidate.prefix, []string{candidate.key})
		require.ErrorIs(t, err, ErrMediaAssetInUse)
	}
	require.Zero(t, env.provider.deleteCalls)
}

func TestOrderEvidenceDetachmentAndReadinessRollbackTogether(t *testing.T) {
	env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
	_, input := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
	attachment, err := env.service.RegisterReference(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, input)
	require.NoError(t, err)
	_, err = NewOrderEvidenceService().UpdateItem(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, OrderEvidenceItemUpdateInput{ItemID: env.item.ID, Status: orderevidence.EvidenceItemStatusComplete, DataJSON: []byte(`{"serial":"proof"}`), CapturedBy: 7})
	require.NoError(t, err)
	callback := "reject_evidence_package_refresh_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Update().After("gorm:update").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "order_evidence_packages" {
			tx.AddError(errors.New("package refresh failed"))
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Update().Remove(callback) })
	require.ErrorContains(t, env.service.Delete(t.Context(), env.item.OrderID, env.item.ID, attachment.ID), "package refresh failed")
	_, err = env.service.evidenceRepo.FindAttachmentByIDForOrder(env.item.OrderID, env.item.ID, attachment.ID)
	require.NoError(t, err)
	item, err := env.service.evidenceRepo.FindItemByID(env.item.ID)
	require.NoError(t, err)
	require.Equal(t, orderevidence.EvidenceItemStatusComplete, item.Status)
	assertOrderEvidenceRegressionObjectExists(t, env, input.StorageKey, true)
}

func (s *orderEvidenceCompensationRegressionStorage) WritePrivateObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	s.preallocatedKeys = append(s.preallocatedKeys, key)
	if s.beforePreallocatedWrite != nil {
		s.beforePreallocatedWrite(ctx, key)
	}
	if s.preallocatedReader != nil {
		reader = s.preallocatedReader(reader)
	}
	var reference string
	var err error
	if !s.skipPreallocatedWrite {
		reference, err = s.StorageService.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(ctx, reader, key)
	}
	if err != nil {
		return reference, err
	}
	s.lastUploadReference = reference
	if s.afterPreallocatedWrite != nil {
		s.afterPreallocatedWrite(ctx, key)
	}
	if s.uploadResultOverride != "" {
		reference = s.uploadResultOverride
	}
	if s.preallocatedResponse != nil {
		reference = s.preallocatedResponse(reference)
	}
	return reference, s.uploadError
}
