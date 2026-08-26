package repository

import (
	"errors"
	"strings"

	procurementdomain "commerce-platform/internal/domain/procurement"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductProfitCalculationRepository struct {
	db *gorm.DB
}

func NewProductProfitCalculationRepository(db *gorm.DB) *ProductProfitCalculationRepository {
	return &ProductProfitCalculationRepository{db: db}
}

func (r *ProductProfitCalculationRepository) WithTx(tx *gorm.DB) *ProductProfitCalculationRepository {
	return &ProductProfitCalculationRepository{db: tx}
}

func (r *ProductProfitCalculationRepository) Transaction(fn func(*gorm.DB) error) error {
	if r == nil || r.db == nil {
		return errors.New("product profitability repository is unavailable")
	}
	return r.db.Transaction(fn)
}

func (r *ProductProfitCalculationRepository) FindByProductCodes(codes []string) ([]procurementdomain.ProductProfitCalculation, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("product profitability repository is unavailable")
	}

	normalized := normalizeProductCodes(codes)
	if len(normalized) == 0 {
		return []procurementdomain.ProductProfitCalculation{}, nil
	}

	var records []procurementdomain.ProductProfitCalculation
	if err := r.db.
		Where("product_code IN ?", normalized).
		Order("product_code ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *ProductProfitCalculationRepository) FindByProductCode(code string) (*procurementdomain.ProductProfitCalculation, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("product profitability repository is unavailable")
	}

	var record procurementdomain.ProductProfitCalculation
	if err := r.db.Where("product_code = ?", strings.TrimSpace(code)).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// BulkUpsert writes only the profitability domain. The transaction intentionally
// does not include the product repository or any product-domain event/outbox.
func (r *ProductProfitCalculationRepository) BulkUpsert(records []procurementdomain.ProductProfitCalculation) error {
	return r.ReplaceCurrentSnapshots(records, nil)
}

// ReplaceCurrentSnapshots atomically clears codes with no known cost and
// upserts the remaining current snapshots. It only operates in this domain.
func (r *ProductProfitCalculationRepository) ReplaceCurrentSnapshots(
	records []procurementdomain.ProductProfitCalculation,
	clearCodes []string,
) error {
	if r == nil || r.db == nil {
		return errors.New("product profitability repository is unavailable")
	}
	return r.Transaction(func(tx *gorm.DB) error {
		return r.ReplaceCurrentSnapshotsInTx(tx, records, clearCodes)
	})
}

func (r *ProductProfitCalculationRepository) ReplaceCurrentSnapshotsInTx(
	tx *gorm.DB,
	records []procurementdomain.ProductProfitCalculation,
	clearCodes []string,
) error {
	if tx == nil {
		return errors.New("product profitability transaction is unavailable")
	}
	normalizedClearCodes := normalizeProductCodes(clearCodes)
	if len(records) == 0 && len(normalizedClearCodes) == 0 {
		return nil
	}

	if len(normalizedClearCodes) > 0 {
		if err := tx.
			Where("product_code IN ?", normalizedClearCodes).
			Delete(&procurementdomain.ProductProfitCalculation{}).Error; err != nil {
			return err
		}
	}
	if len(records) == 0 {
		return nil
	}

	for i := range records {
		procurementdomain.NormalizeProductProfitCalculation(&records[i])
	}
	return tx.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "product_code"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"product_name",
				"currency",
				"list_price",
				"sale_price",
				"effective_selling_price",
				"purchase_price",
				"inbound_shipping_unit_cost",
				"packaging_unit_cost",
				"other_unit_cost",
				"landed_cost",
				"gross_profit",
				"gross_margin_bps",
				"calculation_status",
				"formula_version",
				"warnings",
				"calculated_at",
				"updated_at",
			}),
		}).
		Select("*").
		Create(&records).Error
}

func normalizeProductCodes(codes []string) []string {
	normalized := make([]string, 0, len(codes))
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		normalized = append(normalized, code)
	}
	return normalized
}
