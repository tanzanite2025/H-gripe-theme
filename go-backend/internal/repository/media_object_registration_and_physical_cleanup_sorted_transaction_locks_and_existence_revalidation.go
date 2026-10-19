package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"

	"gorm.io/gorm"
)

var (
	ErrMediaObjectRegistrationVerificationUnavailable = errors.New("media object registration provider verification is unavailable")
	ErrMediaObjectRegistrationDedicatedOwnerConflict  = errors.New("dedicated storage objects cannot acquire shared media registrations")
)

// A lock alone cannot prevent a waiting writer from registering already-deleted
// bytes. Every registration therefore requires the configured origin provider.
type registeredMediaObjectExistenceValidator interface {
	ValidateMediaObjectExistsBeforeRegisteringOwnership(context.Context, string) error
}

func mediaObjectOwnershipReadCommittedTransactionOptions(db *gorm.DB) []*sql.TxOptions {
	if db.Name() == "postgres" {
		return []*sql.TxOptions{{Isolation: sql.LevelReadCommitted}}
	}
	return nil
}

func (r *MediaRepository) registeredMediaOriginalAndDerivativeObjectIdentityKeys(asset *media.MediaAsset) []string {
	keys := r.registeredMediaObjectIdentityKeys(asset.StorageKey, asset.URL)
	for _, derivative := range asset.Derivatives {
		keys = append(keys, r.registeredMediaObjectIdentityKeys(derivative.StorageKey, derivative.URL)...)
	}
	return keys
}

func (r *MediaRepository) registeredMediaObjectIdentityKeys(storageKey, reference string) []string {
	var keys []string
	if key, valid := storage.NormalizeObjectKey(storageKey); valid {
		keys = append(keys, key)
	}
	for _, value := range []string{storageKey, reference} {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if key, valid := storage.NormalizeObjectKey(r.resolveRegisteredMediaObjectKeyForCleanupOwnership(value)); valid {
			keys = append(keys, key)
		}
	}
	return keys
}

// The caller already owns all affected asset rows in ascending ID order. No
// additional existing media asset row may be locked after these object locks.
// Namespace admission also covers private derivatives absent from SQL metadata.
func (r *MediaRepository) LockRegisteredMediaAssetObjectsBeforePhysicalCleanupWithinTransaction(asset *media.MediaAsset) error {
	keys := r.registeredMediaOriginalAndDerivativeObjectIdentityKeys(asset)
	var namespaces []string
	if asset.Visibility == "private" {
		for _, key := range r.registeredMediaObjectIdentityKeys(asset.StorageKey, asset.URL) {
			if storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(key) {
				namespaces = append(namespaces, fmt.Sprintf("%s/%d", storage.CustomerServiceConversationAttachmentDerivativesPrefix, asset.ID))
				break
			}
		}
	}
	return r.lockMediaObjectKeysAndPrivateDerivativeNamespacesWithinTransaction(keys, namespaces)
}

func (r *MediaRepository) lockAndValidateMediaObjectsBeforeRegisteringOwnershipWithinTransaction(keys []string) error {
	if err := r.lockMediaObjectKeysAndPrivateDerivativeNamespacesWithinTransaction(keys, nil); err != nil {
		return err
	}
	validator, configured := r.businessMediaReferenceResolver.(registeredMediaObjectExistenceValidator)
	if !configured {
		return ErrMediaObjectRegistrationVerificationUnavailable
	}
	unique := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		// Dedicated cleanup uses its own owner namespace protocol. A shared
		// registration cannot enter that lifecycle, even if bytes still exist.
		if storage.DedicatedPublicObjectNamespaceForStorageKey(key) != "" {
			return ErrMediaObjectRegistrationDedicatedOwnerConflict
		}
		unique[key] = struct{}{}
	}
	if len(unique) == 0 {
		return fmt.Errorf("registered media object identity is unavailable")
	}
	for key := range unique {
		if err := r.RejectUnpreparedBusinessDraftMediaObjectRegistration(key); err != nil {
			return err
		}
		// This must happen after waiting for the lock: cleanup may have removed
		// both the bytes and the last registration while this writer waited.
		if err := validator.ValidateMediaObjectExistsBeforeRegisteringOwnership(r.db.Statement.Context, key); err != nil {
			return err
		}
	}
	return nil
}

func (r *MediaRepository) lockAndValidateMediaDerivativeObjectsBeforePersistingMetadataWithinTransaction(derivatives []media.MediaAssetDerivative) error {
	var keys []string
	for _, derivative := range derivatives {
		keys = append(keys, r.registeredMediaObjectIdentityKeys(derivative.StorageKey, derivative.URL)...)
	}
	return r.lockAndValidateMediaObjectsBeforeRegisteringOwnershipWithinTransaction(keys)
}

func (r *MediaRepository) lockMediaObjectKeysAndPrivateDerivativeNamespacesWithinTransaction(keys, namespaces []string) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if r.db.Name() != "postgres" {
		return nil // SQLite tests serialize writes through their database transaction.
	}
	if _, transactionActive := r.db.Statement.ConnPool.(gorm.TxCommitter); !transactionActive {
		return fmt.Errorf("media object ownership locks require an active transaction")
	}
	var isolation string
	if err := r.db.Raw("SHOW transaction_isolation").Scan(&isolation).Error; err != nil {
		return err
	}
	if isolation != "read committed" {
		return fmt.Errorf("media object ownership requires read committed transaction isolation")
	}
	lockNames := make(map[string]struct{}, len(keys)+len(namespaces))
	for _, key := range keys {
		normalized, valid := storage.NormalizeObjectKey(key)
		if !valid {
			return fmt.Errorf("media object ownership lock identity is invalid")
		}
		lockNames["object/"+normalized] = struct{}{}
		if remainder, privateDerivative := strings.CutPrefix(normalized, storage.CustomerServiceConversationAttachmentDerivativesPrefix+"/"); privateDerivative {
			owner, _, _ := strings.Cut(remainder, "/")
			lockNames["namespace/"+storage.CustomerServiceConversationAttachmentDerivativesPrefix+"/"+owner] = struct{}{}
		}
	}
	for _, namespace := range namespaces {
		lockNames["namespace/"+namespace] = struct{}{}
	}
	// Sort actual advisory IDs, including the namespace IDs. Even a hash
	// collision cannot reverse the ordering of two overlapping lock plans.
	uniqueLockIDs := make(map[int64]struct{}, len(lockNames))
	for name := range lockNames {
		digest := sha256.Sum256([]byte("media-object-registration-and-physical-cleanup/v1/" + name))
		uniqueLockIDs[int64(binary.BigEndian.Uint64(digest[:8]))] = struct{}{} // #nosec G115 -- Preserve all digest bits as a stable signed PostgreSQL advisory lock identity.
	}
	lockIDs := make([]int64, 0, len(uniqueLockIDs))
	for id := range uniqueLockIDs {
		lockIDs = append(lockIDs, id)
	}
	sort.Slice(lockIDs, func(i, j int) bool { return lockIDs[i] < lockIDs[j] })
	for _, id := range lockIDs {
		if err := r.db.Exec("SELECT pg_advisory_xact_lock(?)", id).Error; err != nil {
			return err
		}
	}
	return nil
}
