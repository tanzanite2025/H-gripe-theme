package merchant

import "time"

const GoogleMerchantProvider = "google_merchant"

type GoogleMerchantConnection struct {
	ID                     uint       `gorm:"primarykey" json:"id"`
	Provider               string     `gorm:"type:varchar(40);uniqueIndex;not null" json:"provider"`
	GoogleSubject          string     `gorm:"type:varchar(160)" json:"-"`
	GoogleAccountEmail     string     `gorm:"type:varchar(320)" json:"google_account_email"`
	MerchantAccountID      string     `gorm:"type:varchar(40)" json:"merchant_account_id"`
	DataSourceID           string     `gorm:"type:varchar(40)" json:"data_source_id"`
	StorefrontBaseURL      string     `gorm:"type:varchar(255)" json:"storefront_base_url"`
	RefreshTokenEncrypted  string     `gorm:"type:text" json:"-"`
	GrantedScopes          string     `gorm:"type:text" json:"-"`
	TokenExpiresAt         *time.Time `json:"-"`
	Status                 string     `gorm:"type:varchar(24);not null;default:'disconnected';index" json:"status"`
	OAuthStateHash         string     `gorm:"type:varchar(128)" json:"-"`
	OAuthStateExpiresAt    *time.Time `json:"-"`
	OAuthInitiatedByUserID *uint      `json:"-"`
	LastConnectedAt        *time.Time `json:"last_connected_at"`
	LastError              string     `gorm:"type:text" json:"last_error"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

func (GoogleMerchantConnection) TableName() string {
	return "google_merchant_connections"
}
