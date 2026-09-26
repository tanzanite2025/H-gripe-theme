package fitmentcatalog

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrUnknownCassetteSpec   = errors.New("unknown cassette specification")
	ErrUnknownFreehub        = errors.New("unknown freehub standard")
	ErrIncompatibleFreehub   = errors.New("freehub is incompatible with cassette")
	ErrInvalidDrivetrainRule = errors.New("drivetrain rule is invalid")
)

// IsKnownFreehubStandard reports whether standard is one of the physical
// interfaces understood by the fitment matrix. Keeping this allow-list in the
// domain layer prevents an arbitrary client string from being treated as a
// merely incompatible (but otherwise valid) hub choice.
func IsKnownFreehubStandard(standard FreehubStandard) bool {
	switch standard {
	case FreehubStandardHG11,
		FreehubStandardHG,
		FreehubStandardHGL2,
		FreehubStandardXDR,
		FreehubStandardXD,
		FreehubStandardMicroSpline,
		FreehubStandardN3W,
		FreehubStandardCampyClassic:
		return true
	default:
		return false
	}
}

// DrivetrainFitmentEngine performs deterministic lookups over the immutable
// rule set. It intentionally has no repository or product/order dependency.
type DrivetrainFitmentEngine struct {
	rules      []CassetteFitmentRule
	rulesIndex map[string]int
}

// NewDrivetrainFitmentEngine validates and indexes a rule set.
func NewDrivetrainFitmentEngine(rules []CassetteFitmentRule) (*DrivetrainFitmentEngine, error) {
	engine := &DrivetrainFitmentEngine{
		rules:      make([]CassetteFitmentRule, len(rules)),
		rulesIndex: make(map[string]int, len(rules)),
	}
	copy(engine.rules, rules)

	for index := range engine.rules {
		rule := &engine.rules[index]
		if err := validateDrivetrainRule(rule); err != nil {
			return nil, fmt.Errorf("%w: rule %q: %w", ErrInvalidDrivetrainRule, rule.RuleID, err)
		}
		key := drivetrainRuleKey(rule.Brand, rule.CassetteSpec)
		if _, exists := engine.rulesIndex[key]; exists {
			return nil, fmt.Errorf("%w: duplicate brand/spec %q", ErrInvalidDrivetrainRule, key)
		}
		engine.rulesIndex[key] = index
	}

	return engine, nil
}

// NewDefaultDrivetrainFitmentEngine creates the application rule engine.
func NewDefaultDrivetrainFitmentEngine() *DrivetrainFitmentEngine {
	engine, err := NewDrivetrainFitmentEngine(DefaultDrivetrainRules())
	if err != nil {
		// The checked-in rule set is a program invariant. Panicking here prevents
		// the application from serving silently incomplete mechanical guidance.
		panic(err)
	}
	return engine
}

