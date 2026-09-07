package productrequirement

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductQualityRequirementRuleValidate(t *testing.T) {
	variantID := uint(22)
	tests := []struct {
		name    string
		rule    ProductQualityRequirementRule
		wantErr string
	}{
		{
			name: "valid product default rule",
			rule: ProductQualityRequirementRule{
				ProductID:       10,
				RequirementType: RequirementTypeSpokeTensionQC,
				Status:          RuleStatusActive,
				RuleVersion:     "2026-09-04-1",
			},
		},
		{
			name: "valid variant rule",
			rule: ProductQualityRequirementRule{
				ProductID:       10,
				VariantID:       &variantID,
				RequirementType: RequirementTypeSpokeTensionQC,
				Status:          RuleStatusInactive,
				RuleVersion:     "2026-09-04-2",
			},
		},
		{
			name: "missing product",
			rule: ProductQualityRequirementRule{
				RequirementType: RequirementTypeSpokeTensionQC,
				Status:          RuleStatusActive,
				RuleVersion:     "2026-09-04-1",
			},
			wantErr: "product_id is required",
		},
		{
			name: "zero variant",
			rule: ProductQualityRequirementRule{
				ProductID:       10,
				VariantID:       new(uint),
				RequirementType: RequirementTypeSpokeTensionQC,
				Status:          RuleStatusActive,
				RuleVersion:     "2026-09-04-1",
			},
			wantErr: "variant_id must be greater than zero",
		},
		{
			name: "unsupported type",
			rule: ProductQualityRequirementRule{
				ProductID:       10,
				RequirementType: "wheelset",
				Status:          RuleStatusActive,
				RuleVersion:     "2026-09-04-1",
			},
			wantErr: "unsupported product quality requirement type",
		},
		{
			name: "unsupported status",
			rule: ProductQualityRequirementRule{
				ProductID:       10,
				RequirementType: RequirementTypeSpokeTensionQC,
				Status:          "deleted",
				RuleVersion:     "2026-09-04-1",
			},
			wantErr: "unsupported product quality requirement status",
		},
		{
			name: "missing version",
			rule: ProductQualityRequirementRule{
				ProductID:       10,
				RequirementType: RequirementTypeSpokeTensionQC,
				Status:          RuleStatusActive,
			},
			wantErr: "rule_version is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestProductQualityRequirementRuleNormalize(t *testing.T) {
	rule := ProductQualityRequirementRule{
		ProductID:       10,
		RequirementType: " SPOKE_TENSION_QC ",
		Status:          " ACTIVE ",
		RuleVersion:     " version-1 ",
		Reason:          " product default ",
	}

	rule.Normalize()

	assert.Equal(t, RequirementTypeSpokeTensionQC, rule.RequirementType)
	assert.Equal(t, RuleStatusActive, rule.Status)
	assert.Equal(t, "version-1", rule.RuleVersion)
	assert.Equal(t, "product default", rule.Reason)
	require.NoError(t, rule.Validate())
}

func TestResolveSpokeTensionQCUsesVariantThenProductDefault(t *testing.T) {
	variantID := uint(22)
	tests := []struct {
		name            string
		inputVariantID  *uint
		rules           []ProductQualityRequirementRule
		wantRequired    bool
		wantMatched     bool
		wantSource      string
		wantRuleID      uint
		wantRuleVersion string
		wantReason      string
	}{
		{
			name:           "exact variant rule overrides product default",
			inputVariantID: &variantID,
			rules: []ProductQualityRequirementRule{
				testRule(101, 10, nil, false, RuleStatusActive, "product-v1", "product default"),
				testRule(102, 10, &variantID, true, RuleStatusActive, "variant-v1", "variant override"),
			},
			wantRequired:    true,
			wantMatched:     true,
			wantSource:      ResolutionSourceVariant,
			wantRuleID:      102,
			wantRuleVersion: "variant-v1",
			wantReason:      "variant override",
		},
		{
			name:           "exact variant rule can explicitly disable product default",
			inputVariantID: &variantID,
			rules: []ProductQualityRequirementRule{
				testRule(103, 10, nil, true, RuleStatusActive, "product-v1", "all variants require QC"),
				testRule(104, 10, &variantID, false, RuleStatusActive, "variant-v2", "variant does not require QC"),
			},
			wantRequired:    false,
			wantMatched:     true,
			wantSource:      ResolutionSourceVariant,
			wantRuleID:      104,
			wantRuleVersion: "variant-v2",
			wantReason:      "variant does not require QC",
		},
		{
			name:           "product default applies when variant has no override",
			inputVariantID: &variantID,
			rules: []ProductQualityRequirementRule{
				testRule(105, 10, nil, true, RuleStatusActive, "product-v3", "product default"),
			},
			wantRequired:    true,
			wantMatched:     true,
			wantSource:      ResolutionSourceProduct,
			wantRuleID:      105,
			wantRuleVersion: "product-v3",
			wantReason:      "product default",
		},
		{
			name: "product default applies without a selected variant",
			rules: []ProductQualityRequirementRule{
				testRule(106, 10, nil, true, RuleStatusActive, "product-v4", "product default"),
			},
			wantRequired:    true,
			wantMatched:     true,
			wantSource:      ResolutionSourceProduct,
			wantRuleID:      106,
			wantRuleVersion: "product-v4",
			wantReason:      "product default",
		},
		{
			name:           "inactive exact rule does not override active default",
			inputVariantID: &variantID,
			rules: []ProductQualityRequirementRule{
				testRule(107, 10, nil, true, RuleStatusActive, "product-v5", "product default"),
				testRule(108, 10, &variantID, false, RuleStatusInactive, "variant-v3", "inactive override"),
			},
			wantRequired:    true,
			wantMatched:     true,
			wantSource:      ResolutionSourceProduct,
			wantRuleID:      107,
			wantRuleVersion: "product-v5",
			wantReason:      "product default",
		},
		{
			name:           "unmatched product resolves to no requirement",
			inputVariantID: &variantID,
			rules: []ProductQualityRequirementRule{
				testRule(109, 99, nil, true, RuleStatusActive, "other-product-v1", "other product"),
				testRuleWithType(110, 10, nil, true, RuleStatusActive, "other_type", "other-type-v1", "other type"),
			},
			wantRequired: false,
			wantMatched:  false,
			wantSource:   ResolutionSourceNone,
			wantReason:   "no_active_rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolution, err := ResolveSpokeTensionQC(10, tt.inputVariantID, tt.rules)
			require.NoError(t, err)
			assert.Equal(t, tt.wantRequired, resolution.Required)
			assert.Equal(t, tt.wantMatched, resolution.Matched)
			assert.Equal(t, tt.wantSource, resolution.Source)
			assert.Equal(t, tt.wantReason, resolution.Reason)
			assert.Equal(t, RequirementTypeSpokeTensionQC, resolution.RequirementType)
			if tt.wantRuleID == 0 {
				assert.Nil(t, resolution.RuleID)
				assert.Empty(t, resolution.RuleVersion)
				return
			}
			require.NotNil(t, resolution.RuleID)
			assert.Equal(t, tt.wantRuleID, *resolution.RuleID)
			assert.Equal(t, tt.wantRuleVersion, resolution.RuleVersion)
		})
	}
}

func TestResolveSpokeTensionQCRejectsAmbiguousActiveRules(t *testing.T) {
	variantID := uint(22)
	_, err := ResolveSpokeTensionQC(10, &variantID, []ProductQualityRequirementRule{
		testRule(201, 10, &variantID, true, RuleStatusActive, "variant-v1", "first"),
		testRule(202, 10, &variantID, false, RuleStatusActive, "variant-v2", "second"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple active spoke tension QC rules")

	_, err = ResolveSpokeTensionQC(10, nil, []ProductQualityRequirementRule{
		testRule(203, 10, nil, true, RuleStatusActive, "product-v1", "first"),
		testRule(204, 10, nil, false, RuleStatusActive, "product-v2", "second"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple active default spoke tension QC rules")
}

func TestResolveSpokeTensionQCRequiresProductID(t *testing.T) {
	_, err := ResolveSpokeTensionQC(0, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "product_id is required")
}

func testRule(
	id uint,
	productID uint,
	variantID *uint,
	required bool,
	status string,
	version string,
	reason string,
) ProductQualityRequirementRule {
	return testRuleWithType(
		id,
		productID,
		variantID,
		required,
		status,
		RequirementTypeSpokeTensionQC,
		version,
		reason,
	)
}

func testRuleWithType(
	id uint,
	productID uint,
	variantID *uint,
	required bool,
	status string,
	requirementType string,
	version string,
	reason string,
) ProductQualityRequirementRule {
	return ProductQualityRequirementRule{
		ID:                     id,
		ProductID:              productID,
		VariantID:              variantID,
		RequirementType:        requirementType,
		SpokeTensionQCRequired: required,
		Status:                 status,
		RuleVersion:            version,
		Reason:                 reason,
	}
}
