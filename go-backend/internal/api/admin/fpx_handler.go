package admin

import (
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FpxHandler owns only 4PX domain data. It deliberately does not read or
// mutate generic shipping templates, carriers, or carrier services.
type FpxHandler struct {
	shippingService *service.ShippingService
}

func NewFpxHandler(shippingService *service.ShippingService) *FpxHandler {
	return &FpxHandler{shippingService: shippingService}
}

func (h *FpxHandler) GetOverview(c *gin.Context) {
	overview, err := h.shippingService.GetFpxOverview()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, overview)
}

func (h *FpxHandler) ListChannels(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"
	channels, err := h.shippingService.ListFpxChannels(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": channels})
}

// ListPublishedCollection is the only 4PX output consumed by downstream
// logistics management. It intentionally exposes enabled service references
// only, never credentials, orders, prices, or direct-shipping tasks.
func (h *FpxHandler) ListPublishedCollection(c *gin.Context) {
	channels, err := h.shippingService.ListFpxChannels(true)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": channels})
}

func (h *FpxHandler) UpdateChannel(c *gin.Context) {
	id, err := parseFpxChannelID(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	var channel shippingdomain.FpxChannel
	if err := c.ShouldBindJSON(&channel); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	channel.ID = id
	existing, err := h.shippingService.GetFpxChannel(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierror.RespondNotFound(c, "4PX channel")
		} else {
			apierror.RespondInternalError(c, err)
		}
		return
	}
	existing.Enabled = channel.Enabled
	if err := h.shippingService.UpdateFpxChannel(existing); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierror.RespondNotFound(c, "4PX channel")
			return
		}
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, existing)
}

func parseFpxChannelID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		return 0, errors.New("invalid 4PX channel id")
	}
	return uint(id), nil
}
