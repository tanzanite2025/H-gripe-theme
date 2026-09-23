package service

import (
	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/warranty"
	"commerce-platform/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"commerce-platform/internal/pkg/emailtoken"
)

var (
	ErrWarrantyEmailMismatch        = errors.New("email does not match order record")
	ErrWarrantyVerificationRequired = errors.New("warranty email verification is required")
	ErrWarrantyClaimAccessRequired  = errors.New("warranty claim access verification is required")
	ErrWarrantyOrderItemMismatch    = errors.New("order item does not match warranty claim")
	ErrWarrantyOrderItemUnavailable = errors.New("order item binding is unavailable")
	ErrWarrantyExpired              = errors.New("the product warranty has expired")
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

// Claim access tokens are deliberately long lived. They are issued only after
// the customer has completed the one-time order email challenge and are bound
// to the claim id and verified email. This lets guest customers revisit the
// claim without weakening the claim-id authorization boundary.
const warrantyClaimAccessPurpose = "warranty:claim:view"
const warrantyClaimAccessTokenTTL = 100 * 365 * 24 * time.Hour

type WarrantyServiceRecordInput struct {
	ServiceType     string
	Status          string
	Summary         string
	CostAmountMinor int64
	Currency        string
	PerformedAt     *time.Time
}

func (s *WarrantyService) VerifyWarrantyOrder(orderNumber, email string) (*orderdomain.Order, error) {
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

func (s *WarrantyService) RequestWarrantyOrderVerification(orderNumber, email string) error {
	if s == nil || s.txManager == nil {
		return ErrEmailChallengeUnavailable
	}
	orderNumber = strings.TrimSpace(orderNumber)
	email = normalizeWarrantyEmail(email)
	if _, err := s.VerifyWarrantyOrder(orderNumber, email); err != nil {
		if errors.Is(err, ErrWarrantyEmailMismatch) || IsRecordNotFound(err) {
			return nil
		}
		return err
	}

	challengeSubject := warrantyOrderChallengeSubject(orderNumber, email)
	return s.txManager.WithinTx(func(repos repository.EmailChallengeTxRepositories) error {
		_, err := issueEmailChallengeWithDelivery(
			repos.EmailChallenge,
			repos.Outbox,
			s.challengeSecret,
			warrantyOrderChallengePurpose,
			email,
			challengeSubject,
			"Verify your warranty request",
			func(token string) string {
				link := fmt.Sprintf("%s/support/warranty?verification_token=%s#submit-warranty", s.baseURL, url.QueryEscape(token))
				return fmt.Sprintf(
					"Use this link to verify your warranty request:\n\n%s\n\nThe verification token expires in 24 hours and can only be used once when submitting the claim.",
					link,
				)
			},
			24*time.Hour,
		)
		return err
	})
}

func (s *WarrantyService) ValidateWarrantyOrderToken(token string) error {
	if s == nil || s.txManager == nil {
		return ErrWarrantyVerificationRequired
	}
	var validationErr error
	err := s.txManager.WithinTx(func(repos repository.EmailChallengeTxRepositories) error {
		_, validationErr = validateEmailChallenge(repos.EmailChallenge, s.challengeSecret, token, warrantyOrderChallengePurpose)
		return validationErr
	})
	if err != nil || validationErr != nil {
		return ErrWarrantyVerificationRequired
	}
	return nil
}

func (s *WarrantyService) CreateWarrantyClaimForOrder(input WarrantyClaimByOrderInput) (*warranty.WarrantyClaim, error) {
	if s == nil || s.txManager == nil {
		return nil, ErrEmailChallengeUnavailable
	}
	orderNumber := strings.TrimSpace(input.OrderNumber)
	email := normalizeWarrantyEmail(input.Email)
	var claims emailtoken.Claims
	err := s.txManager.WithinTx(func(repos repository.EmailChallengeTxRepositories) error {
		var err error
		claims, err = validateEmailChallenge(repos.EmailChallenge, s.challengeSecret, input.VerificationToken, warrantyOrderChallengePurpose)
		return err
	})
	if err != nil || claims.Email != email || claims.Subject != warrantyOrderChallengeSubject(orderNumber, email) {
		return nil, ErrWarrantyVerificationRequired
	}

	order, err := s.VerifyWarrantyOrder(orderNumber, email)
	if err != nil {
		return nil, err
	}
	if err := s.validateWarrantyWindow(order); err != nil {
		return nil, err
	}

	imagesJSON, err := json.Marshal(input.ImageURLs)
	if err != nil {
		return nil, err
	}

	claim := &warranty.WarrantyClaim{
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

	err = s.txManager.WithinTx(func(repos repository.EmailChallengeTxRepositories) error {
		if repos.Warranty == nil {
			return repository.ErrEmailChallengeTransactionNotConfigured
		}
		consumedClaims, err := consumeEmailChallenge(
			repos.EmailChallenge,
			s.challengeSecret,
			input.VerificationToken,
			warrantyOrderChallengePurpose,
		)
		if err != nil || consumedClaims.Email != email || consumedClaims.Subject != warrantyOrderChallengeSubject(orderNumber, email) {
			return ErrWarrantyVerificationRequired
		}
		return repos.Warranty.CreateWarrantyClaim(claim)
	})
	if err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *WarrantyService) validateWarrantyWindow(orderRecord *orderdomain.Order) error {
	if orderRecord == nil {
		return ErrWarrantyVerificationRequired
	}
	var expires time.Time
	if s.shipmentRepo != nil {
		shipment, err := s.shipmentRepo.FindByOrderID(orderRecord.ID)
		if err == nil && shipment != nil {
			expires = shipment.WarrantyExpires
		} else if err != nil && !repository.IsRecordNotFound(err) {
			// The shipment_records migration is optional for legacy installations;
			// fall back to the order's shipped_at policy when it is not present yet.
			message := strings.ToLower(err.Error())
			if !strings.Contains(message, "no such table") && !strings.Contains(message, "doesn't exist") {
				return err
			}
		}
	}
	// Keep legacy paid-order claims working when no shipment fact exists. Once
	// a shipment timestamp is present, derive the default 12-month warranty
	// boundary if the optional shipment record is unavailable.
	if expires.IsZero() && orderRecord.ShippedAt != nil && !orderRecord.ShippedAt.IsZero() {
		expires = orderRecord.ShippedAt.UTC().AddDate(1, 0, 0)
	}
	if !expires.IsZero() && time.Now().UTC().After(expires.UTC()) {
		return ErrWarrantyExpired
	}
	return nil
}

// IssueWarrantyClaimAccessToken creates a long-lived, signed read token for a
// guest claim. The token contains no mutable claim data and is safe to return
// to the verified customer after submission.
func (s *WarrantyService) IssueWarrantyClaimAccessToken(claim *warranty.WarrantyClaim) (string, error) {
	if claim == nil || claim.ID == 0 || normalizeWarrantyEmail(claim.Email) == "" {
		return "", ErrWarrantyClaimAccessRequired
	}
	if strings.TrimSpace(s.challengeSecret) == "" {
		return "", ErrEmailChallengeUnavailable
	}
	now := time.Now()
	return emailtoken.Sign(s.challengeSecret, emailtoken.Claims{
		Purpose:   warrantyClaimAccessPurpose,
		Email:     normalizeWarrantyEmail(claim.Email),
		Subject:   warrantyClaimAccessSubject(claim.ID, claim.Email),
		ExpiresAt: now.Add(warrantyClaimAccessTokenTTL).Unix(),
	})
}

func (s *WarrantyService) validateWarrantyClaimAccessToken(claim *warranty.WarrantyClaim, token string) error {
	if claim == nil || claim.ID == 0 || strings.TrimSpace(token) == "" {
		return ErrWarrantyClaimAccessRequired
	}
	if strings.TrimSpace(s.challengeSecret) == "" {
		return ErrEmailChallengeUnavailable
	}
	claims, err := emailtoken.Verify(s.challengeSecret, strings.TrimSpace(token), warrantyClaimAccessPurpose, time.Now())
	if err != nil || !strings.EqualFold(claims.Email, normalizeWarrantyEmail(claim.Email)) || claims.Subject != warrantyClaimAccessSubject(claim.ID, claim.Email) {
		return ErrWarrantyClaimAccessRequired
	}
	return nil
}

func warrantyClaimAccessSubject(claimID uint, email string) string {
	return fmt.Sprintf("%d|%s", claimID, normalizeWarrantyEmail(email))
}

func warrantyOrderChallengeSubject(orderNumber, email string) string {
	return strings.TrimSpace(orderNumber) + "|" + normalizeWarrantyEmail(email)
}

func normalizeWarrantyEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// GetWarrantyClaim 获取保修申请
func (s *WarrantyService) GetWarrantyClaim(id uint, userID uint, isAdmin bool) (*warranty.WarrantyClaim, error) {
	return s.GetWarrantyClaimForViewer(id, userID, "", "", isAdmin)
}

// GetWarrantyClaimForViewer authorizes registered users by account ownership,
// and also permits a registered account to access a legacy guest claim when
// its verified login email matches the claim email. Guest claims require the
// signed access token returned after verified submission.
func (s *WarrantyService) GetWarrantyClaimForViewer(id uint, userID uint, userEmail, accessToken string, isAdmin bool) (*warranty.WarrantyClaim, error) {
	claim, err := s.warrantyRepo.FindWarrantyClaimByID(id)
	if err != nil {
		return nil, err
	}

	if isAdmin {
		return claim, nil
	}
	if claim.UserID == userID && userID > 0 {
		return claim, nil
	}
	if claim.UserID == 0 {
		if normalizedUserEmail := normalizeWarrantyEmail(userEmail); normalizedUserEmail != "" && normalizedUserEmail == normalizeWarrantyEmail(claim.Email) {
			return claim, nil
		}
		if err := s.validateWarrantyClaimAccessToken(claim, accessToken); err == nil {
			return claim, nil
		}
	}

	return nil, errors.New("unauthorized")
}

// GetAllWarrantyClaims 获取所有保修申请（管理员）
func (s *WarrantyService) GetAllWarrantyClaims(page, pageSize int, status string) ([]warranty.WarrantyClaim, int64, error) {
	return s.warrantyRepo.FindAllWarrantyClaims(page, pageSize, status)
}

// UpdateWarrantyClaim 更新保修申请
func (s *WarrantyService) UpdateWarrantyClaim(claim *warranty.WarrantyClaim, userID uint, isAdmin bool) error {
	existing, err := s.warrantyRepo.FindWarrantyClaimByID(claim.ID)
	if err != nil {
		return err
	}

	// 验证权限
	if !isAdmin && existing.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.warrantyRepo.UpdateWarrantyClaim(claim)
}

// UpdateWarrantyClaimStatus 更新保修申请状态
func (s *WarrantyService) UpdateWarrantyClaimStatus(id uint, status string, processedBy uint) error {
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
	claim, err := s.warrantyRepo.FindWarrantyClaimByID(id)
	if err != nil {
		return err
	}

	// 更新状态
	claim.Status = status
	claim.ProcessedBy = processedBy
	now := time.Now()
	claim.ProcessedAt = &now

	return s.warrantyRepo.UpdateWarrantyClaim(claim)
}

// UpdateWarrantyClaimResolution 更新保修申请处理备注
func (s *WarrantyService) UpdateWarrantyClaimResolution(id uint, resolution string, processedBy uint) error {
	if _, err := s.warrantyRepo.FindWarrantyClaimByID(id); err != nil {
		return err
	}

	return s.warrantyRepo.UpdateWarrantyClaimResolution(id, strings.TrimSpace(resolution), processedBy)
}

// ListWarrantyClaimOrderItems 获取保修申请可绑定订单行
func (s *WarrantyService) ListWarrantyClaimOrderItems(id uint) ([]orderdomain.OrderItem, error) {
	if s.orderRepo == nil {
		return nil, ErrWarrantyOrderItemUnavailable
	}

	claim, err := s.warrantyRepo.FindWarrantyClaimByID(id)
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
func (s *WarrantyService) BindWarrantyClaimOrderItem(id uint, orderItemID *uint) error {
	if orderItemID == nil || *orderItemID == 0 {
		return s.warrantyRepo.UpdateWarrantyClaimOrderItem(id, nil)
	}

	if s.orderRepo == nil {
		return ErrWarrantyOrderItemUnavailable
	}

	claim, err := s.warrantyRepo.FindWarrantyClaimByID(id)
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
	return s.warrantyRepo.UpdateWarrantyClaimOrderItem(id, orderItemID)
}

// ListWarrantyServiceRecords 获取保修申请服务记录
func (s *WarrantyService) ListWarrantyServiceRecords(claimID uint) ([]warranty.WarrantyServiceRecord, error) {
	if _, err := s.warrantyRepo.FindWarrantyClaimByID(claimID); err != nil {
		return nil, err
	}
	return s.warrantyRepo.FindWarrantyServiceRecords(claimID)
}

// CreateWarrantyServiceRecord 创建保修服务记录
func (s *WarrantyService) CreateWarrantyServiceRecord(claimID uint, input WarrantyServiceRecordInput, createdBy uint) (*warranty.WarrantyServiceRecord, error) {
	claim, err := s.warrantyRepo.FindWarrantyClaimByID(claimID)
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
	if input.CostAmountMinor < 0 {
		return nil, errors.New("service cost amount cannot be negative")
	}

	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		return nil, errors.New("service record currency is required")
	}

	record := &warranty.WarrantyServiceRecord{
		ClaimID:         claim.ID,
		ServiceType:     serviceType,
		Status:          status,
		Summary:         summary,
		CostAmountMinor: input.CostAmountMinor,
		Currency:        currency,
		PerformedBy:     createdBy,
		CreatedBy:       createdBy,
		PerformedAt:     input.PerformedAt,
	}

	if err := s.warrantyRepo.CreateWarrantyServiceRecord(record); err != nil {
		return nil, err
	}

	return record, nil
}

// DeleteWarrantyClaim 删除保修申请
func (s *WarrantyService) DeleteWarrantyClaim(id uint, userID uint, isAdmin bool) error {
	claim, err := s.warrantyRepo.FindWarrantyClaimByID(id)
	if err != nil {
		return err
	}

	// 验证权限
	if !isAdmin && claim.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.warrantyRepo.DeleteWarrantyClaim(id)
}
