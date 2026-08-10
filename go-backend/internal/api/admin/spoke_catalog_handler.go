package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"tanzanite/internal/domain/spoke"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type SpokeCatalogHandler struct {
	spokeService *service.SpokeService
}

func NewSpokeCatalogHandler(spokeService *service.SpokeService) *SpokeCatalogHandler {
	return &SpokeCatalogHandler{spokeService: spokeService}
}

func (h *SpokeCatalogHandler) Get(c *gin.Context) {
	export, err := h.spokeService.GetExport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spoke_catalog_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, export)
}

func (h *SpokeCatalogHandler) Replace(c *gin.Context) {
	var payload spoke.ExportResponse
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	h.replaceCatalog(c, payload)
}

func (h *SpokeCatalogHandler) Import(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "JSON file is required"})
		return
	}
	if fileHeader.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "JSON file must be 2MB or smaller"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	defer file.Close()

	var payload spoke.ExportResponse
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json", "message": err.Error()})
		return
	}

	h.replaceCatalog(c, payload)
}

func (h *SpokeCatalogHandler) DownloadPresetTemplate(c *gin.Context) {
	template, err := h.spokeService.BuildPresetTemplate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spoke_catalog_template_error", "message": err.Error()})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="spoke-preset-template.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", template)
}

func (h *SpokeCatalogHandler) ImportPresetTemplate(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "XLSX preset template is required"})
		return
	}
	if fileHeader.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "XLSX preset template must be 10MB or smaller"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	defer file.Close()

	export, err := h.spokeService.ImportPresetTemplate(file)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSpokeCatalog) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_spoke_preset_template", "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spoke_catalog_template_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, export)
}

func (h *SpokeCatalogHandler) replaceCatalog(c *gin.Context, payload spoke.ExportResponse) {
	export, err := h.spokeService.ReplaceCatalog(payload)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSpokeCatalog) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_spoke_catalog", "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spoke_catalog_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, export)
}
