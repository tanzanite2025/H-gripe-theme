package admin

import (
	"errors"
	"net/http"
	"strconv"

	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const adminAuditResourceMediaDerivativePreset = "media_derivative_preset"
const adminAuditResourceMediaDerivativeRebuild = "media_derivative_rebuild"

type mediaDerivativePresetRequest struct {
	Code      string `json:"code"`
	Label     string `json:"label"`
	MaxWidth  int    `json:"max_width"`
	SortOrder int    `json:"sort_order"`
	Enabled   *bool  `json:"enabled"`
}

type mediaDerivativePresetEnabledRequest struct {
	Enabled *bool `json:"enabled" binding:"required"`
}

func (h *MediaHandler) ListDerivativePresets(c *gin.Context) {
	if h == nil || h.mediaService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": service.ErrMediaDerivativePresetUnavailable.Error()})
		return
	}

	presets, err := h.mediaService.ListMediaDerivativePresets()
	if err != nil {
		respondMediaDerivativePresetError(c, err)
		return
	}
	response.Success(c, gin.H{"presets": presets})
}

func (h *MediaHandler) CreateDerivativePreset(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	var req mediaDerivativePresetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h == nil || h.mediaService == nil {
		err := service.ErrMediaDerivativePresetUnavailable
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		respondMediaDerivativePresetError(c, err)
		return
	}

	preset, err := h.mediaService.CreateMediaDerivativePreset(mediaDerivativePresetInput(req))
	if err != nil {
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		respondMediaDerivativePresetError(c, err)
		return
	}
	h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionCreate,
		ResourceID: preset.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    req,
		NewValue:   preset,
	})
	response.Created(c, gin.H{"preset": preset})
}

func (h *MediaHandler) UpdateDerivativePreset(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, ok := parseMediaDerivativePresetID(c)
	if !ok {
		return
	}
	var req mediaDerivativePresetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h == nil || h.mediaService == nil {
		err := service.ErrMediaDerivativePresetUnavailable
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		respondMediaDerivativePresetError(c, err)
		return
	}

	oldPreset, err := h.mediaService.GetMediaDerivativePreset(id)
	if err != nil {
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		respondMediaDerivativePresetError(c, err)
		return
	}
	preset, err := h.mediaService.UpdateMediaDerivativePreset(id, mediaDerivativePresetInput(req))
	if err != nil {
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldPreset,
		})
		respondMediaDerivativePresetError(c, err)
		return
	}
	h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		Changes:    req,
		OldValue:   oldPreset,
		NewValue:   preset,
	})
	response.Success(c, gin.H{"preset": preset})
}

func (h *MediaHandler) UpdateDerivativePresetEnabled(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, ok := parseMediaDerivativePresetID(c)
	if !ok {
		return
	}
	var req mediaDerivativePresetEnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h == nil || h.mediaService == nil {
		err := service.ErrMediaDerivativePresetUnavailable
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      gin.H{"enabled": *req.Enabled},
		})
		respondMediaDerivativePresetError(c, err)
		return
	}

	oldPreset, err := h.mediaService.GetMediaDerivativePreset(id)
	if err != nil {
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      gin.H{"enabled": *req.Enabled},
		})
		respondMediaDerivativePresetError(c, err)
		return
	}
	preset, err := h.mediaService.SetMediaDerivativePresetEnabled(id, *req.Enabled)
	if err != nil {
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      gin.H{"enabled": *req.Enabled},
			OldValue:     oldPreset,
		})
		respondMediaDerivativePresetError(c, err)
		return
	}
	h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		Changes:    gin.H{"enabled": *req.Enabled},
		OldValue:   oldPreset,
		NewValue:   preset,
	})
	response.Success(c, gin.H{"preset": preset})
}

func (h *MediaHandler) DeleteDerivativePreset(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, ok := parseMediaDerivativePresetID(c)
	if !ok {
		return
	}
	if h == nil || h.mediaService == nil {
		err := service.ErrMediaDerivativePresetUnavailable
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		respondMediaDerivativePresetError(c, err)
		return
	}

	preset, err := h.mediaService.GetMediaDerivativePreset(id)
	if err != nil {
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		respondMediaDerivativePresetError(c, err)
		return
	}
	if err := h.mediaService.DeleteMediaDerivativePreset(id); err != nil {
		h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			OldValue:     preset,
		})
		respondMediaDerivativePresetError(c, err)
		return
	}
	h.recordMediaDerivativePresetAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionDelete,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		OldValue:   preset,
	})
	response.Success(c, gin.H{"preset": preset})
}

func (h *MediaHandler) ListDerivativeRebuildJobs(c *gin.Context) {
	if h == nil || h.mediaService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": service.ErrMediaDerivativePresetUnavailable.Error()})
		return
	}
	jobs, err := h.mediaService.ListMediaDerivativeRebuildJobs()
	if err != nil {
		respondMediaDerivativePresetError(c, err)
		return
	}
	response.Success(c, gin.H{"jobs": jobs})
}

func (h *MediaHandler) RequestDerivativeRebuild(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.mediaService == nil {
		err := service.ErrMediaDerivativePresetUnavailable
		h.recordMediaDerivativeRebuildAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		respondMediaDerivativePresetError(c, err)
		return
	}
	job, err := h.mediaService.RequestMediaDerivativeRebuild("manual_request")
	if err != nil {
		h.recordMediaDerivativeRebuildAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		respondMediaDerivativePresetError(c, err)
		return
	}
	h.recordMediaDerivativeRebuildAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionCreate,
		ResourceID: job.ID,
		Status:     adminAuditStatusSuccess,
		NewValue:   job,
	})
	response.Success(c, gin.H{"job": job})
}

func (h *MediaHandler) recordMediaDerivativePresetAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	if event.Resource == "" {
		event.Resource = adminAuditResourceMediaDerivativePreset
	}
	recordAdminAudit(h.auditService, c, event)
}

func (h *MediaHandler) recordMediaDerivativeRebuildAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	event.Resource = adminAuditResourceMediaDerivativeRebuild
	recordAdminAudit(h.auditService, c, event)
}

func mediaDerivativePresetInput(req mediaDerivativePresetRequest) service.MediaDerivativePresetInput {
	return service.MediaDerivativePresetInput{
		Code:      req.Code,
		Label:     req.Label,
		MaxWidth:  req.MaxWidth,
		SortOrder: req.SortOrder,
		Enabled:   req.Enabled,
	}
}

func parseMediaDerivativePresetID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media derivative preset id"})
		return 0, false
	}
	return uint(id), true
}

func respondMediaDerivativePresetError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidMediaDerivativePreset),
		errors.Is(err, service.ErrMediaDerivativePresetCodeImmutable):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrMediaDerivativePresetNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrMediaDerivativePresetConflict),
		errors.Is(err, service.ErrMediaDerivativePresetInUse),
		errors.Is(err, service.ErrMediaDerivativePresetProtected),
		errors.Is(err, service.ErrMediaDerivativePresetLimitReached):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrMediaDerivativePresetUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "media derivative preset operation failed"})
	}
}
