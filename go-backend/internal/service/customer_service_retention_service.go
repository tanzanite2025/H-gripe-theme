package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/metrics"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"gorm.io/datatypes"
)

var (
	ErrCustomerServiceRetentionAdminRequired  = errors.New("customer-service retention requires administrator access")
	ErrCustomerServiceRetentionReasonRequired = errors.New("customer-service retention reason is required")
	ErrCustomerServiceRetentionIneligible     = errors.New("customer-service conversation is not retention eligible")
	ErrCustomerServiceRetentionWindow         = errors.New("customer-service conversation is still inside the recovery window")
	ErrCustomerServiceRetentionIDsRequired    = errors.New("at least one customer-service conversation is required")
	ErrCustomerServiceRetentionTooManyIDs     = errors.New("customer-service retention is limited to 100 conversations")
)

const (
	DefaultCustomerServiceRetentionMonths = 24
	CustomerServiceRetentionRecoveryDays  = 30
)

type CustomerServiceRetentionRuntimeConfig struct {
	Enabled              bool `json:"enabled"`
	IntervalSeconds      int  `json:"interval_seconds"`
	MinimumRetentionDays int  `json:"minimum_retention_days"`
	RecoveryWindowDays   int  `json:"recovery_window_days"`
	BatchLimit           int  `json:"batch_limit"`
}

type CustomerServiceRetentionEligibility struct {
	TicketID        uint       `json:"ticket_id"`
	Eligible        bool       `json:"eligible"`
	Action          string     `json:"action"`
	Reason          string     `json:"reason,omitempty"`
	LifecycleStatus string     `json:"lifecycle_status"`
	LastActivityAt  time.Time  `json:"last_activity_at"`
	EligibleAfter   time.Time  `json:"eligible_after"`
	SoftDeletedAt   *time.Time `json:"soft_deleted_at,omitempty"`
	PurgeAfter      *time.Time `json:"purge_after,omitempty"`
}

// CustomerServiceRetentionBatchResult describes one bounded worker pass. A
// held row is intentionally left soft-deleted for a later retry; an error is
// also non-fatal to the rest of the batch.
type CustomerServiceRetentionBatchResult struct {
	Scanned int `json:"scanned"`
	Purged  int `json:"purged"`
	Blocked int `json:"blocked"`
	Errors  int `json:"errors"`
}

// CustomerServiceConversationSearchIndex is deliberately optional. The
// current deployment does not require a separate search service, but a
// deployment that projects conversations into one must remove the projection
// before the ticket-owned rows are physically deleted.
type CustomerServiceConversationSearchIndex interface {
	DeleteConversation(ctx context.Context, ticketID uint) error
}

type CustomerServiceRetentionService struct {
	ticketRepo        *repository.TicketRepository
	orderRepo         *repository.OrderRepository
	afterSalesRepo    *repository.AfterSalesCaseRepository
	paymentRepo       *repository.PaymentRepository
	evidenceRepo      *repository.OrderEvidenceRepository
	auditService      *AuditService
	cleanupOutbox     *repository.OutboxRepository
	attachmentStorage storage.StorageService
	mediaService      *MediaService
	searchIndex       CustomerServiceConversationSearchIndex
	cdnPurger         *mediaCDNPurger
	minimumAge        time.Duration
	recoveryWindow    time.Duration
	now               func() time.Time
	policyRepo        *repository.CustomerServiceRetentionPolicyRepository
	legacyEnabled     bool
}

func (s *CustomerServiceRetentionService) ConfigurePolicyRepository(repo *repository.CustomerServiceRetentionPolicyRepository) {
	if s != nil {
		s.policyRepo = repo
	}
}

func (s *CustomerServiceRetentionService) ConfigureLegacyWorker(enabled bool) {
	if s != nil {
		s.legacyEnabled = enabled
	}
}

