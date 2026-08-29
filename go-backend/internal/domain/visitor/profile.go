package visitor

import (
	"time"

	"gorm.io/gorm"
)

const (
	ProfileStatusCandidate  = "candidate"
	ProfileStatusActive     = "active"
	ProfileStatusArchived   = "archived"
	ProfileStatusSuppressed = "suppressed"
)

type Profile struct {
	ID                         uint           `gorm:"primarykey" json:"id"`
	UserID                     *uint          `gorm:"index" json:"user_id,omitempty"`
	CustomerServiceVisitorHash string         `gorm:"type:varchar(64);index" json:"customer_service_visitor_hash,omitempty"`
	CartSessionID              string         `gorm:"type:varchar(64);index" json:"cart_session_id,omitempty"`
	Email                      string         `gorm:"type:varchar(255);index" json:"email,omitempty"`
	EmailSource                string         `gorm:"type:varchar(40)" json:"email_source,omitempty"`
	Locale                     string         `gorm:"type:varchar(20);index" json:"locale,omitempty"`
	LocaleSource               string         `gorm:"type:varchar(40)" json:"locale_source,omitempty"`
	CountryCode                string         `gorm:"type:varchar(8);index" json:"country_code,omitempty"`
	Region                     string         `gorm:"type:varchar(80)" json:"region,omitempty"`
	City                       string         `gorm:"type:varchar(80)" json:"city,omitempty"`
	Timezone                   string         `gorm:"type:varchar(80)" json:"timezone,omitempty"`
	IPAddress                  string         `gorm:"type:varchar(45);index" json:"-"`
	IPHash                     string         `gorm:"type:varchar(64);index" json:"-"`
	UserAgentHash              string         `gorm:"type:varchar(64)" json:"-"`
	LastSeenAt                 time.Time      `gorm:"index" json:"last_seen_at"`
	ProfileQualityScore        int            `gorm:"not null;default:0;index" json:"profile_quality_score"`
	ProfileStatus              string         `gorm:"type:varchar(24);not null;default:'active';index" json:"profile_status"`
	LastMeaningfulAction       string         `gorm:"type:varchar(64)" json:"last_meaningful_action,omitempty"`
	FirstMeaningfulSeenAt      *time.Time     `gorm:"index" json:"first_meaningful_seen_at,omitempty"`
	LastMeaningfulSeenAt       *time.Time     `gorm:"index" json:"last_meaningful_seen_at,omitempty"`
	RetentionUntil             *time.Time     `gorm:"index" json:"retention_until,omitempty"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
	DeletedAt                  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Profile) TableName() string {
	return "visitor_profiles"
}
