package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"commerce-platform/internal/repository"
)

type StorefrontURLIssueService struct {
	issues        *repository.StorefrontURLIssueRepository
	catalog       *repository.StorefrontRouteCatalogRepository
	redirectRules *repository.StorefrontRedirectRuleRepository
}

func NewStorefrontURLIssueService(
	issues *repository.StorefrontURLIssueRepository,
	catalog *repository.StorefrontRouteCatalogRepository,
	redirectRules *repository.StorefrontRedirectRuleRepository,
) *StorefrontURLIssueService {
	return &StorefrontURLIssueService{
		issues:        issues,
		catalog:       catalog,
		redirectRules: redirectRules,
	}
}

func (s *StorefrontURLIssueService) List(
	filter repository.StorefrontURLIssueListFilter,
) ([]urlmanagementdomain.StorefrontURLIssue, int64, error) {
	if s == nil || s.issues == nil {
		return nil, 0, errors.New("storefront URL issue service is unavailable")
	}
	return s.issues.List(filter)
}

func (s *StorefrontURLIssueService) Stats() (urlmanagementdomain.StorefrontURLIssueStats, error) {
	if s == nil || s.issues == nil {
		return urlmanagementdomain.StorefrontURLIssueStats{}, errors.New("storefront URL issue service is unavailable")
	}
	return s.issues.Stats()
}

func (s *StorefrontURLIssueService) Get(id uint) (*urlmanagementdomain.StorefrontURLIssue, error) {
	if s == nil || s.issues == nil {
		return nil, errors.New("storefront URL issue service is unavailable")
	}
	if id == 0 {
		return nil, errors.New("URL issue ID is required")
	}
	return s.issues.FindByID(id)
}

func (s *StorefrontURLIssueService) ListEvents(
	issueID uint,
	page int,
	pageSize int,
) ([]urlmanagementdomain.StorefrontURLIssueEvent, int64, error) {
	if s == nil || s.issues == nil {
		return nil, 0, errors.New("storefront URL issue service is unavailable")
	}
	if _, err := s.issues.FindByID(issueID); err != nil {
		return nil, 0, err
	}
	return s.issues.ListEvents(issueID, page, pageSize)
}

