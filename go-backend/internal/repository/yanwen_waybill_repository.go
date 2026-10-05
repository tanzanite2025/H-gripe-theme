package repository

import (
	"strings"

	"commerce-platform/internal/domain/shipping"

	"gorm.io/gorm"
)

type YanwenWaybillRepository struct{ db *gorm.DB }

func NewYanwenWaybillRepository(db *gorm.DB) *YanwenWaybillRepository {
	return &YanwenWaybillRepository{db: db}
}

func (r *YanwenWaybillRepository) FindYanwenWaybillByEnvironmentAndOrderIDAndProductCode(environment string, orderID uint, productCode string) (*shipping.YanwenWaybill, error) {
	var waybill shipping.YanwenWaybill
	err := r.db.Where("environment = ? AND order_id = ? AND product_code = ?", strings.TrimSpace(environment), orderID, strings.TrimSpace(productCode)).First(&waybill).Error
	if err != nil {
		return nil, err
	}
	return &waybill, nil
}

// FindYanwenWaybillByID loads one locally persisted Yanwen waybill.
func (r *YanwenWaybillRepository) FindYanwenWaybillByID(id uint) (*shipping.YanwenWaybill, error) {
	if id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var waybill shipping.YanwenWaybill
	if err := r.db.First(&waybill, id).Error; err != nil {
		return nil, err
	}
	return &waybill, nil
}

func (r *YanwenWaybillRepository) FindYanwenWaybills(environment, keyword, status, warehouseCode string) ([]shipping.YanwenWaybill, error) {
	waybills := make([]shipping.YanwenWaybill, 0)
	query := r.db.Order("created_at DESC").Order("id DESC")
	if environment = strings.TrimSpace(environment); environment != "" {
		query = query.Where("environment = ?", environment)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("LOWER(order_number) LIKE LOWER(?) OR LOWER(waybill_number) LIKE LOWER(?) OR LOWER(reference_number) LIKE LOWER(?) OR LOWER(yanwen_order_number) LIKE LOWER(?) OR LOWER(consignee_name) LIKE LOWER(?)", like, like, like, like, like)
	}
	if status = strings.TrimSpace(status); status != "" {
		query = query.Where("status = ?", status)
	}
	if warehouseCode = strings.TrimSpace(warehouseCode); warehouseCode != "" {
		query = query.Where("warehouse_code = ?", warehouseCode)
	}
	if err := query.Find(&waybills).Error; err != nil {
		return nil, err
	}
	return waybills, nil
}

func (r *YanwenWaybillRepository) CreateYanwenWaybill(waybill *shipping.YanwenWaybill) error {
	if waybill == nil {
		return gorm.ErrInvalidData
	}
	if err := waybill.Validate(); err != nil {
		return err
	}
	return r.db.Create(waybill).Error
}

// UpdateYanwenWaybillOfficialDetails persists one validated express.order.get result.
func (r *YanwenWaybillRepository) UpdateYanwenWaybillOfficialDetails(id uint, details shipping.YanwenOfficialWaybillDetails) error {
	if id == 0 {
		return gorm.ErrRecordNotFound
	}
	if !shipping.IsKnownYanwenOfficialWaybillStatus(details.OfficialStatus) {
		return gorm.ErrInvalidData
	}
	if details.SyncedAt.IsZero() {
		return gorm.ErrInvalidData
	}
	if len(details.ResponseData) == 0 {
		return gorm.ErrInvalidData
	}
	result := r.db.Model(&shipping.YanwenWaybill{}).Where("id = ?", id).Updates(map[string]interface{}{
		"reference_number":        strings.TrimSpace(details.ReferenceNumber),
		"yanwen_order_number":     strings.TrimSpace(details.YanwenOrderNumber),
		"official_status":         details.OfficialStatus,
		"is_printed":              details.IsPrinted,
		"last_official_synced_at": details.SyncedAt.UTC(),
		"response_data":           details.ResponseData,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateYanwenWaybillOfficialDetailsBatch persists a fully validated batch of
// express.order.getlist records atomically.
func (r *YanwenWaybillRepository) UpdateYanwenWaybillOfficialDetailsBatch(updates map[uint]shipping.YanwenOfficialWaybillDetails) error {
	if len(updates) == 0 {
		return gorm.ErrInvalidData
	}
	for id, details := range updates {
		if id == 0 || !shipping.IsKnownYanwenOfficialWaybillStatus(details.OfficialStatus) || details.SyncedAt.IsZero() || len(details.ResponseData) == 0 {
			return gorm.ErrInvalidData
		}
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for id, details := range updates {
			result := tx.Model(&shipping.YanwenWaybill{}).Where("id = ?", id).Updates(map[string]interface{}{
				"reference_number":        strings.TrimSpace(details.ReferenceNumber),
				"yanwen_order_number":     strings.TrimSpace(details.YanwenOrderNumber),
				"official_status":         details.OfficialStatus,
				"is_printed":              details.IsPrinted,
				"last_official_synced_at": details.SyncedAt.UTC(),
				"response_data":           details.ResponseData,
			})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}
		return nil
	})
}
