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
	ErrFitmentHubSpecificationNotFound  = errors.New("fitment hub specification not found")
	ErrFitmentHubSpecificationInvalid   = errors.New("fitment hub specification is invalid")
	ErrFitmentHubSpecificationDuplicate = errors.New("fitment hub specification already exists")
	ErrFitmentHubSpecificationInUse     = errors.New("fitment hub specification is in use")
)

type FitmentHubSpecificationService struct {
	repo                *repository.FitmentHubSpecificationRepository
	associationRepo     *repository.FitmentFrameHubSpecificationRepository
	forkAssociationRepo *repository.FitmentForkHubSpecificationRepository
	spokeRepo           *repository.SpokeRepository
}

type FitmentHubSpecificationListInput struct {
	Page      int
	PageSize  int
	Search    string
	Position  fitmentcatalogdomain.HubPosition
	AxleType  fitmentcatalogdomain.HubAxleType
	IsEnabled *bool
}

type FitmentHubSpecificationInput struct {
	SpecCode      string
	DisplayName   string
	Position      fitmentcatalogdomain.HubPosition
	AxleType      fitmentcatalogdomain.HubAxleType
	AxleSpacingMM int
	WRMM          *float64
	WLMM          *float64
	PCDRMM        *float64
	PCDLMM        *float64
	Notes         string
	IsEnabled     bool
	SortOrder     int
}

func NewFitmentHubSpecificationService(
	repo *repository.FitmentHubSpecificationRepository,
	associationRepos ...*repository.FitmentFrameHubSpecificationRepository,
) *FitmentHubSpecificationService {
	var associationRepo *repository.FitmentFrameHubSpecificationRepository
	if len(associationRepos) > 0 {
		associationRepo = associationRepos[0]
	}
	if associationRepo == nil && repo != nil {
		associationRepo = repo.FrameHubSpecificationRepository()
	}
	return &FitmentHubSpecificationService{
		repo:            repo,
		associationRepo: associationRepo,
	}
}

func (s *FitmentHubSpecificationService) ConfigureForkHubSpecificationRepository(
	associationRepo *repository.FitmentForkHubSpecificationRepository,
) {
	if s == nil {
		return
	}
	s.forkAssociationRepo = associationRepo
}

func (s *FitmentHubSpecificationService) ConfigureSpokeRepository(
	spokeRepo *repository.SpokeRepository,
) {
	if s == nil {
		return
	}
	s.spokeRepo = spokeRepo
}

func (s *FitmentHubSpecificationService) List(
	input FitmentHubSpecificationListInput,
) ([]fitmentcatalogdomain.HubSpecification, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, errors.New("fitment hub specification service is unavailable")
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}
	if len(strings.TrimSpace(input.Search)) > 120 {
		return nil, 0, fmt.Errorf("%w: search is too long", ErrFitmentHubSpecificationInvalid)
	}
	if input.Position != "" && !isValidFitmentHubPosition(input.Position) {
		return nil, 0, fmt.Errorf("%w: unsupported position %q", ErrFitmentHubSpecificationInvalid, input.Position)
	}
	if input.AxleType != "" && !isValidFitmentHubAxleType(input.AxleType) {
		return nil, 0, fmt.Errorf("%w: unsupported axle_type %q", ErrFitmentHubSpecificationInvalid, input.AxleType)
	}

	var position *fitmentcatalogdomain.HubPosition
	if input.Position != "" {
		position = &input.Position
	}
	var axleType *fitmentcatalogdomain.HubAxleType
	if input.AxleType != "" {
		axleType = &input.AxleType
	}

	return s.repo.List(input.Page, input.PageSize, repository.FitmentHubSpecificationFilter{
		Search:    input.Search,
		Position:  position,
		AxleType:  axleType,
		IsEnabled: input.IsEnabled,
	})
}

func (s *FitmentHubSpecificationService) Get(id uint) (*fitmentcatalogdomain.HubSpecification, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("fitment hub specification service is unavailable")
	}
	specification, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFitmentHubSpecificationNotFound
	}
	return specification, err
}

