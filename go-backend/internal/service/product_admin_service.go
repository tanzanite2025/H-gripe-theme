package service

import (
	"errors"
	"fmt"
	"strings"
	"tanzanite/internal/domain/currency"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/pkg/safehtml"
	"tanzanite/internal/repository"
)

type ProductMediaInput struct {
	ID                   *uint
	VariantID            *uint
	VariantOptionValueID *uint
	MediaAssetID         *uint
	MediaType            string
	Role                 string
	URL                  string
	ThumbnailURL         string
	PosterURL            string
	Alt                  string
	Title                string
	Locale               string
	SortOrder            int
	IsPrimary            bool
	IsVisible            *bool
}

type ProductVariantOptionValueInput struct {
	ID                 *uint
	SpecDefinitionID   uint
	ValueKey           string
	Label              string
	ColorHex           string
	SwatchMediaAssetID *uint
	SwatchURL          string
	SortOrder          int
	IsEnabled          *bool
}

type ProductCreateInput struct {
	ProductTypeID        *uint
	ShippingTemplateID   *uint
	AfterSalesTemplateID *uint
	PackagingTemplateID  *uint
	Name                 string
	Slug                 string
	Description          string
	ShortDesc            string
	Currency             string
	Status               string
	Locale               string
	ParentID             *uint
	Featured             bool
	MetaTitle            string
	MetaDesc             string
	SpecValues           map[string]string
	Variants             []ProductVariantInput
	VariantOptionValues  []ProductVariantOptionValueInput
	Media                []ProductMediaInput
}

type ProductUpdateInput struct {
	ProductTypeID              *uint
	UpdateProductTypeID        bool
	ShippingTemplateID         *uint
	UpdateShippingTemplateID   bool
	AfterSalesTemplateID       *uint
	UpdateAfterSalesTemplateID bool
	PackagingTemplateID        *uint
	UpdatePackagingTemplateID  bool
	Name                       *string
	Slug                       *string
	Description                *string
	ShortDesc                  *string
	Currency                   *string
	UpdateCurrency             bool
	Status                     *string
	Locale                     *string
	ParentID                   *uint
	UpdateParentID             bool
	Featured                   *bool
	MetaTitle                  *string
	MetaDesc                   *string
	SpecValues                 map[string]string
	UpdateSpecValues           bool
	Variants                   []ProductVariantInput
	UpdateVariants             bool
	VariantOptionValues        []ProductVariantOptionValueInput
	UpdateVariantOptionValues  bool
	Media                      []ProductMediaInput
	UpdateMedia                bool
}

func (s *ProductService) ListAdmin(page, pageSize int, status, locale, search, featured string) ([]product.Product, int64, error) {
	return s.productRepo.FindAllWithFilters(page, pageSize, status, locale, search, featured)
}

func (s *ProductService) GetAdminProduct(id uint) (*product.Product, error) {
	foundProduct, err := s.findProduct(id)
	if err != nil {
		return nil, err
	}
	return sanitizeProductHTML(foundProduct), nil
}

func (s *ProductService) GetStats() (map[string]interface{}, error) {
	return s.productRepo.GetStats()
}

