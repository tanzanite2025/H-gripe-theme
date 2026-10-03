package brandwheelsetspoke

import (
	"net/http"
	"sort"
	"strings"

	"commerce-platform/internal/domain/wheelsetcatalog"

	"github.com/gin-gonic/gin"
)

// The complete repair-kit directory is kept on the backend. The public model
// index exposes verified spoke and nipple specifications; product editors use
// a reduced selector response. Router-level rate limits bound catalog access.
type brandCatalog = wheelsetcatalog.Brand
type wheelsetSpec = wheelsetcatalog.Wheelset
type rimSpec = wheelsetcatalog.Rim
type wheelPositionSpec = wheelsetcatalog.Wheel
type spokeSideSpec = wheelsetcatalog.Side

type modelIndex struct {
	BrandSlug       string            `json:"brandSlug"`
	BrandName       string            `json:"brandName"`
	Slug            string            `json:"slug"`
	Model           string            `json:"model"`
	LifecycleStatus string            `json:"lifecycleStatus"`
	Rim             publicRimSpec     `json:"rim"`
	Wheels          []publicWheelSpec `json:"wheels"`
	NippleModel     string            `json:"nippleModel"`
	NippleLengthMM  *float64          `json:"nippleLengthMm"`
}

type selectorModel struct {
	BrandSlug       string `json:"brandSlug"`
	BrandName       string `json:"brandName"`
	Slug            string `json:"slug"`
	Model           string `json:"model"`
	LifecycleStatus string `json:"lifecycleStatus"`
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
	Side       string   `json:"side"`
	LengthMM   *float64 `json:"lengthMm"`
	SpokeModel string   `json:"spokeModel"`
	HeadType   string   `json:"headType"`
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
	brands, err := wheelsetcatalog.LoadWheelsetSpokeCatalogBrands()
	if err != nil {
		panic("brand wheelset spoke catalog is invalid: " + err.Error())
	}
	return &Handler{brands: brands}
}

// RegisterRoutes exposes the published model index and single-model details.
// Optional authentication and rate limiting are attached by the router so all
// replicas share the same API policy while anonymous visitors can use the
// reference data.
func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/models", h.ListModels)
	group.GET("/selector-models", h.ListSelectorModels)
	group.GET("/models/:slug", h.GetModel)
}

// ListSelectorModels returns only the finite model labels and keys needed by
// product editors. It does not expose spoke, nipple, or dimensional data.
func (h *Handler) ListSelectorModels(c *gin.Context) {
	models := make([]selectorModel, 0)
	for _, item := range h.publicModels() {
		models = append(models, selectorModel{
			BrandSlug: item.BrandSlug, BrandName: item.BrandName,
			Slug: item.Slug, Model: item.Model, LifecycleStatus: item.LifecycleStatus,
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"models": models}})
}

func (h *Handler) ListModels(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=86400, stale-while-revalidate=3600")
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
			// Published specifications are identical for every visitor, so a
			// shared cache can safely serve the response to search crawlers and
			// anonymous repair-kit shoppers.
			c.Header("Cache-Control", "public, max-age=86400, stale-while-revalidate=3600")
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
				LengthMM:   side.LengthMM,
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
		Wheels:         wheels,
		NippleModel:    wheelset.NippleModel,
		NippleLengthMM: wheelset.NippleLengthMM,
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
