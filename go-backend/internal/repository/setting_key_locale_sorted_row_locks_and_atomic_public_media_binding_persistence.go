package repository

import (
	"errors"
	"sort"

	"commerce-platform/internal/domain/setting"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *SettingRepository) ConfigurePublicBusinessMediaReferenceIdentityResolver(resolver PublicBusinessMediaReferenceObjectKeyResolver) {
	r.businessMediaReferenceResolver = resolver
}

// SetMultipleValuesPreservingExistingVisibilityAndDescription retains Set's
// metadata contract while validating every imported value in one transaction.
func (r *SettingRepository) SetMultipleValuesPreservingExistingVisibilityAndDescription(values []setting.Setting) error {
	return r.persistSettingsWithOneTransactionalMediaBindingPlan(values, false)
}

type settingKeyAndLocaleIdentity struct{ Key, Locale string }

// Empty rows are transaction-local reservations for missing keys. The existing
// unique index also serializes concurrent first writes without another table.
// All parents are locked before the single sorted asset lock plan.
func (r *SettingRepository) persistSettingsWithOneTransactionalMediaBindingPlan(values []setting.Setting, updateVisibilityAndDescription bool) error {
	if len(values) == 0 {
		return nil
	}
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	requested := make(map[settingKeyAndLocaleIdentity]setting.Setting, len(values))
	for _, value := range values {
		// Match GORM's existing default locale for newly created settings.
		if value.Locale == "" {
			value.Locale = "en"
		}
		requested[settingKeyAndLocaleIdentity{value.Key, value.Locale}] = value
	}
	identities := make([]settingKeyAndLocaleIdentity, 0, len(requested))
	for identity := range requested {
		identities = append(identities, identity)
	}
	sort.Slice(identities, func(i, j int) bool {
		if identities[i].Key == identities[j].Key {
			return identities[i].Locale < identities[j].Locale
		}
		return identities[i].Key < identities[j].Key
	})
	committed := make(map[settingKeyAndLocaleIdentity]setting.Setting, len(requested))
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := lockSiteIdentitySettingNamespacesBeforeSortedSettingRows(tx, identities); err != nil {
			return err
		}
		previous := make(map[settingKeyAndLocaleIdentity]setting.Setting, len(requested))
		newlyReserved := make(map[settingKeyAndLocaleIdentity]bool, len(requested))
		for _, identity := range identities {
			current, created, err := lockOrReserveSettingKeyAndLocaleForTransactionalMediaWrite(tx, identity)
			if err != nil {
				return err
			}
			previous[identity] = current
			newlyReserved[identity] = created
			value := requested[identity]
			value.ID, value.CreatedAt, value.DeletedAt = current.ID, current.CreatedAt, current.DeletedAt
			if created && value.Type == "" {
				value.Type = "string"
			}
			if !updateVisibilityAndDescription {
				if !created {
					value.IsPublic, value.Description = current.IsPublic, current.Description
				} else if !value.IsPublic {
					value.IsPublic = true
				}
			}
			committed[identity] = value
		}
		if err := r.validateAndCanonicalizeSharedImageAndWebsiteProfileRichTextSettingsWithOneSortedAssetLockPlan(tx, identities, previous, committed, newlyReserved); err != nil {
			return err
		}
		for _, identity := range identities {
			value := committed[identity]
			updates := map[string]interface{}{"value": value.Value, "type": value.Type, "group": value.Group, "is_public": value.IsPublic, "description": value.Description}
			if err := tx.Model(&value).Where("id = ?", value.ID).Updates(updates).Error; err != nil {
				return err
			}
			committed[identity] = value
		}
		return nil
	})
	if err != nil {
		return err
	}
	for index := range values {
		locale := values[index].Locale
		if locale == "" {
			locale = "en"
		}
		values[index] = committed[settingKeyAndLocaleIdentity{values[index].Key, locale}]
	}
	return nil
}

func lockOrReserveSettingKeyAndLocaleForTransactionalMediaWrite(tx *gorm.DB, identity settingKeyAndLocaleIdentity) (setting.Setting, bool, error) {
	var current setting.Setting
	loadCurrentSettingWithNoKeyUpdateLock := func() error {
		return tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).Where("key = ? AND locale = ?", identity.Key, identity.Locale).First(&current).Error
	}
	err := loadCurrentSettingWithNoKeyUpdateLock()
	if err == nil {
		return current, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return current, false, err
	}
	// Reserve only missing rows: ordinary edits need no speculative insert and
	// consume no sequence IDs. A competing first writer waits on the unique key.
	reservation := setting.Setting{Key: identity.Key, Locale: identity.Locale}
	insert := tx.Omit(clause.Associations).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "key"}, {Name: "locale"}}, DoNothing: true}).Create(&reservation)
	if insert.Error != nil {
		return current, false, insert.Error
	}
	err = loadCurrentSettingWithNoKeyUpdateLock()
	return current, insert.RowsAffected != 0, err
}
