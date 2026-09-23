package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/user"
	referralcookie "commerce-platform/internal/pkg/referral"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

var (
	ErrReferralProgramDisabled      = errors.New("referral program is disabled")
	ErrReferralCodeNotFound         = errors.New("referral code not found")
	ErrSelfReferralForbidden        = errors.New("self referral is forbidden")
	ErrRefereeNotEligible           = errors.New("customer is not eligible for a referral")
	ErrRefereeAlreadyAttributed     = errors.New("customer is already attributed to another referrer")
	ErrReferralServiceUnavailable   = errors.New("referral service is unavailable")
	ErrInvalidReferralProgramConfig = errors.New("invalid referral program config")
	ErrReferralActionInvalid        = errors.New("referral action is invalid for the current state")
	ErrReferralReasonRequired       = errors.New("referral action reason is required")
)

type ReferralValidation struct {
	Valid                    bool   `json:"valid"`
	ReferrerNameMask         string `json:"referrer_name_mask"`
	RefereeBenefitType       string `json:"referee_benefit_type"`
	RefereeBenefitValue      int64  `json:"referee_benefit_value"`
	BenefitIssuance          string `json:"benefit_issuance"`
	MinOrderAmountMinor      int64  `json:"min_order_amount_minor"`
	VestingPeriodDays        int    `json:"vesting_period_days"`
	AttributionCookieTTLDays int    `json:"attribution_cookie_ttl_days"`
}

type ReferralDashboard struct {
	Enabled      bool                     `json:"enabled"`
	ReferralCode string                   `json:"referral_code,omitempty"`
	CustomSlug   *string                  `json:"custom_slug,omitempty"`
	ShareURL     string                   `json:"share_url,omitempty"`
	RewardRules  ReferralDashboardRules   `json:"reward_rules"`
	Stats        repository.ReferralStats `json:"stats"`
}

type ReferralDashboardRules struct {
	MinOrderAmountMinor      int64  `json:"min_order_amount_minor"`
	ReferrerRewardPoints     int    `json:"referrer_reward_points"`
	RefereeBenefitType       string `json:"referee_benefit_type"`
	RefereeBenefitValue      int64  `json:"referee_benefit_value"`
	VestingPeriodDays        int    `json:"vesting_period_days"`
	AttributionCookieTTLDays int    `json:"attribution_cookie_ttl_days"`
}

type ReferralHistoryItem struct {
	ID                    uint       `json:"id"`
	RefereeNameMask       string     `json:"referee_name_mask"`
	Status                string     `json:"status"`
	OrderDate             *time.Time `json:"order_date,omitempty"`
	PotentialRewardPoints int        `json:"potential_reward_points"`
	VestingUntil          *time.Time `json:"vesting_until,omitempty"`
}

type ReferralHistory struct {
	Items    []ReferralHistoryItem `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

type ReferralAdminItem struct {
	ID            uint                `json:"id"`
	ReferralCode  string              `json:"referral_code"`
	Referrer      ReferralAdminPerson `json:"referrer"`
	Referee       ReferralAdminPerson `json:"referee"`
	Order         *ReferralAdminOrder `json:"order,omitempty"`
	RewardPoints  int                 `json:"reward_points"`
	Status        string              `json:"status"`
	VestingUntil  *time.Time          `json:"vesting_until,omitempty"`
	DaysRemaining int                 `json:"days_remaining"`
	RiskFlags     []map[string]any    `json:"risk_flags"`
	CreatedAt     time.Time           `json:"created_at"`
}

type ReferralAdminPerson struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	EmailMasked string `json:"email_masked,omitempty"`
}

type ReferralAdminOrder struct {
	ID          uint       `json:"id"`
	OrderNumber string     `json:"order_number"`
	AmountMinor int64      `json:"amount_minor"`
	Currency    string     `json:"currency"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
}

type ReferralAdminLedger struct {
	Items    []ReferralAdminItem           `json:"items"`
	Overview repository.ReferralAdminStats `json:"overview"`
	Total    int64                         `json:"total"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"page_size"`
}

// ReferralProgramConfigInput is the only mutable surface of the v2 referral
// policy. A new input always becomes a new immutable version.
type ReferralProgramConfigInput struct {
	Enabled                 bool
	MinOrderAmountMinor     int64
	ReferrerRewardPoints    int
	RefereeBenefitType      string
	RefereeBenefitValue     int64
	VestingPeriodDays       int
	UndeliveredFallbackDays int
	AttributionTTLDays      int
	MonthlyCapPerReferrer   int
	AntiFraudMode           string
	CreatedBy               *uint
}

type ReferralAdminDetail struct {
	Item        ReferralAdminItem            `json:"item"`
	Transitions []loyalty.ReferralTransition `json:"transitions"`
	Rewards     []loyalty.ReferralReward     `json:"rewards"`
}

type ReferralActionResult struct {
	Record *loyalty.ReferralRecord `json:"record"`
	Reward *loyalty.ReferralReward `json:"reward,omitempty"`
}

type ReferralBindContext struct {
	ClientIP          string
	DeviceFingerprint string
	RefereeEmail      string
}

type ReferralService struct {
	txManager     *repository.TxManager
	repo          *repository.ReferralRepository
	programRepo   *repository.ReferralProgramRepository
	userRepo      *repository.UserRepository
	cookieSigner  *referralcookie.Signer
	hashKey       []byte
	invitations   *ReferralInvitationService
	storefrontURL string
	now           func() time.Time
}

