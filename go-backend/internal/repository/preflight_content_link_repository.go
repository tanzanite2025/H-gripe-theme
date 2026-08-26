package repository

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	preflightdomain "commerce-platform/internal/domain/preflight"
	sitequalitydomain "commerce-platform/internal/domain/sitequality"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PreflightContentLinkRepository struct {
	db *gorm.DB
}

type PreflightContentLinkIssueListFilter struct {
	Page      int
	PageSize  int
	State     string
	RuleID    string
	RunID     uint
	TargetURL string
	Search    string
	Fixable   *bool
}

type PreflightContentLinkStatsFilter struct {
	RuleID    string
	RunID     uint
	TargetURL string
}

func NewPreflightContentLinkRepository(db *gorm.DB) *PreflightContentLinkRepository {
	return &PreflightContentLinkRepository{db: db}
}

func (r *PreflightContentLinkRepository) CreateRun(
	run *preflightdomain.ContentLinkRun,
) error {
	if r == nil || r.db == nil {
		return errors.New("content link preflight repository is unavailable")
	}
	if run == nil {
		return errors.New("content link run is required")
	}
	normalizeContentLinkRunIdentity(run)
	if run.CheckedAt.IsZero() {
		run.CheckedAt = time.Now().UTC()
	} else {
		run.CheckedAt = run.CheckedAt.UTC()
	}
	return r.db.Create(run).Error
}

func (r *PreflightContentLinkRepository) RecordDetections(
	run *preflightdomain.ContentLinkRun,
	detections []preflightdomain.ContentLinkDetection,
) error {
	if r == nil || r.db == nil {
		return errors.New("content link preflight repository is unavailable")
	}
	if run == nil || run.ID == 0 {
		return errors.New("content link run is required")
	}
	if run.Status != preflightdomain.ContentLinkRunStatusSuccess {
		return nil
	}

	now := run.CheckedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()

	issueKeys := make([]string, 0, len(detections))
	seen := make(map[string]struct{}, len(detections))
	return r.db.Transaction(func(tx *gorm.DB) error {
		for index := range detections {
			detection := detections[index]
			detection.RunID = run.ID
			if detection.LastDetectedAt.IsZero() {
				detection.LastDetectedAt = now
			}
			if detection.FirstDetectedAt.IsZero() {
				detection.FirstDetectedAt = detection.LastDetectedAt
			}
			if _, ok := seen[detection.IssueKey]; !ok {
				seen[detection.IssueKey] = struct{}{}
				issueKeys = append(issueKeys, detection.IssueKey)
			}
			if err := r.recordDetection(tx, detection); err != nil {
				return err
			}
		}
		return r.verifyMissingIssues(tx, run, issueKeys, now)
	})
}

