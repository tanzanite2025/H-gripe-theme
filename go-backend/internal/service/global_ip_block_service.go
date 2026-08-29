package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/security"
	appLogger "commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/metrics"
	"commerce-platform/internal/repository"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	defaultGlobalIPBlockRefreshInterval          = 5 * time.Second
	globalIPBlockCacheInvalidationChannel        = "commerce-platform:security:global-ip-block:invalidate:v1"
	globalIPBlockCacheInvalidationRetryDelay     = time.Second
	globalIPBlockCacheInvalidationPublishTimeout = 2 * time.Second
)

var (
	ErrIPBlockRuleInvalid      = errors.New("invalid IP block rule")
	ErrIPBlockRuleNotFound     = errors.New("IP block rule not found")
	ErrIPBlockCacheUnavailable = errors.New("IP block cache unavailable")
	ErrIPBlockCacheRefresh     = errors.New("IP block cache refresh failed")
	ErrIPBlockAuditWrite       = errors.New("IP block audit write failed")
)

type GlobalIPBlockService struct {
	ruleRepo                *repository.GlobalIPBlockRuleRepository
	auditRecorderFactory    IPBlockAuditRecorderFactory
	refreshInterval         time.Duration
	deferCacheRefresh       bool
	cacheInvalidationClient redis.UniversalClient

	cacheMu        sync.RWMutex
	refreshMu      sync.Mutex
	mutationMu     sync.Mutex
	listenerMu     sync.Mutex
	listenerCancel context.CancelFunc
	listenerDone   chan struct{}
	listenerPubSub *redis.PubSub
	cacheLoaded    bool
	cacheExpiresAt time.Time
	cacheRetryAt   time.Time
	cacheLastErr   error
	activeRules    []compiledIPBlockRule
}

type IPBlockRuleInput struct {
	CIDR            string
	Source          string
	SourceReference string
	Reason          string
	ExpiresAt       *time.Time
	CreatedBy       uint
}

type IPBlockRuleListInput struct {
	Search string
	Source string
	Status string
}

type IPBlockRuleSnapshot struct {
	ID              uint       `json:"id"`
	CIDR            string     `json:"cidr"`
	Source          string     `json:"source"`
	SourceReference string     `json:"source_reference,omitempty"`
	Reason          string     `json:"reason,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	Enabled         bool       `json:"enabled"`
	Status          string     `json:"status"`
	CreatedBy       *uint      `json:"created_by,omitempty"`
	DisabledBy      *uint      `json:"disabled_by,omitempty"`
	DisabledAt      *time.Time `json:"disabled_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type IPBlockAuditRecorder interface {
	CreateAuditLog(log *audit.AuditLog) error
}

type IPBlockAuditRecorderFactory func(tx *gorm.DB) IPBlockAuditRecorder

type IPBlockAuditLogFactory func(before *IPBlockRuleSnapshot, after IPBlockRuleSnapshot) (*audit.AuditLog, error)
type IPBlockCollectionAuditLogFactory func(before, after []IPBlockRuleSnapshot) (*audit.AuditLog, error)

type compiledIPBlockRule struct {
	rule    security.IPBlockRule
	network *net.IPNet
	prefix  int
}

func NewGlobalIPBlockService(ruleRepo *repository.GlobalIPBlockRuleRepository) *GlobalIPBlockService {
	return &GlobalIPBlockService{
		ruleRepo:        ruleRepo,
		refreshInterval: defaultGlobalIPBlockRefreshInterval,
	}
}

func (s *GlobalIPBlockService) ConfigureAuditRepository(auditRepo *repository.AuditRepository) {
	if s == nil {
		return
	}
	if auditRepo == nil {
		s.auditRecorderFactory = nil
		return
	}
	s.auditRecorderFactory = func(tx *gorm.DB) IPBlockAuditRecorder {
		return auditRepo.WithTx(tx)
	}
}

func (s *GlobalIPBlockService) ConfigureAuditRecorderFactory(factory IPBlockAuditRecorderFactory) {
	if s == nil {
		return
	}
	s.auditRecorderFactory = factory
}

// ConfigureCacheInvalidation wires the shared Redis channel used to make
// other API instances refresh their in-process rule snapshot immediately.
// The database remains the source of truth and the existing TTL is retained
// as a fallback when Redis notifications are delayed or unavailable.
func (s *GlobalIPBlockService) ConfigureCacheInvalidation(client redis.UniversalClient) {
	if s == nil {
		return
	}
	s.listenerMu.Lock()
	s.cacheInvalidationClient = client
	s.listenerMu.Unlock()
}

