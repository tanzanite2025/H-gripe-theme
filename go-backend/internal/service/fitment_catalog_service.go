package service

import (
	"errors"
	"fmt"
	"strings"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrFrameFitmentEntryNotFound  = errors.New("frame fitment entry not found")
	ErrFrameFitmentEntryInvalid   = errors.New("frame fitment entry is invalid")
	ErrFrameFitmentEntryDuplicate = errors.New("frame fitment entry already exists")
)

type FrameFitmentEntryService struct {
	repo            *repository.FrameFitmentEntryRepository
	hubRepo         *repository.FitmentHubSpecificationRepository
	associationRepo *repository.FitmentFrameHubSpecificationRepository
}

type FrameFitmentEntryListInput struct {
	Page      int
	PageSize  int
	Search    string
	IsEnabled *bool
	Year      *int
}

type FrameFitmentEntryInput struct {
	BrandName           string
	ModelName           string
	SeriesName          string
	GenerationName      string
	YearMode            fitmentcatalogdomain.YearMode
	YearFrom            *int
	YearTo              *int
	MarketCode          string
	Notes               string
	IsEnabled           bool
	SortOrder           int
	HubSpecificationIDs []uint
}

func NewFrameFitmentEntryService(
	repo *repository.FrameFitmentEntryRepository,
	hubRepo *repository.FitmentHubSpecificationRepository,
	associationRepos ...*repository.FitmentFrameHubSpecificationRepository,
) *FrameFitmentEntryService {
	var associationRepo *repository.FitmentFrameHubSpecificationRepository
	if len(associationRepos) > 0 {
		associationRepo = associationRepos[0]
	}
	if associationRepo == nil && hubRepo != nil {
		associationRepo = hubRepo.FrameHubSpecificationRepository()
	}
	return &FrameFitmentEntryService{
		repo:            repo,
		hubRepo:         hubRepo,
		associationRepo: associationRepo,
	}
}

func (s *FrameFitmentEntryService) ConfigureHubSpecificationRepository(
	hubRepo *repository.FitmentHubSpecificationRepository,
) {
	if s == nil {
		return
	}
	s.hubRepo = hubRepo
}

func (s *FrameFitmentEntryService) List(input FrameFitmentEntryListInput) ([]fitmentcatalogdomain.FrameFitmentEntry, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, errors.New("frame fitment service is unavailable")
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}
	if len(strings.TrimSpace(input.Search)) > 120 {
		return nil, 0, fmt.Errorf("%w: search is too long", ErrFrameFitmentEntryInvalid)
	}
	if err := validateFitmentYearFilter(input.Year, ErrFrameFitmentEntryInvalid); err != nil {
		return nil, 0, err
	}

	entries, total, err := s.repo.List(input.Page, input.PageSize, repository.FrameFitmentEntryFilter{
		Search:    input.Search,
		IsEnabled: input.IsEnabled,
		Year:      input.Year,
	})
	if err != nil {
		return nil, 0, err
	}
	if s.associationRepo == nil || len(entries) == 0 {
		for index := range entries {
			entries[index].HubSpecifications = []fitmentcatalogdomain.HubSpecification{}
		}
		return entries, total, nil
	}

	ids := make([]uint, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.ID)
	}
	counts, err := s.associationRepo.CountFrameReferencesByFrameIDs(ids)
	if err != nil {
		return nil, 0, err
	}
	for index := range entries {
		entries[index].HubSpecificationCount = counts[entries[index].ID]
		entries[index].HubSpecifications = []fitmentcatalogdomain.HubSpecification{}
	}
	return entries, total, nil
}

func (s *FrameFitmentEntryService) Get(id uint) (*fitmentcatalogdomain.FrameFitmentEntry, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("frame fitment service is unavailable")
	}
	entry, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFrameFitmentEntryNotFound
	}
	if err != nil {
		return nil, err
	}
	if s.associationRepo != nil {
		specifications, loadErr := s.associationRepo.ListForFrame(id)
		if loadErr != nil {
			return nil, loadErr
		}
		entry.HubSpecifications = specifications
		if entry.HubSpecifications == nil {
			entry.HubSpecifications = []fitmentcatalogdomain.HubSpecification{}
		}
		entry.HubSpecificationCount = len(specifications)
	}
	return entry, err
}