func (s *CustomerServiceRetentionService) RuntimeConfig() (CustomerServiceRetentionRuntimeConfig, error) {
	if s == nil {
		return CustomerServiceRetentionRuntimeConfig{}, errors.New("customer-service retention service is not configured")
	}
	cfg := CustomerServiceRetentionRuntimeConfig{
		Enabled:              s.legacyEnabled,
		IntervalSeconds:      86400,
		MinimumRetentionDays: int(s.minimumAge / (24 * time.Hour)),
		RecoveryWindowDays:   int(s.recoveryWindow / (24 * time.Hour)),
		BatchLimit:           100,
	}
	if s.policyRepo == nil {
		return cfg, nil
	}
	policy, err := s.policyRepo.Get()
	if err != nil && !repository.IsRecordNotFound(err) {
		return cfg, err
	}
	if policy != nil {
		cfg.Enabled = policy.Enabled
		cfg.IntervalSeconds = policy.IntervalSeconds
		cfg.MinimumRetentionDays = policy.MinimumRetentionDays
		cfg.RecoveryWindowDays = policy.RecoveryWindowDays
		cfg.BatchLimit = policy.BatchLimit
	}
	if cfg.MinimumRetentionDays > 0 && cfg.RecoveryWindowDays > 0 {
		s.ConfigureWindows(time.Duration(cfg.MinimumRetentionDays)*24*time.Hour, time.Duration(cfg.RecoveryWindowDays)*24*time.Hour)
	}
	return cfg, nil
}

func (s *CustomerServiceRetentionService) UpdateRuntimeConfig(cfg CustomerServiceRetentionRuntimeConfig) error {
	if s == nil || s.policyRepo == nil {
		return errors.New("customer-service retention policy repository is not configured")
	}
	if cfg.IntervalSeconds < 60 || cfg.IntervalSeconds > 604800 || cfg.MinimumRetentionDays < 1 || cfg.MinimumRetentionDays > 36500 || cfg.RecoveryWindowDays < 1 || cfg.RecoveryWindowDays > 3650 || cfg.BatchLimit < 1 || cfg.BatchLimit > 1000 {
		return errors.New("invalid customer-service retention settings")
	}
	policy := &ticket.CustomerServiceRetentionPolicy{Enabled: cfg.Enabled, IntervalSeconds: cfg.IntervalSeconds, MinimumRetentionDays: cfg.MinimumRetentionDays, RecoveryWindowDays: cfg.RecoveryWindowDays, BatchLimit: cfg.BatchLimit}
	if existing, err := s.policyRepo.Get(); err == nil {
		policy.ID = existing.ID
		policy.CreatedAt = existing.CreatedAt
	} else if !repository.IsRecordNotFound(err) {
		return err
	}
	if err := s.policyRepo.Save(policy); err != nil {
		return err
	}
	s.ConfigureWindows(time.Duration(cfg.MinimumRetentionDays)*24*time.Hour, time.Duration(cfg.RecoveryWindowDays)*24*time.Hour)
	return nil
}

// ConfigureAttachmentStorage enables cleanup of first-party objects referenced
// by ticket messages. It is optional to preserve compatibility with tests and
// installations that store attachments outside the application bucket.
func (s *CustomerServiceRetentionService) ConfigureAttachmentStorage(storageSvc storage.StorageService) {
	if s == nil {
		return
	}
	s.attachmentStorage = storageSvc
}

func (s *CustomerServiceRetentionService) ConfigureMediaService(mediaService *MediaService) {
	if s == nil {
		return
	}
	s.mediaService = mediaService
}

func (s *CustomerServiceRetentionService) ConfigureSearchIndex(index CustomerServiceConversationSearchIndex) {
	if s == nil {
		return
	}
	s.searchIndex = index
}

// ConfigureCleanupOutbox enables durable post-commit cleanup for attachment,
// CDN, and search-index projections. Without it, the compatibility path keeps
// performing the cleanup synchronously before the SQL purge.
func (s *CustomerServiceRetentionService) ConfigureCleanupOutbox(outboxRepo *repository.OutboxRepository) {
	if s == nil {
		return
	}
	s.cleanupOutbox = outboxRepo
}