func (s *ProductService) CreateAdminProduct(input ProductCreateInput) (*product.Product, error) {
	locale, err := requireSupportedLocale(input.Locale)
	if err != nil {
		return nil, err
	}
	description, err := safehtml.Sanitize(input.Description)
	if err != nil {
		return nil, err
	}
	shortDesc, err := safehtml.Sanitize(input.ShortDesc)
	if err != nil {
		return nil, err
	}

	specValues, err := s.buildSpecValues(input.ProductTypeID, input.SpecValues)
	if err != nil {
		return nil, err
	}

	priceCurrency, err := s.normalizeAdminProductPriceCurrency(input.Currency)
	if err != nil {
		return nil, err
	}

	optionValues, err := s.buildVariantOptionValues(input.ProductTypeID, input.VariantOptionValues)
	if err != nil {
		return nil, err
	}
	variants, err := s.buildVariants(input.ProductTypeID, input.Variants, priceCurrency, optionValues)
	if err != nil {
		return nil, err
	}
	if err := s.ensureVariantSKUsAvailable(variants, 0); err != nil {
		return nil, err
	}
	mediaItems, err := s.buildProductMedia(input.Media)
	if err != nil {
		return nil, err
	}
	if err := s.validateInformationTemplate(input.AfterSalesTemplateID, product.ProductInformationTemplateKindAfterSales, locale, false); err != nil {
		return nil, err
	}
	if err := s.validateInformationTemplate(input.PackagingTemplateID, product.ProductInformationTemplateKindPackaging, locale, false); err != nil {
		return nil, err
	}

	newProduct := &product.Product{
		ProductTypeID:        input.ProductTypeID,
		ShippingTemplateID:   input.ShippingTemplateID,
		AfterSalesTemplateID: input.AfterSalesTemplateID,
		PackagingTemplateID:  input.PackagingTemplateID,
		SKU:                  defaultVariantSKU(variants),
		Name:                 input.Name,
		Slug:                 input.Slug,
		Description:          description,
		ShortDesc:            shortDesc,
		Currency:             priceCurrency,
		Status:               input.Status,
		Locale:               locale,
		ParentID:             input.ParentID,
		Featured:             input.Featured,
		MetaTitle:            input.MetaTitle,
		MetaDesc:             input.MetaDesc,
	}

	if err := s.productRepo.CreateWithSpecValuesVariantsOptionValuesAndMedia(newProduct, specValues, variants, optionValues, mediaItems); err != nil {
		return nil, mapProductRepositoryMutationError(err)
	}

	s.invalidateStorefrontHTMLCache("admin product create")

	return s.findProduct(newProduct.ID)
}