func (s *FrameFitmentEntryService) Create(input FrameFitmentEntryInput) (*fitmentcatalogdomain.FrameFitmentEntry, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("frame fitment service is unavailable")
	}

	entry := frameFitmentEntryFromInput(input)
	entry.IsEnabled = false
	if err := validateFrameFitmentEntry(entry); err != nil {
		return nil, err
	}
	hubSpecificationIDs, err := normalizeHubSpecificationIDs(input.HubSpecificationIDs, ErrFrameFitmentEntryInvalid)
	if err != nil {
		return nil, err
	}
	if err := s.validateFrameHubSpecifications(entry.IsEnabled, nil, hubSpecificationIDs); err != nil {
		return nil, err
	}

	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		frameRepo := s.repo.WithTx(tx)
		if _, err := frameRepo.FindDuplicate(entry, 0); err == nil {
			return ErrFrameFitmentEntryDuplicate
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := frameRepo.Create(entry); err != nil {
			if isUniqueViolation(err) {
				return ErrFrameFitmentEntryDuplicate
			}
			return err
		}
		if s.associationRepo == nil {
			return nil
		}
		return s.associationRepo.WithTx(tx).ReplaceFrameSpecifications(entry.ID, hubSpecificationIDs)
	}); err != nil {
		if errors.Is(err, ErrFrameFitmentEntryDuplicate) || isUniqueViolation(err) {
			return nil, ErrFrameFitmentEntryDuplicate
		}
		return nil, err
	}
	return s.Get(entry.ID)
}

func (s *FrameFitmentEntryService) Update(id uint, input FrameFitmentEntryInput) (*fitmentcatalogdomain.FrameFitmentEntry, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("frame fitment service is unavailable")
	}

	existing, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	entry := frameFitmentEntryFromInput(input)
	entry.ID = id
	if err := validateFrameFitmentEntry(entry); err != nil {
		return nil, err
	}
	hubSpecificationIDs, err := normalizeHubSpecificationIDs(input.HubSpecificationIDs, ErrFrameFitmentEntryInvalid)
	if err != nil {
		return nil, err
	}
	existingHubIDs := hubSpecificationIDsFromEntry(existing)
	if err := s.validateFrameHubSpecifications(entry.IsEnabled, existingHubIDs, hubSpecificationIDs); err != nil {
		return nil, err
	}

	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		frameRepo := s.repo.WithTx(tx)
		if _, err := frameRepo.FindDuplicate(entry, id); err == nil {
			return ErrFrameFitmentEntryDuplicate
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := frameRepo.Update(entry); err != nil {
			if isUniqueViolation(err) {
				return ErrFrameFitmentEntryDuplicate
			}
			return err
		}
		if s.associationRepo == nil {
			return nil
		}
		return s.associationRepo.WithTx(tx).ReplaceFrameSpecifications(id, hubSpecificationIDs)
	}); err != nil {
		if errors.Is(err, ErrFrameFitmentEntryDuplicate) || isUniqueViolation(err) {
			return nil, ErrFrameFitmentEntryDuplicate
		}
		return nil, err
	}
	return s.Get(id)
}

