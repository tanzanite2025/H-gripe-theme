package repository

import (
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	"gorm.io/gorm"
)

type OrderEvidenceRepository struct {
	db *gorm.DB
}

type OrderEvidenceAdminListQuery struct {
	Page           int
	PageSize       int
	Search         string
	PackageStatus  string
	HighValue      *bool
	SpokeTensionQC *bool
}

type OrderEvidenceAdminListRow struct {
	OrderID                    uint      `gorm:"column:order_id"`
	OrderNumber                string    `gorm:"column:order_number"`
	CustomerFirstName          string    `gorm:"column:customer_first_name"`
	CustomerLastName           string    `gorm:"column:customer_last_name"`
	CustomerEmail              string    `gorm:"column:customer_email"`
	OrderStatus                string    `gorm:"column:order_status"`
	PaymentStatus              string    `gorm:"column:payment_status"`
	ShippingStatus             string    `gorm:"column:shipping_status"`
	TotalAmountMinor           int64     `gorm:"column:total_amount_minor"`
	Currency                   string    `gorm:"column:currency"`
	CreatedAt                  time.Time `gorm:"column:created_at"`
	PackageID                  uint      `gorm:"column:package_id"`
	PackageVersion             int       `gorm:"column:package_version"`
	PackageStatus              string    `gorm:"column:package_status"`
	OrderTotalUSDSnapshotMinor int64     `gorm:"column:order_total_usd_snapshot_minor"`
	IsHighValue                bool      `gorm:"column:is_high_value"`
	HasSpokeTensionQC          bool      `gorm:"column:has_spoke_tension_qc"`
	TotalEvidenceItems         int       `gorm:"column:total_evidence_items"`
	CompleteEvidenceItems      int       `gorm:"column:complete_evidence_items"`
	WaivedEvidenceItems        int       `gorm:"column:waived_evidence_items"`
	PendingEvidenceItems       int       `gorm:"column:pending_evidence_items"`
}

func NewOrderEvidenceRepository(db *gorm.DB) *OrderEvidenceRepository {
	return &OrderEvidenceRepository{db: db}
}

func (r *OrderEvidenceRepository) WithTx(tx *gorm.DB) *OrderEvidenceRepository {
	return &OrderEvidenceRepository{db: tx}
}

func (r *OrderEvidenceRepository) CreatePackageWithItems(
	pkg *orderevidence.OrderEvidencePackage,
	items []orderevidence.OrderEvidenceItem,
) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if pkg == nil {
		return errors.New("order evidence package is required")
	}
	if len(items) == 0 {
		return errors.New("order evidence package items are required")
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(pkg).Error; err != nil {
			return err
		}
		for index := range items {
			if items[index].PackageID != 0 {
				return errors.New("order evidence item package_id must be empty before package creation")
			}
			if items[index].OrderID != pkg.OrderID {
				return errors.New("order evidence item order_id does not match package order_id")
			}
			items[index].PackageID = pkg.ID
		}
		return tx.Create(&items).Error
	})
}

func (r *OrderEvidenceRepository) FindLatestPackageByOrderID(orderID uint) (*orderevidence.OrderEvidencePackage, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if orderID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var pkg orderevidence.OrderEvidencePackage
	if err := r.packageQuery().
		Where("order_id = ?", orderID).
		Order("package_version DESC, id DESC").
		First(&pkg).Error; err != nil {
		return nil, err
	}
	return &pkg, nil
}

