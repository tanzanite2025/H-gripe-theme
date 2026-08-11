package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/pkg/antifraud"
	"commerce-platform/internal/pkg/config"
	pgateway "commerce-platform/internal/pkg/payment"
)

const (
	PaymentThreeDSModeAutomatic = "automatic"
	PaymentThreeDSModeAny       = "any"
	PaymentThreeDSModeChallenge = "challenge"
)

type paymentThreeDSVisitorAssessor interface {
	AssessIdentity(ctx context.Context, input VisitorRiskIdentityAssessmentInput) (VisitorRiskIdentityAssessment, error)
}

type paymentThreeDSRiskEvaluator interface {
	Evaluate(ctx context.Context, key string, signals antifraud.Signals) (antifraud.Decision, error)
}

type paymentThreeDSPortfolioRiskProvider interface {
	CurrentCheckoutPolicy(provider string) (PaymentRiskCheckoutPolicy, error)
}

type paymentThreeDSPaymentProtectionProvider interface {
	Evaluate(input paymentdomain.PaymentProtectionEvaluationInput) (paymentdomain.PaymentProtectionDecision, error)
}

type paymentThreeDSOrderHistory interface {
	CountPaidOrdersForUserBefore(userID uint, excludeOrderID uint) (int64, error)
}

type PaymentThreeDSPolicyService struct {
	cfg           config.PaymentThreeDSConfig
	visitorRisk   paymentThreeDSVisitorAssessor
	paymentRisk   paymentThreeDSRiskEvaluator
	portfolioRisk paymentThreeDSPortfolioRiskProvider
	protection    paymentThreeDSPaymentProtectionProvider
	orderHistory  paymentThreeDSOrderHistory
}

type PaymentThreeDSDecisionInput struct {
	Provider        string
	UserID          uint
	OrderID         uint
	Amount          float64
	Currency        string
	BaseMode        string
	IPAddress       string
	IPCountry       string
	UserAgent       string
	SessionID       string
	BillingCountry  string
	ShippingCountry string
	PaymentMethod   string
}

type PaymentThreeDSDecision struct {
	Mode               string
	Strategy           string
	ExemptionCandidate bool
	RiskLevel          string
	RiskScore          int
	PortfolioRiskLevel string
	Reasons            []string
}

func NewPaymentThreeDSPolicyService(
	orderHistory paymentThreeDSOrderHistory,
	visitorRisk paymentThreeDSVisitorAssessor,
	paymentRisk paymentThreeDSRiskEvaluator,
	cfg config.PaymentThreeDSConfig,
) *PaymentThreeDSPolicyService {
	return &PaymentThreeDSPolicyService{
		cfg:          normalizePaymentThreeDSConfig(cfg),
		visitorRisk:  visitorRisk,
		paymentRisk:  paymentRisk,
		orderHistory: orderHistory,
	}
}

func (s *PaymentThreeDSPolicyService) ConfigureRiskMonitoring(provider paymentThreeDSPortfolioRiskProvider) {
	if s == nil {
		return
	}
	s.portfolioRisk = provider
}

func (s *PaymentThreeDSPolicyService) ConfigurePaymentProtection(provider paymentThreeDSPaymentProtectionProvider) {
	if s == nil {
		return
	}
	s.protection = provider
}

