package service

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/safehtml"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
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
	ProductSpecificationTemplateID *uint
	ProductCategoryID              *uint
	BrandID                        *uint
	ShippingTemplateID             *uint
	AfterSalesTemplateID           *uint
	PackagingTemplateID            *uint
	CustomsClassificationProfileID *uint
	HSCode                         string
	CNCode                         string
	CountryOfOrigin                string
	CustomsDescription             string
	Name                           string
	Slug                           string
	Description                    string
	ShortDesc                      string
	Currency                       string
	Status                         string
	Locale                         string
	ParentID                       *uint
	Featured                       bool
	SpecValues                     map[string]string
	Variants                       []ProductVariantInput
	VariantOptionValues            []ProductVariantOptionValueInput
	Media                          []ProductMediaInput
}

type ProductUpdateInput struct {
	ProductSpecificationTemplateID       *uint
	UpdateProductSpecificationTemplateID bool
	ProductCategoryID                    *uint
	UpdateProductCategoryID              bool
	BrandID                              *uint
	UpdateBrandID                        bool
	ShippingTemplateID                   *uint
	UpdateShippingTemplateID             bool
	AfterSalesTemplateID                 *uint
	UpdateAfterSalesTemplateID           bool
	PackagingTemplateID                  *uint
	UpdatePackagingTemplateID            bool
	CustomsClassificationProfileID       *uint
	UpdateCustomsClassificationProfileID bool
	HSCode                               *string
	UpdateHSCode                         bool
	CNCode                               *string
	UpdateCNCode                         bool
	CountryOfOrigin                      *string
	UpdateCountryOfOrigin                bool
	CustomsDescription                   *string
	UpdateCustomsDescription             bool
	Name                                 *string
	Slug                                 *string
	Description                          *string
	ShortDesc                            *string
	Currency                             *string
	UpdateCurrency                       bool
	Status                               *string
	Locale                               *string
	ParentID                             *uint
	UpdateParentID                       bool
	Featured                             *bool
	SpecValues                           map[string]string
	UpdateSpecValues                     bool
	Variants                             []ProductVariantInput
	UpdateVariants                       bool
	VariantOptionValues                  []ProductVariantOptionValueInput
	UpdateVariantOptionValues            bool
	Media                                []ProductMediaInput
	UpdateMedia                          bool
}

