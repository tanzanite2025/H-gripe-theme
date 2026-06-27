package service

import (
	"encoding/json"
	"errors"
	"strings"
	orderdomain "tanzanite/internal/domain/order"
	"tanzanite/internal/domain/registration"
	"tanzanite/internal/repository"
	"time"
)

type RegistrationService struct {
	registrationRepo *repository.RegistrationRepository
	productRepo      *repository.ProductRepository
	orderRepo        *repository.OrderRepository
}

var ErrWarrantyEmailMismatch = errors.New("email does not match order record")

type WarrantyClaimByOrderInput struct {
	OrderNumber  string
	Email        string
	Description  string
	TirePressure string
	IsTubeless   bool
	ImageURLs    []string
	VideoURL     string
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

// WarrantyClaim 相关方法

func (s *RegistrationService) VerifyWarrantyOrder(orderNumber, email string) (*orderdomain.Order, error) {
	if s.orderRepo == nil {
		return nil, errors.New("order verification is unavailable")
	}

	orderNumber = strings.TrimSpace(orderNumber)
	email = strings.ToLower(strings.TrimSpace(email))
	order, err := s.orderRepo.FindByOrderNumberForVerification(orderNumber)
	if err != nil {
		return nil, err
	}

	shippingEmail := strings.ToLower(strings.TrimSpace(order.ShippingAddress.Email))
	billingEmail := strings.ToLower(strings.TrimSpace(order.BillingAddress.Email))
	if email == "" || (email != shippingEmail && email != billingEmail) {
		return nil, ErrWarrantyEmailMismatch
	}

	return order, nil
}

func (s *RegistrationService) CreateWarrantyClaimForOrder(input WarrantyClaimByOrderInput) (*registration.WarrantyClaim, error) {
	order, err := s.VerifyWarrantyOrder(input.OrderNumber, input.Email)
	if err != nil {
		return nil, err
	}

	imagesJSON, err := json.Marshal(input.ImageURLs)
	if err != nil {
		return nil, err
	}

	claim := &registration.WarrantyClaim{
		UserID:       order.UserID,
		IssueType:    "warranty",
		Description:  strings.TrimSpace(input.Description),
		Images:       string(imagesJSON),
		OrderNumber:  strings.TrimSpace(input.OrderNumber),
		Email:        strings.TrimSpace(input.Email),
		TirePressure: strings.TrimSpace(input.TirePressure),
		IsTubeless:   input.IsTubeless,
		VideoURL:     input.VideoURL,
		Status:       "submitted",
	}

	if err := s.registrationRepo.CreateWarrantyClaim(claim); err != nil {
		return nil, err
	}

	return claim, nil
}

// CreateWarrantyClaim 创建保修申请
func (s *RegistrationService) CreateWarrantyClaim(claim *registration.WarrantyClaim, userID uint) error {
	// 验证注册记录
	reg, err := s.registrationRepo.FindRegistrationByID(claim.RegistrationID)
	if err != nil {
		return errors.New("registration not found")
	}

	// 验证权限
	if reg.UserID != userID {
		return errors.New("unauthorized")
	}

	// 验证保修是否有效
	if reg.Status != "active" {
		return errors.New("warranty is not active")
	}

	if time.Now().After(reg.WarrantyExpires) {
		return errors.New("warranty has expired")
	}

	// 设置默认值
	claim.UserID = userID
	claim.Status = "submitted"
	claim.ProcessedBy = 0
	claim.ProcessedAt = nil

	return s.registrationRepo.CreateWarrantyClaim(claim)
}

// GetWarrantyClaim 获取保修申请
func (s *RegistrationService) GetWarrantyClaim(id uint, userID uint, isAdmin bool) (*registration.WarrantyClaim, error) {
	claim, err := s.registrationRepo.FindWarrantyClaimByID(id)
	if err != nil {
		return nil, err
	}

	// 验证权限
	if !isAdmin && claim.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return claim, nil
}

// GetRegistrationClaims 获取注册记录的保修申请
func (s *RegistrationService) GetRegistrationClaims(registrationID uint, userID uint, isAdmin bool) ([]registration.WarrantyClaim, error) {
	// 验证注册记录权限
	reg, err := s.registrationRepo.FindRegistrationByID(registrationID)
	if err != nil {
		return nil, err
	}

	if !isAdmin && reg.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return s.registrationRepo.FindWarrantyClaimsByRegistrationID(registrationID)
}

// GetAllWarrantyClaims 获取所有保修申请（管理员）
func (s *RegistrationService) GetAllWarrantyClaims(page, pageSize int, status string) ([]registration.WarrantyClaim, int64, error) {
	return s.registrationRepo.FindAllWarrantyClaims(page, pageSize, status)
}

// UpdateWarrantyClaim 更新保修申请
func (s *RegistrationService) UpdateWarrantyClaim(claim *registration.WarrantyClaim, userID uint, isAdmin bool) error {
	existing, err := s.registrationRepo.FindWarrantyClaimByID(claim.ID)
	if err != nil {
		return err
	}

	// 验证权限
	if !isAdmin && existing.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.registrationRepo.UpdateWarrantyClaim(claim)
}

// UpdateWarrantyClaimStatus 更新保修申请状态
func (s *RegistrationService) UpdateWarrantyClaimStatus(id uint, status string, processedBy uint) error {
	// 验证状态
	validStatuses := []string{"submitted", "reviewing", "approved", "rejected", "completed"}
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

	// 获取申请
	claim, err := s.registrationRepo.FindWarrantyClaimByID(id)
	if err != nil {
		return err
	}

	// 更新状态
	claim.Status = status
	claim.ProcessedBy = processedBy
	now := time.Now()
	claim.ProcessedAt = &now

	return s.registrationRepo.UpdateWarrantyClaim(claim)
}

// DeleteWarrantyClaim 删除保修申请
func (s *RegistrationService) DeleteWarrantyClaim(id uint, userID uint, isAdmin bool) error {
	claim, err := s.registrationRepo.FindWarrantyClaimByID(id)
	if err != nil {
		return err
	}

	// 验证权限
	if !isAdmin && claim.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.registrationRepo.DeleteWarrantyClaim(id)
}
