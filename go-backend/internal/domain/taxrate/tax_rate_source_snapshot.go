package taxrate

import "time"

const (
	DefaultTaxRateSourceProviderCode         = "vatcomply"
	DefaultTaxRateSourceRefreshIntervalHours = 720
)

// TaxRateSourceConfig controls collection of a reference tax-rate dataset.
// This configuration is deliberately separate from TaxRate, which checkout
// continues to read from the existing tax_rates table.
type TaxRateSourceConfig struct {
	ID                   uint       `gorm:"primaryKey;autoIncrement:false" json:"id"`
	ProviderCode         string     `gorm:"size:40;not null" json:"provider_code"`
	Enabled              bool       `gorm:"not null;default:false" json:"enabled"`
	RefreshIntervalHours int        `gorm:"not null;default:720" json:"refresh_interval_hours"`
	LastCheckedAt        *time.Time `json:"last_checked_at,omitempty"`
	LastSuccessfulSyncAt *time.Time `json:"last_successful_sync_at,omitempty"`
	LastSyncError        string     `gorm:"type:text;not null;default:''" json:"last_sync_error,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (TaxRateSourceConfig) TableName() string {
	return "tax_rate_source_configs"
}

// TaxRateSourceSnapshot is an immutable, versioned copy of one provider
// response after normalization. It is an audit/reference record and is not a
// checkout tax-rate source.
type TaxRateSourceSnapshot struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ProviderCode   string    `gorm:"size:40;not null;index:idx_tax_rate_source_snapshots_provider_created,priority:1" json:"provider_code"`
	SourceEndpoint string    `gorm:"size:255;not null" json:"source_endpoint"`
	Version        string    `gorm:"size:100;not null;uniqueIndex" json:"version"`
	ContentSHA256  string    `gorm:"size:64;not null" json:"content_sha256"`
	CapturedAt     time.Time `gorm:"not null;index:idx_tax_rate_source_snapshots_provider_created,priority:2,sort:desc" json:"captured_at"`
	CountryCount   int       `gorm:"not null" json:"country_count"`
	RateCount      int       `gorm:"not null" json:"rate_count"`
	CreatedAt      time.Time `json:"created_at"`
}

func (TaxRateSourceSnapshot) TableName() string {
	return "tax_rate_source_snapshots"
}

// TaxRateSourceSnapshotEntry stores one rate or product-category rate from a
// snapshot. CountryCode uses the storefront's ISO code; SourceCountryCode
// preserves the provider's original code (for example EL for Greece).
type TaxRateSourceSnapshotEntry struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	SnapshotID        uint      `gorm:"not null;index:idx_tax_rate_source_snapshot_entries_snapshot_country,priority:1" json:"snapshot_id"`
	CountryCode       string    `gorm:"size:2;not null;index:idx_tax_rate_source_snapshot_entries_snapshot_country,priority:2" json:"country_code"`
	SourceCountryCode string    `gorm:"size:2;not null" json:"source_country_code"`
	CountryName       string    `gorm:"size:120;not null" json:"country_name"`
	Currency          string    `gorm:"size:3;not null;default:''" json:"currency"`
	RateType          string    `gorm:"size:32;not null" json:"rate_type"`
	RateCategory      string    `gorm:"size:96;not null;default:''" json:"rate_category"`
	RateDecimal       string    `gorm:"type:numeric(12,8);not null" json:"rate_decimal"`
	CreatedAt         time.Time `json:"created_at"`
}

func (TaxRateSourceSnapshotEntry) TableName() string {
	return "tax_rate_source_snapshot_entries"
}
