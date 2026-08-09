package admin

import (
	"errors"

	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type StorefrontMarketHandler struct {
	marketService *service.StorefrontMarketService
	auditService  adminAuditRecorder
}

func NewStorefrontMarketHandler(marketService *service.StorefrontMarketService) *StorefrontMarketHandler {
	return &StorefrontMarketHandler{marketService: marketService}
}

func (h *StorefrontMarketHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *StorefrontMarketHandler) ListMarkets(c *gin.Context) {
	if h == nil || h.marketService == nil {
		apierror.RespondInternalError(c, errors.New("storefront market service is not configured"))
		return
	}
	markets, err := h.marketService.List(false)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": markets})
}

func (h *StorefrontMarketHandler) GetMarket(c *gin.Context) {
	if h == nil || h.marketService == nil {
		apierror.RespondInternalError(c, errors.New("storefront market service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid market id")
	if err != nil {
		return
	}
	market, err := h.marketService.Get(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Storefront market")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, market)
}

func (h *StorefrontMarketHandler) CreateMarket(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	var req storefrontMarketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordStorefrontMarketAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceStorefrontMarket,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.marketService == nil {
		err := errors.New("storefront market service is not configured")
		h.recordStorefrontMarketAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceStorefrontMarket,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		apierror.RespondInternalError(c, err)
		return
	}
	market, err := h.marketService.Create(req.toServiceInput())
	if err != nil {
		h.recordStorefrontMarketAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceStorefrontMarket,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		respondStorefrontMarketError(c, err)
		return
	}
	h.recordStorefrontMarketAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionCreate,
		Resource:   adminAuditResourceStorefrontMarket,
		ResourceID: market.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    market,
		NewValue:   market,
	})
	response.Created(c, market)
}

func (h *StorefrontMarketHandler) UpdateMarket(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.marketService == nil {
		err := errors.New("storefront market service is not configured")
		apierror.RespondInternalError(c, err)
		return
	}
	id, err := parseUintParam(c, "id", "invalid market id")
	if err != nil {
		return
	}
	oldMarket, _ := h.marketService.Get(id)

	var req storefrontMarketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordStorefrontMarketAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceStorefrontMarket,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldMarket,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	market, err := h.marketService.Update(id, req.toServiceInput())
	if err != nil {
		h.recordStorefrontMarketAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceStorefrontMarket,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldMarket,
		})
		respondStorefrontMarketError(c, err)
		return
	}
	h.recordStorefrontMarketAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceStorefrontMarket,
		ResourceID: market.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    market,
		OldValue:   oldMarket,
		NewValue:   market,
	})
	response.Success(c, market)
}

func (h *StorefrontMarketHandler) DeleteMarket(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.marketService == nil {
		err := errors.New("storefront market service is not configured")
		apierror.RespondInternalError(c, err)
		return
	}
	id, err := parseUintParam(c, "id", "invalid market id")
	if err != nil {
		return
	}
	oldMarket, _ := h.marketService.Get(id)
	if err := h.marketService.Delete(id); err != nil {
		h.recordStorefrontMarketAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceStorefrontMarket,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			OldValue:     oldMarket,
		})
		respondStorefrontMarketError(c, err)
		return
	}
	h.recordStorefrontMarketAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionDelete,
		Resource:   adminAuditResourceStorefrontMarket,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		OldValue:   oldMarket,
	})
	response.SuccessWithMessage(c, "storefront market deleted", nil)
}

func (h *StorefrontMarketHandler) GetOptions(c *gin.Context) {
	if h == nil || h.marketService == nil {
		apierror.RespondInternalError(c, errors.New("storefront market service is not configured"))
		return
	}
	response.Success(c, h.marketService.Options())
}

func respondStorefrontMarketError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidStorefrontMarket) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if repository.IsRecordNotFound(err) {
		apierror.RespondNotFound(c, "Storefront market")
		return
	}
	apierror.RespondInternalError(c, err)
}
