package admin

import (
	"errors"
	"net/http"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type TaxRateSourceSnapshotHandler struct {
	snapshotService *service.TaxRateSourceSnapshotService
	auditService    adminAuditRecorder
}

func NewTaxRateSourceSnapshotHandler(snapshotService *service.TaxRateSourceSnapshotService) *TaxRateSourceSnapshotHandler {
	return &TaxRateSourceSnapshotHandler{snapshotService: snapshotService}
}

func (h *TaxRateSourceSnapshotHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *TaxRateSourceSnapshotHandler) GetSourceConfiguration(c *gin.Context) {
	if h == nil || h.snapshotService == nil {
		apierror.RespondInternalError(c, errors.New("tax rate source snapshot service is not configured"))
		return
	}
	configuration, err := h.snapshotService.GetSourceConfiguration()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"config": configuration})
}

func (h *TaxRateSourceSnapshotHandler) UpdateSourceConfiguration(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.snapshotService == nil {
		apierror.RespondInternalError(c, errors.New("tax rate source snapshot service is not configured"))
		return
	}
	var request service.TaxRateSourceConfigurationUpdate
	if err := c.ShouldBindJSON(&request); err != nil {
		apierror.RespondBadRequest(c, "invalid tax rate source configuration")
		return
	}
	oldConfiguration, err := h.snapshotService.GetSourceConfiguration()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	newConfiguration, err := h.snapshotService.UpdateSourceConfiguration(request)
	if err != nil {
		_ = recordAdminAudit(h.auditService, c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceTaxRateSource,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		if errors.Is(err, service.ErrTaxRateSourceConfigInvalid) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	_ = recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt: startedAt,
		Action:    adminAuditActionUpdate,
		Resource:  adminAuditResourceTaxRateSource,
		Status:    adminAuditStatusSuccess,
		Changes: gin.H{
			"enabled":                newConfiguration.Enabled,
			"refresh_interval_hours": newConfiguration.RefreshIntervalHours,
		},
		OldValue: gin.H{
			"enabled":                oldConfiguration.Enabled,
			"refresh_interval_hours": oldConfiguration.RefreshIntervalHours,
		},
		NewValue: gin.H{
			"enabled":                newConfiguration.Enabled,
			"refresh_interval_hours": newConfiguration.RefreshIntervalHours,
		},
	})
	response.Success(c, gin.H{"config": newConfiguration})
}

func (h *TaxRateSourceSnapshotHandler) GetCurrentSnapshot(c *gin.Context) {
	if h == nil || h.snapshotService == nil {
		apierror.RespondInternalError(c, errors.New("tax rate source snapshot service is not configured"))
		return
	}
	snapshot, err := h.snapshotService.GetCurrentSnapshot()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, snapshot)
}

func (h *TaxRateSourceSnapshotHandler) SyncSourceSnapshot(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.snapshotService == nil {
		err := errors.New("tax rate source snapshot service is not configured")
		apierror.RespondInternalError(c, err)
		return
	}
	result, err := h.snapshotService.Sync(c.Request.Context())
	if err != nil {
		_ = recordAdminAudit(h.auditService, c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionExecute,
			Resource:     adminAuditResourceTaxRateSource,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		if errors.Is(err, service.ErrTaxRateSourceDisabled) || errors.Is(err, service.ErrTaxRateSourceConfigInvalid) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrTaxRateSourceSyncInProgress) {
			apierror.RespondConflict(c, err.Error())
			return
		}
		apierror.RespondError(c, http.StatusBadGateway, "tax_rate_source_sync_failed", err.Error())
		return
	}
	_ = recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt: startedAt,
		Action:    adminAuditActionExecute,
		Resource:  adminAuditResourceTaxRateSource,
		Status:    adminAuditStatusSuccess,
		Changes: gin.H{
			"provider_code":  result.Snapshot.ProviderCode,
			"version":        result.Snapshot.Version,
			"content_sha256": result.Snapshot.ContentSHA256,
			"country_count":  result.Snapshot.CountryCount,
			"rate_count":     result.Snapshot.RateCount,
			"changed":        result.Changed,
		},
		NewValue: result,
	})
	response.Success(c, result)
}
