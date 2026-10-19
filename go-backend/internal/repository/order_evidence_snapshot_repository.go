package repository

import (
	"commerce-platform/internal/domain/orderevidence"

	"gorm.io/gorm"
)

type OrderEvidenceSnapshotRepository struct {
	db                     *gorm.DB
	mediaReferenceResolver PublicBusinessMediaReferenceObjectKeyResolver
}

func NewOrderEvidenceSnapshotRepository(db *gorm.DB) *OrderEvidenceSnapshotRepository {
	return &OrderEvidenceSnapshotRepository{db: db}
}

func (r *OrderEvidenceSnapshotRepository) WithTx(tx *gorm.DB) *OrderEvidenceSnapshotRepository {
	return &OrderEvidenceSnapshotRepository{db: tx, mediaReferenceResolver: r.mediaReferenceResolver}
}

func (r *OrderEvidenceSnapshotRepository) Create(snapshot *orderevidence.OrderEvidenceSnapshot) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if snapshot == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := admitImmutableOrderEvidenceSnapshotMediaReferencesWithinTransaction(tx, r.mediaReferenceResolver, snapshot.SnapshotData, immutableOrderEvidenceSnapshotMediaAdmissionScope{orderID: snapshot.OrderID}); err != nil {
			return err
		}
		return tx.Create(snapshot).Error
	}, mediaObjectOwnershipReadCommittedTransactionOptions(r.db)...)
}

func (r *OrderEvidenceSnapshotRepository) FindByOrderID(orderID uint) (*orderevidence.OrderEvidenceSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var snapshot orderevidence.OrderEvidenceSnapshot
	if err := r.db.Where("order_id = ?", orderID).First(&snapshot).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *OrderEvidenceSnapshotRepository) ConfigureImmutableSnapshotMediaReferenceIdentityResolver(resolver PublicBusinessMediaReferenceObjectKeyResolver) {
	r.mediaReferenceResolver = resolver
}
