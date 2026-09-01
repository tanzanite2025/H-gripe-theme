package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/quickbuy"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/repository"
)

func (s *QuickBuyService) findCurrentPublishedVersion(surface string) (*quickbuy.Version, error) {
	surface = normalizeQuickBuySurface(surface)

	versions, err := s.repo.ListPublishedVersions(surface, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, nil
	}
	version := versions[0]
	return &version, nil
}

func (s *QuickBuyService) resolveSessionVersion(input QuickBuySessionInput) (*quickbuy.Version, string, string, error) {
	locale := locales.ResolveSupported(input.Locale)
	country := normalizeQuickBuyCountry(input.MarketCountry)
	if input.FlowVersionID == 0 {
		version, err := s.findCurrentPublishedVersion(input.Surface)
		return version, locale, country, err
	}

	version, err := s.repo.FindVersionByID(input.FlowVersionID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, locale, country, ErrQuickBuyNotFound
		}
		return nil, locale, country, err
	}
	if version.Status != quickbuy.FlowVersionStatusPublished || version.Flow == nil || !version.Flow.IsEnabled {
		return nil, locale, country, ErrQuickBuyNotFound
	}
	if !quickBuyVersionIsActive(*version, time.Now().UTC()) {
		return nil, locale, country, ErrQuickBuyNotFound
	}
	if input.FlowID > 0 && version.FlowID != input.FlowID {
		return nil, locale, country, fmt.Errorf("%w: flow_id does not match flow_version_id", ErrQuickBuyInvalid)
	}
	return version, locale, country, nil
}

func (s *QuickBuyService) findActiveSession(token string) (*quickbuy.Session, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrQuickBuySessionNotFound
	}
	session, err := s.repo.FindSessionByToken(token)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuySessionNotFound
		}
		return nil, err
	}
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrQuickBuySessionNotFound
	}
	return session, nil
}

func (s *QuickBuyService) sessionVersion(session *quickbuy.Session) (*quickbuy.Version, error) {
	if session.Version != nil {
		return session.Version, nil
	}
	version, err := s.repo.FindVersionByID(session.FlowVersionID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrQuickBuyNotFound
		}
		return nil, err
	}
	return version, nil
}

func (s *QuickBuyService) listVersionStepCandidates(version quickbuy.Version, input QuickBuyCandidateInput) (*QuickBuyCandidateResult, error) {
	if s.productRepo == nil {
		return nil, errors.New("product repository is not configured")
	}

	stepKey := normalizeQuickBuyKey(input.StepKey)
	if stepKey == "" {
		return nil, fmt.Errorf("%w: step_key is required", ErrQuickBuyInvalid)
	}
	step := quickBuyStepByKey(version, stepKey)
	if step == nil {
		return nil, fmt.Errorf("%w: step %q does not exist in this QUICK version", ErrQuickBuyInvalid, stepKey)
	}

	locale := locales.ResolveSupported(input.Locale)
	currency := normalizeQuickBuyCurrency(input.Currency)
	page, pageSize := normalizeQuickBuyCandidatePaging(input.Page, input.PageSize)
	exposeProductSpecificationTemplates := !isDefaultQuickBuyFlow(version) && len(step.ProductSpecificationTemplates) > 0
	if !exposeProductSpecificationTemplates {
		input.SpecFilters = nil
	}
	specFilters, err := normalizeQuickBuySpecFilters(*step, input.SpecFilters)
	if err != nil {
		return nil, err
	}
	result := &QuickBuyCandidateResult{
		FlowID:        version.FlowID,
		FlowVersionID: version.ID,
		Locale:        locale,
		Currency:      currency,
		Step:          quickBuyStepView(*step, locale, exposeProductSpecificationTemplates, s.mediaURLResolver),
		Products:      []productdomain.Product{},
		Page:          page,
		PageSize:      pageSize,
	}
	if step.SelectionMode == quickbuy.SelectionModeAuto {
		return result, nil
	}

	products, total, err := s.productRepo.ListQuickBuyCandidates(repository.ProductQuickBuyCandidateQuery{
		Locale:                          locale,
		ProductSpecificationTemplateIDs: quickBuyStepProductSpecificationTemplateIDsForVersion(version, *step),
		ProductCategoryIDs:              quickBuyStepProductCategoryIDs(*step),
		Keyword:                         strings.TrimSpace(input.Keyword),
		SpecFilters:                     specFilters,
		Offset:                          (page - 1) * pageSize,
		Limit:                           pageSize,
	})
	if err != nil {
		return nil, err
	}
	result.Products = products
	result.Total = total
	result.HasMore = int64(page*pageSize) < total
	if !exposeProductSpecificationTemplates {
		return result, nil
	}
	filterValues, err := s.productRepo.ListQuickBuyFilterValues(repository.ProductQuickBuyCandidateQuery{
		Locale:                          locale,
		ProductSpecificationTemplateIDs: quickBuyStepProductSpecificationTemplateIDsForVersion(version, *step),
		ProductCategoryIDs:              quickBuyStepProductCategoryIDs(*step),
		Keyword:                         strings.TrimSpace(input.Keyword),
		SpecFilters:                     specFilters,
	}, quickBuyFilterableSpecSlugs(*step))
	if err != nil {
		return nil, err
	}
	result.Step.Filters = quickBuyStepFiltersForScope(*step, filterValues, exposeProductSpecificationTemplates)
	return result, nil
}