func (s *ProductService) ListAdmin(page, pageSize int, status, locale, search, featured, customsStatus, productSpecificationTemplateID string) ([]product.Product, int64, error) {
	products, total, err := s.productRepo.FindAllWithFilters(page, pageSize, status, locale, search, featured, customsStatus, productSpecificationTemplateID)
	if err != nil {
		return nil, 0, err
	}
	if err := s.ListAdminTranslationGroups(products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
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

func (s *ProductService) GetCustomsSummary(locale string) (repository.ProductCustomsSummary, error) {
	return s.productRepo.GetCustomsSummary(locale)
}

func (s *ProductService) ListFilterableSpecificationsWithDynamicValuesForCategory(categorySlug string) ([]repository.ProductCategoryFilterableSpecification, error) {
	if s == nil || s.productRepo == nil {
		return []repository.ProductCategoryFilterableSpecification{}, errors.New("product repository is not configured")
	}
	return s.productRepo.ListFilterableSpecificationsWithDynamicValuesForCategory(categorySlug)
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

	specValues, err := s.buildSpecValues(input.ProductSpecificationTemplateID, input.SpecValues)
	if err != nil {
		return nil, err
	}

	priceCurrency, err := s.normalizeAdminProductPriceCurrency(input.Currency)
	if err != nil {
		return nil, err
	}

	optionValues, err := s.buildVariantOptionValues(input.ProductSpecificationTemplateID, input.VariantOptionValues)
	if err != nil {
		return nil, err
	}
	variants, err := s.buildVariants(input.ProductSpecificationTemplateID, input.Variants, priceCurrency, optionValues)
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
	if err := s.validateProductTranslationParent(input.ParentID, 0, locale); err != nil {
		return nil, err
	}
	if err := s.validateProductBrand(input.BrandID, false); err != nil {
		return nil, err
	}
	if err := s.validateProductCategory(input.ProductCategoryID, false); err != nil {
		return nil, err
	}
	if err := s.validateInformationTemplate(input.AfterSalesTemplateID, product.ProductInformationTemplateKindAfterSales, locale, false); err != nil {
		return nil, err
	}
	if err := s.validateInformationTemplate(input.PackagingTemplateID, product.ProductInformationTemplateKindPackaging, locale, false); err != nil {
		return nil, err
	}
	customsInfo, err := s.resolveProductCustomsInfo(
		input.CustomsClassificationProfileID,
		input.ProductSpecificationTemplateID,
		input.HSCode,
		input.CNCode,
		input.CountryOfOrigin,
		input.CustomsDescription,
	)
	if err != nil {
		return nil, err
	}

	newProduct := &product.Product{
		ProductSpecificationTemplateID: input.ProductSpecificationTemplateID,
		ProductCategoryID:              input.ProductCategoryID,
		BrandID:                        input.BrandID,
		ShippingTemplateID:             input.ShippingTemplateID,
		AfterSalesTemplateID:           input.AfterSalesTemplateID,
		PackagingTemplateID:            input.PackagingTemplateID,
		CustomsClassificationProfileID: input.CustomsClassificationProfileID,
		HSCode:                         customsInfo.HSCode,
		CNCode:                         customsInfo.CNCode,
		CountryOfOrigin:                customsInfo.CountryOfOrigin,
		CustomsDescription:             customsInfo.CustomsDescription,
		SKU:                            defaultVariantSKU(variants),
		Name:                           input.Name,
		Slug:                           input.Slug,
		Description:                    description,
		ShortDesc:                      shortDesc,
		Currency:                       priceCurrency,
		Status:                         input.Status,
		Locale:                         locale,
		ParentID:                       input.ParentID,
		Featured:                       input.Featured,
	}

	if s.txManager != nil {
		var createdProduct *product.Product
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.Product == nil {
				return errors.New("transactional product repository is not configured")
			}
			if err := tx.Product.CreateWithSpecValuesVariantsOptionValuesAndMedia(newProduct, specValues, variants, optionValues, mediaItems); err != nil {
				return mapProductRepositoryMutationError(err)
			}
			var err error
			createdProduct, err = tx.Product.FindByID(newProduct.ID)
			if err != nil {
				return err
			}
			_, merchantEvents, err := s.transactionalProductPublishers(tx.Outbox)
			if err != nil {
				return err
			}
			if err := enqueueMerchantProductChangeWithPublisher(merchantEvents, createdProduct, "product_created"); err != nil {
				return requireMerchantEvent(err, createdProduct, "product_created")
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		s.invalidateStorefrontHTMLCache("admin product create")
		return createdProduct, nil
	}

	if err := s.productRepo.CreateWithSpecValuesVariantsOptionValuesAndMedia(newProduct, specValues, variants, optionValues, mediaItems); err != nil {
		return nil, mapProductRepositoryMutationError(err)
	}

	s.invalidateStorefrontHTMLCache("admin product create")

	createdProduct, err := s.findProduct(newProduct.ID)
	if err != nil {
		return nil, err
	}
	if err := s.enqueueMerchantProductChange(createdProduct, "product_created"); err != nil {
		return nil, requireMerchantEvent(err, createdProduct, "product_created")
	}
	return createdProduct, nil
}

func (s *ProductService) UpdateAdminProduct(id uint, input ProductUpdateInput) (*product.Product, error) {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return nil, err
	}

	previousProduct := *existingProduct
	previousAfterSalesTemplateID := existingProduct.AfterSalesTemplateID
	previousPackagingTemplateID := existingProduct.PackagingTemplateID

	if input.UpdateProductSpecificationTemplateID {
		existingProduct.ProductSpecificationTemplateID = input.ProductSpecificationTemplateID
		existingProduct.CustomsClassificationProfileID = nil
		existingProduct.CustomsClassificationProfile = nil
	}
	if input.UpdateProductCategoryID {
		allowDisabled := existingProduct.ProductCategoryID != nil &&
			input.ProductCategoryID != nil &&
			*existingProduct.ProductCategoryID == *input.ProductCategoryID
		if err := s.validateProductCategory(input.ProductCategoryID, allowDisabled); err != nil {
			return nil, err
		}
		existingProduct.ProductCategoryID = input.ProductCategoryID
		existingProduct.ProductCategory = nil
	}
	if input.UpdateBrandID {
		allowDisabled := existingProduct.BrandID != nil && input.BrandID != nil && *existingProduct.BrandID == *input.BrandID
		if err := s.validateProductBrand(input.BrandID, allowDisabled); err != nil {
			return nil, err
		}
		existingProduct.BrandID = input.BrandID
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
	if input.UpdateCustomsClassificationProfileID {
		customsInfo, err := s.resolveProductCustomsInfo(
			input.CustomsClassificationProfileID,
			existingProduct.ProductSpecificationTemplateID,
			existingProduct.HSCode,
			existingProduct.CNCode,
			existingProduct.CountryOfOrigin,
			existingProduct.CustomsDescription,
		)
		if err != nil {
			return nil, err
		}
		existingProduct.CustomsClassificationProfileID = input.CustomsClassificationProfileID
		existingProduct.HSCode = customsInfo.HSCode
		existingProduct.CNCode = customsInfo.CNCode
		existingProduct.CountryOfOrigin = customsInfo.CountryOfOrigin
		existingProduct.CustomsDescription = customsInfo.CustomsDescription
		// The existing product may have its previous profile preloaded. Keep the
		// foreign key update explicit and prevent Save from persisting that stale
		// belongs-to association.
		existingProduct.CustomsClassificationProfile = nil
	}
	if input.UpdateHSCode || input.UpdateCNCode || input.UpdateCountryOfOrigin || input.UpdateCustomsDescription {
		hsCode := existingProduct.HSCode
		cnCode := existingProduct.CNCode
		countryOfOrigin := existingProduct.CountryOfOrigin
		customsDescription := existingProduct.CustomsDescription
		if input.UpdateHSCode && input.HSCode != nil {
			hsCode = *input.HSCode
		}
		if input.UpdateCNCode && input.CNCode != nil {
			cnCode = *input.CNCode
		}
		if input.UpdateCountryOfOrigin && input.CountryOfOrigin != nil {
			countryOfOrigin = *input.CountryOfOrigin
		}
		if input.UpdateCustomsDescription && input.CustomsDescription != nil {
			customsDescription = *input.CustomsDescription
		}
		customsInfo, err := normalizeProductCustomsInfo(hsCode, cnCode, countryOfOrigin, customsDescription)
		if err != nil {
			return nil, err
		}
		existingProduct.HSCode = customsInfo.HSCode
		existingProduct.CNCode = customsInfo.CNCode
		existingProduct.CountryOfOrigin = customsInfo.CountryOfOrigin
		existingProduct.CustomsDescription = customsInfo.CustomsDescription
		if !input.UpdateCustomsClassificationProfileID {
			existingProduct.CustomsClassificationProfileID = nil
			existingProduct.CustomsClassificationProfile = nil
		}
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
	priceCurrency := normalizeStoredProductPriceCurrency(existingProduct.Currency)
	if input.UpdateCurrency {
		requestedCurrency := priceCurrency
		if input.Currency != nil {
			requestedCurrency = currency.NormalizeCode(*input.Currency)
		}
		if requestedCurrency == "" {
			requestedCurrency = priceCurrency
		}
		if requestedCurrency != priceCurrency {
			priceCurrency, err = s.normalizeAdminProductPriceCurrency(requestedCurrency)
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
		if err := s.validateProductTranslationParent(input.ParentID, existingProduct.ID, existingProduct.Locale); err != nil {
			return nil, err
		}
		existingProduct.ParentID = input.ParentID
	}
	if input.Featured != nil {
		existingProduct.Featured = *input.Featured
	}
	var specValues []product.ProductSpecValue
	if input.UpdateSpecValues {
		specValues, err = s.buildSpecValues(existingProduct.ProductSpecificationTemplateID, input.SpecValues)
		if err != nil {
			return nil, err
		}
	}

	var optionValues []product.ProductVariantOptionValue
	effectiveOptionValues := existingProduct.VariantOptionValues
	if input.UpdateVariantOptionValues {
		optionValues, err = s.buildVariantOptionValues(existingProduct.ProductSpecificationTemplateID, input.VariantOptionValues)
		if err != nil {
			return nil, err
		}
		effectiveOptionValues = optionValues
	}

	var variants []product.ProductVariant
	if input.UpdateVariants {
		variants, err = s.buildVariants(existingProduct.ProductSpecificationTemplateID, input.Variants, priceCurrency, effectiveOptionValues)
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

	if s.txManager != nil {
		var updatedProduct *product.Product
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.Product == nil {
				return errors.New("transactional product repository is not configured")
			}
			if err := tx.Product.UpdateWithSpecValuesVariantsOptionValuesAndMedia(existingProduct, specValues, input.UpdateSpecValues, variants, input.UpdateVariants, optionValues, input.UpdateVariantOptionValues, mediaItems, input.UpdateMedia); err != nil {
				return mapProductRepositoryMutationError(err)
			}
			var err error
			updatedProduct, err = tx.Product.FindByID(existingProduct.ID)
			if err != nil {
				return err
			}
			cacheEvents, merchantEvents, err := s.transactionalProductPublishers(tx.Outbox)
			if err != nil {
				return err
			}
			if err := enqueueProductCacheInvalidationByIDsWithPublisher(cacheEvents, []uint{previousProduct.ID, updatedProduct.ID}, "admin product update"); err != nil {
				return err
			}
			if merchantProductUpdateAffectsChannel(input) {
				reason := merchantProductSourceChangeReason(input)
				if err := enqueueMerchantProductChangeWithPublisher(merchantEvents, updatedProduct, reason); err != nil {
					return requireMerchantEvent(err, updatedProduct, reason)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		s.clearProductCache(&previousProduct)
		s.clearProductCache(updatedProduct)
		s.invalidateStorefrontHTMLCache("admin product update")
		return updatedProduct, nil
	}

	if err := s.productRepo.UpdateWithSpecValuesVariantsOptionValuesAndMedia(existingProduct, specValues, input.UpdateSpecValues, variants, input.UpdateVariants, optionValues, input.UpdateVariantOptionValues, mediaItems, input.UpdateMedia); err != nil {
		return nil, mapProductRepositoryMutationError(err)
	}

	s.clearProductCache(&previousProduct)
	s.clearProductCache(existingProduct)
	if err := s.enqueueProductCacheInvalidationByIDs([]uint{previousProduct.ID, existingProduct.ID}, "admin product update"); err != nil {
		return nil, err
	}
	s.invalidateStorefrontHTMLCache("admin product update")

	updatedProduct, err := s.findProduct(existingProduct.ID)
	if err != nil {
		return nil, err
	}
	if merchantProductUpdateAffectsChannel(input) {
		reason := merchantProductSourceChangeReason(input)
		if err := s.enqueueMerchantProductChange(updatedProduct, reason); err != nil {
			return nil, requireMerchantEvent(err, updatedProduct, reason)
		}
	}
	return updatedProduct, nil
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

func (s *ProductService) Delete(id uint) error {
	return s.deleteProductByID(id, true)
}

func (s *ProductService) deleteProductByID(id uint, shouldInvalidateHTML bool) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if s.txManager != nil {
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.Product == nil {
				return errors.New("transactional product repository is not configured")
			}
			if err := tx.Product.Delete(id); err != nil {
				return err
			}
			cacheEvents, merchantEvents, err := s.transactionalProductPublishers(tx.Outbox)
			if err != nil {
				return err
			}
			if err := enqueueProductCacheInvalidationByIDsWithPublisher(cacheEvents, []uint{existingProduct.ID}, "admin product delete"); err != nil {
				return err
			}
			if err := enqueueMerchantProductWithdrawWithPublisher(merchantEvents, existingProduct, "product_deleted"); err != nil {
				return requireMerchantEvent(err, existingProduct, "product_deleted")
			}
			return nil
		})
		if err != nil {
			return err
		}
		s.clearProductCache(existingProduct)
		if shouldInvalidateHTML {
			s.invalidateStorefrontHTMLCache("admin product delete")
		}
		return nil
	}

	if err := s.productRepo.Delete(id); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)
	if err := s.enqueueProductCacheInvalidationByIDs([]uint{existingProduct.ID}, "admin product delete"); err != nil {
		return err
	}
	if shouldInvalidateHTML {
		s.invalidateStorefrontHTMLCache("admin product delete")
	}

	if err := s.enqueueMerchantProductWithdraw(existingProduct, "product_deleted"); err != nil {
		return requireMerchantEvent(err, existingProduct, "product_deleted")
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

	if s.txManager != nil {
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.Product == nil {
				return errors.New("transactional product repository is not configured")
			}
			if err := tx.Product.UpdateStatus(id, status); err != nil {
				return err
			}
			updatedProduct, err := tx.Product.FindByID(id)
			if err != nil {
				return err
			}
			cacheEvents, merchantEvents, err := s.transactionalProductPublishers(tx.Outbox)
			if err != nil {
				return err
			}
			if err := enqueueProductCacheInvalidationByIDsWithPublisher(cacheEvents, []uint{id}, "admin product status update"); err != nil {
				return err
			}
			if err := enqueueMerchantProductChangeWithPublisher(merchantEvents, updatedProduct, "product_status_changed"); err != nil {
				return requireMerchantEvent(err, updatedProduct, "product_status_changed")
			}
			return nil
		})
		if err != nil {
			return err
		}
		s.clearProductCache(existingProduct)
		if shouldInvalidateHTML {
			s.invalidateStorefrontHTMLCache("admin product status update")
		}
		return nil
	}

	if err := s.productRepo.UpdateStatus(id, status); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)
	if err := s.enqueueProductCacheInvalidationByIDs([]uint{existingProduct.ID}, "admin product status update"); err != nil {
		return err
	}
	if shouldInvalidateHTML {
		s.invalidateStorefrontHTMLCache("admin product status update")
	}

	existingProduct.Status = status
	if err := s.enqueueMerchantProductChange(existingProduct, "product_status_changed"); err != nil {
		return requireMerchantEvent(err, existingProduct, "product_status_changed")
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
