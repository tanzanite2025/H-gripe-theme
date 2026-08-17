package service

import (
	"commerce-platform/internal/domain/feedback"
	"commerce-platform/internal/pkg/ugc"
	"commerce-platform/internal/repository"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"gorm.io/gorm"
)

var (
	ErrFeedbackMissingThread     = errors.New("thread is required")
	ErrFeedbackMissingContent    = errors.New("content is required")
	ErrFeedbackContentTooLong    = errors.New("feedback content is too long")
	ErrFeedbackNameTooLong       = errors.New("feedback name is too long")
	ErrFeedbackThreadTooLong     = errors.New("feedback thread is too long")
	ErrFeedbackInvalidThread     = errors.New("feedback thread contains invalid characters")
	ErrFeedbackEmailTooLong      = errors.New("feedback email is too long")
	ErrFeedbackLocaleTooLong     = errors.New("feedback locale is too long")
	ErrFeedbackInvalidLocale     = errors.New("feedback locale contains invalid characters")
	ErrFeedbackPagePathTooLong   = errors.New("feedback page path is too long")
	ErrFeedbackPageTitleTooLong  = errors.New("feedback page title is too long")
	ErrFeedbackSourceHashTooLong = errors.New("feedback source hash is too long")
	ErrFeedbackInvalidStatus     = errors.New("invalid feedback status")
	ErrFeedbackNotFound          = errors.New("feedback not found")
)

const (
	maxFeedbackThreadKeyRunes  = 160
	maxFeedbackContentRunes    = 3000
	maxFeedbackNameRunes       = 120
	maxFeedbackEmailRunes      = 254
	maxFeedbackLocaleRunes     = 32
	maxFeedbackPagePathRunes   = 512
	maxFeedbackPageTitleRunes  = 240
	maxFeedbackSourceHashRunes = 80
)