func (r *PreflightContentLinkRepository) ListIssues(
	filter PreflightContentLinkIssueListFilter,
) ([]preflightdomain.ContentLinkIssue, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("content link preflight repository is unavailable")
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 200 {
		filter.PageSize = 50
	}

	query := r.db.Model(&preflightdomain.ContentLinkIssue{})
	switch strings.TrimSpace(filter.State) {
	case "active":
		query = query.Where("state IN ?", []string{
			preflightdomain.ContentLinkIssueStateOpen,
			preflightdomain.ContentLinkIssueStateResolved,
		})
	case "", "all":
	default:
		query = query.Where("state = ?", strings.TrimSpace(filter.State))
	}
	if targetURL := strings.TrimSpace(filter.TargetURL); targetURL != "" {
		query = query.Where("target_url = ?", targetURL)
	}
	if ruleID := strings.TrimSpace(filter.RuleID); ruleID != "" {
		query = query.Where("rule_id = ?", sitequalitydomain.RuleIDForAuditID(ruleID))
	}
	if filter.RunID != 0 {
		query = query.Where("run_id = ?", filter.RunID)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"LOWER(target_url) LIKE LOWER(?) OR LOWER(link_url) LIKE LOWER(?) OR LOWER(link_text) LIKE LOWER(?) OR LOWER(suggested_text) LIKE LOWER(?)",
			like,
			like,
			like,
			like,
		)
	}
	if filter.Fixable != nil {
		if *filter.Fixable {
			query = query.Where("fix_status IN ?", []string{
				preflightdomain.ContentLinkFixStatusPending,
				preflightdomain.ContentLinkFixStatusApplied,
			})
		} else {
			query = query.Where("fix_status = ?", preflightdomain.ContentLinkFixStatusNotFixable)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var issues []preflightdomain.ContentLinkIssue
	err := query.
		Order("CASE state WHEN 'open' THEN 1 WHEN 'resolved' THEN 2 WHEN 'verified' THEN 3 ELSE 4 END").
		Order("CASE fix_status WHEN 'pending' THEN 1 WHEN 'failed' THEN 2 WHEN 'applied' THEN 3 ELSE 4 END").
		Order("last_detected_at DESC").
		Order("id DESC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&issues).Error
	for index := range issues {
		normalizeContentLinkIssueIdentity(&issues[index])
	}
	return issues, total, err
}

func (r *PreflightContentLinkRepository) FindIssueByID(
	id uint,
) (*preflightdomain.ContentLinkIssue, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("content link preflight repository is unavailable")
	}
	var issue preflightdomain.ContentLinkIssue
	if err := r.db.First(&issue, id).Error; err != nil {
		return nil, err
	}
	normalizeContentLinkIssueIdentity(&issue)
	return &issue, nil
}

func (r *PreflightContentLinkRepository) Stats(
	filter PreflightContentLinkStatsFilter,
) (preflightdomain.ContentLinkIssueStats, error) {
	if r == nil || r.db == nil {
		return preflightdomain.ContentLinkIssueStats{}, errors.New("content link preflight repository is unavailable")
	}
	query := r.db.Model(&preflightdomain.ContentLinkIssue{})
	if ruleID := strings.TrimSpace(filter.RuleID); ruleID != "" {
		query = query.Where("rule_id = ?", sitequalitydomain.RuleIDForAuditID(ruleID))
	}
	if filter.RunID != 0 {
		query = query.Where("run_id = ?", filter.RunID)
	}
	if targetURL := strings.TrimSpace(filter.TargetURL); targetURL != "" {
		query = query.Where("target_url = ?", targetURL)
	}
	var stats preflightdomain.ContentLinkIssueStats
	err := query.
		Select(`
			COALESCE(SUM(CASE WHEN state IN ('open', 'resolved') THEN 1 ELSE 0 END), 0) AS active,
			COALESCE(SUM(CASE WHEN state = 'open' THEN 1 ELSE 0 END), 0) AS open,
			COALESCE(SUM(CASE WHEN state = 'resolved' THEN 1 ELSE 0 END), 0) AS resolved,
			COALESCE(SUM(CASE WHEN state = 'verified' THEN 1 ELSE 0 END), 0) AS verified,
			COALESCE(SUM(CASE WHEN state = 'ignored' THEN 1 ELSE 0 END), 0) AS ignored,
			COALESCE(SUM(CASE WHEN state IN ('open', 'resolved') AND fix_status IN ('pending', 'applied') THEN 1 ELSE 0 END), 0) AS fixable,
			COALESCE(SUM(CASE WHEN state IN ('open', 'resolved') AND fix_status = 'not_fixable' THEN 1 ELSE 0 END), 0) AS needs_source,
			COALESCE(SUM(CASE WHEN fix_status = 'applied' THEN 1 ELSE 0 END), 0) AS applied
		`).
		Scan(&stats).Error
	return stats, err
}

func (r *PreflightContentLinkRepository) ListIssueEvents(
	issueID uint,
	page int,
	pageSize int,
) ([]preflightdomain.ContentLinkIssueEvent, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("content link preflight repository is unavailable")
	}
	if issueID == 0 {
		return nil, 0, errors.New("content link issue ID is required")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	query := r.db.Model(&preflightdomain.ContentLinkIssueEvent{}).Where("issue_id = ?", issueID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var events []preflightdomain.ContentLinkIssueEvent
	err := query.
		Order("created_at DESC").
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&events).Error
	return events, total, err
}

func (r *PreflightContentLinkRepository) UpdateIssueWithEvent(
	id uint,
	updates map[string]interface{},
	eventType string,
	actorUserID uint,
	note string,
	metadata map[string]interface{},
) (*preflightdomain.ContentLinkIssue, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("content link preflight repository is unavailable")
	}
	if id == 0 {
		return nil, errors.New("content link issue ID is required")
	}
	if updates == nil {
		updates = map[string]interface{}{}
	}

	var result preflightdomain.ContentLinkIssue
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var issue preflightdomain.ContentLinkIssue
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&issue, id).Error; err != nil {
			return err
		}
		normalizeContentLinkIssueIdentity(&issue)
		now := time.Now().UTC()
		updates["updated_at"] = now
		if err := tx.Model(&preflightdomain.ContentLinkIssue{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			return err
		}
		if eventType != "" {
			if err := r.createEvent(tx, id, eventType, actorUserID, note, withContentLinkRuleMetadata(metadata, issue), now); err != nil {
				return err
			}
		}
		return tx.First(&result, id).Error
	})
	if err != nil {
		return nil, err
	}
	normalizeContentLinkIssueIdentity(&result)
	return &result, nil
}

