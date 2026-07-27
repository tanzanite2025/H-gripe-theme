package admin

import (
	"errors"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type MarketingHandler struct {
	marketingService *service.MarketingService
	programService   *service.LoyaltyProgramService
}

func NewMarketingHandler(marketingService *service.MarketingService, programServices ...*service.LoyaltyProgramService) *MarketingHandler {
	handler := &MarketingHandler{marketingService: marketingService}
	if len(programServices) > 0 {
		handler.programService = programServices[0]
	}
	return handler
}

func respondMarketingError(c *gin.Context, err error, notFoundResource string) {
	switch {
	case errors.Is(err, service.ErrMarketingNotFound):
		apierror.RespondNotFound(c, notFoundResource)
	case errors.Is(err, service.ErrCouponCodeExists), errors.Is(err, service.ErrGiftCardCodeExists):
		apierror.RespondConflict(c, err.Error())
	case errors.Is(err, service.ErrInvalidGiftCardStatusTransition), errors.Is(err, service.ErrInvalidMemberLevel), errors.Is(err, service.ErrInvalidLoyaltyProgramConfig):
		apierror.RespondBadRequest(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
