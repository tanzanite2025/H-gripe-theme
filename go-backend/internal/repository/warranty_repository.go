package repository

import (
	"commerce-platform/internal/domain/warranty"
	"time"

	"gorm.io/gorm"
)

type WarrantyRepository struct {
	db *gorm.DB
}

func NewWarrantyRepository(db *gorm.DB) *WarrantyRepository {
	return &WarrantyRepository{db: db}
}

// CreateWarrantyClaim 创建保修申请
func (r *WarrantyRepository) CreateWarrantyClaim(claim *warranty.WarrantyClaim) error {
	return r.db.Create(claim).Error
}

// FindWarrantyClaimByID 根据ID查找保修申请
func (r *WarrantyRepository) FindWarrantyClaimByID(id uint) (*warranty.WarrantyClaim, error) {
	var claim warranty.WarrantyClaim
	err := r.db.Preload("OrderItem").
		Preload("ServiceRecords", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&claim, id).Error
	if err != nil {
		return nil, err
	}
	return &claim, nil
}

// FindAllWarrantyClaims 查找所有保修申请（管理员）
func (r *WarrantyRepository) FindAllWarrantyClaims(page, pageSize int, status string) ([]warranty.WarrantyClaim, int64, error) {
	var claims []warranty.WarrantyClaim
	var total int64

	query := r.db.Model(&warranty.WarrantyClaim{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("OrderItem").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&claims).Error

	return claims, total, err
}

// UpdateWarrantyClaim 更新保修申请
func (r *WarrantyRepository) UpdateWarrantyClaim(claim *warranty.WarrantyClaim) error {
	return r.db.Save(claim).Error
}

// UpdateWarrantyClaimStatus 更新保修申请状态
func (r *WarrantyRepository) UpdateWarrantyClaimStatus(id uint, status string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": &now,
	}

	return r.db.Model(&warranty.WarrantyClaim{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateWarrantyClaimResolution 更新保修申请处理备注
func (r *WarrantyRepository) UpdateWarrantyClaimResolution(id uint, resolution string, processedBy uint) error {
	now := time.Now()
	updates := map[string]interface{}{
		"resolution":   resolution,
		"processed_by": processedBy,
		"processed_at": &now,
	}

	return r.db.Model(&warranty.WarrantyClaim{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateWarrantyClaimOrderItem 绑定或解绑保修申请订单行
func (r *WarrantyRepository) UpdateWarrantyClaimOrderItem(id uint, orderItemID *uint) error {
	return r.db.Model(&warranty.WarrantyClaim{}).Where("id = ?", id).
		Update("order_item_id", orderItemID).Error
}

// FindWarrantyServiceRecords 查找保修申请服务记录
func (r *WarrantyRepository) FindWarrantyServiceRecords(claimID uint) ([]warranty.WarrantyServiceRecord, error) {
	var records []warranty.WarrantyServiceRecord
	err := r.db.Where("claim_id = ?", claimID).
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}

// CreateWarrantyServiceRecord 创建保修服务记录
func (r *WarrantyRepository) CreateWarrantyServiceRecord(record *warranty.WarrantyServiceRecord) error {
	return r.db.Create(record).Error
}

// DeleteWarrantyClaim 删除保修申请
func (r *WarrantyRepository) DeleteWarrantyClaim(id uint) error {
	return r.db.Delete(&warranty.WarrantyClaim{}, id).Error
}