func (s *PaymentThreeDSPolicyService) Decide(ctx context.Context, input PaymentThreeDSDecisionInput) PaymentThreeDSDecision {
	baseMode := normalizePaymentThreeDSMode(input.BaseMode)
	decision := PaymentThreeDSDecision{
		Mode:               baseMode,
		Strategy:           "configured",
		RiskLevel:          visitor.RiskLevelNormal,
		PortfolioRiskLevel: "normal",
		Reasons:            []string{},
	}
	manualProtectionApplied := false
	if s != nil {
		manualProtection, protectionErr := s.evaluateManualProtection(input)
		if protectionErr != nil {
			decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeAny)
			decision.Strategy = "manual_protection_unavailable"
			decision.Reasons = append(decision.Reasons, "manual_protection_unavailable")
			manualProtectionApplied = true
		} else if manualProtection.Force3DS {
			decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeAny)
			decision.Strategy = strongerPaymentThreeDSStrategy(decision.Strategy, "manual_protection")
			decision.Reasons = append(decision.Reasons, manualProtection.Reasons...)
			manualProtectionApplied = true
		}
	}
	if s == nil || !s.cfg.AdaptiveEnabled {
		decision.Reasons = append(decision.Reasons, "adaptive_3ds_disabled")
		return decision
	}

	if !manualProtectionApplied {
		decision.Strategy = "adaptive_default"
	}
	portfolioRiskPolicy, portfolioRiskErr := s.evaluatePortfolioRisk(input)
	if portfolioRiskErr != nil {
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeAny)
		decision.Strategy = "adaptive_step_up"
		decision.Reasons = append(decision.Reasons, "portfolio_risk_unavailable")
	} else {
		decision.PortfolioRiskLevel = string(portfolioRiskPolicy.Level)
		if portfolioRiskPolicy.Force3DS {
			decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeAny)
			decision.Strategy = strongerPaymentThreeDSStrategy(decision.Strategy, "adaptive_step_up")
			decision.Reasons = append(decision.Reasons, "portfolio_risk_step_up")
		}
	}

	paymentRiskDecision, paymentRiskErr := s.evaluatePaymentRisk(ctx, input)
	if paymentRiskErr != nil {
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeAny)
		decision.Strategy = "adaptive_step_up"
		decision.Reasons = append(decision.Reasons, "payment_risk_unavailable")
	} else {
		decision.RiskScore = paymentThreeDSMaxInt(decision.RiskScore, paymentRiskDecision.Score)
		s.applyPaymentRiskDecision(&decision, paymentRiskDecision)
	}

	visitorAssessment, visitorRiskErr := s.evaluateVisitorRisk(ctx, input)
	if visitorRiskErr != nil {
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeAny)
		decision.Strategy = "adaptive_step_up"
		decision.Reasons = append(decision.Reasons, "visitor_risk_unavailable")
	} else {
		decision.RiskScore = paymentThreeDSMaxInt(decision.RiskScore, visitorAssessment.RiskScore)
		if visitorAssessment.RiskLevel != "" {
			decision.RiskLevel = strongerVisitorRiskLevel(decision.RiskLevel, visitorAssessment.RiskLevel)
		}
		s.applyVisitorRiskAssessment(&decision, visitorAssessment)
	}

	paidOrders, err := s.paidOrderCount(input)
	if err != nil {
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeAny)
		decision.Strategy = "adaptive_step_up"
		decision.Reasons = append(decision.Reasons, "customer_history_unavailable")
	} else if decision.Mode == PaymentThreeDSModeAutomatic && s.lowRiskExemptionCandidate(input, paidOrders, paymentRiskDecision, visitorAssessment) {
		decision.Strategy = "adaptive_exemption_candidate"
		decision.ExemptionCandidate = true
		decision.Reasons = append(decision.Reasons, paymentThreeDSLowRiskReason(input.Amount, s.cfg.LowRiskMaxAmount, paidOrders, s.cfg.TrustedPaidOrders))
	}

	if len(decision.Reasons) == 0 {
		decision.Reasons = append(decision.Reasons, "stripe_automatic_sca_handling")
	}
	return decision
}

func (s *PaymentThreeDSPolicyService) evaluateManualProtection(
	input PaymentThreeDSDecisionInput,
) (paymentdomain.PaymentProtectionDecision, error) {
	if s == nil || s.protection == nil {
		return paymentdomain.PaymentProtectionDecision{}, nil
	}
	country := input.BillingCountry
	if strings.TrimSpace(country) == "" {
		country = input.ShippingCountry
	}
	return s.protection.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider:      input.Provider,
		Country:       country,
		PaymentMethod: input.PaymentMethod,
	})
}

func (d PaymentThreeDSDecision) Metadata() map[string]string {
	reasons := strings.Join(d.Reasons, ",")
	if len(reasons) > 450 {
		reasons = reasons[:450]
	}
	portfolioRiskLevel := d.PortfolioRiskLevel
	if portfolioRiskLevel == "" {
		portfolioRiskLevel = "normal"
	}
	return map[string]string{
		"three_ds_mode":                 d.Mode,
		"three_ds_policy":               d.Strategy,
		"three_ds_exemption_candidate":  strconv.FormatBool(d.ExemptionCandidate),
		"three_ds_risk_level":           d.RiskLevel,
		"three_ds_risk_score":           strconv.Itoa(d.RiskScore),
		"three_ds_portfolio_risk_level": portfolioRiskLevel,
		"three_ds_reasons":              reasons,
	}
}

func (s *PaymentThreeDSPolicyService) evaluatePortfolioRisk(input PaymentThreeDSDecisionInput) (PaymentRiskCheckoutPolicy, error) {
	if s == nil || s.portfolioRisk == nil {
		return PaymentRiskCheckoutPolicy{
			Level: "normal",
		}, nil
	}
	provider := strings.ToLower(strings.TrimSpace(input.Provider))
	if provider == "" {
		return PaymentRiskCheckoutPolicy{}, fmt.Errorf("payment provider is required for portfolio risk evaluation")
	}
	return s.portfolioRisk.CurrentCheckoutPolicy(provider)
}

