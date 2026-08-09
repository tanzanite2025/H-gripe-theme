package repository

import (
	"strings"
	"tanzanite/internal/domain/recommendation"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RecommendationSignalQuery struct {
	ProductIDs  []uint
	AnonymousID string
	SessionID   string
	Since       time.Time
}

type RecommendationProductSignal struct {
	ProductID            uint
	ProductViews         int64
	ProductDwells        int64
	RecommendationClicks int64
	CartAdds             int64
	WishlistAdds         int64
	CheckoutStarts       int64
	Purchases            int64
}

type RecommendationEventRepository struct {
	db *gorm.DB
}

func NewRecommendationEventRepository(db *gorm.DB) *RecommendationEventRepository {
	return &RecommendationEventRepository{db: db}
}

// CreateBatch is idempotent on the client-generated event_id.
func (r *RecommendationEventRepository) CreateBatch(events []recommendation.Event) (int64, error) {
	if len(events) == 0 {
		return 0, nil
	}

	result := r.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "event_id"}},
			DoNothing: true,
		}).
		CreateInBatches(&events, 50)

	return result.RowsAffected, result.Error
}

func (r *RecommendationEventRepository) ListProductSignals(input RecommendationSignalQuery) (map[uint]RecommendationProductSignal, error) {
	signals := map[uint]RecommendationProductSignal{}
	if r == nil || r.db == nil {
		return signals, nil
	}

	productIDs := uniqueRecommendationSignalProductIDs(input.ProductIDs)
	if len(productIDs) == 0 {
		return signals, nil
	}

	since := input.Since
	if since.IsZero() {
		since = time.Now().UTC().AddDate(0, 0, -30)
	}

	query := r.db.Model(&recommendation.Event{}).
		Select(`
			product_id,
			COALESCE(SUM(CASE WHEN event_type = 'product_view' THEN 1 ELSE 0 END), 0) AS product_views,
			COALESCE(SUM(CASE WHEN event_type = 'product_dwell' THEN 1 ELSE 0 END), 0) AS product_dwells,
			COALESCE(SUM(CASE WHEN event_type = 'recommendation_click' THEN 1 ELSE 0 END), 0) AS recommendation_clicks,
			COALESCE(SUM(CASE WHEN event_type = 'add_to_cart' THEN 1 ELSE 0 END), 0) AS cart_adds,
			COALESCE(SUM(CASE WHEN event_type = 'wishlist_add' THEN 1 ELSE 0 END), 0) AS wishlist_adds,
			COALESCE(SUM(CASE WHEN event_type = 'begin_checkout' THEN 1 ELSE 0 END), 0) AS checkout_starts,
			COALESCE(SUM(CASE WHEN event_type = 'purchase' THEN 1 ELSE 0 END), 0) AS purchases
		`).
		Where("product_id IN ?", productIDs).
		Where("occurred_at >= ?", since.UTC()).
		Group("product_id")

	identityConditions := []string{}
	identityArgs := []interface{}{}
	if anonymousID := strings.TrimSpace(input.AnonymousID); anonymousID != "" {
		identityConditions = append(identityConditions, "anonymous_id = ?")
		identityArgs = append(identityArgs, anonymousID)
	}
	if sessionID := strings.TrimSpace(input.SessionID); sessionID != "" {
		identityConditions = append(identityConditions, "session_id = ?")
		identityArgs = append(identityArgs, sessionID)
	}
	if len(identityConditions) > 0 {
		query = query.Where("("+strings.Join(identityConditions, " OR ")+")", identityArgs...)
	}

	var rows []RecommendationProductSignal
	if err := query.Scan(&rows).Error; err != nil {
		return signals, err
	}
	for _, row := range rows {
		if row.ProductID > 0 {
			signals[row.ProductID] = row
		}
	}
	return signals, nil
}

func (r *RecommendationEventRepository) DeleteExpiredByTypes(eventTypes []string, cutoff time.Time, batchLimit int) (int64, error) {
	if r == nil || r.db == nil || len(eventTypes) == 0 {
		return 0, nil
	}
	if cutoff.IsZero() {
		return 0, nil
	}
	if batchLimit <= 0 {
		batchLimit = 5000
	}

	var ids []uint
	if err := r.db.Model(&recommendation.Event{}).
		Select("id").
		Where("event_type IN ?", eventTypes).
		Where("occurred_at < ?", cutoff.UTC()).
		Order("occurred_at ASC").
		Limit(batchLimit).
		Find(&ids).Error; err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	result := r.db.Where("id IN ?", ids).Delete(&recommendation.Event{})
	return result.RowsAffected, result.Error
}

func uniqueRecommendationSignalProductIDs(productIDs []uint) []uint {
	if len(productIDs) == 0 {
		return nil
	}
	seen := make(map[uint]struct{}, len(productIDs))
	unique := make([]uint, 0, len(productIDs))
	for _, productID := range productIDs {
		if productID == 0 {
			continue
		}
		if _, ok := seen[productID]; ok {
			continue
		}
		seen[productID] = struct{}{}
		unique = append(unique, productID)
	}
	return unique
}
