package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strings"
	"time"

	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/repository"
)

type VisitorProfileService struct {
	visitorRepo *repository.VisitorProfileRepository
}

func NewVisitorProfileService(visitorRepo *repository.VisitorProfileRepository) *VisitorProfileService {
	return &VisitorProfileService{visitorRepo: visitorRepo}
}

const (
	VisitorProfileActionCart            = "cart_action"
	VisitorProfileActionCustomerService = "customer_service"
	VisitorProfileActionEmailCapture    = "email_capture"
	VisitorProfileActionAccount         = "account"
	VisitorProfileActionIdentityBind    = "identity_bind"
	VisitorProfileActionMeaningful      = "meaningful_action"

	VisitorProfileQualityCartAction      = 8
	VisitorProfileQualityCustomerService = 12
	VisitorProfileQualityEmailCapture    = 12
	VisitorProfileQualityAccount         = 20
)

const activeAnonymousProfileRetention = 180 * 24 * time.Hour

type VisitorProfileTouchInput struct {
	UserID                     *uint
	CustomerServiceVisitorHash string
	CartSessionID              string
	Email                      string
	EmailSource                string
	Locale                     string
	LocaleSource               string
	CountryCode                string
	Region                     string
	City                       string
	Timezone                   string
	IPAddress                  string
	UserAgent                  string
	SeenAt                     time.Time
	MeaningfulAction           string
	QualityScoreDelta          int
}

type VisitorProfileListInput struct {
	Search                 string
	Identity               string
	CountryCode            string
	Locale                 string
	Email                  string
	CartSession            string
	CustomerServiceVisitor string
	LastSeen               string
	LastMeaningful         string
	Status                 string
}

type VisitorProfileSnapshot struct {
	ID                                uint       `json:"id"`
	UserID                            *uint      `json:"user_id,omitempty"`
	Identity                          string     `json:"identity"`
	CustomerServiceVisitorHashPreview string     `json:"customer_service_visitor_hash_preview,omitempty"`
	HasCustomerServiceVisitor         bool       `json:"has_customer_service_visitor"`
	CartSessionID                     string     `json:"cart_session_id,omitempty"`
	HasCartSession                    bool       `json:"has_cart_session"`
	Email                             string     `json:"email,omitempty"`
	EmailSource                       string     `json:"email_source,omitempty"`
	HasEmail                          bool       `json:"has_email"`
	Locale                            string     `json:"locale,omitempty"`
	LocaleSource                      string     `json:"locale_source,omitempty"`
	CountryCode                       string     `json:"country_code,omitempty"`
	Region                            string     `json:"region,omitempty"`
	City                              string     `json:"city,omitempty"`
	Timezone                          string     `json:"timezone,omitempty"`
	RegionLabel                       string     `json:"region_label,omitempty"`
	HasIPFingerprint                  bool       `json:"has_ip_fingerprint"`
	HasUserAgentFingerprint           bool       `json:"has_user_agent_fingerprint"`
	ProfileQualityScore               int        `json:"profile_quality_score"`
	ProfileStatus                     string     `json:"profile_status"`
	LastMeaningfulAction              string     `json:"last_meaningful_action,omitempty"`
	FirstMeaningfulSeenAt             *time.Time `json:"first_meaningful_seen_at,omitempty"`
	LastMeaningfulSeenAt              *time.Time `json:"last_meaningful_seen_at,omitempty"`
	RetentionUntil                    *time.Time `json:"retention_until,omitempty"`
	LastSeenAt                        time.Time  `json:"last_seen_at"`
	CreatedAt                         time.Time  `json:"created_at"`
	UpdatedAt                         time.Time  `json:"updated_at"`
}

type VisitorProfileStats struct {
	Total                int64 `json:"total"`
	AccountCount         int64 `json:"account_count"`
	AnonymousCount       int64 `json:"anonymous_count"`
	EmailCount           int64 `json:"email_count"`
	CartLinkedCount      int64 `json:"cart_linked_count"`
	CustomerServiceCount int64 `json:"customer_service_count"`
	RegionCount          int64 `json:"region_count"`
	Recent24hCount       int64 `json:"recent_24h_count"`
	ActiveCount          int64 `json:"active_count"`
	CandidateCount       int64 `json:"candidate_count"`
	ArchivedCount        int64 `json:"archived_count"`
	SuppressedCount      int64 `json:"suppressed_count"`
}