func (s *PaymentThreeDSPolicyService) evaluatePaymentRisk(ctx context.Context, input PaymentThreeDSDecisionInput) (antifraud.Decision, error) {
	if s == nil || s.paymentRisk == nil {
		return antifraud.Decision{}, nil
	}
	return s.paymentRisk.Evaluate(ctx, paymentThreeDSRiskKey(input.UserID, input.SessionID), antifraud.Signals{
		IPCountry:      input.IPCountry,
		BillingCountry: paymentThreeDSBillingCountry(input),
		UserAgent:      input.UserAgent,
	})
}

func (s *PaymentThreeDSPolicyService) evaluateVisitorRisk(ctx context.Context, input PaymentThreeDSDecisionInput) (VisitorRiskIdentityAssessment, error) {
	if s == nil || s.visitorRisk == nil {
		return VisitorRiskIdentityAssessment{RiskLevel: visitor.RiskLevelNormal}, nil
	}
	return s.visitorRisk.AssessIdentity(ctx, VisitorRiskIdentityAssessmentInput{
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
		Now:          time.Now().UTC(),
		LookbackDays: s.cfg.VisitorRiskLookback,
	})
}

func (s *PaymentThreeDSPolicyService) applyPaymentRiskDecision(decision *PaymentThreeDSDecision, riskDecision antifraud.Decision) {
	if riskDecision.HighRisk || riskDecision.Score >= s.cfg.ChallengeRiskScore {
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeChallenge)
		decision.Strategy = "adaptive_challenge"
		decision.Reasons = append(decision.Reasons, "payment_risk_high")
		return
	}
	if riskDecision.Score >= s.cfg.StepUpRiskScore || riskDecision.Failures > 0 {
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeAny)
		decision.Strategy = strongerPaymentThreeDSStrategy(decision.Strategy, "adaptive_step_up")
		decision.Reasons = append(decision.Reasons, "payment_risk_step_up")
	}
}

func (s *PaymentThreeDSPolicyService) applyVisitorRiskAssessment(decision *PaymentThreeDSDecision, assessment VisitorRiskIdentityAssessment) {
	switch assessment.DecisionAction {
	case visitor.RiskDecisionActionIgnore:
		decision.Reasons = append(decision.Reasons, "visitor_risk_ignored")
	case visitor.RiskDecisionActionTemporaryBlock, visitor.RiskDecisionActionBlockCandidate:
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeChallenge)
		decision.Strategy = "adaptive_challenge"
		decision.Reasons = append(decision.Reasons, "visitor_risk_manual_challenge")
		return
	case visitor.RiskDecisionActionWatch:
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeAny)
		decision.Strategy = strongerPaymentThreeDSStrategy(decision.Strategy, "adaptive_step_up")
		decision.Reasons = append(decision.Reasons, "visitor_risk_manual_watch")
	}

	switch assessment.RiskLevel {
	case visitor.RiskLevelBlock, visitor.RiskLevelSuspicious:
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeChallenge)
		decision.Strategy = "adaptive_challenge"
		decision.Reasons = append(decision.Reasons, "visitor_risk_high")
	case visitor.RiskLevelWatch:
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeAny)
		decision.Strategy = strongerPaymentThreeDSStrategy(decision.Strategy, "adaptive_step_up")
		decision.Reasons = append(decision.Reasons, "visitor_risk_watch")
	}

	if assessment.RiskScore >= s.cfg.ChallengeRiskScore || assessment.CheckoutFailureCount >= 2 {
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeChallenge)
		decision.Strategy = "adaptive_challenge"
		decision.Reasons = append(decision.Reasons, "visitor_risk_score_challenge")
		return
	}
	if assessment.RiskScore >= s.cfg.StepUpRiskScore || assessment.CheckoutFailureCount > 0 || assessment.BotLikeUserAgentCount > 0 {
		decision.Mode = strongerThreeDSMode(decision.Mode, PaymentThreeDSModeAny)
		decision.Strategy = strongerPaymentThreeDSStrategy(decision.Strategy, "adaptive_step_up")
		decision.Reasons = append(decision.Reasons, "visitor_risk_score_step_up")
	}
}

func (s *PaymentThreeDSPolicyService) paidOrderCount(input PaymentThreeDSDecisionInput) (int64, error) {
	if s == nil || s.orderHistory == nil || input.UserID == 0 {
		return 0, nil
	}
	return s.orderHistory.CountPaidOrdersForUserBefore(input.UserID, input.OrderID)
}

