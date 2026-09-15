package repository

import (
	"errors"
	"strings"

	suppliercostdomain "commerce-platform/internal/domain/productsuppliercost"

	"gorm.io/gorm"
)

// ProductSupplierCostCatalogRepository is a read-only adapter for the
// supplier-cost SKU picker. It deliberately does not use ProductRepository,
// so the picker cannot enter product write transactions or preload the full
// catalog.
type ProductSupplierCostCatalogRepository struct {
	db *gorm.DB
}

type ProductSupplierCostCatalogFilter struct {
	Page     int
	PageSize int
	Search   string
	SKU      string
}

func NewProductSupplierCostCatalogRepository(db *gorm.DB) *ProductSupplierCostCatalogRepository {
	return &ProductSupplierCostCatalogRepository{db: db}
}

func (r *ProductSupplierCostCatalogRepository) ListOptions(filter ProductSupplierCostCatalogFilter) ([]suppliercostdomain.ProductOption, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("product supplier cost catalog repository is unavailable")
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	query := r.optionQuery()

	if sku := strings.TrimSpace(filter.SKU); sku != "" {
		// Exact SKU lookup is used only to reopen an existing record. It can
		// return an inactive item so historical cost data remains editable.
		query = query.Where("pv.sku = ?", sku)
	} else {
		query = query.Where("p.status = ?", "active").Where("pv.is_active = ?", true)
		if search := strings.TrimSpace(filter.Search); search != "" {
			like := "%" + strings.ToLower(search) + "%"
			query = query.Where(
				"LOWER(pv.sku) LIKE ? OR LOWER(p.name) LIKE ? OR LOWER(COALESCE(pv.title, '')) LIKE ?",
				like, like, like,
			)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var options []suppliercostdomain.ProductOption
	if err := query.
		Order("p.name ASC").
		Order("pv.title ASC").
		Order("pv.sku ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&options).Error; err != nil {
		return nil, 0, err
	}
	if options == nil {
		options = []suppliercostdomain.ProductOption{}
	}
	return options, total, nil
}

func (r *ProductSupplierCostCatalogRepository) FindOptionBySKU(sku string) (*suppliercostdomain.ProductOption, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("product supplier cost catalog repository is unavailable")
	}

	sku = strings.TrimSpace(sku)
	if sku == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var options []suppliercostdomain.ProductOption
	if err := r.optionQuery().
		Where("pv.sku = ?", sku).
		Limit(2).
		Scan(&options).Error; err != nil {
		return nil, err
	}
	if len(options) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if len(options) > 1 {
		return nil, errors.New("product supplier cost catalog contains duplicate SKU")
	}
	return &options[0], nil
}

func (r *ProductSupplierCostCatalogRepository) optionQuery() *gorm.DB {
	return r.db.Table("product_variants AS pv").
		Select(`
			p.name AS product_name,
			COALESCE(pv.title, '') AS variant_title,
			pv.sku AS sku,
			(p.status = 'active' AND pv.is_active = TRUE) AS available
		`).
		Joins("JOIN products AS p ON p.id = pv.product_id").
		Where("p.deleted_at IS NULL").
		Where("pv.deleted_at IS NULL")
}