func (s *ProductService) UpdateAdminProduct(id uint, input ProductUpdateInput) (*product.Product, error) {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return nil, err
	}

	previousProduct := *existingProduct
	previousAfterSalesTemplateID := existingProduct.AfterSalesTemplateID
	previousPackagingTemplateID := existingProduct.PackagingTemplateID

	if input.UpdateProductTypeID {
		existingProduct.ProductTypeID = input.ProductTypeID
	}
	if input.UpdateShippingTemplateID {
		existingProduct.ShippingTemplateID = input.ShippingTemplateID
	}
	if input.UpdateAfterSalesTemplateID {
		existingProduct.AfterSalesTemplateID = input.AfterSalesTemplateID
	}
	if input.UpdatePackagingTemplateID {
		existingProduct.PackagingTemplateID = input.PackagingTemplateID
	}
	if input.Name != nil {
		existingProduct.Name = *input.Name
	}
	if input.Slug != nil {
		existingProduct.Slug = *input.Slug
	}
	if input.Description != nil {
		description, err := safehtml.Sanitize(*input.Description)
		if err != nil {
			return nil, err
		}
		existingProduct.Description = description
	}
	if input.ShortDesc != nil {
		shortDesc, err := safehtml.Sanitize(*input.ShortDesc)
		if err != nil {
			return nil, err
		}
		existingProduct.ShortDesc = shortDesc
	}
	priceCurrency := existingProduct.Currency
	if input.UpdateCurrency {
		if input.Currency == nil {
			priceCurrency = s.primaryPricingCurrency()
		} else {
			priceCurrency, err = s.normalizeAdminProductPriceCurrency(*input.Currency)
			if err != nil {
				return nil, err
			}
		}
		existingProduct.Currency = priceCurrency
	}
	if input.Status != nil {
		existingProduct.Status = *input.Status
	}
	if input.Locale != nil {
		locale, err := requireSupportedLocale(*input.Locale)
		if err != nil {
			return nil, err
		}
		currentLocale, err := requireSupportedLocale(existingProduct.Locale)
		if err != nil {
			return nil, err
		}
		if locale != currentLocale {
			return nil, ErrProductLocaleImmutable
		}
		existingProduct.Locale = locale
	}
	if input.Locale != nil || input.UpdateAfterSalesTemplateID {
		allowDisabled := sameProductInformationTemplateID(existingProduct.AfterSalesTemplateID, previousAfterSalesTemplateID)
		if err := s.validateInformationTemplate(existingProduct.AfterSalesTemplateID, product.ProductInformationTemplateKindAfterSales, existingProduct.Locale, allowDisabled); err != nil {
			return nil, err
		}
	}
	if input.Locale != nil || input.UpdatePackagingTemplateID {
		allowDisabled := sameProductInformationTemplateID(existingProduct.PackagingTemplateID, previousPackagingTemplateID)
		if err := s.validateInformationTemplate(existingProduct.PackagingTemplateID, product.ProductInformationTemplateKindPackaging, existingProduct.Locale, allowDisabled); err != nil {
			return nil, err
		}
	}
	if input.UpdateParentID {
		existingProduct.ParentID = input.ParentID
	}
	if input.Featured != nil {
		existingProduct.Featured = *input.Featured
	}
	if input.MetaTitle != nil {
		existingProduct.MetaTitle = *input.MetaTitle
	}
	if input.MetaDesc != nil {
		existingProduct.MetaDesc = *input.MetaDesc
	}
	var specValues []product.ProductSpecValue
	if input.UpdateSpecValues {
		specValues, err = s.buildSpecValues(existingProduct.ProductTypeID, input.SpecValues)
		if err != nil {
			return nil, err
		}
	}

	var optionValues []product.ProductVariantOptionValue
	effectiveOptionValues := existingProduct.VariantOptionValues
	if input.UpdateVariantOptionValues {
		optionValues, err = s.buildVariantOptionValues(existingProduct.ProductTypeID, input.VariantOptionValues)
		if err != nil {
			return nil, err
		}
		effectiveOptionValues = optionValues
	}

	var variants []product.ProductVariant
	if input.UpdateVariants {
		variants, err = s.buildVariants(existingProduct.ProductTypeID, input.Variants, priceCurrency, effectiveOptionValues)
		if err != nil {
			return nil, err
		}
		if err := s.ensureVariantSKUsAvailable(variants, existingProduct.ID); err != nil {
			return nil, err
		}
	}

	var mediaItems []product.ProductMedia
	if input.UpdateMedia {
		mediaItems, err = s.buildProductMedia(input.Media)
		if err != nil {
			return nil, err
		}
	}

	if err := s.productRepo.UpdateWithSpecValuesVariantsOptionValuesAndMedia(existingProduct, specValues, input.UpdateSpecValues, variants, input.UpdateVariants, optionValues, input.UpdateVariantOptionValues, mediaItems, input.UpdateMedia); err != nil {
		return nil, mapProductRepositoryMutationError(err)
	}

	s.clearProductCache(&previousProduct)
	s.clearProductCache(existingProduct)
	s.invalidateStorefrontHTMLCache("admin product update")

	return s.findProduct(existingProduct.ID)
}

func mapProductRepositoryMutationError(err error) error {
	switch {
	case errors.Is(err, repository.ErrProductMediaReferenceInvalid):
		return fmt.Errorf("%w: %v", ErrProductMediaInvalid, err)
	case errors.Is(err, repository.ErrProductVariantOptionValueReferenceInvalid):
		return fmt.Errorf("%w: %v", ErrProductVariantInvalid, err)
	default:
		return err
	}
}

func sameProductInformationTemplateID(current, previous *uint) bool {
	return current != nil && previous != nil && *current == *previous
}

func (s *ProductService) normalizeAdminProductPriceCurrency(value string) (string, error) {
	primaryCurrency := s.primaryPricingCurrency()
	code := currency.NormalizeCode(value)
	if code == "" {
		code = primaryCurrency
	}
	if !currency.IsValidCode(code) || !currency.IsCatalogCode(code) {
		return "", fmt.Errorf("%w: unsupported product price currency", ErrProductVariantInvalid)
	}
	if code != primaryCurrency {
		return "", fmt.Errorf("%w: product price currency must match primary pricing currency %s", ErrProductVariantInvalid, primaryCurrency)
	}
	return code, nil
}

func (s *ProductService) primaryPricingCurrency() string {
	if s == nil || s.currencyPolicy == nil {
		return product.DefaultPriceCurrency
	}
	value, err := s.currencyPolicy.PrimaryCurrency()
	if err != nil {
		return product.DefaultPriceCurrency
	}
	value = currency.NormalizeCode(value)
	if !currency.IsCatalogCode(value) {
		return product.DefaultPriceCurrency
	}
	return value
}

