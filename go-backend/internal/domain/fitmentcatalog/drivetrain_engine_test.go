package fitmentcatalog

import (
	"errors"
	"strings"
	"testing"
)

func TestDefaultDrivetrainEngineBuildsCompleteMatrix(t *testing.T) {
	engine := NewDefaultDrivetrainFitmentEngine()
	matrix := engine.Matrix()
	if len(matrix) < 20 {
		t.Fatalf("Matrix() length = %d, want at least 20 rules", len(matrix))
	}

	for _, rule := range matrix {
		if rule.RuleVersion != "v1.0" {
			t.Fatalf("rule %q has RuleVersion %q, want v1.0", rule.RuleID, rule.RuleVersion)
		}
		if len(rule.FitmentOptions) == 0 {
			t.Fatalf("rule %q has no fitment options", rule.RuleID)
		}
	}
}

func TestDrivetrainEngineRejectsUnknownCassetteWithoutFallback(t *testing.T) {
	engine := NewDefaultDrivetrainFitmentEngine()

	_, err := engine.CalculateByCassette("SRAM", "sram_12s_unknown")
	if !errors.Is(err, ErrUnknownCassetteSpec) {
		t.Fatalf("CalculateByCassette() error = %v, want ErrUnknownCassetteSpec", err)
	}
	if !strings.Contains(err.Error(), "[CRITICAL]") {
		t.Fatalf("CalculateByCassette() error = %v, want critical marker", err)
	}
}

func TestDrivetrainEngineBlocksTenToothCassetteOnHG(t *testing.T) {
	engine := NewDefaultDrivetrainFitmentEngine()

	_, _, err := engine.CalculateByCassetteAndFreehub(
		"SRAM",
		"sram_12s_10_52t",
		FreehubStandardHG11,
	)
	if !errors.Is(err, ErrIncompatibleFreehub) {
		t.Fatalf("CalculateByCassetteAndFreehub() error = %v, want ErrIncompatibleFreehub", err)
	}
	if !strings.Contains(err.Error(), "physical interference") || !strings.Contains(err.Error(), "33.16mm") {
		t.Fatalf("CalculateByCassetteAndFreehub() error = %v, want 10T physical interference details", err)
	}
}

func TestDrivetrainEngineRejectsUnknownFreehubStandard(t *testing.T) {
	engine := NewDefaultDrivetrainFitmentEngine()

	_, _, err := engine.CalculateByCassetteAndFreehub(
		"SRAM",
		"sram_12s_10_52t",
		FreehubStandard("mystery-freehub"),
	)
	if !errors.Is(err, ErrUnknownFreehub) {
		t.Fatalf("CalculateByCassetteAndFreehub() error = %v, want ErrUnknownFreehub", err)
	}
	if errors.Is(err, ErrIncompatibleFreehub) {
		t.Fatalf("CalculateByCassetteAndFreehub() error = %v, unknown standard must not be reported as incompatibility", err)
	}
}

func TestDrivetrainEngineAllowsShimanoRoadTwelveSpeedOnHGWithoutSpacer(t *testing.T) {
	engine := NewDefaultDrivetrainFitmentEngine()

	option, rule, err := engine.CalculateByCassetteAndFreehub(
		"shimano",
		"shimano_12s_road_11_34t",
		FreehubStandardHG11,
	)
	if err != nil {
		t.Fatalf("CalculateByCassetteAndFreehub() error = %v", err)
	}
	if rule.RecommendedFreehub != FreehubStandardHG11 {
		t.Fatalf("RecommendedFreehub = %q, want %q", rule.RecommendedFreehub, FreehubStandardHG11)
	}
	if option.Spacer.Required || option.Spacer.ThicknessMM != 0 {
		t.Fatalf("Spacer = %+v, want no spacer", option.Spacer)
	}

	if _, _, err := engine.CalculateByCassetteAndFreehub(
		"Shimano",
		"shimano_12s_road_11_34t",
		FreehubStandardHGL2,
	); err != nil {
		t.Fatalf("HG L2 compatibility error = %v", err)
	}
}

func TestDrivetrainEngineReturnsXDRSpacerForXDOrXDRCassette(t *testing.T) {
	engine := NewDefaultDrivetrainFitmentEngine()

	option, _, err := engine.CalculateByCassetteAndFreehub(
		"SRAM",
		"sram_12s_10_52t",
		FreehubStandardXDR,
	)
	if err != nil {
		t.Fatalf("CalculateByCassetteAndFreehub() error = %v", err)
	}
	if option.Spacer.ThicknessMM != 1.85 || option.Spacer.Position != spacerPositionInnerBase {
		t.Fatalf("Spacer = %+v, want 1.85mm inner-base spacer", option.Spacer)
	}

	option, _, err = engine.CalculateByCassetteAndFreehub(
		"SRAM",
		"sram_12s_10_52t",
		FreehubStandardXD,
	)
	if err != nil {
		t.Fatalf("XD compatibility error = %v", err)
	}
	if option.Spacer.Required {
		t.Fatalf("XD Spacer = %+v, want no spacer", option.Spacer)
	}
}