func NewReferralService(
	txManager *repository.TxManager,
	repo *repository.ReferralRepository,
	programRepo *repository.ReferralProgramRepository,
	userRepo *repository.UserRepository,
	baseURL string,
	secret string,
	storefrontURL ...string,
) *ReferralService {
	signer, _ := referralcookie.NewSigner(secret)
	resolvedStorefrontURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if len(storefrontURL) > 0 && strings.TrimSpace(storefrontURL[0]) != "" {
		resolvedStorefrontURL = strings.TrimRight(strings.TrimSpace(storefrontURL[0]), "/")
	}
	return &ReferralService{
		txManager:     txManager,
		repo:          repo,
		programRepo:   programRepo,
		userRepo:      userRepo,
		cookieSigner:  signer,
		hashKey:       []byte(strings.TrimSpace(secret)),
		invitations:   NewReferralInvitationService(repo, nil, StorefrontReferralInvitationLinks{StorefrontURL: resolvedStorefrontURL}),
		storefrontURL: resolvedStorefrontURL,
		now:           func() time.Time { return time.Now().UTC() },
	}
}

func (s *ReferralService) StorefrontURL() string {
	if s == nil {
		return ""
	}
	return s.storefrontURL
}

// ConfigureReferralInvitationService is a startup composition hook for code or
// short-link providers. Existing identities remain persisted across upgrades.
func (s *ReferralService) ConfigureReferralInvitationService(invitations *ReferralInvitationService) {
	s.invitations = invitations
}

func (s *ReferralService) ValidateCode(code string) (*ReferralValidation, error) {
	config, identity, err := s.activeIdentity(code)
	if err != nil {
		return nil, err
	}
	referrer, err := s.userRepo.FindByID(identity.UserID)
	if err != nil {
		return nil, err
	}
	return &ReferralValidation{
		Valid:                    true,
		ReferrerNameMask:         maskReferralName(referrer),
		RefereeBenefitType:       config.RefereeBenefitType,
		RefereeBenefitValue:      config.RefereeBenefitValue,
		BenefitIssuance:          referralBenefitIssuance(config.RefereeBenefitType),
		MinOrderAmountMinor:      config.MinOrderAmountMinor,
		VestingPeriodDays:        config.VestingPeriodDays,
		AttributionCookieTTLDays: config.AttributionTTLDays,
	}, nil
}

func referralBenefitIssuance(benefitType string) string {
	if strings.EqualFold(strings.TrimSpace(benefitType), loyalty.ReferralBenefitPoints) {
		return "points_on_registration"
	}
	return "none"
}

func (s *ReferralService) CreateAttributionToken(code, source string) (string, int, error) {
	config, _, err := s.activeIdentity(code)
	if err != nil {
		return "", 0, err
	}
	if s.cookieSigner == nil {
		return "", 0, ErrReferralServiceUnavailable
	}
	ttl := time.Duration(config.AttributionTTLDays) * 24 * time.Hour
	token, _, err := s.cookieSigner.Encode(code, source, ttl)
	if err != nil {
		return "", 0, err
	}
	return token, int(ttl / time.Second), nil
}

func (s *ReferralService) BindFromToken(userID uint, token, clientIP string) (*loyalty.ReferralRecord, error) {
	return s.BindFromTokenWithContext(userID, token, ReferralBindContext{ClientIP: clientIP})
}

func (s *ReferralService) BindFromTokenWithContext(userID uint, token string, bindContext ReferralBindContext) (*loyalty.ReferralRecord, error) {
	if s == nil || s.cookieSigner == nil {
		return nil, ErrReferralServiceUnavailable
	}
	claims, err := s.cookieSigner.Decode(token)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(bindContext.RefereeEmail) == "" && s.userRepo != nil {
		if account, findErr := s.userRepo.FindByID(userID); findErr == nil && account != nil {
			bindContext.RefereeEmail = account.Email
		}
	}
	return s.bind(userID, claims, bindContext)
}

func (s *ReferralService) Dashboard(userID uint) (*ReferralDashboard, error) {
	if s == nil || s.programRepo == nil || s.repo == nil || s.invitations == nil || userID == 0 {
		return nil, ErrReferralServiceUnavailable
	}
	config, err := s.programRepo.FindActive()
	if err != nil {
		return nil, err
	}
	if err := validateReferralProgramConfig(config); err != nil {
		return nil, err
	}
	dashboard := &ReferralDashboard{
		Enabled:     config.Enabled,
		RewardRules: referralDashboardRules(config),
	}
	if !config.Enabled {
		return dashboard, nil
	}
	identity, shareURL, err := s.invitations.GetOrCreateReferralInvitation(userID)
	if err != nil {
		return nil, err
	}
	stats, err := s.repo.StatsByReferrerID(userID)
	if err != nil {
		return nil, err
	}
	dashboard.ReferralCode = identity.ReferralCode
	dashboard.CustomSlug = identity.CustomSlug
	dashboard.ShareURL = shareURL
	dashboard.Stats = stats
	return dashboard, nil
}