func withContentLinkRuleMetadata(
	metadata map[string]interface{},
	issue preflightdomain.ContentLinkIssue,
) map[string]interface{} {
	normalizeContentLinkIssueIdentity(&issue)
	next := make(map[string]interface{}, len(metadata)+2)
	for key, value := range metadata {
		next[key] = value
	}
	if _, exists := next["rule_id"]; !exists {
		next["rule_id"] = issue.RuleID
	}
	if _, exists := next["provider_audit_id"]; !exists {
		next["provider_audit_id"] = issue.ProviderAuditID
	}
	return next
}

func (r *PreflightContentLinkRepository) recordDetection(
	tx *gorm.DB,
	detection preflightdomain.ContentLinkDetection,
) error {
	if detection.IssueKey == "" || detection.TargetURL == "" {
		return errors.New("content link detection is incomplete")
	}
	ruleID, providerAuditID := sitequalitydomain.NormalizeRuleIdentity(
		detection.RuleID,
		"",
		detection.ProviderAuditID,
	)
	if ruleID == "" {
		ruleID = preflightdomain.ContentLinkRuleID
	}
	if providerAuditID == "" && ruleID == preflightdomain.ContentLinkRuleID {
		providerAuditID = preflightdomain.ContentLinkProviderAuditID
	}

	var issue preflightdomain.ContentLinkIssue
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("issue_key = ?", detection.IssueKey).
		First(&issue).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		issue = preflightdomain.ContentLinkIssue{
			RouteEntryID:    detection.RouteEntryID,
			RunID:           detection.RunID,
			TargetURL:       detection.TargetURL,
			RuleID:          ruleID,
			ProviderAuditID: providerAuditID,
			FinalURL:        detection.FinalURL,
			LinkURL:         detection.LinkURL,
			LinkText:        detection.LinkText,
			Selector:        detection.Selector,
			Snippet:         detection.Snippet,
			SourceType:      detection.SourceType,
			SourceID:        detection.SourceID,
			SourceKey:       detection.SourceKey,
			SourceField:     detection.SourceField,
			IssueKey:        detection.IssueKey,
			Severity:        defaultContentLinkSeverity(detection.Severity),
			State:           preflightdomain.ContentLinkIssueStateOpen,
			SuggestedText:   detection.SuggestedText,
			FixStatus:       defaultContentLinkFixStatus(detection.FixStatus),
			FixError:        "",
			LatestEvidence:  normalizeContentLinkEvidence(detection.LatestEvidence, ruleID, providerAuditID, detection),
			FirstDetectedAt: detection.FirstDetectedAt,
			LastDetectedAt:  detection.LastDetectedAt,
		}
		if err := tx.Create(&issue).Error; err != nil {
			return err
		}
		return r.createEvent(
			tx,
			issue.ID,
			preflightdomain.ContentLinkEventDetected,
			0,
			"",
			withContentLinkRuleMetadata(map[string]interface{}{
				"link_text": issue.LinkText,
				"link_url":  issue.LinkURL,
				"selector":  issue.Selector,
			}, issue),
			detection.LastDetectedAt,
		)
	}
	if err != nil {
		return err
	}
	normalizeContentLinkIssueIdentity(&issue)

	nextState := issue.State
	nextFixStatus := defaultContentLinkFixStatus(detection.FixStatus)
	nextFixError := ""
	reopened := false
	if issue.State == preflightdomain.ContentLinkIssueStateResolved ||
		issue.State == preflightdomain.ContentLinkIssueStateVerified {
		nextState = preflightdomain.ContentLinkIssueStateOpen
		reopened = true
		if issue.FixStatus == preflightdomain.ContentLinkFixStatusApplied {
			nextFixStatus = preflightdomain.ContentLinkFixStatusFailed
			nextFixError = "detected again after automatic fix"
		}
	}
	if issue.State == preflightdomain.ContentLinkIssueStateIgnored {
		nextState = preflightdomain.ContentLinkIssueStateIgnored
		nextFixStatus = issue.FixStatus
		nextFixError = issue.FixError
	}

	updates := map[string]interface{}{
		"route_entry_id":    detection.RouteEntryID,
		"run_id":            detection.RunID,
		"target_url":        detection.TargetURL,
		"rule_id":           ruleID,
		"provider_audit_id": providerAuditID,
		"final_url":         detection.FinalURL,
		"link_url":          detection.LinkURL,
		"link_text":         detection.LinkText,
		"selector":          detection.Selector,
		"snippet":           detection.Snippet,
		"source_type":       detection.SourceType,
		"source_id":         detection.SourceID,
		"source_key":        detection.SourceKey,
		"source_field":      detection.SourceField,
		"severity":          defaultContentLinkSeverity(detection.Severity),
		"state":             nextState,
		"suggested_text":    detection.SuggestedText,
		"fix_status":        nextFixStatus,
		"fix_error":         nextFixError,
		"latest_evidence":   normalizeContentLinkEvidence(detection.LatestEvidence, ruleID, providerAuditID, detection),
		"last_detected_at":  detection.LastDetectedAt,
		"resolved_at":       nil,
		"verified_at":       nil,
		"updated_at":        detection.LastDetectedAt,
	}
	if err := tx.Model(&preflightdomain.ContentLinkIssue{}).
		Where("id = ?", issue.ID).
		Updates(updates).Error; err != nil {
		return err
	}
	if reopened {
		return r.createEvent(
			tx,
			issue.ID,
			preflightdomain.ContentLinkEventReopened,
			0,
			"",
			withContentLinkRuleMetadata(map[string]interface{}{
				"link_text": detection.LinkText,
				"link_url":  detection.LinkURL,
			}, issue),
			detection.LastDetectedAt,
		)
	}
	return nil
}