type VisitorProfileRetentionCleanupResult struct {
	DeletedCandidates         int64     `json:"deleted_candidates"`
	ArchivedAnonymous         int64     `json:"archived_anonymous"`
	TotalChanged              int64     `json:"total_changed"`
	CleanupReferenceTimestamp time.Time `json:"cleanup_reference_timestamp"`
}

func (s *VisitorProfileService) Touch(input VisitorProfileTouchInput) (*visitor.Profile, error) {
	return s.TouchMeaningfulAction(input)
}

func (s *VisitorProfileService) TouchMeaningfulAction(input VisitorProfileTouchInput) (*visitor.Profile, error) {
	return s.touch(input, true, true)
}

func (s *VisitorProfileService) BindIdentityFact(input VisitorProfileTouchInput) (*visitor.Profile, error) {
	if input.MeaningfulAction == "" {
		input.MeaningfulAction = VisitorProfileActionIdentityBind
	}
	return s.touch(input, true, true)
}

func (s *VisitorProfileService) TouchPassiveSeen(input VisitorProfileTouchInput) (*visitor.Profile, error) {
	return s.touch(input, false, false)
}

func (s *VisitorProfileService) touch(input VisitorProfileTouchInput, meaningful bool, createIfMissing bool) (*visitor.Profile, error) {
	if s == nil || s.visitorRepo == nil {
		return nil, nil
	}

	input = normalizeVisitorProfileTouch(input)
	if !input.hasIdentity() {
		return nil, nil
	}

	profile, err := s.findProfile(input)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		if !createIfMissing {
			return nil, nil
		}
		profile = &visitor.Profile{}
		applyVisitorProfileTouch(profile, input)
		if meaningful {
			applyMeaningfulVisitorProfileTouch(profile, input)
		} else {
			applyCandidateVisitorProfileTouch(profile, input)
		}
		if err := s.visitorRepo.Create(profile); err != nil {
			return nil, err
		}
		return profile, nil
	}

	applyVisitorProfileTouch(profile, input)
	if meaningful {
		applyMeaningfulVisitorProfileTouch(profile, input)
	}
	if err := s.visitorRepo.Update(profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *VisitorProfileService) FindByCustomerServiceVisitorHash(hash string) (*visitor.Profile, error) {
	if s == nil || s.visitorRepo == nil || strings.TrimSpace(hash) == "" {
		return nil, repository.ErrRecordNotFound
	}
	return s.visitorRepo.FindByCustomerServiceVisitorHash(hash)
}

func (s *VisitorProfileService) FindByUserID(userID uint) (*visitor.Profile, error) {
	if s == nil || s.visitorRepo == nil || userID == 0 {
		return nil, repository.ErrRecordNotFound
	}
	return s.visitorRepo.FindByUserID(userID)
}

func (s *VisitorProfileService) ListProfiles(page, pageSize int, input VisitorProfileListInput) ([]VisitorProfileSnapshot, int64, error) {
	if s == nil || s.visitorRepo == nil {
		return nil, 0, nil
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := repository.VisitorProfileListFilters{
		Search:                    strings.TrimSpace(input.Search),
		Identity:                  normalizeVisitorProfileAdminFilter(input.Identity),
		CountryCode:               strings.ToUpper(strings.TrimSpace(input.CountryCode)),
		Locale:                    normalizeVisitorLocale(input.Locale),
		HasEmail:                  parseVisitorProfileTriState(input.Email),
		HasCartSession:            parseVisitorProfileTriState(input.CartSession),
		HasCustomerServiceVisitor: parseVisitorProfileTriState(input.CustomerServiceVisitor),
		LastSeenAfter:             visitorProfileLastSeenBoundary(input.LastSeen),
		LastMeaningfulAfter:       visitorProfileLastSeenBoundary(input.LastMeaningful),
		Status:                    normalizeVisitorProfileStatusFilter(input.Status),
	}

	profiles, total, err := s.visitorRepo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}

	items := make([]VisitorProfileSnapshot, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, visitorProfileSnapshot(profile))
	}
	return items, total, nil
}