func (s *ProductService) buildProductMedia(input []ProductMediaInput) ([]product.ProductMedia, error) {
	if len(input) == 0 {
		return nil, nil
	}

	items := make([]product.ProductMedia, 0, len(input))
	hasPrimaryImage := false
	for index, item := range input {
		if item.VariantID != nil && item.VariantOptionValueID != nil {
			return nil, fmt.Errorf("%w: media cannot target both a SKU and an option value", ErrProductMediaInvalid)
		}
		mediaType := strings.ToLower(strings.TrimSpace(item.MediaType))
		if mediaType == "" {
			mediaType = "image"
		}
		if mediaType != "image" && mediaType != "video" {
			return nil, fmt.Errorf("%w: unsupported media type %q", ErrProductMediaInvalid, item.MediaType)
		}

		role, err := normalizeAdminProductMediaRole(mediaType, item.Role, item.IsPrimary)
		if err != nil {
			return nil, err
		}

		url := strings.TrimSpace(item.URL)
		if url == "" {
			return nil, fmt.Errorf("%w: media url is required", ErrProductMediaInvalid)
		}

		isVisible := true
		if item.IsVisible != nil {
			isVisible = *item.IsVisible
		}

		isPrimary := mediaType == "image" && (item.IsPrimary || role == "primary")
		if isPrimary {
			if hasPrimaryImage {
				return nil, fmt.Errorf("%w: only one primary image media is allowed", ErrProductMediaInvalid)
			}
			hasPrimaryImage = true
			role = "primary"
		}

		id := uint(0)
		if item.ID != nil {
			id = *item.ID
		}
		mediaLocale, err := optionalSupportedLocale(item.Locale)
		if err != nil {
			return nil, fmt.Errorf("%w: media %d locale is invalid: %v", ErrProductMediaInvalid, index+1, err)
		}
		items = append(items, product.ProductMedia{
			ID:                   id,
			VariantID:            item.VariantID,
			VariantOptionValueID: item.VariantOptionValueID,
			MediaAssetID:         item.MediaAssetID,
			MediaType:            mediaType,
			Role:                 role,
			URL:                  url,
			ThumbnailURL:         strings.TrimSpace(item.ThumbnailURL),
			PosterURL:            strings.TrimSpace(item.PosterURL),
			Alt:                  strings.TrimSpace(item.Alt),
			Title:                strings.TrimSpace(item.Title),
			Locale:               mediaLocale,
			SortOrder:            item.SortOrder,
			IsPrimary:            isPrimary,
			IsVisible:            isVisible,
		})

		if items[len(items)-1].SortOrder == 0 && index > 0 {
			items[len(items)-1].SortOrder = index * 10
		}
	}

	if !hasPrimaryImage {
		for i := range items {
			if items[i].MediaType == "image" && items[i].IsVisible {
				items[i].IsPrimary = true
				items[i].Role = "primary"
				break
			}
		}
	}

	return items, nil
}