// StartCacheInvalidationListener starts an idempotent best-effort Redis
// subscriber. Redis connection failures are retried in the background because
// cache invalidation is an acceleration path, not the durable rule store.
func (s *GlobalIPBlockService) StartCacheInvalidationListener(ctx context.Context) {
	if s == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	s.listenerMu.Lock()
	if s.listenerCancel != nil || s.cacheInvalidationClient == nil {
		s.listenerMu.Unlock()
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	client := s.cacheInvalidationClient
	s.listenerCancel = cancel
	s.listenerDone = done
	s.listenerMu.Unlock()

	go s.listenForCacheInvalidations(runCtx, client, done)
}

func (s *GlobalIPBlockService) StopCacheInvalidationListener() {
	if s == nil {
		return
	}

	s.listenerMu.Lock()
	cancel := s.listenerCancel
	done := s.listenerDone
	pubsub := s.listenerPubSub
	s.listenerCancel = nil
	s.listenerDone = nil
	s.listenerPubSub = nil
	s.listenerMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if pubsub != nil {
		_ = pubsub.Close()
	}
	if done != nil {
		<-done
	}
}

func (s *GlobalIPBlockService) Refresh(ctx context.Context) error {
	if s == nil || s.ruleRepo == nil {
		return ErrIPBlockCacheUnavailable
	}
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	return s.refreshWithLock(ctx)
}

func (s *GlobalIPBlockService) refreshWithLock(ctx context.Context) error {
	err := s.refreshLocked(ctx)
	if err != nil {
		s.noteRefreshFailure(err)
		return err
	}
	s.noteRefreshSuccess()
	return nil
}

func (s *GlobalIPBlockService) refreshLocked(ctx context.Context) error {
	if s == nil || s.ruleRepo == nil {
		return ErrIPBlockCacheUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	now := time.Now().UTC()
	rules, err := s.ruleRepo.ListActive(now)
	if err != nil {
		return err
	}

	compiled := make([]compiledIPBlockRule, 0, len(rules))
	var compileErr error
	for _, rule := range rules {
		compiledRule, parseErr := compileIPBlockRule(rule)
		if parseErr != nil {
			if compileErr == nil {
				compileErr = fmt.Errorf("%w: stored CIDR %q cannot be parsed", ErrIPBlockRuleInvalid, rule.CIDR)
			}
			appLogger.Error(
				"global IP block cache skipped invalid active rule",
				zap.Uint("rule_id", rule.ID),
				zap.String("cidr", strings.TrimSpace(rule.CIDR)),
				zap.Error(parseErr),
			)
			continue
		}
		compiled = append(compiled, compiledRule)
	}
	if compileErr != nil {
		return compileErr
	}

	s.cacheMu.Lock()
	s.activeRules = compiled
	s.cacheLoaded = true
	s.cacheExpiresAt = now.Add(s.refreshInterval)
	s.cacheMu.Unlock()
	return nil
}

func (s *GlobalIPBlockService) Block(ctx context.Context, input IPBlockRuleInput) (IPBlockRuleSnapshot, error) {
	_, snapshot, err := s.BlockWithPrevious(ctx, input)
	return snapshot, err
}

func (s *GlobalIPBlockService) BlockWithPrevious(
	ctx context.Context,
	input IPBlockRuleInput,
) (*IPBlockRuleSnapshot, IPBlockRuleSnapshot, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, IPBlockRuleSnapshot{}, ErrIPBlockRuleInvalid
	}
	s.mutationMu.Lock()
	defer s.mutationMu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, IPBlockRuleSnapshot{}, err
	}

	cidr, err := NormalizeIPOrCIDR(input.CIDR)
	if err != nil {
		return nil, IPBlockRuleSnapshot{}, err
	}
	source := normalizeIPBlockSource(input.Source)
	sourceReference := strings.TrimSpace(input.SourceReference)
	if source != security.IPBlockRuleSourceManual && sourceReference == "" {
		return nil, IPBlockRuleSnapshot{}, fmt.Errorf("%w: source reference is required for non-manual rules", ErrIPBlockRuleInvalid)
	}
	if utf8.RuneCountInString(sourceReference) > 160 {
		return nil, IPBlockRuleSnapshot{}, fmt.Errorf("%w: source reference is too long", ErrIPBlockRuleInvalid)
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" || utf8.RuneCountInString(reason) > 500 {
		return nil, IPBlockRuleSnapshot{}, fmt.Errorf("%w: reason must be between 1 and 500 characters", ErrIPBlockRuleInvalid)
	}

	now := time.Now().UTC()
	expiresAt := normalizeIPBlockExpiry(input.ExpiresAt)
	if expiresAt != nil && !expiresAt.After(now) {
		return nil, IPBlockRuleSnapshot{}, fmt.Errorf("%w: expiry must be in the future", ErrIPBlockRuleInvalid)
	}

	desired := security.IPBlockRule{
		CIDR:            cidr,
		Source:          source,
		SourceReference: sourceReference,
		Reason:          reason,
		ExpiresAt:       expiresAt,
		Enabled:         true,
	}
	if input.CreatedBy > 0 {
		createdBy := input.CreatedBy
		desired.CreatedBy = &createdBy
	}

	var beforeRule *security.IPBlockRule
	var rule *security.IPBlockRule
	for attempt := 0; attempt < 2; attempt++ {
		beforeRule, rule, err = s.ruleRepo.UpsertEnabledWithPrevious(ctx, desired)
		if err == nil {
			break
		}
		if !isIPBlockDuplicateError(err) || attempt == 1 {
			return nil, IPBlockRuleSnapshot{}, err
		}
	}
	if rule == nil {
		return nil, IPBlockRuleSnapshot{}, ErrIPBlockRuleNotFound
	}

	snapshot := ipBlockRuleSnapshot(*rule, now)
	var beforeSnapshot *IPBlockRuleSnapshot
	if beforeRule != nil {
		before := ipBlockRuleSnapshot(*beforeRule, now)
		beforeSnapshot = &before
	}
	if s.deferCacheRefresh {
		return beforeSnapshot, snapshot, nil
	}
	s.upsertCachedRule(*rule, now)
	if err := s.refreshAfterMutationAndNotify(ctx); err != nil {
		return beforeSnapshot, snapshot, fmt.Errorf("%w: %v", ErrIPBlockCacheRefresh, err)
	}
	return beforeSnapshot, snapshot, nil
}

