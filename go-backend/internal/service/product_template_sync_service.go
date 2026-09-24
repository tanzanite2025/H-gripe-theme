package service

import (
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	ProductTemplateSyncChangeAdded        = "added"
	ProductTemplateSyncChangeDisabled     = "disabled"
	ProductTemplateSyncChangeEnabled      = "enabled"
	ProductTemplateSyncChangeLabelChanged = "label_changed"
	ProductTemplateSyncChangePriceChanged = "price_changed"
	ProductTemplateSyncChangeRemoved      = "removed"
)

var ErrProductTemplateSyncRevisionConflict = errors.New("product template sync revision conflict")

// ProductTemplateSyncDiff is the operator-facing preview of changes between a
// product's materialized option values and its current template revision.
type ProductTemplateSyncDiff struct {
	ProductID            uint                           `json:"product_id"`
	TemplateID           uint                           `json:"template_id"`
	MaterializedRevision int                            `json:"materialized_revision"`
	TemplateRevision     int                            `json:"template_revision"`
	RevisionChanged      bool                           `json:"revision_changed"`
	HasChanges           bool                           `json:"has_changes"`
	Items                []ProductTemplateSyncDiffItem  `json:"items"`
	Summary              ProductTemplateSyncDiffSummary `json:"summary"`
}

type ProductTemplateSyncDiffSummary struct {
	Added        int `json:"added"`
	Disabled     int `json:"disabled"`
	Enabled      int `json:"enabled"`
	LabelChanged int `json:"label_changed"`
	PriceChanged int `json:"price_changed"`
	Removed      int `json:"removed"`
}

type ProductTemplateSyncDiffItem struct {
	SpecDefinitionID        uint     `json:"spec_definition_id"`
	GroupSlug               string   `json:"group_slug"`
	GroupName               string   `json:"group_name"`
	ValueKey                string   `json:"value_key"`
	Changes                 []string `json:"changes"`
	CurrentLabel            string   `json:"current_label,omitempty"`
	TemplateLabel           string   `json:"template_label,omitempty"`
	CurrentEnabled          *bool    `json:"current_enabled,omitempty"`
	TemplateEnabled         *bool    `json:"template_enabled,omitempty"`
	CurrentPriceDeltaMinor  *int64   `json:"current_price_delta_minor,omitempty"`
	TemplatePriceDeltaMinor *int64   `json:"template_price_delta_minor,omitempty"`
}

func (s *ProductService) PreviewProductTemplateSync(productID uint) (*ProductTemplateSyncDiff, error) {
	item, err := s.findProduct(productID)
	if err != nil {
		return nil, err
	}
	if item.ProductSpecificationTemplateID == nil || *item.ProductSpecificationTemplateID == 0 {
		return nil, fmt.Errorf("%w: product has no specification template", ErrProductSpecificationTemplateInvalid)
	}
	template, err := s.productRepo.FindProductSpecificationTemplateByID(*item.ProductSpecificationTemplateID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrProductSpecificationTemplateNotFound
		}
		return nil, err
	}
	return buildProductTemplateSyncDiff(item, template), nil
}

// SyncProductTemplate applies a previously previewed template revision. The
// expected revision makes the confirmation safe when another operator edits
// the template between preview and confirmation.
func (s *ProductService) SyncProductTemplate(productID uint, expectedRevision int) (*product.Product, *ProductTemplateSyncDiff, error) {
	if expectedRevision <= 0 {
		return nil, nil, fmt.Errorf("%w: expected revision is required", ErrProductTemplateSyncRevisionConflict)
	}
	diff, err := s.PreviewProductTemplateSync(productID)
	if err != nil {
		return nil, nil, err
	}
	if diff.TemplateRevision != expectedRevision {
		return nil, diff, fmt.Errorf("%w: expected revision %d, current revision %d", ErrProductTemplateSyncRevisionConflict, expectedRevision, diff.TemplateRevision)
	}

	item, err := s.findProduct(productID)
	if err != nil {
		return nil, diff, err
	}
	template, err := s.productRepo.FindProductSpecificationTemplateByID(diff.TemplateID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, diff, ErrProductSpecificationTemplateNotFound
		}
		return nil, diff, err
	}
	inputs := templateSyncOptionValueInputs(item, template)
	updated, err := s.UpdateAdminProduct(productID, ProductUpdateInput{
		VariantOptionValues:       inputs,
		UpdateVariantOptionValues: true,
	})
	if err != nil {
		return nil, diff, err
	}
	updatedDiff, err := s.PreviewProductTemplateSync(productID)
	if err != nil {
		return updated, diff, err
	}
	return updated, updatedDiff, nil
}

