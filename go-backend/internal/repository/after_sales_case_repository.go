package repository

import (
	"commerce-platform/internal/domain/aftersales"
	"strings"
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
		return r.WithTx(tx).CreateWithItemsAndAttachmentsInTx(caseRecord, items, attachments)
	})
}

// CreateWithItemsAndAttachmentsInTx persists a case, its immutable item and
// attachment snapshots, and the initial status transition on the caller's
// transaction. The persisted transition is returned through caseRecord.Events
// so the service layer can append a canonical Outbox fact before commit.
func (r *AfterSalesCaseRepository) CreateWithItemsAndAttachmentsInTx(
	caseRecord *aftersales.AfterSalesCase,
	items []aftersales.AfterSalesCaseItem,
	attachments []aftersales.AfterSalesCaseAttachment,
) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if caseRecord == nil {
		return gorm.ErrInvalidData
	}
	eventTime := time.Now().UTC()
	record := *caseRecord
	record.Items = nil
	record.Events = nil
	record.Attachments = nil
	record.ReturnShipments = nil
	if err := r.db.Create(&record).Error; err != nil {
		return err
	}
	caseRecord.ID = record.ID
	for index := range items {
		items[index].CaseID = caseRecord.ID
		if err := r.db.Create(&items[index]).Error; err != nil {
			return err
		}
	}
	for index := range attachments {
		attachments[index].CaseID = caseRecord.ID
		if err := r.db.Create(&attachments[index]).Error; err != nil {
			return err
		}
	}
	initialEvent := &aftersales.AfterSalesCaseEvent{
		CaseID:     caseRecord.ID,
		ToStatus:   record.Status,
		Resolution: "售后单创建",
		UpdatedBy:  record.CreatedBy,
		CreatedAt:  eventTime,
	}
	if err := r.db.Create(initialEvent).Error; err != nil {
		return err
	}
	caseRecord.Events = []aftersales.AfterSalesCaseEvent{*initialEvent}
	return nil
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

// FindByReturnTrackingNumber resolves the latest return package carrying a
// tracking number and returns its full after-sales case projection. Tracking
// numbers are normalized at lookup time so scanner input can include casing or
// incidental whitespace without creating a second lookup path.
func (r *AfterSalesCaseRepository) FindByReturnTrackingNumber(trackingNumber string) (*aftersales.AfterSalesCase, error) {
	trackingNumber = strings.TrimSpace(trackingNumber)
	if trackingNumber == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var shipment aftersales.AfterSalesReturnShipment
	if err := r.db.Where("LOWER(TRIM(tracking_number)) = LOWER(?)", trackingNumber).
		Order("created_at DESC, id DESC").First(&shipment).Error; err != nil {
		return nil, err
	}
	return r.FindByID(shipment.CaseID)
}

// FindReturnShipmentByTrackingNumber returns the latest physical return parcel
// without loading customer-facing case history. It is used by signed carrier
// webhooks before applying a guarded status transition.
func (r *AfterSalesCaseRepository) FindReturnShipmentByTrackingNumber(trackingNumber string) (*aftersales.AfterSalesReturnShipment, error) {
	trackingNumber = strings.TrimSpace(trackingNumber)
	if trackingNumber == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var shipment aftersales.AfterSalesReturnShipment
	if err := r.db.Where("LOWER(TRIM(tracking_number)) = LOWER(?)", trackingNumber).
		Order("created_at DESC, id DESC").First(&shipment).Error; err != nil {
		return nil, err
	}
	return &shipment, nil
}

func (r *AfterSalesCaseRepository) UpdateReturnShipmentTrackingStatus(
	shipmentID uint,
	caseID uint,
	status string,
	receivedAt *time.Time,
	receivedBy *uint,
	updatedAt time.Time,
) error {
	if shipmentID == 0 || caseID == 0 {
		return gorm.ErrRecordNotFound
	}
	updates := map[string]interface{}{"updated_at": updatedAt}
	if receivedAt != nil {
		updates["received_at"] = receivedAt
	}
	if receivedBy != nil {
		updates["received_by"] = receivedBy
	}
	if status == "received" {
		updates["received_at"] = receivedAt
	}
	result := r.db.Model(&aftersales.AfterSalesReturnShipment{}).
		Where("id = ? AND case_id = ?", shipmentID, caseID).
		Updates(updates)
	return result.Error
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
	transition, err := r.UpdateStatusIfCurrentInTxWithEvent(id, currentStatus, status, resolution, updatedBy)
	return transition != nil, err
}

// UpdateStatusIfCurrentInTxWithEvent is the event-returning form used by
// callers that must publish a canonical Outbox fact in the same transaction.
func (r *AfterSalesCaseRepository) UpdateStatusIfCurrentInTxWithEvent(
	id uint,
	currentStatus string,
	status string,
	resolution string,
	updatedBy uint,
) (*aftersales.AfterSalesCaseEvent, error) {
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
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	transition := &aftersales.AfterSalesCaseEvent{
		CaseID:     id,
		FromStatus: currentStatus,
		ToStatus:   status,
		Resolution: resolution,
		UpdatedBy:  updatedBy,
		CreatedAt:  eventTime,
	}
	if err := r.db.Create(transition).Error; err != nil {
		return nil, err
	}
	return transition, nil
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
		transition, err := r.WithTx(tx).updateStatusAndSaveReturnShipmentIfCurrentInTx(
			id, currentStatus, status, resolution, updatedBy, shipment,
		)
		updated = transition != nil
		return err
	})
	return updated, err
}

// UpdateStatusAndSaveReturnShipmentIfCurrentInTx performs the same guarded
// transition on the caller's transaction. It returns the persisted transition
// row so callers can use its ID as an idempotency key for an Outbox fact.
func (r *AfterSalesCaseRepository) UpdateStatusAndSaveReturnShipmentIfCurrentInTx(
	id uint,
	currentStatus string,
	status string,
	resolution string,
	updatedBy uint,
	shipment *aftersales.AfterSalesReturnShipment,
) (*aftersales.AfterSalesCaseEvent, error) {
	return r.updateStatusAndSaveReturnShipmentIfCurrentInTx(
		id, currentStatus, status, resolution, updatedBy, shipment,
	)
}

func (r *AfterSalesCaseRepository) updateStatusAndSaveReturnShipmentIfCurrentInTx(
	id uint,
	currentStatus string,
	status string,
	resolution string,
	updatedBy uint,
	shipment *aftersales.AfterSalesReturnShipment,
) (*aftersales.AfterSalesCaseEvent, error) {
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
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	if shipment != nil {
		shipment.CaseID = id
		shipment.UpdatedBy = updatedBy
		shipment.UpdatedAt = eventTime
		if shipment.ID == 0 {
			shipment.CreatedBy = updatedBy
			if err := r.db.Create(shipment).Error; err != nil {
				return nil, err
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
			shipmentResult := r.db.Model(&aftersales.AfterSalesReturnShipment{}).
				Where("id = ? AND case_id = ?", shipment.ID, id).
				Updates(shipmentUpdates)
			if shipmentResult.Error != nil {
				return nil, shipmentResult.Error
			}
			if shipmentResult.RowsAffected == 0 {
				return nil, gorm.ErrRecordNotFound
			}
		}
	}

	transition := &aftersales.AfterSalesCaseEvent{
		CaseID:     id,
		FromStatus: currentStatus,
		ToStatus:   status,
		Resolution: resolution,
		UpdatedBy:  updatedBy,
		CreatedAt:  eventTime,
	}
	if err := r.db.Create(transition).Error; err != nil {
		return nil, err
	}
	return transition, nil
}