func (s *GlobalIPBlockService) Disable(ctx context.Context, id, disabledBy uint) (IPBlockRuleSnapshot, error) {
	_, after, err := s.DisableWithPrevious(ctx, id, disabledBy)
	return after, err
}

func (s *GlobalIPBlockService) DisableWithPrevious(
	ctx context.Context,
	id,
	disabledBy uint,
) (IPBlockRuleSnapshot, IPBlockRuleSnapshot, error) {
	if s == nil || s.ruleRepo == nil {
		return IPBlockRuleSnapshot{}, IPBlockRuleSnapshot{}, ErrIPBlockRuleNotFound
	}
	s.mutationMu.Lock()
	defer s.mutationMu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return IPBlockRuleSnapshot{}, IPBlockRuleSnapshot{}, err
	}

	now := time.Now().UTC()
	beforeRule, afterRule, err := s.ruleRepo.DisableWithPrevious(ctx, id, disabledBy, now)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return IPBlockRuleSnapshot{}, IPBlockRuleSnapshot{}, ErrIPBlockRuleNotFound
		}
		return IPBlockRuleSnapshot{}, IPBlockRuleSnapshot{}, err
	}
	if beforeRule == nil || afterRule == nil {
		s.Invalidate()
		return IPBlockRuleSnapshot{}, IPBlockRuleSnapshot{}, errors.New("disable IP block rule returned incomplete state")
	}

	beforeSnapshot := ipBlockRuleSnapshot(*beforeRule, now)
	afterSnapshot := ipBlockRuleSnapshot(*afterRule, now)
	if s.deferCacheRefresh {
		return beforeSnapshot, afterSnapshot, nil
	}
	if beforeRule.Enabled && !afterRule.Enabled {
		s.removeCachedRules(map[uint]struct{}{id: {}}, now)
		if err := s.refreshAfterMutationAndNotify(ctx); err != nil {
			return beforeSnapshot, afterSnapshot, fmt.Errorf("%w: %v", ErrIPBlockCacheRefresh, err)
		}
	} else {
		// Another service instance may have already disabled the row while
		// this instance still had it in its enforcement cache.
		s.Invalidate()
		s.notifyCacheInvalidation()
	}
	return beforeSnapshot, afterSnapshot, nil
}

func (s *GlobalIPBlockService) DisableBySourceReference(ctx context.Context, source, sourceReference string, disabledBy uint) (IPBlockRuleSnapshot, error) {
	rules, err := s.DisableBySourceReferenceAll(ctx, source, sourceReference, disabledBy)
	if err != nil {
		return IPBlockRuleSnapshot{}, err
	}
	if len(rules) == 0 {
		return IPBlockRuleSnapshot{}, ErrIPBlockRuleNotFound
	}
	return rules[0], nil
}

func (s *GlobalIPBlockService) DisableBySourceReferenceAll(
	ctx context.Context,
	source,
	sourceReference string,
	disabledBy uint,
) ([]IPBlockRuleSnapshot, error) {
	_, after, err := s.DisableBySourceReferenceAllWithPrevious(ctx, source, sourceReference, disabledBy)
	return after, err
}

