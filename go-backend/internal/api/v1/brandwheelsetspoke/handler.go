package brandwheelsetspoke

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// The complete repair-kit directory is kept on the backend. The browser only
// receives the public model index, then requests one exact model after login.
// This prevents the full directory from becoming part of the Nuxt JavaScript
// bundle or one bulk response.
//
//go:embed brand-wheelset-spoke-specs-catalog.json
var catalogJSON []byte

type brandCatalog struct {
	BrandSlug         string         `json:"brandSlug"`
	BrandName         string         `json:"brandName"`
	PublicationStatus string         `json:"publicationStatus"`
	SourceCheckedAt   *string        `json:"sourceCheckedAt"`
	Wheelsets         []wheelsetSpec `json:"wheelsets"`
}

type wheelsetSpec struct {
	Slug               string              `json:"slug"`
	Model              string              `json:"model"`
	LifecycleStatus    string              `json:"lifecycleStatus"`
	VerificationStatus string              `json:"verificationStatus"`
	Rim                rimSpec             `json:"rim"`
	NippleModel        string              `json:"nippleModel"`
	NippleLengthMM     *float64            `json:"nippleLengthMm"`
	Wheels             []wheelPositionSpec `json:"wheels"`
}

type rimSpec struct {
	DepthMM      float64  `json:"depthMm"`
	DepthFrontMM *float64 `json:"depthFrontMm"`
	DepthRearMM  *float64 `json:"depthRearMm"`
	InnerWidthMM float64  `json:"innerWidthMm"`
	OuterWidthMM float64  `json:"outerWidthMm"`
}

type wheelPositionSpec struct {
	Position      string          `json:"position"`
	SpokeCount    int             `json:"spokeCount"`
	LacingPattern string          `json:"lacingPattern"`
	Sides         []spokeSideSpec `json:"sides"`
}

type spokeSideSpec struct {
	Side       string   `json:"side"`
	LengthMM   *float64 `json:"lengthMm"`
	SpokeModel string   `json:"spokeModel"`
	HeadType   string   `json:"headType"`
}

type modelIndex struct {
	BrandSlug       string            `json:"brandSlug"`
	BrandName       string            `json:"brandName"`
	Slug            string            `json:"slug"`
	Model           string            `json:"model"`
	LifecycleStatus string            `json:"lifecycleStatus"`
	Rim             publicRimSpec     `json:"rim"`
	Wheels          []publicWheelSpec `json:"wheels"`
}

type publicRimSpec struct {
	DepthMM      float64  `json:"depthMm"`
	DepthFrontMM *float64 `json:"depthFrontMm,omitempty"`
	DepthRearMM  *float64 `json:"depthRearMm,omitempty"`
}

type publicWheelSpec struct {
	Position      string            `json:"position"`
	SpokeCount    int               `json:"spokeCount"`
	LacingPattern string            `json:"lacingPattern"`
	Sides         []publicSpokeSide `json:"sides"`
}

type publicSpokeSide struct {
	Side       string `json:"side"`
	SpokeModel string `json:"spokeModel"`
	HeadType   string `json:"headType"`
}

type modelDetail struct {
	BrandSlug       string            `json:"brandSlug"`
	BrandName       string            `json:"brandName"`
	Slug            string            `json:"slug"`
	Model           string            `json:"model"`
	LifecycleStatus string            `json:"lifecycleStatus"`
	Rim             publicRimSpec     `json:"rim"`
	Wheels          []detailWheelSpec `json:"wheels"`
	NippleModel     string            `json:"nippleModel"`
	NippleLengthMM  *float64          `json:"nippleLengthMm"`
}

type detailWheelSpec struct {
	Position      string            `json:"position"`
	SpokeCount    int               `json:"spokeCount"`
	LacingPattern string            `json:"lacingPattern"`
	Sides         []detailSpokeSide `json:"sides"`
}

type detailSpokeSide struct {
	Side       string   `json:"side"`
	SpokeModel string   `json:"spokeModel"`
	HeadType   string   `json:"headType"`
	LengthMM   *float64 `json:"lengthMm"`
}

type Handler struct {
	brands []brandCatalog
}

func NewHandler() *Handler {
	var brands []brandCatalog
	if err := json.Unmarshal(catalogJSON, &brands); err != nil {
		panic("brand wheelset spoke catalog is invalid: " + err.Error())
	}
	return &Handler{brands: brands}
}

// RegisterRoutes exposes a small public model index and a protected
// single-model detail endpoint. Authentication and rate limiting are attached
// by the router so all replicas share the same API policy.
func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/models", h.ListModels)
	group.GET("/models/:slug", h.GetModel)
}