func (r *PreflightContentLinkRepository) verifyMissingIssues(
	tx *gorm.DB,
	run *preflightdomain.ContentLinkRun,
	issueKeys []string,
	verifiedAt time.Time,
) error {
	ruleID := strings.TrimSpace(run.RuleID)
	if ruleID == "" {
		ruleID = preflightdomain.ContentLinkRuleID
	}
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("state IN ?", []string{
			preflightdomain.ContentLinkIssueStateOpen,
			preflightdomain.ContentLinkIssueStateResolved,
		}).
		Where("rule_id = ?", ruleID)
	if run.RouteEntryID != nil && *run.RouteEntryID != 0 {
		query = query.Where("route_entry_id = ?", *run.RouteEntryID)
	} else {
		query = query.Where("target_url = ?", run.TargetURL)
	}
	if len(issueKeys) > 0 {
		query = query.Where("issue_key NOT IN ?", issueKeys)
	}

	var stale []preflightdomain.ContentLinkIssue
	if err := query.Find(&stale).Error; err != nil {
		return err
	}
	for _, issue := range stale {
		if err := tx.Model(&preflightdomain.ContentLinkIssue{}).
			Where("id = ?", issue.ID).
			Updates(map[string]interface{}{
				"run_id":          run.ID,
				"state":           preflightdomain.ContentLinkIssueStateVerified,
				"verified_at":     verifiedAt,
				"fix_error":       "",
				"updated_at":      verifiedAt,
				"latest_evidence": defaultJSON(contentLinkVerificationEvidence(issue, run)),
			}).Error; err != nil {
			return err
		}
		if err := r.createEvent(
			tx,
			issue.ID,
			preflightdomain.ContentLinkEventVerificationPassed,
			0,
			"",
			withContentLinkRuleMetadata(map[string]interface{}{"run_id": run.ID}, issue),
			verifiedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func contentLinkVerificationEvidence(
	issue preflightdomain.ContentLinkIssue,
	run *preflightdomain.ContentLinkRun,
) string {
	evidence := map[string]interface{}{}
	if strings.TrimSpace(issue.LatestEvidence) != "" {
		_ = json.Unmarshal([]byte(issue.LatestEvidence), &evidence)
	}
	if evidence == nil {
		evidence = map[string]interface{}{}
	}
	evidence = contentLinkEvidenceWithIssueFields(evidence, issue)
	evidence["rule_id"] = issue.RuleID
	evidence["provider_audit_id"] = issue.ProviderAuditID
	if run != nil {
		evidence["verified_by_run_id"] = run.ID
		evidence["target_url"] = run.TargetURL
	}
	return mustContentLinkJSON(evidence)
}

func normalizeContentLinkEvidence(
	raw string,
	ruleID string,
	providerAuditID string,
	detection ...preflightdomain.ContentLinkDetection,
) string {
	evidence := map[string]interface{}{}
	if strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &evidence)
	}
	if evidence == nil {
		evidence = map[string]interface{}{}
	}
	evidence["rule_id"] = ruleID
	evidence["provider_audit_id"] = providerAuditID
	if len(detection) == 0 {
		return mustContentLinkJSON(evidence)
	}
	item := detection[0]
	if _, exists := evidence["href"]; !exists {
		evidence["href"] = item.LinkURL
	}
	if _, exists := evidence["link_url"]; !exists {
		evidence["link_url"] = item.LinkURL
	}
	if _, exists := evidence["link_text"]; !exists {
		evidence["link_text"] = item.LinkText
	}
	if _, exists := evidence["selector"]; !exists {
		evidence["selector"] = item.Selector
	}
	if _, exists := evidence["snippet"]; !exists {
		evidence["snippet"] = item.Snippet
	}
	if _, exists := evidence["source_type"]; !exists {
		evidence["source_type"] = item.SourceType
	}
	if _, exists := evidence["source_key"]; !exists {
		evidence["source_key"] = item.SourceKey
	}
	if _, exists := evidence["source_field"]; !exists {
		evidence["source_field"] = item.SourceField
	}
	return mustContentLinkJSON(evidence)
}

func normalizeContentLinkIssueIdentity(issue *preflightdomain.ContentLinkIssue) {
	if issue == nil {
		return
	}
	issue.RuleID, issue.ProviderAuditID = sitequalitydomain.NormalizeRuleIdentity(
		issue.RuleID,
		"",
		issue.ProviderAuditID,
	)
	if issue.RuleID == "" {
		issue.RuleID = preflightdomain.ContentLinkRuleID
	}
	if issue.ProviderAuditID == "" && issue.RuleID == preflightdomain.ContentLinkRuleID {
		issue.ProviderAuditID = preflightdomain.ContentLinkProviderAuditID
	}
	issue.LatestEvidence = contentLinkVerificationEvidence(*issue, nil)
}

func normalizeContentLinkRunIdentity(run *preflightdomain.ContentLinkRun) {
	if run == nil {
		return
	}
	run.RuleID, run.ProviderAuditID = sitequalitydomain.NormalizeRuleIdentity(
		run.RuleID,
		"",
		run.ProviderAuditID,
	)
	if run.RuleID == "" {
		run.RuleID = preflightdomain.ContentLinkRuleID
	}
	if run.ProviderAuditID == "" && run.RuleID == preflightdomain.ContentLinkRuleID {
		run.ProviderAuditID = preflightdomain.ContentLinkProviderAuditID
	}
}

func contentLinkEvidenceWithIssueFields(
	evidence map[string]interface{},
	issue preflightdomain.ContentLinkIssue,
) map[string]interface{} {
	if evidence == nil {
		evidence = map[string]interface{}{}
	}
	if _, exists := evidence["rule_id"]; !exists {
		evidence["rule_id"] = issue.RuleID
	}
	if _, exists := evidence["provider_audit_id"]; !exists {
		evidence["provider_audit_id"] = issue.ProviderAuditID
	}
	if _, exists := evidence["href"]; !exists {
		evidence["href"] = issue.LinkURL
	}
	if _, exists := evidence["link_url"]; !exists {
		evidence["link_url"] = issue.LinkURL
	}
	if _, exists := evidence["link_text"]; !exists {
		evidence["link_text"] = issue.LinkText
	}
	if _, exists := evidence["selector"]; !exists {
		evidence["selector"] = issue.Selector
	}
	if _, exists := evidence["snippet"]; !exists {
		evidence["snippet"] = issue.Snippet
	}
	if _, exists := evidence["source_type"]; !exists {
		evidence["source_type"] = issue.SourceType
	}
	if _, exists := evidence["source_key"]; !exists {
		evidence["source_key"] = issue.SourceKey
	}
	if _, exists := evidence["source_field"]; !exists {
		evidence["source_field"] = issue.SourceField
	}
	return evidence
}

func (r *PreflightContentLinkRepository) createEvent(
	tx *gorm.DB,
	issueID uint,
	eventType string,
	actorUserID uint,
	note string,
	metadata map[string]interface{},
	createdAt time.Time,
) error {
	encodedMetadata, err := encodeContentLinkEventMetadata(metadata)
	if err != nil {
		return err
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	return tx.Create(&preflightdomain.ContentLinkIssueEvent{
		IssueID:     issueID,
		EventType:   eventType,
		ActorUserID: actorUserID,
		Note:        strings.TrimSpace(note),
		Metadata:    encodedMetadata,
		CreatedAt:   createdAt.UTC(),
	}).Error
}

func encodeContentLinkEventMetadata(metadata map[string]interface{}) (string, error) {
	if len(metadata) == 0 {
		return "{}", nil
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func mustContentLinkJSON(value map[string]interface{}) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func defaultContentLinkSeverity(value string) string {
	switch strings.TrimSpace(value) {
	case "low", "medium", "high", "critical":
		return strings.TrimSpace(value)
	default:
		return "medium"
	}
}

func defaultContentLinkFixStatus(value string) string {
	switch strings.TrimSpace(value) {
	case preflightdomain.ContentLinkFixStatusPending,
		preflightdomain.ContentLinkFixStatusApplied,
		preflightdomain.ContentLinkFixStatusFailed,
		preflightdomain.ContentLinkFixStatusNotFixable:
		return strings.TrimSpace(value)
	default:
		return preflightdomain.ContentLinkFixStatusNotFixable
	}
}

func defaultJSON(value string) string {
	if strings.TrimSpace(value) == "" {
		return "{}"
	}
	return value
}