func (s *FitmentHubSpecificationService) Create(
	input FitmentHubSpecificationInput,
) (*fitmentcatalogdomain.HubSpecification, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("fitment hub specification service is unavailable")
	}

	specification := fitmentHubSpecificationFromInput(input)
	specification.IsEnabled = false
	if err := validateFitmentHubSpecification(specification); err != nil {
		return nil, err
	}
	if _, err := s.repo.FindDuplicate(specification, 0); err == nil {
		return nil, ErrFitmentHubSpecificationDuplicate
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.repo.Create(specification); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrFitmentHubSpecificationDuplicate
		}
		return nil, err
	}
	return s.Get(specification.ID)
}

func (s *FitmentHubSpecificationService) Update(
	id uint,
	input FitmentHubSpecificationInput,
) (*fitmentcatalogdomain.HubSpecification, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("fitment hub specification service is unavailable")
	}
	existing, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	specification := fitmentHubSpecificationFromInput(input)
	specification.ID = id
	if err := validateFitmentHubSpecification(specification); err != nil {
		return nil, err
	}
	if existing.Position != specification.Position && (existing.FrameReferenceCount > 0 || existing.ForkReferenceCount > 0) {
		return nil, fmt.Errorf("%w: referenced hub specifications cannot change front/rear position", ErrFitmentHubSpecificationInvalid)
	}
	if _, err := s.repo.FindDuplicate(specification, id); err == nil {
		return nil, ErrFitmentHubSpecificationDuplicate
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.repo.Update(specification); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrFitmentHubSpecificationDuplicate
		}
		return nil, err
	}
	return s.Get(id)
}

func (s *FitmentHubSpecificationService) UpdateStatus(
	id uint,
	enabled bool,
) (*fitmentcatalogdomain.HubSpecification, error) {
	specification, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if enabled {
		if err := validateFitmentHubSpecification(specification); err != nil {
			return nil, err
		}
	}
	if err := s.repo.UpdateStatus(id, enabled); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *FitmentHubSpecificationService) Delete(id uint) error {
	if s == nil || s.repo == nil {
		return errors.New("fitment hub specification service is unavailable")
	}
	specification, err := s.Get(id)
	if err != nil {
		return err
	}

	if s.associationRepo == nil {
		return errors.New("fitment frame hub specification repository is unavailable")
	}
	count, err := s.associationRepo.FrameReferenceCount(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrFitmentHubSpecificationInUse
	}
	if s.forkAssociationRepo != nil {
		count, err := s.forkAssociationRepo.ForkReferenceCount(id)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrFitmentHubSpecificationInUse
		}
	}
	if s.spokeRepo != nil {
		count, err := s.spokeRepo.CountFitmentHubSpecificationReferences(id, specification.SpecCode)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrFitmentHubSpecificationInUse
		}
	}
	return s.repo.Delete(id)
}

func fitmentHubSpecificationFromInput(
	input FitmentHubSpecificationInput,
) *fitmentcatalogdomain.HubSpecification {
	return &fitmentcatalogdomain.HubSpecification{
		SpecCode:      input.SpecCode,
		DisplayName:   input.DisplayName,
		Position:      input.Position,
		AxleType:      input.AxleType,
		AxleSpacingMM: input.AxleSpacingMM,
		WRMM:          input.WRMM,
		WLMM:          input.WLMM,
		PCDRMM:        input.PCDRMM,
		PCDLMM:        input.PCDLMM,
		Notes:         input.Notes,
		IsEnabled:     input.IsEnabled,
		SortOrder:     input.SortOrder,
	}
}

func validateFitmentHubSpecification(
	specification *fitmentcatalogdomain.HubSpecification,
) error {
	if err := specification.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrFitmentHubSpecificationInvalid, err)
	}
	return nil
}

func isValidFitmentHubPosition(position fitmentcatalogdomain.HubPosition) bool {
	return position == fitmentcatalogdomain.HubPositionFront || position == fitmentcatalogdomain.HubPositionRear
}

func isValidFitmentHubAxleType(axleType fitmentcatalogdomain.HubAxleType) bool {
	switch axleType {
	case fitmentcatalogdomain.HubAxleTypeQuickRelease,
		fitmentcatalogdomain.HubAxleTypeThruAxle,
		fitmentcatalogdomain.HubAxleTypeBoltOn,
		fitmentcatalogdomain.HubAxleTypeOther:
		return true
	default:
		return false
	}
}
