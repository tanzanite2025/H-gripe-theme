package order

import "time"

// PolicyDisclosure records the exact public policy representation shown during
// order creation so later evidence does not depend on the current policy page.
type PolicyDisclosure struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	OrderID         uint       `gorm:"not null;index;uniqueIndex:idx_order_policy_disclosure_key" json:"order_id"`
	PolicyKey       string     `gorm:"not null;uniqueIndex:idx_order_policy_disclosure_key" json:"policy_key"`
	Locale          string     `gorm:"not null" json:"locale"`
	RequestedLocale string     `gorm:"not null" json:"requested_locale"`
	Fallback        bool       `gorm:"not null;default:false" json:"fallback"`
	PolicyVersion   string     `gorm:"not null" json:"policy_version"`
	PolicyHash      string     `gorm:"not null" json:"policy_hash"`
	PolicyJSON      string     `gorm:"type:text;not null" json:"policy_json"`
	PolicyURL       string     `gorm:"type:text;not null" json:"policy_url"`
	DisclosedAt     time.Time  `gorm:"not null" json:"disclosed_at"`
	ConsentedAt     *time.Time `json:"consented_at,omitempty"`
	Source          string     `gorm:"not null" json:"source"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (PolicyDisclosure) TableName() string {
	return "order_policy_disclosures"
}