func (s *GlobalIPBlockService) DisableBySourceReferenceAllWithPrevious(
	ctx context.Context,
	source,
	sourceReference string,
	disabledBy uint,
) ([]IPBlockRuleSnapshot, []IPBlockRuleSnapshot, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, nil, ErrIPBlockRuleNotFound
	}
	s.mutationMu.Lock()
	defer s.mutationMu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	normalizedSource := normalizeIPBlockSource(source)
	normalizedReference := strings.TrimSpace(sourceReference)
	now := time.Now().UTC()
	beforeRules, afterRules, err := s.ruleRepo.DisableBySourceReferenceWithPrevious(
		ctx,
		normalizedSource,
		normalizedReference,
		disabledBy,
		now,
		now,
	)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			s.Invalidate()
			return nil, nil, ErrIPBlockRuleNotFound
		}
		return nil, nil, err
	}
	if len(afterRules) == 0 {
		s.Invalidate()
		return nil, nil, ErrIPBlockRuleNotFound
	}
	if len(beforeRules) != len(afterRules) {
		s.Invalidate()
		return nil, nil, errors.New("disable IP block rules returned incomplete state")
	}

	ids := make(map[uint]struct{}, len(afterRules))
	for _, rule := range afterRules {
		ids[rule.ID] = struct{}{}
	}
	s.removeCachedRules(ids, now)

	beforeSnapshots := make([]IPBlockRuleSnapshot, 0, len(beforeRules))
	for _, rule := range beforeRules {
		beforeSnapshots = append(beforeSnapshots, ipBlockRuleSnapshot(rule, now))
	}
	afterSnapshots := make([]IPBlockRuleSnapshot, 0, len(afterRules))
	for _, rule := range afterRules {
		afterSnapshots = append(afterSnapshots, ipBlockRuleSnapshot(rule, now))
	}
	if s.deferCacheRefresh {
		return beforeSnapshots, afterSnapshots, nil
	}
	if err := s.refreshAfterMutationAndNotify(ctx); err != nil {
		return beforeSnapshots, afterSnapshots, fmt.Errorf("%w: %v", ErrIPBlockCacheRefresh, err)
	}
	return beforeSnapshots, afterSnapshots, nil
}

func (s *GlobalIPBlockService) BlockWithPreviousAndAudit(
	ctx context.Context,
	input IPBlockRuleInput,
	auditFactory IPBlockAuditLogFactory,
) (*IPBlockRuleSnapshot, IPBlockRuleSnapshot, error) {
	var before *IPBlockRuleSnapshot
	var after IPBlockRuleSnapshot
	err := s.runAuditedMutation(ctx, func(txService *GlobalIPBlockService, auditRecorder IPBlockAuditRecorder) error {
		txBefore, txAfter, err := txService.BlockWithPrevious(ctx, input)
		if err != nil {
			return err
		}
		before = cloneIPBlockRuleSnapshot(txBefore)
		after = txAfter
		return writeIPBlockAuditLog(auditRecorder, auditFactory, before, after)
	})
	return before, after, err
}

func (s *GlobalIPBlockService) DisableWithPreviousAndAudit(
	ctx context.Context,
	id,
	disabledBy uint,
	auditFactory IPBlockAuditLogFactory,
) (IPBlockRuleSnapshot, IPBlockRuleSnapshot, error) {
	var before IPBlockRuleSnapshot
	var after IPBlockRuleSnapshot
	err := s.runAuditedMutation(ctx, func(txService *GlobalIPBlockService, auditRecorder IPBlockAuditRecorder) error {
		txBefore, txAfter, err := txService.DisableWithPrevious(ctx, id, disabledBy)
		if err != nil {
			return err
		}
		before = txBefore
		after = txAfter
		return writeIPBlockAuditLog(auditRecorder, auditFactory, &before, after)
	})
	return before, after, err
}

func (s *GlobalIPBlockService) DisableBySourceReferenceAllWithPreviousAndAudit(
	ctx context.Context,
	source,
	sourceReference string,
	disabledBy uint,
	auditFactory IPBlockCollectionAuditLogFactory,
) ([]IPBlockRuleSnapshot, []IPBlockRuleSnapshot, error) {
	var before []IPBlockRuleSnapshot
	var after []IPBlockRuleSnapshot
	err := s.runAuditedMutation(ctx, func(txService *GlobalIPBlockService, auditRecorder IPBlockAuditRecorder) error {
		txBefore, txAfter, err := txService.DisableBySourceReferenceAllWithPrevious(ctx, source, sourceReference, disabledBy)
		if err != nil {
			return err
		}
		before = append([]IPBlockRuleSnapshot(nil), txBefore...)
		after = append([]IPBlockRuleSnapshot(nil), txAfter...)
		if auditFactory == nil {
			return ErrAuditLogRequired
		}
		log, err := auditFactory(before, after)
		if err != nil {
			return fmt.Errorf("%w: build audit log: %w", ErrIPBlockAuditWrite, err)
		}
		if log == nil {
			return fmt.Errorf("%w: %w", ErrIPBlockAuditWrite, ErrAuditLogRequired)
		}
		if err := auditRecorder.CreateAuditLog(log); err != nil {
			return fmt.Errorf("%w: %w", ErrIPBlockAuditWrite, err)
		}
		return nil
	})
	return before, after, err
}

func (s *GlobalIPBlockService) runAuditedMutation(
	ctx context.Context,
	fn func(txService *GlobalIPBlockService, auditRecorder IPBlockAuditRecorder) error,
) error {
	if s == nil || s.ruleRepo == nil {
		return ErrIPBlockRuleInvalid
	}
	if s.auditRecorderFactory == nil {
		return ErrAuditRepoUnavailable
	}
	if fn == nil {
		return ErrAuditLogRequired
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mutationMu.Lock()
	defer s.mutationMu.Unlock()

	err := s.ruleRepo.Transaction(ctx, func(tx *gorm.DB, txRuleRepo *repository.GlobalIPBlockRuleRepository) error {
		txService := NewGlobalIPBlockService(txRuleRepo)
		txService.refreshInterval = s.refreshInterval
		txService.deferCacheRefresh = true
		auditRecorder := s.auditRecorderFactory(tx)
		if auditRecorder == nil {
			return ErrAuditRepoUnavailable
		}
		return fn(txService, auditRecorder)
	})
	if err != nil {
		return err
	}

	if err := s.refreshAfterMutationAndNotify(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrIPBlockCacheRefresh, err)
	}
	return nil
}