func NewCustomerServiceRetentionService(
	ticketRepo *repository.TicketRepository,
	orderRepo *repository.OrderRepository,
	afterSalesRepo *repository.AfterSalesCaseRepository,
	paymentRepo *repository.PaymentRepository,
	evidenceRepo *repository.OrderEvidenceRepository,
	auditService *AuditService,
) *CustomerServiceRetentionService {
	return &CustomerServiceRetentionService{
		ticketRepo:     ticketRepo,
		orderRepo:      orderRepo,
		afterSalesRepo: afterSalesRepo,
		paymentRepo:    paymentRepo,
		evidenceRepo:   evidenceRepo,
		auditService:   auditService,
		cdnPurger:      newMediaCDNPurgerFromEnv(),
		minimumAge:     time.Duration(DefaultCustomerServiceRetentionMonths) * 30 * 24 * time.Hour,
		recoveryWindow: time.Duration(CustomerServiceRetentionRecoveryDays) * 24 * time.Hour,
		now:            func() time.Time { return time.Now().UTC() },
	}
}

func (s *CustomerServiceRetentionService) ConfigureWindows(minimumAge, recoveryWindow time.Duration) {
	if s == nil {
		return
	}
	if minimumAge > 0 {
		s.minimumAge = minimumAge
	}
	if recoveryWindow > 0 {
		s.recoveryWindow = recoveryWindow
	}
}

// PurgeEligibleBatch is used only by the scheduled system worker. It never
// soft-deletes a live conversation and always re-runs the dependency checks
// before deleting ticket-owned records, attachments, and optional search
// projections. User-triggered purge remains exposed through Purge and still
// requires an administrator actor.
func (s *CustomerServiceRetentionService) PurgeEligibleBatch(limit int) (CustomerServiceRetentionBatchResult, error) {
	var result CustomerServiceRetentionBatchResult
	if s == nil || s.ticketRepo == nil {
		return result, errors.New("customer-service retention service is not configured")
	}
	if s.auditService == nil {
		return result, errors.New("retention audit service is not configured")
	}
	if cfg, err := s.RuntimeConfig(); err != nil {
		return result, err
	} else if !cfg.Enabled {
		return result, nil
	}
	if limit <= 0 {
		return result, nil
	}
	now := s.currentTime()
	records, err := s.ticketRepo.FindCustomerServiceRetentionCandidates(now.Add(-s.recoveryWindow), limit)
	if err != nil {
		metrics.CustomerServiceRetentionEligibility.WithLabelValues("error").Inc()
		return result, err
	}
	result.Scanned = len(records)
	for _, record := range records {
		item := s.evaluateRecord(record, now)
		if !item.Eligible || item.Action != "purge" {
			result.Blocked++
			metrics.CustomerServiceRetentionEligibility.WithLabelValues("blocked").Inc()
			continue
		}
		if err := s.purgeConversationWithAudit(item.TicketID, s.newAuditEntry(0, item.TicketID, "customer_service_retention_purge", "scheduled_retention_worker", "success")); err != nil {
			result.Errors++
			metrics.CustomerServiceRetentionEligibility.WithLabelValues("error").Inc()
			_ = s.writeAudit(0, item.TicketID, "customer_service_retention_purge", "scheduled_retention_worker", "failed")
			continue
		}
		result.Purged++
		metrics.CustomerServiceRetentionPurged.Inc()
	}
	return result, nil
}

