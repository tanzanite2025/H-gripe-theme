package currency

import (
	"errors"
	"net/http"
	"strings"

	domaincurrency "commerce-platform/internal/domain/currency"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	policyService       *service.CurrencyPolicyService
	exchangeRateService *service.ExchangeRateService
}

func NewHandler(policyService *service.CurrencyPolicyService, exchangeRateService *service.ExchangeRateService) *Handler {
	return &Handler{policyService: policyService, exchangeRateService: exchangeRateService}
}

func (h *Handler) GetPolicy(c *gin.Context) {
	policy, err := h.policyService.GetPolicy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": policy})
}

func (h *Handler) ListExchangeRates(c *gin.Context) {
	if h == nil || h.exchangeRateService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "exchange rate service is not configured"})
		return
	}
	defaultBase := currencyDefaultBase(h.policyService)
	base := strings.ToUpper(strings.TrimSpace(c.DefaultQuery("base", defaultBase)))
	rates, err := h.exchangeRateService.List(base)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	config, err := h.exchangeRateService.GetConfig()
	if err != nil && !errors.Is(err, service.ErrExchangeRateDisabled) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"base_currency":    base,
		"provider":         config.Provider,
		"refresh_minutes":  config.RefreshMinutes,
		"quote_currencies": config.QuoteCurrencies,
		"rates":            rates,
	}})
}

func currencyDefaultBase(policyService *service.CurrencyPolicyService) string {
	if policyService == nil {
		return domaincurrency.DefaultPrimaryCurrency
	}
	primary, err := policyService.PrimaryCurrency()
	if err != nil || strings.TrimSpace(primary) == "" {
		return domaincurrency.DefaultPrimaryCurrency
	}
	return strings.ToUpper(strings.TrimSpace(primary))
}