func writeIPBlockAuditLog(
	auditRecorder IPBlockAuditRecorder,
	auditFactory IPBlockAuditLogFactory,
	before *IPBlockRuleSnapshot,
	after IPBlockRuleSnapshot,
) error {
	if auditRecorder == nil {
		return ErrAuditRepoUnavailable
	}
	if auditFactory == nil {
		return ErrAuditLogRequired
	}
	log, err := auditFactory(before, after)
	if err != nil {
		return fmt.Errorf("%w: build audit log: %w", ErrIPBlockAuditWrite, err)
	}
	if log == nil {
		return fmt.Errorf("%w: %w", ErrIPBlockAuditWrite, ErrAuditLogRequired)
	}
	if err := auditRecorder.CreateAuditLog(log); err != nil {
		return fmt.Errorf("%w: %w", ErrIPBlockAuditWrite, err)
	}
	return nil
}

func cloneIPBlockRuleSnapshot(snapshot *IPBlockRuleSnapshot) *IPBlockRuleSnapshot {
	if snapshot == nil {
		return nil
	}
	clone := *snapshot
	return &clone
}

func (s *GlobalIPBlockService) Get(ctx context.Context, id uint) (IPBlockRuleSnapshot, error) {
	if s == nil || s.ruleRepo == nil {
		return IPBlockRuleSnapshot{}, ErrIPBlockRuleNotFound
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return IPBlockRuleSnapshot{}, err
	}
	rule, err := s.ruleRepo.FindByID(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return IPBlockRuleSnapshot{}, ErrIPBlockRuleNotFound
		}
		return IPBlockRuleSnapshot{}, err
	}
	return ipBlockRuleSnapshot(*rule, time.Now().UTC()), nil
}

func (s *GlobalIPBlockService) GetActiveBySourceReference(
	ctx context.Context,
	source,
	sourceReference string,
) (IPBlockRuleSnapshot, error) {
	if s == nil || s.ruleRepo == nil {
		return IPBlockRuleSnapshot{}, ErrIPBlockRuleNotFound
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return IPBlockRuleSnapshot{}, err
	}
	rule, err := s.ruleRepo.FindActiveBySourceReference(
		normalizeIPBlockSource(source),
		strings.TrimSpace(sourceReference),
		time.Now().UTC(),
	)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return IPBlockRuleSnapshot{}, ErrIPBlockRuleNotFound
		}
		return IPBlockRuleSnapshot{}, err
	}
	return ipBlockRuleSnapshot(*rule, time.Now().UTC()), nil
}

func (s *GlobalIPBlockService) ListActiveBySourceReference(
	ctx context.Context,
	source,
	sourceReference string,
) ([]IPBlockRuleSnapshot, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, ErrIPBlockRuleNotFound
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	rules, err := s.ruleRepo.ListActiveBySourceReference(
		normalizeIPBlockSource(source),
		strings.TrimSpace(sourceReference),
		now,
	)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, ErrIPBlockRuleNotFound
	}

	snapshots := make([]IPBlockRuleSnapshot, 0, len(rules))
	for _, rule := range rules {
		snapshots = append(snapshots, ipBlockRuleSnapshot(rule, now))
	}
	return snapshots, nil
}

func (s *GlobalIPBlockService) List(ctx context.Context, page, pageSize int, input IPBlockRuleListInput) ([]IPBlockRuleSnapshot, int64, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, 0, ErrIPBlockCacheUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	rules, total, err := s.ruleRepo.List(page, pageSize, repository.IPBlockRuleListFilters{
		Search: strings.TrimSpace(input.Search),
		Source: normalizeIPBlockSourceFilter(input.Source),
		Status: normalizeIPBlockStatusFilter(input.Status),
	}, time.Now().UTC())
	if err != nil {
		return nil, 0, err
	}

	now := time.Now().UTC()
	snapshots := make([]IPBlockRuleSnapshot, 0, len(rules))
	for _, rule := range rules {
		snapshots = append(snapshots, ipBlockRuleSnapshot(rule, now))
	}
	return snapshots, total, nil
}