// ListAdminOrders returns every order with the latest evidence package
// summary when one exists. The left join is intentional: historical orders
// created before evidence-package generation must remain visible as
// "not_generated" instead of disappearing from the evidence workbench.
func (r *OrderEvidenceRepository) ListAdminOrders(
	input OrderEvidenceAdminListQuery,
) ([]OrderEvidenceAdminListRow, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, gorm.ErrInvalidDB
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

	query := r.adminOrderEvidenceBaseQuery(input)

	var total int64
	if err := query.Session(&gorm.Session{}).
		Distinct("orders.id").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []OrderEvidenceAdminListRow
	err := query.
		Joins("LEFT JOIN order_evidence_items AS evidence_items ON evidence_items.package_id = latest_package.id").
		Select(`
			orders.id AS order_id,
			orders.order_number AS order_number,
			orders.shipping_first_name AS customer_first_name,
			orders.shipping_last_name AS customer_last_name,
			orders.shipping_email AS customer_email,
			orders.status AS order_status,
			orders.payment_status AS payment_status,
			orders.shipping_status AS shipping_status,
			orders.total_amount_minor AS total_amount_minor,
			orders.currency AS currency,
			orders.created_at AS created_at,
			COALESCE(latest_package.id, 0) AS package_id,
			COALESCE(latest_package.package_version, 0) AS package_version,
			COALESCE(latest_package.status, '') AS package_status,
			COALESCE(latest_package.order_total_usd_snapshot_minor, 0) AS order_total_usd_snapshot_minor,
			COALESCE(latest_package.is_high_value, false) AS is_high_value,
			COALESCE(latest_package.has_spoke_tension_qc, false) AS has_spoke_tension_qc,
			COUNT(evidence_items.id) AS total_evidence_items,
			SUM(CASE WHEN evidence_items.status = 'complete' THEN 1 ELSE 0 END) AS complete_evidence_items,
			SUM(CASE WHEN evidence_items.status = 'waived' THEN 1 ELSE 0 END) AS waived_evidence_items,
			SUM(CASE
				WHEN evidence_items.id IS NOT NULL
					AND evidence_items.status NOT IN ('complete', 'waived')
				THEN 1
				ELSE 0
			END) AS pending_evidence_items
		`).
		Group(`
			orders.id,
			orders.order_number,
			orders.shipping_first_name,
			orders.shipping_last_name,
			orders.shipping_email,
			orders.status,
			orders.payment_status,
			orders.shipping_status,
			orders.total_amount_minor,
			orders.currency,
			orders.created_at,
			latest_package.id,
			latest_package.package_version,
			latest_package.status,
			latest_package.order_total_usd_snapshot_minor,
			latest_package.is_high_value,
			latest_package.has_spoke_tension_qc
		`).
		Order("orders.created_at DESC, orders.id DESC").
		Offset((input.Page - 1) * input.PageSize).
		Limit(input.PageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *OrderEvidenceRepository) adminOrderEvidenceBaseQuery(
	input OrderEvidenceAdminListQuery,
) *gorm.DB {
	latestPackage := r.db.
		Table("order_evidence_packages AS candidate_package").
		Select("candidate_package.*").
		Joins(`
			INNER JOIN (
				SELECT order_id, MAX(package_version) AS package_version
				FROM order_evidence_packages
				GROUP BY order_id
			) AS latest_version
			ON latest_version.order_id = candidate_package.order_id
			AND latest_version.package_version = candidate_package.package_version
		`)

	query := r.db.Model(&order.Order{}).
		Joins("LEFT JOIN (?) AS latest_package ON latest_package.order_id = orders.id", latestPackage)

	search := strings.TrimSpace(input.Search)
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where(`
			LOWER(orders.order_number) LIKE ?
			OR LOWER(COALESCE(orders.shipping_email, '')) LIKE ?
			OR LOWER(COALESCE(orders.shipping_first_name, '')) LIKE ?
			OR LOWER(COALESCE(orders.shipping_last_name, '')) LIKE ?
			OR CAST(orders.id AS TEXT) LIKE ?
		`, like, like, like, like, like)
	}
	if status := strings.TrimSpace(input.PackageStatus); status != "" {
		query = query.Where("latest_package.status = ?", status)
	}
	if input.HighValue != nil {
		query = query.Where("COALESCE(latest_package.is_high_value, false) = ?", *input.HighValue)
	}
	if input.SpokeTensionQC != nil {
		query = query.Where("COALESCE(latest_package.has_spoke_tension_qc, false) = ?", *input.SpokeTensionQC)
	}
	return query
}

