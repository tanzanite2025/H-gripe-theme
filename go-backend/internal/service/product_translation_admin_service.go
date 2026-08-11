package service

import (
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/repository"
)

var (
	ErrProductTranslationExists = errors.New("product translation already exists")
)

func (s *ProductService) ListAdminTranslationGroups(products []product.Product) error {
	if len(products) == 0 {
		return nil
	}

	rootIDs := make([]uint, 0, len(products))
	seenRootIDs := make(map[uint]struct{}, len(products))
	for index := range products {
		rootID := productTranslationRootID(&products[index])
		if _, exists := seenRootIDs[rootID]; exists {
			continue
		}
		seenRootIDs[rootID] = struct{}{}
		rootIDs = append(rootIDs, rootID)
	}

	membersByRootID, err := s.productRepo.FindTranslationGroupMembers(rootIDs)
	if err != nil {
		return err
	}

	for index := range products {
		rootID := productTranslationRootID(&products[index])
		products[index].TranslationGroup = buildProductTranslationGroup(
			rootID,
			products[index].ID,
			membersByRootID[rootID],
		)
	}
	return nil
}

func (s *ProductService) GetAdminProductTranslationGroup(id uint) (*product.ProductTranslationGroup, error) {
	source, err := s.findProduct(id)
	if err != nil {
		return nil, err
	}

	rootID := productTranslationRootID(source)
	membersByRootID, err := s.productRepo.FindTranslationGroupMembers([]uint{rootID})
	if err != nil {
		return nil, err
	}

	return buildProductTranslationGroup(rootID, source.ID, membersByRootID[rootID]), nil
}

func (s *ProductService) CopyAdminProductTranslation(id uint, targetLocale string) (*product.Product, *product.ProductTranslationGroup, error) {
	source, err := s.findProduct(id)
	if err != nil {
		return nil, nil, err
	}

	locale, err := requireSupportedLocale(targetLocale)
	if err != nil {
		return nil, nil, err
	}
	sourceLocale, err := requireSupportedLocale(source.Locale)
	if err != nil {
		return nil, nil, err
	}
	if locale == sourceLocale {
		return nil, nil, fmt.Errorf("%w: target locale is already present", ErrProductTranslationInvalid)
	}

	rootID := productTranslationRootID(source)
	groupBeforeCopy, err := s.GetAdminProductTranslationGroup(source.ID)
	if err != nil {
		return nil, nil, err
	}
	if productTranslationGroupHasLocale(groupBeforeCopy, locale) {
		return nil, nil, ErrProductTranslationExists
	}

	slug, err := s.nextAvailableProductSlug(source.Slug, locale)
	if err != nil {
		return nil, nil, err
	}

	variantSKUs := make(map[uint]string, len(source.Variants))
	usedSKUs := make(map[string]struct{}, len(source.Variants))
	for _, sourceVariant := range source.Variants {
		sku, err := s.nextAvailableProductSKU(sourceVariant.SKU, locale, usedSKUs)
		if err != nil {
			return nil, nil, err
		}
		variantSKUs[sourceVariant.ID] = sku
		usedSKUs[strings.ToLower(sku)] = struct{}{}
	}
	targetSKU, err := s.nextAvailableProductSKU(source.SKU, locale, usedSKUs)
	if err != nil {
		return nil, nil, err
	}

	parentID := rootID
	target := &product.Product{
		ProductTypeID:        source.ProductTypeID,
		ShippingTemplateID:   source.ShippingTemplateID,
		AfterSalesTemplateID: copyProductInformationTemplateID(source.AfterSalesTemplate, source.AfterSalesTemplateID, locale),
		PackagingTemplateID:  copyProductInformationTemplateID(source.PackagingTemplate, source.PackagingTemplateID, locale),
		SKU:                  targetSKU,
		Name:                 source.Name,
		Slug:                 slug,
		Description:          source.Description,
		ShortDesc:            source.ShortDesc,
		Currency:             source.Currency,
		Price:                source.Price,
		SalePrice:            source.SalePrice,
		DisplayPriceData:     append([]byte(nil), source.DisplayPriceData...),
		Stock:                source.Stock,
		Status:               source.Status,
		Locale:               locale,
		ParentID:             &parentID,
		Featured:             source.Featured,
	}
	if len(source.Variants) > 0 {
		target.SKU = ""
	}

	if err := s.productRepo.CreateTranslatedCopy(source, target, variantSKUs); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, nil, ErrProductNotFound
		}
		return nil, nil, mapProductRepositoryMutationError(err)
	}

	s.invalidateStorefrontHTMLCache("admin product translation copy")

	createdProduct, err := s.findProduct(target.ID)
	if err != nil {
		return nil, nil, err
	}
	if err := s.enqueueMerchantProductChange(createdProduct, "product_translation_created"); err != nil {
		return nil, nil, requireMerchantEvent(err, createdProduct, "product_translation_created")
	}

	group, err := s.GetAdminProductTranslationGroup(createdProduct.ID)
	if err != nil {
		return nil, nil, err
	}
	return createdProduct, group, nil
}

