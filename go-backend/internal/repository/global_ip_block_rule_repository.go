package repository

import (
	"context"
	"hash/fnv"
	"strings"
	"time"

	"commerce-platform/internal/domain/security"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GlobalIPBlockRuleRepository struct {
	db *gorm.DB
}

func NewGlobalIPBlockRuleRepository(db *gorm.DB) *GlobalIPBlockRuleRepository {
	return &GlobalIPBlockRuleRepository{db: db}
}

func (r *GlobalIPBlockRuleRepository) WithTx(tx *gorm.DB) *GlobalIPBlockRuleRepository {
	if tx == nil {
		return r
	}
	return &GlobalIPBlockRuleRepository{db: tx}
}

func (r *GlobalIPBlockRuleRepository) Transaction(
	ctx context.Context,
	fn func(tx *gorm.DB, repo *GlobalIPBlockRuleRepository) error,
) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if fn == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx, r.WithTx(tx))
	})
}

type IPBlockRuleListFilters struct {
	Search string
	Source string
	Status string
}

// UpsertEnabled serializes mutations for one source reference and updates an
// existing enabled row, including an expired one, instead of creating a
// second active identity.
func (r *GlobalIPBlockRuleRepository) UpsertEnabled(
	ctx context.Context,
	desired security.IPBlockRule,
) (*security.IPBlockRule, error) {
	_, saved, err := r.UpsertEnabledWithPrevious(ctx, desired)
	return saved, err
}

func (r *GlobalIPBlockRuleRepository) UpsertEnabledWithPrevious(
	ctx context.Context,
	desired security.IPBlockRule,
) (*security.IPBlockRule, *security.IPBlockRule, error) {
	if r == nil || r.db == nil {
		return nil, nil, gorm.ErrInvalidDB
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var before *security.IPBlockRule
	var saved security.IPBlockRule
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockGlobalIPBlockSourceReference(tx, desired.Source, desired.SourceReference); err != nil {
			return err
		}

		var existing security.IPBlockRule
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("cidr = ?", strings.TrimSpace(desired.CIDR)).
			Where("source = ?", strings.TrimSpace(desired.Source)).
			Where("source_reference = ?", strings.TrimSpace(desired.SourceReference)).
			Where("enabled = ?", true).
			Order("updated_at DESC").
			Order("created_at DESC").
			Order("id DESC").
			First(&existing).Error
		switch {
		case err == nil:
			beforeCopy := existing
			before = &beforeCopy
			updates := map[string]interface{}{
				"reason":      desired.Reason,
				"expires_at":  desired.ExpiresAt,
				"enabled":     true,
				"disabled_by": nil,
				"disabled_at": nil,
			}
			if desired.CreatedBy != nil && existing.CreatedBy == nil {
				updates["created_by"] = desired.CreatedBy
			}
			if err := tx.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
			if err := tx.First(&existing, existing.ID).Error; err != nil {
				return err
			}
			saved = existing
			return nil
		case !IsRecordNotFound(err):
			return err
		}

		if err := tx.Create(&desired).Error; err != nil {
			return err
		}
		saved = desired
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return before, &saved, nil
}

func (r *GlobalIPBlockRuleRepository) FindByID(id uint) (*security.IPBlockRule, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}

	var rule security.IPBlockRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *GlobalIPBlockRuleRepository) FindActiveByIdentity(cidr, source, sourceReference string, now time.Time) (*security.IPBlockRule, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	var rule security.IPBlockRule
	err := r.db.
		Where("cidr = ?", strings.TrimSpace(cidr)).
		Where("source = ?", strings.TrimSpace(source)).
		Where("source_reference = ?", strings.TrimSpace(sourceReference)).
		Where("enabled = ?", true).
		Where("(expires_at IS NULL OR expires_at > ?)", now).
		Order("updated_at DESC").
		Order("created_at DESC").
		Order("id DESC").
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *GlobalIPBlockRuleRepository) FindEnabledByIdentity(cidr, source, sourceReference string) (*security.IPBlockRule, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}

	var rule security.IPBlockRule
	err := r.db.
		Where("cidr = ?", strings.TrimSpace(cidr)).
		Where("source = ?", strings.TrimSpace(source)).
		Where("source_reference = ?", strings.TrimSpace(sourceReference)).
		Where("enabled = ?", true).
		Order("updated_at DESC").
		Order("created_at DESC").
		Order("id DESC").
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *GlobalIPBlockRuleRepository) FindActiveBySourceReference(source, sourceReference string, now time.Time) (*security.IPBlockRule, error) {
	rules, err := r.ListActiveBySourceReference(source, sourceReference, now)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &rules[0], nil
}