func (s *ReferralService) History(userID uint, page, pageSize int) (*ReferralHistory, error) {
	if s == nil || s.repo == nil || s.programRepo == nil || s.userRepo == nil || userID == 0 {
		return nil, ErrReferralServiceUnavailable
	}
	records, total, err := s.repo.ListRecordsByReferrerID(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	refereeIDs := make([]uint, 0, len(records))
	configIDs := make([]uint, 0, len(records))
	for _, record := range records {
		if record.RefereeID != nil {
			refereeIDs = append(refereeIDs, *record.RefereeID)
		}
		configIDs = append(configIDs, record.ProgramConfigID)
	}
	referees, err := s.userRepo.FindByIDs(refereeIDs)
	if err != nil {
		return nil, err
	}
	configs, err := s.programRepo.FindByIDs(configIDs)
	if err != nil {
		return nil, err
	}
	refereesByID := make(map[uint]*user.User, len(referees))
	for index := range referees {
		refereesByID[referees[index].ID] = &referees[index]
	}
	configsByID := make(map[uint]*loyalty.ReferralProgramConfig, len(configs))
	for index := range configs {
		configsByID[configs[index].ID] = &configs[index]
	}

	result := &ReferralHistory{Items: make([]ReferralHistoryItem, 0, len(records)), Total: total, Page: page, PageSize: pageSize}
	for _, record := range records {
		item := ReferralHistoryItem{
			ID:           record.ID,
			Status:       record.Status,
			OrderDate:    record.OrderedAt,
			VestingUntil: record.VestingUntil,
		}
		if record.RefereeID != nil {
			referee, exists := refereesByID[*record.RefereeID]
			if !exists {
				return nil, fmt.Errorf("referral history referee %d is missing", *record.RefereeID)
			}
			item.RefereeNameMask = maskReferralName(referee)
		}
		config, exists := configsByID[record.ProgramConfigID]
		if !exists {
			return nil, fmt.Errorf("referral history config %d is missing", record.ProgramConfigID)
		}
		item.PotentialRewardPoints = config.ReferrerRewardPoints
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func (s *ReferralService) AdminLedger(filters repository.ReferralAdminFilters, page, pageSize int, now time.Time) (*ReferralAdminLedger, error) {
	if s == nil || s.repo == nil || s.userRepo == nil || s.txManager == nil {
		return nil, ErrReferralServiceUnavailable
	}
	records, total, err := s.repo.ListAdminRecords(filters, page, pageSize)
	if err != nil {
		return nil, err
	}
	overview, err := s.repo.AdminStats(filters)
	if err != nil {
		return nil, err
	}
	userIDs := make([]uint, 0, len(records)*2)
	orderIDs := make([]uint, 0, len(records))
	configIDs := make([]uint, 0, len(records))
	for _, record := range records {
		userIDs = append(userIDs, record.ReferrerID)
		if record.RefereeID != nil {
			userIDs = append(userIDs, *record.RefereeID)
		}
		if record.OrderID != nil {
			orderIDs = append(orderIDs, *record.OrderID)
		}
		configIDs = append(configIDs, record.ProgramConfigID)
	}
	users, err := s.userRepo.FindByIDs(userIDs)
	if err != nil {
		return nil, err
	}
	orders, err := s.txManager.OrderRepository().FindByIDsBasic(orderIDs)
	if err != nil {
		return nil, err
	}
	configs, err := s.programRepo.FindByIDs(configIDs)
	if err != nil {
		return nil, err
	}
	usersByID := make(map[uint]*user.User, len(users))
	for index := range users {
		usersByID[users[index].ID] = &users[index]
	}
	ordersByID := make(map[uint]*order.Order, len(orders))
	for index := range orders {
		ordersByID[orders[index].ID] = &orders[index]
	}
	configsByID := make(map[uint]*loyalty.ReferralProgramConfig, len(configs))
	for index := range configs {
		configsByID[configs[index].ID] = &configs[index]
	}
	if now.IsZero() {
		now = s.now().UTC()
	} else {
		now = now.UTC()
	}
	result := &ReferralAdminLedger{Items: make([]ReferralAdminItem, 0, len(records)), Overview: overview, Total: total, Page: page, PageSize: pageSize}
	for _, record := range records {
		item := ReferralAdminItem{ID: record.ID, ReferralCode: record.ReferralCodeSnapshot, Status: record.Status, VestingUntil: record.VestingUntil, CreatedAt: record.CreatedAt, RiskFlags: []map[string]any{}}
		if record.VestingUntil != nil && record.VestingUntil.After(now) {
			item.DaysRemaining = int(record.VestingUntil.Sub(now).Hours() / 24)
			if item.DaysRemaining < 1 {
				item.DaysRemaining = 1
			}
		}
		if referrer := usersByID[record.ReferrerID]; referrer != nil {
			item.Referrer = ReferralAdminPerson{ID: referrer.ID, Name: maskReferralName(referrer), EmailMasked: maskReferralEmail(referrer.Email)}
		}
		if record.RefereeID != nil {
			if referee := usersByID[*record.RefereeID]; referee != nil {
				item.Referee = ReferralAdminPerson{ID: referee.ID, Name: maskReferralName(referee), EmailMasked: maskReferralEmail(referee.Email)}
			}
		}
		if record.OrderID != nil {
			if orderRecord := ordersByID[*record.OrderID]; orderRecord != nil {
				item.Order = &ReferralAdminOrder{ID: orderRecord.ID, OrderNumber: orderRecord.OrderNumber, AmountMinor: record.OrderAmountMinor, Currency: record.Currency, PaidAt: orderRecord.PaidAt, DeliveredAt: orderRecord.DeliveredAt}
			}
		}
		if config := configsByID[record.ProgramConfigID]; config != nil {
			item.RewardPoints = config.ReferrerRewardPoints
		}
		if len(record.RiskFlags) > 0 {
			_ = json.Unmarshal(record.RiskFlags, &item.RiskFlags)
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func (s *ReferralService) AdminDetail(id uint, now time.Time) (*ReferralAdminDetail, error) {
	if s == nil || s.repo == nil || s.userRepo == nil || s.txManager == nil || s.programRepo == nil || id == 0 {
		return nil, ErrReferralServiceUnavailable
	}
	record, err := s.repo.FindRecordByID(id)
	if err != nil {
		return nil, err
	}
	users, err := s.userRepo.FindByIDs(append([]uint{record.ReferrerID}, valueUint(record.RefereeID)...))
	if err != nil {
		return nil, err
	}
	usersByID := make(map[uint]*user.User, len(users))
	for i := range users {
		usersByID[users[i].ID] = &users[i]
	}
	var orders []order.Order
	if record.OrderID != nil {
		orders, err = s.txManager.OrderRepository().FindByIDsBasic([]uint{*record.OrderID})
		if err != nil {
			return nil, err
		}
	}
	var orderRecord *order.Order
	if len(orders) > 0 {
		orderRecord = &orders[0]
	}
	config, err := s.programRepo.FindByID(record.ProgramConfigID)
	if err != nil {
		return nil, err
	}
	item := s.adminItem(record, usersByID, orderRecord, config, now)
	transitions, err := s.repo.ListTransitionsByRecordID(id)
	if err != nil {
		return nil, err
	}
	rewards, err := s.repo.ListRewardsByRecordID(id)
	if err != nil {
		return nil, err
	}
	return &ReferralAdminDetail{Item: item, Transitions: transitions, Rewards: rewards}, nil
}

func valueUint(value *uint) []uint {
	if value == nil {
		return nil
	}
	return []uint{*value}
}

func (s *ReferralService) adminItem(record *loyalty.ReferralRecord, usersByID map[uint]*user.User, orderRecord *order.Order, config *loyalty.ReferralProgramConfig, now time.Time) ReferralAdminItem {
	if now.IsZero() {
		now = s.now().UTC()
	} else {
		now = now.UTC()
	}
	item := ReferralAdminItem{ID: record.ID, ReferralCode: record.ReferralCodeSnapshot, Status: record.Status, VestingUntil: record.VestingUntil, CreatedAt: record.CreatedAt, RiskFlags: []map[string]any{}}
	if record.VestingUntil != nil && record.VestingUntil.After(now) {
		item.DaysRemaining = int(record.VestingUntil.Sub(now).Hours() / 24)
		if item.DaysRemaining < 1 {
			item.DaysRemaining = 1
		}
	}
	if referrer := usersByID[record.ReferrerID]; referrer != nil {
		item.Referrer = ReferralAdminPerson{ID: referrer.ID, Name: maskReferralName(referrer), EmailMasked: maskReferralEmail(referrer.Email)}
	}
	if record.RefereeID != nil {
		if referee := usersByID[*record.RefereeID]; referee != nil {
			item.Referee = ReferralAdminPerson{ID: referee.ID, Name: maskReferralName(referee), EmailMasked: maskReferralEmail(referee.Email)}
		}
	}
	if orderRecord != nil {
		item.Order = &ReferralAdminOrder{ID: orderRecord.ID, OrderNumber: orderRecord.OrderNumber, AmountMinor: record.OrderAmountMinor, Currency: record.Currency, PaidAt: orderRecord.PaidAt, DeliveredAt: orderRecord.DeliveredAt}
	}
	if config != nil {
		item.RewardPoints = config.ReferrerRewardPoints
	}
	if len(record.RiskFlags) > 0 {
		_ = json.Unmarshal(record.RiskFlags, &item.RiskFlags)
	}
	return item
}

func (s *ReferralService) GetAdminProgramConfig() (*loyalty.ReferralProgramConfig, error) {
	if s == nil || s.programRepo == nil {
		return nil, ErrReferralServiceUnavailable
	}
	config, err := s.programRepo.FindActive()
	if err != nil {
		return nil, err
	}
	if err := validateReferralProgramConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}

func (s *ReferralService) PublishAdminProgramConfig(input ReferralProgramConfigInput, expectedVersion int) (*loyalty.ReferralProgramConfig, error) {
	if s == nil || s.programRepo == nil || expectedVersion <= 0 {
		return nil, ErrInvalidReferralProgramConfig
	}
	// Referral benefits have exactly one type: points.
	if strings.ToLower(strings.TrimSpace(input.RefereeBenefitType)) != loyalty.ReferralBenefitPoints {
		return nil, fmt.Errorf("%w: referral benefits must use points", ErrInvalidReferralProgramConfig)
	}
	current, err := s.programRepo.FindActive()
	if err != nil {
		return nil, err
	}
	configuredCurrency := currency.NormalizeCode(current.Currency)
	if configuredCurrency == "" {
		return nil, ErrInvalidReferralProgramConfig
	}
	config := &loyalty.ReferralProgramConfig{
		Version: expectedVersion + 1,
		Enabled: input.Enabled, Currency: configuredCurrency,
		MinOrderAmountMinor: input.MinOrderAmountMinor, ReferrerRewardPoints: input.ReferrerRewardPoints,
		RefereeBenefitType: loyalty.ReferralBenefitPoints, RefereeBenefitValue: input.RefereeBenefitValue,
		VestingPeriodDays: input.VestingPeriodDays, UndeliveredFallbackDays: input.UndeliveredFallbackDays,
		AttributionTTLDays: input.AttributionTTLDays, MonthlyCapPerReferrer: input.MonthlyCapPerReferrer,
		AntiFraudMode: strings.ToLower(strings.TrimSpace(input.AntiFraudMode)), CreatedBy: input.CreatedBy,
	}
	if err := validateReferralProgramConfig(config); err != nil {
		return nil, err
	}
	if err := s.programRepo.CreateVersion(config, expectedVersion); err != nil {
		return nil, err
	}
	return config, nil
}

// releaseRefereeBenefitInTx credits the new account's registration points. The
// reward log is the idempotency boundary; this is ordinary unified loyalty
// balance, not a referral-specific wallet.
func (s *ReferralService) releaseRefereeBenefitInTx(
	repos repository.TxRepositories,
	record *loyalty.ReferralRecord,
	config *loyalty.ReferralProgramConfig,
	releasedAt time.Time,
) error {
	if record == nil || config == nil || record.RefereeID == nil || repos.Referral == nil {
		return ErrReferralServiceUnavailable
	}
	benefitType := strings.ToLower(strings.TrimSpace(config.RefereeBenefitType))
	if benefitType != loyalty.ReferralBenefitPoints || config.RefereeBenefitValue <= 0 {
		return ErrInvalidReferralProgramConfig
	}
	rewardKey := fmt.Sprintf("referral:%d:referee:%s:v1", record.ID, benefitType)
	existing, err := repos.Referral.FindRewardByIdempotencyKey(rewardKey)
	if err == nil {
		if existing.Status == loyalty.ReferralRewardStatusReleased {
			return nil
		}
		if existing.Status != loyalty.ReferralRewardStatusLocked {
			return ErrReferralActionInvalid
		}
	} else if !repository.IsRecordNotFound(err) {
		return err
	}

	snapshot, err := json.Marshal(map[string]any{
		"version":               config.Version,
		"referee_benefit_type":  config.RefereeBenefitType,
		"referee_benefit_value": config.RefereeBenefitValue,
	})
	if err != nil {
		return err
	}

	reward := existing
	if reward == nil {
		reward = &loyalty.ReferralReward{
			ReferralRecordID: record.ID,
			ProgramConfigID:  record.ProgramConfigID,
			RecipientUserID:  *record.RefereeID,
			RecipientRole:    loyalty.ReferralRecipientReferee,
			RewardType:       loyalty.ReferralRewardTypePoints,
			IdempotencyKey:   rewardKey,
			Status:           loyalty.ReferralRewardStatusLocked,
			RuleSnapshot:     snapshot,
		}
	}

	if benefitType == loyalty.ReferralBenefitPoints {
		if repos.Loyalty == nil {
			return ErrReferralServiceUnavailable
		}
		if config.RefereeBenefitValue > int64(^uint(0)>>1) {
			return ErrInvalidReferralProgramConfig
		}
		reward.PointsAmount = int(config.RefereeBenefitValue)
		if reward.ID == 0 {
			if err := repos.Referral.CreateReward(reward); err != nil {
				return err
			}
		}
		transaction, err := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
			*record.RefereeID,
			reward.PointsAmount,
			"earn",
			"referral_referee",
			record.ID,
			fmt.Sprintf("Referral welcome benefit for record #%d", record.ID),
			&record.ProgramConfigID,
		)
		if err != nil {
			return err
		}
		reward.LoyaltyTransactionID = &transaction.ID
	}

	if reward.ID == 0 {
		return ErrReferralActionInvalid
	}
	updates := map[string]any{
		"status":        loyalty.ReferralRewardStatusReleased,
		"released_at":   releasedAt,
		"rule_snapshot": snapshot,
		"updated_at":    releasedAt,
	}
	if reward.LoyaltyTransactionID != nil {
		updates["loyalty_transaction_id"] = *reward.LoyaltyTransactionID
	}
	if err := repos.Referral.UpdateReward(reward.ID, updates); err != nil {
		return err
	}
	return nil
}

// SettleReferral releases the referrer points and closes the lifecycle record
// in one transaction. The idempotency key is stable across retries, so a
// retried admin request cannot create a second points transaction.
func (s *ReferralService) SettleReferral(id uint, reason string, actorID *uint) (*ReferralActionResult, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, ErrReferralReasonRequired
	}
	if s == nil || s.txManager == nil {
		return nil, ErrReferralServiceUnavailable
	}
	var result ReferralActionResult
	trigger, actorType := "admin_settle", "admin"
	if actorID == nil {
		trigger, actorType = "lifecycle_settle", "system"
	}
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Referral == nil || repos.ReferralProgram == nil || repos.Loyalty == nil {
			return ErrReferralServiceUnavailable
		}
		record, err := repos.Referral.FindRecordByIDForUpdate(id)
		if err != nil {
			return err
		}
		result.Record = record
		if record.Status == loyalty.ReferralStatusSettled {
			return nil
		}
		if record.Status != loyalty.ReferralStatusVesting || record.ReferrerID == 0 {
			return ErrReferralActionInvalid
		}
		config, err := repos.ReferralProgram.FindByID(record.ProgramConfigID)
		if err != nil {
			return err
		}
		rewardKey := fmt.Sprintf("referral:%d:referrer:points:v1", record.ID)
		reward, rewardErr := repos.Referral.FindRewardByIdempotencyKey(rewardKey)
		if rewardErr != nil && !repository.IsRecordNotFound(rewardErr) {
			return rewardErr
		}
		if reward == nil {
			snapshot, marshalErr := json.Marshal(map[string]any{
				"version":                config.Version,
				"referrer_reward_points": config.ReferrerRewardPoints,
				"vesting_period_days":    config.VestingPeriodDays,
			})
			if marshalErr != nil {
				return marshalErr
			}
			reward = &loyalty.ReferralReward{
				ReferralRecordID: record.ID, ProgramConfigID: record.ProgramConfigID,
				RecipientUserID: record.ReferrerID, RecipientRole: loyalty.ReferralRecipientReferrer,
				RewardType: loyalty.ReferralRewardTypePoints, PointsAmount: config.ReferrerRewardPoints,
				IdempotencyKey: rewardKey, Status: loyalty.ReferralRewardStatusLocked, RuleSnapshot: snapshot,
			}
			if err := repos.Referral.CreateReward(reward); err != nil {
				return err
			}
		}
		if reward.Status == loyalty.ReferralRewardStatusReleased {
			result.Reward = reward
			settledAt := s.now().UTC()
			return transitionReferralRecordAs(repos, record, loyalty.ReferralStatusSettled, trigger, strings.TrimSpace(reason), fmt.Sprintf("referral.settle:%d:%d", record.ID, reward.ID), map[string]any{"settled_at": settledAt}, actorType, actorID)
		}
		if reward.Status != loyalty.ReferralRewardStatusLocked || reward.PointsAmount <= 0 {
			return ErrReferralActionInvalid
		}
		transaction, err := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
			record.ReferrerID, reward.PointsAmount, "earn", "referral", record.ID,
			fmt.Sprintf("Referral reward for record #%d", record.ID), &record.ProgramConfigID,
		)
		if err != nil {
			return err
		}
		releasedAt := s.now().UTC()
		if err := repos.Referral.UpdateReward(reward.ID, map[string]any{"status": loyalty.ReferralRewardStatusReleased, "released_at": releasedAt, "loyalty_transaction_id": transaction.ID, "updated_at": releasedAt}); err != nil {
			return err
		}
		reward.Status = loyalty.ReferralRewardStatusReleased
		reward.ReleasedAt = &releasedAt
		reward.LoyaltyTransactionID = &transaction.ID
		result.Reward = reward
		return transitionReferralRecordAs(repos, record, loyalty.ReferralStatusSettled, trigger, strings.TrimSpace(reason), fmt.Sprintf("referral.settle:%d:%d", record.ID, reward.ID), map[string]any{"settled_at": releasedAt}, actorType, actorID)
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ReferralService) RevokeReferral(id uint, reason string, actorID *uint) (*ReferralActionResult, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, ErrReferralReasonRequired
	}
	if s == nil || s.txManager == nil {
		return nil, ErrReferralServiceUnavailable
	}
	var result ReferralActionResult
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Referral == nil {
			return ErrReferralServiceUnavailable
		}
		record, err := repos.Referral.FindRecordByIDForUpdate(id)
		if err != nil {
			return err
		}
		result.Record = record
		switch record.Status {
		case loyalty.ReferralStatusPending, loyalty.ReferralStatusOrdered, loyalty.ReferralStatusVesting:
			rewardKey := fmt.Sprintf("referral:%d:referrer:points:v1", record.ID)
			if reward, rewardErr := repos.Referral.FindRewardByIdempotencyKey(rewardKey); rewardErr == nil && reward.Status == loyalty.ReferralRewardStatusLocked {
				forfeitedAt := s.now().UTC()
				if err := repos.Referral.UpdateReward(reward.ID, map[string]any{"status": loyalty.ReferralRewardStatusForfeited, "forfeited_at": forfeitedAt, "updated_at": forfeitedAt}); err != nil {
					return err
				}
			} else if rewardErr != nil && !repository.IsRecordNotFound(rewardErr) {
				return rewardErr
			}
			revokedAt := s.now().UTC()
			return transitionReferralRecordAs(repos, record, loyalty.ReferralStatusRevoked, "admin_revoke", strings.TrimSpace(reason), fmt.Sprintf("referral.admin.revoke:%d:%d", record.ID, revokedAt.UnixNano()), map[string]any{"revoked_at": revokedAt, "revoke_reason": strings.TrimSpace(reason)}, "admin", actorID)
		case loyalty.ReferralStatusRevoked:
			return nil
		default:
			return ErrReferralActionInvalid
		}
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ReferralService) SettleMaturedReferrals(ctx context.Context, now time.Time, batchLimit int) (int, error) {
	if s == nil || s.repo == nil {
		return 0, ErrReferralServiceUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if now.IsZero() {
		now = s.now().UTC()
	} else {
		now = now.UTC()
	}
	if batchLimit <= 0 || batchLimit > 500 {
		batchLimit = 100
	}
	records, err := s.repo.ListMaturedVesting(now, batchLimit)
	if err != nil {
		return 0, err
	}
	settled := 0
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return settled, err
		}
		if _, err := s.SettleReferral(record.ID, "vesting period completed", nil); err != nil {
			if errors.Is(err, ErrReferralActionInvalid) {
				continue
			}
			return settled, err
		}
		settled++
	}
	return settled, nil
}

func (s *ReferralService) activeIdentity(code string) (*loyalty.ReferralProgramConfig, *loyalty.ReferralIdentity, error) {
	if s == nil || s.programRepo == nil || s.repo == nil || s.userRepo == nil {
		return nil, nil, ErrReferralServiceUnavailable
	}
	config, err := s.programRepo.FindActive()
	if err != nil {
		return nil, nil, err
	}
	if err := validateReferralProgramConfig(config); err != nil {
		return nil, nil, err
	}
	if !config.Enabled {
		return nil, nil, ErrReferralProgramDisabled
	}
	identity, err := s.repo.FindActiveIdentityByCode(code)
	if repository.IsRecordNotFound(err) || errors.Is(err, loyalty.ErrInvalidReferralCode) {
		return nil, nil, ErrReferralCodeNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	return config, identity, nil
}

func (s *ReferralService) bind(userID uint, claims referralcookie.Claims, bindContext ReferralBindContext) (*loyalty.ReferralRecord, error) {
	if s == nil || s.txManager == nil || userID == 0 || len(s.hashKey) == 0 {
		return nil, ErrReferralServiceUnavailable
	}
	var bound *loyalty.ReferralRecord
	var expectedReferrerID uint
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Referral == nil || repos.ReferralProgram == nil || repos.Order == nil {
			return ErrReferralServiceUnavailable
		}
		config, err := repos.ReferralProgram.FindActive()
		if err != nil {
			return err
		}
		if err := validateReferralProgramConfig(config); err != nil {
			return err
		}
		if !config.Enabled {
			return ErrReferralProgramDisabled
		}
		identity, err := repos.Referral.FindActiveIdentityByCode(claims.Code)
		if repository.IsRecordNotFound(err) {
			return ErrReferralCodeNotFound
		}
		if err != nil {
			return err
		}
		expectedReferrerID = identity.UserID
		if identity.UserID == userID {
			return ErrSelfReferralForbidden
		}
		referrerEmail := ""
		if s.userRepo != nil {
			if referrer, findErr := s.userRepo.FindByID(identity.UserID); findErr == nil && referrer != nil {
				referrerEmail = strings.ToLower(strings.TrimSpace(referrer.Email))
			}
		}
		refereeEmail := strings.ToLower(strings.TrimSpace(bindContext.RefereeEmail))
		if referrerEmail != "" && refereeEmail != "" && referrerEmail == refereeEmail {
			return ErrSelfReferralForbidden
		}
		existing, err := repos.Referral.FindRecordByRefereeID(userID)
		if err == nil {
			if existing.ReferrerID == identity.UserID {
				bound = existing
				return nil
			}
			return ErrRefereeAlreadyAttributed
		}
		if !repository.IsRecordNotFound(err) {
			return err
		}
		paidOrders, err := repos.Order.CountEverPaidOrdersForUserBefore(userID, 0)
		if err != nil {
			return err
		}
		if paidOrders > 0 {
			return ErrRefereeNotEligible
		}

		now := s.now().UTC()
		if !claims.ExpiresAt.After(now) {
			return referralcookie.ErrInvalidCookie
		}

		clientIPHash := s.hashSensitiveValue("client-ip", bindContext.ClientIP)
		clientIPSubnetHash := s.hashSensitiveValue("client-ip-subnet", referralIPSubnetKey(bindContext.ClientIP))
		deviceFingerprintHash := s.hashSensitiveValue("device-fingerprint", bindContext.DeviceFingerprint)
		var bindRiskFlags []map[string]any
		addBindRiskFlag := func(flagType, level, source string) {
			bindRiskFlags = append(bindRiskFlags, map[string]any{
				"type": flagType, "level": level, "source": source,
			})
		}
		if clientIPSubnetHash != "" {
			recentBindings, countErr := repos.Referral.CountRecentBindingsByIPSubnetHash(clientIPSubnetHash, now.Add(-24*time.Hour))
			if countErr != nil {
				return countErr
			}
			if recentBindings >= 3 {
				if config.AntiFraudMode == loyalty.ReferralFraudModeStrict {
					return ErrRefereeNotEligible
				}
				addBindRiskFlag("ip_subnet_velocity", "high", "referral_binding_24h")
			}
		}
		if deviceFingerprintHash != "" {
			referrerRecord, recordErr := repos.Referral.FindRecordByRefereeID(identity.UserID)
			if recordErr == nil && referrerRecord.DeviceFingerprintHash == deviceFingerprintHash {
				if config.AntiFraudMode == loyalty.ReferralFraudModeStrict {
					return ErrRefereeNotEligible
				}
				addBindRiskFlag("device_fingerprint_match", "high", "referrer_historical_attribution")
			} else if recordErr != nil && !repository.IsRecordNotFound(recordErr) {
				return recordErr
			}
		}
		refereeID := userID
		// Keep the signed token expiry as an audit snapshot only. The token has
		// already been validated at bind time; after this point the referral
		// relationship and any registration points are permanent.
		record := &loyalty.ReferralRecord{
			ReferralIdentityID:    identity.ID,
			ProgramConfigID:       config.ID,
			ReferrerID:            identity.UserID,
			RefereeID:             &refereeID,
			ReferralCodeSnapshot:  identity.ReferralCode,
			AttributionSource:     claims.Source,
			RefereeEmailHash:      s.hashSensitiveValue("referee-email", bindContext.RefereeEmail),
			ClientIPHash:          clientIPHash,
			ClientIPSubnetHash:    clientIPSubnetHash,
			DeviceFingerprintHash: deviceFingerprintHash,
			HashKeyVersion:        1,
			Currency:              config.Currency,
			Status:                loyalty.ReferralStatusPending,
			RecordVersion:         1,
			ExpiresAt:             claims.ExpiresAt.UTC(),
		}
		if len(bindRiskFlags) > 0 {
			encodedRiskFlags, marshalErr := json.Marshal(bindRiskFlags)
			if marshalErr != nil {
				return marshalErr
			}
			record.RiskFlags = datatypes.JSON(encodedRiskFlags)
		}
		if err := repos.Referral.CreateRecord(record); err != nil {
			if repository.IsDuplicatedKey(err) {
				return ErrRefereeAlreadyAttributed
			}
			return err
		}
		eventKey := fmt.Sprintf("referral.bound:%d", record.ID)
		if err := repos.Referral.AppendTransition(&loyalty.ReferralTransition{
			ReferralRecordID: record.ID,
			FromStatus:       "unbound",
			ToStatus:         loyalty.ReferralStatusPending,
			Trigger:          "account_binding",
			ActorType:        "customer",
			ActorID:          &refereeID,
			EventKey:         &eventKey,
		}); err != nil {
			return err
		}
		if config.RefereeBenefitType == loyalty.ReferralBenefitPoints {
			if err := s.releaseRefereeBenefitInTx(repos, record, config, now); err != nil {
				return err
			}
		}
		bound = record
		return nil
	})
	if errors.Is(err, ErrRefereeAlreadyAttributed) && expectedReferrerID > 0 {
		existing, findErr := s.repo.FindRecordByRefereeID(userID)
		if findErr == nil && existing.ReferrerID == expectedReferrerID {
			return existing, nil
		}
	}
	return bound, err
}

