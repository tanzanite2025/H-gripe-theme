package repository

import (
	"context"

	"commerce-platform/internal/domain/aftersales"
)

func (r *AfterSalesCaseRepository) FindAfterSalesAttachmentsForCurrentOrderOwner(ctx context.Context, orderNumber string, userID uint) ([]aftersales.AfterSalesCaseAttachment, error) {
	attachments := make([]aftersales.AfterSalesCaseAttachment, 0)
	if userID == 0 {
		return attachments, nil
	}
	err := r.db.WithContext(ctx).Model(&aftersales.AfterSalesCaseAttachment{}).
		Select("after_sales_case_attachments.*").
		Joins("JOIN after_sales_cases ON after_sales_cases.id = after_sales_case_attachments.case_id").
		Joins("JOIN orders ON orders.id = after_sales_cases.order_id").
		Where("orders.order_number = ? AND orders.user_id = ? AND orders.deleted_at IS NULL", orderNumber, userID).
		Find(&attachments).Error
	return attachments, err
}
