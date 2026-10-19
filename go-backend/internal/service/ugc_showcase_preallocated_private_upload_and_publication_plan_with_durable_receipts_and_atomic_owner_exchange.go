package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"reflect"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func newUGCShowcasePrecommittedObjectWriteCleanupReceipt(keys []string, expectedBytes ...map[string]int64) (*outbox.Event, error) {
	id := uuid.NewString()
	event, err := newObjectStorageCleanupEvent(objectCleanupResourceFailedUGCShowcaseUpload, id, outbox.AggregateTypeUGCShowcase, id, keys, expectedBytes...)
	if err != nil {
		return nil, err
	}
	event.EventKey = repository.UGCShowcasePrecommittedObjectWriteReceiptEventKeyPrefix + id
	return event, nil
}

func (s *UGCShowcaseService) uploadPendingShowcaseImagesWithPrecommittedKeysAndAtomicSubmission(ctx context.Context, userID, orderID uint, files []*multipart.FileHeader, params map[string]string) (*ugcshowcase.UGCShowcase, error) {
	if err := s.repo.RequireIndependentUGCShowcaseProviderOperationTransaction(); err != nil {
		return nil, err
	}
	writer, available := s.storage.(storage.PreallocatedPrivateObjectWriter)
	if !available {
		return nil, ErrShowcaseStorageUnavailable
	}
	keys := make([]string, len(files))
	digests := make([]string, len(files))
	for index, file := range files {
		if file == nil {
			return nil, ErrShowcaseImagesInvalid
		}
		key, err := storage.AllocatePrivateObjectKeyForPendingUpload(file.Filename, showcasePendingUploadPrefix)
		if err != nil {
			return nil, err
		}
		keys[index] = key
		digest, err := contentSHA256FromMultipartFile(file)
		if err != nil {
			return nil, err
		}
		digests[index] = digest
	}
	expectedBytes := make(map[string]int64, len(keys))
	for index, key := range keys {
		expectedBytes[key] = files[index].Size
	}
	receipt, err := newUGCShowcasePrecommittedObjectWriteCleanupReceipt(keys, expectedBytes)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(keys)
	if err != nil {
		return nil, err
	}
	item := &ugcshowcase.UGCShowcase{UserID: userID, OrderID: showcaseUploadOrderID(orderID), Kind: ugcshowcase.KindUser, Region: params["region"], Location: params["location"], Nickname: params["nickname"], Notes: params["notes"], Images: encoded, Status: ugcshowcase.StatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err = s.outbox.RunUGCShowcaseObjectWritesWithPrecommittedCleanupReceiptAndAtomicOwnerExchange(ctx, receipt, func(tx *gorm.DB) error {
		lockedService := *s
		lockedService.repo = s.repo.WithTx(tx)
		// Keep the existing per-user admission lock and count inside the owner
		// transaction. Limit failure never promotes partial uploads to owners.
		if s.pendingLimit > 0 {
			if err := lockedService.repo.LockUserForSubmissionLimit(userID); err != nil {
				if repository.IsRecordNotFound(err) {
					return ErrShowcaseUploadOrderNotEligible
				}
				return err
			}
			count, err := lockedService.repo.CountByUserAndStatus(userID, ugcshowcase.StatusPending)
			if err != nil {
				return err
			}
			if count >= int64(s.pendingLimit) {
				return ErrShowcaseUploadPendingLimitExceeded
			}
		}
		for index, file := range files {
			reader, err := file.Open()
			if err != nil {
				return err
			}
			reference, writeErr := writer.WritePrivateObjectAtPreallocatedKey(tx.Statement.Context, reader, keys[index])
			closeErr := reader.Close()
			if err := errors.Join(writeErr, closeErr); err != nil {
				return err
			}
			returnedKey, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, showcasePendingUploadPrefix, reference)
			if err != nil || returnedKey != keys[index] {
				return ErrShowcaseImagesInvalid
			}
			if err := s.verifyUGCShowcaseObjectSizeAndContentDigest(tx.Statement.Context, keys[index], file.Size, digests[index]); err != nil {
				return err
			}
		}
		return lockedService.repo.Create(item)
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

type ugcShowcasePreallocatedPublicationCopy struct{ sourceKey, destinationKey string }
type ugcShowcasePreallocatedPublicationPlan struct {
	references        []string
	publishedKeys     []string
	copies            []ugcShowcasePreallocatedPublicationCopy
	pendingReferences []string
}

func (s *UGCShowcaseService) prepareUGCShowcasePublicationDestinationPlanWithoutObjectIO(item *ugcshowcase.UGCShowcase) (ugcShowcasePreallocatedPublicationPlan, error) {
	var plan ugcShowcasePreallocatedPublicationPlan
	if item.Status != ugcshowcase.StatusPending {
		return plan, ErrShowcaseInvalidTransition
	}
	references, err := decodeShowcaseImageURLs(item.Images)
	if err != nil {
		return plan, err
	}
	plan.references = references
	plan.publishedKeys = make([]string, 0, len(references))
	destinationBySource := make(map[string]string)
	for _, reference := range references {
		key, err := s.resolveUGCShowcaseStoredImageObjectKey(reference)
		if err != nil || key == showcasePendingUploadPrefix || key == showcaseApprovedUploadPrefix {
			return plan, ErrShowcaseImagesInvalid
		}
		if showcaseStorageKeyIsApproved(key) {
			plan.publishedKeys = append(plan.publishedKeys, key)
			continue
		}
		if storage.IsPrivateObjectKey(key) && !showcaseStorageKeyIsPending(key) {
			return plan, ErrShowcaseImagesInvalid
		}
		if namespace := storage.DedicatedPublicObjectNamespaceForStorageKey(key); namespace != "" && namespace != showcasePendingUploadPrefix {
			return plan, ErrShowcaseImagesInvalid
		}
		destination, exists := destinationBySource[key]
		if !exists {
			// A fresh directory per distinct source prevents a legacy public key
			// and a pending key with the same suffix from sharing a destination.
			destination, err = approvedShowcaseStorageKeyForSource(key, uuid.NewString())
			if err != nil {
				return plan, err
			}
			plan.copies = append(plan.copies, ugcShowcasePreallocatedPublicationCopy{key, destination})
			destinationBySource[key] = destination
		}
		plan.publishedKeys = append(plan.publishedKeys, destination)
		if showcaseStorageKeyIsPending(key) {
			plan.pendingReferences = append(plan.pendingReferences, reference)
		}
	}
	return plan, nil
}

func (s *UGCShowcaseService) approveUGCShowcaseWithPrecommittedCopyDestinationsAndAtomicPublication(ctx context.Context, id uint) error {
	if err := s.repo.RequireIndependentUGCShowcaseProviderOperationTransaction(); err != nil {
		return err
	}
	if err := s.outbox.RequireIndependentUGCShowcaseObjectWriteReceiptTransaction(); err != nil {
		return err
	}
	// This read creates no files and holds no lifecycle lock. The real owner is
	// locked and its references compared again after receipt commitment.
	snapshot, err := s.repo.WithContext(ctx).GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrShowcaseNotFound
	}
	if err != nil {
		return err
	}
	plan, err := s.prepareUGCShowcasePublicationDestinationPlanWithoutObjectIO(snapshot)
	if err != nil {
		return err
	}
	operation := func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		current, err := repo.FindUGCShowcaseByIDForUpdateWithinLifecycleTransaction(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShowcaseNotFound
		}
		if err != nil {
			return err
		}
		if current.Status != ugcshowcase.StatusPending {
			return ErrShowcaseInvalidTransition
		}
		references, err := decodeShowcaseImageURLs(current.Images)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(references, plan.references) {
			return ErrShowcaseInvalidTransition
		}
		for _, copyPlan := range plan.copies {
			size, digest, err := s.readUGCShowcaseObjectSizeAndContentDigest(tx.Statement.Context, copyPlan.sourceKey)
			if err != nil {
				return err
			}
			if err := s.storage.CopyObject(tx.Statement.Context, copyPlan.sourceKey, copyPlan.destinationKey); err != nil {
				return err
			}
			if err := s.verifyUGCShowcaseObjectSizeAndContentDigest(tx.Statement.Context, copyPlan.destinationKey, size, digest); err != nil {
				return err
			}
		}
		encoded, err := json.Marshal(plan.publishedKeys)
		if err != nil {
			return err
		}
		updated, err := repo.UpdatePendingImagesAndStatus(id, encoded, ugcshowcase.StatusApproved, "")
		if err != nil {
			return err
		}
		if !updated {
			return ErrShowcaseInvalidTransition
		}
		return s.enqueueShowcaseCleanup(tx, id, s.showcasePendingCleanupKeys(plan.pendingReferences))
	}
	if len(plan.copies) == 0 {
		return s.repo.WithLockedUGCShowcaseLifecycleUntilProviderOperationReturns(ctx, func(_ *repository.UGCShowcaseRepository, tx *gorm.DB) error { return operation(tx) })
	}
	keys := make([]string, 0, len(plan.copies))
	for _, copyPlan := range plan.copies {
		keys = append(keys, copyPlan.destinationKey)
	}
	receipt, err := newUGCShowcasePrecommittedObjectWriteCleanupReceipt(keys)
	if err != nil {
		return err
	}
	return s.outbox.RunUGCShowcaseObjectWritesWithPrecommittedCleanupReceiptAndAtomicOwnerExchange(ctx, receipt, operation)
}

func (s *UGCShowcaseService) readUGCShowcaseObjectSizeAndContentDigest(ctx context.Context, key string) (int64, string, error) {
	opener, available := s.storage.(storage.ObjectOpener)
	if !available {
		return 0, "", ErrShowcaseStorageUnavailable
	}
	object, err := opener.Open(ctx, key)
	if err != nil {
		return 0, "", err
	}
	if object == nil || object.ReadCloser == nil {
		return 0, "", ErrShowcaseImagesInvalid
	}
	digest := sha256.New()
	size, readErr := io.Copy(digest, object.ReadCloser)
	closeErr := object.ReadCloser.Close()
	if err := errors.Join(readErr, closeErr, ctx.Err()); err != nil {
		return 0, "", err
	}
	return size, fmt.Sprintf("%x", digest.Sum(nil)), nil
}

func (s *UGCShowcaseService) verifyUGCShowcaseObjectSizeAndContentDigest(ctx context.Context, key string, expectedSize int64, expectedDigest string) error {
	size, digest, err := s.readUGCShowcaseObjectSizeAndContentDigest(ctx, key)
	if err != nil {
		return err
	}
	if size != expectedSize || digest != expectedDigest {
		return fmt.Errorf("%w: stored UGC image integrity mismatch", ErrShowcaseImagesInvalid)
	}
	return nil
}
