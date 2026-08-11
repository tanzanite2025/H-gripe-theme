package seo

import (
	"net/http"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type HomeHandler struct {
	seoService *service.SEOService
	audit      seoAuditRecorder
}

func NewHomeHandler(seoService *service.SEOService) *HomeHandler {
	return &HomeHandler{seoService: seoService}
}

func (h *HomeHandler) ConfigureAuditService(recorder seoAuditRecorder) {
	if h == nil {
		return
	}
	h.audit = recorder
}

func (h *HomeHandler) Get(c *gin.Context) {
	settings, err := h.seoService.GetHome(localeFromQuery(c))
	if err != nil {
		writeGetError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func (h *HomeHandler) Update(c *gin.Context) {
	var request seodomain.UpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if request.Locale == "" {
		request.Locale = localeFromQuery(c)
	}

	startedAt := time.Now().UTC()
	oldSettings, _ := h.seoService.GetHome(request.Locale)
	settings, err := h.seoService.UpdateHome(request)
	if err != nil {
		recordSEOAudit(h.audit, c, seoAuditEvent{
			StartedAt:    startedAt,
			Resource:     seoAuditResourceHome,
			Status:       seoAuditStatusFailed,
			ErrorMessage: err.Error(),
			OldValue:     homeAuditValue(oldSettings),
			Changes: seoFieldChanges(map[string]interface{}{
				"meta_title":       request.MetaTitle != nil,
				"meta_description": request.MetaDescription != nil,
			}),
		})
		writeUpdateError(c, err)
		return
	}

	recordSEOAudit(h.audit, c, seoAuditEvent{
		StartedAt: startedAt,
		Resource:  seoAuditResourceHome,
		Status:    seoAuditStatusOK,
		Changes: seoFieldChanges(map[string]interface{}{
			"meta_title":       request.MetaTitle != nil,
			"meta_description": request.MetaDescription != nil,
		}),
		OldValue: homeAuditValue(oldSettings),
		NewValue: homeAuditValue(settings),
	})

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func homeAuditValue(settings *seodomain.Settings) map[string]string {
	if settings == nil {
		return nil
	}
	return map[string]string{
		"meta_title":       settings.MetaTitle,
		"meta_description": settings.MetaDescription,
	}
}
