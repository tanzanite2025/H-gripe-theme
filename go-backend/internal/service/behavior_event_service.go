package service

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"commerce-platform/internal/domain/recommendation"
	attributionpkg "commerce-platform/internal/pkg/attribution"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

const (
	MaxBehaviorEventBatchSize = 50
	maxBehaviorEventAge       = 30 * 24 * time.Hour
	maxBehaviorEventFuture    = 10 * time.Minute
	maxBehaviorMetadataKeys   = 20
	maxBehaviorMetadataBytes  = 4096
)

var (
	ErrBehaviorEventBatchEmpty         = errors.New("event batch is empty")
	ErrBehaviorEventBatchTooLarge      = errors.New("event batch is too large")
	ErrBehaviorEventIDRequired         = errors.New("event_id is required")
	ErrBehaviorEventIDInvalid          = errors.New("event_id contains invalid characters")
	ErrBehaviorEventTypeInvalid        = errors.New("event_type is not supported")
	ErrBehaviorEventIdentityInvalid    = errors.New("identity token contains invalid characters")
	ErrBehaviorEventIdentityRequired   = errors.New("anonymous_id, session_id, or authenticated user is required")
	ErrBehaviorEventTimestampInvalid   = errors.New("occurred_at is outside the accepted time window")
	ErrBehaviorEventMetadataInvalid    = errors.New("metadata is invalid")
	ErrBehaviorEventAttributionInvalid = errors.New("advertising attribution metadata is invalid")
)

var behaviorEventIDPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)
var behaviorIdentityPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)

var supportedBehaviorEventTypes = map[string]struct{}{
	"page_view":                 {},
	"product_view":              {},
	"product_dwell":             {},
	"search_submit":             {},
	"filter_apply":              {},
	"category_navigation_click": {},
	"calculator_use":            {},
	"recommendation_impression": {},
	"recommendation_click":      {},
	"add_to_cart":               {},
	"wishlist_add":              {},
	"begin_checkout":            {},
	"ad_landing":                {},
	"quiz_completed":            {},
}

type BehaviorEventInput struct {
	EventID     string
	EventType   string
	AnonymousID string
	SessionID   string
	ProductID   *uint
	CategoryID  *uint
	Locale      string
	Path        string
	Referrer    string
	Metadata    map[string]any
	OccurredAt  time.Time
}

type BehaviorEventIngestResult struct {
	Received   int   `json:"received"`
	Accepted   int64 `json:"accepted"`
	Duplicates int   `json:"duplicates"`
}

type BehaviorEventCleanupResult struct {
	DeletedLowIntent      int64     `json:"deleted_low_intent"`
	DeletedStandardIntent int64     `json:"deleted_standard_intent"`
	DeletedHighIntent     int64     `json:"deleted_high_intent"`
	TotalDeleted          int64     `json:"total_deleted"`
	LowIntentCutoff       time.Time `json:"low_intent_cutoff"`
	StandardIntentCutoff  time.Time `json:"standard_intent_cutoff"`
	HighIntentCutoff      time.Time `json:"high_intent_cutoff"`
}

type BehaviorEventService struct {
	eventRepo       *repository.RecommendationEventRepository
	retentionPolicy BehaviorEventRetentionPolicy
}

type BehaviorEventRetentionPolicy struct {
	LowIntentRetentionDays      int
	StandardIntentRetentionDays int
	HighIntentRetentionDays     int
	CleanupBatchLimit           int
}

func NewBehaviorEventService(eventRepo *repository.RecommendationEventRepository, cfg ...config.BehaviorEventsConfig) *BehaviorEventService {
	policy := defaultBehaviorEventRetentionPolicy()
	if len(cfg) > 0 {
		policy = behaviorEventRetentionPolicyFromConfig(cfg[0])
	}
	return &BehaviorEventService{eventRepo: eventRepo, retentionPolicy: policy}
}

func (s *BehaviorEventService) Ingest(userID *uint, inputs []BehaviorEventInput) (BehaviorEventIngestResult, error) {
	result := BehaviorEventIngestResult{Received: len(inputs)}
	if len(inputs) == 0 {
		return result, ErrBehaviorEventBatchEmpty
	}
	if len(inputs) > MaxBehaviorEventBatchSize {
		return result, ErrBehaviorEventBatchTooLarge
	}
	if s == nil || s.eventRepo == nil {
		return result, errors.New("behavior event repository is not configured")
	}

	events := make([]recommendation.Event, 0, len(inputs))
	now := time.Now().UTC()
	for _, input := range inputs {
		event, err := normalizeBehaviorEvent(userID, input, now)
		if err != nil {
			return result, err
		}
		events = append(events, event)
	}

	accepted, err := s.eventRepo.CreateBatch(events)
	if err != nil {
		return result, err
	}

	result.Accepted = accepted
	result.Duplicates = len(inputs) - int(accepted)
	return result, nil
}

