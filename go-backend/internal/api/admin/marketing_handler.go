package admin

import (
	"errors"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type MarketingHandler struct {
	marketingService *service.MarketingService
}

func NewMarketingHandler(marketingService *service.MarketingService) *MarketingHandler {
	return &MarketingHandler{marketingService: marketingService}
}

func respondMarketingError(c *gin.Context, err error, notFoundResource string) {
	switch {
	case errors.Is(err, service.ErrMarketingNotFound):
		apierror.RespondNotFound(c, notFoundResource)
	case errors.Is(err, service.ErrCouponCodeExists), errors.Is(err, service.ErrGiftCardCodeExists):
		apierror.RespondConflict(c, err.Error())
	case errors.Is(err, service.ErrInvalidGiftCardStatusTransition), errors.Is(err, service.ErrInvalidMemberLevel):
		apierror.RespondBadRequest(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