// ReconcileCatalog projects current route observations into durable issue work
// items. It only opens or reopens detected issues; it never auto-closes work
// that has not been explicitly resolved and verified by an operator.
func (s *StorefrontURLIssueService) ReconcileCatalog(ctx context.Context) error {
	if s == nil || s.catalog == nil {
		return errors.New("storefront route catalog is unavailable")
	}
	ids, err := s.catalog.ListIssueCandidateIDs()
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := s.ReconcileEntry(ctx, id, nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *StorefrontURLIssueService) ReconcileEntry(
	_ context.Context,
	routeEntryID uint,
	latestCheckResultID *uint,
) error {
	if s == nil || s.issues == nil || s.catalog == nil {
		return errors.New("storefront URL issue service is unavailable")
	}
	entry, err := s.catalog.FindByID(routeEntryID)
	if err != nil {
		return err
	}
	for _, definition := range deriveStorefrontURLIssueDefinitions(*entry) {
		if _, err := s.issues.RecordDetection(
			entry.ID,
			definition.issueType,
			definition.severity,
			latestCheckResultID,
			time.Now().UTC(),
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *StorefrontURLIssueService) Acknowledge(
	id uint,
	actorUserID uint,
	input urlmanagementdomain.StorefrontURLIssueActionInput,
) (*urlmanagementdomain.StorefrontURLIssue, error) {
	issue, err := s.requireActionableIssue(id)
	if err != nil {
		return nil, err
	}
	if issue.State != urlmanagementdomain.URLIssueStateOpen &&
		issue.State != urlmanagementdomain.URLIssueStateAcknowledged {
		return nil, errors.New("only open URL issues can be acknowledged")
	}
	return s.issues.UpdateWithEvent(
		id,
		map[string]interface{}{"state": urlmanagementdomain.URLIssueStateAcknowledged},
		urlmanagementdomain.URLIssueEventAcknowledged,
		actorUserID,
		strings.TrimSpace(input.Note),
		nil,
	)
}

func (s *StorefrontURLIssueService) Claim(
	id uint,
	actorUserID uint,
) (*urlmanagementdomain.StorefrontURLIssue, error) {
	if actorUserID == 0 {
		return nil, errors.New("an authenticated user is required to claim a URL issue")
	}
	issue, err := s.requireActionableIssue(id)
	if err != nil {
		return nil, err
	}
	if issue.State != urlmanagementdomain.URLIssueStateOpen &&
		issue.State != urlmanagementdomain.URLIssueStateAcknowledged {
		return nil, errors.New("only open URL issues can be claimed")
	}
	return s.issues.UpdateWithEvent(
		id,
		map[string]interface{}{
			"assignee_id": actorUserID,
			"state":       urlmanagementdomain.URLIssueStateAcknowledged,
		},
		urlmanagementdomain.URLIssueEventClaimed,
		actorUserID,
		"",
		nil,
	)
}

func (s *StorefrontURLIssueService) AddComment(
	id uint,
	actorUserID uint,
	input urlmanagementdomain.StorefrontURLIssueActionInput,
) (*urlmanagementdomain.StorefrontURLIssue, error) {
	if _, err := s.requireActionableIssue(id); err != nil {
		return nil, err
	}
	note := strings.TrimSpace(input.Note)
	if note == "" {
		return nil, errors.New("comment is required")
	}
	return s.issues.UpdateWithEvent(
		id,
		nil,
		urlmanagementdomain.URLIssueEventCommented,
		actorUserID,
		note,
		nil,
	)
}

func (s *StorefrontURLIssueService) LinkRedirect(
	id uint,
	actorUserID uint,
	input urlmanagementdomain.StorefrontURLIssueLinkRedirectInput,
) (*urlmanagementdomain.StorefrontURLIssue, error) {
	if _, err := s.requireActionableIssue(id); err != nil {
		return nil, err
	}
	if input.RedirectRuleID == 0 {
		return nil, errors.New("redirect rule ID is required")
	}
	if s.redirectRules == nil {
		return nil, errors.New("storefront redirect rule repository is unavailable")
	}
	rule, err := s.redirectRules.FindByID(input.RedirectRuleID)
	if err != nil {
		return nil, err
	}
	return s.issues.UpdateWithEvent(
		id,
		map[string]interface{}{"linked_redirect_rule_id": input.RedirectRuleID},
		urlmanagementdomain.URLIssueEventRedirectLinked,
		actorUserID,
		"",
		map[string]interface{}{
			"redirect_rule_id": input.RedirectRuleID,
			"source_path":      rule.SourcePath,
			"target_path":      rule.TargetPath,
			"state":            rule.State,
		},
	)
}

func (s *StorefrontURLIssueService) Resolve(
	id uint,
	actorUserID uint,
	input urlmanagementdomain.StorefrontURLIssueResolutionInput,
) (*urlmanagementdomain.StorefrontURLIssue, error) {
	issue, err := s.requireActionableIssue(id)
	if err != nil {
		return nil, err
	}
	if issue.State != urlmanagementdomain.URLIssueStateOpen &&
		issue.State != urlmanagementdomain.URLIssueStateAcknowledged {
		return nil, errors.New("only open URL issues can be resolved")
	}

	resolutionType := strings.TrimSpace(input.ResolutionType)
	if !isValidStorefrontURLIssueResolution(resolutionType) {
		return nil, errors.New("resolution type is invalid")
	}
	resolutionNote := strings.TrimSpace(input.ResolutionNote)
	if resolutionNote == "" {
		return nil, errors.New("resolution note is required")
	}
	if input.LinkedRedirectRuleID != nil {
		if err := s.validateResolutionRedirect(*input.LinkedRedirectRuleID, resolutionType); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"state":           urlmanagementdomain.URLIssueStateResolved,
		"resolution_type": resolutionType,
		"resolution_note": resolutionNote,
		"resolved_at":     now,
		"verified_at":     nil,
	}
	if input.LinkedRedirectRuleID != nil {
		updates["linked_redirect_rule_id"] = *input.LinkedRedirectRuleID
	}
	return s.issues.UpdateWithEvent(
		id,
		updates,
		urlmanagementdomain.URLIssueEventResolutionRecorded,
		actorUserID,
		resolutionNote,
		map[string]interface{}{
			"resolution_type":         resolutionType,
			"linked_redirect_rule_id": input.LinkedRedirectRuleID,
		},
	)
}

func (s *StorefrontURLIssueService) Suppress(
	id uint,
	actorUserID uint,
	input urlmanagementdomain.StorefrontURLIssueSuppressionInput,
) (*urlmanagementdomain.StorefrontURLIssue, error) {
	issue, err := s.requireActionableIssue(id)
	if err != nil {
		return nil, err
	}
	if issue.State != urlmanagementdomain.URLIssueStateOpen &&
		issue.State != urlmanagementdomain.URLIssueStateAcknowledged {
		return nil, errors.New("only open URL issues can be suppressed")
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return nil, errors.New("suppression reason is required")
	}
	until, err := parseStorefrontURLIssueTimestamp(input.SuppressedUntil)
	if err != nil {
		return nil, err
	}
	if !until.After(time.Now().UTC()) {
		return nil, errors.New("suppression review time must be in the future")
	}

	return s.issues.UpdateWithEvent(
		id,
		map[string]interface{}{
			"state":              urlmanagementdomain.URLIssueStateSuppressed,
			"suppressed_until":   until,
			"suppression_reason": reason,
		},
		urlmanagementdomain.URLIssueEventSuppressed,
		actorUserID,
		reason,
		map[string]interface{}{"suppressed_until": until},
	)
}

func (s *StorefrontURLIssueService) Verify(
	id uint,
	actorUserID uint,
) (*urlmanagementdomain.StorefrontURLIssue, error) {
	issue, err := s.requireActionableIssue(id)
	if err != nil {
		return nil, err
	}
	if issue.State != urlmanagementdomain.URLIssueStateResolved {
		return nil, errors.New("record a resolution before verification")
	}
	if issue.RouteEntry == nil {
		return nil, errors.New("URL issue route entry is unavailable")
	}
	entry, err := s.catalog.FindByID(issue.RouteEntryID)
	if err != nil {
		return nil, err
	}
	if storefrontURLIssueStillDetected(issue.IssueType, *entry) {
		return s.issues.UpdateWithEvent(
			id,
			map[string]interface{}{
				"state":           urlmanagementdomain.URLIssueStateOpen,
				"resolution_type": "",
				"resolution_note": "",
				"resolved_at":     nil,
				"verified_at":     nil,
			},
			urlmanagementdomain.URLIssueEventVerificationFailed,
			actorUserID,
			"",
			map[string]interface{}{"last_check_status": entry.LastCheckStatus},
		)
	}

	now := time.Now().UTC()
	return s.issues.UpdateWithEvent(
		id,
		map[string]interface{}{
			"state":       urlmanagementdomain.URLIssueStateVerified,
			"verified_at": now,
		},
		urlmanagementdomain.URLIssueEventVerificationPassed,
		actorUserID,
		"",
		nil,
	)
}

func (s *StorefrontURLIssueService) RequiresRouteCheck(issue *urlmanagementdomain.StorefrontURLIssue) bool {
	if issue == nil {
		return false
	}
	switch issue.IssueType {
	case urlmanagementdomain.URLIssueTypePathCollision, urlmanagementdomain.URLIssueTypeStaleRoute:
		return false
	default:
		return true
	}
}

func (s *StorefrontURLIssueService) requireActionableIssue(
	id uint,
) (*urlmanagementdomain.StorefrontURLIssue, error) {
	if s == nil || s.issues == nil || s.catalog == nil {
		return nil, errors.New("storefront URL issue service is unavailable")
	}
	if id == 0 {
		return nil, errors.New("URL issue ID is required")
	}
	issue, err := s.issues.FindByID(id)
	if err != nil {
		return nil, err
	}
	if issue.RouteEntry == nil {
		return nil, errors.New("URL issue route entry is unavailable")
	}
	return issue, nil
}

func (s *StorefrontURLIssueService) validateResolutionRedirect(
	redirectRuleID uint,
	resolutionType string,
) error {
	if s.redirectRules == nil {
		return errors.New("storefront redirect rule repository is unavailable")
	}
	rule, err := s.redirectRules.FindByID(redirectRuleID)
	if err != nil {
		return err
	}
	if resolutionType == urlmanagementdomain.URLIssueResolutionRedirectPublished &&
		rule.State != urlmanagementdomain.RedirectRuleStatePublished {
		return errors.New("a redirect-published resolution requires a published redirect rule")
	}
	return nil
}

type storefrontURLIssueDefinition struct {
	issueType string
	severity  string
}

func deriveStorefrontURLIssueDefinitions(
	entry seodomain.StorefrontRouteCatalogEntry,
) []storefrontURLIssueDefinition {
	definitions := make([]storefrontURLIssueDefinition, 0, 2)
	switch entry.EntryStatus {
	case seodomain.RouteEntryStatusDuplicate:
		definitions = append(definitions, storefrontURLIssueDefinition{
			issueType: urlmanagementdomain.URLIssueTypePathCollision,
			severity:  urlmanagementdomain.URLIssueSeverityCritical,
		})
	case seodomain.RouteEntryStatusStale:
		definitions = append(definitions, storefrontURLIssueDefinition{
			issueType: urlmanagementdomain.URLIssueTypeStaleRoute,
			severity:  urlmanagementdomain.URLIssueSeverityMedium,
		})
	}

	switch entry.LastCheckStatus {
	case seodomain.RouteCheckStatusRedirectChain:
		definitions = append(definitions, storefrontURLIssueDefinition{
			issueType: urlmanagementdomain.URLIssueTypeRedirectChain,
			severity:  urlmanagementdomain.URLIssueSeverityHigh,
		})
	case seodomain.RouteCheckStatusRedirectTarget:
		definitions = append(definitions, storefrontURLIssueDefinition{
			issueType: urlmanagementdomain.URLIssueTypeRedirectTargetMismatch,
			severity:  urlmanagementdomain.URLIssueSeverityHigh,
		})
	case seodomain.RouteCheckStatusRedirect:
		if !entry.IsAlias {
			definitions = append(definitions, storefrontURLIssueDefinition{
				issueType: urlmanagementdomain.URLIssueTypeRedirectStatusMismatch,
				severity:  urlmanagementdomain.URLIssueSeverityMedium,
			})
		}
	case seodomain.RouteCheckStatusNotFound:
		definitions = append(definitions, storefrontURLIssueDefinition{
			issueType: urlmanagementdomain.URLIssueTypeNotFound,
			severity:  urlmanagementdomain.URLIssueSeverityHigh,
		})
	case seodomain.RouteCheckStatusServerError:
		definitions = append(definitions, storefrontURLIssueDefinition{
			issueType: urlmanagementdomain.URLIssueTypeServerError,
			severity:  urlmanagementdomain.URLIssueSeverityHigh,
		})
	case seodomain.RouteCheckStatusCanonicalMisfit:
		definitions = append(definitions, storefrontURLIssueDefinition{
			issueType: urlmanagementdomain.URLIssueTypeCanonicalMismatch,
			severity:  urlmanagementdomain.URLIssueSeverityMedium,
		})
	case seodomain.RouteCheckStatusError:
		definitions = append(definitions, storefrontURLIssueDefinition{
			issueType: urlmanagementdomain.URLIssueTypeCheckError,
			severity:  urlmanagementdomain.URLIssueSeverityMedium,
		})
	}
	return definitions
}

func storefrontURLIssueStillDetected(
	issueType string,
	entry seodomain.StorefrontRouteCatalogEntry,
) bool {
	for _, definition := range deriveStorefrontURLIssueDefinitions(entry) {
		if definition.issueType == issueType {
			return true
		}
	}
	return false
}

func isValidStorefrontURLIssueResolution(value string) bool {
	switch value {
	case urlmanagementdomain.URLIssueResolutionRedirectPublished,
		urlmanagementdomain.URLIssueResolutionSourceRestored,
		urlmanagementdomain.URLIssueResolutionSourcePathChanged,
		urlmanagementdomain.URLIssueResolutionCanonicalFixed,
		urlmanagementdomain.URLIssueResolutionRuntimeFixed,
		urlmanagementdomain.URLIssueResolutionRetired,
		urlmanagementdomain.URLIssueResolutionNotApplicable:
		return true
	default:
		return false
	}
}

func parseStorefrontURLIssueTimestamp(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, errors.New("suppression review time is required")
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04"} {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("suppression review time is invalid")
}