func (s *FrameFitmentEntryService) UpdateStatus(id uint, enabled bool) (*fitmentcatalogdomain.FrameFitmentEntry, error) {
	entry, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if enabled {
		if err := validateFrameFitmentEntry(entry); err != nil {
			return nil, fmt.Errorf("%w: cannot enable incomplete entry: %v", ErrFrameFitmentEntryInvalid, err)
		}
		hubSpecificationIDs := hubSpecificationIDListFromEntry(entry)
		if err := s.validateFrameHubSpecifications(true, hubSpecificationIDsFromEntry(entry), hubSpecificationIDs); err != nil {
			return nil, err
		}
	}
	if err := s.repo.UpdateStatus(id, enabled); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *FrameFitmentEntryService) Delete(id uint) error {
	if s == nil || s.repo == nil {
		return errors.New("frame fitment service is unavailable")
	}
	if _, err := s.Get(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func frameFitmentEntryFromInput(input FrameFitmentEntryInput) *fitmentcatalogdomain.FrameFitmentEntry {
	return &fitmentcatalogdomain.FrameFitmentEntry{
		BrandName:      input.BrandName,
		ModelName:      input.ModelName,
		SeriesName:     input.SeriesName,
		GenerationName: input.GenerationName,
		YearMode:       input.YearMode,
		YearFrom:       input.YearFrom,
		YearTo:         input.YearTo,
		MarketCode:     input.MarketCode,
		Notes:          input.Notes,
		IsEnabled:      input.IsEnabled,
		SortOrder:      input.SortOrder,
	}
}

func validateFrameFitmentEntry(entry *fitmentcatalogdomain.FrameFitmentEntry) error {
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrFrameFitmentEntryInvalid, err)
	}
	return nil
}

func validateFitmentYearFilter(year *int, invalidErr error) error {
	if year == nil {
		return nil
	}
	if *year < 1800 || *year > 2200 {
		return fmt.Errorf("%w: year must be between 1800 and 2200", invalidErr)
	}
	return nil
}

func (s *FrameFitmentEntryService) validateFrameHubSpecifications(
	frameEnabled bool,
	existingHubIDs map[uint]struct{},
	hubSpecificationIDs []uint,
) error {
	if len(hubSpecificationIDs) == 0 {
		if frameEnabled {
			return fmt.Errorf("%w: at least one enabled rear hub specification is required", ErrFrameFitmentEntryInvalid)
		}
		return nil
	}
	if s.hubRepo == nil {
		return fmt.Errorf("%w: hub specification repository is unavailable", ErrFrameFitmentEntryInvalid)
	}

	specifications, err := s.hubRepo.FindByIDs(hubSpecificationIDs)
	if err != nil {
		return err
	}
	if len(specifications) != len(hubSpecificationIDs) {
		return fmt.Errorf("%w: one or more hub specifications do not exist", ErrFrameFitmentEntryInvalid)
	}

	for _, specification := range specifications {
		if specification.Position != fitmentcatalogdomain.HubPositionRear {
			return fmt.Errorf("%w: frame entries can only use rear hub specifications", ErrFrameFitmentEntryInvalid)
		}
		if !specification.IsEnabled {
			if _, wasAlreadyAssociated := existingHubIDs[specification.ID]; !wasAlreadyAssociated || frameEnabled {
				return fmt.Errorf("%w: disabled hub specifications cannot be selected or enabled", ErrFrameFitmentEntryInvalid)
			}
		}
	}

	return nil
}

func normalizeHubSpecificationIDs(ids []uint, invalidErr error) ([]uint, error) {
	if len(ids) == 0 {
		return []uint{}, nil
	}

	normalized := make([]uint, 0, len(ids))
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return nil, fmt.Errorf("%w: hub specification IDs must be positive", invalidErr)
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized, nil
}

func hubSpecificationIDsFromEntry(entry *fitmentcatalogdomain.FrameFitmentEntry) map[uint]struct{} {
	ids := make(map[uint]struct{})
	if entry == nil {
		return ids
	}
	for _, specification := range entry.HubSpecifications {
		ids[specification.ID] = struct{}{}
	}
	return ids
}

func hubSpecificationIDListFromEntry(entry *fitmentcatalogdomain.FrameFitmentEntry) []uint {
	if entry == nil || len(entry.HubSpecifications) == 0 {
		return []uint{}
	}
	ids := make([]uint, 0, len(entry.HubSpecifications))
	for _, specification := range entry.HubSpecifications {
		ids = append(ids, specification.ID)
	}
	return ids
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