func (s *QuickBuyService) sessionItemFromSelection(session quickbuy.Session, version quickbuy.Version, selection QuickBuySelectionInput, index int) (*quickbuy.SessionItem, bool, error) {
	stepKey := normalizeQuickBuyKey(selection.StepKey)
	if stepKey == "" {
		return nil, false, fmt.Errorf("%w: selection step_key is required", ErrQuickBuyInvalid)
	}
	step := quickBuyStepByKey(version, stepKey)
	if step == nil {
		return nil, false, fmt.Errorf("%w: step %q does not exist in this QUICK version", ErrQuickBuyInvalid, stepKey)
	}
	if selection.ProductID == 0 {
		return nil, true, nil
	}
	if step.SelectionMode == quickbuy.SelectionModeAuto {
		return nil, false, fmt.Errorf("%w: step %q does not accept manual selections", ErrQuickBuyInvalid, stepKey)
	}
	quantity := selection.Quantity
	if quantity <= 0 {
		quantity = step.DefaultQuantity
	}
	if quantity <= 0 {
		quantity = 1
	}
	if quantity > 999 {
		return nil, false, fmt.Errorf("%w: step %q quantity is too large", ErrQuickBuyInvalid, stepKey)
	}
	if s.productRepo == nil {
		return nil, false, errors.New("product repository is not configured")
	}
	productItem, variant, err := s.productRepo.FindPurchasableVariant(selection.ProductID, selection.VariantID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, false, fmt.Errorf("%w: product %d is not available", ErrQuickBuyInvalid, selection.ProductID)
		}
		return nil, false, err
	}
	if err := s.validateQuickBuyProductAllowedForStep(*step, *productItem, !isDefaultQuickBuyFlow(version)); err != nil {
		return nil, false, err
	}
	if variant.Stock < quantity {
		return nil, false, fmt.Errorf("%w: product %d does not have enough stock for step %q", ErrQuickBuyInvalid, selection.ProductID, stepKey)
	}

	variantID := variant.ID
	price := variant.EffectivePrice()
	currency := normalizeQuickBuyCurrency(variant.Currency)
	if currency == "" {
		currency = normalizeQuickBuyCurrency(productItem.DisplayPriceCurrency())
	}
	if currency == "" {
		currency = session.Currency
	}
	return &quickbuy.SessionItem{
		StepID:            step.ID,
		StepKey:           step.StepKey,
		ProductID:         productItem.ID,
		VariantID:         &variantID,
		Quantity:          quantity,
		UnitPriceSnapshot: price,
		CurrencySnapshot:  currency,
		WeightSnapshotG:   variant.Weight,
		ProductSnapshot:   quickBuyProductSnapshot(*productItem, s.mediaURLResolver),
		VariantSnapshot:   quickBuyVariantSnapshot(*variant),
		SortOrder:         step.SortOrder*100 + index + 1,
	}, false, nil
}

func (s *QuickBuyService) validateQuickBuySession(version quickbuy.Version, items []quickbuy.SessionItem) QuickBuySessionValidationResult {
	result := QuickBuySessionValidationResult{Valid: true, Issues: []QuickBuySessionValidationIssue{}}
	itemsByStep := make(map[string][]quickbuy.SessionItem, len(items))
	for _, item := range items {
		itemsByStep[item.StepKey] = append(itemsByStep[item.StepKey], item)
		step := quickBuyStepByKey(version, item.StepKey)
		if step == nil {
			result.addIssue("step_missing", fmt.Sprintf("selection references missing step %q", item.StepKey), item.StepKey, item.ProductID, item.VariantID)
			continue
		}
		if item.Quantity <= 0 {
			result.addIssue("invalid_quantity", fmt.Sprintf("step %q has a non-positive quantity", item.StepKey), item.StepKey, item.ProductID, item.VariantID)
		}
		if s.productRepo == nil {
			continue
		}
		productItem, variant, err := s.productRepo.FindPurchasableVariant(item.ProductID, item.VariantID)
		if err != nil {
			result.addIssue("product_unavailable", fmt.Sprintf("product %d is no longer available", item.ProductID), item.StepKey, item.ProductID, item.VariantID)
			continue
		}
		if err := s.validateQuickBuyProductAllowedForStep(*step, *productItem, !isDefaultQuickBuyFlow(version)); err != nil {
			result.addIssue("product_not_allowed", err.Error(), item.StepKey, item.ProductID, item.VariantID)
		}
		if variant.Stock < item.Quantity {
			result.addIssue("stock_unavailable", fmt.Sprintf("product %d no longer has enough stock", item.ProductID), item.StepKey, item.ProductID, item.VariantID)
		}
	}

	for _, step := range version.Steps {
		stepItems := itemsByStep[step.StepKey]
		if step.IsRequired && step.SelectionMode != quickbuy.SelectionModeAuto && len(stepItems) == 0 {
			result.addIssue("required_step_missing", fmt.Sprintf("required step %q has no selection", step.StepKey), step.StepKey, 0, nil)
		}
		if step.SelectionMode == quickbuy.SelectionModeSingle && len(stepItems) > 1 {
			result.addIssue("single_step_multiple_items", fmt.Sprintf("single-select step %q has more than one selection", step.StepKey), step.StepKey, 0, nil)
		}
		if step.MaxSelect > 0 && len(stepItems) > step.MaxSelect {
			result.addIssue("max_select_exceeded", fmt.Sprintf("step %q exceeds max_select", step.StepKey), step.StepKey, 0, nil)
		}
		if step.MinSelect > 0 && len(stepItems) < step.MinSelect {
			result.addIssue("min_select_missing", fmt.Sprintf("step %q has fewer selections than min_select", step.StepKey), step.StepKey, 0, nil)
		}
	}
	return result
}
