package admin

import (
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"
	"errors"

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
	case errors.Is(err, service.ErrCouponCodeExists):
		apierror.RespondConflict(c, err.Error())
	case errors.Is(err, service.ErrInvalidGiftCardStatusTransition), errors.Is(err, service.ErrInvalidMemberLevel), errors.Is(err, service.ErrInvalidLoyaltyProgramConfig), errors.Is(err, service.ErrInvalidCurrencyPolicy):
		apierror.RespondBadRequest(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
