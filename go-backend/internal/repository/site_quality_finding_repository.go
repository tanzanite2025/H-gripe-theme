package repository

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SiteQualityFindingRepository struct {
	db *gorm.DB
}

type SiteQualityFindingListFilter struct {
	Page        int
	PageSize    int
	State       string
	Severity    string
	RuleID      string
	TargetURL   string
	Strategy    string
	FindingKind string
}

func NewSiteQualityFindingRepository(db *gorm.DB) *SiteQualityFindingRepository {
	return &SiteQualityFindingRepository{db: db}
}

func (r *SiteQualityFindingRepository) WithTx(tx *gorm.DB) *SiteQualityFindingRepository {
	return &SiteQualityFindingRepository{db: tx}
}

func (r *SiteQualityFindingRepository) List(
	filter SiteQualityFindingListFilter,
) ([]sitequalitydomain.SiteQualityFinding, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("SiteQuality finding repository is unavailable")
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 200 {
		filter.PageSize = 50
	}

	query := r.db.Model(&sitequalitydomain.SiteQualityFinding{})
	switch filter.State {
	case "active":
		query = query.Where("state IN ?", []string{
			sitequalitydomain.SiteQualityFindingStateOpen,
			sitequalitydomain.SiteQualityFindingStateAcknowledged,
			sitequalitydomain.SiteQualityFindingStateResolved,
		})
	case "", "all":
	default:
		query = query.Where("state = ?", filter.State)
	}
	if filter.Severity != "" {
		query = query.Where("severity = ?", filter.Severity)
	}
	if ruleID := strings.TrimSpace(filter.RuleID); ruleID != "" {
		query = query.Where("rule_id = ?", sitequalitydomain.RuleIDForAuditID(ruleID))
	}
	if filter.TargetURL != "" {
		query = query.Where("target_url = ?", filter.TargetURL)
	}
	if filter.Strategy != "" {
		query = query.Where("strategy = ?", filter.Strategy)
	}
	if filter.FindingKind != "" {
		query = query.Where("finding_kind = ?", filter.FindingKind)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var findings []sitequalitydomain.SiteQualityFinding
	err := query.
		Order("CASE severity WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END").
		Order("last_detected_at ASC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&findings).Error
	for index := range findings {
		normalizeSiteQualityFindingIdentity(&findings[index])
	}
	return findings, total, err
}

func (r *SiteQualityFindingRepository) Stats() (sitequalitydomain.SiteQualityFindingStats, error) {
	if r == nil || r.db == nil {
		return sitequalitydomain.SiteQualityFindingStats{}, errors.New("SiteQuality finding repository is unavailable")
	}
	var stats sitequalitydomain.SiteQualityFindingStats
	err := r.db.Model(&sitequalitydomain.SiteQualityFinding{}).
		Select(`
			COALESCE(SUM(CASE WHEN state IN ('open', 'acknowledged', 'resolved') THEN 1 ELSE 0 END), 0) AS active,
			COALESCE(SUM(CASE WHEN state = 'open' THEN 1 ELSE 0 END), 0) AS open,
			COALESCE(SUM(CASE WHEN state = 'acknowledged' THEN 1 ELSE 0 END), 0) AS acknowledged,
			COALESCE(SUM(CASE WHEN state = 'resolved' THEN 1 ELSE 0 END), 0) AS resolved,
			COALESCE(SUM(CASE WHEN state = 'verified' THEN 1 ELSE 0 END), 0) AS verified,
			COALESCE(SUM(CASE WHEN severity = 'critical' AND state IN ('open', 'acknowledged', 'resolved') THEN 1 ELSE 0 END), 0) AS critical,
			COALESCE(SUM(CASE WHEN severity = 'high' AND state IN ('open', 'acknowledged', 'resolved') THEN 1 ELSE 0 END), 0) AS high
		`).
		Scan(&stats).Error
	return stats, err
}

func (r *SiteQualityFindingRepository) FindByID(
	id uint,
) (*sitequalitydomain.SiteQualityFinding, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("SiteQuality finding repository is unavailable")
	}
	var finding sitequalitydomain.SiteQualityFinding
	if err := r.db.First(&finding, id).Error; err != nil {
		return nil, err
	}
	normalizeSiteQualityFindingIdentity(&finding)
	return &finding, nil
}

func (r *SiteQualityFindingRepository) ListEvents(
	findingID uint,
	page int,
	pageSize int,
) ([]sitequalitydomain.SiteQualityFindingEvent, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("SiteQuality finding repository is unavailable")
	}
	if findingID == 0 {
		return nil, 0, errors.New("SiteQuality finding ID is required")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	query := r.db.Model(&sitequalitydomain.SiteQualityFindingEvent{}).Where("finding_id = ?", findingID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var events []sitequalitydomain.SiteQualityFindingEvent
	err := query.
		Order("created_at DESC").
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&events).Error
	return events, total, err
}

// RecordDetection updates the current work item while preserving the immutable
// SiteQuality run evidence. A resolved or verified item is reopened when the
// audit appears again.
func (r *SiteQualityFindingRepository) RecordDetection(
	detection sitequalitydomain.SiteQualityFindingDetection,
) (*sitequalitydomain.SiteQualityFinding, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("SiteQuality finding repository is unavailable")
	}
	if detection.TargetURL == "" || detection.Strategy == "" || detection.AuditID == "" || detection.LatestRunID == 0 {
		return nil, errors.New("SiteQuality finding detection is incomplete")
	}
	if detection.DetectedAt.IsZero() {
		detection.DetectedAt = time.Now().UTC()
	}
	detection.DetectedAt = detection.DetectedAt.UTC()
	ruleID := strings.TrimSpace(detection.RuleID)
	if ruleID == "" {
		ruleID = sitequalitydomain.RuleIDForAuditID(detection.AuditID)
	} else {
		ruleID = sitequalitydomain.RuleIDForAuditID(ruleID)
	}
	providerAuditID := strings.TrimSpace(detection.ProviderAuditID)
	if providerAuditID == "" {
		providerAuditID = sitequalitydomain.ProviderAuditIDForAuditID(detection.AuditID)
	}

	var result sitequalitydomain.SiteQualityFinding
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var finding sitequalitydomain.SiteQualityFinding
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("strategy = ? AND audit_id = ?", detection.Strategy, detection.AuditID)
		if detection.TargetID != nil && *detection.TargetID != 0 {
			query = query.Where("target_id = ?", *detection.TargetID)
		} else {
			query = query.Where("target_url = ?", detection.TargetURL)
		}
		err := query.First(&finding).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findingKind := strings.TrimSpace(detection.FindingKind)
			if findingKind == "" {
				findingKind = "opportunity"
			}
			sampleCount := detection.SampleCount
			if sampleCount <= 0 {
				sampleCount = 1
			}
			confirmations := detection.Confirmations
			if confirmations <= 0 {
				confirmations = 1
			}
			finding = sitequalitydomain.SiteQualityFinding{
				TargetID:           detection.TargetID,
				TargetURL:          detection.TargetURL,
				Strategy:           detection.Strategy,
				AuditID:            detection.AuditID,
				RuleID:             ruleID,
				ProviderAuditID:    providerAuditID,
				FindingKind:        findingKind,
				RuleVersion:        strings.TrimSpace(detection.RuleVersion),
				Confidence:         detection.Confidence,
				SampleCount:        sampleCount,
				Confirmations:      confirmations,
				Title:              detection.Title,
				Description:        detection.Description,
				Severity:           detection.Severity,
				State:              sitequalitydomain.SiteQualityFindingStateOpen,
				FirstDetectedAt:    detection.DetectedAt,
				LastDetectedAt:     detection.DetectedAt,
				LatestRunID:        detection.LatestRunID,
				LatestSavingsMS:    detection.LatestSavingsMS,
				LatestSavingsBytes: detection.LatestSavingsBytes,
				ResourceCount:      detection.ResourceCount,
				LatestEvidence:     detection.LatestEvidence,
			}
			if err := tx.Create(&finding).Error; err != nil {
				return err
			}
			runID := detection.LatestRunID
			if err := r.createEvent(
				tx,
				finding.ID,
				&runID,
				sitequalitydomain.SiteQualityFindingEventDetected,
				0,
				"",
				map[string]interface{}{
					"audit_id": detection.AuditID,
					"rule_id":  ruleID,
					"severity": detection.Severity,
				},
				detection.DetectedAt,
			); err != nil {
				return err
			}
			result = finding
			return nil
		}
		if err != nil {
			return err
		}

		updates := map[string]interface{}{
			"target_id":            detection.TargetID,
			"target_url":           detection.TargetURL,
			"rule_id":              ruleID,
			"provider_audit_id":    providerAuditID,
			"title":                detection.Title,
			"description":          detection.Description,
			"severity":             detection.Severity,
			"finding_kind":         defaultSiteQualityFindingKind(detection.FindingKind),
			"rule_version":         strings.TrimSpace(detection.RuleVersion),
			"confidence":           detection.Confidence,
			"sample_count":         maxSiteQualityFindingCount(detection.SampleCount),
			"confirmations":        maxSiteQualityFindingCount(detection.Confirmations),
			"consecutive_clean":    0,
			"last_detected_at":     detection.DetectedAt,
			"latest_run_id":        detection.LatestRunID,
			"latest_savings_ms":    detection.LatestSavingsMS,
			"latest_savings_bytes": detection.LatestSavingsBytes,
			"resource_count":       detection.ResourceCount,
			"latest_evidence":      detection.LatestEvidence,
			"updated_at":           detection.DetectedAt,
		}
		reopened := finding.State == sitequalitydomain.SiteQualityFindingStateResolved ||
			finding.State == sitequalitydomain.SiteQualityFindingStateVerified
		if reopened {
			updates["state"] = sitequalitydomain.SiteQualityFindingStateOpen
			updates["resolution_note"] = ""
			updates["resolved_at"] = nil
			updates["verified_at"] = nil
		}
		if err := tx.Model(&sitequalitydomain.SiteQualityFinding{}).
			Where("id = ?", finding.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		if reopened {
			runID := detection.LatestRunID
			if err := r.createEvent(
				tx,
				finding.ID,
				&runID,
				sitequalitydomain.SiteQualityFindingEventReopened,
				0,
				"",
				map[string]interface{}{
					"audit_id":          detection.AuditID,
					"rule_id":           ruleID,
					"provider_audit_id": providerAuditID,
				},
				detection.DetectedAt,
			); err != nil {
				return err
			}
		}
		return tx.First(&result, finding.ID).Error
	})
	if err != nil {
		return nil, err
	}
	normalizeSiteQualityFindingIdentity(&result)
	return &result, nil
}

