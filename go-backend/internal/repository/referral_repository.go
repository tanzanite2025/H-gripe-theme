package repository

import (
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/loyalty"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrReferralRecordVersionConflict  = errors.New("referral record version conflict")
	ErrReferralProgramVersionConflict = errors.New("referral program version conflict")
)

type ReferralRepository struct {
	db *gorm.DB
}

func NewReferralRepository(db *gorm.DB) *ReferralRepository {
	return &ReferralRepository{db: db}
}

func (r *ReferralRepository) WithTx(tx *gorm.DB) *ReferralRepository {
	return &ReferralRepository{db: tx}
}

func (r *ReferralRepository) CreateIdentity(identity *loyalty.ReferralIdentity) error {
	if r == nil || r.db == nil || identity == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.Create(identity).Error
}

func (r *ReferralRepository) FindIdentityByUserID(userID uint) (*loyalty.ReferralIdentity, error) {
	if r == nil || r.db == nil || userID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var identity loyalty.ReferralIdentity
	if err := r.db.Where("user_id = ?", userID).First(&identity).Error; err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *ReferralRepository) FindActiveIdentityByCode(code string) (*loyalty.ReferralIdentity, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrRecordNotFound
	}
	code = loyalty.NormalizeReferralCode(code)
	if err := loyalty.ValidateReferralCode(code); err != nil {
		return nil, err
	}
	var identity loyalty.ReferralIdentity
	if err := r.db.Where("UPPER(referral_code) = ? AND is_active = ?", code, true).First(&identity).Error; err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *ReferralRepository) CreateRecord(record *loyalty.ReferralRecord) error {
	if r == nil || r.db == nil || record == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.Create(record).Error
}