func productTranslationGroupHasLocale(group *product.ProductTranslationGroup, locale string) bool {
	if group == nil {
		return false
	}
	for _, translation := range group.Translations {
		if translation.Locale == locale {
			return true
		}
	}
	return false
}

func productTranslationRootID(item *product.Product) uint {
	if item == nil || item.ParentID == nil || *item.ParentID == 0 {
		if item == nil {
			return 0
		}
		return item.ID
	}
	return *item.ParentID
}

func buildProductTranslationGroup(rootID, sourceID uint, translations []product.ProductTranslation) *product.ProductTranslationGroup {
	if translations == nil {
		translations = make([]product.ProductTranslation, 0)
	}

	presentLocales := make(map[string]struct{}, len(translations))
	for _, translation := range translations {
		if translation.Locale != "" {
			presentLocales[translation.Locale] = struct{}{}
		}
	}

	missingLocales := make([]string, 0)
	for _, locale := range locales.EnabledLocaleCodes() {
		if _, exists := presentLocales[locale]; !exists {
			missingLocales = append(missingLocales, locale)
		}
	}

	return &product.ProductTranslationGroup{
		RootID:         rootID,
		SourceID:       sourceID,
		Translations:   translations,
		MissingLocales: missingLocales,
	}
}

func copyProductInformationTemplateID(template *product.ProductInformationTemplate, id *uint, targetLocale string) *uint {
	if id == nil || *id == 0 {
		return nil
	}
	if template == nil {
		return id
	}

	templateLocale := strings.TrimSpace(template.Locale)
	if templateLocale == "" || templateLocale == "en" || templateLocale == targetLocale {
		return id
	}
	return nil
}

func (s *ProductService) nextAvailableProductSKU(sourceSKU, locale string, used map[string]struct{}) (string, error) {
	base := strings.TrimSpace(sourceSKU)
	if base == "" {
		base = "product"
	}

	for index := 0; ; index++ {
		suffix := "-" + locale
		if index > 0 {
			suffix = fmt.Sprintf("-%s-%d", locale, index+1)
		}
		candidate := trimProductIdentifier(base+suffix, 100)
		key := strings.ToLower(candidate)
		if _, exists := used[key]; exists {
			continue
		}

		_, err := s.productRepo.FindBySKU(candidate)
		if err != nil && !repository.IsRecordNotFound(err) {
			return "", err
		}
		if err == nil {
			continue
		}

		_, err = s.productRepo.FindVariantBySKU(candidate)
		if err != nil && !repository.IsRecordNotFound(err) {
			return "", err
		}
		if err == nil {
			continue
		}
		return candidate, nil
	}
}

func (s *ProductService) nextAvailableProductSlug(sourceSlug, locale string) (string, error) {
	base := strings.Trim(strings.TrimSpace(sourceSlug), "/")
	if base == "" {
		base = "product"
	}

	for index := 0; ; index++ {
		candidate := base
		if index == 1 {
			candidate = base + "-" + locale
		} else if index > 1 {
			candidate = fmt.Sprintf("%s-%s-%d", base, locale, index)
		}
		candidate = trimProductIdentifier(candidate, 255)

		exists, err := s.productRepo.ProductSlugExists(candidate, locale)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
}

func trimProductIdentifier(value string, maxLength int) string {
	value = strings.TrimSpace(value)
	if len(value) <= maxLength {
		return value
	}
	return strings.TrimSpace(value[:maxLength])
}
