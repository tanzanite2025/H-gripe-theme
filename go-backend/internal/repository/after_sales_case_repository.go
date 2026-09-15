package repository

import (
	"commerce-platform/internal/domain/aftersales"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AfterSalesCaseRepository struct {
	db *gorm.DB
}

func NewAfterSalesCaseRepository(db *gorm.DB) *AfterSalesCaseRepository {
	return &AfterSalesCaseRepository{db: db}
}

func (r *AfterSalesCaseRepository) WithTx(tx *gorm.DB) *AfterSalesCaseRepository {
	return &AfterSalesCaseRepository{db: tx}
}

func (r *AfterSalesCaseRepository) lockForUpdate(query *gorm.DB) *gorm.DB {
	switch r.db.Dialector.Name() {
	case "postgres", "mysql", "sqlserver":
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	default:
		return query
	}
}

func (r *AfterSalesCaseRepository) preloadReturnShipments(query *gorm.DB) *gorm.DB {
	if r == nil || r.db == nil || !r.db.Migrator().HasTable(&aftersales.AfterSalesReturnShipment{}) {
		return query
	}
	return query.Preload("ReturnShipments", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC, id ASC")
	})
}

func (r *AfterSalesCaseRepository) CreateWithItems(
	caseRecord *aftersales.AfterSalesCase,
	items []aftersales.AfterSalesCaseItem,
) error {
	return r.CreateWithItemsAndAttachments(caseRecord, items, nil)
}

func (r *AfterSalesCaseRepository) CreateWithItemsAndAttachments(
	caseRecord *aftersales.AfterSalesCase,
	items []aftersales.AfterSalesCaseItem,
	attachments []aftersales.AfterSalesCaseAttachment,
) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		eventTime := time.Now().UTC()
		record := *caseRecord
		record.Items = nil
		record.Events = nil
		record.Attachments = nil
		record.ReturnShipments = nil
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		caseRecord.ID = record.ID
		for index := range items {
			items[index].CaseID = caseRecord.ID
			if err := tx.Create(&items[index]).Error; err != nil {
				return err
			}
		}
		for index := range attachments {
			attachments[index].CaseID = caseRecord.ID
			if err := tx.Create(&attachments[index]).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&aftersales.AfterSalesCaseEvent{
			CaseID:     caseRecord.ID,
			ToStatus:   record.Status,
			Resolution: "售后单创建",
			UpdatedBy:  record.CreatedBy,
			CreatedAt:  eventTime,
		}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *AfterSalesCaseRepository) FindByID(id uint) (*aftersales.AfterSalesCase, error) {
	var record aftersales.AfterSalesCase
	query := r.db.Preload("Items").Preload("Attachments").Preload("RefundReview")
	query = r.preloadReturnShipments(query)
	err := query.
		First(&record, id).Error
	if err != nil {
		return nil, err
	}
	var events []aftersales.AfterSalesCaseEvent
	eventQuery := r.db.Model(&aftersales.AfterSalesCaseEvent{})
	if r.db.Migrator().HasTable(&aftersales.AfterSalesCaseEventArchive{}) {
		eventQuery = r.db.Table(
			"(SELECT * FROM after_sales_case_events UNION ALL SELECT * FROM after_sales_case_events_archive) AS after_sales_case_events",
		)
	}
	if err := eventQuery.
		Where("case_id = ?", id).
		Order("created_at ASC, id ASC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	record.Events = events
	return &record, nil
}

func (r *AfterSalesCaseRepository) FindByIDForUpdate(id uint) (*aftersales.AfterSalesCase, error) {
	var record aftersales.AfterSalesCase
	query := r.lockForUpdate(r.db).Preload("Items").Preload("Attachments").Preload("RefundReview")
	query = r.preloadReturnShipments(query)
	err := query.
		First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *AfterSalesCaseRepository) FindAttachment(
	caseID uint,
	attachmentID uint,
) (*aftersales.AfterSalesCaseAttachment, error) {
	var attachment aftersales.AfterSalesCaseAttachment
	err := r.db.
		Where("case_id = ? AND id = ?", caseID, attachmentID).
		First(&attachment).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r *AfterSalesCaseRepository) FindByOrderID(orderID uint, status string) ([]aftersales.AfterSalesCase, error) {
	var records []aftersales.AfterSalesCase
	query := r.db.Preload("Items").Where("order_id = ?", orderID)
	query = r.preloadReturnShipments(query)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC, id DESC").Find(&records).Error
	return records, err
}

func (r *AfterSalesCaseRepository) List(
	page int,
	pageSize int,
	status string,
	caseType string,
	search string,
) ([]aftersales.AfterSalesCase, int64, error) {
	var records []aftersales.AfterSalesCase
	var total int64

	query := r.db.Model(&aftersales.AfterSalesCase{}).
		Joins("JOIN orders ON orders.id = after_sales_cases.order_id")

	if status != "" {
		query = query.Where("after_sales_cases.status = ?", status)
	}
	if caseType != "" {
		query = query.Where("after_sales_cases.type = ?", caseType)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"LOWER(orders.order_number) LIKE LOWER(?) OR LOWER(after_sales_cases.reason) LIKE LOWER(?)",
			like,
			like,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query = query.Select("after_sales_cases.*, orders.order_number AS order_number").Preload("Items")
	err := query.
		Order("after_sales_cases.created_at DESC, after_sales_cases.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error

	return records, total, err
}

// SumActiveQuantity counts quantity currently reserved by an in-flight,
// operator-managed after-sales case. Customer-originated request snapshots
// are informational and never reserve eligibility; completed, rejected, and
// cancelled cases have released their reservation as well.
func (r *AfterSalesCaseRepository) SumActiveQuantity(orderItemID uint) (int, error) {
	var quantity int64
	err := r.db.
		Model(&aftersales.AfterSalesCaseItem{}).
		Joins("JOIN after_sales_cases ON after_sales_cases.id = after_sales_case_items.case_id").
		Where("after_sales_case_items.order_item_id = ?", orderItemID).
		Where("after_sales_cases.status NOT IN ?", []string{
			aftersales.StatusCompleted,
			aftersales.StatusRejected,
			aftersales.StatusCancelled,
		}).
		Where("after_sales_cases.type <> ?", aftersales.TypeCustomerRequest).
		Select("COALESCE(SUM(after_sales_case_items.quantity), 0)").
		Scan(&quantity).Error
	return int(quantity), err
}

func (r *AfterSalesCaseRepository) UpdateStatusIfCurrent(
	id uint,
	currentStatus string,
	status string,
	resolution string,
	updatedBy uint,
) (bool, error) {
	var updated bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		eventTime := time.Now().UTC()
		updates := map[string]interface{}{
			"status":     status,
			"resolution": resolution,
			"updated_by": updatedBy,
			"updated_at": eventTime,
		}
		if aftersales.IsTerminalStatus(status) {
			updates["closed_at"] = eventTime
		}
		result := tx.Model(&aftersales.AfterSalesCase{}).
			Where("id = ? AND status = ?", id, currentStatus).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		if err := tx.Create(&aftersales.AfterSalesCaseEvent{
			CaseID:     id,
			FromStatus: currentStatus,
			ToStatus:   status,
			Resolution: resolution,
			UpdatedBy:  updatedBy,
			CreatedAt:  eventTime,
		}).Error; err != nil {
			return err
		}
		updated = true
		return nil
	})
	return updated, err
}

// UpdateStatusIfCurrentInTx performs the same guarded transition without
// opening a nested transaction. Callers that already own the surrounding
// transaction can use it to keep the case event atomic with other changes.
func (r *AfterSalesCaseRepository) UpdateStatusIfCurrentInTx(
	id uint,
	currentStatus string,
	status string,
	resolution string,
	updatedBy uint,
) (bool, error) {
	eventTime := time.Now().UTC()
	updates := map[string]interface{}{
		"status":     status,
		"resolution": resolution,
		"updated_by": updatedBy,
		"updated_at": eventTime,
	}
	if aftersales.IsTerminalStatus(status) {
		updates["closed_at"] = eventTime
	}
	result := r.db.Model(&aftersales.AfterSalesCase{}).
		Where("id = ? AND status = ?", id, currentStatus).
		Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, nil
	}

	if err := r.db.Create(&aftersales.AfterSalesCaseEvent{
		CaseID:     id,
		FromStatus: currentStatus,
		ToStatus:   status,
		Resolution: resolution,
		UpdatedBy:  updatedBy,
		CreatedAt:  eventTime,
	}).Error; err != nil {
		return false, err
	}
	return true, nil
}

