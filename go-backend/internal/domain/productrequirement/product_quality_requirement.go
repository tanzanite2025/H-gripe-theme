package productrequirement

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	RequirementTypeSpokeTensionQC = "spoke_tension_qc"

	RuleStatusActive   = "active"
	RuleStatusInactive = "inactive"

	ResolutionSourceNone    = "none"
	ResolutionSourceProduct = "product"
	ResolutionSourceVariant = "variant"
)

// ProductQualityRequirementRule is an explicit fulfillment requirement for a
// product or one of its variants. A nil VariantID means the product default.
type ProductQualityRequirementRule struct {
	ID                     uint      `gorm:"primaryKey" json:"id"`
	ProductID              uint      `gorm:"not null;index" json:"product_id"`
	VariantID              *uint     `gorm:"index" json:"variant_id,omitempty"`
	RequirementType        string    `gorm:"size:64;not null;index" json:"requirement_type"`
	SpokeTensionQCRequired bool      `gorm:"not null;default:false" json:"spoke_tension_qc_required"`
	Status                 string    `gorm:"size:16;not null;default:'active';index" json:"status"`
	RuleVersion            string    `gorm:"size:64;not null" json:"rule_version"`
	Reason                 string    `gorm:"type:text;not null;default:''" json:"reason"`
	CreatedBy              uint      `gorm:"not null;index" json:"created_by"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (ProductQualityRequirementRule) TableName() string {
	return "product_quality_requirement_rules"
}

func (r ProductQualityRequirementRule) Validate() error {
	if r.ProductID == 0 {
		return errors.New("product quality requirement product_id is required")
	}
	if r.VariantID != nil && *r.VariantID == 0 {
		return errors.New("product quality requirement variant_id must be greater than zero")
	}
	if normalizeRuleValue(r.RequirementType) != RequirementTypeSpokeTensionQC {
		return fmt.Errorf("unsupported product quality requirement type %q", r.RequirementType)
	}
	switch normalizeRuleValue(r.Status) {
	case RuleStatusActive, RuleStatusInactive:
	default:
		return fmt.Errorf("unsupported product quality requirement status %q", r.Status)
	}
	if strings.TrimSpace(r.RuleVersion) == "" {
		return errors.New("product quality requirement rule_version is required")
	}
	return nil
}

func (r *ProductQualityRequirementRule) Normalize() {
	if r == nil {
		return
	}
	r.RequirementType = normalizeRuleValue(r.RequirementType)
	r.Status = normalizeRuleValue(r.Status)
	r.RuleVersion = strings.TrimSpace(r.RuleVersion)
	r.Reason = strings.TrimSpace(r.Reason)
}

func (r *ProductQualityRequirementRule) BeforeCreate(tx *gorm.DB) error {
	r.Normalize()
	return r.Validate()
}

func (r *ProductQualityRequirementRule) BeforeSave(tx *gorm.DB) error {
	r.Normalize()
	return r.Validate()
}

func (r ProductQualityRequirementRule) IsActive() bool {
	return normalizeRuleValue(r.Status) == RuleStatusActive
}

// SpokeTensionQCResolution is the order-creation input for the future
// immutable product-requirement snapshot. It intentionally contains only the
// selected rule and its result, never live product metadata.
type SpokeTensionQCResolution struct {
	Required        bool   `json:"required"`
	Matched         bool   `json:"matched"`
	RuleID          *uint  `json:"rule_id,omitempty"`
	RuleVersion     string `json:"rule_version,omitempty"`
	Reason          string `json:"reason"`
	Source          string `json:"source"`
	RequirementType string `json:"requirement_type"`
}

// ResolveSpokeTensionQC selects one explicit rule. Variant rules take
// precedence over product defaults; inactive and unrelated rules are ignored.
// Multiple active rules at the same specificity are rejected as ambiguous so
// behavior cannot depend on database or slice ordering.
func ResolveSpokeTensionQC(
	productID uint,
	variantID *uint,
	rules []ProductQualityRequirementRule,
) (SpokeTensionQCResolution, error) {
	if productID == 0 {
		return SpokeTensionQCResolution{}, errors.New("product_id is required")
	}

	variantMatches := make([]ProductQualityRequirementRule, 0, 1)
	productMatches := make([]ProductQualityRequirementRule, 0, 1)

	for _, rule := range rules {
		if rule.ProductID != productID ||
			normalizeRuleValue(rule.RequirementType) != RequirementTypeSpokeTensionQC ||
			!rule.IsActive() {
			continue
		}
		if variantID != nil && rule.VariantID != nil && *rule.VariantID == *variantID {
			variantMatches = append(variantMatches, rule)
			continue
		}
		if rule.VariantID == nil {
			productMatches = append(productMatches, rule)
		}
	}

	switch {
	case len(variantMatches) > 1:
		return SpokeTensionQCResolution{}, fmt.Errorf(
			"multiple active spoke tension QC rules match product %d variant %d",
			productID,
			*variantID,
		)
	case len(variantMatches) == 1:
		return resolutionFromRule(variantMatches[0], ResolutionSourceVariant), nil
	case len(productMatches) > 1:
		return SpokeTensionQCResolution{}, fmt.Errorf(
			"multiple active default spoke tension QC rules match product %d",
			productID,
		)
	case len(productMatches) == 1:
		return resolutionFromRule(productMatches[0], ResolutionSourceProduct), nil
	default:
		return SpokeTensionQCResolution{
			Required:        false,
			Matched:         false,
			Reason:          "no_active_rule",
			Source:          ResolutionSourceNone,
			RequirementType: RequirementTypeSpokeTensionQC,
		}, nil
	}
}

func normalizeRuleValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func resolutionFromRule(rule ProductQualityRequirementRule, source string) SpokeTensionQCResolution {
	ruleID := rule.ID
	return SpokeTensionQCResolution{
		Required:        rule.SpokeTensionQCRequired,
		Matched:         true,
		RuleID:          &ruleID,
		RuleVersion:     strings.TrimSpace(rule.RuleVersion),
		Reason:          strings.TrimSpace(rule.Reason),
		Source:          source,
		RequirementType: RequirementTypeSpokeTensionQC,
	}
}
