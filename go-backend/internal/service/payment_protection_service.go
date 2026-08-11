package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/config"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"
)

var (
	ErrPaymentProtectionDisabled        = errors.New("payment protection is disabled")
	ErrInvalidPaymentProtectionControl  = errors.New("invalid payment protection control")
	ErrPaymentProtectionControlNotFound = errors.New("payment protection control not found")
	ErrPaymentProtectionControlRevoked  = errors.New("payment protection control is already revoked")
)

type PaymentProtectionActor struct {
	UserID    uint
	Username  string
	IPAddress string
	UserAgent string
	Method    string
	Path      string
}

type CreatePaymentProtectionControlInput struct {
	Action     string
	ScopeType  string
	ScopeValue string
	Reason     string
	ExpiresAt  time.Time
}

type PaymentProtectionService struct {
	repo   *repository.PaymentProtectionRepository
	config config.PaymentProtectionConfig
}

func NewPaymentProtectionService(
	repo *repository.PaymentProtectionRepository,
	cfg config.PaymentProtectionConfig,
) *PaymentProtectionService {
	if cfg.MaxControlDurationHours <= 0 {
		cfg.MaxControlDurationHours = 168
	}
	if cfg.MaxPausePaymentDurationHours <= 0 {
		cfg.MaxPausePaymentDurationHours = 24
	}
	if cfg.MaxPausePaymentDurationHours > cfg.MaxControlDurationHours {
		cfg.MaxPausePaymentDurationHours = cfg.MaxControlDurationHours
	}
	if cfg.MaxGlobalPausePaymentDurationHours <= 0 {
		cfg.MaxGlobalPausePaymentDurationHours = 2
	}
	if cfg.MaxGlobalPausePaymentDurationHours > cfg.MaxPausePaymentDurationHours {
		cfg.MaxGlobalPausePaymentDurationHours = cfg.MaxPausePaymentDurationHours
	}
	return &PaymentProtectionService{
		repo:   repo,
		config: cfg,
	}
}

func (s *PaymentProtectionService) Enabled() bool {
	return s != nil && s.config.Enabled && s.repo != nil
}

func (s *PaymentProtectionService) MaxControlDuration() time.Duration {
	if s == nil || s.config.MaxControlDurationHours <= 0 {
		return 168 * time.Hour
	}
	return time.Duration(s.config.MaxControlDurationHours) * time.Hour
}

func (s *PaymentProtectionService) MaxDurationForControl(
	action paymentdomain.PaymentProtectionAction,
	scopeType paymentdomain.PaymentProtectionScope,
) time.Duration {
	if s == nil {
		return 168 * time.Hour
	}
	if action == paymentdomain.PaymentProtectionActionPausePayment {
		if scopeType == paymentdomain.PaymentProtectionScopeGlobal {
			return time.Duration(s.config.MaxGlobalPausePaymentDurationHours) * time.Hour
		}
		return time.Duration(s.config.MaxPausePaymentDurationHours) * time.Hour
	}
	return s.MaxControlDuration()
}

func (s *PaymentProtectionService) PolicySummary() map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"max_control_duration_hours":              int(s.MaxControlDuration().Hours()),
		"max_pause_payment_duration_hours":        s.config.MaxPausePaymentDurationHours,
		"max_global_pause_payment_duration_hours": s.config.MaxGlobalPausePaymentDurationHours,
	}
}

