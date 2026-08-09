package product

import (
	"strings"

	"tanzanite/internal/api/middleware"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type publicProductContext struct {
	Locale          string
	DisplayCurrency string
	Response        *service.StorefrontContext
}

func (h *Handler) resolvePublicContext(c *gin.Context) publicProductContext {
	locale := middleware.GetLocale(c)
	result := publicProductContext{Locale: locale}
	if h == nil || h.storefrontContext == nil {
		return result
	}
	country, countrySource := productRequestCountry(c)
	context, err := h.storefrontContext.Resolve(service.StorefrontContextInput{
		Country:           country,
		CountrySource:     countrySource,
		RequestedLocale:   productFirstNonBlank(c.Query("locale"), c.GetHeader("X-Locale")),
		CookieLocale:      productCookieValue(c, "locale"),
		AcceptLanguage:    c.GetHeader("Accept-Language"),
		RequestedCurrency: productFirstNonBlank(c.Query("currency"), c.GetHeader("X-Display-Currency")),
		CookieCurrency:    productFirstNonBlank(productCookieValue(c, "display_currency"), productCookieValue(c, "currency")),
	})
	if err != nil || context == nil {
		return result
	}
	if context.Locale.Resolved != "" {
		result.Locale = context.Locale.Resolved
	}
	result.Response = context
	result.DisplayCurrency = context.Currency.Resolved
	return result
}

func productRequestCountry(c *gin.Context) (string, string) {
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

func productCookieValue(c *gin.Context, name string) string {
	if c == nil || strings.TrimSpace(name) == "" {
		return ""
	}
	value, err := c.Cookie(name)
	if err != nil {
		return ""
	}
	return value
}

func productFirstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
