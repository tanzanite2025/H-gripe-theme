package repository

import (
	"context"
	"sort"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/setting"
)

func (r *SettingRepository) ListAllLocalesForSettingsConfigurationBackup(ctx context.Context, limit int) ([]setting.Setting, error) {
	var rows []setting.Setting
	err := r.db.WithContext(ctx).Order("key ASC, locale ASC").Limit(limit).Find(&rows).Error
	return rows, err
}

// The same row reservations and lock order as ordinary setting writes protect
// both missing and existing keys. A stale preview rolls back all reservations.
func (r *SettingRepository) ImportSettingsConfigurationAfterLockedSnapshotValidation(ctx context.Context, values []setting.Setting, validateSnapshot func([]setting.Setting) error) error {
	identities := make([]settingKeyAndLocaleIdentity, 0, len(values))
	for _, value := range values {
		identities = append(identities, settingKeyAndLocaleIdentity{Key: value.Key, Locale: value.Locale})
	}
	sort.Slice(identities, func(i, j int) bool {
		if identities[i].Key == identities[j].Key {
			return identities[i].Locale < identities[j].Locale
		}
		return identities[i].Key < identities[j].Key
	})
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockSiteIdentitySettingNamespacesBeforeSortedSettingRows(tx, identities); err != nil {
			return err
		}
		current := make([]setting.Setting, 0, len(values))
		for _, identity := range identities {
			row, created, err := lockOrReserveSettingKeyAndLocaleForTransactionalMediaWrite(tx, identity)
			if err != nil {
				return err
			}
			if !created {
				current = append(current, row)
			}
		}
		if err := validateSnapshot(current); err != nil {
			return err
		}
		return r.WithTx(tx).BatchSet(values)
	})
}
