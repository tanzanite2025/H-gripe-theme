package service

import (
	"fmt"
	"sort"
	"strings"

	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/quickbuy"
	"commerce-platform/internal/repository"
)

func (s *QuickBuyService) normalizeFlowAndVersion(input QuickBuyFlowInput) (*quickbuy.Flow, *quickbuy.Version, error) {
	flow, err := normalizeFlowInput(input)
	if err != nil {
		return nil, nil, err
	}
	normalizeDefaultQuickBuyFlow(flow)
	version, err := s.normalizeVersionInputForFlow(flow.Slug, input.Version)
	if err != nil {
		return nil, nil, err
	}
	if err := validateQuickBuyDefaultSteps(flow.Slug, version.Steps); err != nil {
		return nil, nil, err
	}
	version.VersionNumber = 1
	version.Status = quickbuy.FlowVersionStatusDraft
	return flow, version, nil
}

func normalizeFlowInput(input QuickBuyFlowInput) (*quickbuy.Flow, error) {
	slug := normalizeQuickBuyKey(input.Slug)
	if slug == "" {
		return nil, fmt.Errorf("%w: slug is required", ErrQuickBuyInvalid)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrQuickBuyInvalid)
	}
	enabled := true
	if input.IsEnabled != nil {
		enabled = *input.IsEnabled
	}
	sortOrder := input.SortOrder
	if sortOrder <= 0 {
		sortOrder = 100
	}
	translations, err := normalizeQuickBuyFlowTranslations(input.Translations)
	if err != nil {
		return nil, err
	}
	flow := &quickbuy.Flow{
		Slug:         slug,
		Name:         name,
		Description:  strings.TrimSpace(input.Description),
		HelpText:     strings.TrimSpace(input.HelpText),
		Translations: translations,
		EntrySurface: normalizeQuickBuySurface(input.EntrySurface),
		IsEnabled:    enabled,
		SortOrder:    sortOrder,
	}
	return flow, nil
}

func normalizeQuickBuyFlowTranslations(input []QuickBuyFlowTranslationInput) ([]quickbuy.FlowTranslation, error) {
	if len(input) == 0 {
		return nil, nil
	}

	result := make([]quickbuy.FlowTranslation, 0, len(input))
	seenLocales := make(map[string]struct{}, len(input))
	for index, item := range input {
		localeValue := strings.TrimSpace(item.Locale)
		if localeValue == "" {
			continue
		}
		locale, err := requireSupportedLocale(localeValue)
		if err != nil {
			return nil, fmt.Errorf("%w: flow translation %d has invalid locale", ErrQuickBuyInvalid, index+1)
		}
		helpText := strings.TrimSpace(item.HelpText)
		if helpText == "" {
			continue
		}
		if _, exists := seenLocales[locale]; exists {
			return nil, fmt.Errorf("%w: flow has duplicate translation locale %q", ErrQuickBuyInvalid, locale)
		}
		seenLocales[locale] = struct{}{}
		result = append(result, quickbuy.FlowTranslation{
			ID:       item.ID,
			Locale:   locale,
			HelpText: helpText,
		})
	}
	return result, nil
}

func (s *QuickBuyService) normalizeVersionInput(input QuickBuyVersionInput) (*quickbuy.Version, error) {
	if input.EndsAt != nil && input.StartsAt != nil && !input.EndsAt.After(*input.StartsAt) {
		return nil, fmt.Errorf("%w: ends_at must be after starts_at", ErrQuickBuyInvalid)
	}
	steps, err := s.normalizeStepInputs(input.Steps)
	if err != nil {
		return nil, err
	}
	return &quickbuy.Version{
		Status:   quickbuy.FlowVersionStatusDraft,
		StartsAt: input.StartsAt,
		EndsAt:   input.EndsAt,
		Steps:    steps,
	}, nil
}

func (s *QuickBuyService) normalizeVersionInputForFlow(flowSlug string, input QuickBuyVersionInput) (*quickbuy.Version, error) {
	version, err := s.normalizeVersionInput(input)
	if err != nil {
		return nil, err
	}
	if isDefaultQuickBuyFlowSlug(flowSlug) {
		if err := normalizeDefaultQuickBuySteps(version.Steps); err != nil {
			return nil, err
		}
	}
	return version, nil
}