// CalculateByCassette returns the authoritative rule for a brand/spec pair.
// Unknown inputs fail loudly and are never mapped to a generic HG fallback.
func (e *DrivetrainFitmentEngine) CalculateByCassette(brand, cassetteSpec string) (*CassetteFitmentRule, error) {
	if e == nil {
		return nil, fmt.Errorf("%w: engine is nil", ErrInvalidDrivetrainRule)
	}
	index, exists := e.rulesIndex[drivetrainRuleKey(brand, cassetteSpec)]
	if !exists {
		return nil, fmt.Errorf("%w: [CRITICAL] unknown cassette spec %q for brand %q", ErrUnknownCassetteSpec, strings.TrimSpace(cassetteSpec), strings.TrimSpace(brand))
	}

	rule := e.rules[index]
	if err := validatePhysicalInterference(&rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// CalculateByCassetteAndFreehub validates a caller's actual hub choice. The
// optional freehub input is what makes the physical-interference block
// testable; CalculateByCassette remains useful for recommendation-only UIs.
func (e *DrivetrainFitmentEngine) CalculateByCassetteAndFreehub(brand, cassetteSpec string, freehub FreehubStandard) (*FreehubFitmentOption, *CassetteFitmentRule, error) {
	rule, err := e.CalculateByCassette(brand, cassetteSpec)
	if err != nil {
		return nil, nil, err
	}
	standard := FreehubStandard(strings.ToUpper(strings.TrimSpace(string(freehub))))
	if !IsKnownFreehubStandard(standard) {
		if standard == "" {
			return nil, rule, fmt.Errorf("%w: freehub standard is required", ErrUnknownFreehub)
		}
		return nil, rule, fmt.Errorf("%w: unsupported freehub standard %q", ErrUnknownFreehub, standard)
	}
	// Check the diameter constraint before looking for a declared option so a
	// caller forcing a 10T cassette onto HG receives the physical explanation,
	// rather than a generic "option not found" response.
	if err := validatePhysicalInterferenceForFreehub(rule, standard); err != nil {
		return nil, rule, err
	}
	for index := range rule.FitmentOptions {
		option := rule.FitmentOptions[index]
		if option.Standard != standard {
			continue
		}
		return &option, rule, nil
	}
	return nil, rule, fmt.Errorf("%w: [CRITICAL] %s cannot accept %s", ErrIncompatibleFreehub, rule.DisplayName, standard)
}

// Matrix returns a copy of the full rule matrix for SSR and API consumers.
func (e *DrivetrainFitmentEngine) Matrix() []CassetteFitmentRule {
	if e == nil {
		return nil
	}
	matrix := make([]CassetteFitmentRule, len(e.rules))
	copy(matrix, e.rules)
	return matrix
}

func drivetrainRuleKey(brand, cassetteSpec string) string {
	return strings.ToUpper(strings.TrimSpace(brand)) + ":" + strings.ToLower(strings.TrimSpace(cassetteSpec))
}

func validateDrivetrainRule(rule *CassetteFitmentRule) error {
	if rule == nil {
		return errors.New("rule is nil")
	}
	if strings.TrimSpace(rule.RuleID) == "" {
		return errors.New("rule_id is required")
	}
	if strings.TrimSpace(rule.Brand) == "" {
		return errors.New("brand is required")
	}
	if strings.TrimSpace(rule.CassetteSpec) == "" {
		return errors.New("cassette_spec is required")
	}
	if rule.Speed <= 0 || rule.MinCogTeeth <= 0 || rule.MaxCogTeeth < rule.MinCogTeeth {
		return errors.New("speed and cog range must be positive")
	}
	if strings.TrimSpace(rule.RuleVersion) == "" {
		return errors.New("rule_version is required")
	}
	if len(rule.FitmentOptions) == 0 {
		return errors.New("at least one fitment option is required")
	}
	if rule.RecommendedFreehub == "" {
		return errors.New("recommended_freehub is required")
	}
	if !IsKnownFreehubStandard(rule.RecommendedFreehub) {
		return fmt.Errorf("%w: recommended_freehub %q", ErrUnknownFreehub, rule.RecommendedFreehub)
	}
	foundRecommended := false
	for index := range rule.FitmentOptions {
		option := &rule.FitmentOptions[index]
		if option.Standard == "" || strings.TrimSpace(option.DisplayName) == "" {
			return fmt.Errorf("fitment option %d is missing standard or display name", index)
		}
		if !IsKnownFreehubStandard(option.Standard) {
			return fmt.Errorf("%w: fitment option %d standard %q", ErrUnknownFreehub, index, option.Standard)
		}
		if option.Standard == rule.RecommendedFreehub {
			foundRecommended = true
		}
		if err := validateSpacerRequirement(option.Spacer); err != nil {
			return fmt.Errorf("fitment option %d spacer: %v", index, err)
		}
	}
	if !foundRecommended {
		return fmt.Errorf("recommended freehub %q is not present in fitment options", rule.RecommendedFreehub)
	}
	if err := validatePhysicalInterference(rule); err != nil {
		return err
	}
	return nil
}

func validateSpacerRequirement(requirement SpacerRequirement) error {
	if !requirement.Required {
		if requirement.ThicknessMM != 0 || len(requirement.Parts) != 0 {
			return errors.New("optional spacer cannot have thickness or parts")
		}
		return nil
	}
	if requirement.ThicknessMM <= 0 || math.IsNaN(requirement.ThicknessMM) || math.IsInf(requirement.ThicknessMM, 0) {
		return errors.New("required spacer thickness must be finite and positive")
	}
	if len(requirement.Parts) == 0 {
		return nil
	}
	var total float64
	for _, part := range requirement.Parts {
		if part.ThicknessMM <= 0 || math.IsNaN(part.ThicknessMM) || math.IsInf(part.ThicknessMM, 0) {
			return errors.New("spacer part thickness must be finite and positive")
		}
		total += part.ThicknessMM
	}
	if math.Abs(total-requirement.ThicknessMM) > 0.0001 {
		return fmt.Errorf("spacer parts total %.4f does not equal thickness %.4f", total, requirement.ThicknessMM)
	}
	return nil
}

func validatePhysicalInterference(rule *CassetteFitmentRule) error {
	if rule == nil {
		return fmt.Errorf("%w: rule is nil", ErrInvalidDrivetrainRule)
	}
	for _, option := range rule.FitmentOptions {
		if err := validatePhysicalInterferenceForFreehub(rule, option.Standard); err != nil {
			// An invalid option in the authoritative matrix must never be hidden by
			// another valid option. Return it as a critical data error.
			return err
		}
	}
	return nil
}

func validatePhysicalInterferenceForFreehub(rule *CassetteFitmentRule, standard FreehubStandard) error {
	if rule.MinCogTeeth <= 10 && (standard == FreehubStandardHG || standard == FreehubStandardHG11) {
		return fmt.Errorf("%w: [CRITICAL] physical interference: %dT cog root diameter (%.2fmm) < HG outer diameter (34.80mm)", ErrIncompatibleFreehub, rule.MinCogTeeth, calculateRootDiameter(rule.MinCogTeeth))
	}
	return nil
}

func calculatePitchDiameter(teeth int) float64 {
	if teeth <= 0 {
		return 0
	}
	return 12.7 / math.Sin(math.Pi/float64(teeth))
}

func calculateRootDiameter(teeth int) float64 {
	if teeth <= 0 {
		return 0
	}
	return calculatePitchDiameter(teeth) - 7.94
}
