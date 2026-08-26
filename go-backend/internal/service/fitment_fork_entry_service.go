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
	ErrForkFitmentEntryNotFound  = errors.New("fork fitment entry not found")
	ErrForkFitmentEntryInvalid   = errors.New("fork fitment entry is invalid")
	ErrForkFitmentEntryDuplicate = errors.New("fork fitment entry already exists")
)

type ForkFitmentEntryService struct {
	repo            *repository.ForkFitmentEntryRepository
	hubRepo         *repository.FitmentHubSpecificationRepository
	associationRepo *repository.FitmentForkHubSpecificationRepository
}

type ForkFitmentEntryListInput struct {
	Page      int
	PageSize  int
	Search    string
	IsEnabled *bool
	Year      *int
}

type ForkFitmentEntryInput struct {
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

func NewForkFitmentEntryService(
	repo *repository.ForkFitmentEntryRepository,
	hubRepo *repository.FitmentHubSpecificationRepository,
	associationRepo *repository.FitmentForkHubSpecificationRepository,
) *ForkFitmentEntryService {
	return &ForkFitmentEntryService{
		repo:            repo,
		hubRepo:         hubRepo,
		associationRepo: associationRepo,
	}
}

func (s *ForkFitmentEntryService) List(input ForkFitmentEntryListInput) ([]fitmentcatalogdomain.ForkFitmentEntry, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, errors.New("fork fitment service is unavailable")
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}
	if len(strings.TrimSpace(input.Search)) > 120 {
		return nil, 0, fmt.Errorf("%w: search is too long", ErrForkFitmentEntryInvalid)
	}
	if err := validateFitmentYearFilter(input.Year, ErrForkFitmentEntryInvalid); err != nil {
		return nil, 0, err
	}

	entries, total, err := s.repo.List(input.Page, input.PageSize, repository.ForkFitmentEntryFilter{
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
	counts, err := s.associationRepo.CountForkReferencesByForkIDs(ids)
	if err != nil {
		return nil, 0, err
	}
	for index := range entries {
		entries[index].HubSpecificationCount = counts[entries[index].ID]
		entries[index].HubSpecifications = []fitmentcatalogdomain.HubSpecification{}
	}
	return entries, total, nil
}

func (s *ForkFitmentEntryService) Get(id uint) (*fitmentcatalogdomain.ForkFitmentEntry, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("fork fitment service is unavailable")
	}
	entry, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrForkFitmentEntryNotFound
	}
	if err != nil {
		return nil, err
	}
	if s.associationRepo != nil {
		specifications, loadErr := s.associationRepo.ListForFork(id)
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

func (s *ForkFitmentEntryService) Create(input ForkFitmentEntryInput) (*fitmentcatalogdomain.ForkFitmentEntry, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("fork fitment service is unavailable")
	}

	entry := forkFitmentEntryFromInput(input)
	entry.IsEnabled = false
	if err := validateForkFitmentEntry(entry); err != nil {
		return nil, err
	}
	hubSpecificationIDs, err := normalizeHubSpecificationIDs(input.HubSpecificationIDs, ErrForkFitmentEntryInvalid)
	if err != nil {
		return nil, err
	}
	if err := s.validateForkHubSpecifications(entry.IsEnabled, nil, hubSpecificationIDs); err != nil {
		return nil, err
	}

	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		forkRepo := s.repo.WithTx(tx)
		if _, err := forkRepo.FindDuplicate(entry, 0); err == nil {
			return ErrForkFitmentEntryDuplicate
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := forkRepo.Create(entry); err != nil {
			if isUniqueViolation(err) {
				return ErrForkFitmentEntryDuplicate
			}
			return err
		}
		if s.associationRepo == nil {
			return nil
		}
		return s.associationRepo.WithTx(tx).ReplaceForkSpecifications(entry.ID, hubSpecificationIDs)
	}); err != nil {
		if errors.Is(err, ErrForkFitmentEntryDuplicate) || isUniqueViolation(err) {
			return nil, ErrForkFitmentEntryDuplicate
		}
		return nil, err
	}
	return s.Get(entry.ID)
}