func (s *QuickBuyService) normalizeStepInputs(input []QuickBuyStepInput) ([]quickbuy.Step, error) {
	steps := make([]quickbuy.Step, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for index, item := range input {
		stepKey := normalizeQuickBuyKey(item.StepKey)
		if stepKey == "" {
			return nil, fmt.Errorf("%w: step %d key is required", ErrQuickBuyInvalid, index+1)
		}
		if _, exists := seen[stepKey]; exists {
			return nil, fmt.Errorf("%w: duplicate step key %q", ErrQuickBuyInvalid, stepKey)
		}
		seen[stepKey] = struct{}{}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: step %q name is required", ErrQuickBuyInvalid, stepKey)
		}
		productCategories, err := s.normalizeStepProductCategories(item.ProductCategoryIDs)
		if err != nil {
			return nil, fmt.Errorf("%w: step %q product categories: %v", ErrQuickBuyInvalid, stepKey, err)
		}
		productSpecificationTemplates, err := s.normalizeStepProductSpecificationTemplates(item.ProductSpecificationTemplateIDs)
		if err != nil {
			return nil, fmt.Errorf("%w: step %q product specification templates: %v", ErrQuickBuyInvalid, stepKey, err)
		}
		steps = append(steps, quickbuy.Step{
			StepKey:                       stepKey,
			Name:                          name,
			SortOrder:                     (index + 1) * 10,
			SelectionMode:                 quickbuy.SelectionModeSingle,
			IsRequired:                    true,
			MinSelect:                     0,
			MaxSelect:                     1,
			DefaultQuantity:               1,
			AllowSkip:                     false,
			ProductCategories:             productCategories,
			ProductSpecificationTemplates: productSpecificationTemplates,
		})
	}
	sort.SliceStable(steps, func(i, j int) bool {
		if steps[i].SortOrder == steps[j].SortOrder {
			return steps[i].StepKey < steps[j].StepKey
		}
		return steps[i].SortOrder < steps[j].SortOrder
	})
	return steps, nil
}

func (s *QuickBuyService) normalizeStepProductCategories(ids []uint) ([]quickbuy.StepProductCategory, error) {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]quickbuy.StepProductCategory, 0, len(ids))
	for index, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		if s.productCategoryRepo != nil {
			if _, err := s.productCategoryRepo.FindByID(id); err != nil {
				if repository.IsRecordNotFound(err) {
					return nil, fmt.Errorf("product category %d does not exist", id)
				}
				return nil, err
			}
		}
		result = append(result, quickbuy.StepProductCategory{
			ProductCategoryID: id,
			IsPrimary:         len(result) == 0,
			SortOrder:         (index + 1) * 10,
		})
	}
	return result, nil
}

func (s *QuickBuyService) normalizeStepProductSpecificationTemplates(ids []uint) ([]quickbuy.StepProductSpecificationTemplate, error) {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]quickbuy.StepProductSpecificationTemplate, 0, len(ids))
	for index, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		if s.productRepo != nil {
			if _, err := s.productRepo.FindProductSpecificationTemplateByID(id); err != nil {
				if repository.IsRecordNotFound(err) {
					return nil, fmt.Errorf("product specification template %d does not exist", id)
				}
				return nil, err
			}
		}
		result = append(result, quickbuy.StepProductSpecificationTemplate{
			ProductSpecificationTemplateID: id,
			IsPrimary:                      len(result) == 0,
			SortOrder:                      (index + 1) * 10,
		})
	}
	return result, nil
}

func validateQuickBuyVersionForPublish(version quickbuy.Version) error {
	result := validateQuickBuyVersion(version)
	if result.Valid {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrQuickBuyInvalid, result.errorSummary())
}

