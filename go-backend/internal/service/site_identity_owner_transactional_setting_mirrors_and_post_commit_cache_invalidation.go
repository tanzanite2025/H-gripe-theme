package service

import (
	"commerce-platform/internal/domain/setting"
	"gorm.io/gorm"
)

func (s *SettingService) synchronizeDedicatedSiteIdentityMirrorsFromOwnerTransaction(tx *gorm.DB, key, locale, publicSiteBaseURL string) ([]setting.Setting, error) {
	if s == nil || s.settingRepo == nil || tx == nil {
		return nil, ErrSettingInvalid
	}
	mirrors, err := s.settingRepo.WithTx(tx).SynchronizeDedicatedSiteIdentitySettingMirrorsFromCurrentOwner(key, locale, publicSiteBaseURL)
	return mirrors, mapSettingRepositoryMediaErrorToInvalidSettingValue(err)
}

func (s *SettingService) invalidateDedicatedSiteIdentityMirrorCachesAfterOwnerCommit(mirrors []setting.Setting) {
	for _, mirror := range mirrors {
		s.invalidateSettingCaches(mirror.Key, mirror.Group, mirror.Locale)
	}
}
