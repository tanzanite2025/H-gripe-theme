package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strings"
	"time"

	"tanzanite/internal/domain/visitor"
	"tanzanite/internal/repository"
)

type VisitorProfileService struct {
	visitorRepo *repository.VisitorProfileRepository
}

func NewVisitorProfileService(visitorRepo *repository.VisitorProfileRepository) *VisitorProfileService {
	return &VisitorProfileService{visitorRepo: visitorRepo}
}

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
}

type VisitorProfileSnapshot struct {
	ID                                uint      `json:"id"`
	UserID                            *uint     `json:"user_id,omitempty"`
	Identity                          string    `json:"identity"`
	CustomerServiceVisitorHashPreview string    `json:"customer_service_visitor_hash_preview,omitempty"`
	HasCustomerServiceVisitor         bool      `json:"has_customer_service_visitor"`
	CartSessionID                     string    `json:"cart_session_id,omitempty"`
	HasCartSession                    bool      `json:"has_cart_session"`
	Email                             string    `json:"email,omitempty"`
	EmailSource                       string    `json:"email_source,omitempty"`
	HasEmail                          bool      `json:"has_email"`
	Locale                            string    `json:"locale,omitempty"`
	LocaleSource                      string    `json:"locale_source,omitempty"`
	CountryCode                       string    `json:"country_code,omitempty"`
	Region                            string    `json:"region,omitempty"`
	City                              string    `json:"city,omitempty"`
	Timezone                          string    `json:"timezone,omitempty"`
	RegionLabel                       string    `json:"region_label,omitempty"`
	HasIPFingerprint                  bool      `json:"has_ip_fingerprint"`
	HasUserAgentFingerprint           bool      `json:"has_user_agent_fingerprint"`
	LastSeenAt                        time.Time `json:"last_seen_at"`
	CreatedAt                         time.Time `json:"created_at"`
	UpdatedAt                         time.Time `json:"updated_at"`
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
}

func (s *VisitorProfileService) Touch(input VisitorProfileTouchInput) (*visitor.Profile, error) {
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
		profile = &visitor.Profile{}
		applyVisitorProfileTouch(profile, input)
		if err := s.visitorRepo.Create(profile); err != nil {
			return nil, err
		}
		return profile, nil
	}

	applyVisitorProfileTouch(profile, input)
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
	input.Timezone = strings.TrimSpace(input.Timezone)
	input.IPAddress = normalizeVisitorIP(input.IPAddress)
	input.UserAgent = strings.TrimSpace(input.UserAgent)
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
		LastSeenAt:                        profile.LastSeenAt,
		CreatedAt:                         profile.CreatedAt,
		UpdatedAt:                         profile.UpdatedAt,
	}
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
