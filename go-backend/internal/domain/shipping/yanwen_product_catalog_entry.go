package shipping

import (
	"errors"
	"strings"
	"time"
)

// YanwenProductCatalogEntry caches a product returned by Yanwen's official
// express.channel.getlist endpoint for one gateway environment.
type YanwenProductCatalogEntry struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Environment  string    `gorm:"size:16;not null;uniqueIndex:idx_yanwen_product_environment" json:"environment"`
	ProductID    string    `gorm:"size:80;not null;uniqueIndex:idx_yanwen_product_environment" json:"product_id"`
	NameChinese  string    `gorm:"column:name_ch;size:200;not null;default:''" json:"name_ch"`
	NameEnglish  string    `gorm:"column:name_en;size:200;not null;default:''" json:"name_en"`
	LastSyncedAt time.Time `gorm:"column:last_synced_at;not null" json:"last_synced_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (YanwenProductCatalogEntry) TableName() string {
	return "shipping_yanwen_product_catalog_entries"
}

func (entry *YanwenProductCatalogEntry) Validate() error {
	if entry == nil {
		return errors.New("Yanwen product catalog entry is required")
	}
	entry.Environment = strings.ToLower(strings.TrimSpace(entry.Environment))
	entry.ProductID = strings.TrimSpace(entry.ProductID)
	entry.NameChinese = strings.TrimSpace(entry.NameChinese)
	entry.NameEnglish = strings.TrimSpace(entry.NameEnglish)
	if entry.Environment != "fat" && entry.Environment != "production" {
		return errors.New("Yanwen product environment must be fat or production")
	}
	if entry.ProductID == "" {
		return errors.New("Yanwen product id is required")
	}
	if entry.NameChinese == "" && entry.NameEnglish == "" {
		return errors.New("Yanwen product name is required")
	}
	if entry.LastSyncedAt.IsZero() {
		return errors.New("Yanwen product last synced time is required")
	}
	return nil
}
