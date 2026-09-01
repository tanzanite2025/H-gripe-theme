package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	currencydomain "commerce-platform/internal/domain/currency"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/quickbuy"

	"github.com/google/uuid"
)

func normalizeQuickBuyKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	if !quickBuyKeyPattern.MatchString(value) {
		return ""
	}
	return value
}

func isDefaultQuickBuyFlowSlug(slug string) bool {
	return normalizeQuickBuyKey(slug) == quickBuyDefaultFlowSlug
}

func isDefaultQuickBuyStepKey(stepKey string) bool {
	normalized := normalizeQuickBuyKey(stepKey)
	for _, defaultStepKey := range quickBuyDefaultStepKeys {
		if normalized == defaultStepKey {
			return true
		}
	}
	return false
}

func isDefaultQuickBuyFlow(version quickbuy.Version) bool {
	return version.Flow != nil && isDefaultQuickBuyFlowSlug(version.Flow.Slug)
}

func normalizeDefaultQuickBuyFlow(flow *quickbuy.Flow) {
	if flow == nil || !isDefaultQuickBuyFlowSlug(flow.Slug) {
		return
	}
	flow.Slug = quickBuyDefaultFlowSlug
	flow.Name = quickBuyDefaultFlowName
	flow.Description = quickBuyDefaultFlowDescription
	flow.EntrySurface = quickBuyDefaultFlowEntrySurface
	flow.IsEnabled = true
	flow.SortOrder = quickBuyDefaultFlowSortOrder
}

func normalizeDefaultQuickBuySteps(steps []quickbuy.Step) error {
	stepByKey := make(map[string]quickbuy.Step, len(steps))
	for _, step := range steps {
		stepByKey[normalizeQuickBuyKey(step.StepKey)] = step
	}

	ordered := make([]quickbuy.Step, 0, len(steps))
	for _, stepKey := range quickBuyDefaultStepKeys {
		step, exists := stepByKey[stepKey]
		if !exists {
			return fmt.Errorf("%w: default quick-build step %q cannot be removed", ErrQuickBuyInvalid, stepKey)
		}
		ordered = append(ordered, step)
		delete(stepByKey, stepKey)
	}

	extraSteps := make([]quickbuy.Step, 0, len(stepByKey))
	for _, step := range stepByKey {
		extraSteps = append(extraSteps, step)
	}
	sort.SliceStable(extraSteps, func(i, j int) bool {
		if extraSteps[i].SortOrder == extraSteps[j].SortOrder {
			return extraSteps[i].StepKey < extraSteps[j].StepKey
		}
		return extraSteps[i].SortOrder < extraSteps[j].SortOrder
	})
	ordered = append(ordered, extraSteps...)

	for index := range ordered {
		step := &ordered[index]
		step.SortOrder = (index + 1) * 10
		step.SelectionMode = quickbuy.SelectionModeSingle
		step.IsRequired = true
		step.MinSelect = 0
		step.MaxSelect = 1
		step.DefaultQuantity = 1
		step.AllowSkip = false
		step.ProductSpecificationTemplates = nil
		if index < len(quickBuyDefaultStepKeys) {
			step.StepKey = quickBuyDefaultStepKeys[index]
		}
	}

	copy(steps, ordered)
	return nil
}

func validateQuickBuyDefaultSteps(flowSlug string, steps []quickbuy.Step) error {
	if !isDefaultQuickBuyFlowSlug(flowSlug) {
		return nil
	}
	stepKeys := make(map[string]struct{}, len(steps))
	for _, step := range steps {
		stepKeys[normalizeQuickBuyKey(step.StepKey)] = struct{}{}
	}
	for _, requiredStepKey := range quickBuyDefaultStepKeys {
		if _, exists := stepKeys[requiredStepKey]; !exists {
			return fmt.Errorf("%w: default quick-build step %q cannot be removed", ErrQuickBuyInvalid, requiredStepKey)
		}
	}
	return nil
}

func normalizeQuickBuySurface(value string) string {
	value = normalizeQuickBuyKey(value)
	if value == "" {
		return "dock"
	}
	return value
}

func quickBuySelectionModeIsValid(value string) bool {
	switch value {
	case quickbuy.SelectionModeSingle, quickbuy.SelectionModeMultiple, quickbuy.SelectionModeQuantity, quickbuy.SelectionModeAuto:
		return true
	default:
		return false
	}
}

func normalizeQuickBuyCountry(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func quickBuyVersionIsActive(version quickbuy.Version, now time.Time) bool {
	if version.Status != quickbuy.FlowVersionStatusPublished || version.Flow == nil || !version.Flow.IsEnabled {
		return false
	}
	if version.StartsAt != nil && version.StartsAt.After(now) {
		return false
	}
	if version.EndsAt != nil && !version.EndsAt.After(now) {
		return false
	}
	return true
}

func normalizeQuickBuyCurrency(value string) string {
	currency := currencydomain.NormalizeCode(value)
	if currency == "" || !currencydomain.IsCatalogCode(currency) {
		return productdomain.DefaultPriceCurrency
	}
	return currency
}

func generateQuickBuySessionToken() string {
	return "qb_" + strings.ReplaceAll(uuid.NewString(), "-", "")
}