func validateQuickBuyVersion(version quickbuy.Version) QuickBuyValidationResult {
	result := QuickBuyValidationResult{Valid: true, Issues: []QuickBuyValidationIssue{}}
	if version.Flow == nil {
		result.addIssue("error", "flow_missing", "version is not linked to a QUICK flow", "", "", 0)
	}
	if version.EndsAt != nil && version.StartsAt != nil && !version.EndsAt.After(*version.StartsAt) {
		result.addIssue("error", "invalid_time_window", "ends_at must be after starts_at", "", "", 0)
	}
	if len(version.Steps) == 0 {
		result.addIssue("error", "steps_required", "at least one step is required before publishing", "", "", 0)
		return result
	}

	stepKeys := make(map[string]struct{}, len(version.Steps))
	for _, step := range version.Steps {
		stepKey := normalizeQuickBuyKey(step.StepKey)
		if stepKey == "" || stepKey != step.StepKey {
			result.addIssue("error", "invalid_step_key", fmt.Sprintf("step key %q is invalid", step.StepKey), step.StepKey, "", 0)
		}
		if _, exists := stepKeys[step.StepKey]; exists {
			result.addIssue("error", "duplicate_step_key", fmt.Sprintf("step key %q is duplicated", step.StepKey), step.StepKey, "", 0)
		}
		stepKeys[step.StepKey] = struct{}{}
		if strings.TrimSpace(step.Name) == "" {
			result.addIssue("error", "step_name_required", "step name is required", step.StepKey, "", 0)
		}
		if !quickBuySelectionModeIsValid(step.SelectionMode) {
			result.addIssue("error", "invalid_selection_mode", fmt.Sprintf("step %q has invalid selection mode %q", step.StepKey, step.SelectionMode), step.StepKey, "", 0)
		}
		if step.MinSelect < 0 || step.MaxSelect < 0 || step.DefaultQuantity <= 0 {
			result.addIssue("error", "invalid_selection_bounds", fmt.Sprintf("step %q has invalid quantity or selection bounds", step.StepKey), step.StepKey, "", 0)
		}
		if step.SelectionMode != quickbuy.SelectionModeAuto && step.MaxSelect > 0 && step.MinSelect > step.MaxSelect {
			result.addIssue("error", "invalid_selection_bounds", fmt.Sprintf("step %q min_select cannot exceed max_select", step.StepKey), step.StepKey, "", 0)
		}
		if step.SelectionMode == quickbuy.SelectionModeSingle && step.MaxSelect > 1 {
			result.addIssue("error", "single_step_max_select", fmt.Sprintf("single-select step %q cannot allow more than one selection", step.StepKey), step.StepKey, "", 0)
		}
		if step.IsRequired && len(step.ProductCategories) == 0 && len(step.ProductSpecificationTemplates) == 0 && step.SelectionMode != quickbuy.SelectionModeAuto && !isDefaultQuickBuyFlow(version) {
			result.addIssue("error", "required_step_product_categories", fmt.Sprintf("required step %q needs at least one product category", step.StepKey), step.StepKey, "", 0)
		}
		if step.IsRequired && step.AllowSkip {
			result.addIssue("warning", "required_step_allows_skip", fmt.Sprintf("required step %q is also marked as skippable", step.StepKey), step.StepKey, "", 0)
		}
		if !step.IsRequired && step.MinSelect > 0 {
			result.addIssue("warning", "optional_step_min_select", fmt.Sprintf("optional step %q has min_select greater than zero", step.StepKey), step.StepKey, "", 0)
		}

		productCategories := make(map[uint]struct{}, len(step.ProductCategories))
		for _, item := range step.ProductCategories {
			if item.ProductCategoryID == 0 {
				result.addProductCategoryIssue("error", "invalid_product_category", fmt.Sprintf("step %q contains an empty product category reference", step.StepKey), step.StepKey, item.ProductCategoryID)
				continue
			}
			if _, exists := productCategories[item.ProductCategoryID]; exists {
				result.addProductCategoryIssue("error", "duplicate_product_category", fmt.Sprintf("step %q references product category %d more than once", step.StepKey, item.ProductCategoryID), step.StepKey, item.ProductCategoryID)
				continue
			}
			productCategories[item.ProductCategoryID] = struct{}{}
			if item.ProductCategory == nil {
				result.addProductCategoryIssue("error", "missing_product_category", fmt.Sprintf("step %q references missing product category %d", step.StepKey, item.ProductCategoryID), step.StepKey, item.ProductCategoryID)
				continue
			}
			if !item.ProductCategory.IsEnabled {
				result.addProductCategoryIssue("error", "disabled_product_category", fmt.Sprintf("step %q references disabled product category %q", step.StepKey, item.ProductCategory.Slug), step.StepKey, item.ProductCategoryID)
			}
		}

		if !isDefaultQuickBuyFlow(version) {
			productSpecificationTemplates := make(map[uint]struct{}, len(step.ProductSpecificationTemplates))
			for _, item := range step.ProductSpecificationTemplates {
				if item.ProductSpecificationTemplateID == 0 {
					result.addIssue("error", "invalid_product_specification_template", fmt.Sprintf("step %q contains an empty product specification template reference", step.StepKey), step.StepKey, "", 0)
					continue
				}
				if _, exists := productSpecificationTemplates[item.ProductSpecificationTemplateID]; exists {
					result.addIssue("error", "duplicate_product_specification_template", fmt.Sprintf("step %q references product specification template %d more than once", step.StepKey, item.ProductSpecificationTemplateID), step.StepKey, "", item.ProductSpecificationTemplateID)
					continue
				}
				productSpecificationTemplates[item.ProductSpecificationTemplateID] = struct{}{}
				if item.ProductSpecificationTemplate == nil {
					result.addIssue("error", "missing_product_specification_template", fmt.Sprintf("step %q references missing product specification template %d", step.StepKey, item.ProductSpecificationTemplateID), step.StepKey, "", item.ProductSpecificationTemplateID)
					continue
				}
				if !item.ProductSpecificationTemplate.IsEnabled {
					result.addIssue("error", "disabled_product_specification_template", fmt.Sprintf("step %q references disabled product specification template %q", step.StepKey, item.ProductSpecificationTemplate.Slug), step.StepKey, "", item.ProductSpecificationTemplateID)
				}
			}
		}
	}
	if version.Flow != nil && isDefaultQuickBuyFlowSlug(version.Flow.Slug) {
		for _, requiredStepKey := range quickBuyDefaultStepKeys {
			if _, exists := stepKeys[requiredStepKey]; !exists {
				result.addIssue(
					"error",
					"default_step_required",
					fmt.Sprintf("default quick-build step %q cannot be removed", requiredStepKey),
					requiredStepKey,
					"",
					0,
				)
			}
		}
	}

	ruleKeys := make(map[string]struct{}, len(version.Rules))
	for _, rule := range version.Rules {
		if !rule.IsEnabled {
			continue
		}
		ruleKey := normalizeQuickBuyKey(rule.RuleKey)
		if ruleKey == "" || ruleKey != rule.RuleKey {
			result.addIssue("error", "invalid_rule_key", fmt.Sprintf("rule key %q is invalid", rule.RuleKey), "", rule.RuleKey, 0)
		}
		if _, exists := ruleKeys[rule.RuleKey]; exists {
			result.addIssue("error", "duplicate_rule_key", fmt.Sprintf("rule key %q is duplicated", rule.RuleKey), "", rule.RuleKey, 0)
		}
		ruleKeys[rule.RuleKey] = struct{}{}
		if rule.SourceStepKey != "" {
			if _, exists := stepKeys[rule.SourceStepKey]; !exists {
				result.addIssue("error", "rule_source_step_missing", fmt.Sprintf("rule %q references missing source step %q", rule.RuleKey, rule.SourceStepKey), rule.SourceStepKey, rule.RuleKey, 0)
			}
		}
		if rule.TargetStepKey != "" {
			if _, exists := stepKeys[rule.TargetStepKey]; !exists {
				result.addIssue("error", "rule_target_step_missing", fmt.Sprintf("rule %q references missing target step %q", rule.RuleKey, rule.TargetStepKey), rule.TargetStepKey, rule.RuleKey, 0)
			}
		}
		if rule.Severity != "error" && rule.Severity != "warning" && rule.Severity != "info" {
			result.addIssue("error", "invalid_rule_severity", fmt.Sprintf("rule %q has invalid severity %q", rule.RuleKey, rule.Severity), "", rule.RuleKey, 0)
		}
	}

	return result
}

