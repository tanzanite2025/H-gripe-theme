package service

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	mediadomain "commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

const mediaDerivativePresetSortOrderLimit = 100000

// MediaDerivativePresetInput is the mutable configuration for one generated
// image size. Code is accepted only when creating a preset.
type MediaDerivativePresetInput struct {
	Code      string `json:"code"`
	Label     string `json:"label"`
	MaxWidth  int    `json:"max_width"`
	SortOrder int    `json:"sort_order"`
	Enabled   *bool  `json:"enabled"`
}

func (s *MediaService) ListMediaDerivativePresets() ([]mediadomain.MediaDerivativePreset, error) {
	repo, err := s.mediaDerivativePresetRepository()
	if err != nil {
		return nil, err
	}

	presets, err := repo.ListAll()
	if err != nil {
		return nil, err
	}
	if presets == nil {
		return nil, ErrMediaDerivativePresetUnavailable
	}
	return presets, nil
}

func (s *MediaService) GetMediaDerivativePreset(id uint) (*mediadomain.MediaDerivativePreset, error) {
	repo, err := s.mediaDerivativePresetRepository()
	if err != nil {
		return nil, err
	}

	preset, err := repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMediaDerivativePresetNotFound
	}
	if err != nil {
		return nil, err
	}
	return preset, nil
}