func (s *CustomerServiceRetentionService) Evaluate(ids []uint) ([]CustomerServiceRetentionEligibility, error) {
	if s == nil || s.ticketRepo == nil {
		return nil, errors.New("customer-service retention service is not configured")
	}
	ids = normalizeCustomerServiceRetentionIDs(ids)
	if len(ids) == 0 {
		return nil, ErrCustomerServiceRetentionIDsRequired
	}
	if len(ids) > 100 {
		return nil, ErrCustomerServiceRetentionTooManyIDs
	}
	records, err := s.ticketRepo.FindCustomerServiceRetentionTickets(ids)
	if err != nil {
		return nil, err
	}
	now := s.currentTime()
	byID := make(map[uint]ticket.Ticket, len(records))
	for _, record := range records {
		byID[record.ID] = record
	}
	result := make([]CustomerServiceRetentionEligibility, 0, len(ids))
	for _, id := range ids {
		record, found := byID[id]
		if !found {
			result = append(result, CustomerServiceRetentionEligibility{
				TicketID: id, Action: "ineligible",
				Reason: "conversation was not found or is not a customer-service conversation",
			})
			continue
		}
		result = append(result, s.evaluateRecord(record, now))
	}
	return result, nil
}

func (s *CustomerServiceRetentionService) SoftDelete(ids []uint, actorID uint, reason string) ([]CustomerServiceRetentionEligibility, error) {
	if actorID == 0 {
		return nil, ErrCustomerServiceRetentionAdminRequired
	}
	if strings.TrimSpace(reason) == "" {
		return nil, ErrCustomerServiceRetentionReasonRequired
	}
	ids = normalizeCustomerServiceRetentionIDs(ids)
	if len(ids) == 0 {
		return nil, ErrCustomerServiceRetentionIDsRequired
	}
	if len(ids) > 100 {
		return nil, ErrCustomerServiceRetentionTooManyIDs
	}
	if s == nil || s.auditService == nil {
		return nil, errors.New("retention audit service is not configured")
	}
	eligibility, err := s.Evaluate(ids)
	if err != nil {
		return nil, err
	}
	now := s.currentTime()
	for _, item := range eligibility {
		if !item.Eligible || item.Action != "soft_delete" {
			return eligibility, fmt.Errorf("%w: ticket %d: %s", ErrCustomerServiceRetentionIneligible, item.TicketID, item.Reason)
		}
	}
	updated := make([]CustomerServiceRetentionEligibility, 0, len(eligibility))
	for _, item := range eligibility {
		if err := s.ticketRepo.SoftDeleteCustomerServiceConversation(item.TicketID, now); err != nil {
			_ = s.writeAudit(actorID, item.TicketID, "customer_service_retention_soft_delete", reason, "failed")
			return eligibility, err
		}
		item.SoftDeletedAt = &now
		purgeAfter := now.Add(s.recoveryWindow)
		item.PurgeAfter = &purgeAfter
		item.Action = "soft_deleted"
		updated = append(updated, item)
		metrics.CustomerServiceRetentionSoftDeleted.Inc()
		if err := s.writeAudit(actorID, item.TicketID, "customer_service_retention_soft_delete", reason, "success"); err != nil {
			return eligibility, err
		}
	}
	return updated, nil
}

func (s *CustomerServiceRetentionService) Purge(ids []uint, actorID uint, reason string) ([]CustomerServiceRetentionEligibility, error) {
	if actorID == 0 {
		return nil, ErrCustomerServiceRetentionAdminRequired
	}
	if strings.TrimSpace(reason) == "" {
		return nil, ErrCustomerServiceRetentionReasonRequired
	}
	ids = normalizeCustomerServiceRetentionIDs(ids)
	if len(ids) == 0 {
		return nil, ErrCustomerServiceRetentionIDsRequired
	}
	if len(ids) > 100 {
		return nil, ErrCustomerServiceRetentionTooManyIDs
	}
	if s == nil || s.auditService == nil {
		return nil, errors.New("retention audit service is not configured")
	}
	eligibility, err := s.Evaluate(ids)
	if err != nil {
		return nil, err
	}
	for _, item := range eligibility {
		if !item.Eligible || item.Action != "purge" {
			if item.Action == "soft_delete" || strings.Contains(item.Reason, "recovery window") {
				return eligibility, fmt.Errorf("%w: ticket %d: %s", ErrCustomerServiceRetentionWindow, item.TicketID, item.Reason)
			}
			return eligibility, fmt.Errorf("%w: ticket %d: %s", ErrCustomerServiceRetentionIneligible, item.TicketID, item.Reason)
		}
	}
	for _, item := range eligibility {
		if err := s.purgeConversationWithAudit(item.TicketID, s.newAuditEntry(actorID, item.TicketID, "customer_service_retention_purge", reason, "success")); err != nil {
			_ = s.writeAudit(actorID, item.TicketID, "customer_service_retention_purge", reason, "failed")
			return eligibility, err
		}
		metrics.CustomerServiceRetentionPurged.Inc()
	}
	return eligibility, nil
}