func buildProductTemplateSyncDiff(item *product.Product, template *product.ProductSpecificationTemplate) *ProductTemplateSyncDiff {
	diff := &ProductTemplateSyncDiff{
		ProductID:        item.ID,
		TemplateID:       template.ID,
		TemplateRevision: template.Revision,
		Items:            make([]ProductTemplateSyncDiffItem, 0),
	}
	currentRevision := 0
	currentByKey := make(map[string]product.ProductVariantOptionValue, len(item.VariantOptionValues))
	for _, value := range item.VariantOptionValues {
		if value.SourceTemplateRevision > currentRevision {
			currentRevision = value.SourceTemplateRevision
		}
		currentByKey[templateSyncValueKey(value.SpecDefinitionID, value.ValueKey)] = value
	}
	diff.MaterializedRevision = currentRevision
	diff.RevisionChanged = currentRevision != template.Revision

	definitions := make(map[uint]product.SpecDefinition, len(template.SpecDefinitions))
	for _, definition := range template.SpecDefinitions {
		if role := definition.RuntimeRole(); role == "variant" || role == ProductSpecRoleCustomOption {
			definitions[definition.ID] = definition
			for _, candidate := range definition.OptionItems {
				key := templateSyncValueKey(definition.ID, candidate.ValueKey)
				current, exists := currentByKey[key]
				entry := ProductTemplateSyncDiffItem{
					SpecDefinitionID: definition.ID,
					GroupSlug:        definition.Slug,
					GroupName:        definition.Name,
					ValueKey:         candidate.ValueKey,
					TemplateLabel:    templateOptionLabel(candidate),
				}
				templateEnabled := candidate.IsEnabledByDefault
				entry.TemplateEnabled = &templateEnabled
				if candidate.DefaultPriceDeltaMinor != nil {
					value := *candidate.DefaultPriceDeltaMinor
					entry.TemplatePriceDeltaMinor = &value
				}
				if !exists {
					entry.Changes = append(entry.Changes, ProductTemplateSyncChangeAdded)
				} else {
					currentEnabled := current.IsEnabled
					entry.CurrentEnabled = &currentEnabled
					entry.CurrentLabel = current.Label
					if current.Label != templateOptionLabel(candidate) {
						entry.Changes = append(entry.Changes, ProductTemplateSyncChangeLabelChanged)
					}
					if !templateEnabled && currentEnabled {
						entry.Changes = append(entry.Changes, ProductTemplateSyncChangeDisabled)
					} else if templateEnabled && !currentEnabled {
						entry.Changes = append(entry.Changes, ProductTemplateSyncChangeEnabled)
					}
					if definition.RuntimeRole() == ProductSpecRoleCustomOption {
						currentPrice := productOptionPriceDelta(current)
						entry.CurrentPriceDeltaMinor = &currentPrice
						if !int64PointersEqual(entry.TemplatePriceDeltaMinor, entry.CurrentPriceDeltaMinor) {
							entry.Changes = append(entry.Changes, ProductTemplateSyncChangePriceChanged)
						}
					}
				}
				if len(entry.Changes) > 0 {
					diff.Items = append(diff.Items, entry)
					for _, change := range entry.Changes {
						switch change {
						case ProductTemplateSyncChangeAdded:
							diff.Summary.Added++
						case ProductTemplateSyncChangeDisabled:
							diff.Summary.Disabled++
						case ProductTemplateSyncChangeEnabled:
							diff.Summary.Enabled++
						case ProductTemplateSyncChangeLabelChanged:
							diff.Summary.LabelChanged++
						case ProductTemplateSyncChangePriceChanged:
							diff.Summary.PriceChanged++
						}
					}
				}
			}
		}
	}

	for _, current := range item.VariantOptionValues {
		if _, exists := definitions[current.SpecDefinitionID]; !exists {
			continue
		}
		if templateSyncCandidateExists(template, current.SpecDefinitionID, current.ValueKey) {
			continue
		}
		// A removed candidate that was already explicitly synchronized is an
		// archived historical row, not a new pending difference.
		if !current.IsEnabled && current.SourceTemplateRevision == template.Revision {
			continue
		}
		definition := definitions[current.SpecDefinitionID]
		entry := ProductTemplateSyncDiffItem{
			SpecDefinitionID: current.SpecDefinitionID,
			GroupSlug:        definition.Slug,
			GroupName:        definition.Name,
			ValueKey:         current.ValueKey,
			Changes:          []string{ProductTemplateSyncChangeRemoved},
			CurrentLabel:     current.Label,
		}
		currentEnabled := current.IsEnabled
		entry.CurrentEnabled = &currentEnabled
		if definition.RuntimeRole() == ProductSpecRoleCustomOption {
			price := productOptionPriceDelta(current)
			entry.CurrentPriceDeltaMinor = &price
		}
		diff.Items = append(diff.Items, entry)
		diff.Summary.Removed++
	}

	sort.Slice(diff.Items, func(i, j int) bool {
		if diff.Items[i].GroupSlug != diff.Items[j].GroupSlug {
			return diff.Items[i].GroupSlug < diff.Items[j].GroupSlug
		}
		return diff.Items[i].ValueKey < diff.Items[j].ValueKey
	})
	diff.HasChanges = diff.RevisionChanged || len(diff.Items) > 0
	return diff
}

