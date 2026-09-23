package workbenchfeed

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

type Catalog struct {
	db *gorm.DB
}

func NewCatalog(db *gorm.DB) *Catalog {
	return &Catalog{db: db}
}

type CatalogQuery struct {
	Page     int
	PageSize int
	Search   string
}

func (c *Catalog) ListOptions(input CatalogQuery) ([]ProductCatalogOption, int64, error) {
	if c == nil || c.db == nil {
		return nil, 0, errors.New("workbench product catalog is unavailable")
	}
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	query := c.optionQuery().Where("p.status = ?", "active")
	if search := strings.TrimSpace(input.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(p.name) LIKE ? OR LOWER(p.slug) LIKE ? OR LOWER(COALESCE(pv.title, '')) LIKE ? OR LOWER(COALESCE(pv.sku, '')) LIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var options []ProductCatalogOption
	if err := query.
		Order("p.name ASC").
		Order("pv.is_default DESC").
		Order("pv.title ASC").
		Order("p.id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&options).Error; err != nil {
		return nil, 0, err
	}
	if options == nil {
		options = []ProductCatalogOption{}
	}
	return options, total, nil
}

func (c *Catalog) Resolve(productID uint, variantID *uint) (*ProductCatalogOption, error) {
	if c == nil || c.db == nil {
		return nil, errors.New("workbench product catalog is unavailable")
	}
	query := c.optionQuery().Where("p.id = ?", productID)
	if variantID != nil && *variantID > 0 {
		query = query.Where("pv.id = ?", *variantID)
	} else {
		query = query.Order("pv.is_default DESC").Order("pv.id ASC")
	}

	var option ProductCatalogOption
	if err := query.Take(&option).Error; err != nil {
		return nil, err
	}
	return &option, nil
}

func (c *Catalog) optionQuery() *gorm.DB {
	return c.db.Table("products AS p").
		Select(`
			p.id AS product_id,
			pv.id AS variant_id,
			p.name AS product_name,
			p.slug AS product_slug,
			COALESCE(pv.title, '') AS variant_title,
			COALESCE(pv.sale_price_minor, pv.price_minor, p.sale_price_minor, p.price_minor)::bigint AS price_minor,
			COALESCE(NULLIF(pv.currency, ''), p.currency, 'USD') AS currency,
			(p.status = 'active' AND (pv.id IS NULL OR pv.is_active = TRUE)) AS available
		`).
		Joins("LEFT JOIN product_variants AS pv ON pv.product_id = p.id AND pv.deleted_at IS NULL").
		Where("p.deleted_at IS NULL").
		Where("pv.id IS NULL OR pv.is_active = TRUE")
}
