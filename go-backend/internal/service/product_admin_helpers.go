package service

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

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

func normalizeStoredProductPriceCurrency(value string) string {
	code := currency.NormalizeCode(value)
	if !currency.IsValidCode(code) || !currency.IsCatalogCode(code) {
		return product.DefaultPriceCurrency
	}
	return code
}

type productCustomsInfo struct {
	HSCode             string
	CNCode             string
	CountryOfOrigin    string
	CustomsDescription string
}

func (s *ProductService) resolveProductCustomsInfo(profileID *uint, productSpecificationTemplateID *uint, hsCode, cnCode, countryOfOrigin, customsDescription string) (productCustomsInfo, error) {
	if profileID == nil || *profileID == 0 {
		return normalizeProductCustomsInfo(hsCode, cnCode, countryOfOrigin, customsDescription)
	}
	if s.customsClassificationRepo == nil {
		return productCustomsInfo{}, fmt.Errorf("%w: customs classification repository is not configured", ErrProductCustomsProfileInvalid)
	}
	profile, err := s.customsClassificationRepo.FindByID(*profileID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return productCustomsInfo{}, ErrProductCustomsProfileNotFound
	}
	if err != nil {
		return productCustomsInfo{}, err
	}
	if profile.Status != product.CustomsClassificationStatusActive {
		return productCustomsInfo{}, fmt.Errorf("%w: profile is not active", ErrProductCustomsProfileInvalid)
	}
	if profile.ProductSpecificationTemplateID != nil && (productSpecificationTemplateID == nil || *profile.ProductSpecificationTemplateID != *productSpecificationTemplateID) {
		return productCustomsInfo{}, fmt.Errorf("%w: profile does not match product specification template", ErrProductCustomsProfileInvalid)
	}
	return normalizeProductCustomsInfo(
		profile.HSCode,
		profile.CNCode,
		profile.CountryOfOrigin,
		profile.CustomsDescription,
	)
}

func normalizeProductCustomsInfo(hsCode, cnCode, countryOfOrigin, customsDescription string) (productCustomsInfo, error) {
	hsCode = strings.TrimSpace(hsCode)
	cnCode = strings.TrimSpace(cnCode)
	countryOfOrigin = strings.ToUpper(strings.TrimSpace(countryOfOrigin))
	customsDescription = strings.TrimSpace(customsDescription)

	if hsCode != "" && (len(hsCode) != 6 || !isDigitsOnly(hsCode)) {
		return productCustomsInfo{}, fmt.Errorf("%w: HS Code must contain exactly 6 digits", ErrProductCustomsInfoInvalid)
	}
	if cnCode != "" && (len(cnCode) != 8 || !isDigitsOnly(cnCode)) {
		return productCustomsInfo{}, fmt.Errorf("%w: CN Code must contain exactly 8 digits", ErrProductCustomsInfoInvalid)
	}
	if countryOfOrigin != "" && (len(countryOfOrigin) != 2 || !isUppercaseLettersOnly(countryOfOrigin)) {
		return productCustomsInfo{}, fmt.Errorf("%w: country of origin must be a 2-letter ISO code", ErrProductCustomsInfoInvalid)
	}
	if len(customsDescription) > 255 {
		return productCustomsInfo{}, fmt.Errorf("%w: customs description is too long", ErrProductCustomsInfoInvalid)
	}

	return productCustomsInfo{
		HSCode:             hsCode,
		CNCode:             cnCode,
		CountryOfOrigin:    countryOfOrigin,
		CustomsDescription: customsDescription,
	}, nil
}

func isDigitsOnly(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return value != ""
}

func isUppercaseLettersOnly(value string) bool {
	for _, char := range value {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return value != ""
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
		thumbnailURL := strings.TrimSpace(item.ThumbnailURL)
		imageVariants := map[string]product.ProductMediaImageVariant{}
		if mediaType == "image" && item.MediaAssetID != nil && *item.MediaAssetID > 0 {
			if resolver, ok := s.mediaService.(productMediaImageVariantResolver); ok {
				variants, derivativeThumbnail, err := resolver.ProductMediaImageVariants(*item.MediaAssetID)
				if err != nil {
					return nil, fmt.Errorf("%w: resolve media derivatives: %v", ErrProductMediaInvalid, err)
				}
				imageVariants = variants
				if thumbnailURL == "" {
					thumbnailURL = derivativeThumbnail
				}
			}
		}

		items = append(items, product.ProductMedia{
			ID:                   id,
			VariantID:            item.VariantID,
			VariantOptionValueID: item.VariantOptionValueID,
			MediaAssetID:         item.MediaAssetID,
			MediaType:            mediaType,
			Role:                 role,
			URL:                  url,
			ThumbnailURL:         thumbnailURL,
			PosterURL:            strings.TrimSpace(item.PosterURL),
			ImageVariantData:     product.ProductMediaImageVariantsJSON(imageVariants),
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

func (s *ProductService) buildVariantOptionValues(productSpecificationTemplateID *uint, input []ProductVariantOptionValueInput) ([]product.ProductVariantOptionValue, error) {
	if len(input) == 0 {
		return nil, nil
	}
	if productSpecificationTemplateID == nil {
		return nil, fmt.Errorf("%w: product_specification_template_id is required when option display values are provided", ErrProductVariantInvalid)
	}

	productSpecificationTemplate, err := s.productRepo.FindProductSpecificationTemplateByID(*productSpecificationTemplateID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrProductSpecificationTemplateNotFound
		}
		return nil, err
	}

	definitions := make(map[uint]product.SpecDefinition)
	for _, definition := range productSpecificationTemplate.SpecDefinitions {
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
