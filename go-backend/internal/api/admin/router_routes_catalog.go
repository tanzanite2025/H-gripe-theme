package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerProductRoutes(
	authenticated *gin.RouterGroup,
	productHandler *ProductHandler,
	productProcurementHandler *ProductProcurementHandler,
	frameFitmentEntryHandler *FrameFitmentEntryHandler,
	forkFitmentEntryHandler *ForkFitmentEntryHandler,
	fitmentHubSpecificationHandler *FitmentHubSpecificationHandler,
	productProfitabilityHandler *ProductProfitabilityHandler,
	productCategoryHandler *ProductCategoryHandler,
	productBrandHandler *ProductBrandHandler,
	productInformationTemplateHandler *ProductInformationTemplateHandler,
	customsClassificationHandler *CustomsClassificationHandler,
	spokeCatalogHandler *SpokeCatalogHandler,
	quickBuyHandler *QuickBuyHandler,
	selectionAssistantHandler *SelectionAssistantHandler,
	selectionConfigurationKeyHandler *SelectionConfigurationKeyHandler,
	wheelsetFitQuestionnaireHandler *WheelsetFitQuestionnaireHandler,
) {
	// 商品管理（需要商品管理权限）
	// 采购资料是独立附加域，只读写自身产品编码/名称快照。
	productProcurementGroup := authenticated.Group("/procurement/records")
	productProcurementGroup.Use(middleware.RequirePermission(auth.PermProcurementView))
	{
		productProcurementGroup.GET("", productProcurementHandler.List)
		productProcurementGroup.GET("/by-codes", productProcurementHandler.ListByCodes)
		productProcurementGroup.GET("/:id", productProcurementHandler.Get)
		productProcurementGroup.POST("", middleware.RequirePermission(auth.PermProcurementCreate), productProcurementHandler.Create)
		productProcurementGroup.PUT("/:id", middleware.RequirePermission(auth.PermProcurementEdit), productProcurementHandler.Update)
		productProcurementGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProcurementDelete), productProcurementHandler.Delete)
	}

	productProcurementOptionsGroup := authenticated.Group("/procurement/product-options")
	productProcurementOptionsGroup.Use(middleware.RequirePermission(auth.PermProcurementView))
	{
		productProcurementOptionsGroup.GET("", productProcurementHandler.ProductOptions)
	}

	frameFitmentEntriesGroup := authenticated.Group("/fitment-catalog/frame-entries")
	frameFitmentEntriesGroup.Use(middleware.RequirePermission(auth.PermFitmentCatalogView))
	{
		frameFitmentEntriesGroup.GET("", frameFitmentEntryHandler.List)
		frameFitmentEntriesGroup.GET("/:id", frameFitmentEntryHandler.Get)
		frameFitmentEntriesGroup.POST("", middleware.RequirePermission(auth.PermFitmentCatalogCreate), frameFitmentEntryHandler.Create)
		frameFitmentEntriesGroup.PUT("/:id", middleware.RequirePermission(auth.PermFitmentCatalogEdit), frameFitmentEntryHandler.Update)
		frameFitmentEntriesGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermFitmentCatalogEdit), frameFitmentEntryHandler.UpdateStatus)
		frameFitmentEntriesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermFitmentCatalogDelete), frameFitmentEntryHandler.Delete)
	}

	forkFitmentEntriesGroup := authenticated.Group("/fitment-catalog/fork-entries")
	forkFitmentEntriesGroup.Use(middleware.RequirePermission(auth.PermFitmentCatalogView))
	{
		forkFitmentEntriesGroup.GET("", forkFitmentEntryHandler.List)
		forkFitmentEntriesGroup.GET("/:id", forkFitmentEntryHandler.Get)
		forkFitmentEntriesGroup.POST("", middleware.RequirePermission(auth.PermFitmentCatalogCreate), forkFitmentEntryHandler.Create)
		forkFitmentEntriesGroup.PUT("/:id", middleware.RequirePermission(auth.PermFitmentCatalogEdit), forkFitmentEntryHandler.Update)
		forkFitmentEntriesGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermFitmentCatalogEdit), forkFitmentEntryHandler.UpdateStatus)
		forkFitmentEntriesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermFitmentCatalogDelete), forkFitmentEntryHandler.Delete)
	}

	fitmentHubSpecificationsGroup := authenticated.Group("/fitment-catalog/hub-specifications")
	fitmentHubSpecificationsGroup.Use(middleware.RequirePermission(auth.PermFitmentCatalogView))
	{
		fitmentHubSpecificationsGroup.GET("", fitmentHubSpecificationHandler.List)
		fitmentHubSpecificationsGroup.GET("/:id", fitmentHubSpecificationHandler.Get)
		fitmentHubSpecificationsGroup.POST("", middleware.RequirePermission(auth.PermFitmentCatalogCreate), fitmentHubSpecificationHandler.Create)
		fitmentHubSpecificationsGroup.PUT("/:id", middleware.RequirePermission(auth.PermFitmentCatalogEdit), fitmentHubSpecificationHandler.Update)
		fitmentHubSpecificationsGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermFitmentCatalogEdit), fitmentHubSpecificationHandler.UpdateStatus)
		fitmentHubSpecificationsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermFitmentCatalogDelete), fitmentHubSpecificationHandler.Delete)
	}

	productProfitabilityGroup := authenticated.Group("/procurement/profitability")
	productProfitabilityGroup.Use(middleware.RequirePermission(auth.PermProcurementView))
	{
		productProfitabilityGroup.GET("/by-codes", productProfitabilityHandler.ListByCodes)
		productProfitabilityGroup.POST("/preview", productProfitabilityHandler.Preview)
		productProfitabilityGroup.POST("/bulk-upsert", middleware.RequirePermission(auth.PermProcurementEdit), productProfitabilityHandler.BulkUpsert)
	}

	productSpecificationTemplatesGroup := authenticated.Group("/product-specification-templates")
	productSpecificationTemplatesGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		productSpecificationTemplatesGroup.GET("", productHandler.ListProductSpecificationTemplates)
		productSpecificationTemplatesGroup.GET("/:id", productHandler.GetProductSpecificationTemplate)
		productSpecificationTemplatesGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productHandler.CreateProductSpecificationTemplate)
		productSpecificationTemplatesGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateProductSpecificationTemplate)
		productSpecificationTemplatesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productHandler.DeleteProductSpecificationTemplate)
	}

	productBrandsGroup := authenticated.Group("/product-brands")
	productBrandsGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		productBrandsGroup.GET("", productBrandHandler.List)
		productBrandsGroup.GET("/:id", productBrandHandler.Get)
		productBrandsGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productBrandHandler.Create)
		productBrandsGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productBrandHandler.Update)
		productBrandsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productBrandHandler.Delete)
	}

	productCategoriesGroup := authenticated.Group("/product-categories")
	productCategoriesGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		productCategoriesGroup.GET("", productCategoryHandler.List)
		productCategoriesGroup.GET("/:id", productCategoryHandler.Get)
		productCategoriesGroup.GET("/:id/translations", productCategoryHandler.ListTranslations)
		productCategoriesGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productCategoryHandler.Create)
		productCategoriesGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productCategoryHandler.Update)
		productCategoriesGroup.PUT("/:id/translations", middleware.RequirePermission(auth.PermProductEdit), productCategoryHandler.UpdateTranslations)
		productCategoriesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productCategoryHandler.Delete)
	}

	productsGroup := authenticated.Group("/products")
	productsGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		productsGroup.GET("", productHandler.ListProducts)
		productsGroup.GET("/stats", productHandler.GetProductStats)
		productsGroup.GET("/customs-summary", productHandler.GetCustomsSummary)
		productsGroup.GET("/:id", productHandler.GetProduct)
		productsGroup.GET("/:id/translations", productHandler.GetProductTranslations)
		productsGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productHandler.CreateProduct)
		productsGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateProduct)
		productsGroup.POST("/:id/translations/copy", middleware.RequirePermission(auth.PermProductCreate), productHandler.CopyProductTranslation)
		productsGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateProductStatus)
		productsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productHandler.DeleteProduct)
		productsGroup.POST("/batch-status", middleware.RequirePermission(auth.PermProductEdit), productHandler.BatchUpdateStatus)
		productsGroup.POST("/batch-delete", middleware.RequirePermission(auth.PermProductDelete), productHandler.BatchDelete)
	}

	productInformationTemplatesGroup := authenticated.Group("/product-information-templates")
	productInformationTemplatesGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		productInformationTemplatesGroup.GET("", productInformationTemplateHandler.List)
		productInformationTemplatesGroup.GET("/:id", productInformationTemplateHandler.Get)
		productInformationTemplatesGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productInformationTemplateHandler.Create)
		productInformationTemplatesGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productInformationTemplateHandler.Update)
		productInformationTemplatesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productInformationTemplateHandler.Delete)
	}

	customsClassificationsGroup := authenticated.Group("/customs-classifications")
	customsClassificationsGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		customsClassificationsGroup.GET("", customsClassificationHandler.List)
		customsClassificationsGroup.GET("/lookup", customsClassificationHandler.Lookup)
		customsClassificationsGroup.GET("/:id", customsClassificationHandler.Get)
		customsClassificationsGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), customsClassificationHandler.Create)
		customsClassificationsGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), customsClassificationHandler.Update)
		customsClassificationsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), customsClassificationHandler.Delete)
	}

	spokeCatalogGroup := authenticated.Group("/spoke-catalog")
	spokeCatalogGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		spokeCatalogGroup.GET("", spokeCatalogHandler.Get)
		spokeCatalogGroup.PUT("", middleware.RequirePermission(auth.PermProductEdit), spokeCatalogHandler.Replace)
		spokeCatalogGroup.POST("/import", middleware.RequirePermission(auth.PermProductEdit), spokeCatalogHandler.Import)
		spokeCatalogGroup.GET("/preset-template", spokeCatalogHandler.DownloadPresetTemplate)
		spokeCatalogGroup.POST("/preset-template/import", middleware.RequirePermission(auth.PermProductEdit), spokeCatalogHandler.ImportPresetTemplate)
	}

	quickBuyGroup := authenticated.Group("/quick-buy")
	quickBuyGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		quickBuyGroup.GET("/flows", quickBuyHandler.ListFlows)
		quickBuyGroup.GET("/flows/:id", quickBuyHandler.GetFlow)
		quickBuyGroup.POST("/flows", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.CreateFlow)
		quickBuyGroup.PUT("/flows/:id", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.UpdateFlow)
		quickBuyGroup.PUT("/flows/:id/configuration", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.SaveFlowConfiguration)
		quickBuyGroup.POST("/flows/:id/draft", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.CreateDraftVersion)
		quickBuyGroup.PUT("/flow-versions/:version_id", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.UpdateDraftVersion)
		quickBuyGroup.POST("/flow-versions/:version_id/validate", quickBuyHandler.ValidateVersion)
		quickBuyGroup.POST("/flow-versions/:version_id/preview", quickBuyHandler.PreviewVersionStepCandidates)
		quickBuyGroup.POST("/flow-versions/:version_id/publish", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.PublishVersion)
	}

	selectionAssistantGroup := authenticated.Group("/selection-assistant")
	selectionAssistantGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		selectionAssistantGroup.GET("/flows", selectionAssistantHandler.ListFlows)
		selectionAssistantGroup.GET("/flows/:id", selectionAssistantHandler.GetFlow)
		selectionAssistantGroup.POST("/flows", middleware.RequirePermission(auth.PermProductEdit), selectionAssistantHandler.CreateFlow)
		selectionAssistantGroup.PUT("/flows/:id/configuration", middleware.RequirePermission(auth.PermProductEdit), selectionAssistantHandler.SaveFlowConfiguration)
		selectionAssistantGroup.POST("/flow-versions/:version_id/validate", selectionAssistantHandler.ValidateVersion)
		selectionAssistantGroup.POST("/flow-versions/:version_id/publish", middleware.RequirePermission(auth.PermProductEdit), selectionAssistantHandler.PublishVersion)
	}

	selectionConfigurationKeyGroup := authenticated.Group("/selection-configuration")
	selectionConfigurationKeyGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		selectionConfigurationKeyGroup.GET("/keys", selectionConfigurationKeyHandler.List)
		selectionConfigurationKeyGroup.GET("/keys/options", selectionConfigurationKeyHandler.ListOptions)
		selectionConfigurationKeyGroup.POST("/keys", middleware.RequirePermission(auth.PermProductEdit), selectionConfigurationKeyHandler.Create)
		selectionConfigurationKeyGroup.PUT("/keys/:id", middleware.RequirePermission(auth.PermProductEdit), selectionConfigurationKeyHandler.Update)
	}

	wheelsetFitQuestionnaireGroup := authenticated.Group("/wheelset-fit-questionnaire")
	wheelsetFitQuestionnaireGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		wheelsetFitQuestionnaireGroup.GET("/current", wheelsetFitQuestionnaireHandler.GetCurrentVersion)
		wheelsetFitQuestionnaireGroup.GET("/product-filter-options", wheelsetFitQuestionnaireHandler.GetProductFilterOptions)
		wheelsetFitQuestionnaireGroup.POST("/draft", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.CreateDraft)
		wheelsetFitQuestionnaireGroup.POST("/questions", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.CreateQuestion)
		wheelsetFitQuestionnaireGroup.PUT("/questions/order", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.ReorderQuestions)
		wheelsetFitQuestionnaireGroup.PUT("/questions/:id", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.UpdateQuestion)
		wheelsetFitQuestionnaireGroup.DELETE("/questions/:id", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.DeleteQuestion)
		wheelsetFitQuestionnaireGroup.POST("/versions/:version_id/validate", wheelsetFitQuestionnaireHandler.ValidateVersion)
		wheelsetFitQuestionnaireGroup.POST("/versions/:version_id/publish", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.PublishVersion)
	}

	attributesGroup := authenticated.Group("/attributes")
	attributesGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		attributesGroup.GET("", productHandler.ListAttributes)
		attributesGroup.GET("/:id", productHandler.GetAttribute)
		attributesGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productHandler.CreateAttribute)
		attributesGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateAttribute)
		attributesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productHandler.DeleteAttribute)

		// 属性值管理
		attributesGroup.GET("/:id/values", productHandler.GetAttributeValues)
		attributesGroup.POST("/:id/values", middleware.RequirePermission(auth.PermProductCreate), productHandler.CreateAttributeValue)
		attributesGroup.PUT("/:id/values/:valueId", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateAttributeValue)
		attributesGroup.DELETE("/:id/values/:valueId", middleware.RequirePermission(auth.PermProductDelete), productHandler.DeleteAttributeValue)
	}
}

