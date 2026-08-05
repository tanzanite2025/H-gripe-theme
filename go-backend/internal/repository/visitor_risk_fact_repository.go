package repository

import (
	"encoding/json"
	"strings"
	"time"

	"tanzanite/internal/domain/visitor"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type VisitorRiskFactRepository struct {
	db *gorm.DB
}

func NewVisitorRiskFactRepository(db *gorm.DB) *VisitorRiskFactRepository {
	return &VisitorRiskFactRepository{db: db}
}

type VisitorRiskFactDelta struct {
	Day                   time.Time
	IPHash                string
	UserAgentHash         string
	CountryCode           string
	FirstSeenAt           time.Time
	LastSeenAt            time.Time
	RequestCount          int
	UniquePathCount       int
	UniqueAnonymousCount  int
	UniqueSessionCount    int
	InvalidRequestCount   int
	AuthFailureCount      int
	CheckoutFailureCount  int
	BotLikeUserAgentCount int
	NoCookieRequestCount  int
	MeaningfulActionCount int
	RiskScoreDelta        int
	SamplePaths           []string
	SamplePathLimit       int
}

type VisitorRiskFactListFilters struct {
	Search       string
	RiskLevel    string
	DayAfter     *time.Time
	MinRiskScore *int
}

type VisitorRiskFactStats struct {
	TotalFacts            int64
	NormalCount           int64
	WatchCount            int64
	SuspiciousCount       int64
	BlockCount            int64
	RequestCount          int64
	InvalidRequestCount   int64
	AuthFailureCount      int64
	CheckoutFailureCount  int64
	BotLikeUserAgentCount int64
	NoCookieRequestCount  int64
	MeaningfulActionCount int64
}

func (r *VisitorRiskFactRepository) FindFactByID(id uint) (visitor.RiskDailyFact, error) {
	if r == nil || r.db == nil {
		return visitor.RiskDailyFact{}, gorm.ErrInvalidDB
	}

	var fact visitor.RiskDailyFact
	if err := r.db.First(&fact, id).Error; err != nil {
		return visitor.RiskDailyFact{}, err
	}
	return fact, nil
}

func (r *VisitorRiskFactRepository) FindLatestByIdentity(dayAfter time.Time, ipHash, userAgentHash string) (*visitor.RiskDailyFact, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	ipHash = strings.TrimSpace(ipHash)
	userAgentHash = strings.TrimSpace(userAgentHash)
	if ipHash == "" {
		return nil, gorm.ErrRecordNotFound
	}

	query := r.db.Model(&visitor.RiskDailyFact{}).
		Where("ip_hash = ?", ipHash)
	if !dayAfter.IsZero() {
		query = query.Where("day >= ?", visitorRiskDay(dayAfter))
	}
	if userAgentHash != "" {
		query = query.Where("(user_agent_hash = ? OR user_agent_hash = '')", userAgentHash)
	} else {
		query = query.Where("user_agent_hash = ''")
	}

	var fact visitor.RiskDailyFact
	if err := query.Order("day DESC").Order("last_seen_at DESC").First(&fact).Error; err != nil {
		return nil, err
	}
	return &fact, nil
}

func (r *VisitorRiskFactRepository) CreateDecision(decision *visitor.RiskDecision) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if decision == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Create(decision).Error
}