func templateSyncOptionValueInputs(item *product.Product, template *product.ProductSpecificationTemplate) []ProductVariantOptionValueInput {
	currentByKey := make(map[string]product.ProductVariantOptionValue, len(item.VariantOptionValues))
	for _, value := range item.VariantOptionValues {
		currentByKey[templateSyncValueKey(value.SpecDefinitionID, value.ValueKey)] = value
	}
	inputs := make([]ProductVariantOptionValueInput, 0)
	for _, definition := range template.SpecDefinitions {
		role := definition.RuntimeRole()
		if role != "variant" && role != ProductSpecRoleCustomOption {
			continue
		}
		for _, candidate := range definition.OptionItems {
			current, exists := currentByKey[templateSyncValueKey(definition.ID, candidate.ValueKey)]
			itemID := candidate.ID
			enabled := candidate.IsEnabledByDefault
			input := ProductVariantOptionValueInput{
				SpecDefinitionID:       definition.ID,
				TemplateOptionItemID:   &itemID,
				SourceTemplateRevision: template.Revision,
				ValueKey:               candidate.ValueKey,
				Label:                  templateOptionLabel(candidate),
				ColorHex:               candidate.ColorHex,
				SwatchMediaAssetID:     candidate.SwatchMediaAssetID,
				SwatchURL:              candidate.SwatchURL,
				SortOrder:              candidate.SortOrder,
				IsEnabled:              &enabled,
			}
			if exists {
				input.ID = &current.ID
			}
			if role == ProductSpecRoleCustomOption {
				input.PriceDeltaMinor = candidate.DefaultPriceDeltaMinor
				input.IsDefault = candidate.IsDefault
				input.InventoryPolicy = ProductCustomOptionInventoryNone
				// Template revisions own presentation and default pricing only. Keep
				// product-level fulfillment and policy semantics when a value already
				// exists so a catalog sync cannot silently change order behavior.
				if exists && current.CustomOptionPolicy != nil {
					policy := current.CustomOptionPolicy
					input.IsDefault = policy.IsDefault
					input.InventoryPolicy = policy.InventoryPolicy
					input.ComponentVariantID = policy.ComponentVariantID
					input.ComponentQuantity = policy.ComponentQuantity
					input.WeightDeltaGrams = policy.WeightDeltaGrams
					input.PackagingWeightDeltaGrams = policy.PackagingWeightDeltaGrams
					input.ProductionLeadTimeDays = policy.ProductionLeadTimeDays
					input.RequiresProduction = policy.RequiresProduction
					input.CancellationPolicy = policy.CancellationPolicy
					input.ReturnPolicy = policy.ReturnPolicy
				}
			}
			inputs = append(inputs, input)
		}
	}

	// Keep removed candidates as disabled product rows when their definition
	// still exists. This preserves cart/order references without making them
	// purchasable after an explicit sync.
	for _, current := range item.VariantOptionValues {
		definition := templateDefinitionByID(template, current.SpecDefinitionID)
		if definition == nil || templateSyncCandidateExists(template, current.SpecDefinitionID, current.ValueKey) {
			continue
		}
		enabled := false
		input := ProductVariantOptionValueInput{
			ID:                     &current.ID,
			SpecDefinitionID:       current.SpecDefinitionID,
			SourceTemplateRevision: template.Revision,
			ValueKey:               current.ValueKey,
			Label:                  current.Label,
			ColorHex:               current.ColorHex,
			SwatchMediaAssetID:     current.SwatchMediaAssetID,
			SwatchURL:              current.SwatchURL,
			SortOrder:              current.SortOrder,
			IsEnabled:              &enabled,
		}
		if definition.RuntimeRole() == ProductSpecRoleCustomOption {
			price := productOptionPriceDelta(current)
			input.PriceDeltaMinor = &price
			if current.CustomOptionPolicy != nil {
				policy := current.CustomOptionPolicy
				input.IsDefault = policy.IsDefault
				input.InventoryPolicy = policy.InventoryPolicy
				input.ComponentVariantID = policy.ComponentVariantID
				input.ComponentQuantity = policy.ComponentQuantity
				input.WeightDeltaGrams = policy.WeightDeltaGrams
				input.PackagingWeightDeltaGrams = policy.PackagingWeightDeltaGrams
				input.ProductionLeadTimeDays = policy.ProductionLeadTimeDays
				input.RequiresProduction = policy.RequiresProduction
				input.CancellationPolicy = policy.CancellationPolicy
				input.ReturnPolicy = policy.ReturnPolicy
			}
		}
		inputs = append(inputs, input)
	}
	return inputs
}

