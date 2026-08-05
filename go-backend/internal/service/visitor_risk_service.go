package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"tanzanite/internal/domain/visitor"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/repository"
)

type VisitorRiskService struct {
	riskRepo        *repository.VisitorRiskFactRepository
	enabled         bool
	hashSalt        string
	maxPendingFacts int
	samplePathLimit int
	retentionDays   int
	mu              sync.Mutex
	pending         map[string]*visitorRiskAccumulator
}

var (
	ErrVisitorRiskFactNotFound       = errors.New("visitor risk fact not found")
	ErrVisitorRiskDecisionInvalid    = errors.New("invalid visitor risk decision")
	ErrVisitorRiskDecisionNoIdentity = errors.New("visitor risk fact has no decision identity")
)

type VisitorRiskRecordInput struct {
	IPAddress        string
	UserAgent        string
	Path             string
	CountryCode      string
	AnonymousID      string
	SessionID        string
	HasCookieHeader  bool
	StatusCode       int
	OccurredAt       time.Time
	MeaningfulAction bool
}

type VisitorRiskFlushResult struct {
	FlushedFacts int `json:"flushed_facts"`
	DroppedFacts int `json:"dropped_facts"`
}

type VisitorRiskCleanupResult struct {
	DeletedFacts     int64     `json:"deleted_facts"`
	DeletedDecisions int64     `json:"deleted_decisions"`
	CutoffDay        time.Time `json:"cutoff_day"`
}

type VisitorRiskFactListInput struct {
	Search       string
	RiskLevel    string
	DayRange     string
	MinRiskScore string
}

type VisitorRiskFactSnapshot struct {
	ID                    uint                         `json:"id"`
	Day                   time.Time                    `json:"day"`
	IPHashPreview         string                       `json:"ip_hash_preview"`
	UserAgentHashPreview  string                       `json:"user_agent_hash_preview,omitempty"`
	CountryCode           string                       `json:"country_code,omitempty"`
	FirstSeenAt           time.Time                    `json:"first_seen_at"`
	LastSeenAt            time.Time                    `json:"last_seen_at"`
	RequestCount          int                          `json:"request_count"`
	UniquePathCount       int                          `json:"unique_path_count"`
	UniqueAnonymousCount  int                          `json:"unique_anonymous_count"`
	UniqueSessionCount    int                          `json:"unique_session_count"`
	InvalidRequestCount   int                          `json:"invalid_request_count"`
	AuthFailureCount      int                          `json:"auth_failure_count"`
	CheckoutFailureCount  int                          `json:"checkout_failure_count"`
	BotLikeUserAgentCount int                          `json:"bot_like_user_agent_count"`
	NoCookieRequestCount  int                          `json:"no_cookie_request_count"`
	MeaningfulActionCount int                          `json:"meaningful_action_count"`
	RiskScore             int                          `json:"risk_score"`
	RiskLevel             string                       `json:"risk_level"`
	SamplePaths           []string                     `json:"sample_paths"`
	Decision              *VisitorRiskDecisionSnapshot `json:"decision,omitempty"`
	CreatedAt             time.Time                    `json:"created_at"`
	UpdatedAt             time.Time                    `json:"updated_at"`
}

