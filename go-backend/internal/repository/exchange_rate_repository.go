package repository

import (
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExchangeRateRepository struct {
	db *gorm.DB
}

func NewExchangeRateRepository(db *gorm.DB) *ExchangeRateRepository {
	return &ExchangeRateRepository{db: db}
}

func (r *ExchangeRateRepository) WithTx(tx *gorm.DB) *ExchangeRateRepository {
	return &ExchangeRateRepository{db: tx}
}

func (r *ExchangeRateRepository) Find(baseCurrency, quoteCurrency string) (*currency.ExchangeRate, error) {
	var rate currency.ExchangeRate
	if err := r.db.Where(
		"base_currency = ? AND quote_currency = ?",
		strings.ToUpper(strings.TrimSpace(baseCurrency)),
		strings.ToUpper(strings.TrimSpace(quoteCurrency)),
	).First(&rate).Error; err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *ExchangeRateRepository) FindFresh(baseCurrency, quoteCurrency string, now time.Time) (*currency.ExchangeRate, error) {
	var rate currency.ExchangeRate
	now = now.UTC()
	if err := r.db.Where(
		"base_currency = ? AND quote_currency = ? AND (expires_at IS NULL OR expires_at > ?)",
		strings.ToUpper(strings.TrimSpace(baseCurrency)),
		strings.ToUpper(strings.TrimSpace(quoteCurrency)),
		now,
	).First(&rate).Error; err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *ExchangeRateRepository) List(baseCurrency string) ([]currency.ExchangeRate, error) {
	var rates []currency.ExchangeRate
	query := r.db.Order("quote_currency ASC")
	baseCurrency = strings.ToUpper(strings.TrimSpace(baseCurrency))
	if baseCurrency != "" {
		query = query.Where("base_currency = ?", baseCurrency)
	}
	if err := query.Find(&rates).Error; err != nil {
		return nil, err
	}
	return rates, nil
}

func (r *ExchangeRateRepository) UpsertRates(rates []currency.ExchangeRate) error {
	if len(rates) == 0 {
		return nil
	}

	now := time.Now().UTC()
	for i := range rates {
		rates[i].BaseCurrency = strings.ToUpper(strings.TrimSpace(rates[i].BaseCurrency))
		rates[i].QuoteCurrency = strings.ToUpper(strings.TrimSpace(rates[i].QuoteCurrency))
		rates[i].UpdatedAt = now
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "base_currency"},
			{Name: "quote_currency"},
		},
		TargetWhere: clause.Where{
			Exprs: []clause.Expression{
				clause.Eq{
					Column: clause.Column{Name: "deleted_at"},
					Value:  nil,
				},
			},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"rate",
			"source",
			"fetched_at",
			"expires_at",
			"updated_at",
		}),
	}).Create(&rates).Error
}

func (r *ExchangeRateRepository) TryAcquireSyncLease(
	leaseKey string,
	ownerID string,
	now time.Time,
	leaseTTL time.Duration,
) (bool, error) {
	leaseKey = strings.TrimSpace(leaseKey)
	ownerID = strings.TrimSpace(ownerID)
	if leaseKey == "" || ownerID == "" || leaseTTL <= 0 {
		return false, nil
	}

	now = now.UTC()
	expiresAt := now.Add(leaseTTL)
	result := r.db.Exec(`
		INSERT INTO currency_exchange_sync_leases
			(lease_key, owner_id, lease_expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (lease_key) DO UPDATE SET
			owner_id = excluded.owner_id,
			lease_expires_at = excluded.lease_expires_at,
			updated_at = excluded.updated_at
		WHERE currency_exchange_sync_leases.lease_expires_at <= ?
		   OR currency_exchange_sync_leases.owner_id = excluded.owner_id
	`, leaseKey, ownerID, expiresAt, now, now, now)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *ExchangeRateRepository) ReleaseSyncLease(leaseKey, ownerID string, now time.Time) error {
	leaseKey = strings.TrimSpace(leaseKey)
	ownerID = strings.TrimSpace(ownerID)
	if leaseKey == "" || ownerID == "" {
		return nil
	}

	now = now.UTC()
	return r.db.Model(&currency.ExchangeRateSyncLease{}).
		Where("lease_key = ? AND owner_id = ?", leaseKey, ownerID).
		Updates(map[string]interface{}{
			"lease_expires_at": now,
			"updated_at":       now,
		}).Error
}