func (s *ReferralService) hashSensitiveValue(scope, value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || len(s.hashKey) == 0 {
		return ""
	}
	mac := hmac.New(sha256.New, s.hashKey)
	_, _ = mac.Write([]byte("referral-hash:v1:" + scope + ":" + value))
	return hex.EncodeToString(mac.Sum(nil))
}

// referralIPSubnetKey returns the comparison-only network identity used for
// the rolling binding cap. IPv4 is grouped by /24 and IPv6 by /64. The raw IP
// is never persisted; callers HMAC this value before storing it.
func referralIPSubnetKey(raw string) string {
	ip := net.ParseIP(strings.TrimSpace(raw))
	if ip == nil {
		return ""
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		return fmt.Sprintf("%d.%d.%d.0/24", ipv4[0], ipv4[1], ipv4[2])
	}
	mask := net.CIDRMask(64, 128)
	return ip.Mask(mask).String() + "/64"
}

func validateReferralProgramConfig(config *loyalty.ReferralProgramConfig) error {
	if config == nil || config.Version <= 0 || !currency.IsCatalogCode(config.Currency) ||
		config.MinOrderAmountMinor < 0 || config.ReferrerRewardPoints < 0 || config.RefereeBenefitValue <= 0 || config.VestingPeriodDays <= 0 ||
		config.UndeliveredFallbackDays < config.VestingPeriodDays || config.AttributionTTLDays <= 0 ||
		config.AttributionTTLDays > 90 || config.MonthlyCapPerReferrer <= 0 {
		return ErrInvalidReferralProgramConfig
	}
	if config.RefereeBenefitType != loyalty.ReferralBenefitPoints {
		return fmt.Errorf("%w: invalid referral benefit type", ErrInvalidReferralProgramConfig)
	}
	if config.AntiFraudMode != loyalty.ReferralFraudModeMonitor && config.AntiFraudMode != loyalty.ReferralFraudModeStrict {
		return fmt.Errorf("%w: invalid referral anti-fraud mode", ErrInvalidReferralProgramConfig)
	}
	return nil
}