func (s *PaymentProtectionService) ListControls(
	now time.Time,
	includeExpired bool,
) ([]paymentdomain.PaymentProtectionControl, error) {
	if s == nil || s.repo == nil {
		return nil, ErrPaymentProtectionDisabled
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	controls, err := s.repo.ListControls(now, includeExpired)
	if err != nil {
		return nil, err
	}
	for index := range controls {
		controls[index].Active = controls[index].Enabled && controls[index].ExpiresAt.After(now)
		switch {
		case controls[index].Active:
			controls[index].Status = "active"
		case controls[index].Enabled:
			controls[index].Status = "expired"
		default:
			controls[index].Status = "revoked"
		}
	}
	return controls, nil
}

func (s *PaymentProtectionService) CreateControl(
	input CreatePaymentProtectionControlInput,
	actor PaymentProtectionActor,
) (*paymentdomain.PaymentProtectionControl, error) {
	if !s.Enabled() {
		return nil, ErrPaymentProtectionDisabled
	}
	now := time.Now().UTC()
	action, scopeType, scopeValue, reason, expiresAt, err := s.normalizeAndValidateControl(input, now)
	if err != nil {
		return nil, err
	}

	control := &paymentdomain.PaymentProtectionControl{
		Action:     action,
		ScopeType:  scopeType,
		ScopeValue: scopeValue,
		Reason:     reason,
		ExpiresAt:  expiresAt,
		Enabled:    true,
		CreatedBy:  actor.UserID,
		UpdatedBy:  actor.UserID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.CreateControlWithAudit(control, toPaymentProtectionAuditContext(actor)); err != nil {
		return nil, err
	}
	control.Active = true
	control.Status = "active"
	return control, nil
}

func (s *PaymentProtectionService) RevokeControl(
	id uint,
	actor PaymentProtectionActor,
) (*paymentdomain.PaymentProtectionControl, error) {
	if !s.Enabled() {
		return nil, ErrPaymentProtectionDisabled
	}
	control, err := s.repo.FindControlByID(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrPaymentProtectionControlNotFound
		}
		return nil, err
	}
	if !control.Enabled {
		return nil, ErrPaymentProtectionControlRevoked
	}
	control, err = s.repo.RevokeControlWithAudit(id, actor.UserID, toPaymentProtectionAuditContext(actor))
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrPaymentProtectionControlNotFound
		}
		return nil, err
	}
	control.Active = false
	control.Status = "revoked"
	return control, nil
}

func (s *PaymentProtectionService) ListControlAuditLogs(
	controlID uint,
	page int,
	pageSize int,
) ([]map[string]interface{}, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, ErrPaymentProtectionDisabled
	}
	logs, total, err := s.repo.ListControlAuditLogs(controlID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]map[string]interface{}, 0, len(logs))
	for _, log := range logs {
		result = append(result, map[string]interface{}{
			"id":         log.ID,
			"user_id":    log.UserID,
			"username":   log.Username,
			"action":     log.Action,
			"old_value":  log.OldValue,
			"new_value":  log.NewValue,
			"ip_address": log.IPAddress,
			"created_at": log.CreatedAt,
			"status":     log.Status,
			"error":      log.ErrorMessage,
		})
	}
	return result, total, nil
}

func (s *PaymentProtectionService) Evaluate(
	input paymentdomain.PaymentProtectionEvaluationInput,
) (paymentdomain.PaymentProtectionDecision, error) {
	if s == nil || s.repo == nil || !s.config.Enabled {
		return paymentdomain.PaymentProtectionDecision{}, nil
	}
	controls, err := s.repo.FindActiveControlsForEvaluation(time.Now().UTC())
	if err != nil {
		return paymentdomain.PaymentProtectionDecision{}, err
	}
	provider := strings.ToLower(strings.TrimSpace(input.Provider))
	country := strings.ToUpper(strings.TrimSpace(input.Country))
	paymentMethod := strings.ToLower(strings.TrimSpace(input.PaymentMethod))
	decision := paymentdomain.PaymentProtectionDecision{Reasons: []string{}}
	for _, control := range controls {
		if !paymentProtectionScopeMatches(control, provider, country, paymentMethod) {
			continue
		}
		if control.Action == paymentdomain.PaymentProtectionActionForce3DS {
			decision.Force3DS = true
			decision.Reasons = append(decision.Reasons, fmt.Sprintf("manual_force_3ds_control_%d", control.ID))
		}
		if control.Action == paymentdomain.PaymentProtectionActionPausePayment {
			decision.PausePayment = true
			decision.Reasons = append(decision.Reasons, fmt.Sprintf("manual_pause_payment_control_%d", control.ID))
		}
	}
	return decision, nil
}