func (s *ForkFitmentEntryService) Update(id uint, input ForkFitmentEntryInput) (*fitmentcatalogdomain.ForkFitmentEntry, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("fork fitment service is unavailable")
	}

	existing, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	entry := forkFitmentEntryFromInput(input)
	entry.ID = id
	if err := validateForkFitmentEntry(entry); err != nil {
		return nil, err
	}
	hubSpecificationIDs, err := normalizeHubSpecificationIDs(input.HubSpecificationIDs, ErrForkFitmentEntryInvalid)
	if err != nil {
		return nil, err
	}
	existingHubIDs := hubSpecificationIDsFromForkEntry(existing)
	if err := s.validateForkHubSpecifications(entry.IsEnabled, existingHubIDs, hubSpecificationIDs); err != nil {
		return nil, err
	}

	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		forkRepo := s.repo.WithTx(tx)
		if _, err := forkRepo.FindDuplicate(entry, id); err == nil {
			return ErrForkFitmentEntryDuplicate
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := forkRepo.Update(entry); err != nil {
			if isUniqueViolation(err) {
				return ErrForkFitmentEntryDuplicate
			}
			return err
		}
		if s.associationRepo == nil {
			return nil
		}
		return s.associationRepo.WithTx(tx).ReplaceForkSpecifications(id, hubSpecificationIDs)
	}); err != nil {
		if errors.Is(err, ErrForkFitmentEntryDuplicate) || isUniqueViolation(err) {
			return nil, ErrForkFitmentEntryDuplicate
		}
		return nil, err
	}
	return s.Get(id)
}

func (s *ForkFitmentEntryService) UpdateStatus(id uint, enabled bool) (*fitmentcatalogdomain.ForkFitmentEntry, error) {
	entry, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if enabled {
		if err := validateForkFitmentEntry(entry); err != nil {
			return nil, fmt.Errorf("%w: cannot enable incomplete entry: %v", ErrForkFitmentEntryInvalid, err)
		}
		hubSpecificationIDs := hubSpecificationIDListFromForkEntry(entry)
		if err := s.validateForkHubSpecifications(true, hubSpecificationIDsFromForkEntry(entry), hubSpecificationIDs); err != nil {
			return nil, err
		}
	}
	if err := s.repo.UpdateStatus(id, enabled); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *ForkFitmentEntryService) Delete(id uint) error {
	if s == nil || s.repo == nil {
		return errors.New("fork fitment service is unavailable")
	}
	if _, err := s.Get(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func forkFitmentEntryFromInput(input ForkFitmentEntryInput) *fitmentcatalogdomain.ForkFitmentEntry {
	return &fitmentcatalogdomain.ForkFitmentEntry{
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

func validateForkFitmentEntry(entry *fitmentcatalogdomain.ForkFitmentEntry) error {
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrForkFitmentEntryInvalid, err)
	}
	return nil
}

func (s *ForkFitmentEntryService) validateForkHubSpecifications(
	forkEnabled bool,
	existingHubIDs map[uint]struct{},
	hubSpecificationIDs []uint,
) error {
	if len(hubSpecificationIDs) == 0 {
		if forkEnabled {
			return fmt.Errorf("%w: at least one enabled front hub specification is required", ErrForkFitmentEntryInvalid)
		}
		return nil
	}
	if s.hubRepo == nil {
		return fmt.Errorf("%w: hub specification repository is unavailable", ErrForkFitmentEntryInvalid)
	}

	specifications, err := s.hubRepo.FindByIDs(hubSpecificationIDs)
	if err != nil {
		return err
	}
	if len(specifications) != len(hubSpecificationIDs) {
		return fmt.Errorf("%w: one or more hub specifications do not exist", ErrForkFitmentEntryInvalid)
	}

	for _, specification := range specifications {
		if specification.Position != fitmentcatalogdomain.HubPositionFront {
			return fmt.Errorf("%w: fork entries can only use front hub specifications", ErrForkFitmentEntryInvalid)
		}
		if !specification.IsEnabled {
			if _, wasAlreadyAssociated := existingHubIDs[specification.ID]; !wasAlreadyAssociated || forkEnabled {
				return fmt.Errorf("%w: disabled hub specifications cannot be selected or enabled", ErrForkFitmentEntryInvalid)
			}
		}
	}

	return nil
}

func hubSpecificationIDsFromForkEntry(entry *fitmentcatalogdomain.ForkFitmentEntry) map[uint]struct{} {
	ids := make(map[uint]struct{})
	if entry == nil {
		return ids
	}
	for _, specification := range entry.HubSpecifications {
		ids[specification.ID] = struct{}{}
	}
	return ids
}

func hubSpecificationIDListFromForkEntry(entry *fitmentcatalogdomain.ForkFitmentEntry) []uint {
	if entry == nil || len(entry.HubSpecifications) == 0 {
		return []uint{}
	}
	ids := make([]uint, 0, len(entry.HubSpecifications))
	for _, specification := range entry.HubSpecifications {
		ids = append(ids, specification.ID)
	}
	return ids
}