func (r *VisitorRiskFactRepository) ListActiveDecisionsForIdentity(ipHash, userAgentHash string, now time.Time) ([]visitor.RiskDecision, error) {
	if r == nil || r.db == nil {
		return []visitor.RiskDecision{}, nil
	}
	ipHash = strings.TrimSpace(ipHash)
	userAgentHash = strings.TrimSpace(userAgentHash)
	if ipHash == "" {
		return []visitor.RiskDecision{}, nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	clauses := []string{"(scope = ? AND value_hash = ?)"}
	args := []interface{}{visitor.RiskDecisionScopeIPHash, ipHash}
	if userAgentHash != "" {
		if valueHash := visitor.RiskDecisionIPUAValueHash(ipHash, userAgentHash); valueHash != "" {
			clauses = append(clauses, "(scope = ? AND value_hash = ?)")
			args = append(args, visitor.RiskDecisionScopeIPUAHash, valueHash)
		}
	}

	var decisions []visitor.RiskDecision
	query := r.db.
		Where(strings.Join(clauses, " OR "), args...).
		Where("(expires_at IS NULL OR expires_at > ?)", now).
		Order("created_at DESC").
		Order("id DESC")
	if err := query.Find(&decisions).Error; err != nil {
		return nil, err
	}
	return decisions, nil
}

func (r *VisitorRiskFactRepository) ListActiveDecisionsForFacts(facts []visitor.RiskDailyFact, now time.Time) ([]visitor.RiskDecision, error) {
	if r == nil || r.db == nil || len(facts) == 0 {
		return []visitor.RiskDecision{}, nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	clauses := make([]string, 0, len(facts)*2)
	args := make([]interface{}, 0, len(facts)*4)
	for _, fact := range facts {
		if valueHash := visitor.RiskDecisionIPUAValueHash(fact.IPHash, fact.UserAgentHash); valueHash != "" {
			clauses = append(clauses, "(scope = ? AND value_hash = ?)")
			args = append(args, visitor.RiskDecisionScopeIPUAHash, valueHash)
		}
		if ipHash := strings.TrimSpace(fact.IPHash); ipHash != "" {
			clauses = append(clauses, "(scope = ? AND value_hash = ?)")
			args = append(args, visitor.RiskDecisionScopeIPHash, ipHash)
		}
	}
	if len(clauses) == 0 {
		return []visitor.RiskDecision{}, nil
	}

	var decisions []visitor.RiskDecision
	query := r.db.
		Where(strings.Join(clauses, " OR "), args...).
		Where("(expires_at IS NULL OR expires_at > ?)", now).
		Order("created_at DESC").
		Order("id DESC")
	if err := query.Find(&decisions).Error; err != nil {
		return nil, err
	}
	return decisions, nil
}

func (r *VisitorRiskFactRepository) UpsertDelta(delta VisitorRiskFactDelta) error {
	if r == nil || r.db == nil {
		return nil
	}

	delta = normalizeVisitorRiskDelta(delta)
	if delta.IPHash == "" {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var fact visitor.RiskDailyFact
		err := tx.
			Where("day = ? AND ip_hash = ? AND user_agent_hash = ?", delta.Day, delta.IPHash, delta.UserAgentHash).
			First(&fact).Error
		if err != nil {
			if !IsRecordNotFound(err) {
				return err
			}
			fact = visitor.RiskDailyFact{
				Day:                   delta.Day,
				IPHash:                delta.IPHash,
				UserAgentHash:         delta.UserAgentHash,
				CountryCode:           delta.CountryCode,
				FirstSeenAt:           delta.FirstSeenAt,
				LastSeenAt:            delta.LastSeenAt,
				RequestCount:          delta.RequestCount,
				UniquePathCount:       delta.UniquePathCount,
				UniqueAnonymousCount:  delta.UniqueAnonymousCount,
				UniqueSessionCount:    delta.UniqueSessionCount,
				InvalidRequestCount:   delta.InvalidRequestCount,
				AuthFailureCount:      delta.AuthFailureCount,
				CheckoutFailureCount:  delta.CheckoutFailureCount,
				BotLikeUserAgentCount: delta.BotLikeUserAgentCount,
				NoCookieRequestCount:  delta.NoCookieRequestCount,
				MeaningfulActionCount: delta.MeaningfulActionCount,
				RiskScore:             delta.RiskScoreDelta,
				RiskLevel:             visitorRiskLevel(delta.RiskScoreDelta),
				SamplePaths:           encodeVisitorRiskSamplePaths(delta.SamplePaths, delta.SamplePathLimit),
			}
			return tx.Create(&fact).Error
		}

		paths := mergeVisitorRiskSamplePaths(fact.SamplePaths, delta.SamplePaths, delta.SamplePathLimit)
		updates := map[string]interface{}{
			"last_seen_at":              maxVisitorRiskTime(fact.LastSeenAt, delta.LastSeenAt),
			"request_count":             fact.RequestCount + delta.RequestCount,
			"unique_path_count":         fact.UniquePathCount + delta.UniquePathCount,
			"unique_anonymous_count":    fact.UniqueAnonymousCount + delta.UniqueAnonymousCount,
			"unique_session_count":      fact.UniqueSessionCount + delta.UniqueSessionCount,
			"invalid_request_count":     fact.InvalidRequestCount + delta.InvalidRequestCount,
			"auth_failure_count":        fact.AuthFailureCount + delta.AuthFailureCount,
			"checkout_failure_count":    fact.CheckoutFailureCount + delta.CheckoutFailureCount,
			"bot_like_user_agent_count": fact.BotLikeUserAgentCount + delta.BotLikeUserAgentCount,
			"no_cookie_request_count":   fact.NoCookieRequestCount + delta.NoCookieRequestCount,
			"meaningful_action_count":   fact.MeaningfulActionCount + delta.MeaningfulActionCount,
			"risk_score":                fact.RiskScore + delta.RiskScoreDelta,
			"risk_level":                visitorRiskLevel(fact.RiskScore + delta.RiskScoreDelta),
			"sample_paths":              paths,
		}
		if !delta.FirstSeenAt.IsZero() && delta.FirstSeenAt.Before(fact.FirstSeenAt) {
			updates["first_seen_at"] = delta.FirstSeenAt
		}
		if strings.TrimSpace(fact.CountryCode) == "" && delta.CountryCode != "" {
			updates["country_code"] = delta.CountryCode
		}

		return tx.Model(&fact).Updates(updates).Error
	})
}

func (r *VisitorRiskFactRepository) DeleteBeforeDay(cutoffDay time.Time) (int64, error) {
	if r == nil || r.db == nil {
		return 0, nil
	}
	cutoffDay = visitorRiskDay(cutoffDay)
	result := r.db.
		Where("day < ?", cutoffDay).
		Delete(&visitor.RiskDailyFact{})
	return result.RowsAffected, result.Error
}

func (r *VisitorRiskFactRepository) DeleteExpiredDecisionsBefore(cutoff time.Time) (int64, error) {
	if r == nil || r.db == nil {
		return 0, nil
	}
	if cutoff.IsZero() {
		cutoff = time.Now().UTC()
	} else {
		cutoff = cutoff.UTC()
	}

	result := r.db.
		Where("expires_at IS NOT NULL AND expires_at < ?", cutoff).
		Delete(&visitor.RiskDecision{})
	return result.RowsAffected, result.Error
}

func (r *VisitorRiskFactRepository) List(page, pageSize int, filters VisitorRiskFactListFilters) ([]visitor.RiskDailyFact, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, nil
	}

	var facts []visitor.RiskDailyFact
	var total int64
	query := r.applyVisitorRiskFactFilters(r.db.Model(&visitor.RiskDailyFact{}), filters)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	err := query.
		Order("day DESC").
		Order("risk_score DESC").
		Order("last_seen_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&facts).Error
	return facts, total, err
}

func (r *VisitorRiskFactRepository) Stats(dayAfter *time.Time) (VisitorRiskFactStats, error) {
	if r == nil || r.db == nil {
		return VisitorRiskFactStats{}, nil
	}

	query := r.db.Model(&visitor.RiskDailyFact{})
	if dayAfter != nil {
		query = query.Where("day >= ?", visitorRiskDay(*dayAfter))
	}

	type aggregateResult struct {
		TotalFacts            int64
		RequestCount          int64
		InvalidRequestCount   int64
		AuthFailureCount      int64
		CheckoutFailureCount  int64
		BotLikeUserAgentCount int64
		NoCookieRequestCount  int64
		MeaningfulActionCount int64
	}
	var aggregate aggregateResult
	if err := query.Select(`
		COUNT(*) AS total_facts,
		COALESCE(SUM(request_count), 0) AS request_count,
		COALESCE(SUM(invalid_request_count), 0) AS invalid_request_count,
		COALESCE(SUM(auth_failure_count), 0) AS auth_failure_count,
		COALESCE(SUM(checkout_failure_count), 0) AS checkout_failure_count,
		COALESCE(SUM(bot_like_user_agent_count), 0) AS bot_like_user_agent_count,
		COALESCE(SUM(no_cookie_request_count), 0) AS no_cookie_request_count,
		COALESCE(SUM(meaningful_action_count), 0) AS meaningful_action_count
	`).Scan(&aggregate).Error; err != nil {
		return VisitorRiskFactStats{}, err
	}
	stats := VisitorRiskFactStats{
		TotalFacts:            aggregate.TotalFacts,
		RequestCount:          aggregate.RequestCount,
		InvalidRequestCount:   aggregate.InvalidRequestCount,
		AuthFailureCount:      aggregate.AuthFailureCount,
		CheckoutFailureCount:  aggregate.CheckoutFailureCount,
		BotLikeUserAgentCount: aggregate.BotLikeUserAgentCount,
		NoCookieRequestCount:  aggregate.NoCookieRequestCount,
		MeaningfulActionCount: aggregate.MeaningfulActionCount,
	}

	counts := map[string]*int64{
		visitor.RiskLevelNormal:     &stats.NormalCount,
		visitor.RiskLevelWatch:      &stats.WatchCount,
		visitor.RiskLevelSuspicious: &stats.SuspiciousCount,
		visitor.RiskLevelBlock:      &stats.BlockCount,
	}
	for level, target := range counts {
		levelQuery := r.db.Model(&visitor.RiskDailyFact{}).Where("risk_level = ?", level)
		if dayAfter != nil {
			levelQuery = levelQuery.Where("day >= ?", visitorRiskDay(*dayAfter))
		}
		if err := levelQuery.Count(target).Error; err != nil {
			return VisitorRiskFactStats{}, err
		}
	}

	return stats, nil
}

func (r *VisitorRiskFactRepository) applyVisitorRiskFactFilters(query *gorm.DB, filters VisitorRiskFactListFilters) *gorm.DB {
	if search := strings.ToLower(strings.TrimSpace(filters.Search)); search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"LOWER(ip_hash) LIKE ? OR LOWER(user_agent_hash) LIKE ? OR LOWER(country_code) LIKE ? OR CAST(id AS TEXT) LIKE ?",
			like, like, like, like,
		)
	}
	switch strings.ToLower(strings.TrimSpace(filters.RiskLevel)) {
	case "", "all":
	case visitor.RiskLevelNormal, visitor.RiskLevelWatch, visitor.RiskLevelSuspicious, visitor.RiskLevelBlock:
		query = query.Where("risk_level = ?", strings.ToLower(strings.TrimSpace(filters.RiskLevel)))
	}
	if filters.DayAfter != nil {
		query = query.Where("day >= ?", visitorRiskDay(*filters.DayAfter))
	}
	if filters.MinRiskScore != nil {
		query = query.Where("risk_score >= ?", *filters.MinRiskScore)
	}
	return query
}

