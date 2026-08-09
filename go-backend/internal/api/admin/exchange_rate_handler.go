package admin

import (
	"errors"
	"math"

	"tanzanite/internal/domain/currency"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type ExchangeRateHandler struct {
	exchangeRateService *service.ExchangeRateService
	auditService        adminAuditRecorder
}

type exchangeRateConvertRequest struct {
	Amount          float64  `json:"amount"`
	BaseCurrency    string   `json:"base_currency"`
	QuoteCurrencies []string `json:"quote_currencies"`
}

type exchangeRateDisplayPrice struct {
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	QuoteCurrency  string  `json:"quote_currency"`
	Rate           float64 `json:"rate"`
	Source         string  `json:"source"`
	Converted      bool    `json:"converted"`
	FallbackReason string  `json:"fallback_reason,omitempty"`
}

func NewExchangeRateHandler(exchangeRateService *service.ExchangeRateService) *ExchangeRateHandler {
	return &ExchangeRateHandler{exchangeRateService: exchangeRateService}
}

func (h *ExchangeRateHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *ExchangeRateHandler) GetExchangeRates(c *gin.Context) {
	if h == nil || h.exchangeRateService == nil {
		apierror.RespondInternalError(c, errors.New("exchange rate service is not configured"))
		return
	}
	config, err := h.exchangeRateService.GetConfig()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	rates, err := h.exchangeRateService.List(config.BaseCurrency)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"config": config, "rates": rates})
}

func (h *ExchangeRateHandler) SyncExchangeRates(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.exchangeRateService == nil {
		err := errors.New("exchange rate service is not configured")
		apierror.RespondInternalError(c, err)
		return
	}
	result, err := h.exchangeRateService.Sync()
	if err != nil {
		recordAdminAudit(h.auditService, c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionExecute,
			Resource:     adminAuditResourceExchangeRate,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		if errors.Is(err, service.ErrExchangeRateDisabled) || errors.Is(err, service.ErrExchangeRateNotConfigured) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt: startedAt,
		Action:    adminAuditActionExecute,
		Resource:  adminAuditResourceExchangeRate,
		Status:    adminAuditStatusSuccess,
		Changes: gin.H{
			"base_currency": result.Config.BaseCurrency,
			"rate_count":    len(result.Rates),
			"provider":      result.Config.Provider,
		},
		NewValue: result,
	})
	response.Success(c, result)
}

func (h *ExchangeRateHandler) ConvertDisplayPrices(c *gin.Context) {
	if h == nil || h.exchangeRateService == nil {
		apierror.RespondInternalError(c, errors.New("exchange rate service is not configured"))
		return
	}

	var req exchangeRateConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, "invalid exchange-rate conversion request")
		return
	}
	if req.Amount <= 0 {
		apierror.RespondBadRequest(c, "amount must be greater than zero")
		return
	}

	config, err := h.exchangeRateService.GetConfig()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	baseCurrency := currency.NormalizeCode(req.BaseCurrency)
	if baseCurrency == "" {
		baseCurrency = config.BaseCurrency
	}
	if !currency.IsCatalogCode(baseCurrency) {
		apierror.RespondBadRequest(c, "unsupported base currency")
		return
	}

	quoteCurrencies := normalizeQuoteCurrencies(req.QuoteCurrencies, config.QuoteCurrencies, baseCurrency)
	if len(quoteCurrencies) == 0 {
		apierror.RespondBadRequest(c, "quote currencies are required")
		return
	}

	prices := make([]exchangeRateDisplayPrice, 0, len(quoteCurrencies))
	for _, quoteCurrency := range quoteCurrencies {
		converted := h.exchangeRateService.Convert(req.Amount, baseCurrency, quoteCurrency)
		prices = append(prices, exchangeRateDisplayPrice{
			Amount:         roundCurrencyAmount(converted.Amount, converted.Currency),
			Currency:       converted.Currency,
			QuoteCurrency:  quoteCurrency,
			Rate:           converted.Rate,
			Source:         converted.Source,
			Converted:      converted.Converted,
			FallbackReason: converted.FallbackReason,
		})
	}

	response.Success(c, gin.H{
		"amount":           req.Amount,
		"base_currency":    baseCurrency,
		"quote_currencies": quoteCurrencies,
		"prices":           prices,
	})
}

func normalizeQuoteCurrencies(requested []string, defaults []string, baseCurrency string) []string {
	values := requested
	if len(values) == 0 {
		values = defaults
	}
	values = currency.NormalizeCodes(values)
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, code := range values {
		if code == baseCurrency || !currency.IsCatalogCode(code) {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, code)
	}
	return result
}

func roundCurrencyAmount(amount float64, currencyCode string) float64 {
	minorUnits, ok := currency.MinorUnits(currencyCode)
	if !ok {
		minorUnits = 2
	}
	factor := math.Pow10(minorUnits)
	return math.Round(amount*factor) / factor
}
