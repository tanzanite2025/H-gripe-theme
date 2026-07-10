package service

import (
	"encoding/json"
	"errors"
	"strings"
	orderdomain "tanzanite/internal/domain/order"
	"tanzanite/internal/domain/registration"
	"time"
)

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
