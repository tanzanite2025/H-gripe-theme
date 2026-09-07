package service

import (
	"errors"
	"fmt"
	"strings"

	productrequirement "commerce-platform/internal/domain/productrequirement"
	"commerce-platform/internal/repository"
)

var (
	ErrProductQualityRequirementStoreUnavailable = errors.New("product quality requirement store is unavailable")
	ErrProductQualityRequirementProductNotFound  = errors.New("product quality requirement product not found")
	ErrProductQualityRequirementVariantNotFound  = errors.New("product quality requirement variant not found")
	ErrProductQualityRequirementInvalid          = errors.New("product quality requirement is invalid")
)

type ProductQualityRequirementService struct {
	repo *repository.ProductQualityRequirementRepository
}

type ProductQualityRequirementInput struct {
	ProductID              uint
	VariantID              *uint
	SpokeTensionQCRequired bool
	Status                 string
	RuleVersion            string
	Reason                 string
	CreatedBy              uint
}

func NewProductQualityRequirementService(
	repo *repository.ProductQualityRequirementRepository,
) *ProductQualityRequirementService {
	return &ProductQualityRequirementService{repo: repo}
}

func (s *ProductQualityRequirementService) UpsertSpokeTensionQC(
	input ProductQualityRequirementInput,
) (*productrequirement.ProductQualityRequirementRule, bool, error) {
	if s == nil || s.repo == nil {
		return nil, false, ErrProductQualityRequirementStoreUnavailable
	}
	if input.ProductID == 0 {
		return nil, false, ErrProductQualityRequirementProductNotFound
	}
	if input.VariantID != nil && *input.VariantID == 0 {
		return nil, false, ErrProductQualityRequirementVariantNotFound
	}
	if err := s.repo.EnsureProductScope(input.ProductID, input.VariantID); err != nil {
		if input.VariantID != nil && repository.IsRecordNotFound(err) {
			return nil, false, fmt.Errorf("%w: product %d", ErrProductQualityRequirementVariantNotFound, input.ProductID)
		}
		if repository.IsRecordNotFound(err) {
			return nil, false, fmt.Errorf("%w: product %d", ErrProductQualityRequirementProductNotFound, input.ProductID)
		}
		return nil, false, err
	}

	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = productrequirement.RuleStatusActive
	}
	rule := productrequirement.ProductQualityRequirementRule{
		ProductID:              input.ProductID,
		VariantID:              copyUintPointer(input.VariantID),
		RequirementType:        productrequirement.RequirementTypeSpokeTensionQC,
		SpokeTensionQCRequired: input.SpokeTensionQCRequired,
		Status:                 status,
		RuleVersion:            strings.TrimSpace(input.RuleVersion),
		Reason:                 strings.TrimSpace(input.Reason),
		CreatedBy:              input.CreatedBy,
	}
	rule.Normalize()
	if err := rule.Validate(); err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrProductQualityRequirementInvalid, err)
	}

	existing, err := s.repo.FindByScope(
		input.ProductID,
		input.VariantID,
		productrequirement.RequirementTypeSpokeTensionQC,
	)
	if repository.IsRecordNotFound(err) {
		if err := s.repo.Create(&rule); err != nil {
			return nil, false, err
		}
		return &rule, true, nil
	}
	if err != nil {
		return nil, false, err
	}

	existing.SpokeTensionQCRequired = rule.SpokeTensionQCRequired
	existing.Status = rule.Status
	existing.RuleVersion = rule.RuleVersion
	existing.Reason = rule.Reason
	if err := s.repo.Update(existing); err != nil {
		return nil, false, err
	}
	return existing, false, nil
}

func (s *ProductQualityRequirementService) ResolveSpokeTensionQC(
	productID uint,
	variantID *uint,
) (productrequirement.SpokeTensionQCResolution, error) {
	if s == nil || s.repo == nil {
		return productrequirement.SpokeTensionQCResolution{}, ErrProductQualityRequirementStoreUnavailable
	}
	if productID == 0 {
		return productrequirement.SpokeTensionQCResolution{}, ErrProductQualityRequirementProductNotFound
	}
	if err := s.repo.EnsureProductScope(productID, variantID); err != nil {
		if variantID != nil && repository.IsRecordNotFound(err) {
			return productrequirement.SpokeTensionQCResolution{}, fmt.Errorf("%w: product %d", ErrProductQualityRequirementVariantNotFound, productID)
		}
		if repository.IsRecordNotFound(err) {
			return productrequirement.SpokeTensionQCResolution{}, fmt.Errorf("%w: product %d", ErrProductQualityRequirementProductNotFound, productID)
		}
		return productrequirement.SpokeTensionQCResolution{}, err
	}

	rules, err := s.repo.ListByProduct(productID)
	if err != nil {
		return productrequirement.SpokeTensionQCResolution{}, err
	}
	return productrequirement.ResolveSpokeTensionQC(productID, variantID, rules)
}

func (s *ProductQualityRequirementService) ListSpokeTensionQCRules(
	productID uint,
) ([]productrequirement.ProductQualityRequirementRule, error) {
	if s == nil || s.repo == nil {
		return nil, ErrProductQualityRequirementStoreUnavailable
	}
	if productID == 0 {
		return nil, ErrProductQualityRequirementProductNotFound
	}
	if err := s.repo.EnsureProductScope(productID, nil); repository.IsRecordNotFound(err) {
		return nil, fmt.Errorf("%w: product %d", ErrProductQualityRequirementProductNotFound, productID)
	} else if err != nil {
		return nil, err
	}
	rules, err := s.repo.ListByProduct(productID)
	if err != nil {
		return nil, err
	}
	spokeRules := make([]productrequirement.ProductQualityRequirementRule, 0, len(rules))
	for _, rule := range rules {
		if strings.EqualFold(strings.TrimSpace(rule.RequirementType), productrequirement.RequirementTypeSpokeTensionQC) {
			spokeRules = append(spokeRules, rule)
		}
	}
	return spokeRules, nil
}

func copyUintPointer(value *uint) *uint {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