func (s *GlobalIPBlockService) FindMatch(ctx context.Context, ip string, now time.Time) (*IPBlockRuleSnapshot, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, ErrIPBlockCacheUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if err := s.ensureFresh(ctx, now); err != nil {
		return nil, err
	}

	normalizedIP := NormalizeIP(ip)
	if normalizedIP == "" {
		return nil, nil
	}
	parsedIP := net.ParseIP(normalizedIP)
	if parsedIP == nil {
		return nil, nil
	}

	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	var best *compiledIPBlockRule
	for index := range s.activeRules {
		candidate := &s.activeRules[index]
		if candidate.rule.ExpiresAt != nil && !candidate.rule.ExpiresAt.After(now) {
			continue
		}
		ipForNetwork := parsedIP
		if candidate.network.IP.To4() != nil {
			ipForNetwork = parsedIP.To4()
		}
		if ipForNetwork == nil || !candidate.network.Contains(ipForNetwork) {
			continue
		}
		if best == nil ||
			candidate.prefix > best.prefix ||
			(candidate.prefix == best.prefix && ipBlockRuleRecency(candidate.rule).After(ipBlockRuleRecency(best.rule))) ||
			(candidate.prefix == best.prefix &&
				ipBlockRuleRecency(candidate.rule).Equal(ipBlockRuleRecency(best.rule)) &&
				candidate.rule.ID > best.rule.ID) {
			best = candidate
		}
	}
	if best == nil {
		return nil, nil
	}

	snapshot := ipBlockRuleSnapshot(best.rule, now)
	return &snapshot, nil
}

func (s *GlobalIPBlockService) FindActiveBySourceReference(
	ctx context.Context,
	source,
	sourceReference string,
	now time.Time,
) (*IPBlockRuleSnapshot, error) {
	rules, err := s.FindActiveRulesBySourceReference(ctx, source, sourceReference, now)
	if err != nil || len(rules) == 0 {
		return nil, err
	}
	return &rules[0], nil
}

func (s *GlobalIPBlockService) FindActiveRulesBySourceReference(
	ctx context.Context,
	source,
	sourceReference string,
	now time.Time,
) ([]IPBlockRuleSnapshot, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, ErrIPBlockCacheUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if err := s.ensureFresh(ctx, now); err != nil {
		return nil, err
	}

	normalizedSource := normalizeIPBlockSource(source)
	normalizedReference := strings.TrimSpace(sourceReference)
	if normalizedReference == "" {
		return nil, nil
	}

	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	snapshots := make([]IPBlockRuleSnapshot, 0)
	for index := range s.activeRules {
		candidate := &s.activeRules[index]
		if candidate.rule.Source != normalizedSource ||
			candidate.rule.SourceReference != normalizedReference {
			continue
		}
		if candidate.rule.ExpiresAt != nil && !candidate.rule.ExpiresAt.After(now) {
			continue
		}
		snapshots = append(snapshots, ipBlockRuleSnapshot(candidate.rule, now))
	}
	sort.SliceStable(snapshots, func(left, right int) bool {
		leftRecency := ipBlockRuleRecencySnapshot(snapshots[left])
		rightRecency := ipBlockRuleRecencySnapshot(snapshots[right])
		if leftRecency.Equal(rightRecency) {
			return snapshots[left].ID > snapshots[right].ID
		}
		return leftRecency.After(rightRecency)
	})
	return snapshots, nil
}

func ipBlockRuleRecencySnapshot(snapshot IPBlockRuleSnapshot) time.Time {
	if !snapshot.UpdatedAt.IsZero() {
		return snapshot.UpdatedAt
	}
	return snapshot.CreatedAt
}

func (s *GlobalIPBlockService) IsBlocked(ctx context.Context, ip string, now time.Time) (bool, *IPBlockRuleSnapshot, error) {
	match, err := s.FindMatch(ctx, ip, now)
	return match != nil, match, err
}

func (s *GlobalIPBlockService) ensureFresh(ctx context.Context, now time.Time) error {
	if s == nil || s.ruleRepo == nil {
		return ErrIPBlockCacheUnavailable
	}
	s.cacheMu.RLock()
	loaded := s.cacheLoaded
	fresh := loaded && s.cacheLastErr == nil && now.Before(s.cacheExpiresAt)
	retryAt := s.cacheRetryAt
	lastErr := s.cacheLastErr
	s.cacheMu.RUnlock()
	if fresh {
		return nil
	}
	if now.Before(retryAt) {
		return unavailableIPBlockCacheError(lastErr)
	}

	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()

	s.cacheMu.RLock()
	loaded = s.cacheLoaded
	fresh = loaded && s.cacheLastErr == nil && now.Before(s.cacheExpiresAt)
	retryAt = s.cacheRetryAt
	lastErr = s.cacheLastErr
	s.cacheMu.RUnlock()
	if fresh {
		return nil
	}
	if now.Before(retryAt) {
		return unavailableIPBlockCacheError(lastErr)
	}

	if err := s.refreshWithLock(ctx); err != nil {
		s.cacheMu.RLock()
		lastErr = s.cacheLastErr
		s.cacheMu.RUnlock()
		return unavailableIPBlockCacheError(lastErr)
	}
	return nil
}

func (s *GlobalIPBlockService) Invalidate() {
	if s == nil {
		return
	}
	s.cacheMu.Lock()
	s.cacheExpiresAt = time.Time{}
	s.cacheRetryAt = time.Time{}
	s.cacheLastErr = nil
	s.cacheMu.Unlock()
}

func (s *GlobalIPBlockService) refreshAfterMutation(ctx context.Context) error {
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	return s.refreshWithLock(ctx)
}