func (r *OrderEvidenceRepository) FindPackageByIDAndOrderID(
	packageID uint,
	orderID uint,
) (*orderevidence.OrderEvidencePackage, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if packageID == 0 || orderID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var pkg orderevidence.OrderEvidencePackage
	if err := r.packageQuery().
		Where("id = ? AND order_id = ?", packageID, orderID).
		First(&pkg).Error; err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *OrderEvidenceRepository) ListItemsByPackageID(packageID uint) ([]orderevidence.OrderEvidenceItem, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var items []orderevidence.OrderEvidenceItem
	err := r.db.Where("package_id = ?", packageID).
		Order("id ASC").
		Find(&items).Error
	return items, err
}

func (r *OrderEvidenceRepository) FindItemByID(id uint) (*orderevidence.OrderEvidenceItem, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var item orderevidence.OrderEvidenceItem
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OrderEvidenceRepository) FindItemAndPackageByIDForOrder(
	itemID uint,
	orderID uint,
) (*orderevidence.OrderEvidenceItem, *orderevidence.OrderEvidencePackage, error) {
	if r == nil || r.db == nil {
		return nil, nil, gorm.ErrInvalidDB
	}
	if itemID == 0 || orderID == 0 {
		return nil, nil, gorm.ErrRecordNotFound
	}

	var item orderevidence.OrderEvidenceItem
	if err := r.db.
		Where("id = ? AND order_id = ?", itemID, orderID).
		First(&item).Error; err != nil {
		return nil, nil, err
	}

	pkg, err := r.FindPackageByIDAndOrderID(item.PackageID, orderID)
	if err != nil {
		return nil, nil, err
	}
	if pkg.ID != item.PackageID || pkg.OrderID != item.OrderID {
		return nil, nil, errors.New("order evidence item package order scope mismatch")
	}
	if item.OrderItemID != nil {
		var orderItemCount int64
		if err := r.db.Model(&order.OrderItem{}).
			Where("id = ? AND order_id = ?", *item.OrderItemID, orderID).
			Count(&orderItemCount).Error; err != nil {
			return nil, nil, err
		}
		if orderItemCount == 0 {
			return nil, nil, gorm.ErrRecordNotFound
		}
	}
	return &item, pkg, nil
}

func (r *OrderEvidenceRepository) CreateAttachment(attachment *orderevidence.OrderEvidenceAttachment) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.Create(attachment).Error
}

func (r *OrderEvidenceRepository) DeleteAttachment(attachmentID uint) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if attachmentID == 0 {
		return gorm.ErrRecordNotFound
	}
	// Attachment rows are mutable package references. The underlying object
	// remains immutable and is intentionally retained for later audit/review.
	// Use controlled SQL because the domain hook correctly blocks generic
	// GORM deletes on immutable attachment records.
	result := r.db.Exec(
		"DELETE FROM order_evidence_attachments WHERE id = ?",
		attachmentID,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *OrderEvidenceRepository) ListAttachmentsByItemID(itemID uint) ([]orderevidence.OrderEvidenceAttachment, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var attachments []orderevidence.OrderEvidenceAttachment
	err := r.db.Where("evidence_item_id = ?", itemID).
		Order("id ASC").
		Find(&attachments).Error
	return attachments, err
}

func (r *OrderEvidenceRepository) FindAttachmentByIDForOrder(
	orderID uint,
	itemID uint,
	attachmentID uint,
) (*orderevidence.OrderEvidenceAttachment, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if orderID == 0 || itemID == 0 || attachmentID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var attachment orderevidence.OrderEvidenceAttachment
	err := r.db.
		Joins("JOIN order_evidence_items ON order_evidence_items.id = order_evidence_attachments.evidence_item_id").
		Joins("JOIN order_evidence_packages ON order_evidence_packages.id = order_evidence_items.package_id").
		Where(`
			order_evidence_attachments.id = ?
			AND order_evidence_attachments.evidence_item_id = ?
			AND order_evidence_items.order_id = ?
			AND order_evidence_packages.order_id = ?
		`, attachmentID, itemID, orderID, orderID).
		First(&attachment).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r *OrderEvidenceRepository) packageQuery() *gorm.DB {
	return r.db.
		Preload("Items.Attachments").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		})
}