func normalizeCustomerServiceRetentionIDs(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func normalizeCustomerServiceRetentionAttachmentReferences(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func (s *CustomerServiceRetentionService) newCleanupOutboxEvent(ticketID uint, refs []string) (*outbox.Event, error) {
	if ticketID == 0 {
		return nil, errors.New("customer-service retention cleanup ticket id is required")
	}
	refs = normalizeCustomerServiceRetentionAttachmentReferences(refs)
	requestedAt := s.currentTime()
	payload, err := json.Marshal(outbox.CustomerServiceRetentionCleanupPayload{
		TicketID:             ticketID,
		AttachmentReferences: refs,
		RequestedAt:          requestedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("encode customer-service retention cleanup event: %w", err)
	}
	identity := fmt.Sprintf("%d\x00%s", ticketID, strings.Join(refs, "\x00"))
	digest := sha256.Sum256([]byte(identity))
	return &outbox.Event{
		EventKey:      fmt.Sprintf("%s:%x", outbox.EventTypeCustomerServiceRetentionCleanup, digest[:]),
		EventType:     outbox.EventTypeCustomerServiceRetentionCleanup,
		AggregateType: outbox.AggregateTypeCustomerServiceConversation,
		AggregateID:   strconv.FormatUint(uint64(ticketID), 10),
		Payload:       datatypes.JSON(payload),
		MaxAttempts:   1000000,
		AvailableAt:   requestedAt,
	}, nil
}

// purgeConversation performs external projection/object cleanup before the
// ticket-owned hard delete. Object deletion is limited to references accepted
// by the configured storage service, while a configured CDN purger may evict
// both first-party and separately managed CDN references.
func (s *CustomerServiceRetentionService) purgeConversation(ticketID uint) error {
	if err := s.cleanupConversationProjections(context.Background(), ticketID, nil); err != nil {
		return err
	}
	return s.ticketRepo.PurgeCustomerServiceConversation(ticketID)
}

// purgeConversationWithAudit commits the ticket-owned delete and its success
// audit row together. Projection/object cleanup is kept outside the SQL
// transaction because those systems cannot participate in the database
// rollback; their operations are designed to be repeatable.
func (s *CustomerServiceRetentionService) purgeConversationWithAudit(ticketID uint, entry *audit.AuditLog) error {
	if s == nil || s.ticketRepo == nil {
		return errors.New("customer-service retention service is not configured")
	}
	scan, err := s.ticketRepo.FindCustomerServiceConversationAttachmentReferenceScan(ticketID)
	if err != nil {
		return fmt.Errorf("load customer-service attachment references: %w", err)
	}
	if scan.MalformedPayloads > 0 {
		metrics.CustomerServiceRetentionAttachmentReferenceSkips.WithLabelValues("malformed_payload").Add(float64(scan.MalformedPayloads))
	}
	if s.cleanupOutbox != nil {
		cleanupEvent, eventErr := s.newCleanupOutboxEvent(ticketID, scan.References)
		if eventErr != nil {
			return eventErr
		}
		return s.ticketRepo.PurgeCustomerServiceConversationWithAuditAndCleanup(ticketID, entry, cleanupEvent)
	}
	if err := s.cleanupConversationProjections(context.Background(), ticketID, scan.References); err != nil {
		return err
	}
	return s.ticketRepo.PurgeCustomerServiceConversationWithAudit(ticketID, entry)
}

// cleanupConversationProjections performs external projection/object cleanup
// before the ticket-owned hard delete.
func (s *CustomerServiceRetentionService) cleanupConversationProjections(ctx context.Context, ticketID uint, refs []string) error {
	if s == nil || s.ticketRepo == nil {
		return errors.New("customer-service retention service is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if s.searchIndex != nil {
		if err := s.searchIndex.DeleteConversation(ctx, ticketID); err != nil {
			return fmt.Errorf("delete customer-service search index: %w", err)
		}
	}
	if refs == nil {
		scan, err := s.ticketRepo.FindCustomerServiceConversationAttachmentReferenceScan(ticketID)
		if err != nil {
			return fmt.Errorf("load customer-service attachment references: %w", err)
		}
		refs = scan.References
		if scan.MalformedPayloads > 0 {
			metrics.CustomerServiceRetentionAttachmentReferenceSkips.WithLabelValues("malformed_payload").Add(float64(scan.MalformedPayloads))
		}
	}
	if s.attachmentStorage != nil {
		for _, ref := range refs {
			if s.mediaService != nil {
				shared, shareErr := s.mediaService.customerServiceAttachmentReferencedElsewhere(ref, ticketID)
				if shareErr != nil {
					return fmt.Errorf("check customer-service attachment references: %w", shareErr)
				}
				if shared {
					metrics.CustomerServiceRetentionAttachmentReferenceSkips.WithLabelValues("shared_reference").Inc()
					continue
				}
			}
			if _, keyErr := s.attachmentStorage.ObjectKey(ref); keyErr == nil {
				if err := s.attachmentStorage.Delete(ctx, ref); err != nil {
					return fmt.Errorf("delete customer-service attachment: %w", err)
				}
			}
			if s.cdnPurger != nil {
				s.cdnPurger.PurgeAsync(ref)
			}
		}
	} else if s.cdnPurger != nil {
		// A deployment may keep conversation media in a separately managed
		// bucket. The CDN purge is still safe and useful even when the app
		// storage adapter is intentionally not configured here.
		for _, ref := range refs {
			s.cdnPurger.PurgeAsync(ref)
		}
	}
	return nil
}

// HandleCustomerServiceRetentionCleanup is registered with the shared SQL
// outbox dispatcher. The ticket and its messages have already been deleted;
// attachment references therefore come from the durable event payload.
func (s *CustomerServiceRetentionService) HandleCustomerServiceRetentionCleanup(ctx context.Context, event outbox.Event) error {
	if event.EventType != outbox.EventTypeCustomerServiceRetentionCleanup {
		return fmt.Errorf("unsupported customer-service retention event %s", event.EventType)
	}
	var payload outbox.CustomerServiceRetentionCleanupPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode customer-service retention cleanup event: %w", err)
	}
	if payload.TicketID == 0 {
		return errors.New("customer-service retention cleanup ticket id is required")
	}
	return s.cleanupConversationProjections(ctx, payload.TicketID, payload.AttachmentReferences)
}

func (s *CustomerServiceRetentionService) evaluateRecord(record ticket.Ticket, now time.Time) CustomerServiceRetentionEligibility {
	lastActivity := record.UpdatedAt
	if record.ClosedAt != nil && record.ClosedAt.After(lastActivity) {
		lastActivity = *record.ClosedAt
	}
	if record.ResolvedAt != nil && record.ResolvedAt.After(lastActivity) {
		lastActivity = *record.ResolvedAt
	}
	item := CustomerServiceRetentionEligibility{
		TicketID: record.ID, LifecycleStatus: record.Status, LastActivityAt: lastActivity,
		EligibleAfter: lastActivity.Add(s.minimumAge), Action: "ineligible",
	}
	if record.DeletedAt.Valid {
		deletedAt := record.DeletedAt.Time.UTC()
		purgeAfter := deletedAt.Add(s.recoveryWindow)
		item.SoftDeletedAt, item.PurgeAfter = &deletedAt, &purgeAfter
		if now.Before(purgeAfter) {
			item.Reason = "soft-delete recovery window has not elapsed"
			return item
		}
		if hold := s.dependencyHold(record); hold != "" {
			item.Reason = hold
			return item
		}
		item.Eligible, item.Action = true, "purge"
		return item
	}
	if record.Status != "resolved" && record.Status != "closed" {
		item.Reason = "conversation is not resolved or closed"
		return item
	}
	if now.Before(item.EligibleAfter) {
		item.Reason = "minimum retention age has not elapsed"
		return item
	}
	if hold := s.dependencyHold(record); hold != "" {
		item.Reason = hold
		return item
	}
	item.Eligible, item.Action = true, "soft_delete"
	return item
}

func (s *CustomerServiceRetentionService) dependencyHold(record ticket.Ticket) string {
	if s.orderRepo == nil || record.CustomerUserID == nil || *record.CustomerUserID == 0 {
		return ""
	}
	orders, _, err := s.orderRepo.FindByUserID(*record.CustomerUserID, 1, 100)
	if err != nil {
		return "order dependency check failed"
	}
	for _, orderRecord := range orders {
		if s.afterSalesRepo != nil {
			cases, caseErr := s.afterSalesRepo.FindByOrderID(orderRecord.ID, "")
			if caseErr != nil {
				return "after-sales dependency check failed"
			}
			for _, item := range cases {
				if !aftersales.IsTerminalStatus(item.Status) {
					return "active after-sales case hold"
				}
			}
		}
		if s.paymentRepo != nil {
			stripe, stripeErr := s.paymentRepo.HasActiveStripeDisputeByOrderID(orderRecord.ID)
			paypal, paypalErr := s.paymentRepo.HasActivePayPalDisputeByOrderID(orderRecord.ID)
			if stripeErr != nil || paypalErr != nil {
				return "payment dispute dependency check failed"
			}
			if stripe || paypal {
				return "active payment dispute hold"
			}
		}
		if s.evidenceRepo != nil {
			if _, evidenceErr := s.evidenceRepo.FindLatestPackageByOrderID(orderRecord.ID); evidenceErr == nil {
				return "order evidence hold"
			} else if !repository.IsRecordNotFound(evidenceErr) {
				return "order evidence dependency check failed"
			}
		}
	}
	return ""
}

func (s *CustomerServiceRetentionService) newAuditEntry(actorID, ticketID uint, action, reason, status string) *audit.AuditLog {
	if s.auditService == nil {
		return nil
	}
	pathAction := strings.TrimPrefix(action, "customer_service_retention_")
	pathAction = strings.ReplaceAll(pathAction, "_", "-")
	oldValue, newValue := `{"retention_state":"soft_deleted"}`, `{"retention_state":"purged"}`
	if strings.Contains(action, "soft_delete") {
		oldValue, newValue = `{"deleted_at":null}`, `{"deleted_at":"tombstone"}`
	}
	entry := &audit.AuditLog{
		UserID: actorID, Action: "delete", Resource: action, ResourceID: ticketID,
		Method: "POST", Path: "/api/admin/customer-service/conversations/retention/" + pathAction,
		Changes: fmt.Sprintf(`{"reason":%q,"projection":"ticket-owned-records-conversation-attachments-and-search-index"}`, strings.TrimSpace(reason)), Status: status,
		OldValue: oldValue, NewValue: newValue,
	}
	if actorID == 0 {
		entry.Username = "system"
	}
	return entry
}

func (s *CustomerServiceRetentionService) writeAudit(actorID, ticketID uint, action, reason, status string) error {
	if s.auditService == nil {
		return errors.New("retention audit service is not configured")
	}
	entry := s.newAuditEntry(actorID, ticketID, action, reason, status)
	return s.auditService.CreateAuditLog(entry)
}

func (s *CustomerServiceRetentionService) currentTime() time.Time {
	if s != nil && s.now != nil {
		return s.now().UTC()
	}
	return time.Now().UTC()
}