func (s *GlobalIPBlockService) refreshAfterMutationAndNotify(ctx context.Context) error {
	err := s.refreshAfterMutation(ctx)
	s.notifyCacheInvalidation()
	return err
}

func (s *GlobalIPBlockService) listenForCacheInvalidations(
	ctx context.Context,
	client redis.UniversalClient,
	done chan<- struct{},
) {
	defer close(done)

	for {
		pubsub := client.Subscribe(ctx, globalIPBlockCacheInvalidationChannel)
		s.listenerMu.Lock()
		listenerActive := s.listenerCancel != nil && ctx.Err() == nil
		if listenerActive {
			s.listenerPubSub = pubsub
		}
		s.listenerMu.Unlock()
		if !listenerActive {
			_ = pubsub.Close()
			return
		}

		if _, err := pubsub.Receive(ctx); err != nil {
			_ = pubsub.Close()
			s.listenerMu.Lock()
			if s.listenerPubSub == pubsub {
				s.listenerPubSub = nil
			}
			s.listenerMu.Unlock()
			if ctx.Err() != nil {
				return
			}
			metrics.GlobalIPBlockCacheInvalidations.WithLabelValues("listener", "error").Inc()
			appLogger.Warn("global IP block cache invalidation subscription failed", zap.Error(err))
			if !waitForGlobalIPBlockInvalidationRetry(ctx) {
				return
			}
			continue
		}

		for {
			_, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				_ = pubsub.Close()
				s.listenerMu.Lock()
				if s.listenerPubSub == pubsub {
					s.listenerPubSub = nil
				}
				s.listenerMu.Unlock()
				if ctx.Err() != nil {
					return
				}
				metrics.GlobalIPBlockCacheInvalidations.WithLabelValues("listener", "error").Inc()
				appLogger.Warn("global IP block cache invalidation receive failed", zap.Error(err))
				break
			}
			s.Invalidate()
			metrics.GlobalIPBlockCacheInvalidations.WithLabelValues("receive", "success").Inc()
		}

		s.listenerMu.Lock()
		if s.listenerPubSub == pubsub {
			s.listenerPubSub = nil
		}
		s.listenerMu.Unlock()
		if !waitForGlobalIPBlockInvalidationRetry(ctx) {
			return
		}
	}
}

