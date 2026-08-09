package market

import (
	"time"

	"gorm.io/gorm"
)

type StorefrontMarket struct {
	ID                  uint            `gorm:"primarykey" json:"id"`
	Code                string          `gorm:"size:32;not null;index" json:"code"`
	Name                string          `gorm:"size:120;not null" json:"name"`
	DefaultLocale       string          `gorm:"size:32;not null" json:"default_locale"`
	SupportedLocales    StringList      `gorm:"type:json;not null" json:"supported_locales"`
	DefaultCurrency     string          `gorm:"size:3;not null" json:"default_currency"`
	DisplayCurrencies   StringList      `gorm:"type:json;not null" json:"display_currencies"`
	PaymentMethodPolicy string          `gorm:"size:80;not null;default:''" json:"payment_method_policy"`
	LogisticsPolicy     string          `gorm:"size:80;not null;default:''" json:"logistics_policy"`
	TaxPolicy           string          `gorm:"size:80;not null;default:''" json:"tax_policy"`
	Enabled             bool            `gorm:"not null;default:true" json:"enabled"`
	Priority            int             `gorm:"not null;default:100" json:"priority"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	DeletedAt           gorm.DeletedAt  `gorm:"index" json:"-"`
	Countries           []MarketCountry `gorm:"foreignKey:MarketID;constraint:OnDelete:CASCADE" json:"countries"`
}

func (StorefrontMarket) TableName() string {
	return "storefront_markets"
}

type MarketCountry struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	MarketID  uint           `gorm:"not null;index" json:"market_id"`
	Code      string         `gorm:"size:2;not null;uniqueIndex" json:"code"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MarketCountry) TableName() string {
	return "storefront_market_countries"
}
