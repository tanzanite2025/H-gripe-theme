package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	orderdomain "tanzanite/internal/domain/order"
	"tanzanite/internal/domain/registration"
	"time"
)

var (
	ErrWarrantyEmailMismatch        = errors.New("email does not match order record")
	ErrWarrantyVerificationRequired = errors.New("warranty email verification is required")
	ErrWarrantyOrderItemMismatch    = errors.New("order item does not match warranty claim")
	ErrWarrantyOrderItemUnavailable = errors.New("order item binding is unavailable")
)

var validWarrantyServiceTypes = map[string]struct{}{
	"inspection":  {},
	"repair":      {},
	"replacement": {},
	"refund":      {},
	"shipping":    {},
}

var validWarrantyServiceStatuses = map[string]struct{}{
	"open":       {},
	"processing": {},
	"resolved":   {},
	"closed":     {},
}

type WarrantyClaimByOrderInput struct {
	OrderNumber       string
	Email             string
	VerificationToken string
	Description       string
	TirePressure      string
	IsTubeless        bool
	ImageURLs         []string
	VideoURL          string
}

const warrantyOrderChallengePurpose = "warranty:order"

type WarrantyServiceRecordInput struct {
	ServiceType string
	Status      string
	Summary     string
	CostAmount  float64
	Currency    string
	PerformedAt *time.Time
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

func (s *RegistrationService) RequestWarrantyOrderVerification(orderNumber, email string) error {
	orderNumber = strings.TrimSpace(orderNumber)
	email = normalizeWarrantyEmail(email)
	if _, err := s.VerifyWarrantyOrder(orderNumber, email); err != nil {
		if errors.Is(err, ErrWarrantyEmailMismatch) || IsRecordNotFound(err) {
			return nil
		}
		return err
	}

	token, err := issueEmailChallenge(
		s.challengeRepo,
		s.challengeSecret,
		warrantyOrderChallengePurpose,
		email,
		warrantyOrderChallengeSubject(orderNumber, email),
		24*time.Hour,
	)
	if err != nil {
		return err
	}
	if s.emailSender == nil {
		return ErrEmailChallengeUnavailable
	}

	link := fmt.Sprintf("%s/support/warranty?verification_token=%s#submit-warranty", s.baseURL, url.QueryEscape(token))
	body := fmt.Sprintf(
		"Use this link to verify your warranty request:\n\n%s\n\nThe verification token expires in 24 hours and can only be used once when submitting the claim.",
		link,
	)
	return s.emailSender.SendEmail([]string{email}, "Verify your Tanzanite warranty request", body)
}

func (s *RegistrationService) ValidateWarrantyOrderToken(token string) error {
	if _, err := validateEmailChallenge(s.challengeRepo, s.challengeSecret, token, warrantyOrderChallengePurpose); err != nil {
		return ErrWarrantyVerificationRequired
	}
	return nil
}

func (s *RegistrationService) CreateWarrantyClaimForOrder(input WarrantyClaimByOrderInput) (*registration.WarrantyClaim, error) {
	orderNumber := strings.TrimSpace(input.OrderNumber)
	email := normalizeWarrantyEmail(input.Email)
	claims, err := consumeEmailChallenge(
		s.challengeRepo,
		s.challengeSecret,
		input.VerificationToken,
		warrantyOrderChallengePurpose,
	)
	if err != nil || claims.Email != email || claims.Subject != warrantyOrderChallengeSubject(orderNumber, email) {
		return nil, ErrWarrantyVerificationRequired
	}

	order, err := s.VerifyWarrantyOrder(orderNumber, email)
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
		OrderNumber:  orderNumber,
		Email:        email,
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

func warrantyOrderChallengeSubject(orderNumber, email string) string {
	return strings.TrimSpace(orderNumber) + "|" + normalizeWarrantyEmail(email)
}

func normalizeWarrantyEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// CreateWarrantyClaim 创建保修申请
func (s *RegistrationService) CreateWarrantyClaim(claim *registration.WarrantyClaim, userID uint) error {
	if claim.RegistrationID == nil || *claim.RegistrationID == 0 {
		return errors.New("registration is required")
	}

	// 验证注册记录
	reg, err := s.registrationRepo.FindRegistrationByID(*claim.RegistrationID)
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

// UpdateWarrantyClaimResolution 更新保修申请处理备注
func (s *RegistrationService) UpdateWarrantyClaimResolution(id uint, resolution string, processedBy uint) error {
	if _, err := s.registrationRepo.FindWarrantyClaimByID(id); err != nil {
		return err
	}

	return s.registrationRepo.UpdateWarrantyClaimResolution(id, strings.TrimSpace(resolution), processedBy)
}

// ListWarrantyClaimOrderItems 获取保修申请可绑定订单行
func (s *RegistrationService) ListWarrantyClaimOrderItems(id uint) ([]orderdomain.OrderItem, error) {
	if s.orderRepo == nil {
		return nil, ErrWarrantyOrderItemUnavailable
	}

	claim, err := s.registrationRepo.FindWarrantyClaimByID(id)
	if err != nil {
		return nil, err
	}

	orderNumber := strings.TrimSpace(claim.OrderNumber)
	if orderNumber == "" {
		return []orderdomain.OrderItem{}, nil
	}

	order, err := s.orderRepo.FindByOrderNumber(orderNumber)
	if err != nil {
		return nil, err
	}

	return order.Items, nil
}

// BindWarrantyClaimOrderItem 绑定或解绑保修申请订单行
func (s *RegistrationService) BindWarrantyClaimOrderItem(id uint, orderItemID *uint) error {
	if orderItemID == nil || *orderItemID == 0 {
		return s.registrationRepo.UpdateWarrantyClaimOrderItem(id, nil)
	}

	if s.orderRepo == nil {
		return ErrWarrantyOrderItemUnavailable
	}

	claim, err := s.registrationRepo.FindWarrantyClaimByID(id)
	if err != nil {
		return err
	}

	if strings.TrimSpace(claim.OrderNumber) == "" {
		return ErrWarrantyOrderItemMismatch
	}

	item, err := s.orderRepo.FindOrderItemByID(*orderItemID)
	if err != nil {
		return err
	}

	order, err := s.orderRepo.FindByID(item.OrderID)
	if err != nil {
		return err
	}

	if order.OrderNumber != claim.OrderNumber {
		return ErrWarrantyOrderItemMismatch
	}
	if order.UserID != 0 && claim.UserID != 0 && order.UserID != claim.UserID {
		return ErrWarrantyOrderItemMismatch
	}
	if claim.Registration != nil && claim.Registration.ProductID != 0 && claim.Registration.ProductID != item.ProductID {
		return ErrWarrantyOrderItemMismatch
	}

	return s.registrationRepo.UpdateWarrantyClaimOrderItem(id, orderItemID)
}

// ListWarrantyServiceRecords 获取保修申请服务记录
func (s *RegistrationService) ListWarrantyServiceRecords(claimID uint) ([]registration.WarrantyServiceRecord, error) {
	if _, err := s.registrationRepo.FindWarrantyClaimByID(claimID); err != nil {
		return nil, err
	}
	return s.registrationRepo.FindWarrantyServiceRecords(claimID)
}

// CreateWarrantyServiceRecord 创建保修服务记录
func (s *RegistrationService) CreateWarrantyServiceRecord(claimID uint, input WarrantyServiceRecordInput, createdBy uint) (*registration.WarrantyServiceRecord, error) {
	claim, err := s.registrationRepo.FindWarrantyClaimByID(claimID)
	if err != nil {
		return nil, err
	}

	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		return nil, errors.New("service summary is required")
	}

	serviceType := strings.ToLower(strings.TrimSpace(input.ServiceType))
	if serviceType == "" {
		serviceType = "inspection"
	}
	if _, ok := validWarrantyServiceTypes[serviceType]; !ok {
		return nil, errors.New("invalid service record type")
	}

	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = "open"
	}
	if _, ok := validWarrantyServiceStatuses[status]; !ok {
		return nil, errors.New("invalid service record status")
	}
	if input.CostAmount < 0 {
		return nil, errors.New("service cost amount cannot be negative")
	}

	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		return nil, errors.New("service record currency is required")
	}

	record := &registration.WarrantyServiceRecord{
		ClaimID:        claim.ID,
		RegistrationID: claim.RegistrationID,
		ServiceType:    serviceType,
		Status:         status,
		Summary:        summary,
		CostAmount:     input.CostAmount,
		Currency:       currency,
		PerformedBy:    createdBy,
		CreatedBy:      createdBy,
		PerformedAt:    input.PerformedAt,
	}

	if err := s.registrationRepo.CreateWarrantyServiceRecord(record); err != nil {
		return nil, err
	}

	return record, nil
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
