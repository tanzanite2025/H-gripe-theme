package service

import (
	productdomain "commerce-platform/internal/domain/product"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	domainmoney "commerce-platform/internal/domain/money"
)

const (
	ProductConfigurationSchemaVersion = 1
	ProductSpecRoleCustomOption       = "custom_option"
	ProductSpecSelectionSingle        = "single"
	ProductSpecSelectionMultiple      = "multiple"
	ProductCustomOptionInventoryNone  = "none"
)

var (
	ErrProductConfigurationInvalid      = errors.New("product configuration is invalid")
	ErrProductConfigurationRequired     = errors.New("required product option group is missing")
	ErrProductConfigurationConflict     = errors.New("product configuration is no longer available")
	ErrProductConfigurationPriceChanged = errors.New("product configuration price has changed")
)

// SelectedOption is the only client-controlled part of a configuration. Labels,
// prices and IDs are deliberately absent and are resolved from the catalog.
type SelectedOption struct {
	GroupSlug string   `json:"group_slug"`
	ValueKeys []string `json:"value_keys"`
}

type productConfigurationSelection struct {
	GroupSlug string   `json:"group_slug"`
	ValueKeys []string `json:"value_keys"`
}

type productConfigurationSnapshot struct {
	SchemaVersion int                             `json:"schema_version"`
	Selections    []productConfigurationSelection `json:"selections"`
}

type ProductConfigurationResult struct {
	Data       []byte
	Hash       string
	Delta      domainmoney.Money
	Selections []SelectedOption
	Snapshot   ProductConfigurationSnapshot
}

type ProductConfigurationSnapshot struct {
	SchemaVersion           int                                 `json:"schema_version"`
	ProductID               uint                                `json:"product_id"`
	VariantID               uint                                `json:"variant_id"`
	ProductName             string                              `json:"product_name"`
	VariantName             string                              `json:"variant_name"`
	ConfigurationHash       string                              `json:"configuration_hash"`
	NormalizedConfiguration json.RawMessage                     `json:"normalized_configuration"`
	Currency                string                              `json:"currency"`
	Selections              []ProductConfigurationGroupSnapshot `json:"selections"`
	PriceBreakdown          ProductConfigurationPriceBreakdown  `json:"price_breakdown"`
}

type ProductConfigurationGroupSnapshot struct {
	GroupSlug     string                              `json:"group_slug"`
	GroupName     string                              `json:"group_name"`
	SelectionMode string                              `json:"selection_mode"`
	Values        []ProductConfigurationValueSnapshot `json:"values"`
}

type ProductConfigurationValueSnapshot struct {
	ValueKey        string `json:"value_key"`
	ValueLabel      string `json:"value_label"`
	PriceDeltaMinor int64  `json:"price_delta_minor"`
	InventoryPolicy string `json:"inventory_policy"`
}

type ProductConfigurationPriceBreakdown struct {
	BaseUnitPriceMinor    int64 `json:"base_unit_price_minor"`
	OptionsUnitPriceMinor int64 `json:"options_unit_price_minor"`
	FinalUnitPriceMinor   int64 `json:"final_unit_price_minor"`
}