func TestDrivetrainEngineReturnsCampagnoloAdapterForTraditionalCassetteOnN3W(t *testing.T) {
	engine := NewDefaultDrivetrainFitmentEngine()

	option, _, err := engine.CalculateByCassetteAndFreehub(
		"Campagnolo",
		"campagnolo_11_12s_11_34t",
		FreehubStandardN3W,
	)
	if err != nil {
		t.Fatalf("CalculateByCassetteAndFreehub() error = %v", err)
	}
	if option.Spacer.PartCode != "AC21-N3W" || option.Spacer.ThicknessMM != 4.4 {
		t.Fatalf("Spacer = %+v, want AC21-N3W 4.4mm adapter", option.Spacer)
	}
	if option.Spacer.Position != spacerPositionAdapterSleeve || len(option.Spacer.Parts) != 1 {
		t.Fatalf("Spacer = %+v, want one adapter sleeve part", option.Spacer)
	}
}

func TestDrivetrainEnginePreservesCompoundShimanoSpacerParts(t *testing.T) {
	engine := NewDefaultDrivetrainFitmentEngine()

	option, _, err := engine.CalculateByCassetteAndFreehub(
		"Shimano",
		"shimano_10s_road_11_30t",
		FreehubStandardHG11,
	)
	if err != nil {
		t.Fatalf("CalculateByCassetteAndFreehub() error = %v", err)
	}
	if option.Spacer.ThicknessMM != 2.85 || len(option.Spacer.Parts) != 2 {
		t.Fatalf("Spacer = %+v, want two parts totalling 2.85mm", option.Spacer)
	}
	if option.Spacer.Parts[0].ThicknessMM != 1.85 || option.Spacer.Parts[1].ThicknessMM != 1.0 {
		t.Fatalf("Spacer parts = %+v, want 1.85mm + 1.0mm", option.Spacer.Parts)
	}
}

func TestDrivetrainEngineRejectsInvalidRuleAtConstruction(t *testing.T) {
	_, err := NewDrivetrainFitmentEngine([]CassetteFitmentRule{{
		RuleID:             "bad",
		Brand:              "Shimano",
		CassetteSpec:       "bad",
		Speed:              12,
		MinCogTeeth:        10,
		MaxCogTeeth:        50,
		RecommendedFreehub: FreehubStandardHG11,
		RuleVersion:        "v1.0",
		FitmentOptions: []FreehubFitmentOption{{
			Standard:    FreehubStandardHG11,
			DisplayName: "HG-11",
			Spacer:      noSpacer(),
		}},
	}})
	if !errors.Is(err, ErrInvalidDrivetrainRule) {
		t.Fatalf("NewDrivetrainFitmentEngine() error = %v, want ErrInvalidDrivetrainRule", err)
	}
}

func TestDrivetrainEngineRejectsUnknownFreehubInRule(t *testing.T) {
	_, err := NewDrivetrainFitmentEngine([]CassetteFitmentRule{{
		RuleID:             "bad-freehub",
		Brand:              "Shimano",
		CassetteSpec:       "bad-freehub",
		Speed:              12,
		MinCogTeeth:        11,
		MaxCogTeeth:        34,
		RecommendedFreehub: FreehubStandard("mystery-freehub"),
		RuleVersion:        "v1.0",
		FitmentOptions: []FreehubFitmentOption{{
			Standard:    FreehubStandard("mystery-freehub"),
			DisplayName: "Mystery",
			Spacer:      noSpacer(),
		}},
	}})
	if !errors.Is(err, ErrInvalidDrivetrainRule) {
		t.Fatalf("NewDrivetrainFitmentEngine() error = %v, want ErrInvalidDrivetrainRule", err)
	}
	if !errors.Is(err, ErrUnknownFreehub) {
		t.Fatalf("NewDrivetrainFitmentEngine() error = %v, want ErrUnknownFreehub cause", err)
	}
}

func TestDrivetrainEnginePhysicalDiameterFormula(t *testing.T) {
	tests := []struct {
		teeth int
		want  float64
	}{
		{teeth: 11, want: 37.14},
		{teeth: 10, want: 33.16},
		{teeth: 9, want: 29.19},
	}

	for _, test := range tests {
		got := calculateRootDiameter(test.teeth)
		if difference := got - test.want; difference > 0.01 || difference < -0.01 {
			t.Errorf("calculateRootDiameter(%d) = %.4f, want %.2f", test.teeth, got, test.want)
		}
	}
}