func registerIntegrationRoutes(
	authenticated *gin.RouterGroup,
	googleMerchantHandler *GoogleMerchantHandler,
	socialOAuthHandler *SocialOAuthHandler,
) {
	googleMerchantGroup := authenticated.Group("/google-merchant")
	googleMerchantGroup.Use(middleware.RequirePermission(auth.PermMerchantView))
	{
		googleMerchantGroup.GET("/connection", googleMerchantHandler.GetConnection)
		googleMerchantGroup.PATCH("/connection", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.UpdateConnection)
		googleMerchantGroup.POST("/oauth/start", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.StartOAuth)
		googleMerchantGroup.POST("/disconnect", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.Disconnect)
		googleMerchantGroup.GET("/remote-products", googleMerchantHandler.ListRemoteProducts)
		googleMerchantGroup.GET("/offers", googleMerchantHandler.ListOffers)
		googleMerchantGroup.POST("/reconcile", middleware.RequirePermission(auth.PermMerchantSync), googleMerchantHandler.Reconcile)
		googleMerchantGroup.POST("/offers", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.CreateOffer)
		googleMerchantGroup.PUT("/offers/:id", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.UpdateOffer)
		googleMerchantGroup.POST("/offers/:id/validate", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.ValidateOffer)
		googleMerchantGroup.POST("/offers/:id/sync", middleware.RequirePermission(auth.PermMerchantSync), googleMerchantHandler.SyncOffer)
		googleMerchantGroup.POST("/offers/:id/remove-remote", middleware.RequirePermission(auth.PermMerchantSync), googleMerchantHandler.RemoveRemoteOffer)
		googleMerchantGroup.DELETE("/offers/:id", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.DeleteOffer)
	}

	socialOAuthGroup := authenticated.Group("/social/oauth")
	socialOAuthGroup.Use(middleware.RequirePermission(auth.PermSettingsView))
	{
		socialOAuthGroup.GET("", socialOAuthHandler.ListConnections)
		socialOAuthGroup.POST(
			"/:provider/start",
			middleware.RequirePermission(auth.PermSettingsEdit),
			socialOAuthHandler.Start,
		)
		socialOAuthGroup.DELETE(
			"/:provider",
			middleware.RequirePermission(auth.PermSettingsEdit),
			socialOAuthHandler.Disconnect,
		)
	}
}

