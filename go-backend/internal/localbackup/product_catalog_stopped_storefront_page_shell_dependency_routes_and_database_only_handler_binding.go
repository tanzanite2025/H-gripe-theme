package localbackup

import (
	publiccontent "commerce-platform/internal/api/v1/content"
	publiccurrency "commerce-platform/internal/api/v1/currency"
	publicproduct "commerce-platform/internal/api/v1/product"
	publicseo "commerce-platform/internal/api/v1/seo/home"
	publicsettings "commerce-platform/internal/api/v1/settings"
	publicstorefront "commerce-platform/internal/api/v1/storefront"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Only public, database-backed page-shell reads belong here. Payment gateway
// availability, authentication, tracking and personalized recommendations need
// their own side-effect and identity contracts before they can be bound.
var stoppedStorefrontPageShellReadOnlyGETPaths = map[string]string{
	"redirects":               "/api/v1/storefront/redirects",
	"seo_home":                "/api/v1/seo/home",
	"social_settings":         "/api/v1/settings/social",
	"website_profile":         "/api/v1/settings/website-profile",
	"website_name":            "/api/v1/settings/website-name",
	"specification_templates": "/api/v1/products/specification-templates",
	"currency_policy":         "/api/v1/settings/currency-policy",
	"product_page_faq":        "/api/v1/content/faq-pages/by-route",
}

func bindStoppedProductReadOnlyStorefrontPageShellDatabaseHandlers(router *gin.Engine, database *gorm.DB, settings *service.SettingService, settingsHandler *publicsettings.Handler, productHandler *publicproduct.Handler, policy *service.CurrencyPolicyService, media *service.MediaService) {
	settingsHandler.ConfigureWebsiteNameService(service.NewWebsiteNameService(settings))
	redirects := publicstorefront.NewRedirectsHandler(service.NewStorefrontRedirectRuleService(repository.NewStorefrontRedirectRuleRepository(database), nil))
	router.GET(stoppedStorefrontPageShellReadOnlyGETPaths["redirects"], redirects.ListPublished)
	router.GET(stoppedStorefrontPageShellReadOnlyGETPaths["seo_home"], publicseo.NewHandler(service.NewSEOService(settings)).Get)
	router.GET(stoppedStorefrontPageShellReadOnlyGETPaths["social_settings"], settingsHandler.GetSocialSettings)
	router.GET(stoppedStorefrontPageShellReadOnlyGETPaths["website_profile"], settingsHandler.GetWebsiteProfile)
	router.GET(stoppedStorefrontPageShellReadOnlyGETPaths["website_name"], settingsHandler.GetWebsiteName)
	router.GET(stoppedStorefrontPageShellReadOnlyGETPaths["specification_templates"], productHandler.ListProductSpecificationTemplates)
	router.GET(stoppedStorefrontPageShellReadOnlyGETPaths["currency_policy"], publiccurrency.NewHandler(policy, nil).GetPolicy)
	faq := service.NewFAQService(repository.NewFAQRepository(database))
	faq.ConfigureMediaService(media)
	router.GET(stoppedStorefrontPageShellReadOnlyGETPaths["product_page_faq"], publiccontent.NewHandler(nil, faq, media).GetFAQPageByRoute)
}