func (r *GlobalIPBlockRuleRepository) ListActiveBySourceReference(source, sourceReference string, now time.Time) ([]security.IPBlockRule, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	var rules []security.IPBlockRule
	err := r.db.
		Where("source = ?", strings.TrimSpace(source)).
		Where("source_reference = ?", strings.TrimSpace(sourceReference)).
		Where("enabled = ?", true).
		Where("(expires_at IS NULL OR expires_at > ?)", now).
		Order("updated_at DESC").
		Order("created_at DESC").
		Order("id DESC").
		Find(&rules).Error
	return rules, err
}

func (r *GlobalIPBlockRuleRepository) Create(rule *security.IPBlockRule) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if rule == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Create(rule).Error
}

func (r *GlobalIPBlockRuleRepository) Update(rule *security.IPBlockRule) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if rule == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Save(rule).Error
}

func (r *GlobalIPBlockRuleRepository) Disable(id, disabledBy uint, disabledAt time.Time) (*security.IPBlockRule, error) {
	_, after, err := r.DisableWithPrevious(context.Background(), id, disabledBy, disabledAt)
	return after, err
}

func (r *GlobalIPBlockRuleRepository) DisableWithPrevious(
	ctx context.Context,
	id,
	disabledBy uint,
	disabledAt time.Time,
) (*security.IPBlockRule, *security.IPBlockRule, error) {
	if r == nil || r.db == nil {
		return nil, nil, gorm.ErrInvalidDB
	}
	if id == 0 {
		return nil, nil, gorm.ErrRecordNotFound
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if disabledAt.IsZero() {
		disabledAt = time.Now().UTC()
	} else {
		disabledAt = disabledAt.UTC()
	}

	var before *security.IPBlockRule
	var after *security.IPBlockRule
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var probe security.IPBlockRule
		if err := tx.First(&probe, id).Error; err != nil {
			return err
		}
		if err := lockGlobalIPBlockSourceReference(tx, probe.Source, probe.SourceReference); err != nil {
			return err
		}

		var current security.IPBlockRule
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current, id).Error; err != nil {
			return err
		}
		beforeCopy := current
		before = &beforeCopy
		if !current.Enabled {
			afterCopy := current
			after = &afterCopy
			return nil
		}

		updates := map[string]interface{}{
			"enabled":     false,
			"disabled_at": disabledAt,
		}
		if disabledBy > 0 {
			updates["disabled_by"] = disabledBy
		}
		result := tx.Model(&security.IPBlockRule{}).
			Where("id = ?", id).
			Where("enabled = ?", true).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.First(&current, id).Error; err != nil {
			return err
		}
		afterCopy := current
		after = &afterCopy
		return nil
	})
	return before, after, err
}

func (r *GlobalIPBlockRuleRepository) DisableBySourceReference(
	source,
	sourceReference string,
	disabledBy uint,
	disabledAt time.Time,
	now time.Time,
) ([]security.IPBlockRule, error) {
	_, after, err := r.DisableBySourceReferenceWithPrevious(
		context.Background(),
		source,
		sourceReference,
		disabledBy,
		disabledAt,
		now,
	)
	return after, err
}