type VisitorRiskDecisionSnapshot struct {
	ID        uint       `json:"id"`
	Scope     string     `json:"scope"`
	Action    string     `json:"action"`
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedBy *uint      `json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type VisitorRiskDecisionInput struct {
	Action    string
	Reason    string
	ExpiresAt *time.Time
}

type VisitorRiskStatsSnapshot struct {
	TotalFacts            int64 `json:"total_facts"`
	NormalCount           int64 `json:"normal_count"`
	WatchCount            int64 `json:"watch_count"`
	SuspiciousCount       int64 `json:"suspicious_count"`
	BlockCount            int64 `json:"block_count"`
	RequestCount          int64 `json:"request_count"`
	InvalidRequestCount   int64 `json:"invalid_request_count"`
	AuthFailureCount      int64 `json:"auth_failure_count"`
	CheckoutFailureCount  int64 `json:"checkout_failure_count"`
	BotLikeUserAgentCount int64 `json:"bot_like_user_agent_count"`
	NoCookieRequestCount  int64 `json:"no_cookie_request_count"`
	MeaningfulActionCount int64 `json:"meaningful_action_count"`
}

type VisitorRiskIdentityAssessmentInput struct {
	IPAddress    string
	UserAgent    string
	Now          time.Time
	LookbackDays int
}

type VisitorRiskIdentityAssessment struct {
	Known                 bool
	RiskLevel             string
	RiskScore             int
	RequestCount          int
	CheckoutFailureCount  int
	BotLikeUserAgentCount int
	NoCookieRequestCount  int
	DecisionAction        string
	DecisionReason        string
}

type visitorRiskAccumulator struct {
	Day                   time.Time
	IPHash                string
	UserAgentHash         string
	CountryCode           string
	FirstSeenAt           time.Time
	LastSeenAt            time.Time
	RequestCount          int
	InvalidRequestCount   int
	AuthFailureCount      int
	CheckoutFailureCount  int
	BotLikeUserAgentCount int
	NoCookieRequestCount  int
	MeaningfulActionCount int
	RiskScoreDelta        int
	pathSet               map[string]struct{}
	anonymousSet          map[string]struct{}
	sessionSet            map[string]struct{}
	samplePaths           []string
}

func NewVisitorRiskService(riskRepo *repository.VisitorRiskFactRepository, cfg config.VisitorRiskConfig, fallbackHashSalt string) *VisitorRiskService {
	maxPendingFacts := cfg.MaxPendingFacts
	if maxPendingFacts <= 0 {
		maxPendingFacts = 5000
	}
	samplePathLimit := cfg.SamplePathLimit
	if samplePathLimit <= 0 {
		samplePathLimit = 8
	}
	retentionDays := cfg.RetentionDays
	if retentionDays <= 0 {
		retentionDays = 365
	}
	hashSalt := strings.TrimSpace(cfg.HashSalt)
	if hashSalt == "" {
		hashSalt = strings.TrimSpace(fallbackHashSalt)
	}

	return &VisitorRiskService{
		riskRepo:        riskRepo,
		enabled:         cfg.Enabled,
		hashSalt:        hashSalt,
		maxPendingFacts: maxPendingFacts,
		samplePathLimit: samplePathLimit,
		retentionDays:   retentionDays,
		pending:         map[string]*visitorRiskAccumulator{},
	}
}

func (s *VisitorRiskService) Enabled() bool {
	return s != nil && s.enabled && s.riskRepo != nil
}

func (s *VisitorRiskService) RecordRequest(input VisitorRiskRecordInput) {
	if !s.Enabled() {
		return
	}

	input = normalizeVisitorRiskRecordInput(input)
	if input.IPAddress == "" {
		return
	}

	ipHash := s.hash(input.IPAddress)
	if ipHash == "" {
		return
	}
	uaHash := ""
	if input.UserAgent != "" {
		uaHash = s.hash(input.UserAgent)
	}
	day := visitorRiskDay(input.OccurredAt)
	key := visitorRiskAccumulatorKey(day, ipHash, uaHash)

	s.mu.Lock()
	defer s.mu.Unlock()

	acc := s.pending[key]
	if acc == nil {
		if len(s.pending) >= s.maxPendingFacts {
			return
		}
		acc = &visitorRiskAccumulator{
			Day:           day,
			IPHash:        ipHash,
			UserAgentHash: uaHash,
			CountryCode:   input.CountryCode,
			FirstSeenAt:   input.OccurredAt,
			LastSeenAt:    input.OccurredAt,
			pathSet:       map[string]struct{}{},
			anonymousSet:  map[string]struct{}{},
			sessionSet:    map[string]struct{}{},
		}
		s.pending[key] = acc
	}

	acc.RequestCount++
	if input.CountryCode != "" && acc.CountryCode == "" {
		acc.CountryCode = input.CountryCode
	}
	if input.OccurredAt.Before(acc.FirstSeenAt) {
		acc.FirstSeenAt = input.OccurredAt
	}
	if input.OccurredAt.After(acc.LastSeenAt) {
		acc.LastSeenAt = input.OccurredAt
	}
	if input.Path != "" {
		if _, exists := acc.pathSet[input.Path]; !exists {
			acc.pathSet[input.Path] = struct{}{}
			if len(acc.samplePaths) < s.samplePathLimit {
				acc.samplePaths = append(acc.samplePaths, input.Path)
			}
		}
	}
	if input.AnonymousID != "" {
		acc.anonymousSet[s.hash(input.AnonymousID)] = struct{}{}
	}
	if input.SessionID != "" {
		acc.sessionSet[s.hash(input.SessionID)] = struct{}{}
	}
	if input.StatusCode >= 400 {
		acc.InvalidRequestCount++
	}
	if isVisitorRiskAuthFailure(input.Path, input.StatusCode) {
		acc.AuthFailureCount++
	}
	if isVisitorRiskCheckoutFailure(input.Path, input.StatusCode) {
		acc.CheckoutFailureCount++
	}
	if isVisitorRiskBotLikeUserAgent(input.UserAgent) {
		acc.BotLikeUserAgentCount++
	}
	if !input.HasCookieHeader {
		acc.NoCookieRequestCount++
	}
	if input.MeaningfulAction {
		acc.MeaningfulActionCount++
	}
	acc.RiskScoreDelta += visitorRiskScoreDelta(input)
}

func (s *VisitorRiskService) Flush(ctx context.Context) (VisitorRiskFlushResult, error) {
	result := VisitorRiskFlushResult{}
	if !s.Enabled() {
		return result, nil
	}

	deltas := s.takePendingDeltas()
	if len(deltas) == 0 {
		return result, nil
	}

	for _, delta := range deltas {
		select {
		case <-ctx.Done():
			s.restorePendingDeltas(deltas[result.FlushedFacts:])
			return result, ctx.Err()
		default:
		}
		if err := s.riskRepo.UpsertDelta(delta); err != nil {
			s.restorePendingDeltas(deltas[result.FlushedFacts:])
			return result, err
		}
		result.FlushedFacts++
	}

	return result, nil
}

func (s *VisitorRiskService) CleanupExpiredFacts(now time.Time) (VisitorRiskCleanupResult, error) {
	result := VisitorRiskCleanupResult{}
	if s == nil || s.riskRepo == nil {
		return result, nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	cutoffDay := visitorRiskDay(now.AddDate(0, 0, -s.retentionDays))
	deleted, err := s.riskRepo.DeleteBeforeDay(cutoffDay)
	if err != nil {
		return result, err
	}
	deletedDecisions, err := s.riskRepo.DeleteExpiredDecisionsBefore(now.AddDate(0, 0, -s.retentionDays))
	if err != nil {
		return result, err
	}
	result.DeletedFacts = deleted
	result.DeletedDecisions = deletedDecisions
	result.CutoffDay = cutoffDay
	return result, nil
}

func (s *VisitorRiskService) ListFacts(page, pageSize int, input VisitorRiskFactListInput) ([]VisitorRiskFactSnapshot, int64, error) {
	if s == nil || s.riskRepo == nil {
		return nil, 0, nil
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := repository.VisitorRiskFactListFilters{
		Search:       strings.TrimSpace(input.Search),
		RiskLevel:    normalizeVisitorRiskLevelFilter(input.RiskLevel),
		DayAfter:     visitorRiskDayRangeBoundary(input.DayRange),
		MinRiskScore: parseVisitorRiskMinScore(input.MinRiskScore),
	}
	facts, total, err := s.riskRepo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}

	decisions, err := s.riskRepo.ListActiveDecisionsForFacts(facts, time.Now().UTC())
	if err != nil {
		return nil, 0, err
	}
	decisionByKey := latestVisitorRiskDecisionsByKey(decisions)
	items := make([]VisitorRiskFactSnapshot, 0, len(facts))
	for _, fact := range facts {
		items = append(items, visitorRiskFactSnapshot(fact, activeVisitorRiskDecisionForFact(fact, decisionByKey)))
	}
	return items, total, nil
}

func (s *VisitorRiskService) CreateDecision(factID uint, input VisitorRiskDecisionInput, createdBy uint) (VisitorRiskDecisionSnapshot, error) {
	if s == nil || s.riskRepo == nil {
		return VisitorRiskDecisionSnapshot{}, ErrVisitorRiskFactNotFound
	}
	fact, err := s.riskRepo.FindFactByID(factID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return VisitorRiskDecisionSnapshot{}, ErrVisitorRiskFactNotFound
		}
		return VisitorRiskDecisionSnapshot{}, err
	}

	action := strings.ToLower(strings.TrimSpace(input.Action))
	switch action {
	case visitor.RiskDecisionActionIgnore,
		visitor.RiskDecisionActionWatch,
		visitor.RiskDecisionActionTemporaryBlock,
		visitor.RiskDecisionActionBlockCandidate:
	default:
		return VisitorRiskDecisionSnapshot{}, fmt.Errorf("%w: unsupported action", ErrVisitorRiskDecisionInvalid)
	}

	reason := strings.TrimSpace(input.Reason)
	if reason == "" || utf8.RuneCountInString(reason) > 500 {
		return VisitorRiskDecisionSnapshot{}, fmt.Errorf("%w: reason must be between 1 and 500 characters", ErrVisitorRiskDecisionInvalid)
	}

	now := time.Now().UTC()
	expiresAt := normalizeVisitorRiskDecisionExpiry(input.ExpiresAt)
	if expiresAt != nil && !expiresAt.After(now) {
		return VisitorRiskDecisionSnapshot{}, fmt.Errorf("%w: expiry must be in the future", ErrVisitorRiskDecisionInvalid)
	}
	if action == visitor.RiskDecisionActionTemporaryBlock && expiresAt == nil {
		return VisitorRiskDecisionSnapshot{}, fmt.Errorf("%w: temporary_block requires an expiry", ErrVisitorRiskDecisionInvalid)
	}

	scope, valueHash := visitorRiskDecisionIdentity(fact)
	if scope == "" || valueHash == "" {
		return VisitorRiskDecisionSnapshot{}, ErrVisitorRiskDecisionNoIdentity
	}

	decision := &visitor.RiskDecision{
		Scope:     scope,
		ValueHash: valueHash,
		Action:    action,
		Reason:    reason,
		ExpiresAt: expiresAt,
	}
	if createdBy > 0 {
		decision.CreatedBy = &createdBy
	}
	if err := s.riskRepo.CreateDecision(decision); err != nil {
		return VisitorRiskDecisionSnapshot{}, err
	}
	return visitorRiskDecisionSnapshot(*decision), nil
}

func (s *VisitorRiskService) GetDecision(factID uint) (*VisitorRiskDecisionSnapshot, error) {
	if s == nil || s.riskRepo == nil {
		return nil, ErrVisitorRiskFactNotFound
	}
	fact, err := s.riskRepo.FindFactByID(factID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrVisitorRiskFactNotFound
		}
		return nil, err
	}
	decisions, err := s.riskRepo.ListActiveDecisionsForFacts([]visitor.RiskDailyFact{fact}, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	decisionByKey := latestVisitorRiskDecisionsByKey(decisions)
	decision := activeVisitorRiskDecisionForFact(fact, decisionByKey)
	if decision == nil {
		return nil, nil
	}
	snapshot := visitorRiskDecisionSnapshot(*decision)
	return &snapshot, nil
}

func (s *VisitorRiskService) GetStats(dayRange string) (VisitorRiskStatsSnapshot, error) {
	if s == nil || s.riskRepo == nil {
		return VisitorRiskStatsSnapshot{}, nil
	}

	stats, err := s.riskRepo.Stats(visitorRiskDayRangeBoundary(dayRange))
	if err != nil {
		return VisitorRiskStatsSnapshot{}, err
	}
	return VisitorRiskStatsSnapshot{
		TotalFacts:            stats.TotalFacts,
		NormalCount:           stats.NormalCount,
		WatchCount:            stats.WatchCount,
		SuspiciousCount:       stats.SuspiciousCount,
		BlockCount:            stats.BlockCount,
		RequestCount:          stats.RequestCount,
		InvalidRequestCount:   stats.InvalidRequestCount,
		AuthFailureCount:      stats.AuthFailureCount,
		CheckoutFailureCount:  stats.CheckoutFailureCount,
		BotLikeUserAgentCount: stats.BotLikeUserAgentCount,
		NoCookieRequestCount:  stats.NoCookieRequestCount,
		MeaningfulActionCount: stats.MeaningfulActionCount,
	}, nil
}

func (s *VisitorRiskService) AssessIdentity(ctx context.Context, input VisitorRiskIdentityAssessmentInput) (VisitorRiskIdentityAssessment, error) {
	assessment := VisitorRiskIdentityAssessment{
		RiskLevel: visitor.RiskLevelNormal,
	}
	if !s.Enabled() {
		return assessment, nil
	}
	input.IPAddress = normalizeVisitorRiskIP(input.IPAddress)
	input.UserAgent = strings.TrimSpace(input.UserAgent)
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	} else {
		input.Now = input.Now.UTC()
	}
	if input.LookbackDays <= 0 {
		input.LookbackDays = 30
	}
	if input.IPAddress == "" {
		return assessment, nil
	}

	ipHash := s.hash(input.IPAddress)
	uaHash := ""
	if input.UserAgent != "" {
		uaHash = s.hash(input.UserAgent)
	}
	if ipHash == "" {
		return assessment, nil
	}

	if err := ctx.Err(); err != nil {
		return assessment, err
	}

	dayAfter := visitorRiskDay(input.Now.AddDate(0, 0, -input.LookbackDays))
	fact, err := s.riskRepo.FindLatestByIdentity(dayAfter, ipHash, uaHash)
	if err != nil && !repository.IsRecordNotFound(err) {
		return assessment, err
	}
	if fact != nil {
		assessment.Known = true
		assessment.RiskLevel = strings.TrimSpace(fact.RiskLevel)
		if assessment.RiskLevel == "" {
			assessment.RiskLevel = visitor.RiskLevelNormal
		}
		assessment.RiskScore = fact.RiskScore
		assessment.RequestCount = fact.RequestCount
		assessment.CheckoutFailureCount = fact.CheckoutFailureCount
		assessment.BotLikeUserAgentCount = fact.BotLikeUserAgentCount
		assessment.NoCookieRequestCount = fact.NoCookieRequestCount
	}

	decisions, err := s.riskRepo.ListActiveDecisionsForIdentity(ipHash, uaHash, input.Now)
	if err != nil {
		return assessment, err
	}
	if len(decisions) > 0 {
		assessment.Known = true
		assessment.DecisionAction = strings.TrimSpace(decisions[0].Action)
		assessment.DecisionReason = strings.TrimSpace(decisions[0].Reason)
	}
	return assessment, nil
}

func (s *VisitorRiskService) PendingCount() int {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.pending)
}

func (s *VisitorRiskService) takePendingDeltas() []repository.VisitorRiskFactDelta {
	s.mu.Lock()
	defer s.mu.Unlock()

	deltas := make([]repository.VisitorRiskFactDelta, 0, len(s.pending))
	for _, acc := range s.pending {
		deltas = append(deltas, acc.toDelta(s.samplePathLimit))
	}
	s.pending = map[string]*visitorRiskAccumulator{}
	return deltas
}

func (s *VisitorRiskService) restorePendingDeltas(deltas []repository.VisitorRiskFactDelta) {
	if len(deltas) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, delta := range deltas {
		if len(s.pending) >= s.maxPendingFacts {
			return
		}
		key := visitorRiskAccumulatorKey(delta.Day, delta.IPHash, delta.UserAgentHash)
		acc := &visitorRiskAccumulator{
			Day:                   delta.Day,
			IPHash:                delta.IPHash,
			UserAgentHash:         delta.UserAgentHash,
			CountryCode:           delta.CountryCode,
			FirstSeenAt:           delta.FirstSeenAt,
			LastSeenAt:            delta.LastSeenAt,
			RequestCount:          delta.RequestCount,
			InvalidRequestCount:   delta.InvalidRequestCount,
			AuthFailureCount:      delta.AuthFailureCount,
			CheckoutFailureCount:  delta.CheckoutFailureCount,
			BotLikeUserAgentCount: delta.BotLikeUserAgentCount,
			NoCookieRequestCount:  delta.NoCookieRequestCount,
			MeaningfulActionCount: delta.MeaningfulActionCount,
			RiskScoreDelta:        delta.RiskScoreDelta,
			pathSet:               map[string]struct{}{},
			anonymousSet:          map[string]struct{}{},
			sessionSet:            map[string]struct{}{},
			samplePaths:           append([]string{}, delta.SamplePaths...),
		}
		for _, path := range delta.SamplePaths {
			acc.pathSet[path] = struct{}{}
		}
		s.pending[key] = acc
	}
}

func (acc *visitorRiskAccumulator) toDelta(samplePathLimit int) repository.VisitorRiskFactDelta {
	return repository.VisitorRiskFactDelta{
		Day:                   acc.Day,
		IPHash:                acc.IPHash,
		UserAgentHash:         acc.UserAgentHash,
		CountryCode:           acc.CountryCode,
		FirstSeenAt:           acc.FirstSeenAt,
		LastSeenAt:            acc.LastSeenAt,
		RequestCount:          acc.RequestCount,
		UniquePathCount:       len(acc.pathSet),
		UniqueAnonymousCount:  len(acc.anonymousSet),
		UniqueSessionCount:    len(acc.sessionSet),
		InvalidRequestCount:   acc.InvalidRequestCount,
		AuthFailureCount:      acc.AuthFailureCount,
		CheckoutFailureCount:  acc.CheckoutFailureCount,
		BotLikeUserAgentCount: acc.BotLikeUserAgentCount,
		NoCookieRequestCount:  acc.NoCookieRequestCount,
		MeaningfulActionCount: acc.MeaningfulActionCount,
		RiskScoreDelta:        acc.RiskScoreDelta,
		SamplePaths:           append([]string{}, acc.samplePaths...),
		SamplePathLimit:       samplePathLimit,
	}
}

func normalizeVisitorRiskRecordInput(input VisitorRiskRecordInput) VisitorRiskRecordInput {
	input.IPAddress = normalizeVisitorRiskIP(input.IPAddress)
	input.UserAgent = strings.TrimSpace(input.UserAgent)
	input.Path = normalizeVisitorRiskPath(input.Path)
	input.CountryCode = strings.ToUpper(strings.TrimSpace(input.CountryCode))
	input.AnonymousID = strings.TrimSpace(input.AnonymousID)
	input.SessionID = strings.TrimSpace(input.SessionID)
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now().UTC()
	} else {
		input.OccurredAt = input.OccurredAt.UTC()
	}
	return input
}

func normalizeVisitorRiskIP(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	first, _, found := strings.Cut(value, ",")
	if found {
		value = strings.TrimSpace(first)
	}
	if parsed := net.ParseIP(value); parsed != nil {
		return parsed.String()
	}
	return ""
}

func normalizeVisitorRiskPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if idx := strings.Index(value, "?"); idx >= 0 {
		value = value[:idx]
	}
	if len(value) > 180 {
		value = value[:180]
	}
	return value
}

func visitorRiskDay(value time.Time) time.Time {
	if value.IsZero() {
		value = time.Now().UTC()
	} else {
		value = value.UTC()
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func visitorRiskAccumulatorKey(day time.Time, ipHash string, userAgentHash string) string {
	return fmt.Sprintf("%s|%s|%s", day.Format("2006-01-02"), ipHash, userAgentHash)
}

func (s *VisitorRiskService) hash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(s.hashSalt + ":" + value))
	return hex.EncodeToString(sum[:])
}

func visitorRiskScoreDelta(input VisitorRiskRecordInput) int {
	score := 0
	if input.StatusCode >= 400 {
		score += 3
	}
	if isVisitorRiskAuthFailure(input.Path, input.StatusCode) {
		score += 5
	}
	if isVisitorRiskCheckoutFailure(input.Path, input.StatusCode) {
		score += 8
	}
	if isVisitorRiskBotLikeUserAgent(input.UserAgent) {
		score += 3
	}
	if input.MeaningfulAction {
		score -= 2
	}
	return score
}

func isVisitorRiskAuthFailure(path string, statusCode int) bool {
	if statusCode != 401 && statusCode != 403 {
		return false
	}
	path = strings.ToLower(strings.TrimSpace(path))
	return strings.Contains(path, "/auth") ||
		strings.Contains(path, "/admin") ||
		strings.Contains(path, "/user") ||
		strings.Contains(path, "/orders")
}

func isVisitorRiskCheckoutFailure(path string, statusCode int) bool {
	if statusCode < 400 {
		return false
	}
	path = strings.ToLower(strings.TrimSpace(path))
	return strings.Contains(path, "/checkout") ||
		strings.Contains(path, "/orders") ||
		strings.Contains(path, "/payment")
}

func isVisitorRiskBotLikeUserAgent(userAgent string) bool {
	userAgent = strings.ToLower(strings.TrimSpace(userAgent))
	if userAgent == "" {
		return false
	}
	patterns := []string{
		"bot",
		"crawler",
		"spider",
		"slurp",
		"ahrefs",
		"semrush",
		"python-requests",
		"curl/",
		"wget/",
		"headless",
	}
	for _, pattern := range patterns {
		if strings.Contains(userAgent, pattern) {
			return true
		}
	}
	return false
}

func normalizeVisitorRiskLevelFilter(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "all":
		return "all"
	case visitor.RiskLevelNormal, visitor.RiskLevelWatch, visitor.RiskLevelSuspicious, visitor.RiskLevelBlock:
		return value
	default:
		return "all"
	}
}

func visitorRiskDayRangeBoundary(value string) *time.Time {
	now := time.Now().UTC()
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "24h", "today":
		boundary := now.Add(-24 * time.Hour)
		return &boundary
	case "7d", "week":
		boundary := now.AddDate(0, 0, -7)
		return &boundary
	case "30d", "month":
		boundary := now.AddDate(0, 0, -30)
		return &boundary
	case "90d", "quarter":
		boundary := now.AddDate(0, 0, -90)
		return &boundary
	default:
		return nil
	}
}

func parseVisitorRiskMinScore(value string) *int {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	score, err := strconv.Atoi(value)
	if err != nil || score < 0 {
		return nil
	}
	return &score
}

func visitorRiskFactSnapshot(fact visitor.RiskDailyFact, decision *visitor.RiskDecision) VisitorRiskFactSnapshot {
	return VisitorRiskFactSnapshot{
		ID:                    fact.ID,
		Day:                   fact.Day,
		IPHashPreview:         maskVisitorRiskHash(fact.IPHash),
		UserAgentHashPreview:  maskVisitorRiskHash(fact.UserAgentHash),
		CountryCode:           strings.TrimSpace(fact.CountryCode),
		FirstSeenAt:           fact.FirstSeenAt,
		LastSeenAt:            fact.LastSeenAt,
		RequestCount:          fact.RequestCount,
		UniquePathCount:       fact.UniquePathCount,
		UniqueAnonymousCount:  fact.UniqueAnonymousCount,
		UniqueSessionCount:    fact.UniqueSessionCount,
		InvalidRequestCount:   fact.InvalidRequestCount,
		AuthFailureCount:      fact.AuthFailureCount,
		CheckoutFailureCount:  fact.CheckoutFailureCount,
		BotLikeUserAgentCount: fact.BotLikeUserAgentCount,
		NoCookieRequestCount:  fact.NoCookieRequestCount,
		MeaningfulActionCount: fact.MeaningfulActionCount,
		RiskScore:             fact.RiskScore,
		RiskLevel:             strings.TrimSpace(fact.RiskLevel),
		SamplePaths:           decodeVisitorRiskSamplePaths(fact.SamplePaths),
		Decision:              visitorRiskDecisionSnapshotPtr(decision),
		CreatedAt:             fact.CreatedAt,
		UpdatedAt:             fact.UpdatedAt,
	}
}

func visitorRiskDecisionIdentity(fact visitor.RiskDailyFact) (string, string) {
	if valueHash := visitor.RiskDecisionIPUAValueHash(fact.IPHash, fact.UserAgentHash); valueHash != "" {
		return visitor.RiskDecisionScopeIPUAHash, valueHash
	}
	return visitor.RiskDecisionScopeIPHash, strings.TrimSpace(fact.IPHash)
}

func latestVisitorRiskDecisionsByKey(decisions []visitor.RiskDecision) map[string]visitor.RiskDecision {
	result := make(map[string]visitor.RiskDecision, len(decisions))
	for _, decision := range decisions {
		key := visitorRiskDecisionKey(decision.Scope, decision.ValueHash)
		if _, exists := result[key]; !exists {
			result[key] = decision
		}
	}
	return result
}

func activeVisitorRiskDecisionForFact(fact visitor.RiskDailyFact, decisions map[string]visitor.RiskDecision) *visitor.RiskDecision {
	scope, valueHash := visitorRiskDecisionIdentity(fact)
	if scope == "" || valueHash == "" {
		return nil
	}
	decision, exists := decisions[visitorRiskDecisionKey(scope, valueHash)]
	if !exists {
		if scope != visitor.RiskDecisionScopeIPHash {
			decision, exists = decisions[visitorRiskDecisionKey(visitor.RiskDecisionScopeIPHash, fact.IPHash)]
		}
	}
	if !exists {
		return nil
	}
	return &decision
}

func visitorRiskDecisionKey(scope, valueHash string) string {
	return strings.TrimSpace(scope) + "|" + strings.TrimSpace(valueHash)
}

func visitorRiskDecisionSnapshot(decision visitor.RiskDecision) VisitorRiskDecisionSnapshot {
	return VisitorRiskDecisionSnapshot{
		ID:        decision.ID,
		Scope:     strings.TrimSpace(decision.Scope),
		Action:    strings.TrimSpace(decision.Action),
		Reason:    strings.TrimSpace(decision.Reason),
		ExpiresAt: decision.ExpiresAt,
		CreatedBy: decision.CreatedBy,
		CreatedAt: decision.CreatedAt,
		UpdatedAt: decision.UpdatedAt,
	}
}

func visitorRiskDecisionSnapshotPtr(decision *visitor.RiskDecision) *VisitorRiskDecisionSnapshot {
	if decision == nil {
		return nil
	}
	snapshot := visitorRiskDecisionSnapshot(*decision)
	return &snapshot
}

func normalizeVisitorRiskDecisionExpiry(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}

func decodeVisitorRiskSamplePaths(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var paths []string
	if err := json.Unmarshal(raw, &paths); err != nil {
		return []string{}
	}
	return paths
}

func maskVisitorRiskHash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 16 {
		return value
	}
	return value[:8] + "..." + value[len(value)-8:]
}
