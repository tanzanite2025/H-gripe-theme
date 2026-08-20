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
	err := r.db.
		Preload("Items").
		Preload("Attachments").
		Preload("RefundReview").
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
	err := r.lockForUpdate(r.db).
		Preload("Items").
		Preload("Attachments").
		Preload("RefundReview").
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

	err := query.
		Select("after_sales_cases.*, orders.order_number AS order_number").
		Preload("Items").
		Order("after_sales_cases.created_at DESC, after_sales_cases.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error

	return records, total, err
}

func (r *AfterSalesCaseRepository) SumActiveQuantity(orderItemID uint) (int, error) {
	var quantity int64
	err := r.db.
		Model(&aftersales.AfterSalesCaseItem{}).
		Joins("JOIN after_sales_cases ON after_sales_cases.id = after_sales_case_items.case_id").
		Where("after_sales_case_items.order_item_id = ?", orderItemID).
		Where("after_sales_cases.status NOT IN ?", []string{
			aftersales.StatusRejected,
			aftersales.StatusCancelled,
		}).
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
