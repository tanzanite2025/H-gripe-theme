package repository

import (
	"context"
	"errors"
	"math"
	"sort"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/product"
)

func validateProductViewCountBatchIdentityAndSortProductIDs(batchID string, counts map[uint]int64) ([]uint, error) {
	id, err := uuid.Parse(batchID)
	if err != nil || id == uuid.Nil || id.String() != batchID {
		return nil, errors.New("product view flush requires a canonical nonzero batch UUID")
	}
	ids := make([]uint, 0, len(counts))
	for productID, delta := range counts {
		if productID == 0 || uint64(productID) > math.MaxInt64 || delta <= 0 {
			return nil, errors.New("product view flush requires positive SQL product IDs and deltas")
		}
		ids = append(ids, productID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

// ApplyProductViewCountBatchWithAtomicReplayReceipts serializes duplicate
// (batch, product) claims in SQL. Any failure rolls back all new receipts and
// increments; a committed receipt with another delta is an identity conflict.
func (r *ProductRepository) ApplyProductViewCountBatchWithAtomicReplayReceipts(ctx context.Context, batchID string, counts map[uint]int64) (int, error) {
	ids, err := validateProductViewCountBatchIdentityAndSortProductIDs(batchID, counts)
	if err != nil {
		return 0, err
	}
	applied := 0
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			receipt := product.ProductViewCountFlushReceipt{BatchID: batchID, ProductID: id, Delta: counts[id]}
			insert := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "batch_id"}, {Name: "product_id"}}, DoNothing: true}).Create(&receipt)
			if insert.Error != nil {
				return insert.Error
			}
			if insert.RowsAffected == 0 {
				var existing product.ProductViewCountFlushReceipt
				if err := tx.Where("batch_id = ? AND product_id = ?", batchID, id).Take(&existing).Error; err != nil {
					return err
				}
				if existing.Delta != counts[id] {
					return errors.New("product view flush receipt conflicts with immutable batch delta")
				}
				continue
			}
			update := tx.Model(&product.Product{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("COALESCE(view_count, 0) + ?", counts[id]))
			if update.Error != nil {
				return update.Error
			}
			if update.RowsAffected == 1 {
				if err := tx.Model(&receipt).UpdateColumn("product_increment_applied", true).Error; err != nil {
					return err
				}
				applied++
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return applied, nil
}

// InspectProductViewCountBatchReplayReceipts never creates a missing receipt.
func (r *ProductRepository) InspectProductViewCountBatchReplayReceipts(ctx context.Context, batchID string, counts map[uint]int64) ([]product.ProductViewCountFlushReceipt, error) {
	if batchID == "" && len(counts) == 0 {
		var schemaProbe []product.ProductViewCountFlushReceipt
		return nil, r.db.WithContext(ctx).Select("batch_id, product_id, delta, product_increment_applied, created_at").Limit(0).Find(&schemaProbe).Error
	}
	ids, err := validateProductViewCountBatchIdentityAndSortProductIDs(batchID, counts)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	var receipts []product.ProductViewCountFlushReceipt
	if err := r.db.WithContext(ctx).Where("batch_id = ? AND product_id IN ?", batchID, ids).Order("product_id").Find(&receipts).Error; err != nil {
		return nil, err
	}
	for _, receipt := range receipts {
		if counts[receipt.ProductID] != receipt.Delta {
			return nil, errors.New("product view flush receipt conflicts with immutable batch delta")
		}
	}
	return receipts, nil
}
