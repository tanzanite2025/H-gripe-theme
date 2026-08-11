package storefront

import (
	"net/http"
	"strings"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type ContextHandler struct {
	contextService *service.StorefrontContextService
}

func NewContextHandler(contextService *service.StorefrontContextService) *ContextHandler {
	return &ContextHandler{contextService: contextService}
}

func (h *ContextHandler) GetContext(c *gin.Context) {
	if h == nil || h.contextService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "storefront context service is not configured"})
		return
	}
	country, countrySource := requestCountry(c)
	context, err := h.contextService.Resolve(service.StorefrontContextInput{
		Country:           country,
		CountrySource:     countrySource,
		RequestedLocale:   firstNonBlank(c.Query("locale"), c.GetHeader("X-Locale")),
		CookieLocale:      cookieValue(c, "locale"),
		AcceptLanguage:    c.GetHeader("Accept-Language"),
		RequestedCurrency: firstNonBlank(c.Query("currency"), c.GetHeader("X-Display-Currency")),
		CookieCurrency:    firstNonBlank(cookieValue(c, "display_currency"), cookieValue(c, "currency")),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": context})
}

func requestCountry(c *gin.Context) (string, string) {
	if c == nil {
		return "", "fallback"
	}
	for _, candidate := range []struct {
		value  string
		source string
	}{
		{c.Query("country"), "request"},
		{c.GetHeader("X-Market-Country"), "request_header"},
		{c.GetHeader("CF-IPCountry"), "cf_ip_country"},
		{c.GetHeader("X-Country-Code"), "country_header"},
	} {
		if strings.TrimSpace(candidate.value) != "" {
			return candidate.value, candidate.source
		}
	}
	return "", "fallback"
}

func cookieValue(c *gin.Context, name string) string {
	if c == nil || strings.TrimSpace(name) == "" {
		return ""
	}
	value, err := c.Cookie(name)
	if err != nil {
		return ""
	}
	return value
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
