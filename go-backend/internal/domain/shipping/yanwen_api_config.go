package shipping

import "time"

// YanwenAPIConfig stores one encrypted credential set for a Yanwen gateway environment.
type YanwenAPIConfig struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	Environment       string    `gorm:"size:16;not null;default:'fat';uniqueIndex" json:"environment"`
	Endpoint          string    `gorm:"size:500;not null" json:"endpoint"`
	UserIDEncrypted   string    `gorm:"type:text;not null;default:''" json:"-"`
	APITokenEncrypted string    `gorm:"type:text;not null;default:''" json:"-"`
	Enabled           bool      `gorm:"not null;default:false" json:"enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (YanwenAPIConfig) TableName() string { return "shipping_yanwen_api_configs" }

type YanwenAPIConfigView struct {
	Environment        string `json:"environment"`
	Endpoint           string `json:"endpoint"`
	UserIDConfigured   bool   `json:"user_id_configured"`
	APITokenConfigured bool   `json:"api_token_configured"`
	Enabled            bool   `json:"enabled"`
}