func (s *PaymentThreeDSPolicyService) lowRiskExemptionCandidate(
	input PaymentThreeDSDecisionInput,
	paidOrders int64,
	paymentRiskDecision antifraud.Decision,
	visitorAssessment VisitorRiskIdentityAssessment,
) bool {
	if input.Amount < 0 || paymentRiskDecision.HighRisk {
		return false
	}
	if paymentRiskDecision.Score >= s.cfg.StepUpRiskScore || paymentRiskDecision.Failures > 0 {
		return false
	}
	if visitorAssessment.DecisionAction == visitor.RiskDecisionActionWatch ||
		visitorAssessment.DecisionAction == visitor.RiskDecisionActionTemporaryBlock ||
		visitorAssessment.DecisionAction == visitor.RiskDecisionActionBlockCandidate {
		return false
	}
	if visitorAssessment.RiskLevel == visitor.RiskLevelWatch ||
		visitorAssessment.RiskLevel == visitor.RiskLevelSuspicious ||
		visitorAssessment.RiskLevel == visitor.RiskLevelBlock ||
		visitorAssessment.RiskScore >= s.cfg.StepUpRiskScore ||
		visitorAssessment.CheckoutFailureCount > 0 ||
		visitorAssessment.BotLikeUserAgentCount > 0 {
		return false
	}

	smallAmount := s.cfg.LowRiskMaxAmount > 0 && input.Amount <= s.cfg.LowRiskMaxAmount
	trustedCustomer := paidOrders >= int64(s.cfg.TrustedPaidOrders)
	return smallAmount || trustedCustomer
}

func normalizePaymentThreeDSConfig(cfg config.PaymentThreeDSConfig) config.PaymentThreeDSConfig {
	if cfg.LowRiskMaxAmount < 0 {
		cfg.LowRiskMaxAmount = 0
	}
	if cfg.TrustedPaidOrders <= 0 {
		cfg.TrustedPaidOrders = 1
	}
	if cfg.VisitorRiskLookback <= 0 {
		cfg.VisitorRiskLookback = 30
	}
	if cfg.StepUpRiskScore <= 0 {
		cfg.StepUpRiskScore = 20
	}
	if cfg.ChallengeRiskScore <= 0 {
		cfg.ChallengeRiskScore = 60
	}
	if cfg.StepUpRiskScore > cfg.ChallengeRiskScore {
		cfg.StepUpRiskScore = cfg.ChallengeRiskScore
	}
	return cfg
}

func normalizePaymentThreeDSMode(value string) string {
	return pgateway.NormalizeThreeDSecureMode(value)
}

func strongerThreeDSMode(current, next string) string {
	current = normalizePaymentThreeDSMode(current)
	next = normalizePaymentThreeDSMode(next)
	if paymentThreeDSModePriority(next) > paymentThreeDSModePriority(current) {
		return next
	}
	return current
}

func paymentThreeDSModePriority(mode string) int {
	switch normalizePaymentThreeDSMode(mode) {
	case PaymentThreeDSModeChallenge:
		return 2
	case PaymentThreeDSModeAny:
		return 1
	default:
		return 0
	}
}

func strongerPaymentThreeDSStrategy(current, next string) string {
	if current == "adaptive_challenge" {
		return current
	}
	return next
}

func strongerVisitorRiskLevel(current, next string) string {
	if visitorRiskLevelPriority(next) > visitorRiskLevelPriority(current) {
		return next
	}
	return current
}

func visitorRiskLevelPriority(level string) int {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case visitor.RiskLevelBlock:
		return 3
	case visitor.RiskLevelSuspicious:
		return 2
	case visitor.RiskLevelWatch:
		return 1
	default:
		return 0
	}
}

func paymentThreeDSRiskKey(userID uint, sessionID string) string {
	sessionID = strings.TrimSpace(sessionID)
	if userID == 0 {
		if sessionID != "" {
			return "anonymous:session:" + sessionID
		}
		return "anonymous"
	}
	if sessionID != "" {
		return fmt.Sprintf("user:%d:session:%s", userID, sessionID)
	}
	return fmt.Sprintf("user:%d", userID)
}

func paymentThreeDSBillingCountry(input PaymentThreeDSDecisionInput) string {
	if strings.TrimSpace(input.BillingCountry) != "" {
		return input.BillingCountry
	}
	return input.ShippingCountry
}

func paymentThreeDSLowRiskReason(amount, lowRiskMax float64, paidOrders int64, trustedPaidOrders int) string {
	reasons := []string{}
	if lowRiskMax > 0 && amount <= lowRiskMax {
		reasons = append(reasons, "low_amount")
	}
	if paidOrders >= int64(trustedPaidOrders) {
		reasons = append(reasons, "trusted_customer")
	}
	if len(reasons) == 0 {
		return "low_risk"
	}
	return strings.Join(reasons, "+")
}

func paymentThreeDSMaxInt(a, b int) int {
	if b > a {
		return b
	}
	return a
}
