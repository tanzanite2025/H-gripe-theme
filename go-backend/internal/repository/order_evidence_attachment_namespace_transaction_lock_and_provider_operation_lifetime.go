package repository

import (
	"context"
	"errors"

	"commerce-platform/internal/domain/orderevidence"
	"gorm.io/gorm"
)

const orderEvidenceAttachmentLifecycleNamespace = "order-evidence"

func (r *OrderEvidenceRepository) WithLockedOrderEvidenceAttachmentLifecycleTransaction(ctx context.Context, operation func(context.Context, *OrderEvidenceRepository) error) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if ctx == nil {
		ctx = r.db.Statement.Context
	}
	return runWithDedicatedPublicObjectNamespaceTransactionLock(ctx, r.db, orderEvidenceAttachmentLifecycleNamespace, func(tx *gorm.DB) error {
		return operation(tx.Statement.Context, r.WithTx(tx))
	})
}

func (r *OrderEvidenceRepository) WithLockedOrderEvidenceAttachmentCleanupUntilProviderReturns(ctx context.Context, operation func(*OrderEvidenceRepository, *MediaRepository) error) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(ctx, r.db, orderEvidenceAttachmentLifecycleNamespace, func(tx *gorm.DB) error {
		return operation(r.WithTx(tx), NewMediaRepository(tx))
	})
}

func (r *OrderEvidenceRepository) findMutableOrderEvidenceAttachmentBeforeDetachment(attachmentID uint) (*orderevidence.OrderEvidenceAttachment, error) {
	var attachment orderevidence.OrderEvidenceAttachment
	if err := r.db.First(&attachment, attachmentID).Error; err != nil {
		return nil, err
	}
	_, pkg, err := r.FindItemAndPackageByID(attachment.EvidenceItemID)
	if err != nil {
		return nil, err
	}
	if err := pkg.EnsureMutable(); err != nil {
		return nil, err
	}
	return &attachment, nil
}

func orderEvidenceItemRequiresImmutableAttachment(itemType string) bool {
	switch itemType {
	case orderevidence.EvidenceItemTypeProductIdentity, orderevidence.EvidenceItemTypeOutboundWeightPackaging,
		orderevidence.EvidenceItemTypeSignedPOD, orderevidence.EvidenceItemTypeSpokeQCTension:
		return true
	}
	return false
}

func (r *OrderEvidenceRepository) RequireIndependentOrderEvidenceAttachmentUploadTransaction() error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if _, nested := r.db.Statement.ConnPool.(gorm.TxCommitter); nested {
		return errors.New("order evidence upload requires its own transaction lifetime")
	}
	return nil
}