func (s *VisitorProfileService) GetStats() (VisitorProfileStats, error) {
	if s == nil || s.visitorRepo == nil {
		return VisitorProfileStats{}, nil
	}

	rawStats, err := s.visitorRepo.Stats()
	if err != nil {
		return VisitorProfileStats{}, err
	}

	return VisitorProfileStats{
		Total:                rawStats["total"],
		AccountCount:         rawStats["account_count"],
		AnonymousCount:       rawStats["anonymous_count"],
		EmailCount:           rawStats["email_count"],
		CartLinkedCount:      rawStats["cart_linked_count"],
		CustomerServiceCount: rawStats["customer_service_count"],
		RegionCount:          rawStats["region_count"],
		Recent24hCount:       rawStats["recent_24h_count"],
		ActiveCount:          rawStats["active_count"],
		CandidateCount:       rawStats["candidate_count"],
		ArchivedCount:        rawStats["archived_count"],
		SuppressedCount:      rawStats["suppressed_count"],
	}, nil
}

func (s *VisitorProfileService) CleanupExpiredProfiles(now time.Time) (VisitorProfileRetentionCleanupResult, error) {
	if s == nil || s.visitorRepo == nil {
		return VisitorProfileRetentionCleanupResult{}, nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	deletedCandidates, err := s.visitorRepo.DeleteExpiredCandidates(now)
	if err != nil {
		return VisitorProfileRetentionCleanupResult{}, err
	}

	archivedAnonymous, err := s.visitorRepo.ArchiveExpiredAnonymousProfiles(now)
	if err != nil {
		return VisitorProfileRetentionCleanupResult{}, err
	}

	return VisitorProfileRetentionCleanupResult{
		DeletedCandidates:         deletedCandidates,
		ArchivedAnonymous:         archivedAnonymous,
		TotalChanged:              deletedCandidates + archivedAnonymous,
		CleanupReferenceTimestamp: now,
	}, nil
}

func (s *VisitorProfileService) findProfile(input VisitorProfileTouchInput) (*visitor.Profile, error) {
	if input.CustomerServiceVisitorHash != "" {
		profile, err := s.visitorRepo.FindByCustomerServiceVisitorHash(input.CustomerServiceVisitorHash)
		if err == nil {
			return profile, nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	if input.CartSessionID != "" {
		profile, err := s.visitorRepo.FindByCartSessionID(input.CartSessionID)
		if err == nil {
			return profile, nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	if input.UserID != nil && *input.UserID > 0 {
		profile, err := s.visitorRepo.FindByUserID(*input.UserID)
		if err == nil {
			return profile, nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	return nil, nil
}

func normalizeVisitorProfileTouch(input VisitorProfileTouchInput) VisitorProfileTouchInput {
	input.CustomerServiceVisitorHash = strings.TrimSpace(input.CustomerServiceVisitorHash)
	input.CartSessionID = strings.TrimSpace(input.CartSessionID)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.EmailSource = strings.TrimSpace(input.EmailSource)
	input.Locale = normalizeVisitorLocale(input.Locale)
	input.LocaleSource = strings.TrimSpace(input.LocaleSource)
	input.CountryCode = strings.ToUpper(strings.TrimSpace(input.CountryCode))
	input.Region = strings.TrimSpace(input.Region)
	input.City = strings.TrimSpace(input.City)
	input.Timezone = normalizeVisitorTimezone(input.Timezone)
	input.IPAddress = normalizeVisitorIP(input.IPAddress)
	input.UserAgent = strings.TrimSpace(input.UserAgent)
	input.MeaningfulAction = normalizeVisitorProfileAction(input.MeaningfulAction)
	if input.QualityScoreDelta < 0 {
		input.QualityScoreDelta = 0
	}
	if input.SeenAt.IsZero() {
		input.SeenAt = time.Now().UTC()
	}
	if input.Email != "" && input.EmailSource == "" {
		input.EmailSource = "unknown"
	}
	if input.Locale != "" && input.LocaleSource == "" {
		input.LocaleSource = "request"
	}
	return input
}

func normalizeVisitorTimezone(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if _, err := time.LoadLocation(value); err != nil {
		return ""
	}
	return value
}

func (input VisitorProfileTouchInput) hasIdentity() bool {
	return input.UserID != nil ||
		input.CustomerServiceVisitorHash != "" ||
		input.CartSessionID != "" ||
		input.Email != ""
}

func applyVisitorProfileTouch(profile *visitor.Profile, input VisitorProfileTouchInput) {
	if input.UserID != nil && *input.UserID > 0 {
		profile.UserID = input.UserID
	}
	if input.CustomerServiceVisitorHash != "" {
		profile.CustomerServiceVisitorHash = input.CustomerServiceVisitorHash
	}
	if input.CartSessionID != "" {
		profile.CartSessionID = input.CartSessionID
	}
	if input.Email != "" {
		profile.Email = input.Email
		profile.EmailSource = input.EmailSource
	}
	if input.Locale != "" {
		profile.Locale = input.Locale
		profile.LocaleSource = input.LocaleSource
	}
	if input.CountryCode != "" {
		profile.CountryCode = input.CountryCode
	}
	if input.Region != "" {
		profile.Region = input.Region
	}
	if input.City != "" {
		profile.City = input.City
	}
	if input.Timezone != "" {
		profile.Timezone = input.Timezone
	}
	if input.IPAddress != "" {
		profile.IPHash = stableVisitorHash(input.IPAddress)
	}
	if input.UserAgent != "" {
		profile.UserAgentHash = stableVisitorHash(input.UserAgent)
	}
	profile.LastSeenAt = input.SeenAt
}

func applyCandidateVisitorProfileTouch(profile *visitor.Profile, input VisitorProfileTouchInput) {
	if strings.TrimSpace(profile.ProfileStatus) == "" {
		profile.ProfileStatus = visitor.ProfileStatusCandidate
	}
	if profile.RetentionUntil == nil {
		retentionUntil := input.SeenAt.Add(14 * 24 * time.Hour)
		profile.RetentionUntil = &retentionUntil
	}
}

func applyMeaningfulVisitorProfileTouch(profile *visitor.Profile, input VisitorProfileTouchInput) {
	action := input.MeaningfulAction
	if action == "" {
		action = inferredVisitorProfileAction(input)
	}

	scoreDelta := input.QualityScoreDelta
	if scoreDelta <= 0 {
		scoreDelta = visitorProfileQualityDelta(action, input)
	}
	profile.ProfileQualityScore += scoreDelta
	if profile.ProfileQualityScore < scoreDelta {
		profile.ProfileQualityScore = scoreDelta
	}

	if profile.ProfileStatus == "" ||
		profile.ProfileStatus == visitor.ProfileStatusCandidate ||
		profile.ProfileStatus == visitor.ProfileStatusArchived {
		profile.ProfileStatus = visitor.ProfileStatusActive
	}

	if action != "" {
		profile.LastMeaningfulAction = action
	}

	seenAt := input.SeenAt
	if profile.FirstMeaningfulSeenAt == nil || profile.FirstMeaningfulSeenAt.IsZero() {
		firstSeen := seenAt
		profile.FirstMeaningfulSeenAt = &firstSeen
	}
	lastSeen := seenAt
	profile.LastMeaningfulSeenAt = &lastSeen

	if profile.UserID != nil && *profile.UserID > 0 {
		profile.RetentionUntil = nil
		return
	}
	retentionUntil := seenAt.Add(activeAnonymousProfileRetention)
	profile.RetentionUntil = &retentionUntil
}

func normalizeVisitorProfileAction(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

func inferredVisitorProfileAction(input VisitorProfileTouchInput) string {
	if input.UserID != nil && *input.UserID > 0 {
		return VisitorProfileActionAccount
	}
	if input.Email != "" {
		return VisitorProfileActionEmailCapture
	}
	if input.CustomerServiceVisitorHash != "" {
		return VisitorProfileActionCustomerService
	}
	if input.CartSessionID != "" {
		return VisitorProfileActionCart
	}
	return VisitorProfileActionMeaningful
}

func visitorProfileQualityDelta(action string, input VisitorProfileTouchInput) int {
	switch action {
	case VisitorProfileActionAccount:
		return VisitorProfileQualityAccount
	case VisitorProfileActionEmailCapture:
		return VisitorProfileQualityEmailCapture
	case VisitorProfileActionCustomerService:
		return VisitorProfileQualityCustomerService
	case VisitorProfileActionCart:
		return VisitorProfileQualityCartAction
	case VisitorProfileActionIdentityBind:
		if input.UserID != nil && *input.UserID > 0 {
			return VisitorProfileQualityAccount
		}
		if input.Email != "" {
			return VisitorProfileQualityEmailCapture
		}
		return VisitorProfileQualityCustomerService
	default:
		if input.UserID != nil && *input.UserID > 0 {
			return VisitorProfileQualityAccount
		}
		if input.Email != "" {
			return VisitorProfileQualityEmailCapture
		}
		if input.CustomerServiceVisitorHash != "" {
			return VisitorProfileQualityCustomerService
		}
		return VisitorProfileQualityCartAction
	}
}

func normalizeVisitorLocale(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.Split(value, ",")[0]
	value = strings.Split(value, ";")[0]
	value = strings.ReplaceAll(value, "_", "-")
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeVisitorIP(value string) string {
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

func stableVisitorHash(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func normalizeVisitorProfileAdminFilter(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "all" {
		return ""
	}
	return value
}

func normalizeVisitorProfileStatusFilter(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", visitor.ProfileStatusActive:
		return visitor.ProfileStatusActive
	case "all":
		return "all"
	case visitor.ProfileStatusCandidate, visitor.ProfileStatusArchived, visitor.ProfileStatusSuppressed:
		return value
	default:
		return visitor.ProfileStatusActive
	}
}

func parseVisitorProfileTriState(value string) *bool {
	switch normalizeVisitorProfileAdminFilter(value) {
	case "yes", "true", "1", "linked", "captured", "with":
		result := true
		return &result
	case "no", "false", "0", "missing", "none", "without":
		result := false
		return &result
	default:
		return nil
	}
}

func visitorProfileLastSeenBoundary(value string) *time.Time {
	switch normalizeVisitorProfileAdminFilter(value) {
	case "24h", "today":
		boundary := time.Now().UTC().Add(-24 * time.Hour)
		return &boundary
	case "7d", "week":
		boundary := time.Now().UTC().AddDate(0, 0, -7)
		return &boundary
	case "30d", "month":
		boundary := time.Now().UTC().AddDate(0, 0, -30)
		return &boundary
	default:
		return nil
	}
}

func visitorProfileSnapshot(profile visitor.Profile) VisitorProfileSnapshot {
	identity := "anonymous"
	if profile.UserID != nil && *profile.UserID > 0 {
		identity = "account"
	}

	return VisitorProfileSnapshot{
		ID:                                profile.ID,
		UserID:                            profile.UserID,
		Identity:                          identity,
		CustomerServiceVisitorHashPreview: maskVisitorProfileToken(profile.CustomerServiceVisitorHash),
		HasCustomerServiceVisitor:         strings.TrimSpace(profile.CustomerServiceVisitorHash) != "",
		CartSessionID:                     strings.TrimSpace(profile.CartSessionID),
		HasCartSession:                    strings.TrimSpace(profile.CartSessionID) != "",
		Email:                             strings.TrimSpace(profile.Email),
		EmailSource:                       strings.TrimSpace(profile.EmailSource),
		HasEmail:                          strings.TrimSpace(profile.Email) != "",
		Locale:                            strings.TrimSpace(profile.Locale),
		LocaleSource:                      strings.TrimSpace(profile.LocaleSource),
		CountryCode:                       strings.TrimSpace(profile.CountryCode),
		Region:                            strings.TrimSpace(profile.Region),
		City:                              strings.TrimSpace(profile.City),
		Timezone:                          strings.TrimSpace(profile.Timezone),
		RegionLabel:                       visitorProfileRegionLabel(profile),
		HasIPFingerprint:                  strings.TrimSpace(profile.IPHash) != "",
		HasUserAgentFingerprint:           strings.TrimSpace(profile.UserAgentHash) != "",
		ProfileQualityScore:               profile.ProfileQualityScore,
		ProfileStatus:                     visitorProfileStatusOrActive(profile.ProfileStatus),
		LastMeaningfulAction:              strings.TrimSpace(profile.LastMeaningfulAction),
		FirstMeaningfulSeenAt:             profile.FirstMeaningfulSeenAt,
		LastMeaningfulSeenAt:              profile.LastMeaningfulSeenAt,
		RetentionUntil:                    profile.RetentionUntil,
		LastSeenAt:                        profile.LastSeenAt,
		CreatedAt:                         profile.CreatedAt,
		UpdatedAt:                         profile.UpdatedAt,
	}
}

func visitorProfileStatusOrActive(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return visitor.ProfileStatusActive
	}
	return value
}

func maskVisitorProfileToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 12 {
		return value
	}
	return value[:6] + "..." + value[len(value)-6:]
}

func visitorProfileRegionLabel(profile visitor.Profile) string {
	parts := []string{}
	for _, part := range []string{countryCodeDisplayName(profile.CountryCode), profile.Region, profile.City} {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, " / ")
}