// ResolveProductConfiguration validates and canonicalizes selected custom
// options for one product/variant. It is intentionally strict: the server owns
// option labels, enablement, selection bounds and all price deltas.
func ResolveProductConfiguration(item *productdomain.Product, variant *productdomain.ProductVariant, selected []SelectedOption) (ProductConfigurationResult, error) {
	if item == nil || variant == nil {
		return ProductConfigurationResult{}, fmt.Errorf("%w: product and variant are required", ErrProductConfigurationInvalid)
	}
	currencyCode := strings.TrimSpace(variant.Currency)
	if currencyCode == "" {
		currencyCode = productdomain.DefaultPriceCurrency
	}
	zero, err := domainmoney.New(0, currencyCode)
	if err != nil {
		return ProductConfigurationResult{}, fmt.Errorf("%w: invalid option currency: %v", ErrProductConfigurationInvalid, err)
	}

	definitions := make(map[string]productdomain.SpecDefinition)
	if item.ProductSpecificationTemplate != nil {
		for _, definition := range item.ProductSpecificationTemplate.SpecDefinitions {
			if definition.RuntimeRole() == ProductSpecRoleCustomOption {
				definitions[definition.Slug] = definition
			}
		}
	}

	valuesByGroup := make(map[string]map[string]productdomain.ProductVariantOptionValue)
	for _, value := range item.VariantOptionValues {
		for groupSlug, definition := range definitions {
			if definition.ID != value.SpecDefinitionID {
				continue
			}
			if valuesByGroup[groupSlug] == nil {
				valuesByGroup[groupSlug] = make(map[string]productdomain.ProductVariantOptionValue)
			}
			valuesByGroup[groupSlug][value.ValueKey] = value
		}
	}
	groupRules := make(map[uint]productdomain.ProductOptionGroupVariantRule, len(variant.OptionGroupRules))
	for _, rule := range variant.OptionGroupRules {
		groupRules[rule.SpecDefinitionID] = rule
	}
	valueRules := make(map[uint]productdomain.ProductOptionValueVariantRule, len(variant.OptionValueRules))
	for _, rule := range variant.OptionValueRules {
		valueRules[rule.ProductVariantOptionValueID] = rule
	}

	requested := make(map[string][]string, len(selected))
	snapshotGroups := make(map[string]ProductConfigurationGroupSnapshot, len(selected))
	for _, selection := range selected {
		group := strings.TrimSpace(selection.GroupSlug)
		if group == "" {
			return ProductConfigurationResult{}, fmt.Errorf("%w: group_slug is required", ErrProductConfigurationInvalid)
		}
		if _, exists := requested[group]; exists {
			return ProductConfigurationResult{}, fmt.Errorf("%w: duplicate option group %q", ErrProductConfigurationInvalid, group)
		}
		definition, exists := definitions[group]
		if !exists {
			return ProductConfigurationResult{}, fmt.Errorf("%w: unknown custom option group %q", ErrProductConfigurationConflict, group)
		}
		groupRule, hasGroupRule := groupRules[definition.ID]
		if hasGroupRule && !groupRule.IsApplicable {
			return ProductConfigurationResult{}, fmt.Errorf("%w: option group %q is not applicable to this variant", ErrProductConfigurationConflict, group)
		}
		minSelections := definition.MinSelections
		maxSelections := definition.MaxSelections
		if hasGroupRule {
			if groupRule.MinSelectionsOverride != nil {
				minSelections = *groupRule.MinSelectionsOverride
			}
			if groupRule.MaxSelectionsOverride != nil {
				maxSelections = groupRule.MaxSelectionsOverride
			}
		}
		keys := append([]string(nil), selection.ValueKeys...)
		if len(keys) == 0 {
			return ProductConfigurationResult{}, fmt.Errorf("%w: option group %q has no values", ErrProductConfigurationInvalid, group)
		}
		seen := make(map[string]struct{}, len(keys))
		for i := range keys {
			keys[i] = strings.TrimSpace(keys[i])
			if keys[i] == "" {
				return ProductConfigurationResult{}, fmt.Errorf("%w: empty value_key in group %q", ErrProductConfigurationInvalid, group)
			}
			if _, duplicate := seen[keys[i]]; duplicate {
				return ProductConfigurationResult{}, fmt.Errorf("%w: duplicate value %q", ErrProductConfigurationInvalid, keys[i])
			}
			seen[keys[i]] = struct{}{}
		}
		if definition.SelectionMode == ProductSpecSelectionSingle && len(keys) != 1 {
			return ProductConfigurationResult{}, fmt.Errorf("%w: option group %q accepts exactly one value", ErrProductConfigurationInvalid, group)
		}
		if definition.SelectionMode != ProductSpecSelectionSingle && definition.SelectionMode != ProductSpecSelectionMultiple {
			return ProductConfigurationResult{}, fmt.Errorf("%w: option group %q has unsupported selection mode", ErrProductConfigurationInvalid, group)
		}
		if minSelections > 0 && len(keys) < minSelections {
			return ProductConfigurationResult{}, fmt.Errorf("%w: option group %q requires at least %d values", ErrProductConfigurationInvalid, group, minSelections)
		}
		if maxSelections != nil && len(keys) > *maxSelections {
			return ProductConfigurationResult{}, fmt.Errorf("%w: option group %q accepts at most %d values", ErrProductConfigurationInvalid, group, *maxSelections)
		}
		available := valuesByGroup[group]
		groupSnapshot := ProductConfigurationGroupSnapshot{GroupSlug: group, GroupName: definition.Name, SelectionMode: definition.SelectionMode, Values: make([]ProductConfigurationValueSnapshot, 0, len(keys))}
		for _, key := range keys {
			value, ok := available[key]
			if !ok || !value.IsEnabled {
				return ProductConfigurationResult{}, fmt.Errorf("%w: value %q in group %q is unavailable", ErrProductConfigurationConflict, key, group)
			}
			if rule, ok := valueRules[value.ID]; ok && !rule.IsEnabled {
				return ProductConfigurationResult{}, fmt.Errorf("%w: value %q in group %q is unavailable", ErrProductConfigurationConflict, key, group)
			}
			if value.CustomOptionPolicy == nil {
				return ProductConfigurationResult{}, fmt.Errorf("%w: value %q in group %q has no transaction policy", ErrProductConfigurationConflict, key, group)
			}
			policy := value.CustomOptionPolicy
			if strings.TrimSpace(policy.InventoryPolicy) == "" {
				policy.InventoryPolicy = ProductCustomOptionInventoryNone
			}
			if policy.InventoryPolicy != ProductCustomOptionInventoryNone {
				return ProductConfigurationResult{}, fmt.Errorf("%w: option %q uses unsupported inventory policy", ErrProductConfigurationConflict, key)
			}
			if policy.PriceDeltaMinor < 0 {
				return ProductConfigurationResult{}, fmt.Errorf("%w: option %q has a negative price delta", ErrProductConfigurationConflict, key)
			}
			priceDeltaMinor := policy.PriceDeltaMinor
			if rule, ok := valueRules[value.ID]; ok && rule.PriceDeltaMinorOverride != nil {
				priceDeltaMinor = *rule.PriceDeltaMinorOverride
			}
			groupSnapshot.Values = append(groupSnapshot.Values, ProductConfigurationValueSnapshot{ValueKey: value.ValueKey, ValueLabel: value.Label, PriceDeltaMinor: priceDeltaMinor, InventoryPolicy: policy.InventoryPolicy})
			delta, deltaErr := domainmoney.New(priceDeltaMinor, currencyCode)
			if deltaErr != nil {
				return ProductConfigurationResult{}, fmt.Errorf("%w: option %q price delta: %v", ErrProductConfigurationConflict, key, deltaErr)
			}
			zero, err = zero.Add(delta)
			if err != nil {
				return ProductConfigurationResult{}, fmt.Errorf("%w: option price delta overflow", ErrProductConfigurationInvalid)
			}
		}
		requested[group] = keys
		snapshotGroups[group] = groupSnapshot
	}

	for group, definition := range definitions {
		rule, hasRule := groupRules[definition.ID]
		if hasRule && !rule.IsApplicable {
			continue
		}
		minSelections := definition.MinSelections
		if hasRule && rule.MinSelectionsOverride != nil {
			minSelections = *rule.MinSelectionsOverride
		}
		if (definition.IsRequired || minSelections > 0) && len(requested[group]) == 0 {
			return ProductConfigurationResult{}, fmt.Errorf("%w: group %q", ErrProductConfigurationRequired, group)
		}
	}

	groups := make([]string, 0, len(requested))
	for group := range requested {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	normalized := productConfigurationSnapshot{SchemaVersion: ProductConfigurationSchemaVersion, Selections: make([]productConfigurationSelection, 0, len(groups))}
	resultSelections := make([]SelectedOption, 0, len(groups))
	resultSnapshotGroups := make([]ProductConfigurationGroupSnapshot, 0, len(groups))
	for _, group := range groups {
		keys := append([]string(nil), requested[group]...)
		sort.Strings(keys)
		if groupSnapshot := snapshotGroups[group]; len(groupSnapshot.Values) > 1 {
			sort.Slice(groupSnapshot.Values, func(i, j int) bool { return groupSnapshot.Values[i].ValueKey < groupSnapshot.Values[j].ValueKey })
			snapshotGroups[group] = groupSnapshot
		}
		normalized.Selections = append(normalized.Selections, productConfigurationSelection{GroupSlug: group, ValueKeys: keys})
		resultSelections = append(resultSelections, SelectedOption{GroupSlug: group, ValueKeys: keys})
		resultSnapshotGroups = append(resultSnapshotGroups, snapshotGroups[group])
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return ProductConfigurationResult{}, fmt.Errorf("%w: encode normalized configuration: %v", ErrProductConfigurationInvalid, err)
	}
	digest := sha256.Sum256(data)
	return ProductConfigurationResult{
		Data: data, Hash: hex.EncodeToString(digest[:]), Delta: zero, Selections: resultSelections,
		Snapshot: ProductConfigurationSnapshot{SchemaVersion: ProductConfigurationSchemaVersion, ProductID: item.ID, VariantID: variant.ID, Currency: currencyCode, Selections: resultSnapshotGroups},
	}, nil
}

// ProductCustomOptionDefinitions returns the public-facing custom option
// groups for a product and the variant-specific applicability overlays.
func ProductCustomOptionDefinitions(item *productdomain.Product, variant *productdomain.ProductVariant) []productdomain.SpecDefinition {
	if item == nil || item.ProductSpecificationTemplate == nil {
		return nil
	}
	groupRules := make(map[uint]productdomain.ProductOptionGroupVariantRule)
	if variant != nil {
		for _, rule := range variant.OptionGroupRules {
			groupRules[rule.SpecDefinitionID] = rule
		}
	}
	definitions := make([]productdomain.SpecDefinition, 0)
	for _, definition := range item.ProductSpecificationTemplate.SpecDefinitions {
		if definition.RuntimeRole() != ProductSpecRoleCustomOption {
			continue
		}
		if rule, ok := groupRules[definition.ID]; ok && !rule.IsApplicable {
			continue
		}
		definitions = append(definitions, definition)
	}
	return definitions
}

func ParseSelectedOptions(data []byte) ([]SelectedOption, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}
	var selections []SelectedOption
	if err := json.Unmarshal(data, &selections); err != nil {
		return nil, fmt.Errorf("%w: selected_options must be an array", ErrProductConfigurationInvalid)
	}
	return selections, nil
}

func SelectedOptionsFromConfiguration(data []byte) ([]SelectedOption, error) {
	if len(data) == 0 || string(data) == "null" || string(data) == "{}" {
		return nil, nil
	}
	var snapshot productConfigurationSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("%w: configuration is not valid JSON", ErrProductConfigurationInvalid)
	}
	if snapshot.SchemaVersion != 0 && snapshot.SchemaVersion != ProductConfigurationSchemaVersion {
		return nil, fmt.Errorf("%w: unsupported configuration schema version", ErrProductConfigurationInvalid)
	}
	result := make([]SelectedOption, 0, len(snapshot.Selections))
	for _, selection := range snapshot.Selections {
		result = append(result, SelectedOption{GroupSlug: selection.GroupSlug, ValueKeys: selection.ValueKeys})
	}
	return result, nil
}
