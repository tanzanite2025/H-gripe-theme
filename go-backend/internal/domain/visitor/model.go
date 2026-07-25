package visitor

import (
	"time"

	"gorm.io/gorm"
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
	IPHash                     string         `gorm:"type:varchar(64);index" json:"-"`
	UserAgentHash              string         `gorm:"type:varchar(64)" json:"-"`
	LastSeenAt                 time.Time      `gorm:"index" json:"last_seen_at"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
	DeletedAt                  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Profile) TableName() string {
	return "visitor_profiles"
}