func registerMediaAndPreflightRoutes(
	authenticated *gin.RouterGroup,
	mediaHandler *MediaHandler,
	mediaImageDimensionsHandler *MediaImageDimensionsHandler,
	fontPreflightHandler *FontPreflightHandler,
	contentLinkPreflightHandler *ContentLinkPreflightHandler,
	siteQualityHandler *SiteQualityHandler,
) {
	mediaGroup := authenticated.Group("/media")
	mediaGroup.Use(middleware.RequirePermission(auth.PermMediaView))
	{
		mediaGroup.GET("/derivative-presets", mediaHandler.ListDerivativePresets)
		mediaGroup.POST("/derivative-presets", middleware.RequirePermission(auth.PermMediaConfigure), mediaHandler.CreateDerivativePreset)
		mediaGroup.PUT("/derivative-presets/:id", middleware.RequirePermission(auth.PermMediaConfigure), mediaHandler.UpdateDerivativePreset)
		mediaGroup.PATCH("/derivative-presets/:id/enabled", middleware.RequirePermission(auth.PermMediaConfigure), mediaHandler.UpdateDerivativePresetEnabled)
		mediaGroup.DELETE("/derivative-presets/:id", middleware.RequirePermission(auth.PermMediaConfigure), mediaHandler.DeleteDerivativePreset)
		mediaGroup.GET("/derivative-rebuild-jobs", mediaHandler.ListDerivativeRebuildJobs)
		mediaGroup.POST("/derivative-rebuild-jobs", middleware.RequirePermission(auth.PermMediaConfigure), mediaHandler.RequestDerivativeRebuild)

		mediaGroup.GET("/assets", mediaHandler.ListAssets)
		mediaGroup.GET("/assets/:id/file", mediaHandler.ServeAssetFile)
		mediaGroup.GET("/assets/:id/copyright-evidence", middleware.RequirePermission(auth.PermMediaEdit), mediaHandler.ExportCopyrightEvidence)
		mediaGroup.GET("/assets/:id/references", mediaHandler.GetAssetReferences)
		mediaGroup.GET("/assets/:id", mediaHandler.GetAsset)
		mediaGroup.POST(
			"/assets",
			middleware.RequirePermission(auth.PermMediaCreate),
			middleware.RateLimitByUserPerMinute(3, 2),
			mediaHandler.UploadAsset,
		)
		mediaGroup.PATCH("/assets/:id", middleware.RequirePermission(auth.PermMediaEdit), mediaHandler.UpdateAsset)
		mediaGroup.DELETE("/assets/:id", middleware.RequirePermission(auth.PermMediaDelete), mediaHandler.DeleteAsset)
	}

	preflightGroup := authenticated.Group("/preflight")
	preflightGroup.Use(middleware.RequireAnyPermission(auth.PermServicesView, auth.PermMediaView))
	{
		siteQualityGroup := preflightGroup.Group("/site-quality")
		siteQualityGroup.Use(middleware.RequirePermission(auth.PermServicesView))
		registerSiteQualityRoutes(siteQualityGroup, siteQualityHandler, auth.PermServicesManage)

		imageDimensionsGroup := preflightGroup.Group("/image-dimensions")
		imageDimensionsGroup.Use(middleware.RequirePermission(auth.PermMediaView))
		registerMediaImageDimensionRoutes(imageDimensionsGroup, mediaImageDimensionsHandler)

		contentLinksGroup := preflightGroup.Group("/content-links")
		contentLinksGroup.Use(middleware.RequirePermission(auth.PermServicesView))
		registerPreflightContentLinkRoutes(contentLinksGroup, contentLinkPreflightHandler, auth.PermServicesManage)

		preflightGroup.GET("/fonts", middleware.RequirePermission(auth.PermServicesView), fontPreflightHandler.Get)
	}
}