func (h *Handler) ListModels(c *gin.Context) {
	models := h.publicModels()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"models":            models,
			"source_checked_at": h.sourceCheckedAt(),
		},
	})
}

func (h *Handler) GetModel(c *gin.Context) {
	if _, ok := c.Get("user_id"); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "registration_required",
			"message": "Sign in or create an account to view exact repair-kit specifications.",
		})
		return
	}

	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model slug is required"})
		return
	}

	for _, brand := range h.brands {
		if brand.PublicationStatus != "published" {
			continue
		}
		for _, wheelset := range brand.Wheelsets {
			if wheelset.Slug != slug || wheelset.VerificationStatus != "verified" {
				continue
			}
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"data": gin.H{"model": toModelDetail(brand, wheelset)},
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error":   "wheelset_model_not_found",
		"message": "Wheelset model not found.",
	})
}

func (h *Handler) publicModels() []modelIndex {
	models := make([]modelIndex, 0)
	for _, brand := range h.brands {
		if brand.PublicationStatus != "published" {
			continue
		}
		for _, wheelset := range brand.Wheelsets {
			if wheelset.VerificationStatus != "verified" {
				continue
			}
			models = append(models, toModelIndex(brand, wheelset))
		}
	}
	sort.SliceStable(models, func(i, j int) bool {
		return strings.ToLower(models[i].Model) < strings.ToLower(models[j].Model)
	})
	return models
}

func (h *Handler) sourceCheckedAt() *string {
	for _, brand := range h.brands {
		if brand.SourceCheckedAt != nil && strings.TrimSpace(*brand.SourceCheckedAt) != "" {
			value := strings.TrimSpace(*brand.SourceCheckedAt)
			return &value
		}
	}
	return nil
}

func toModelIndex(brand brandCatalog, wheelset wheelsetSpec) modelIndex {
	wheels := make([]publicWheelSpec, 0, len(wheelset.Wheels))
	for _, wheel := range wheelset.Wheels {
		sides := make([]publicSpokeSide, 0, len(wheel.Sides))
		for _, side := range wheel.Sides {
			sides = append(sides, publicSpokeSide{
				Side:       side.Side,
				SpokeModel: side.SpokeModel,
				HeadType:   side.HeadType,
			})
		}
		wheels = append(wheels, publicWheelSpec{
			Position:      wheel.Position,
			SpokeCount:    wheel.SpokeCount,
			LacingPattern: wheel.LacingPattern,
			Sides:         sides,
		})
	}
	return modelIndex{
		BrandSlug:       brand.BrandSlug,
		BrandName:       brand.BrandName,
		Slug:            wheelset.Slug,
		Model:           wheelset.Model,
		LifecycleStatus: wheelset.LifecycleStatus,
		Rim: publicRimSpec{
			DepthMM:      wheelset.Rim.DepthMM,
			DepthFrontMM: wheelset.Rim.DepthFrontMM,
			DepthRearMM:  wheelset.Rim.DepthRearMM,
		},
		Wheels: wheels,
	}
}

func toModelDetail(brand brandCatalog, wheelset wheelsetSpec) modelDetail {
	wheels := make([]detailWheelSpec, 0, len(wheelset.Wheels))
	for _, wheel := range wheelset.Wheels {
		sides := make([]detailSpokeSide, 0, len(wheel.Sides))
		for _, side := range wheel.Sides {
			sides = append(sides, detailSpokeSide{
				Side:       side.Side,
				SpokeModel: side.SpokeModel,
				HeadType:   side.HeadType,
				LengthMM:   side.LengthMM,
			})
		}
		wheels = append(wheels, detailWheelSpec{
			Position:      wheel.Position,
			SpokeCount:    wheel.SpokeCount,
			LacingPattern: wheel.LacingPattern,
			Sides:         sides,
		})
	}
	return modelDetail{
		BrandSlug:       brand.BrandSlug,
		BrandName:       brand.BrandName,
		Slug:            wheelset.Slug,
		Model:           wheelset.Model,
		LifecycleStatus: wheelset.LifecycleStatus,
		Rim: publicRimSpec{
			DepthMM:      wheelset.Rim.DepthMM,
			DepthFrontMM: wheelset.Rim.DepthFrontMM,
			DepthRearMM:  wheelset.Rim.DepthRearMM,
		},
		Wheels:         wheels,
		NippleModel:    wheelset.NippleModel,
		NippleLengthMM: wheelset.NippleLengthMM,
	}
}