var (
	feedbackThreadKeyPattern = regexp.MustCompile(`^[A-Za-z0-9/_-][A-Za-z0-9:_./-]*$`)
	feedbackLocalePattern    = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

type FeedbackService struct {
	feedbackRepo *repository.FeedbackRepository
}

type FeedbackAdminListFilter struct {
	Status    string
	ThreadKey string
	PagePath  string
	Search    string
	Page      int
	PageSize  int
}

type FeedbackAdminUpdateInput struct {
	Status       string
	ReplyContent string
	AdminID      uint
}

type FeedbackRiskOverviewInput struct {
	WindowHours int
	GeneratedAt time.Time
	RateLimit   FeedbackRateLimitSnapshot
}

type FeedbackRateLimitSnapshot struct {
	WindowHours       int   `json:"window_hours"`
	Total             int64 `json:"total"`
	ReadIP            int64 `json:"read_ip"`
	WriteIP           int64 `json:"write_ip"`
	WriteUser         int64 `json:"write_user"`
	FallbackTotal     int64 `json:"fallback_total"`
	FallbackReadIP    int64 `json:"fallback_read_ip"`
	FallbackWriteIP   int64 `json:"fallback_write_ip"`
	FallbackWriteUser int64 `json:"fallback_write_user"`
	RedisUnavailable  int64 `json:"redis_unavailable"`
	Unavailable       bool  `json:"unavailable"`
}

type FeedbackRiskOverview struct {
	WindowHours  int                       `json:"window_hours"`
	GeneratedAt  time.Time                 `json:"generated_at"`
	Level        string                    `json:"level"`
	Totals       FeedbackRiskTotals        `json:"totals"`
	RateLimit    FeedbackRateLimitSnapshot `json:"rate_limit"`
	HotPages     []FeedbackRiskPage        `json:"hot_pages"`
	SourceBursts []FeedbackRiskSource      `json:"source_bursts"`
}

type FeedbackRiskTotals struct {
	PendingTotal       int64            `json:"pending_total"`
	PendingOver24Hours int64            `json:"pending_over_24h"`
	WindowTotal        int64            `json:"window_total"`
	LastHourTotal      int64            `json:"last_hour_total"`
	ByStatus           map[string]int64 `json:"by_status"`
}

type FeedbackRiskPage struct {
	PagePath       string    `json:"page_path"`
	PageTitle      string    `json:"page_title"`
	ThreadKey      string    `json:"thread_key"`
	FilterKind     string    `json:"filter_kind"`
	FilterValue    string    `json:"filter_value"`
	FeedbackCount  int64     `json:"feedback_count"`
	PendingCount   int64     `json:"pending_count"`
	LastFeedbackAt time.Time `json:"last_feedback_at"`
}

type FeedbackRiskSource struct {
	SourceHashPreview string    `json:"source_hash_preview"`
	FeedbackCount     int64     `json:"feedback_count"`
	PageCount         int64     `json:"page_count"`
	PendingCount      int64     `json:"pending_count"`
	LastFeedbackAt    time.Time `json:"last_feedback_at"`
}

func NewFeedbackService(feedbackRepo *repository.FeedbackRepository) *FeedbackService {
	return &FeedbackService{feedbackRepo: feedbackRepo}
}

func (s *FeedbackService) List(threadKey, status, search string, page, pageSize int) ([]feedback.Feedback, int64, error) {
	normalizedThreadKey, err := normalizeFeedbackThreadKey(threadKey)
	if err != nil {
		return nil, 0, err
	}
	if status != "" && !validFeedbackStatus(status) {
		return nil, 0, ErrFeedbackInvalidStatus
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.feedbackRepo.List(normalizedThreadKey, status, strings.TrimSpace(search), page, pageSize)
}

func (s *FeedbackService) ListPublic(threadKey, search string, page, pageSize int) ([]feedback.Feedback, int64, error) {
	items, total, err := s.List(threadKey, "approved", search, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for index := range items {
		items[index].Name = normalizeFeedbackPublicText(items[index].Name)
		items[index].Content = normalizeFeedbackPublicText(items[index].Content)
		items[index].ReplyContent = normalizeFeedbackPublicText(items[index].ReplyContent)
	}
	return items, total, nil
}

func (s *FeedbackService) Create(item *feedback.Feedback) error {
	threadKey, err := normalizeFeedbackThreadKey(item.ThreadKey)
	if err != nil {
		return err
	}
	content, err := ugc.PlainText(item.Content, maxFeedbackContentRunes)
	if errors.Is(err, ugc.ErrTextTooLong) {
		return ErrFeedbackContentTooLong
	}
	if err != nil {
		return err
	}
	name, err := ugc.PlainText(item.Name, maxFeedbackNameRunes)
	if errors.Is(err, ugc.ErrTextTooLong) {
		return ErrFeedbackNameTooLong
	}
	if err != nil {
		return err
	}
	email, err := normalizeFeedbackMetadata(item.Email, maxFeedbackEmailRunes, ErrFeedbackEmailTooLong)
	if err != nil {
		return err
	}
	locale, err := normalizeFeedbackLocale(item.Locale)
	if err != nil {
		return err
	}
	pagePath, err := normalizeFeedbackPlainMetadata(item.PagePath, maxFeedbackPagePathRunes, ErrFeedbackPagePathTooLong)
	if err != nil {
		return err
	}
	pageTitle, err := normalizeFeedbackPlainMetadata(item.PageTitle, maxFeedbackPageTitleRunes, ErrFeedbackPageTitleTooLong)
	if err != nil {
		return err
	}
	sourceHash, err := normalizeFeedbackMetadata(item.SourceHash, maxFeedbackSourceHashRunes, ErrFeedbackSourceHashTooLong)
	if err != nil {
		return err
	}
	item.ThreadKey = threadKey
	item.Content = content
	item.Name = name
	item.Email = email
	item.Locale = locale
	item.PagePath = pagePath
	item.PageTitle = pageTitle
	item.SourceHash = sourceHash

	if item.Content == "" {
		return ErrFeedbackMissingContent
	}
	if item.Status == "" {
		item.Status = "pending"
	}
	if !validFeedbackStatus(item.Status) {
		return ErrFeedbackInvalidStatus
	}

	return s.feedbackRepo.Create(item)
}

func (s *FeedbackService) UpdateStatus(id uint, status string) error {
	status = strings.TrimSpace(status)
	if !validFeedbackStatus(status) {
		return ErrFeedbackInvalidStatus
	}
	return s.feedbackRepo.UpdateStatus(id, status)
}

func (s *FeedbackService) ListAdmin(filter FeedbackAdminListFilter) ([]feedback.Feedback, int64, error) {
	filter.Status = strings.TrimSpace(filter.Status)
	if filter.Status == "" {
		filter.Status = "pending"
	}
	if filter.Status != "all" && !validFeedbackStatus(filter.Status) {
		return nil, 0, ErrFeedbackInvalidStatus
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	return s.feedbackRepo.ListAdmin(repository.FeedbackAdminListFilter{
		Status:    filter.Status,
		ThreadKey: strings.TrimSpace(filter.ThreadKey),
		PagePath:  strings.TrimSpace(filter.PagePath),
		Search:    strings.TrimSpace(filter.Search),
		Page:      filter.Page,
		PageSize:  filter.PageSize,
	})
}

func (s *FeedbackService) GetAdmin(id uint) (*feedback.Feedback, error) {
	item, err := s.feedbackRepo.Get(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFeedbackNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *FeedbackService) UpdateAdmin(id uint, input FeedbackAdminUpdateInput) (*feedback.Feedback, error) {
	item, err := s.GetAdmin(id)
	if err != nil {
		return nil, err
	}

	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = item.Status
	}
	if !validFeedbackStatus(status) {
		return nil, ErrFeedbackInvalidStatus
	}

	replyContent, err := ugc.PlainText(input.ReplyContent, 3000)
	if errors.Is(err, ugc.ErrTextTooLong) {
		return nil, ErrFeedbackContentTooLong
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	item.Status = status
	item.ReviewedAt = &now
	item.ReviewedBy = input.AdminID
	item.ReplyContent = replyContent
	if replyContent != "" {
		item.RepliedAt = &now
		item.RepliedBy = input.AdminID
	} else {
		item.RepliedAt = nil
		item.RepliedBy = 0
	}

	if err := s.feedbackRepo.UpdateAdmin(item); err != nil {
		return nil, err
	}
	return s.GetAdmin(id)
}

func (s *FeedbackService) RiskOverview(input FeedbackRiskOverviewInput) (*FeedbackRiskOverview, error) {
	windowHours := input.WindowHours
	if windowHours < 1 || windowHours > 168 {
		windowHours = 24
	}
	generatedAt := input.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}
	generatedAt = generatedAt.UTC()
	summary, err := s.feedbackRepo.RiskSummary(
		generatedAt.Add(-time.Duration(windowHours)*time.Hour),
		generatedAt.Add(-24*time.Hour),
		generatedAt.Add(-time.Hour),
		5,
		5,
	)
	if err != nil {
		return nil, err
	}

	overview := &FeedbackRiskOverview{
		WindowHours: windowHours,
		GeneratedAt: generatedAt,
		Totals: FeedbackRiskTotals{
			PendingTotal:       summary.PendingTotal,
			PendingOver24Hours: summary.PendingOver24Hours,
			WindowTotal:        summary.WindowTotal,
			LastHourTotal:      summary.LastHourTotal,
			ByStatus:           mapFeedbackStatusCounts(summary.StatusCounts),
		},
		RateLimit:    input.RateLimit,
		HotPages:     mapFeedbackRiskPages(summary.HotPages),
		SourceBursts: mapFeedbackRiskSources(summary.SourceBursts),
	}
	overview.RateLimit.WindowHours = windowHours
	overview.Level = feedbackRiskLevel(overview)
	return overview, nil
}

func validFeedbackStatus(status string) bool {
	switch status {
	case "pending", "approved", "rejected", "hidden":
		return true
	default:
		return false
	}
}

func normalizeFeedbackThreadKey(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", ErrFeedbackMissingThread
	}
	if utf8.RuneCountInString(normalized) > maxFeedbackThreadKeyRunes {
		return "", ErrFeedbackThreadTooLong
	}
	if !feedbackThreadKeyPattern.MatchString(normalized) {
		return "", ErrFeedbackInvalidThread
	}
	return normalized, nil
}

func mapFeedbackStatusCounts(items []repository.FeedbackStatusCount) map[string]int64 {
	counts := map[string]int64{
		"pending":  0,
		"approved": 0,
		"rejected": 0,
		"hidden":   0,
	}
	for _, item := range items {
		status := strings.TrimSpace(item.Status)
		if status == "" {
			continue
		}
		counts[status] = item.Count
	}
	return counts
}

func mapFeedbackRiskPages(items []repository.FeedbackRiskPage) []FeedbackRiskPage {
	pages := make([]FeedbackRiskPage, 0, len(items))
	for _, item := range items {
		pages = append(pages, FeedbackRiskPage{
			PagePath:       item.PagePath,
			PageTitle:      item.PageTitle,
			ThreadKey:      item.ThreadKey,
			FilterKind:     normalizeFeedbackRiskFilterKind(item.FilterKind),
			FilterValue:    normalizeFeedbackRiskFilterValue(item.FilterValue, item.PagePath, item.ThreadKey),
			FeedbackCount:  item.FeedbackCount,
			PendingCount:   item.PendingCount,
			LastFeedbackAt: parseFeedbackRiskTime(item.LastFeedbackAt.String),
		})
	}
	return pages
}

func mapFeedbackRiskSources(items []repository.FeedbackRiskSource) []FeedbackRiskSource {
	sources := make([]FeedbackRiskSource, 0, len(items))
	for _, item := range items {
		sources = append(sources, FeedbackRiskSource{
			SourceHashPreview: feedbackSourceHashPreview(item.SourceHash),
			FeedbackCount:     item.FeedbackCount,
			PageCount:         item.PageCount,
			PendingCount:      item.PendingCount,
			LastFeedbackAt:    parseFeedbackRiskTime(item.LastFeedbackAt.String),
		})
	}
	return sources
}

func feedbackSourceHashPreview(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}

func normalizeFeedbackRiskFilterKind(value string) string {
	switch strings.TrimSpace(value) {
	case "page_path":
		return "page_path"
	default:
		return "thread_key"
	}
}

func normalizeFeedbackRiskFilterValue(value, pagePath, threadKey string) string {
	if normalized := strings.TrimSpace(value); normalized != "" {
		return normalized
	}
	if normalized := strings.TrimSpace(pagePath); normalized != "" {
		return normalized
	}
	return strings.TrimSpace(threadKey)
}

func parseFeedbackRiskTime(value string) time.Time {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return time.Time{}
	}
	for _, candidate := range []string{
		normalized,
		strings.Replace(normalized, " ", "T", 1),
	} {
		if parsed, err := time.Parse(time.RFC3339Nano, candidate); err == nil {
			return parsed.UTC()
		}
	}
	for _, layout := range []string{
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	} {
		if parsed, err := time.ParseInLocation(layout, normalized, time.UTC); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}

func feedbackRiskLevel(overview *FeedbackRiskOverview) string {
	if overview == nil {
		return "normal"
	}
	if overview.RateLimit.Total >= 20 ||
		overview.RateLimit.RedisUnavailable >= 20 ||
		overview.Totals.LastHourTotal >= 20 ||
		overview.Totals.PendingOver24Hours >= 20 {
		return "critical"
	}
	for _, source := range overview.SourceBursts {
		if source.PageCount >= 3 && source.FeedbackCount >= 5 {
			return "critical"
		}
	}
	if overview.RateLimit.Total > 0 ||
		overview.RateLimit.FallbackTotal > 0 ||
		overview.RateLimit.RedisUnavailable > 0 ||
		overview.RateLimit.Unavailable ||
		overview.Totals.LastHourTotal >= 8 ||
		overview.Totals.PendingOver24Hours >= 5 {
		return "warning"
	}
	for _, source := range overview.SourceBursts {
		if source.PageCount >= 2 && source.FeedbackCount >= 3 {
			return "warning"
		}
	}
	return "normal"
}

func normalizeFeedbackLocale(value string) (string, error) {
	normalized, err := normalizeFeedbackMetadata(value, maxFeedbackLocaleRunes, ErrFeedbackLocaleTooLong)
	if err != nil || normalized == "" {
		return normalized, err
	}
	if !feedbackLocalePattern.MatchString(normalized) {
		return "", ErrFeedbackInvalidLocale
	}
	return normalized, nil
}

func normalizeFeedbackPlainMetadata(value string, maxRunes int, tooLongErr error) (string, error) {
	normalized, err := ugc.PlainText(value, maxRunes)
	if errors.Is(err, ugc.ErrTextTooLong) {
		return "", tooLongErr
	}
	if err != nil {
		return "", err
	}
	return stripFeedbackControlRunes(strings.TrimSpace(normalized)), nil
}

func normalizeFeedbackMetadata(value string, maxRunes int, tooLongErr error) (string, error) {
	normalized := stripFeedbackControlRunes(strings.TrimSpace(value))
	if normalized == "" {
		return "", nil
	}
	if utf8.RuneCountInString(normalized) > maxRunes {
		return "", tooLongErr
	}
	return normalized, nil
}

func stripFeedbackControlRunes(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, value)
}

func normalizeFeedbackPublicText(value string) string {
	normalized, err := ugc.PlainText(value, 0)
	if err != nil {
		return ""
	}
	return normalized
}