func waitForGlobalIPBlockInvalidationRetry(ctx context.Context) bool {
	timer := time.NewTimer(globalIPBlockCacheInvalidationRetryDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (s *GlobalIPBlockService) notifyCacheInvalidation() {
	if s == nil {
		return
	}

	s.listenerMu.Lock()
	client := s.cacheInvalidationClient
	s.listenerMu.Unlock()
	if client == nil {
		metrics.GlobalIPBlockCacheInvalidations.WithLabelValues("publish", "not_configured").Inc()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), globalIPBlockCacheInvalidationPublishTimeout)
	defer cancel()
	if err := client.Publish(
		ctx,
		globalIPBlockCacheInvalidationChannel,
		time.Now().UTC().Format(time.RFC3339Nano),
	).Err(); err != nil {
		metrics.GlobalIPBlockCacheInvalidations.WithLabelValues("publish", "error").Inc()
		appLogger.Warn("global IP block cache invalidation publish failed", zap.Error(err))
		return
	}
	metrics.GlobalIPBlockCacheInvalidations.WithLabelValues("publish", "success").Inc()
}

func (s *GlobalIPBlockService) noteRefreshSuccess() {
	s.cacheMu.Lock()
	s.cacheRetryAt = time.Time{}
	s.cacheLastErr = nil
	s.cacheMu.Unlock()
	metrics.GlobalIPBlockCacheRefreshes.WithLabelValues("success").Inc()
}

func (s *GlobalIPBlockService) noteRefreshFailure(err error) {
	retryAfter := s.refreshInterval
	if retryAfter <= 0 {
		retryAfter = time.Second
	}
	retryAt := time.Now().UTC().Add(retryAfter)

	s.cacheMu.Lock()
	s.cacheRetryAt = retryAt
	s.cacheLastErr = err
	if s.cacheLoaded {
		// A stale snapshot is not authoritative for access control. Keep it
		// available for diagnostics, but require a successful refresh before
		// serving a request through the enforcement middleware again.
		s.cacheExpiresAt = time.Time{}
	}
	s.cacheMu.Unlock()

	metrics.GlobalIPBlockCacheRefreshes.WithLabelValues("error").Inc()
	appLogger.Error("global IP block cache refresh failed", zap.Error(err))
}

func ipBlockRuleRecency(rule security.IPBlockRule) time.Time {
	if !rule.UpdatedAt.IsZero() {
		return rule.UpdatedAt
	}
	return rule.CreatedAt
}

func (s *GlobalIPBlockService) upsertCachedRule(rule security.IPBlockRule, now time.Time) {
	compiledRule, err := compileIPBlockRule(rule)
	if err != nil {
		appLogger.Error(
			"global IP block cache skipped mutated rule",
			zap.Uint("rule_id", rule.ID),
			zap.String("cidr", strings.TrimSpace(rule.CIDR)),
			zap.Error(err),
		)
		return
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if !s.cacheLoaded {
		return
	}

	activeRules := make([]compiledIPBlockRule, 0, len(s.activeRules)+1)
	for _, candidate := range s.activeRules {
		if candidate.rule.ID != rule.ID {
			activeRules = append(activeRules, candidate)
		}
	}
	if rule.Enabled && (rule.ExpiresAt == nil || rule.ExpiresAt.After(now)) {
		activeRules = append(activeRules, compiledRule)
	}
	s.activeRules = activeRules
	s.cacheExpiresAt = now.Add(s.refreshInterval)
}

func (s *GlobalIPBlockService) removeCachedRules(ids map[uint]struct{}, now time.Time) {
	if len(ids) == 0 {
		return
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if !s.cacheLoaded {
		return
	}

	activeRules := make([]compiledIPBlockRule, 0, len(s.activeRules))
	for _, candidate := range s.activeRules {
		if _, remove := ids[candidate.rule.ID]; !remove {
			activeRules = append(activeRules, candidate)
		}
	}
	s.activeRules = activeRules
	s.cacheExpiresAt = now.Add(s.refreshInterval)
}

func compileIPBlockRule(rule security.IPBlockRule) (compiledIPBlockRule, error) {
	_, network, err := net.ParseCIDR(strings.TrimSpace(rule.CIDR))
	if err != nil {
		return compiledIPBlockRule{}, fmt.Errorf("%w: stored CIDR %q cannot be parsed", ErrIPBlockRuleInvalid, rule.CIDR)
	}
	ones, _ := network.Mask.Size()
	return compiledIPBlockRule{
		rule:    rule,
		network: network,
		prefix:  ones,
	}, nil
}

func unavailableIPBlockCacheError(cause error) error {
	if cause == nil {
		return ErrIPBlockCacheUnavailable
	}
	return fmt.Errorf("%w: %v", ErrIPBlockCacheUnavailable, cause)
}

func isIPBlockDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "duplicate entry")
}

func NormalizeIP(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if first, _, found := strings.Cut(value, ","); found {
		value = strings.TrimSpace(first)
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	parsed := net.ParseIP(value)
	if parsed == nil {
		return ""
	}
	if ipv4 := parsed.To4(); ipv4 != nil {
		return ipv4.String()
	}
	return parsed.String()
}

func NormalizeIPOrCIDR(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%w: IP or CIDR is required", ErrIPBlockRuleInvalid)
	}
	if !strings.Contains(value, "/") {
		ip := NormalizeIP(value)
		if ip == "" {
			return "", fmt.Errorf("%w: %q is not a valid IP address", ErrIPBlockRuleInvalid, value)
		}
		parsed := net.ParseIP(ip)
		bits := 128
		if parsed.To4() != nil {
			bits = 32
		}
		return fmt.Sprintf("%s/%d", ip, bits), nil
	}

	_, network, err := net.ParseCIDR(value)
	if err != nil {
		return "", fmt.Errorf("%w: %q is not a valid CIDR", ErrIPBlockRuleInvalid, value)
	}
	if ipv4 := network.IP.To4(); ipv4 != nil {
		ones, _ := network.Mask.Size()
		return fmt.Sprintf("%s/%d", ipv4.String(), ones), nil
	}
	return network.String(), nil
}

func normalizeIPBlockSource(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return security.IPBlockRuleSourceManual
	}
	return value
}

func normalizeIPBlockSourceFilter(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "all" {
		return ""
	}
	return value
}

func normalizeIPBlockStatusFilter(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", security.IPBlockRuleStatusActive:
		return security.IPBlockRuleStatusActive
	case security.IPBlockRuleStatusExpired, security.IPBlockRuleStatusDisabled, "all":
		return value
	default:
		return security.IPBlockRuleStatusActive
	}
}

func normalizeIPBlockExpiry(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}

func ipBlockRuleSnapshot(rule security.IPBlockRule, now time.Time) IPBlockRuleSnapshot {
	status := security.IPBlockRuleStatusDisabled
	if rule.Enabled {
		status = security.IPBlockRuleStatusActive
		if rule.ExpiresAt != nil && !rule.ExpiresAt.After(now) {
			status = security.IPBlockRuleStatusExpired
		}
	}
	return IPBlockRuleSnapshot{
		ID:              rule.ID,
		CIDR:            strings.TrimSpace(rule.CIDR),
		Source:          strings.TrimSpace(rule.Source),
		SourceReference: strings.TrimSpace(rule.SourceReference),
		Reason:          strings.TrimSpace(rule.Reason),
		ExpiresAt:       rule.ExpiresAt,
		Enabled:         rule.Enabled,
		Status:          status,
		CreatedBy:       rule.CreatedBy,
		DisabledBy:      rule.DisabledBy,
		DisabledAt:      rule.DisabledAt,
		CreatedAt:       rule.CreatedAt,
		UpdatedAt:       rule.UpdatedAt,
	}
}

func profileIPBlockSourceReference(profileID uint) string {
	return strconv.FormatUint(uint64(profileID), 10)
}