// UpdateStatusAndSaveReturnShipmentIfCurrent keeps a physical return-package
// update and its case transition in the same transaction. A failed package
// save therefore cannot leave a case claiming that it is in transit or has
// been received without the evidence that proves it.
func (r *AfterSalesCaseRepository) UpdateStatusAndSaveReturnShipmentIfCurrent(
	id uint,
	currentStatus string,
	status string,
	resolution string,
	updatedBy uint,
	shipment *aftersales.AfterSalesReturnShipment,
) (bool, error) {
	var updated bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		eventTime := time.Now().UTC()
		updates := map[string]interface{}{
			"status":     status,
			"resolution": resolution,
			"updated_by": updatedBy,
			"updated_at": eventTime,
		}
		if aftersales.IsTerminalStatus(status) {
			updates["closed_at"] = eventTime
		}
		result := tx.Model(&aftersales.AfterSalesCase{}).
			Where("id = ? AND status = ?", id, currentStatus).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		if shipment != nil {
			shipment.CaseID = id
			shipment.UpdatedBy = updatedBy
			shipment.UpdatedAt = eventTime
			if shipment.ID == 0 {
				shipment.CreatedBy = updatedBy
				if err := tx.Create(shipment).Error; err != nil {
					return err
				}
			} else {
				shipmentUpdates := map[string]interface{}{
					"warehouse_name":    shipment.WarehouseName,
					"warehouse_address": shipment.WarehouseAddress,
					"carrier":           shipment.Carrier,
					"tracking_number":   shipment.TrackingNumber,
					"tracking_url":      shipment.TrackingURL,
					"label_url":         shipment.LabelURL,
					"shipped_at":        shipment.ShippedAt,
					"received_at":       shipment.ReceivedAt,
					"received_by":       shipment.ReceivedBy,
					"updated_by":        updatedBy,
					"updated_at":        eventTime,
				}
				shipmentResult := tx.Model(&aftersales.AfterSalesReturnShipment{}).
					Where("id = ? AND case_id = ?", shipment.ID, id).
					Updates(shipmentUpdates)
				if shipmentResult.Error != nil {
					return shipmentResult.Error
				}
				if shipmentResult.RowsAffected == 0 {
					return gorm.ErrRecordNotFound
				}
			}
		}

		if err := tx.Create(&aftersales.AfterSalesCaseEvent{
			CaseID:     id,
			FromStatus: currentStatus,
			ToStatus:   status,
			Resolution: resolution,
			UpdatedBy:  updatedBy,
			CreatedAt:  eventTime,
		}).Error; err != nil {
			return err
		}
		updated = true
		return nil
	})
	return updated, err
}
