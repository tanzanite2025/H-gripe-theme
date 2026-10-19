package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var ErrDedicatedSiteIdentityTransactionalObjectValidationUnavailable = errors.New("dedicated site identity transactional object validation is unavailable")

// Only the transaction lifetime ignores cancellation. SQL, lock waits and the
// synchronous provider operation retain the caller's context and deadline.
func runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(ctx context.Context, db *gorm.DB, namespace string, operation func(*gorm.DB) error) error {
	if db == nil {
		return gorm.ErrInvalidDB
	}
	if _, nested := db.Statement.ConnPool.(gorm.TxCommitter); nested {
		return fmt.Errorf("dedicated physical cleanup requires its own transaction lifetime")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return db.WithContext(context.WithoutCancel(ctx)).Transaction(func(tx *gorm.DB) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		requestTx := tx.WithContext(ctx)
		if err := acquireDedicatedPublicObjectNamespaceLockWithinTransaction(requestTx, namespace); err != nil {
			return err
		}
		return errors.Join(operation(requestTx), ctx.Err())
	}, mediaObjectOwnershipReadCommittedTransactionOptions(db)...)
}

const (
	dedicatedSiteLogoObjectNamespace              = "site-logo"
	dedicatedSiteFaviconObjectNamespace           = "site-favicon"
	dedicatedHomeVisualObjectNamespace            = "visual-showcase"
	dedicatedCustomerServiceAvatarObjectNamespace = "customer-service/avatars"
	dedicatedUGCShowcaseLifecycleNamespace        = "ugc-showcase"
)

// A namespace lock also exists before the first owner row is inserted. Every
// dedicated owner write and cleanup must acquire it before locking owner rows.
// No registration/lock table or second reference ledger is needed.
func runWithDedicatedPublicObjectNamespaceTransactionLock(ctx context.Context, db *gorm.DB, namespace string, operation func(*gorm.DB) error) error {
	if db == nil {
		return gorm.ErrInvalidDB
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var options []*sql.TxOptions
	if db.Name() == "postgres" {
		options = append(options, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := acquireDedicatedPublicObjectNamespaceLockWithinTransaction(tx, namespace); err != nil {
			return err
		}
		return operation(tx)
	}, options...)
}

func acquireDedicatedPublicObjectNamespaceLockWithinTransaction(tx *gorm.DB, namespace string) error {
	if tx.Name() != "postgres" {
		return nil
	}
	if _, active := tx.Statement.ConnPool.(gorm.TxCommitter); !active {
		return fmt.Errorf("dedicated object ownership locks require an active transaction")
	}
	var isolation string
	if err := tx.Raw("SHOW transaction_isolation").Scan(&isolation).Error; err != nil {
		return err
	}
	if isolation != "read committed" {
		return fmt.Errorf("dedicated object ownership requires read committed transaction isolation")
	}
	digest := sha256.Sum256([]byte("dedicated-public-object-lifecycle/" + namespace))
	lockID := int64(binary.BigEndian.Uint64(digest[:8]))
	if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", lockID).Error; err != nil {
		return fmt.Errorf("lock dedicated public object namespace %s: %w", namespace, err)
	}
	return nil
}
