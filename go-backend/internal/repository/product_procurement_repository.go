package repository

import (
	"errors"
	"strings"

	procurementdomain "commerce-platform/internal/domain/procurement"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductProcurementRepository struct {
	db *gorm.DB
}

type ProductProcurementFilter struct {
	Search string
}

func NewProductProcurementRepository(db *gorm.DB) *ProductProcurementRepository {
	return &ProductProcurementRepository{db: db}
}

func (r *ProductProcurementRepository) WithTx(tx *gorm.DB) *ProductProcurementRepository {
	return &ProductProcurementRepository{db: tx}
}

func (r *ProductProcurementRepository) Transaction(fn func(*gorm.DB) error) error {
	if r == nil || r.db == nil {
		return errors.New("product procurement repository is unavailable")
	}
	return r.db.Transaction(fn)
}

func (r *ProductProcurementRepository) List(page, pageSize int, filter ProductProcurementFilter) ([]procurementdomain.ProductProcurement, int64, error) {
	var total int64
	query := r.db.Model(&procurementdomain.ProductProcurement{})

	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(product_procurement_records.product_code) LIKE ? OR LOWER(product_procurement_records.product_name) LIKE ? OR LOWER(product_procurement_records.supplier_name) LIKE ? OR LOWER(product_procurement_records.supplier_contact_name) LIKE ?",
			like, like, like, like,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var records []procurementdomain.ProductProcurement
	if err := query.
		Order("product_procurement_records.updated_at DESC").
		Order("product_procurement_records.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (r *ProductProcurementRepository) FindByID(id uint) (*procurementdomain.ProductProcurement, error) {
	var record procurementdomain.ProductProcurement
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ProductProcurementRepository) FindByProductCode(productCode string) (*procurementdomain.ProductProcurement, error) {
	var record procurementdomain.ProductProcurement
	if err := r.db.Where("product_code = ?", strings.TrimSpace(productCode)).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ProductProcurementRepository) FindByProductCodes(codes []string) ([]procurementdomain.ProductProcurement, error) {
	normalized := normalizeProductCodes(codes)
	if len(normalized) == 0 {
		return []procurementdomain.ProductProcurement{}, nil
	}

	var records []procurementdomain.ProductProcurement
	if err := r.db.
		Where("product_code IN ?", normalized).
		Order("product_code ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *ProductProcurementRepository) Create(record *procurementdomain.ProductProcurement) error {
	return r.db.Select("*").Create(record).Error
}

func (r *ProductProcurementRepository) Update(record *procurementdomain.ProductProcurement) error {
	return r.db.Model(&procurementdomain.ProductProcurement{}).
		Where("id = ?", record.ID).
		Updates(map[string]interface{}{
			"purchase_price":             record.PurchasePrice,
			"currency":                   record.Currency,
			"supplier_name":              record.SupplierName,
			"supplier_contact_name":      record.SupplierContactName,
			"supplier_phone":             record.SupplierPhone,
			"supplier_email":             record.SupplierEmail,
			"lead_time_days":             record.LeadTimeDays,
			"minimum_order_quantity":     record.MinimumOrderQuantity,
			"inbound_shipping_unit_cost": record.InboundShippingUnitCost,
			"packaging_unit_cost":        record.PackagingUnitCost,
			"other_unit_cost":            record.OtherUnitCost,
		}).Error
}

func (r *ProductProcurementRepository) Delete(id uint) error {
	return r.db.Delete(&procurementdomain.ProductProcurement{}, id).Error
}

func (r *ProductProcurementRepository) UpsertInTx(tx *gorm.DB, records []procurementdomain.ProductProcurement) error {
	if tx == nil || len(records) == 0 {
		return nil
	}

	return tx.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "product_code"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"product_name",
				"purchase_price",
				"currency",
				"supplier_name",
				"supplier_contact_name",
				"supplier_phone",
				"supplier_email",
				"lead_time_days",
				"minimum_order_quantity",
				"inbound_shipping_unit_cost",
				"packaging_unit_cost",
				"other_unit_cost",
				"updated_at",
			}),
		}).
		Select("*").
		Create(&records).Error
}

func (r *ProductProcurementRepository) DeleteByProductCodesInTx(tx *gorm.DB, codes []string) error {
	if tx == nil {
		return nil
	}
	normalized := normalizeProductCodes(codes)
	if len(normalized) == 0 {
		return nil
	}
	return tx.
		Where("product_code IN ?", normalized).
		Delete(&procurementdomain.ProductProcurement{}).Error
}