func normalizeVisitorRiskDelta(delta VisitorRiskFactDelta) VisitorRiskFactDelta {
	delta.Day = visitorRiskDay(delta.Day)
	delta.IPHash = strings.TrimSpace(delta.IPHash)
	delta.UserAgentHash = strings.TrimSpace(delta.UserAgentHash)
	delta.CountryCode = strings.ToUpper(strings.TrimSpace(delta.CountryCode))
	if delta.FirstSeenAt.IsZero() {
		delta.FirstSeenAt = time.Now().UTC()
	}
	if delta.LastSeenAt.IsZero() || delta.LastSeenAt.Before(delta.FirstSeenAt) {
		delta.LastSeenAt = delta.FirstSeenAt
	}
	if delta.RequestCount < 0 {
		delta.RequestCount = 0
	}
	if delta.SamplePathLimit <= 0 {
		delta.SamplePathLimit = 8
	}
	return delta
}

func visitorRiskDay(value time.Time) time.Time {
	if value.IsZero() {
		value = time.Now().UTC()
	} else {
		value = value.UTC()
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func visitorRiskLevel(score int) string {
	switch {
	case score >= 70:
		return visitor.RiskLevelBlock
	case score >= 40:
		return visitor.RiskLevelSuspicious
	case score >= 20:
		return visitor.RiskLevelWatch
	default:
		return visitor.RiskLevelNormal
	}
}

func encodeVisitorRiskSamplePaths(paths []string, limit int) datatypes.JSON {
	merged := mergeVisitorRiskSamplePathValues(nil, paths, limit)
	encoded, err := json.Marshal(merged)
	if err != nil {
		return datatypes.JSON([]byte(`[]`))
	}
	return datatypes.JSON(encoded)
}

func mergeVisitorRiskSamplePaths(existing datatypes.JSON, next []string, limit int) datatypes.JSON {
	var current []string
	if len(existing) > 0 {
		_ = json.Unmarshal(existing, &current)
	}
	return encodeVisitorRiskSamplePaths(mergeVisitorRiskSamplePathValues(current, next, limit), limit)
}

func mergeVisitorRiskSamplePathValues(existing []string, next []string, limit int) []string {
	if limit <= 0 {
		limit = 8
	}
	seen := make(map[string]struct{}, len(existing)+len(next))
	result := make([]string, 0, limit)
	for _, path := range append(existing, next...) {
		path = normalizeVisitorRiskSamplePath(path)
		if path == "" {
			continue
		}
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		result = append(result, path)
		if len(result) >= limit {
			break
		}
	}
	return result
}

func normalizeVisitorRiskSamplePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if len(path) > 180 {
		path = path[:180]
	}
	return path
}

func maxVisitorRiskTime(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() || a.After(b) {
		return a
	}
	return b
}