func (r *GlobalIPBlockRuleRepository) DisableBySourceReferenceWithPrevious(
	ctx context.Context,
	source,
	sourceReference string,
	disabledBy uint,
	disabledAt time.Time,
	now time.Time,
) ([]security.IPBlockRule, []security.IPBlockRule, error) {
	if r == nil || r.db == nil {
		return nil, nil, gorm.ErrInvalidDB
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if disabledAt.IsZero() {
		disabledAt = time.Now().UTC()
	} else {
		disabledAt = disabledAt.UTC()
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	var beforeRules []security.IPBlockRule
	var afterRules []security.IPBlockRule
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockGlobalIPBlockSourceReference(tx, source, sourceReference); err != nil {
			return err
		}

		var activeRules []security.IPBlockRule
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("source = ?", strings.TrimSpace(source)).
			Where("source_reference = ?", strings.TrimSpace(sourceReference)).
			Where("enabled = ?", true).
			Where("(expires_at IS NULL OR expires_at > ?)", now).
			Order("updated_at DESC").
			Order("created_at DESC").
			Order("id DESC").
			Find(&activeRules).Error; err != nil {
			return err
		}
		if len(activeRules) == 0 {
			return gorm.ErrRecordNotFound
		}
		beforeRules = append([]security.IPBlockRule(nil), activeRules...)

		ids := make([]uint, 0, len(activeRules))
		for _, rule := range activeRules {
			ids = append(ids, rule.ID)
		}

		updates := map[string]interface{}{
			"enabled":     false,
			"disabled_at": disabledAt,
		}
		if disabledBy > 0 {
			updates["disabled_by"] = disabledBy
		}
		result := tx.Model(&security.IPBlockRule{}).
			Where("id IN ?", ids).
			Where("enabled = ?", true).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return tx.
			Where("id IN ?", ids).
			Order("updated_at DESC").
			Order("created_at DESC").
			Order("id DESC").
			Find(&afterRules).Error
	})
	return beforeRules, afterRules, err
}

func lockGlobalIPBlockSourceReference(tx *gorm.DB, source, sourceReference string) error {
	if tx == nil || tx.Dialector == nil || tx.Dialector.Name() != "postgres" {
		return nil
	}
	return tx.Exec(
		"SELECT pg_advisory_xact_lock(?)",
		globalIPBlockSourceReferenceLockKey(source, sourceReference),
	).Error
}

func globalIPBlockSourceReferenceLockKey(source, sourceReference string) int64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte("global_ip_block_rules\x00"))
	_, _ = hash.Write([]byte(strings.TrimSpace(source)))
	_, _ = hash.Write([]byte("\x00"))
	_, _ = hash.Write([]byte(strings.TrimSpace(sourceReference)))
	return int64(hash.Sum64())
}

func (r *GlobalIPBlockRuleRepository) ListActive(now time.Time) ([]security.IPBlockRule, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	var rules []security.IPBlockRule
	err := r.db.
		Where("enabled = ?", true).
		Where("(expires_at IS NULL OR expires_at > ?)", now).
		Order("created_at DESC").
		Order("id DESC").
		Find(&rules).Error
	return rules, err
}

func (r *GlobalIPBlockRuleRepository) List(page, pageSize int, filters IPBlockRuleListFilters, now time.Time) ([]security.IPBlockRule, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, gorm.ErrInvalidDB
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	query := r.db.Model(&security.IPBlockRule{})
	switch strings.ToLower(strings.TrimSpace(filters.Status)) {
	case security.IPBlockRuleStatusActive:
		query = query.
			Where("enabled = ?", true).
			Where("(expires_at IS NULL OR expires_at > ?)", now)
	case security.IPBlockRuleStatusExpired:
		query = query.Where("enabled = ?", true).Where("expires_at IS NOT NULL AND expires_at <= ?", now)
	case security.IPBlockRuleStatusDisabled:
		query = query.Where("enabled = ?", false)
	}

	if source := strings.TrimSpace(filters.Source); source != "" {
		query = query.Where("source = ?", source)
	}
	if search := strings.ToLower(strings.TrimSpace(filters.Search)); search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"LOWER(cidr) LIKE ? OR LOWER(source) LIKE ? OR LOWER(source_reference) LIKE ? OR LOWER(reason) LIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var rules []security.IPBlockRule
	err := query.
		Order("enabled DESC").
		Order("created_at DESC").
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&rules).Error
	return rules, total, err
}
