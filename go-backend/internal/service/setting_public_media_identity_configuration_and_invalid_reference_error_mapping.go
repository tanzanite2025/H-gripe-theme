package service

import (
	"errors"
	"fmt"

	"commerce-platform/internal/repository"
)

func (s *SettingService) ConfigureMediaService(assets *MediaService) {
	if s != nil && s.settingRepo != nil {
		s.settingRepo.ConfigurePublicBusinessMediaReferenceIdentityResolver(assets)
	}
}

func mapSettingRepositoryMediaErrorToInvalidSettingValue(err error) error {
	if errors.Is(err, repository.ErrSettingMediaReferenceInvalid) {
		return fmt.Errorf("%w: %v", ErrSettingInvalid, err)
	}
	return err
}
