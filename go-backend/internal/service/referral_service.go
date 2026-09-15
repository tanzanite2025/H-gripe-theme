package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/loyalty"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/user"
	referralcookie "commerce-platform/internal/pkg/referral"
	"commerce-platform/internal/repository"
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

const (
	referralIdentityCreateAttempts = 8
	referralRefereeReversalSource  = "referral_referee_reversal"
)

type ReferralValidation struct {
	Valid                    bool   `json:"valid"`
	ReferrerNameMask         string `json:"referrer_name_mask"`
	RefereeBenefitType       string `json:"referee_benefit_type"`
	RefereeBenefitValue      int64  `json:"referee_benefit_value"`
	BenefitCurrency          string `json:"benefit_currency"`
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
	Currency                 string `json:"currency"`
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
	Enabled                      bool
	Currency                     string
	MinOrderAmountMinor          int64
	ReferrerRewardPoints         int
	RefereeBenefitType           string
	RefereeBenefitValue          int64
	RefereeBenefitMaxAmountMinor int64
	CouponStackable              bool
	VestingPeriodDays            int
	UndeliveredFallbackDays      int
	AttributionTTLDays           int
	MonthlyCapPerReferrer        int
	AntiFraudMode                string
	CreatedBy                    *uint
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
	baseURL       string
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
		baseURL:       strings.TrimRight(strings.TrimSpace(baseURL), "/"),
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
		BenefitCurrency:          config.Currency,
		MinOrderAmountMinor:      config.MinOrderAmountMinor,
		VestingPeriodDays:        config.VestingPeriodDays,
		AttributionCookieTTLDays: config.AttributionTTLDays,
	}, nil
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
	if s == nil || s.programRepo == nil || s.repo == nil || userID == 0 {
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
	identity, err := s.getOrCreateIdentity(userID)
	if err != nil {
		return nil, err
	}
	stats, err := s.repo.StatsByReferrerID(userID)
	if err != nil {
		return nil, err
	}
	dashboard.ReferralCode = identity.ReferralCode
	dashboard.CustomSlug = identity.CustomSlug
	dashboard.ShareURL = s.baseURL + "/r/" + identity.ReferralCode
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
	overview, err := s.repo.AdminStats()
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
	config := &loyalty.ReferralProgramConfig{
		Version: expectedVersion + 1,
		Enabled: input.Enabled, Currency: currency.NormalizeCode(input.Currency),
		MinOrderAmountMinor: input.MinOrderAmountMinor, ReferrerRewardPoints: input.ReferrerRewardPoints,
		RefereeBenefitType: strings.ToLower(strings.TrimSpace(input.RefereeBenefitType)), RefereeBenefitValue: input.RefereeBenefitValue,
		RefereeBenefitMaxAmountMinor: input.RefereeBenefitMaxAmountMinor, CouponStackable: input.CouponStackable,
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

// releaseRefereeBenefitInTx materializes the benefit promised by the active
// referral policy after the referee's qualifying first payment. The reward log
// is the idempotency boundary: retries reuse the same record and never issue a
// second points transaction or coupon.
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
	if benefitType == loyalty.ReferralBenefitNone || config.RefereeBenefitValue <= 0 {
		return nil
	}
	if benefitType != loyalty.ReferralBenefitPoints && benefitType != loyalty.ReferralBenefitPercentCoupon && benefitType != loyalty.ReferralBenefitFixedCoupon {
		return ErrInvalidReferralProgramConfig
	}
	rewardType := loyalty.ReferralRewardTypePoints
	if benefitType != loyalty.ReferralBenefitPoints {
		rewardType = loyalty.ReferralRewardTypeCoupon
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
		"version":                          config.Version,
		"currency":                         config.Currency,
		"referee_benefit_type":             config.RefereeBenefitType,
		"referee_benefit_value":            config.RefereeBenefitValue,
		"referee_benefit_max_amount_minor": config.RefereeBenefitMaxAmountMinor,
		"coupon_stackable":                 config.CouponStackable,
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
			RewardType:       rewardType,
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
	} else {
		if repos.Coupon == nil {
			return ErrReferralServiceUnavailable
		}
		if reward.ID == 0 || reward.CouponID == nil {
			couponRecord, couponErr := buildRefereeCoupon(record, config, releasedAt)
			if couponErr != nil {
				return couponErr
			}
			if reward.CouponID == nil {
				if err := repos.Coupon.CreateCoupon(couponRecord); err != nil {
					return err
				}
				reward.CouponID = &couponRecord.ID
			}
			if reward.ID == 0 {
				if err := repos.Referral.CreateReward(reward); err != nil {
					return err
				}
			}
		}
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

func (s *ReferralService) provisionLockedRefereeCouponInTx(
	repos repository.TxRepositories,
	record *loyalty.ReferralRecord,
	config *loyalty.ReferralProgramConfig,
	createdAt time.Time,
) error {
	if record == nil || config == nil || record.RefereeID == nil {
		return ErrInvalidReferralProgramConfig
	}
	if config.RefereeBenefitType != loyalty.ReferralBenefitPercentCoupon && config.RefereeBenefitType != loyalty.ReferralBenefitFixedCoupon {
		return nil
	}
	if repos.Referral == nil || repos.Coupon == nil {
		return ErrReferralServiceUnavailable
	}
	rewardKey := fmt.Sprintf("referral:%d:referee:%s:v1", record.ID, config.RefereeBenefitType)
	_, err := repos.Referral.FindRewardByIdempotencyKey(rewardKey)
	if err == nil {
		return nil
	}
	if !repository.IsRecordNotFound(err) {
		return err
	}
	couponRecord, err := buildRefereeCoupon(record, config, createdAt)
	if err != nil {
		return err
	}
	if err := repos.Coupon.CreateCoupon(couponRecord); err != nil {
		return err
	}
	snapshot, err := json.Marshal(map[string]any{
		"version":                          config.Version,
		"currency":                         config.Currency,
		"referee_benefit_type":             config.RefereeBenefitType,
		"referee_benefit_value":            config.RefereeBenefitValue,
		"referee_benefit_max_amount_minor": config.RefereeBenefitMaxAmountMinor,
		"coupon_stackable":                 config.CouponStackable,
	})
	if err != nil {
		return err
	}
	return repos.Referral.CreateReward(&loyalty.ReferralReward{
		ReferralRecordID: record.ID,
		ProgramConfigID:  record.ProgramConfigID,
		RecipientUserID:  *record.RefereeID,
		RecipientRole:    loyalty.ReferralRecipientReferee,
		RewardType:       loyalty.ReferralRewardTypeCoupon,
		CouponID:         &couponRecord.ID,
		IdempotencyKey:   rewardKey,
		Status:           loyalty.ReferralRewardStatusLocked,
		RuleSnapshot:     snapshot,
	})
}

func (s *ReferralService) forfeitLockedRefereeBenefitInTx(repos repository.TxRepositories, record *loyalty.ReferralRecord, at time.Time) error {
	if record == nil || repos.Referral == nil {
		return ErrReferralServiceUnavailable
	}
	for _, benefitType := range []string{loyalty.ReferralBenefitFixedCoupon, loyalty.ReferralBenefitPercentCoupon, loyalty.ReferralBenefitPoints} {
		key := fmt.Sprintf("referral:%d:referee:%s:v1", record.ID, benefitType)
		reward, err := repos.Referral.FindRewardByIdempotencyKey(key)
		if repository.IsRecordNotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		if reward.Status != loyalty.ReferralRewardStatusLocked {
			continue
		}
		if err := repos.Referral.UpdateReward(reward.ID, map[string]any{
			"status":       loyalty.ReferralRewardStatusForfeited,
			"forfeited_at": at,
			"updated_at":   at,
		}); err != nil {
			return err
		}
		if reward.CouponID != nil && repos.Coupon != nil {
			couponRecord, findErr := repos.Coupon.FindCouponByIDForUpdate(*reward.CouponID)
			if repository.IsRecordNotFound(findErr) {
				continue
			}
			if findErr != nil {
				return findErr
			}
			couponRecord.Enabled = false
			if err := repos.Coupon.UpdateCoupon(couponRecord); err != nil {
				return err
			}
		}
	}
	return nil
}

// reverseReleasedRefereeBenefitInTx claws back a referee benefit that was
// released at payment time before the order was later invalidated. The
// referral record lock serializes this with payment and settlement handlers;
// the reversal ledger source makes retries idempotent.
func (s *ReferralService) reverseReleasedRefereeBenefitInTx(repos repository.TxRepositories, record *loyalty.ReferralRecord, at time.Time) error {
	if record == nil || repos.Referral == nil {
		return ErrReferralServiceUnavailable
	}
	for _, benefitType := range []string{loyalty.ReferralBenefitFixedCoupon, loyalty.ReferralBenefitPercentCoupon, loyalty.ReferralBenefitPoints} {
		key := fmt.Sprintf("referral:%d:referee:%s:v1", record.ID, benefitType)
		reward, err := repos.Referral.FindRewardByIdempotencyKey(key)
		if repository.IsRecordNotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		if reward.Status != loyalty.ReferralRewardStatusReleased {
			continue
		}

		if reward.RewardType == loyalty.ReferralRewardTypePoints && reward.PointsAmount > 0 {
			if repos.Loyalty == nil {
				return ErrReferralServiceUnavailable
			}
			count, countErr := repos.Loyalty.CountTransactionsByUserTypeSourceAndSourceID(
				reward.RecipientUserID,
				"adjust",
				referralRefereeReversalSource,
				record.ID,
			)
			if countErr != nil {
				return countErr
			}
			if count == 0 {
				if _, adjustErr := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
					reward.RecipientUserID,
					-reward.PointsAmount,
					"adjust",
					referralRefereeReversalSource,
					record.ID,
					fmt.Sprintf("Referral referee reward reversal for record #%d", record.ID),
					&record.ProgramConfigID,
				); adjustErr != nil {
					return adjustErr
				}
			}
		}

		if reward.CouponID != nil && repos.Coupon != nil {
			couponRecord, findErr := repos.Coupon.FindCouponByIDForUpdate(*reward.CouponID)
			if repository.IsRecordNotFound(findErr) {
				continue
			}
			if findErr != nil {
				return findErr
			}
			if couponRecord.Enabled {
				couponRecord.Enabled = false
				if updateErr := repos.Coupon.UpdateCoupon(couponRecord); updateErr != nil {
					return updateErr
				}
			}
		}

		if updateErr := repos.Referral.UpdateReward(reward.ID, map[string]any{
			"status":      loyalty.ReferralRewardStatusReversed,
			"reversed_at": at,
			"updated_at":  at,
		}); updateErr != nil {
			return updateErr
		}
	}
	return nil
}

// refereeCouponMatchesOrderInTx verifies that a qualifying order actually
// consumed the private coupon provisioned for this referral. The order is
// already locked by the lifecycle handler; lock the reward coupon as well so
// a concurrent coupon disable/use cannot race this decision.
func (s *ReferralService) refereeCouponMatchesOrderInTx(
	repos repository.TxRepositories,
	record *loyalty.ReferralRecord,
	config *loyalty.ReferralProgramConfig,
	orderRecord *order.Order,
) (bool, error) {
	if record == nil || config == nil || orderRecord == nil || repos.Referral == nil || repos.Coupon == nil {
		return false, ErrReferralServiceUnavailable
	}
	orderCouponCode := strings.TrimSpace(orderRecord.CouponCode)
	rewardKey := fmt.Sprintf("referral:%d:referee:%s:v1", record.ID, config.RefereeBenefitType)
	reward, err := repos.Referral.FindRewardByIdempotencyKey(rewardKey)
	if repository.IsRecordNotFound(err) {
		// Bind normally provisions the locked reward. If an older or manually
		// repaired record lacks that row, allow the payment path to materialize
		// the configured benefit only when no competing coupon was selected.
		return orderCouponCode == "", nil
	}
	if err != nil {
		return false, err
	}
	if orderCouponCode == "" || reward.CouponID == nil {
		return false, nil
	}
	couponRecord, err := repos.Coupon.FindCouponByIDForUpdate(*reward.CouponID)
	if repository.IsRecordNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return strings.EqualFold(orderCouponCode, strings.TrimSpace(couponRecord.Code)), nil
}

func buildRefereeCoupon(record *loyalty.ReferralRecord, config *loyalty.ReferralProgramConfig, start time.Time) (*coupon.Coupon, error) {
	if record == nil || config == nil {
		return nil, ErrInvalidReferralProgramConfig
	}
	benefitCurrency := currency.NormalizeCode(config.Currency)
	value := float64(config.RefereeBenefitValue)
	typeName := "fixed"
	if config.RefereeBenefitType == loyalty.ReferralBenefitPercentCoupon {
		typeName = "percentage"
		value /= 100
	} else {
		money, err := domainmoney.New(config.RefereeBenefitValue, benefitCurrency)
		if err != nil {
			return nil, err
		}
		value, err = money.MajorFloat()
		if err != nil {
			return nil, err
		}
	}
	if value <= 0 {
		return nil, ErrInvalidReferralProgramConfig
	}
	if start.IsZero() {
		start = time.Now().UTC()
	} else {
		start = start.UTC()
	}
	couponValidityDays := config.AttributionTTLDays
	if couponValidityDays <= 0 {
		couponValidityDays = 30
	}
	end := start.AddDate(0, 0, couponValidityDays)
	minimumMoney, err := domainmoney.New(config.MinOrderAmountMinor, benefitCurrency)
	if err != nil {
		return nil, err
	}
	minimum, err := minimumMoney.MajorFloat()
	if err != nil {
		return nil, err
	}
	maxDiscount := float64(0)
	if config.RefereeBenefitMaxAmountMinor > 0 {
		maxDiscountMoney, moneyErr := domainmoney.New(config.RefereeBenefitMaxAmountMinor, benefitCurrency)
		if moneyErr != nil {
			return nil, moneyErr
		}
		maxDiscount, err = maxDiscountMoney.MajorFloat()
		if err != nil {
			return nil, err
		}
	}
	return &coupon.Coupon{
		Code:                    fmt.Sprintf("REFERRAL-%d-%s", record.ID, strings.ToUpper(strings.TrimSuffix(config.RefereeBenefitType, "_coupon"))),
		Type:                    typeName,
		Value:                   value,
		Currency:                benefitCurrency,
		Description:             fmt.Sprintf("Referral welcome benefit for referral #%d", record.ID),
		MinAmount:               minimum,
		MaxDiscount:             maxDiscount,
		UsageLimit:              1,
		UsageLimitPerUser:       1,
		ReferralRecipientUserID: record.RefereeID,
		StartDate:               start.Add(-time.Minute),
		EndDate:                 end,
		Enabled:                 true,
	}, nil
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
				"version": config.Version, "currency": config.Currency,
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
			if err := s.forfeitLockedRefereeBenefitInTx(repos, record, revokedAt); err != nil {
				return err
			}
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

func (s *ReferralService) getOrCreateIdentity(userID uint) (*loyalty.ReferralIdentity, error) {
	identity, err := s.repo.FindIdentityByUserID(userID)
	if err == nil {
		return identity, nil
	}
	if !repository.IsRecordNotFound(err) {
		return nil, err
	}
	for attempt := 0; attempt < referralIdentityCreateAttempts; attempt++ {
		code, generateErr := loyalty.GenerateReferralCode()
		if generateErr != nil {
			return nil, generateErr
		}
		identity = &loyalty.ReferralIdentity{UserID: userID, ReferralCode: code, IsActive: true}
		if createErr := s.repo.CreateIdentity(identity); createErr == nil {
			return identity, nil
		} else if !repository.IsDuplicatedKey(createErr) {
			return nil, createErr
		}
		if existing, findErr := s.repo.FindIdentityByUserID(userID); findErr == nil {
			return existing, nil
		}
	}
	return nil, errors.New("could not allocate a unique referral code")
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
		refereeID := userID
		record := &loyalty.ReferralRecord{
			ReferralIdentityID:    identity.ID,
			ProgramConfigID:       config.ID,
			ReferrerID:            identity.UserID,
			RefereeID:             &refereeID,
			ReferralCodeSnapshot:  identity.ReferralCode,
			AttributionSource:     claims.Source,
			RefereeEmailHash:      s.hashSensitiveValue("referee-email", bindContext.RefereeEmail),
			ClientIPHash:          s.hashSensitiveValue("client-ip", bindContext.ClientIP),
			DeviceFingerprintHash: s.hashSensitiveValue("device-fingerprint", bindContext.DeviceFingerprint),
			HashKeyVersion:        1,
			Currency:              config.Currency,
			Status:                loyalty.ReferralStatusPending,
			RecordVersion:         1,
			ExpiresAt:             claims.ExpiresAt.UTC(),
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
		if err := s.provisionLockedRefereeCouponInTx(repos, record, config, now); err != nil {
			return err
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

func validateReferralProgramConfig(config *loyalty.ReferralProgramConfig) error {
	if config == nil || config.Version <= 0 || !currency.IsCatalogCode(config.Currency) ||
		config.MinOrderAmountMinor < 0 || config.RefereeBenefitMaxAmountMinor < 0 ||
		config.ReferrerRewardPoints < 0 || config.RefereeBenefitValue < 0 || config.VestingPeriodDays <= 0 ||
		config.UndeliveredFallbackDays < config.VestingPeriodDays || config.AttributionTTLDays <= 0 ||
		config.AttributionTTLDays > 90 || config.MonthlyCapPerReferrer <= 0 {
		return ErrInvalidReferralProgramConfig
	}
	switch config.RefereeBenefitType {
	case loyalty.ReferralBenefitNone, loyalty.ReferralBenefitPoints, loyalty.ReferralBenefitFixedCoupon:
	case loyalty.ReferralBenefitPercentCoupon:
		if config.RefereeBenefitValue > 10000 {
			return fmt.Errorf("%w: invalid referral percent benefit", ErrInvalidReferralProgramConfig)
		}
	default:
		return fmt.Errorf("%w: invalid referral benefit type", ErrInvalidReferralProgramConfig)
	}
	if config.AntiFraudMode != loyalty.ReferralFraudModeMonitor && config.AntiFraudMode != loyalty.ReferralFraudModeStrict {
		return fmt.Errorf("%w: invalid referral anti-fraud mode", ErrInvalidReferralProgramConfig)
	}
	return nil
}

func referralDashboardRules(config *loyalty.ReferralProgramConfig) ReferralDashboardRules {
	return ReferralDashboardRules{
		Currency:                 config.Currency,
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