func templateSyncValueKey(definitionID uint, valueKey string) string {
	return fmt.Sprintf("%d:%s", definitionID, strings.TrimSpace(valueKey))
}

func templateSyncCandidateExists(template *product.ProductSpecificationTemplate, definitionID uint, valueKey string) bool {
	definition := templateDefinitionByID(template, definitionID)
	if definition == nil {
		return false
	}
	for _, candidate := range definition.OptionItems {
		if candidate.ValueKey == valueKey {
			return true
		}
	}
	return false
}

func templateDefinitionByID(template *product.ProductSpecificationTemplate, id uint) *product.SpecDefinition {
	if template == nil {
		return nil
	}
	for index := range template.SpecDefinitions {
		if template.SpecDefinitions[index].ID == id {
			return &template.SpecDefinitions[index]
		}
	}
	return nil
}

func productOptionPriceDelta(value product.ProductVariantOptionValue) int64 {
	if value.CustomOptionPolicy == nil {
		return 0
	}
	return value.CustomOptionPolicy.PriceDeltaMinor
}

func int64PointersEqual(left, right *int64) bool {
	leftValue, rightValue := int64(0), int64(0)
	if left != nil {
		leftValue = *left
	}
	if right != nil {
		rightValue = *right
	}
	return leftValue == rightValue
}

func templateOptionLabel(item product.ProductSpecOptionItem) string {
	label := strings.TrimSpace(item.DefaultLabel)
	if label == "" {
		return strings.TrimSpace(item.ValueKey)
	}
	return label
}