func defaultSiteQualityFindingKind(value string) string {
	if normalized := strings.TrimSpace(value); normalized != "" {
		return normalized
	}
	return "opportunity"
}

func maxSiteQualityFindingCount(value int) int {
	if value > 0 {
		return value
	}
	return 1
}

// ApplyEvaluation applies a confirmed multi-sample decision inside the caller's
// transaction. Only a completely absent audit contributes to clean verification;
// partial recurrence resets the clean streak without reopening the work item.
func (r *SiteQualityFindingRepository) ApplyEvaluation(
	tx *gorm.DB,
	input sitequalitydomain.SiteQualityFindingEvaluationInput,
) error {
	if r == nil || r.db == nil {
		return errors.New("SiteQuality finding repository is unavailable")
	}
	if tx == nil {
		return errors.New("SiteQuality finding evaluation transaction is required")
	}
	if input.TargetID == 0 || input.Strategy == "" || input.LatestRunID == 0 {
		return errors.New("SiteQuality finding evaluation input is incomplete")
	}
	if input.DecidedAt.IsZero() {
		input.DecidedAt = time.Now().UTC()
	} else {
		input.DecidedAt = input.DecidedAt.UTC()
	}
	if input.RequiredCleanEvaluations <= 0 {
		input.RequiredCleanEvaluations = 2
	}

	txRepo := r.WithTx(tx)
	for index := range input.Detections {
		detection := input.Detections[index]
		targetID := input.TargetID
		detection.TargetID = &targetID
		detection.TargetURL = input.TargetURL
		detection.Strategy = input.Strategy
		detection.DetectedAt = input.DecidedAt
		if _, err := txRepo.RecordDetection(detection); err != nil {
			return err
		}
	}

	cleanAuditIDs := make(map[string]struct{}, len(input.CleanAuditIDs))
	for _, auditID := range input.CleanAuditIDs {
		if normalized := strings.TrimSpace(auditID); normalized != "" {
			cleanAuditIDs[normalized] = struct{}{}
		}
	}
	observedAuditIDs := make(map[string]struct{}, len(input.ObservedAuditIDs))
	for _, auditID := range input.ObservedAuditIDs {
		if normalized := strings.TrimSpace(auditID); normalized != "" {
			observedAuditIDs[normalized] = struct{}{}
		}
	}

	var resolved []sitequalitydomain.SiteQualityFinding
	resolvedQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			"target_id = ? AND strategy = ? AND state = ?",
			input.TargetID,
			input.Strategy,
			sitequalitydomain.SiteQualityFindingStateResolved,
		)
	if input.FindingID != nil && *input.FindingID != 0 {
		resolvedQuery = resolvedQuery.Where("id = ?", *input.FindingID)
	}
	if err := resolvedQuery.Find(&resolved).Error; err != nil {
		return err
	}
	for _, finding := range resolved {
		if _, confirmed := observedAuditIDs[finding.AuditID]; confirmed {
			if _, clean := cleanAuditIDs[finding.AuditID]; !clean && finding.ConsecutiveClean != 0 {
				if err := tx.Model(&sitequalitydomain.SiteQualityFinding{}).
					Where("id = ?", finding.ID).
					Updates(map[string]interface{}{
						"consecutive_clean": 0,
						"updated_at":        input.DecidedAt,
					}).Error; err != nil {
					return err
				}
			}
			continue
		}
		if _, clean := cleanAuditIDs[finding.AuditID]; !clean {
			continue
		}

		nextClean := finding.ConsecutiveClean + 1
		updates := map[string]interface{}{
			"consecutive_clean": nextClean,
			"updated_at":        input.DecidedAt,
		}
		eventType := ""
		if nextClean >= input.RequiredCleanEvaluations {
			updates["state"] = sitequalitydomain.SiteQualityFindingStateVerified
			updates["verified_at"] = input.DecidedAt
			eventType = sitequalitydomain.SiteQualityFindingEventVerificationPassed
		}
		if err := tx.Model(&sitequalitydomain.SiteQualityFinding{}).
			Where("id = ?", finding.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		if eventType != "" {
			runID := input.LatestRunID
			if err := r.createEvent(
				tx,
				finding.ID,
				&runID,
				eventType,
				0,
				"",
				map[string]interface{}{
					"audit_id":          finding.AuditID,
					"rule_id":           finding.RuleID,
					"provider_audit_id": finding.ProviderAuditID,
					"clean_evaluations": nextClean,
					"sample_policy":     "all_samples_absent",
				},
				input.DecidedAt,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *SiteQualityFindingRepository) UpdateWithEvent(
	id uint,
	updates map[string]interface{},
	eventType string,
	actorUserID uint,
	note string,
	metadata map[string]interface{},
) (*sitequalitydomain.SiteQualityFinding, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("SiteQuality finding repository is unavailable")
	}
	if id == 0 {
		return nil, errors.New("SiteQuality finding ID is required")
	}
	if updates == nil {
		updates = map[string]interface{}{}
	}

	var result sitequalitydomain.SiteQualityFinding
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var finding sitequalitydomain.SiteQualityFinding
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&finding, id).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		updates["updated_at"] = now
		if len(updates) > 0 {
			if err := tx.Model(&sitequalitydomain.SiteQualityFinding{}).
				Where("id = ?", id).
				Updates(updates).Error; err != nil {
				return err
			}
		}
		if eventType != "" {
			if err := r.createEvent(tx, id, nil, eventType, actorUserID, note, withSiteQualityRuleMetadata(metadata, finding), now); err != nil {
				return err
			}
		}
		return tx.First(&result, id).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func withSiteQualityRuleMetadata(
	metadata map[string]interface{},
	finding sitequalitydomain.SiteQualityFinding,
) map[string]interface{} {
	normalizeSiteQualityFindingIdentity(&finding)
	next := make(map[string]interface{}, len(metadata)+3)
	for key, value := range metadata {
		next[key] = value
	}
	if _, exists := next["audit_id"]; !exists {
		next["audit_id"] = finding.AuditID
	}
	if _, exists := next["rule_id"]; !exists {
		next["rule_id"] = finding.RuleID
	}
	if _, exists := next["provider_audit_id"]; !exists {
		next["provider_audit_id"] = finding.ProviderAuditID
	}
	return next
}

func normalizeSiteQualityFindingIdentity(finding *sitequalitydomain.SiteQualityFinding) {
	if finding == nil {
		return
	}
	finding.RuleID, finding.ProviderAuditID = sitequalitydomain.NormalizeRuleIdentity(
		finding.RuleID,
		finding.AuditID,
		finding.ProviderAuditID,
	)
	finding.LatestEvidence = normalizeSiteQualityFindingEvidence(
		finding.LatestEvidence,
		finding.AuditID,
		finding.RuleID,
		finding.ProviderAuditID,
	)
}

func normalizeSiteQualityFindingEvidence(
	raw string,
	auditID string,
	ruleID string,
	providerAuditID string,
) string {
	evidence := map[string]interface{}{}
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &evidence); err != nil || evidence == nil {
			return raw
		}
	}
	if _, exists := evidence["audit_id"]; !exists {
		evidence["audit_id"] = strings.TrimSpace(auditID)
	}
	if strings.TrimSpace(ruleID) != "" {
		evidence["rule_id"] = strings.TrimSpace(ruleID)
	}
	if strings.TrimSpace(providerAuditID) != "" {
		evidence["provider_audit_id"] = strings.TrimSpace(providerAuditID)
	}
	encoded, err := json.Marshal(evidence)
	if err != nil {
		return raw
	}
	return string(encoded)
}

func (r *SiteQualityFindingRepository) createEvent(
	tx *gorm.DB,
	findingID uint,
	runID *uint,
	eventType string,
	actorUserID uint,
	note string,
	metadata map[string]interface{},
	createdAt time.Time,
) error {
	encodedMetadata, err := encodeSiteQualityFindingEventMetadata(metadata)
	if err != nil {
		return err
	}
	return tx.Create(&sitequalitydomain.SiteQualityFindingEvent{
		FindingID:   findingID,
		RunID:       runID,
		EventType:   eventType,
		ActorUserID: actorUserID,
		Note:        note,
		Metadata:    encodedMetadata,
		CreatedAt:   createdAt,
	}).Error
}

func encodeSiteQualityFindingEventMetadata(metadata map[string]interface{}) (string, error) {
	if len(metadata) == 0 {
		return "{}", nil
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}