func (s *BehaviorEventService) CleanupExpiredEvents(now time.Time) (BehaviorEventCleanupResult, error) {
	result := BehaviorEventCleanupResult{}
	if s == nil || s.eventRepo == nil {
		return result, nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	policy := s.retentionPolicy.normalized()
	result.LowIntentCutoff = now.AddDate(0, 0, -policy.LowIntentRetentionDays)
	result.StandardIntentCutoff = now.AddDate(0, 0, -policy.StandardIntentRetentionDays)
	result.HighIntentCutoff = now.AddDate(0, 0, -policy.HighIntentRetentionDays)

	deletedLow, err := s.eventRepo.DeleteExpiredByTypes(lowIntentBehaviorEventTypes(), result.LowIntentCutoff, policy.CleanupBatchLimit)
	if err != nil {
		return result, err
	}
	result.DeletedLowIntent = deletedLow

	deletedStandard, err := s.eventRepo.DeleteExpiredByTypes(standardIntentBehaviorEventTypes(), result.StandardIntentCutoff, policy.CleanupBatchLimit)
	if err != nil {
		return result, err
	}
	result.DeletedStandardIntent = deletedStandard

	deletedHigh, err := s.eventRepo.DeleteExpiredByTypes(highIntentBehaviorEventTypes(), result.HighIntentCutoff, policy.CleanupBatchLimit)
	if err != nil {
		return result, err
	}
	result.DeletedHighIntent = deletedHigh
	result.TotalDeleted = deletedLow + deletedStandard + deletedHigh
	return result, nil
}

func normalizeBehaviorEvent(userID *uint, input BehaviorEventInput, receivedAt time.Time) (recommendation.Event, error) {
	eventID := strings.TrimSpace(input.EventID)
	if eventID == "" {
		return recommendation.Event{}, ErrBehaviorEventIDRequired
	}
	if !behaviorEventIDPattern.MatchString(eventID) {
		return recommendation.Event{}, ErrBehaviorEventIDInvalid
	}

	eventType := strings.TrimSpace(input.EventType)
	if _, ok := supportedBehaviorEventTypes[eventType]; !ok {
		return recommendation.Event{}, ErrBehaviorEventTypeInvalid
	}

	anonymousID, err := normalizeIdentityToken(input.AnonymousID)
	if err != nil {
		return recommendation.Event{}, err
	}
	sessionID, err := normalizeIdentityToken(input.SessionID)
	if err != nil {
		return recommendation.Event{}, err
	}
	if anonymousID == "" && sessionID == "" && (userID == nil || *userID == 0) {
		return recommendation.Event{}, ErrBehaviorEventIdentityRequired
	}

	occurredAt := input.OccurredAt.UTC()
	if occurredAt.IsZero() ||
		occurredAt.Before(receivedAt.Add(-maxBehaviorEventAge)) ||
		occurredAt.After(receivedAt.Add(maxBehaviorEventFuture)) {
		return recommendation.Event{}, ErrBehaviorEventTimestampInvalid
	}

	metadata, err := normalizeBehaviorMetadata(input.Metadata)
	if err != nil {
		return recommendation.Event{}, err
	}
	if eventType == "ad_landing" {
		attribution, ok := attributionpkg.FromMetadata(input.Metadata)
		if !ok {
			return recommendation.Event{}, ErrBehaviorEventAttributionInvalid
		}
		metadata, err = normalizeBehaviorMetadata(attribution.Metadata())
		if err != nil {
			return recommendation.Event{}, ErrBehaviorEventAttributionInvalid
		}
	}

	return recommendation.Event{
		EventID:      eventID,
		EventType:    eventType,
		AnonymousID:  anonymousID,
		SessionID:    sessionID,
		UserID:       normalizeOptionalID(userID),
		ProductID:    normalizeOptionalID(input.ProductID),
		CategoryID:   normalizeOptionalID(input.CategoryID),
		Locale:       normalizeBehaviorLocale(input.Locale),
		Path:         normalizeText(input.Path, 1024),
		Referrer:     normalizeText(input.Referrer, 1024),
		MetadataJSON: metadata,
		OccurredAt:   occurredAt,
		ReceivedAt:   receivedAt,
	}, nil
}

func defaultBehaviorEventRetentionPolicy() BehaviorEventRetentionPolicy {
	return BehaviorEventRetentionPolicy{
		LowIntentRetentionDays:      30,
		StandardIntentRetentionDays: 60,
		HighIntentRetentionDays:     180,
		CleanupBatchLimit:           5000,
	}
}

func behaviorEventRetentionPolicyFromConfig(cfg config.BehaviorEventsConfig) BehaviorEventRetentionPolicy {
	return BehaviorEventRetentionPolicy{
		LowIntentRetentionDays:      cfg.LowIntentRetentionDays,
		StandardIntentRetentionDays: cfg.StandardIntentRetentionDays,
		HighIntentRetentionDays:     cfg.HighIntentRetentionDays,
		CleanupBatchLimit:           cfg.CleanupBatchLimit,
	}.normalized()
}

func (policy BehaviorEventRetentionPolicy) normalized() BehaviorEventRetentionPolicy {
	defaults := defaultBehaviorEventRetentionPolicy()
	if policy.LowIntentRetentionDays <= 0 {
		policy.LowIntentRetentionDays = defaults.LowIntentRetentionDays
	}
	if policy.StandardIntentRetentionDays <= 0 {
		policy.StandardIntentRetentionDays = defaults.StandardIntentRetentionDays
	}
	if policy.HighIntentRetentionDays <= 0 {
		policy.HighIntentRetentionDays = defaults.HighIntentRetentionDays
	}
	if policy.CleanupBatchLimit <= 0 {
		policy.CleanupBatchLimit = defaults.CleanupBatchLimit
	}
	return policy
}

func lowIntentBehaviorEventTypes() []string {
	return []string{
		"page_view",
		"recommendation_impression",
	}
}

func standardIntentBehaviorEventTypes() []string {
	return []string{
		"product_view",
		"product_dwell",
		"search_submit",
		"filter_apply",
		"category_navigation_click",
		"recommendation_click",
		"quiz_completed",
	}
}

func highIntentBehaviorEventTypes() []string {
	return []string{
		"calculator_use",
		"add_to_cart",
		"wishlist_add",
		"begin_checkout",
		"purchase",
	}
}

func normalizeBehaviorMetadata(metadata map[string]any) (datatypes.JSON, error) {
	if metadata == nil {
		return datatypes.JSON([]byte(`{}`)), nil
	}
	if len(metadata) > maxBehaviorMetadataKeys {
		return nil, ErrBehaviorEventMetadataInvalid
	}

	sanitized := make(map[string]any, len(metadata))
	for key, value := range metadata {
		key = strings.TrimSpace(key)
		if key == "" || len(key) > 64 {
			return nil, ErrBehaviorEventMetadataInvalid
		}

		switch typed := value.(type) {
		case nil, bool, float64, json.Number:
			sanitized[key] = typed
		case string:
			if len(typed) > 512 {
				return nil, ErrBehaviorEventMetadataInvalid
			}
			sanitized[key] = typed
		default:
			return nil, ErrBehaviorEventMetadataInvalid
		}
	}

	encoded, err := json.Marshal(sanitized)
	if err != nil || len(encoded) > maxBehaviorMetadataBytes {
		return nil, ErrBehaviorEventMetadataInvalid
	}

	return datatypes.JSON(encoded), nil
}

func normalizeOptionalID(value *uint) *uint {
	if value == nil || *value == 0 {
		return nil
	}
	result := *value
	return &result
}

func normalizeIdentityToken(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if !behaviorIdentityPattern.MatchString(value) {
		return "", ErrBehaviorEventIdentityInvalid
	}
	return value, nil
}

func normalizeText(value string, maxLength int) string {
	value = strings.TrimSpace(value)
	value = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' || r >= 0x20 {
			return r
		}
		return -1
	}, value)
	if len(value) > maxLength {
		return value[:maxLength]
	}
	return value
}

func normalizeBehaviorLocale(value string) string {
	value = strings.TrimSpace(strings.Split(value, ",")[0])
	value = strings.ReplaceAll(value, "_", "-")
	return normalizeText(strings.ToLower(value), 20)
}

func IsBehaviorEventValidationError(err error) bool {
	return errors.Is(err, ErrBehaviorEventBatchEmpty) ||
		errors.Is(err, ErrBehaviorEventBatchTooLarge) ||
		errors.Is(err, ErrBehaviorEventIDRequired) ||
		errors.Is(err, ErrBehaviorEventIDInvalid) ||
		errors.Is(err, ErrBehaviorEventTypeInvalid) ||
		errors.Is(err, ErrBehaviorEventIdentityInvalid) ||
		errors.Is(err, ErrBehaviorEventIdentityRequired) ||
		errors.Is(err, ErrBehaviorEventTimestampInvalid) ||
		errors.Is(err, ErrBehaviorEventMetadataInvalid) ||
		errors.Is(err, ErrBehaviorEventAttributionInvalid)
}
