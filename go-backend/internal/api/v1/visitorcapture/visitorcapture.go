package visitorcapture

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/pkg/visitorcookie"
	"commerce-platform/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type TouchOptions struct {
	UserID                     *uint
	CustomerServiceVisitorHash string
	CartSessionID              string
	Email                      string
	EmailSource                string
	LocaleSource               string
	MeaningfulAction           string
	QualityScoreDelta          int
}

func BuildVisitorProfileTouchInput(c *gin.Context, opts TouchOptions) service.VisitorProfileTouchInput {
	if c == nil || c.Request == nil {
		return service.VisitorProfileTouchInput{}
	}

	localeSource := strings.TrimSpace(opts.LocaleSource)
	if localeSource == "" {
		localeSource = "accept_language"
	}

	return service.VisitorProfileTouchInput{
		UserID:                     opts.UserID,
		CustomerServiceVisitorHash: strings.TrimSpace(opts.CustomerServiceVisitorHash),
		CartSessionID:              strings.TrimSpace(opts.CartSessionID),
		Email:                      strings.TrimSpace(opts.Email),
		EmailSource:                strings.TrimSpace(opts.EmailSource),
		Locale:                     requestLocale(c),
		LocaleSource:               localeSource,
		CountryCode:                requestCountryCode(c),
		Region:                     firstNonEmptyHeader(c, "CF-Region", "X-Region"),
		City:                       firstNonEmptyHeader(c, "CF-IPCity", "X-City"),
		Timezone:                   firstNonEmptyHeader(c, "CF-Timezone", "X-Timezone"),
		IPAddress:                  requestIP(c),
		UserAgent:                  c.GetHeader("User-Agent"),
		MeaningfulAction:           strings.TrimSpace(opts.MeaningfulAction),
		QualityScoreDelta:          opts.QualityScoreDelta,
	}
}

func ExistingCartSessionID(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(sessionID)
}

func ExistingCustomerServiceVisitorHash(c *gin.Context, visitorSecret []byte) string {
	if c == nil {
		return ""
	}
	hash, _ := visitorcookie.ExistingCustomerServiceVisitorHash(c, visitorSecret)
	return hash
}

func requestLocale(c *gin.Context) string {
	return firstNonEmptyHeader(c, "X-Locale", "Accept-Language")
}

func requestCountryCode(c *gin.Context) string {
	return middleware.TrustedEdgeCountry(c)
}

func requestIP(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	return c.ClientIP()
}

func firstNonEmptyHeader(c *gin.Context, keys ...string) string {
	if c == nil || c.Request == nil {
		return ""
	}
	for _, key := range keys {
		if value := strings.TrimSpace(c.GetHeader(key)); value != "" {
			return value
		}
	}
	return ""
}
