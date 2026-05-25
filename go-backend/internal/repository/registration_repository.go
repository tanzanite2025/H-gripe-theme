package repository

import (
	"tanzanite/internal/domain/registration"
	"time"

	"gorm.io/gorm"
)

type RegistrationRepository struct {
	db *gorm.DB
}

func NewRegistrationRepository(db *gorm.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

// ProductRegistration 相关方法

// CreateRegistration 创建产品注册
func (r *RegistrationRepository) CreateRegistration(reg *registration.ProductRegistration) error {
	return r.db.Create(reg).Error
}

// FindRegistrationByID 根据ID查找注册记录
func (r *RegistrationRepository) FindRegistrationByID(id uint) (*registration.ProductRegistration, error) {
	var reg registration.ProductRegistration
	err := r.db.Preload("User").Preload("Product").First(&reg, id).Error
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

// FindRegistrationBySerialNumber 根据序列号查找注册记录
func (r *RegistrationRepository) FindRegistrationBySerialNumber(serialNumber string) (*registration.ProductRegistration, error) {
	var reg registration.ProductRegistration
	err := r.db.Where("serial_number = ?", serialNumber).
		Preload("User").Preload("Product").First(&reg).Error
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

// FindRegistrationsByUserID 查找用户的注册记录
func (r *RegistrationRepository) FindRegistrationsByUserID(userID uint, page, pageSize int) ([]registration.ProductRegistration, int64, error) {
	var registrations []registration.ProductRegistration
	var total int64

	query := r.db.Model(&registration.ProductRegistration{}).Where("user_id = ?", userID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Product").Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&registrations).Error
	
	return registrations, total, err
}

// FindRegistrationsByProductID 查找产品的注册记录
func (r *RegistrationRepository) FindRegistrationsByProductID(productID uint, page, pageSize int) ([]registration.ProductRegistration, int64, error) {
	var registrations []registration.ProductRegistration
	var total int64

	query := r.db.Model(&registration.ProductRegistration{}).Where("product_id = ?", productID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("User").Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&registrations).Error
	
	return registrations, total, err
}

// FindAllRegistrations 查找所有注册记录（管理员）
func (r *RegistrationRepository) FindAllRegistrations(page, pageSize int, status string) ([]registration.ProductRegistration, int64, error) {
	var registrations []registration.ProductRegistration
	var total int64

	query := r.db.Model(&registration.ProductRegistration{})
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("User").Preload("Product").Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&registrations).Error
	
	return registrations, total, err
}

// UpdateRegistration 更新注册记录
func (r *RegistrationRepository) UpdateRegistration(reg *registration.ProductRegistration) error {
	return r.db.Save(reg).Error
}

// UpdateRegistrationStatus 更新注册状态
func (r *RegistrationRepository) UpdateRegistrationStatus(id uint, status string) error {
	return r.db.Model(&registration.ProductRegistration{}).Where("id = ?", id).
		Update("status", status).Error
}

// DeleteRegistration 删除注册记录
func (r *RegistrationRepository) DeleteRegistration(id uint) error {
	return r.db.Delete(&registration.ProductRegistration{}, id).Error
}

// CheckSerialNumberExists 检查序列号是否已存在
func (r *RegistrationRepository) CheckSerialNumberExists(serialNumber string) (bool, error) {
	var count int64
	err := r.db.Model(&registration.ProductRegistration{}).
		Where("serial_number = ?", serialNumber).Count(&count).Error
	return count > 0, err
}

// FindExpiringWarranties 查找即将过期的保修
func (r *RegistrationRepository) FindExpiringWarranties(days int) ([]registration.ProductRegistration, error) {
	var registrations []registration.ProductRegistration
	expiryDate := time.Now().AddDate(0, 0, days)
	
	err := r.db.Where("warranty_expires <= ? AND warranty_expires >= ? AND status = ?", 
		expiryDate, time.Now(), "active").
		Preload("User").Preload("Product").Find(&registrations).Error
	
	return registrations, err
}

// GetRegistrationStats 获取注册统计
func (r *RegistrationRepository) GetRegistrationStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 总注册数
	var totalCount int64
	if err := r.db.Model(&registration.ProductRegistration{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount
	
	// 各状态统计
	statuses := []string{"active", "expired", "cancelled"}
	for _, status := range statuses {
		var count int64
		if err := r.db.Model(&registration.ProductRegistration{}).
			Where("status = ?", status).Count(&count).Error; err != nil {
			return nil, err
		}
		stats[status+"_count"] = count
	}
	
	// 本月新增
	var monthlyCount int64
	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	if err := r.db.Model(&registration.ProductRegistration{}).
		Where("created_at >= ?", startOfMonth).Count(&monthlyCount).Error; err != nil {
		return nil, err
	}
	stats["monthly_count"] = monthlyCount
	
	return stats, nil
}

// WarrantyClaim 相关方法

// CreateWarrantyClaim 创建保修申请
func (r *RegistrationRepository) CreateWarrantyClaim(claim *registration.WarrantyClaim) error {
	return r.db.Create(claim).Error
}

// FindWarrantyClaimByID 根据ID查找保修申请
func (r *RegistrationRepository) FindWarrantyClaimByID(id uint) (*registration.WarrantyClaim, error) {
	var claim registration.WarrantyClaim
	err := r.db.Preload("Registration").Preload("Registration.User").
		Preload("Registration.Product").First(&claim, id).Error
	if err != nil {
		return nil, err
	}
	return &claim, nil
}

// FindWarrantyClaimsByRegistrationID 查找注册记录的保修申请
func (r *RegistrationRepository) FindWarrantyClaimsByRegistrationID(registrationID uint) ([]registration.WarrantyClaim, error) {
	var claims []registration.WarrantyClaim
	err := r.db.Where("registration_id = ?", registrationID).
		Order("created_at DESC").Find(&claims).Error
	return claims, err
}

// FindAllWarrantyClaims 查找所有保修申请（管理员）
func (r *RegistrationRepository) FindAllWarrantyClaims(page, pageSize int, status string) ([]registration.WarrantyClaim, int64, error) {
	var claims []registration.WarrantyClaim
	var total int64

	query := r.db.Model(&registration.WarrantyClaim{})
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Registration").Preload("Registration.User").
		Preload("Registration.Product").Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&claims).Error
	
	return claims, total, err
}

// UpdateWarrantyClaim 更新保修申请
func (r *RegistrationRepository) UpdateWarrantyClaim(claim *registration.WarrantyClaim) error {
	return r.db.Save(claim).Error
}

// UpdateWarrantyClaimStatus 更新保修申请状态
func (r *RegistrationRepository) UpdateWarrantyClaimStatus(id uint, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	
	if status == "approved" {
		updates["approved_at"] = time.Now()
	} else if status == "completed" {
		updates["completed_at"] = time.Now()
	}
	
	return r.db.Model(&registration.WarrantyClaim{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteWarrantyClaim 删除保修申请
func (r *RegistrationRepository) DeleteWarrantyClaim(id uint) error {
	return r.db.Delete(&registration.WarrantyClaim{}, id).Error
}
