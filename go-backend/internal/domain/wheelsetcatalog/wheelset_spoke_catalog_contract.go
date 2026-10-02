package wheelsetcatalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// The wheelset directory is shared by the public reference endpoint and the
// repair-kit product workflow. Keeping one embedded source prevents the admin
// selector from accepting model keys that are not present in the published
// directory.
//
//go:embed brand-wheelset-spoke-specs-catalog.json
var catalogJSON []byte

type Brand struct {
	BrandSlug         string     `json:"brandSlug"`
	BrandName         string     `json:"brandName"`
	PublicationStatus string     `json:"publicationStatus"`
	SourceCheckedAt   *string    `json:"sourceCheckedAt"`
	Wheelsets         []Wheelset `json:"wheelsets"`
}

type Wheelset struct {
	Slug               string   `json:"slug"`
	Model              string   `json:"model"`
	LifecycleStatus    string   `json:"lifecycleStatus"`
	VerificationStatus string   `json:"verificationStatus"`
	Rim                Rim      `json:"rim"`
	NippleModel        string   `json:"nippleModel"`
	NippleLengthMM     *float64 `json:"nippleLengthMm"`
	Wheels             []Wheel  `json:"wheels"`
}

type Rim struct {
	DepthMM      float64  `json:"depthMm"`
	DepthFrontMM *float64 `json:"depthFrontMm"`
	DepthRearMM  *float64 `json:"depthRearMm"`
	InnerWidthMM float64  `json:"innerWidthMm"`
	OuterWidthMM float64  `json:"outerWidthMm"`
}

type Wheel struct {
	Position      string `json:"position"`
	SpokeCount    int    `json:"spokeCount"`
	LacingPattern string `json:"lacingPattern"`
	Sides         []Side `json:"sides"`
}

type Side struct {
	Side       string   `json:"side"`
	LengthMM   *float64 `json:"lengthMm"`
	SpokeModel string   `json:"spokeModel"`
	HeadType   string   `json:"headType"`
}

type Model struct {
	BrandSlug       string  `json:"brandSlug"`
	BrandName       string  `json:"brandName"`
	Slug            string  `json:"slug"`
	Model           string  `json:"model"`
	LifecycleStatus string  `json:"lifecycleStatus"`
	SourceCheckedAt *string `json:"-"`
}

func LoadWheelsetSpokeCatalogBrands() ([]Brand, error) {
	var brands []Brand
	if err := json.Unmarshal(catalogJSON, &brands); err != nil {
		return nil, err
	}
	if err := validateWheelsetCatalogIdentifiers(brands); err != nil {
		return nil, err
	}
	return brands, nil
}

// validateWheelsetCatalogIdentifiers keeps the backend lookup key unambiguous.
// The storefront source performs the same check, but the embedded JSON is the
// backend runtime authority and must defend itself when it is edited or copied.
func validateWheelsetCatalogIdentifiers(brands []Brand) error {
	brandSlugs := make(map[string]struct{}, len(brands))
	wheelsetSlugs := make(map[string]struct{})
	for brandIndex, brand := range brands {
		brandSlug := strings.ToLower(strings.TrimSpace(brand.BrandSlug))
		if brandSlug == "" {
			return fmt.Errorf("wheelset catalog brand %d has an empty slug", brandIndex)
		}
		if _, exists := brandSlugs[brandSlug]; exists {
			return fmt.Errorf("wheelset catalog has duplicate brand slug %q", brandSlug)
		}
		brandSlugs[brandSlug] = struct{}{}

		for wheelsetIndex, wheelset := range brand.Wheelsets {
			wheelsetSlug := strings.ToLower(strings.TrimSpace(wheelset.Slug))
			if wheelsetSlug == "" {
				return fmt.Errorf("wheelset catalog brand %q model %d has an empty slug", brandSlug, wheelsetIndex)
			}
			if _, exists := wheelsetSlugs[wheelsetSlug]; exists {
				return fmt.Errorf("wheelset catalog has duplicate model slug %q", wheelsetSlug)
			}
			wheelsetSlugs[wheelsetSlug] = struct{}{}
		}
	}
	return nil
}

func ListPublishedWheelsetSpokeCatalogModels() ([]Model, error) {
	brands, err := LoadWheelsetSpokeCatalogBrands()
	if err != nil {
		return nil, err
	}
	models := make([]Model, 0)
	for _, brand := range brands {
		if brand.PublicationStatus != "published" {
			continue
		}
		for _, wheelset := range brand.Wheelsets {
			if wheelset.VerificationStatus != "verified" {
				continue
			}
			models = append(models, Model{
				BrandSlug:       brand.BrandSlug,
				BrandName:       brand.BrandName,
				Slug:            wheelset.Slug,
				Model:           wheelset.Model,
				LifecycleStatus: wheelset.LifecycleStatus,
				SourceCheckedAt: brand.SourceCheckedAt,
			})
		}
	}
	sort.SliceStable(models, func(i, j int) bool {
		left := strings.ToLower(models[i].BrandName + " " + models[i].Model)
		right := strings.ToLower(models[j].BrandName + " " + models[j].Model)
		return left < right
	})
	return models, nil
}

func BuildWheelsetSpokeCatalogModelValueKey(brandSlug, wheelsetSlug string) string {
	return strings.ToLower(strings.TrimSpace(brandSlug) + ":" + strings.TrimSpace(wheelsetSlug))
}

func FindWheelsetSpokeCatalogModelByValueKey(valueKey string) (Model, bool, error) {
	models, err := ListPublishedWheelsetSpokeCatalogModels()
	if err != nil {
		return Model{}, false, err
	}
	key := strings.ToLower(strings.TrimSpace(valueKey))
	for _, model := range models {
		if BuildWheelsetSpokeCatalogModelValueKey(model.BrandSlug, model.Slug) == key {
			return model, true, nil
		}
	}
	return Model{}, false, nil
}
