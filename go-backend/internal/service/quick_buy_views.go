package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/quickbuy"
	"commerce-platform/internal/pkg/locales"

	"gorm.io/datatypes"
)

func quickBuySessionTotals(items []quickbuy.SessionItem) (float64, int) {
	var subtotal float64
	var weightG int
	for _, item := range items {
		quantity := item.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		subtotal += item.UnitPriceSnapshot * float64(quantity)
		weightG += item.WeightSnapshotG * quantity
	}
	return subtotal, weightG
}

func quickBuySessionView(session quickbuy.Session, validation *QuickBuySessionValidationResult, resolvers ...PublicMediaURLResolver) *QuickBuySessionView {
	resolver := quickBuyMediaResolver(resolvers)
	items := make([]QuickBuySessionItemView, 0, len(session.Items))
	for _, item := range session.Items {
		items = append(items, QuickBuySessionItemView{
			ID:                item.ID,
			StepKey:           item.StepKey,
			ProductID:         item.ProductID,
			VariantID:         item.VariantID,
			Quantity:          item.Quantity,
			UnitPriceSnapshot: item.UnitPriceSnapshot,
			CurrencySnapshot:  item.CurrencySnapshot,
			WeightSnapshotG:   item.WeightSnapshotG,
			ProductSnapshot:   quickBuyPublicProductSnapshot(item.ProductSnapshot, resolver),
			VariantSnapshot:   item.VariantSnapshot,
			SortOrder:         item.SortOrder,
		})
	}

	var flow *QuickBuyPublicFlowView
	if session.Version != nil && session.Version.Flow != nil {
		flow = quickBuyPublicFlowView(*session.Version, session.Locale, resolver)
	}
	return &QuickBuySessionView{
		SessionToken:     session.SessionToken,
		FlowID:           session.FlowID,
		FlowVersionID:    session.FlowVersionID,
		Locale:           session.Locale,
		MarketCountry:    session.MarketCountry,
		Currency:         session.Currency,
		Status:           session.Status,
		ValidationStatus: session.ValidationStatus,
		SubtotalSnapshot: session.SubtotalSnapshot,
		WeightSnapshotG:  session.WeightSnapshotG,
		ExpiresAt:        session.ExpiresAt,
		Flow:             flow,
		Items:            items,
		Validation:       validation,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
	}
}

func quickBuyPublicProductSnapshot(raw datatypes.JSON, resolver PublicMediaURLResolver) datatypes.JSON {
	if resolver == nil || len(raw) == 0 {
		return raw
	}

	var snapshot map[string]interface{}
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return raw
	}
	for _, key := range []string{"thumbnail", "featured_image"} {
		if value, ok := snapshot[key].(string); ok {
			snapshot[key] = canonicalPublicMediaURL(resolver, value)
		}
	}
	return quickBuyJSON(snapshot)
}

func quickBuyProductSnapshot(item productdomain.Product, resolvers ...PublicMediaURLResolver) datatypes.JSON {
	thumbnail := quickBuyProductThumbnail(item, resolvers...)
	return quickBuyJSON(map[string]interface{}{
		"id":                                item.ID,
		"product_specification_template_id": item.ProductSpecificationTemplateID,
		"sku":                               item.SKU,
		"name":                              item.Name,
		"slug":                              item.Slug,
		"thumbnail":                         thumbnail,
		"featured_image":                    thumbnail,
		"currency":                          item.Currency,
		"price":                             item.Price,
		"sale_price":                        item.SalePrice,
		"status":                            item.Status,
	})
}

func quickBuyProductThumbnail(item productdomain.Product, resolvers ...PublicMediaURLResolver) string {
	resolver := quickBuyMediaResolver(resolvers)
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible && media.IsPrimary {
			if strings.TrimSpace(media.ThumbnailURL) != "" {
				return canonicalPublicMediaURL(resolver, media.ThumbnailURL)
			}
			return canonicalPublicMediaURL(resolver, media.URL)
		}
	}
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible {
			if strings.TrimSpace(media.ThumbnailURL) != "" {
				return canonicalPublicMediaURL(resolver, media.ThumbnailURL)
			}
			return canonicalPublicMediaURL(resolver, media.URL)
		}
	}
	return ""
}

