package service

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/safehtml"
	"commerce-platform/internal/repository"
	"errors"
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