func (s *ProductService) buildVariantOptionValues(productTypeID *uint, input []ProductVariantOptionValueInput) ([]product.ProductVariantOptionValue, error) {
	if len(input) == 0 {
		return nil, nil
	}
	if productTypeID == nil {
		return nil, fmt.Errorf("%w: product_type_id is required when option display values are provided", ErrProductVariantInvalid)
	}

	productType, err := s.productRepo.FindProductTypeByID(*productTypeID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrProductTypeNotFound
		}
		return nil, err
	}

	definitions := make(map[uint]product.SpecDefinition)
	for _, definition := range productType.SpecDefinitions {
		if definition.IsVariantOption {
			definitions[definition.ID] = definition
		}
	}

	values := make([]product.ProductVariantOptionValue, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for index, item := range input {
		definition, ok := definitions[item.SpecDefinitionID]
		if !ok {
			return nil, fmt.Errorf("%w: option display value %d does not belong to a variant option definition", ErrProductVariantInvalid, index+1)
		}

		valueKey := strings.TrimSpace(item.ValueKey)
		if valueKey == "" {
			return nil, fmt.Errorf("%w: option display value %d requires value_key", ErrProductVariantInvalid, index+1)
		}
		uniqueKey := fmt.Sprintf("%d:%s", definition.ID, valueKey)
		if _, exists := seen[uniqueKey]; exists {
			return nil, fmt.Errorf("%w: duplicate option display value %s", ErrProductVariantInvalid, valueKey)
		}
		seen[uniqueKey] = struct{}{}

		label := strings.TrimSpace(item.Label)
		if label == "" {
			label = valueKey
		}
		colorHex := strings.TrimSpace(item.ColorHex)
		if colorHex != "" && !isValidColorHex(colorHex) {
			return nil, fmt.Errorf("%w: invalid color_hex for %s", ErrProductVariantInvalid, valueKey)
		}
		isEnabled := true
		if item.IsEnabled != nil {
			isEnabled = *item.IsEnabled
		}

		optionValue := product.ProductVariantOptionValue{
			SpecDefinitionID:   definition.ID,
			ValueKey:           valueKey,
			Label:              label,
			ColorHex:           colorHex,
			SwatchMediaAssetID: item.SwatchMediaAssetID,
			SwatchURL:          strings.TrimSpace(item.SwatchURL),
			SortOrder:          item.SortOrder,
			IsEnabled:          isEnabled,
		}
		if item.ID != nil {
			optionValue.ID = *item.ID
		}
		values = append(values, optionValue)
	}

	return values, nil
}

func isValidColorHex(value string) bool {
	if len(value) != 4 && len(value) != 7 {
		return false
	}
	if value[0] != '#' {
		return false
	}
	for _, char := range value[1:] {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}

func normalizeAdminProductMediaRole(mediaType, rawRole string, isPrimary bool) (string, error) {
	role := strings.ToLower(strings.TrimSpace(rawRole))

	switch mediaType {
	case "image":
		if role == "" {
			if isPrimary {
				return "primary", nil
			}
			return "gallery", nil
		}
		if role != "primary" && role != "gallery" {
			return "", fmt.Errorf("%w: unsupported image media role %q", ErrProductMediaInvalid, rawRole)
		}
		if isPrimary {
			return "primary", nil
		}
		return role, nil
	case "video":
		if role == "" {
			return "video", nil
		}
		if role != "video" {
			return "", fmt.Errorf("%w: unsupported video media role %q", ErrProductMediaInvalid, rawRole)
		}
		return "video", nil
	default:
		return "", fmt.Errorf("%w: unsupported media type %q", ErrProductMediaInvalid, mediaType)
	}
}

func (s *ProductService) Delete(id uint) error {
	return s.deleteProductByID(id, true)
}

func (s *ProductService) deleteProductByID(id uint, shouldInvalidateHTML bool) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.Delete(id); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)
	if shouldInvalidateHTML {
		s.invalidateStorefrontHTMLCache("admin product delete")
	}

	return nil
}

func (s *ProductService) UpdateStatus(id uint, status string) error {
	return s.updateProductStatusByID(id, status, true)
}

func (s *ProductService) updateProductStatusByID(id uint, status string, shouldInvalidateHTML bool) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.UpdateStatus(id, status); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)
	if shouldInvalidateHTML {
		s.invalidateStorefrontHTMLCache("admin product status update")
	}

	return nil
}

func (s *ProductService) BatchUpdateStatus(ids []uint, status string) (int, error) {
	updated := 0
	for _, id := range ids {
		if err := s.updateProductStatusByID(id, status, false); err != nil {
			if errors.Is(err, ErrProductNotFound) {
				continue
			}
			return updated, err
		}
		updated++
	}
	if updated > 0 {
		s.invalidateStorefrontHTMLCache("admin product batch status update")
	}

	return updated, nil
}

func (s *ProductService) BatchDelete(ids []uint) (int, error) {
	deleted := 0
	for _, id := range ids {
		if err := s.deleteProductByID(id, false); err != nil {
			if errors.Is(err, ErrProductNotFound) {
				continue
			}
			return deleted, err
		}
		deleted++
	}
	if deleted > 0 {
		s.invalidateStorefrontHTMLCache("admin product batch delete")
	}

	return deleted, nil
}
