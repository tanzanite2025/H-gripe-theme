package service

import (
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/safehtml"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

type ProductInformationTemplateInput struct {
	Kind      string
	Name      string
	Slug      string
	Content   string
	Locale    string
	IsEnabled bool
	SortOrder int
}

var (
	ErrProductInformationTemplateNotFound   = errors.New("product information template not found")
	ErrProductInformationTemplateInvalid    = errors.New("product information template invalid")
	ErrProductInformationTemplateSlugExists = errors.New("product information template slug already exists")
)

type ProductInformationTemplateService struct {
	repo               *repository.ProductInformationTemplateRepository
	productCache       ProductDependencyCacheInvalidator
	productCacheEvents ProductCacheEventPublisher
}

func NewProductInformationTemplateService(repo *repository.ProductInformationTemplateRepository) *ProductInformationTemplateService {
	return &ProductInformationTemplateService{repo: repo}
}

func (s *ProductInformationTemplateService) ConfigureProductCacheInvalidator(invalidator ProductDependencyCacheInvalidator) {
	if s == nil {
		return
	}
	s.productCache = invalidator
}

func (s *ProductInformationTemplateService) ConfigureProductCacheEventPublisher(publisher ProductCacheEventPublisher) {
	if s == nil {
		return
	}
	s.productCacheEvents = publisher
}

func (s *ProductInformationTemplateService) List(kind, locale string, includeDisabled bool) ([]product.ProductInformationTemplate, error) {
	return s.repo.List(strings.TrimSpace(kind), strings.TrimSpace(locale), includeDisabled)
}

func (s *ProductInformationTemplateService) Get(id uint) (*product.ProductInformationTemplate, error) {
	template, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProductInformationTemplateNotFound
	}
	return template, err
}

func (s *ProductInformationTemplateService) Create(input ProductInformationTemplateInput) (*product.ProductInformationTemplate, error) {
	template, err := normalizeProductInformationTemplateInput(input)
	if err != nil {
		return nil, err
	}
	exists, err := s.repo.ExistsByKindSlugLocale(template.Kind, template.Slug, template.Locale, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductInformationTemplateSlugExists
	}
	if err := s.repo.Create(template); err != nil {
		return nil, err
	}
	return s.Get(template.ID)
}

func (s *ProductInformationTemplateService) Update(id uint, input ProductInformationTemplateInput) (*product.ProductInformationTemplate, error) {
	template, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeProductInformationTemplateInput(input)
	if err != nil {
		return nil, err
	}
	if normalized.Kind != template.Kind {
		return nil, fmt.Errorf("%w: template kind cannot be changed after creation", ErrProductInformationTemplateInvalid)
	}
	currentLocale, err := requireSupportedLocale(template.Locale)
	if err != nil {
		return nil, err
	}
	if normalized.Locale != currentLocale {
		return nil, fmt.Errorf("%w: template locale cannot be changed after creation", ErrProductInformationTemplateInvalid)
	}
	exists, err := s.repo.ExistsByKindSlugLocale(normalized.Kind, normalized.Slug, normalized.Locale, template.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductInformationTemplateSlugExists
	}
	normalized.ID = template.ID
	normalized.Kind = template.Kind
	normalized.Locale = currentLocale
	if err := s.repo.Update(normalized); err != nil {
		return nil, err
	}
	s.invalidateProductCacheByInformationTemplateID(id)
	if err := s.enqueueProductCacheInvalidationByInformationTemplateID(id, "product information template update"); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *ProductInformationTemplateService) Delete(id uint) error {
	if _, err := s.Get(id); err != nil {
		return err
	}
	s.invalidateProductCacheByInformationTemplateID(id)
	if err := s.enqueueProductCacheInvalidationByInformationTemplateID(id, "product information template delete"); err != nil {
		return err
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	return nil
}

func (s *ProductInformationTemplateService) invalidateProductCacheByInformationTemplateID(templateID uint) {
	if s == nil || s.productCache == nil {
		return
	}
	s.productCache.InvalidateProductCacheByInformationTemplateID(templateID)
}

func (s *ProductInformationTemplateService) enqueueProductCacheInvalidationByInformationTemplateID(templateID uint, reason string) error {
	if s == nil || s.productCacheEvents == nil {
		return nil
	}
	return s.productCacheEvents.EnqueueProductCacheInvalidateByInformationTemplateID(templateID, reason)
}

func normalizeProductInformationTemplateInput(input ProductInformationTemplateInput) (*product.ProductInformationTemplate, error) {
	kind := strings.TrimSpace(input.Kind)
	name := strings.TrimSpace(input.Name)
	slug := strings.TrimSpace(input.Slug)
	locale, err := requireSupportedLocale(input.Locale)
	if err != nil {
		return nil, err
	}
	if !product.IsProductInformationTemplateKind(kind) || name == "" || slug == "" {
		return nil, fmt.Errorf("%w: kind, name and slug are required", ErrProductInformationTemplateInvalid)
	}
	content, err := safehtml.Sanitize(input.Content)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid content: %v", ErrProductInformationTemplateInvalid, err)
	}
	return &product.ProductInformationTemplate{
		Kind:      kind,
		Name:      name,
		Slug:      slug,
		Content:   content,
		Locale:    locale,
		IsEnabled: input.IsEnabled,
		SortOrder: input.SortOrder,
	}, nil
}