func quickBuyVariantSnapshot(item productdomain.ProductVariant) datatypes.JSON {
	return quickBuyJSON(map[string]interface{}{
		"id":            item.ID,
		"product_id":    item.ProductID,
		"sku":           item.SKU,
		"title":         item.Title,
		"currency":      item.Currency,
		"price":         item.Price,
		"sale_price":    item.SalePrice,
		"stock":         item.Stock,
		"weight_grams":  item.Weight,
		"is_default":    item.IsDefault,
		"option_values": item.OptionValues,
	})
}

func quickBuyJSON(value interface{}) datatypes.JSON {
	encoded, err := json.Marshal(value)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(encoded)
}

func quickBuyFlowSummary(flow quickbuy.Flow) QuickBuyFlowSummary {
	versions := make([]QuickBuyVersionSummary, 0, len(flow.Versions))
	for _, version := range flow.Versions {
		versions = append(versions, QuickBuyVersionSummary{
			ID:            version.ID,
			FlowID:        version.FlowID,
			VersionNumber: version.VersionNumber,
			Status:        version.Status,
			PublishedAt:   version.PublishedAt,
			StartsAt:      version.StartsAt,
			EndsAt:        version.EndsAt,
		})
	}
	return QuickBuyFlowSummary{
		ID:           flow.ID,
		Slug:         flow.Slug,
		Name:         flow.Name,
		Description:  flow.Description,
		HelpText:     flow.HelpText,
		EntrySurface: flow.EntrySurface,
		IsEnabled:    flow.IsEnabled,
		SortOrder:    flow.SortOrder,
		Versions:     versions,
		CreatedAt:    flow.CreatedAt,
		UpdatedAt:    flow.UpdatedAt,
	}
}

func quickBuyFlowView(version quickbuy.Version, locale string) *QuickBuyFlowView {
	if version.Flow == nil {
		return nil
	}
	requestLocale := locales.ResolveSupported(locale)
	steps := quickBuyStepViews(version.Steps, requestLocale, !isDefaultQuickBuyFlow(version))
	return &QuickBuyFlowView{
		ID:           version.Flow.ID,
		Slug:         version.Flow.Slug,
		Name:         version.Flow.Name,
		Description:  version.Flow.Description,
		HelpText:     quickBuyFlowHelpText(*version.Flow, requestLocale),
		Translations: quickBuyFlowTranslationViews(version.Flow.Translations),
		EntrySurface: version.Flow.EntrySurface,
		IsEnabled:    version.Flow.IsEnabled,
		SortOrder:    version.Flow.SortOrder,
		Version: QuickBuyVersionView{
			ID:            version.ID,
			VersionNumber: version.VersionNumber,
			Status:        version.Status,
			PublishedAt:   version.PublishedAt,
			StartsAt:      version.StartsAt,
			EndsAt:        version.EndsAt,
		},
		Steps: steps,
	}
}

func quickBuyPublicFlowView(version quickbuy.Version, locale string, resolvers ...PublicMediaURLResolver) *QuickBuyPublicFlowView {
	if version.Flow == nil {
		return nil
	}
	resolver := quickBuyMediaResolver(resolvers)
	requestLocale := locales.ResolveSupported(locale)
	return &QuickBuyPublicFlowView{
		ID:           version.Flow.ID,
		Slug:         version.Flow.Slug,
		Name:         version.Flow.Name,
		Description:  version.Flow.Description,
		HelpText:     quickBuyFlowHelpText(*version.Flow, requestLocale),
		EntrySurface: version.Flow.EntrySurface,
		IsEnabled:    version.Flow.IsEnabled,
		Version: QuickBuyVersionView{
			ID:            version.ID,
			VersionNumber: version.VersionNumber,
			Status:        version.Status,
			PublishedAt:   version.PublishedAt,
			StartsAt:      version.StartsAt,
			EndsAt:        version.EndsAt,
		},
		Steps: quickBuyStepViews(version.Steps, requestLocale, !isDefaultQuickBuyFlow(version), resolver),
	}
}