func (result *QuickBuyValidationResult) addIssue(severity, code, message, stepKey, ruleKey string, productSpecificationTemplateID uint) {
	if severity == "error" {
		result.Valid = false
	}
	result.Issues = append(result.Issues, QuickBuyValidationIssue{
		Severity:                       severity,
		Code:                           code,
		Message:                        message,
		StepKey:                        stepKey,
		RuleKey:                        ruleKey,
		ProductSpecificationTemplateID: productSpecificationTemplateID,
	})
}

func (result *QuickBuyValidationResult) addProductCategoryIssue(severity, code, message, stepKey string, productCategoryID uint) {
	if severity == "error" {
		result.Valid = false
	}
	result.Issues = append(result.Issues, QuickBuyValidationIssue{
		Severity:          severity,
		Code:              code,
		Message:           message,
		StepKey:           stepKey,
		ProductCategoryID: productCategoryID,
	})
}

func (result QuickBuyValidationResult) errorSummary() string {
	for _, issue := range result.Issues {
		if issue.Severity == "error" {
			return issue.Message
		}
	}
	return "validation failed"
}

func (result *QuickBuySessionValidationResult) addIssue(code, message, stepKey string, productID uint, variantID *uint) {
	result.Valid = false
	result.Issues = append(result.Issues, QuickBuySessionValidationIssue{
		Severity:  "error",
		Code:      code,
		Message:   message,
		StepKey:   stepKey,
		ProductID: productID,
		VariantID: variantID,
	})
}

func quickBuySessionValidationStatus(result QuickBuySessionValidationResult) string {
	hasWarning := false
	for _, issue := range result.Issues {
		if issue.Severity == "error" {
			return quickbuy.ValidationStatusInvalid
		}
		if issue.Severity == "warning" {
			hasWarning = true
		}
	}
	if hasWarning {
		return quickbuy.ValidationStatusWarning
	}
	return quickbuy.ValidationStatusValid
}