func (r *ReferralRepository) FindRecordByIDForUpdate(id uint) (*loyalty.ReferralRecord, error) {
	if r == nil || r.db == nil || id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var record loyalty.ReferralRecord
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// FindRecordByID reads one referral record without taking a write lock. Admin
// detail views use this path so inspecting a record never competes with the
// lifecycle worker.
func (r *ReferralRepository) FindRecordByID(id uint) (*loyalty.ReferralRecord, error) {
	if r == nil || r.db == nil || id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var record loyalty.ReferralRecord
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ReferralRepository) FindRecordByRefereeID(refereeID uint) (*loyalty.ReferralRecord, error) {
	if r == nil || r.db == nil || refereeID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var record loyalty.ReferralRecord
	if err := r.db.Where("referee_id = ?", refereeID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ReferralRepository) FindRecordByRefereeIDForUpdate(refereeID uint) (*loyalty.ReferralRecord, error) {
	if r == nil || r.db == nil || refereeID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var record loyalty.ReferralRecord
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("referee_id = ?", refereeID).
		First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ReferralRepository) FindRecordByOrderIDForUpdate(orderID uint) (*loyalty.ReferralRecord, error) {
	if r == nil || r.db == nil || orderID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var record loyalty.ReferralRecord
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("order_id = ?", orderID).
		First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// ListPendingExpired returns a bounded snapshot of pending records that have
// passed their attribution window. Callers must lock each record again before
// applying a transition because the snapshot may race with a bind or payment.
func (r *ReferralRepository) ListPendingExpired(now time.Time, limit int) ([]loyalty.ReferralRecord, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var records []loyalty.ReferralRecord
	err := r.db.Where("status = ? AND expires_at <= ?", loyalty.ReferralStatusPending, now).
		Order("expires_at ASC, id ASC").Limit(limit).Find(&records).Error
	return records, err
}

// ListOrderedCandidates returns ordered records whose order either has an
// authoritative delivered_at or has reached the configured undelivered
// fallback deadline. The service rechecks the order and record under lock.
func (r *ReferralRepository) ListOrderedCandidates(now time.Time, limit int) ([]loyalty.ReferralRecord, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	// Keep the deadline predicate in SQL so a large backlog cannot hide ready
	// candidates behind newer records. Each supported dialect has an explicit
	// date expression; the service still rechecks the row under lock.
	fallbackPredicate := "orders.shipped_at IS NOT NULL AND orders.shipped_at <= ?"
	switch r.db.Dialector.Name() {
	case "postgres":
		fallbackPredicate = "orders.shipped_at IS NOT NULL AND orders.shipped_at + (referral_program_configs.undelivered_fallback_days * INTERVAL '1 day') <= ?"
	case "mysql":
		fallbackPredicate = "orders.shipped_at IS NOT NULL AND DATE_ADD(orders.shipped_at, INTERVAL referral_program_configs.undelivered_fallback_days DAY) <= ?"
	case "sqlite":
		fallbackPredicate = "orders.shipped_at IS NOT NULL AND datetime(orders.shipped_at, '+' || referral_program_configs.undelivered_fallback_days || ' days') <= ?"
	}
	var records []loyalty.ReferralRecord
	err := r.db.Model(&loyalty.ReferralRecord{}).
		Joins("JOIN orders ON orders.id = referral_records.order_id").
		Joins("JOIN referral_program_configs ON referral_program_configs.id = referral_records.program_config_id").
		Where("referral_records.status = ?", loyalty.ReferralStatusOrdered).
		Where("orders.delivered_at IS NOT NULL AND orders.delivered_at <= ? OR "+fallbackPredicate, now, now).
		Order("referral_records.ordered_at ASC, referral_records.id ASC").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (r *ReferralRepository) ListMaturedVesting(now time.Time, limit int) ([]loyalty.ReferralRecord, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var records []loyalty.ReferralRecord
	err := r.db.Where("status = ? AND vesting_until IS NOT NULL AND vesting_until <= ?", loyalty.ReferralStatusVesting, now).
		Order("vesting_until ASC, id ASC").Limit(limit).Find(&records).Error
	return records, err
}

// CountMonthlyConvertedByReferrer counts referral records that reached a
// qualifying paid-order state during one calendar month. Revoked and expired
// attributions do not consume the cap; reversed orders remain counted so a
// referrer cannot cycle the same month through refunds to bypass the limit.
func (r *ReferralRepository) CountMonthlyConvertedByReferrer(referrerID uint, start, end time.Time) (int64, error) {
	if r == nil || r.db == nil || referrerID == 0 {
		return 0, nil
	}
	if start.IsZero() || end.IsZero() || !end.After(start) {
		return 0, errors.New("invalid referral monthly cap window")
	}
	var count int64
	err := r.db.Model(&loyalty.ReferralRecord{}).
		Where("referrer_id = ? AND ordered_at >= ? AND ordered_at < ?", referrerID, start.UTC(), end.UTC()).
		Where("status IN ?", []string{loyalty.ReferralStatusOrdered, loyalty.ReferralStatusVesting, loyalty.ReferralStatusSettled, loyalty.ReferralStatusReversed}).
		Count(&count).Error
	return count, err
}

// CountRecentBindingsByIPSubnetHash counts still-effective referral bindings
// created in a rolling window. Expired, revoked, and reversed records do not
// consume the anti-fraud allowance because they are no longer valid
// attributions.
func (r *ReferralRepository) CountRecentBindingsByIPSubnetHash(subnetHash string, since time.Time) (int64, error) {
	if r == nil || r.db == nil || strings.TrimSpace(subnetHash) == "" {
		return 0, nil
	}
	if since.IsZero() {
		return 0, errors.New("invalid referral binding window")
	}
	var count int64
	err := r.db.Model(&loyalty.ReferralRecord{}).
		Where("client_ip_subnet_hash = ? AND created_at >= ?", strings.TrimSpace(subnetHash), since.UTC()).
		Where("status IN ?", []string{
			loyalty.ReferralStatusPending,
			loyalty.ReferralStatusOrdered,
			loyalty.ReferralStatusVesting,
			loyalty.ReferralStatusSettled,
		}).
		Count(&count).Error
	return count, err
}

func (r *ReferralRepository) ListRecordsByReferrerID(referrerID uint, page, pageSize int) ([]loyalty.ReferralRecord, int64, error) {
	if r == nil || r.db == nil || referrerID == 0 {
		return nil, 0, gorm.ErrRecordNotFound
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	query := r.db.Model(&loyalty.ReferralRecord{}).Where("referrer_id = ?", referrerID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []loyalty.ReferralRecord
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

type ReferralStats struct {
	TotalInvited        int64 `json:"total_invited_count"`
	SuccessfulOrders    int64 `json:"successful_orders_count"`
	PendingRewardPoints int64 `json:"pending_reward_points"`
	SettledRewardPoints int64 `json:"settled_reward_points"`
}

type ReferralAdminFilters struct {
	Status  string
	Keyword string
	From    *time.Time
	To      *time.Time
}

type ReferralAdminStats struct {
	TotalReferrals       int64 `json:"total_referrals"`
	ConvertedOrders      int64 `json:"converted_orders"`
	AttributedGMVMinor   int64 `json:"attributed_gmv_minor"`
	PendingVestingPoints int64 `json:"pending_vesting_points"`
	SettledPoints        int64 `json:"settled_points"`
	FraudBlockedCount    int64 `json:"fraud_blocked_count"`
}

func (r *ReferralRepository) ListAdminRecords(filters ReferralAdminFilters, page, pageSize int) ([]loyalty.ReferralRecord, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, gorm.ErrInvalidDB
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	query := r.db.Model(&loyalty.ReferralRecord{}).
		Joins("LEFT JOIN users AS referrer_user ON referrer_user.id = referral_records.referrer_id").
		Joins("LEFT JOIN users AS referee_user ON referee_user.id = referral_records.referee_id").
		Joins("LEFT JOIN orders AS referral_order ON referral_order.id = referral_records.order_id")
	query = applyReferralAdminFilters(query, filters, "referral_records")
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []loyalty.ReferralRecord
	if err := query.Order("referral_records.created_at DESC, referral_records.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func applyReferralAdminFilters(query *gorm.DB, filters ReferralAdminFilters, recordAlias string) *gorm.DB {
	if status := strings.TrimSpace(strings.ToLower(filters.Status)); status != "" {
		query = query.Where(recordAlias+".status = ?", status)
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("LOWER("+recordAlias+".referral_code_snapshot) LIKE LOWER(?) OR LOWER(referrer_user.email) LIKE LOWER(?) OR LOWER(referee_user.email) LIKE LOWER(?) OR LOWER(referral_order.order_number) LIKE LOWER(?)", like, like, like, like)
	}
	if filters.From != nil {
		query = query.Where(recordAlias+".created_at >= ?", filters.From.UTC())
	}
	if filters.To != nil {
		query = query.Where(recordAlias+".created_at < ?", filters.To.UTC())
	}
	return query
}

// AdminStats accepts an optional filter for backwards compatibility with
// callers that previously requested the unfiltered overview.
func (r *ReferralRepository) AdminStats(filterArgs ...ReferralAdminFilters) (ReferralAdminStats, error) {
	stats := ReferralAdminStats{}
	if r == nil || r.db == nil {
		return stats, gorm.ErrInvalidDB
	}
	var filters ReferralAdminFilters
	if len(filterArgs) > 0 {
		filters = filterArgs[0]
	}
	query := r.db.Table("referral_records AS rr").
		Select(`COUNT(*) AS total_referrals,
			COALESCE(SUM(CASE WHEN rr.status IN ('ordered','vesting','settled','reversed') THEN 1 ELSE 0 END), 0) AS converted_orders,
			COALESCE(SUM(CASE WHEN rr.status IN ('ordered','vesting','settled','reversed') THEN rr.order_amount_minor ELSE 0 END), 0) AS attributed_gmv_minor,
			COALESCE(SUM(CASE WHEN rr.status IN ('ordered','vesting') THEN rpc.referrer_reward_points ELSE 0 END), 0) AS pending_vesting_points,
			COALESCE(SUM(CASE WHEN rr.status = 'settled' THEN rpc.referrer_reward_points ELSE 0 END), 0) AS settled_points,
			COALESCE(SUM(CASE WHEN CAST(rr.risk_flags AS TEXT) <> '[]' THEN 1 ELSE 0 END), 0) AS fraud_blocked_count`).
		Joins("JOIN referral_program_configs AS rpc ON rpc.id = rr.program_config_id").
		Joins("LEFT JOIN users AS referrer_user ON referrer_user.id = rr.referrer_id").
		Joins("LEFT JOIN users AS referee_user ON referee_user.id = rr.referee_id").
		Joins("LEFT JOIN orders AS referral_order ON referral_order.id = rr.order_id")
	row := applyReferralAdminFilters(query, filters, "rr").Row()
	if err := row.Scan(&stats.TotalReferrals, &stats.ConvertedOrders, &stats.AttributedGMVMinor, &stats.PendingVestingPoints, &stats.SettledPoints, &stats.FraudBlockedCount); err != nil {
		return ReferralAdminStats{}, err
	}
	return stats, nil
}

func (r *ReferralRepository) StatsByReferrerID(referrerID uint) (ReferralStats, error) {
	stats := ReferralStats{}
	if r == nil || r.db == nil || referrerID == 0 {
		return stats, nil
	}
	row := r.db.Table("referral_records AS rr").
		Select(`
			COUNT(*) AS total_invited,
			COALESCE(SUM(CASE WHEN rr.status IN ('ordered', 'vesting', 'settled', 'reversed') THEN 1 ELSE 0 END), 0) AS successful_orders,
			COALESCE(SUM(CASE WHEN rr.status IN ('ordered', 'vesting') THEN rpc.referrer_reward_points ELSE 0 END), 0) AS pending_reward_points,
			COALESCE(SUM(CASE WHEN rr.status = 'settled' THEN rpc.referrer_reward_points ELSE 0 END), 0) AS settled_reward_points
		`).
		Joins("JOIN referral_program_configs AS rpc ON rpc.id = rr.program_config_id").
		Where("rr.referrer_id = ?", referrerID).
		Row()
	if err := row.Scan(
		&stats.TotalInvited,
		&stats.SuccessfulOrders,
		&stats.PendingRewardPoints,
		&stats.SettledRewardPoints,
	); err != nil {
		return ReferralStats{}, err
	}
	return stats, nil
}

// UpdateRecordState uses both the current state and record version as the
// compare-and-swap boundary. Callers must append a ReferralTransition in the
// same transaction after this succeeds.
func (r *ReferralRepository) UpdateRecordState(
	id uint,
	currentStatus string,
	currentVersion int,
	updates map[string]any,
) error {
	if r == nil || r.db == nil || id == 0 || currentVersion <= 0 || updates == nil {
		return gorm.ErrInvalidData
	}
	updates["record_version"] = currentVersion + 1
	result := r.db.Model(&loyalty.ReferralRecord{}).
		Where("id = ? AND status = ? AND record_version = ?", id, currentStatus, currentVersion).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrReferralRecordVersionConflict
	}
	return nil
}

func (r *ReferralRepository) CreateReward(reward *loyalty.ReferralReward) error {
	if r == nil || r.db == nil || reward == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.Create(reward).Error
}

func (r *ReferralRepository) FindRewardByIdempotencyKey(key string) (*loyalty.ReferralReward, error) {
	if r == nil || r.db == nil || strings.TrimSpace(key) == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var reward loyalty.ReferralReward
	if err := r.db.Where("idempotency_key = ?", strings.TrimSpace(key)).First(&reward).Error; err != nil {
		return nil, err
	}
	return &reward, nil
}

func (r *ReferralRepository) UpdateReward(id uint, updates map[string]any) error {
	if r == nil || r.db == nil || id == 0 || updates == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Model(&loyalty.ReferralReward{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ReferralRepository) AppendTransition(transition *loyalty.ReferralTransition) error {
	if r == nil || r.db == nil || transition == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.Create(transition).Error
}

func (r *ReferralRepository) ListTransitionsByRecordID(recordID uint) ([]loyalty.ReferralTransition, error) {
	if r == nil || r.db == nil || recordID == 0 {
		return nil, gorm.ErrInvalidData
	}
	var transitions []loyalty.ReferralTransition
	err := r.db.Where("referral_record_id = ?", recordID).
		Order("created_at ASC, id ASC").Find(&transitions).Error
	return transitions, err
}

func (r *ReferralRepository) ListRewardsByRecordID(recordID uint) ([]loyalty.ReferralReward, error) {
	if r == nil || r.db == nil || recordID == 0 {
		return nil, gorm.ErrInvalidData
	}
	var rewards []loyalty.ReferralReward
	err := r.db.Where("referral_record_id = ?", recordID).
		Order("created_at ASC, id ASC").Find(&rewards).Error
	return rewards, err
}

type ReferralProgramRepository struct {
	db *gorm.DB
}

func NewReferralProgramRepository(db *gorm.DB) *ReferralProgramRepository {
	return &ReferralProgramRepository{db: db}
}

func (r *ReferralProgramRepository) WithTx(tx *gorm.DB) *ReferralProgramRepository {
	return &ReferralProgramRepository{db: tx}
}

func (r *ReferralProgramRepository) FindActive() (*loyalty.ReferralProgramConfig, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var config loyalty.ReferralProgramConfig
	if err := r.db.Where("status = ?", "active").Order("version DESC").First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *ReferralProgramRepository) FindByID(id uint) (*loyalty.ReferralProgramConfig, error) {
	if r == nil || r.db == nil || id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var config loyalty.ReferralProgramConfig
	if err := r.db.First(&config, id).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *ReferralProgramRepository) FindByIDs(ids []uint) ([]loyalty.ReferralProgramConfig, error) {
	ids = uniqueUintValues(ids)
	if len(ids) == 0 {
		return []loyalty.ReferralProgramConfig{}, nil
	}
	var configs []loyalty.ReferralProgramConfig
	err := r.db.Where("id IN ?", ids).Find(&configs).Error
	return configs, err
}

// CreateVersion requires the version read by the admin. This prevents two
// operators from silently overwriting each other's referral policy changes.
func (r *ReferralProgramRepository) CreateVersion(config *loyalty.ReferralProgramConfig, expectedVersion int) error {
	if r == nil || r.db == nil || config == nil || expectedVersion <= 0 {
		return gorm.ErrInvalidData
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if tx.Name() == "postgres" {
			if err := tx.Exec("LOCK TABLE referral_program_configs IN EXCLUSIVE MODE").Error; err != nil {
				return err
			}
		}

		var active loyalty.ReferralProgramConfig
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("status = ?", "active").
			Order("version DESC").
			First(&active).Error; err != nil {
			return err
		}
		if active.Version != expectedVersion {
			return ErrReferralProgramVersionConflict
		}

		if err := tx.Model(&loyalty.ReferralProgramConfig{}).
			Where("id = ? AND status = ?", active.ID, "active").
			Update("status", "archived").Error; err != nil {
			return err
		}

		config.ID = 0
		config.Version = active.Version + 1
		config.Status = "active"
		config.CreatedAt = time.Time{}
		return tx.Create(config).Error
	})
}
