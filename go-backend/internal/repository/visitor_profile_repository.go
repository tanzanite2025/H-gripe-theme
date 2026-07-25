package repository

import (
	"strings"
	"tanzanite/internal/domain/visitor"
	"time"

	"gorm.io/gorm"
)

type VisitorProfileRepository struct {
	db *gorm.DB
}

func NewVisitorProfileRepository(db *gorm.DB) *VisitorProfileRepository {
	return &VisitorProfileRepository{db: db}
}

type VisitorProfileListFilters struct {
	Search                    string
	Identity                  string
	CountryCode               string
	Locale                    string
	HasEmail                  *bool
	HasCartSession            *bool
	HasCustomerServiceVisitor *bool
	LastSeenAfter             *time.Time
}

func (r *VisitorProfileRepository) FindByCustomerServiceVisitorHash(hash string) (*visitor.Profile, error) {
	var profile visitor.Profile
	err := r.db.Where("customer_service_visitor_hash = ?", strings.TrimSpace(hash)).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *VisitorProfileRepository) FindByCartSessionID(sessionID string) (*visitor.Profile, error) {
	var profile visitor.Profile
	err := r.db.Where("cart_session_id = ?", strings.TrimSpace(sessionID)).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *VisitorProfileRepository) FindByUserID(userID uint) (*visitor.Profile, error) {
	var profile visitor.Profile
	err := r.db.Where("user_id = ?", userID).Order("updated_at DESC").First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *VisitorProfileRepository) Create(profile *visitor.Profile) error {
	return r.db.Create(profile).Error
}

func (r *VisitorProfileRepository) Update(profile *visitor.Profile) error {
	return r.db.Save(profile).Error
}

func (r *VisitorProfileRepository) List(page, pageSize int, filters VisitorProfileListFilters) ([]visitor.Profile, int64, error) {
	var profiles []visitor.Profile
	var total int64

	query := r.applyListFilters(r.db.Model(&visitor.Profile{}), filters)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("last_seen_at DESC").
		Order("updated_at DESC").
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&profiles).Error

	return profiles, total, err
}

func (r *VisitorProfileRepository) Stats() (map[string]int64, error) {
	stats := map[string]int64{}

	counts := map[string]*gorm.DB{
		"total":                  r.db.Model(&visitor.Profile{}),
		"account_count":          r.db.Model(&visitor.Profile{}).Where("user_id IS NOT NULL"),
		"anonymous_count":        r.db.Model(&visitor.Profile{}).Where("user_id IS NULL"),
		"email_count":            r.db.Model(&visitor.Profile{}).Where("email IS NOT NULL AND email <> ''"),
		"cart_linked_count":      r.db.Model(&visitor.Profile{}).Where("cart_session_id IS NOT NULL AND cart_session_id <> ''"),
		"customer_service_count": r.db.Model(&visitor.Profile{}).Where("customer_service_visitor_hash IS NOT NULL AND customer_service_visitor_hash <> ''"),
		"region_count":           r.db.Model(&visitor.Profile{}).Where("country_code IS NOT NULL AND country_code <> ''"),
		"recent_24h_count":       r.db.Model(&visitor.Profile{}).Where("last_seen_at >= ?", time.Now().UTC().Add(-24*time.Hour)),
	}

	for key, query := range counts {
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return nil, err
		}
		stats[key] = count
	}

	return stats, nil
}

func (r *VisitorProfileRepository) applyListFilters(query *gorm.DB, filters VisitorProfileListFilters) *gorm.DB {
	if search := strings.ToLower(strings.TrimSpace(filters.Search)); search != "" {
		like := "%" + search + "%"
		query = query.Where(
			`LOWER(email) LIKE ? OR LOWER(customer_service_visitor_hash) LIKE ? OR LOWER(cart_session_id) LIKE ? OR LOWER(locale) LIKE ? OR LOWER(country_code) LIKE ? OR LOWER(region) LIKE ? OR LOWER(city) LIKE ? OR CAST(id AS TEXT) LIKE ? OR CAST(user_id AS TEXT) LIKE ?`,
			like, like, like, like, like, like, like, like, like,
		)
	}

	switch strings.ToLower(strings.TrimSpace(filters.Identity)) {
	case "account", "member", "user":
		query = query.Where("user_id IS NOT NULL")
	case "anonymous", "visitor", "guest":
		query = query.Where("user_id IS NULL")
	}

	if countryCode := strings.ToUpper(strings.TrimSpace(filters.CountryCode)); countryCode != "" {
		query = query.Where("country_code = ?", countryCode)
	}

	if locale := strings.ToLower(strings.TrimSpace(filters.Locale)); locale != "" {
		query = query.Where("LOWER(locale) = ?", locale)
	}

	if filters.HasEmail != nil {
		if *filters.HasEmail {
			query = query.Where("email IS NOT NULL AND email <> ''")
		} else {
			query = query.Where("(email IS NULL OR email = '')")
		}
	}

	if filters.HasCartSession != nil {
		if *filters.HasCartSession {
			query = query.Where("cart_session_id IS NOT NULL AND cart_session_id <> ''")
		} else {
			query = query.Where("(cart_session_id IS NULL OR cart_session_id = '')")
		}
	}

	if filters.HasCustomerServiceVisitor != nil {
		if *filters.HasCustomerServiceVisitor {
			query = query.Where("customer_service_visitor_hash IS NOT NULL AND customer_service_visitor_hash <> ''")
		} else {
			query = query.Where("(customer_service_visitor_hash IS NULL OR customer_service_visitor_hash = '')")
		}
	}

	if filters.LastSeenAfter != nil {
		query = query.Where("last_seen_at >= ?", *filters.LastSeenAfter)
	}

	return query
}