func quickBuyStepViews(steps []quickbuy.Step, locale string, exposeProductSpecificationTemplates bool, resolvers ...PublicMediaURLResolver) []QuickBuyStepView {
	resolver := quickBuyMediaResolver(resolvers)
	result := make([]QuickBuyStepView, 0, len(steps))
	for _, step := range steps {
		result = append(result, quickBuyStepView(step, locale, exposeProductSpecificationTemplates, resolver))
	}
	return result
}

func quickBuyFlowHelpText(flow quickbuy.Flow, locale string) string {
	helpText := strings.TrimSpace(flow.HelpText)
	if translation := quickBuyFlowTranslationForLocale(flow.Translations, locale); translation != nil && strings.TrimSpace(translation.HelpText) != "" {
		helpText = strings.TrimSpace(translation.HelpText)
	}
	return helpText
}

func quickBuyFlowTranslationForLocale(translations []quickbuy.FlowTranslation, locale string) *quickbuy.FlowTranslation {
	requestedLocale := locales.ResolveSupported(locale)
	if requestedLocale == "" {
		return nil
	}
	for index := range translations {
		if translations[index].Locale == requestedLocale {
			return &translations[index]
		}
	}
	return nil
}

func quickBuyFlowTranslationViews(translations []quickbuy.FlowTranslation) []QuickBuyFlowTranslationView {
	if len(translations) == 0 {
		return []QuickBuyFlowTranslationView{}
	}
	result := make([]QuickBuyFlowTranslationView, 0, len(translations))
	for _, translation := range translations {
		result = append(result, QuickBuyFlowTranslationView{
			ID:       translation.ID,
			Locale:   translation.Locale,
			HelpText: translation.HelpText,
		})
	}
	return result
}

func quickBuyStepView(step quickbuy.Step, locale string, exposeProductSpecificationTemplates bool, resolvers ...PublicMediaURLResolver) QuickBuyStepView {
	resolver := quickBuyMediaResolver(resolvers)
	productCategories := make([]QuickBuyProductCategoryView, 0, len(step.ProductCategories))
	productSpecificationTemplates := make([]QuickBuyProductSpecificationTemplateView, 0, len(step.ProductSpecificationTemplates))
	stepSlug := step.StepKey
	for index, item := range step.ProductCategories {
		if item.ProductCategory == nil {
			continue
		}
		if stepSlug == step.StepKey && index == 0 && !isDefaultQuickBuyStepKey(step.StepKey) {
			stepSlug = item.ProductCategory.Slug
		}
		productCategories = append(productCategories, quickBuyProductCategoryView(*item.ProductCategory, locale, item.IsPrimary, resolver))
	}
	if exposeProductSpecificationTemplates {
		for index, item := range step.ProductSpecificationTemplates {
			if item.ProductSpecificationTemplate == nil {
				continue
			}
			if index == 0 {
				stepSlug = item.ProductSpecificationTemplate.Slug
			}
			productSpecificationTemplates = append(productSpecificationTemplates, quickBuyProductSpecificationTemplateView(*item.ProductSpecificationTemplate, item.IsPrimary))
		}
	}
	return QuickBuyStepView{
		ID:                            step.ID,
		StepKey:                       step.StepKey,
		Slug:                          stepSlug,
		Name:                          step.Name,
		SortOrder:                     step.SortOrder,
		ProductCategories:             productCategories,
		ProductSpecificationTemplates: productSpecificationTemplates,
		Filters:                       quickBuyStepFiltersForScope(step, nil, exposeProductSpecificationTemplates),
	}
}

func quickBuyStepFiltersForScope(step quickbuy.Step, valuesBySlug map[string][]string, exposeProductSpecificationTemplates bool) []QuickBuySpecFilterView {
	if !exposeProductSpecificationTemplates {
		return []QuickBuySpecFilterView{}
	}
	return quickBuyStepFilters(step, valuesBySlug)
}