func (s *PaymentProtectionService) normalizeAndValidateControl(
	input CreatePaymentProtectionControlInput,
	now time.Time,
) (
	paymentdomain.PaymentProtectionAction,
	paymentdomain.PaymentProtectionScope,
	string,
	string,
	time.Time,
	error,
) {
	action := paymentdomain.PaymentProtectionAction(strings.ToLower(strings.TrimSpace(input.Action)))
	scopeType := paymentdomain.PaymentProtectionScope(strings.ToLower(strings.TrimSpace(input.ScopeType)))
	scopeValue := strings.TrimSpace(input.ScopeValue)
	reason := strings.TrimSpace(input.Reason)
	expiresAt := input.ExpiresAt.UTC()

	if action != paymentdomain.PaymentProtectionActionForce3DS &&
		action != paymentdomain.PaymentProtectionActionPausePayment {
		return "", "", "", "", time.Time{}, ErrInvalidPaymentProtectionControl
	}
	switch scopeType {
	case paymentdomain.PaymentProtectionScopeGlobal:
		scopeValue = ""
	case paymentdomain.PaymentProtectionScopeProvider:
		scopeValue = strings.ToLower(scopeValue)
		if _, err := pgateway.ParseGatewayType(scopeValue); err != nil {
			return "", "", "", "", time.Time{}, ErrInvalidPaymentProtectionControl
		}
	case paymentdomain.PaymentProtectionScopeCountry:
		scopeValue = strings.ToUpper(scopeValue)
		if len(scopeValue) != 2 {
			return "", "", "", "", time.Time{}, ErrInvalidPaymentProtectionControl
		}
	case paymentdomain.PaymentProtectionScopePaymentMethod:
		scopeValue = strings.ToLower(scopeValue)
		if scopeValue == "" || len(scopeValue) > 128 {
			return "", "", "", "", time.Time{}, ErrInvalidPaymentProtectionControl
		}
	default:
		return "", "", "", "", time.Time{}, ErrInvalidPaymentProtectionControl
	}
	if reason == "" || len(reason) > 2000 {
		return "", "", "", "", time.Time{}, ErrInvalidPaymentProtectionControl
	}
	maxDuration := s.MaxDurationForControl(action, scopeType)
	if input.ExpiresAt.IsZero() || !expiresAt.After(now) || expiresAt.After(now.Add(maxDuration)) {
		return "", "", "", "", time.Time{}, ErrInvalidPaymentProtectionControl
	}
	return action, scopeType, scopeValue, reason, expiresAt, nil
}

func paymentProtectionScopeMatches(
	control paymentdomain.PaymentProtectionControl,
	provider string,
	country string,
	paymentMethod string,
) bool {
	switch control.ScopeType {
	case paymentdomain.PaymentProtectionScopeGlobal:
		return true
	case paymentdomain.PaymentProtectionScopeProvider:
		return provider == strings.ToLower(strings.TrimSpace(control.ScopeValue))
	case paymentdomain.PaymentProtectionScopeCountry:
		return country == strings.ToUpper(strings.TrimSpace(control.ScopeValue))
	case paymentdomain.PaymentProtectionScopePaymentMethod:
		return paymentMethod == strings.ToLower(strings.TrimSpace(control.ScopeValue))
	default:
		return false
	}
}

func toPaymentProtectionAuditContext(actor PaymentProtectionActor) repository.PaymentProtectionAuditContext {
	return repository.PaymentProtectionAuditContext{
		UserID:    actor.UserID,
		Username:  actor.Username,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
		Method:    actor.Method,
		Path:      actor.Path,
	}
}