func validateQuickBuySelectionBounds(version quickbuy.Version, items []quickbuy.SessionItem) error {
	itemsByStep := make(map[string][]quickbuy.SessionItem, len(items))
	for _, item := range items {
		itemsByStep[item.StepKey] = append(itemsByStep[item.StepKey], item)
	}
	for _, step := range version.Steps {
		stepItems := itemsByStep[step.StepKey]
		if step.SelectionMode == quickbuy.SelectionModeSingle && len(stepItems) > 1 {
			return fmt.Errorf("%w: step %q accepts only one selection", ErrQuickBuyInvalid, step.StepKey)
		}
		if step.MaxSelect > 0 && len(stepItems) > step.MaxSelect {
			return fmt.Errorf("%w: step %q exceeds max_select", ErrQuickBuyInvalid, step.StepKey)
		}
	}
	return nil
}

func (s *QuickBuyService) validateQuickBuyProductAllowedForStep(step quickbuy.Step, item productdomain.Product, enforceProductSpecificationTemplates bool) error {
	if err := s.validateQuickBuyProductCategoryAllowedForStep(step, item); err != nil {
		return err
	}
	if !enforceProductSpecificationTemplates {
		return nil
	}
	return validateQuickBuyProductSpecificationTemplateAllowedForStep(step, item)
}

func (s *QuickBuyService) validateQuickBuyProductCategoryAllowedForStep(step quickbuy.Step, item productdomain.Product) error {
	categoryIDs := quickBuyStepProductCategoryIDs(step)
	if len(categoryIDs) == 0 {
		return nil
	}
	if item.ProductCategoryID == nil {
		return fmt.Errorf("%w: product %d has no product category for step %q", ErrQuickBuyInvalid, item.ID, step.StepKey)
	}
	if s.productRepo == nil {
		return nil
	}
	allowed, err := s.productRepo.ProductCategoryInQuickBuyScope(item.ProductCategoryID, categoryIDs)
	if err != nil {
		return err
	}
	if allowed {
		return nil
	}
	return fmt.Errorf("%w: product %d is not in an allowed product category for step %q", ErrQuickBuyInvalid, item.ID, step.StepKey)
}

func validateQuickBuyProductSpecificationTemplateAllowedForStep(step quickbuy.Step, item productdomain.Product) error {
	if len(step.ProductSpecificationTemplates) == 0 {
		return nil
	}
	if item.ProductSpecificationTemplateID == nil {
		return fmt.Errorf("%w: product %d has no product specification template for step %q", ErrQuickBuyInvalid, item.ID, step.StepKey)
	}
	for _, productSpecificationTemplate := range step.ProductSpecificationTemplates {
		if productSpecificationTemplate.ProductSpecificationTemplateID == *item.ProductSpecificationTemplateID {
			return nil
		}
	}
	return fmt.Errorf("%w: product %d is not allowed for step %q", ErrQuickBuyInvalid, item.ID, step.StepKey)
}

func quickBuyStepByKey(version quickbuy.Version, stepKey string) *quickbuy.Step {
	for index := range version.Steps {
		if version.Steps[index].StepKey == stepKey {
			return &version.Steps[index]
		}
	}
	return nil
}

func quickBuyStepProductCategoryIDs(step quickbuy.Step) []uint {
	ids := make([]uint, 0, len(step.ProductCategories))
	for _, item := range step.ProductCategories {
		if item.ProductCategoryID == 0 {
			continue
		}
		ids = append(ids, item.ProductCategoryID)
	}
	return ids
}

func quickBuyStepProductSpecificationTemplateIDs(step quickbuy.Step) []uint {
	ids := make([]uint, 0, len(step.ProductSpecificationTemplates))
	for _, item := range step.ProductSpecificationTemplates {
		if item.ProductSpecificationTemplateID == 0 {
			continue
		}
		ids = append(ids, item.ProductSpecificationTemplateID)
	}
	return ids
}

func quickBuyStepProductSpecificationTemplateIDsForVersion(version quickbuy.Version, step quickbuy.Step) []uint {
	if isDefaultQuickBuyFlow(version) {
		return nil
	}
	return quickBuyStepProductSpecificationTemplateIDs(step)
}

func normalizeQuickBuyCandidatePaging(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = quickBuyCandidateDefaultPageSize
	}
	if pageSize > quickBuyCandidateMaxPageSize {
		pageSize = quickBuyCandidateMaxPageSize
	}
	return page, pageSize
}