func quickBuyFilterableSpecDefinitions(step quickbuy.Step) []productdomain.SpecDefinition {
	definitionsBySlug := make(map[string]productdomain.SpecDefinition)
	for _, item := range step.ProductSpecificationTemplates {
		if item.ProductSpecificationTemplate == nil {
			continue
		}
		for _, definition := range item.ProductSpecificationTemplate.SpecDefinitions {
			slug := strings.TrimSpace(definition.Slug)
			if slug == "" || !definition.IsVisible || !definition.IsFilterable {
				continue
			}
			if _, exists := definitionsBySlug[slug]; !exists {
				definitionsBySlug[slug] = definition
			}
		}
	}

	definitions := make([]productdomain.SpecDefinition, 0, len(definitionsBySlug))
	for _, definition := range definitionsBySlug {
		definitions = append(definitions, definition)
	}
	sort.SliceStable(definitions, func(i, j int) bool {
		if definitions[i].SortOrder == definitions[j].SortOrder {
			return definitions[i].Slug < definitions[j].Slug
		}
		return definitions[i].SortOrder < definitions[j].SortOrder
	})
	return definitions
}

func quickBuyFilterableSpecSlugs(step quickbuy.Step) []string {
	definitions := quickBuyFilterableSpecDefinitions(step)
	slugs := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		slugs = append(slugs, definition.Slug)
	}
	return slugs
}

func normalizeQuickBuySpecFilters(step quickbuy.Step, input map[string][]string) (map[string][]string, error) {
	if len(input) == 0 {
		return nil, nil
	}

	allowed := make(map[string]struct{})
	for _, slug := range quickBuyFilterableSpecSlugs(step) {
		allowed[slug] = struct{}{}
	}

	result := make(map[string][]string)
	for rawSlug, rawValues := range input {
		slug := strings.TrimSpace(rawSlug)
		if slug == "" {
			continue
		}
		if _, exists := allowed[slug]; !exists {
			return nil, fmt.Errorf("%w: step %q does not expose filterable specification %q", ErrQuickBuyInvalid, step.StepKey, slug)
		}
		values := make([]string, 0, len(rawValues))
		seen := make(map[string]struct{}, len(rawValues))
		for _, rawValue := range rawValues {
			value := strings.TrimSpace(rawValue)
			if value == "" {
				continue
			}
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			values = append(values, value)
			if len(values) >= 64 {
				break
			}
		}
		if len(values) > 0 {
			result[slug] = values
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

func quickBuyStepFilters(step quickbuy.Step, valuesBySlug map[string][]string) []QuickBuySpecFilterView {
	definitions := quickBuyFilterableSpecDefinitions(step)
	result := make([]QuickBuySpecFilterView, 0, len(definitions))
	for _, definition := range definitions {
		values := []string{}
		if valuesBySlug != nil && valuesBySlug[definition.Slug] != nil {
			values = append(values, valuesBySlug[definition.Slug]...)
		}
		result = append(result, QuickBuySpecFilterView{
			ID:              definition.ID,
			Name:            definition.Name,
			Slug:            definition.Slug,
			Unit:            definition.Unit,
			FieldType:       definition.FieldType,
			Presentation:    definition.Presentation,
			IsVariantOption: definition.IsVariantOption,
			Multiple:        true,
			Values:          values,
		})
	}
	return result
}

func quickBuyProductSpecificationTemplateView(item productdomain.ProductSpecificationTemplate, primary bool) QuickBuyProductSpecificationTemplateView {
	return QuickBuyProductSpecificationTemplateView{
		ID:      item.ID,
		Slug:    item.Slug,
		Name:    item.Name,
		Primary: primary,
	}
}

func quickBuyProductCategoryView(item productdomain.ProductCategory, locale string, primary bool, resolvers ...PublicMediaURLResolver) QuickBuyProductCategoryView {
	resolver := quickBuyMediaResolver(resolvers)
	name := strings.TrimSpace(item.Name)
	for _, translation := range item.Translations {
		if locales.Normalize(translation.Locale) != locale {
			continue
		}
		if translatedName := strings.TrimSpace(translation.Name); translatedName != "" {
			name = translatedName
		}
		break
	}
	return QuickBuyProductCategoryView{
		ID:       item.ID,
		ParentID: item.ParentID,
		Slug:     item.Slug,
		Name:     name,
		Depth:    item.Depth,
		ImageURL: canonicalPublicMediaURL(resolver, item.ImageURL),
		Primary:  primary,
	}
}

func quickBuyMediaResolver(resolvers []PublicMediaURLResolver) PublicMediaURLResolver {
	if len(resolvers) == 0 {
		return nil
	}
	return resolvers[0]
}
