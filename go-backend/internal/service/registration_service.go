package service

import (
	"errors"
	"tanzanite/internal/domain/registration"
	"tanzanite/internal/repository"
	"time"
)

type RegistrationService struct {
	registrationRepo *repository.RegistrationRepository
	productRepo      *repository.ProductRepository
	orderRepo        *repository.OrderRepository
}

func NewRegistrationService(
	registrationRepo *repository.RegistrationRepository,
	productRepo *repository.ProductRepository,
	orderRepo *repository.OrderRepository,
) *RegistrationService {
	return &RegistrationService{
		registrationRepo: registrationRepo,
		productRepo:      productRepo,
		orderRepo:        orderRepo,
	}
}

// ProductRegistration 相关方法

// CreateRegistration 创建产品注册
func (s *RegistrationService) CreateRegistration(reg *registration.ProductRegistration) error {
	// 检查序列号是否已存在
	exists, err := s.registrationRepo.CheckSerialNumberExists(reg.SerialNumber)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("serial number already registered")
	}

	// 验证产品是否存在
	_, err = s.productRepo.FindByID(reg.ProductID)
	if err != nil {
		return errors.New("product not found")
	}

	// 计算保修到期日期
	if reg.WarrantyPeriod > 0 {
		reg.WarrantyExpires = reg.PurchaseDate.AddDate(0, reg.WarrantyPeriod, 0)
	}

	// 设置默认状态
	if reg.Status == "" {
		reg.Status = "active"
	}

	return s.registrationRepo.CreateRegistration(reg)
}

// GetRegistration 获取注册记录
func (s *RegistrationService) GetRegistration(id uint, userID uint, isAdmin bool) (*registration.ProductRegistration, error) {
	reg, err := s.registrationRepo.FindRegistrationByID(id)
	if err != nil {
		return nil, err
	}

	// 验证权限
	if !isAdmin && reg.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return reg, nil
}

// GetUserRegistrations 获取用户的注册记录
func (s *RegistrationService) GetUserRegistrations(userID uint, page, pageSize int) ([]registration.ProductRegistration, int64, error) {
	return s.registrationRepo.FindRegistrationsByUserID(userID, page, pageSize)
}

// GetAllRegistrations 获取所有注册记录（管理员）
func (s *RegistrationService) GetAllRegistrations(page, pageSize int, status string) ([]registration.ProductRegistration, int64, error) {
	return s.registrationRepo.FindAllRegistrations(page, pageSize, status)
}

// UpdateRegistration 更新注册记录
func (s *RegistrationService) UpdateRegistration(reg *registration.ProductRegistration, userID uint, isAdmin bool) error {
	existing, err := s.registrationRepo.FindRegistrationByID(reg.ID)
	if err != nil {
		return err
	}

	// 验证权限
	if !isAdmin && existing.UserID != userID {
		return errors.New("unauthorized")
	}

	// 重新计算保修到期日期
	if reg.WarrantyPeriod > 0 && reg.PurchaseDate != existing.PurchaseDate {
		reg.WarrantyExpires = reg.PurchaseDate.AddDate(0, reg.WarrantyPeriod, 0)
	}

	return s.registrationRepo.UpdateRegistration(reg)
}

// UpdateRegistrationStatus 更新注册状态
func (s *RegistrationService) UpdateRegistrationStatus(id uint, status string) error {
	// 验证状态
	validStatuses := []string{"active", "expired", "cancelled"}
	isValid := false
	for _, s := range validStatuses {
		if s == status {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.New("invalid status")
	}

	return s.registrationRepo.UpdateRegistrationStatus(id, status)
}

// DeleteRegistration 删除注册记录
func (s *RegistrationService) DeleteRegistration(id uint, userID uint, isAdmin bool) error {
	reg, err := s.registrationRepo.FindRegistrationByID(id)
	if err != nil {
		return err
	}

	// 验证权限
	if !isAdmin && reg.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.registrationRepo.DeleteRegistration(id)
}

// VerifySerialNumber 验证序列号
func (s *RegistrationService) VerifySerialNumber(serialNumber string) (*registration.ProductRegistration, error) {
	return s.registrationRepo.FindRegistrationBySerialNumber(serialNumber)
}

// GetExpiringWarranties 获取即将过期的保修
func (s *RegistrationService) GetExpiringWarranties(days int) ([]registration.ProductRegistration, error) {
	return s.registrationRepo.FindExpiringWarranties(days)
}

// GetRegistrationStats 获取注册统计
func (s *RegistrationService) GetRegistrationStats() (map[string]interface{}, error) {
	return s.registrationRepo.GetRegistrationStats()
}

// CheckExpiredWarranties 检查并更新过期保修
func (s *RegistrationService) CheckExpiredWarranties() error {
	// 查找所有活跃的注册
	registrations, _, err := s.registrationRepo.FindAllRegistrations(1, 1000, "active")
	if err != nil {
		return err
	}

	now := time.Now()
	for _, reg := range registrations {
		if reg.WarrantyExpires.Before(now) {
			// 更新为过期状态
			if err := s.registrationRepo.UpdateRegistrationStatus(reg.ID, "expired"); err != nil {
				return err
			}
		}
	}

	return nil
}