func (s *MediaService) CreateMediaDerivativePreset(input MediaDerivativePresetInput) (*mediadomain.MediaDerivativePreset, error) {
	repo, err := s.mediaDerivativePresetRepository()
	if err != nil {
		return nil, err
	}

	preset, err := mediaDerivativePresetFromInput(input)
	if err != nil {
		return nil, err
	}

	if err := repo.Transaction(func(tx *gorm.DB) error {
		txRepo := repo.WithTx(tx)
		if preset.Enabled {
			if err := ensureMediaDerivativePresetCapacity(txRepo, true, false); err != nil {
				return err
			}
		}
		if _, err := txRepo.FindByCode(preset.Code); err == nil {
			return ErrMediaDerivativePresetConflict
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := txRepo.Create(&preset); err != nil {
			if isMediaDerivativePresetConflictError(err) {
				return ErrMediaDerivativePresetConflict
			}
			return err
		}
		if preset.Enabled {
			return s.enqueueMediaDerivativeRebuildInTx(tx, "preset_created")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &preset, nil
}

func (s *MediaService) UpdateMediaDerivativePreset(id uint, input MediaDerivativePresetInput) (*mediadomain.MediaDerivativePreset, error) {
	repo, err := s.mediaDerivativePresetRepository()
	if err != nil {
		return nil, err
	}

	var next mediadomain.MediaDerivativePreset
	if err := repo.Transaction(func(tx *gorm.DB) error {
		txRepo := repo.WithTx(tx)
		current, err := txRepo.FindByID(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMediaDerivativePresetNotFound
		}
		if err != nil {
			return err
		}
		if requestedCode := normalizeMediaDerivativePresetCode(input.Code); requestedCode != "" && requestedCode != current.Code {
			return ErrMediaDerivativePresetCodeImmutable
		}
		if err := validateMediaDerivativePresetWidth(input.MaxWidth); err != nil {
			return err
		}
		if err := validateMediaDerivativePresetSortOrder(input.SortOrder); err != nil {
			return err
		}

		next = *current
		next.Label = normalizeMediaDerivativePresetLabel(input.Label, current.Label)
		if next.Label == "" {
			next.Label = current.Code
		}
		if utf8.RuneCountInString(next.Label) > 120 {
			return ErrInvalidMediaDerivativePreset
		}
		next.MaxWidth = input.MaxWidth
		next.SortOrder = input.SortOrder
		if input.Enabled != nil {
			next.Enabled = *input.Enabled
		}
		if err := ensureMediaDerivativePresetCapacity(txRepo, next.Enabled, current.Enabled); err != nil {
			return err
		}
		if next.MaxWidth != current.MaxWidth {
			next.GenerationVersion = normalizeMediaDerivativeVersion(current.GenerationVersion) + 1
		}
		if err := txRepo.Update(&next); err != nil {
			return err
		}
		if next.Enabled != current.Enabled || next.MaxWidth != current.MaxWidth {
			return s.enqueueMediaDerivativeRebuildInTx(tx, "preset_changed")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &next, nil
}

func (s *MediaService) SetMediaDerivativePresetEnabled(id uint, enabled bool) (*mediadomain.MediaDerivativePreset, error) {
	repo, err := s.mediaDerivativePresetRepository()
	if err != nil {
		return nil, err
	}

	var next *mediadomain.MediaDerivativePreset
	if err := repo.Transaction(func(tx *gorm.DB) error {
		txRepo := repo.WithTx(tx)
		current, err := txRepo.FindByID(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMediaDerivativePresetNotFound
		}
		if err != nil {
			return err
		}
		if current.Enabled == enabled {
			next = current
			return nil
		}
		if err := ensureMediaDerivativePresetCapacity(txRepo, enabled, current.Enabled); err != nil {
			return err
		}
		current.Enabled = enabled
		if err := txRepo.Update(current); err != nil {
			return err
		}
		next = current
		return s.enqueueMediaDerivativeRebuildInTx(tx, "preset_enabled_changed")
	}); err != nil {
		return nil, err
	}
	return next, nil
}

func (s *MediaService) DeleteMediaDerivativePreset(id uint) error {
	repo, err := s.mediaDerivativePresetRepository()
	if err != nil {
		return err
	}

	return repo.Transaction(func(tx *gorm.DB) error {
		txRepo := repo.WithTx(tx)
		preset, err := txRepo.FindByID(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMediaDerivativePresetNotFound
		}
		if err != nil {
			return err
		}
		if preset.IsSystem {
			return ErrMediaDerivativePresetProtected
		}

		used, err := txRepo.CountDerivativesByPreset(preset.Code)
		if err != nil {
			return err
		}
		if used > 0 {
			return ErrMediaDerivativePresetInUse
		}
		if err := txRepo.Delete(preset); err != nil {
			return err
		}
		return s.enqueueMediaDerivativeRebuildInTx(tx, "preset_deleted")
	})
}

func (s *MediaService) mediaDerivativePresetRepository() (*repository.MediaDerivativePresetRepository, error) {
	if s == nil || s.derivativePresets == nil {
		return nil, ErrMediaDerivativePresetUnavailable
	}
	return s.derivativePresets, nil
}

func mediaDerivativePresetFromInput(input MediaDerivativePresetInput) (mediadomain.MediaDerivativePreset, error) {
	code := normalizeMediaDerivativePresetCode(input.Code)
	if !isValidMediaDerivativePresetCode(code) {
		return mediadomain.MediaDerivativePreset{}, ErrInvalidMediaDerivativePreset
	}
	if err := validateMediaDerivativePresetWidth(input.MaxWidth); err != nil {
		return mediadomain.MediaDerivativePreset{}, err
	}
	if err := validateMediaDerivativePresetSortOrder(input.SortOrder); err != nil {
		return mediadomain.MediaDerivativePreset{}, err
	}

	label := normalizeMediaDerivativePresetLabel(input.Label, code)
	if utf8.RuneCountInString(label) > 120 {
		return mediadomain.MediaDerivativePreset{}, ErrInvalidMediaDerivativePreset
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	return mediadomain.MediaDerivativePreset{
		Code:              code,
		Label:             label,
		MaxWidth:          input.MaxWidth,
		SortOrder:         input.SortOrder,
		Enabled:           enabled,
		GenerationVersion: 1,
	}, nil
}

func normalizeMediaDerivativePresetCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeMediaDerivativePresetLabel(value string, fallback string) string {
	label := strings.TrimSpace(value)
	if label == "" {
		return strings.TrimSpace(fallback)
	}
	return label
}

func isValidMediaDerivativePresetCode(code string) bool {
	if len(code) < 2 || len(code) > 40 {
		return false
	}
	for index := 0; index < len(code); index++ {
		character := code[index]
		if (character >= 'a' && character <= 'z') ||
			(character >= '0' && character <= '9') ||
			character == '_' ||
			character == '-' {
			continue
		}
		return false
	}
	return true
}

func validateMediaDerivativePresetWidth(width int) error {
	if width <= 0 || width > mediaDerivativeMaxDimension {
		return ErrInvalidMediaDerivativePreset
	}
	return nil
}

func validateMediaDerivativePresetSortOrder(sortOrder int) error {
	if sortOrder < -mediaDerivativePresetSortOrderLimit || sortOrder > mediaDerivativePresetSortOrderLimit {
		return ErrInvalidMediaDerivativePreset
	}
	return nil
}

func ensureMediaDerivativePresetCapacity(
	repo *repository.MediaDerivativePresetRepository,
	nextEnabled bool,
	currentEnabled bool,
) error {
	if repo == nil || nextEnabled == currentEnabled {
		return nil
	}
	if err := repo.LockActiveConfiguration(); err != nil {
		return err
	}
	active, err := repo.CountActive()
	if err != nil {
		return err
	}
	if nextEnabled {
		if active >= mediaDerivativeActivePresetLimit {
			return ErrMediaDerivativePresetLimitReached
		}
	}
	return nil
}

func (s *MediaService) enqueueMediaDerivativeRebuildInTx(tx *gorm.DB, reason string) error {
	if s == nil || s.derivativeRebuildJobs == nil {
		return nil
	}
	_, _, err := s.derivativeRebuildJobs.WithTx(tx).EnqueueInTx(reason, time.Now().UTC())
	return err
}

func isMediaDerivativePresetConflictError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique") || strings.Contains(message, "duplicate")
}
