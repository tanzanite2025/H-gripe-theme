package shipping

import (
	"errors"
	"strings"
	"time"
)

// YanwenWarehouseCatalogEntry caches a warehouse returned by Yanwen's
// official common.warehouse.getlist endpoint for one gateway environment.
type YanwenWarehouseCatalogEntry struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	Environment   string    `gorm:"size:16;not null;uniqueIndex:idx_yanwen_warehouse_environment" json:"environment"`
	WarehouseCode string    `gorm:"column:warehouse_code;size:80;not null;uniqueIndex:idx_yanwen_warehouse_environment" json:"warehouse_code"`
	Name          string    `gorm:"size:200;not null" json:"name"`
	Area          string    `gorm:"size:200;not null;default:''" json:"area"`
	LastSyncedAt  time.Time `gorm:"column:last_synced_at;not null" json:"last_synced_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (YanwenWarehouseCatalogEntry) TableName() string {
	return "shipping_yanwen_warehouse_catalog_entries"
}

func (entry *YanwenWarehouseCatalogEntry) Validate() error {
	if entry == nil {
		return errors.New("Yanwen warehouse catalog entry is required")
	}
	entry.Environment = strings.ToLower(strings.TrimSpace(entry.Environment))
	entry.WarehouseCode = strings.TrimSpace(entry.WarehouseCode)
	entry.Name = strings.TrimSpace(entry.Name)
	entry.Area = strings.TrimSpace(entry.Area)
	if entry.Environment != "fat" && entry.Environment != "production" {
		return errors.New("Yanwen warehouse environment must be fat or production")
	}
	if entry.WarehouseCode == "" {
		return errors.New("Yanwen warehouse code is required")
	}
	if entry.Name == "" {
		return errors.New("Yanwen warehouse name is required")
	}
	if entry.LastSyncedAt.IsZero() {
		return errors.New("Yanwen warehouse last synced time is required")
	}
	return nil
}