func referralDashboardRules(config *loyalty.ReferralProgramConfig) ReferralDashboardRules {
	return ReferralDashboardRules{
		MinOrderAmountMinor:      config.MinOrderAmountMinor,
		ReferrerRewardPoints:     config.ReferrerRewardPoints,
		RefereeBenefitType:       config.RefereeBenefitType,
		RefereeBenefitValue:      config.RefereeBenefitValue,
		VestingPeriodDays:        config.VestingPeriodDays,
		AttributionCookieTTLDays: config.AttributionTTLDays,
	}
}

func maskReferralEmail(email string) string {
	email = strings.TrimSpace(email)
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}
	local := []rune(parts[0])
	if len(local) > 1 {
		local = []rune(string(local[:1]) + "***")
	}
	return string(local) + "@" + parts[1]
}

func maskReferralName(value *user.User) string {
	if value == nil {
		return ""
	}
	name := strings.TrimSpace(strings.TrimSpace(value.FirstName) + " " + strings.TrimSpace(value.LastName))
	if name == "" {
		name = strings.TrimSpace(value.Username)
	}
	parts := strings.Fields(name)
	for index, part := range parts {
		runes := []rune(part)
		if len(runes) > 1 {
			parts[index] = string(runes[0]) + strings.Repeat("*", len(runes)-1)
		}
	}
	return strings.Join(parts, " ")
}
