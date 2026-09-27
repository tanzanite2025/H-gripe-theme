package shipping

import "time"

type FpxAPIConfig struct {
	ID                       uint       `gorm:"primarykey" json:"id"`
	Environment              string     `gorm:"size:16;not null;default:'production';uniqueIndex" json:"environment"`
	Endpoint                 string     `gorm:"size:500;not null" json:"endpoint"`
	AppKeyEncrypted          string     `gorm:"type:text;not null;default:''" json:"-"`
	AppSecretEncrypted       string     `gorm:"type:text;not null;default:''" json:"-"`
	AccessTokenEncrypted     string     `gorm:"type:text;not null;default:''" json:"-"`
	Enabled                  bool       `gorm:"not null;default:false" json:"enabled"`
	LastSyncStatus           string     `gorm:"size:32;not null;default:''" json:"last_sync_status"`
	LastSyncedAt             *time.Time `json:"last_synced_at,omitempty"`
	LastError                string     `gorm:"size:500;not null;default:''" json:"last_error,omitempty"`
	LastSyncScanned          int        `gorm:"not null;default:0" json:"last_sync_scanned"`
	LastSyncAdded            int        `gorm:"not null;default:0" json:"last_sync_added"`
	LastSyncUpdated          int        `gorm:"not null;default:0" json:"last_sync_updated"`
	LastSyncPreservedEnabled int        `gorm:"not null;default:0" json:"last_sync_preserved_enabled"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

func (FpxAPIConfig) TableName() string { return "shipping_fpx_api_configs" }

type FpxAPIConfigView struct {
	Environment              string     `json:"environment"`
	Endpoint                 string     `json:"endpoint"`
	AppKeyConfigured         bool       `json:"app_key_configured"`
	AppSecretConfigured      bool       `json:"app_secret_configured"`
	AccessTokenConfigured    bool       `json:"access_token_configured"`
	Enabled                  bool       `json:"enabled"`
	LastSyncStatus           string     `json:"last_sync_status"`
	LastSyncedAt             *time.Time `json:"last_synced_at,omitempty"`
	LastError                string     `json:"last_error,omitempty"`
	LastSyncScanned          int        `json:"last_sync_scanned"`
	LastSyncAdded            int        `json:"last_sync_added"`
	LastSyncUpdated          int        `json:"last_sync_updated"`
	LastSyncPreservedEnabled int        `json:"last_sync_preserved_enabled"`
}
