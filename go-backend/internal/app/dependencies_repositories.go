package app

import (
	"fmt"

	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"gorm.io/gorm"
)

func newDependencyRepositories(db *gorm.DB) (Repositories, error) {
	fitmentFrameHubSpecificationRepository := repository.NewFitmentFrameHubSpecificationRepository(db)
	fitmentForkHubSpecificationRepository := repository.NewFitmentForkHubSpecificationRepository(db)
	fitmentHubSpecificationRepository := repository.NewFitmentHubSpecificationRepository(db, fitmentFrameHubSpecificationRepository)
	fitmentHubSpecificationRepository.ConfigureForkHubSpecificationRepository(fitmentForkHubSpecificationRepository)

	postRepo := repository.NewPostRepository(db)
	blogCategoryRepo := repository.NewBlogCategoryRepository(db)
	postRepo.ConfigureCategoryRepository(blogCategoryRepo)

	repos := Repositories{
		User:                           repository.NewUserRepository(db),
		Post:                           postRepo,
		BlogCategory:                   blogCategoryRepo,
		StorefrontRouteCatalog:         repository.NewStorefrontRouteCatalogRepository(db),
		StorefrontURLSearchProfiles:    repository.NewStorefrontURLSearchProfileRepository(db),
		StorefrontRedirectRules:        repository.NewStorefrontRedirectRuleRepository(db),
		StorefrontURLIssues:            repository.NewStorefrontURLIssueRepository(db),
		PreflightContentLinks:          repository.NewPreflightContentLinkRepository(db),
		SiteQualityTargets:             repository.NewSiteQualityTargetRepository(db),
		SiteQualityJobs:                repository.NewSiteQualityJobRepository(db),
		SiteQualityRuns:                repository.NewSiteQualityRunRepository(db),
		SiteQualityFindings:            repository.NewSiteQualityFindingRepository(db),
		HotDataArchive:                 repository.NewHotDataArchiveRepository(db),
		Product:                        repository.NewProductRepository(db),
		ProductSupplierCostRecord:      repository.NewProductSupplierCostRecordRepository(db),
		FrameFitmentEntry:              repository.NewFrameFitmentEntryRepository(db),
		ForkFitmentEntry:               repository.NewForkFitmentEntryRepository(db),
		FitmentFrameHubSpecification:   fitmentFrameHubSpecificationRepository,
		FitmentForkHubSpecification:    fitmentForkHubSpecificationRepository,
		FitmentHubSpecification:        fitmentHubSpecificationRepository,
		ProductSupplierCostCatalog:     repository.NewProductSupplierCostCatalogRepository(db),
		ProductProfitCalculation:       repository.NewProductProfitCalculationRepository(db),
		ProductCategory:                repository.NewProductCategoryRepository(db),
		ProductBrand:                   repository.NewProductBrandRepository(db),
		ProductInformationTemplate:     repository.NewProductInformationTemplateRepository(db),
		CustomsClassification:          repository.NewCustomsClassificationRepository(db),
		Cart:                           repository.NewCartRepository(db),
		Setting:                        repository.NewSettingRepository(db),
		FAQ:                            repository.NewFAQRepository(db),
		Order:                          repository.NewOrderRepository(db),
		AfterSales:                     repository.NewAfterSalesCaseRepository(db),
		AfterSalesRefundReview:         repository.NewAfterSalesRefundReviewRepository(db),
		OrderIdempotency:               repository.NewOrderIdempotencyRepository(db),
		OrderPolicyDisclosure:          repository.NewOrderPolicyDisclosureRepository(db),
		OrderAttribution:               repository.NewOrderAttributionRepository(db),
		ProductQualityRequirement:      repository.NewProductQualityRequirementRepository(db),
		OrderEvidenceSnapshot:          repository.NewOrderEvidenceSnapshotRepository(db),
		OrderEvidence:                  repository.NewOrderEvidenceRepository(db),
		OrderEvidenceSubmission:        repository.NewOrderEvidenceSubmissionSnapshotRepository(db),
		OrderEvidenceExport:            repository.NewOrderEvidenceExportSnapshotRepository(db),
		Payment:                        repository.NewPaymentRepository(db),
		PaymentOperationIdempotency:    repository.NewPaymentOperationIdempotencyRepository(db),
		PaymentRisk:                    repository.NewPaymentRiskRepository(db),
		PaymentProtection:              repository.NewPaymentProtectionRepository(db),
		PaymentRefundReview:            repository.NewPaymentRefundRecommendationRepository(db),
		PaymentRefundExec:              repository.NewPaymentRefundExecutionRepository(db),
		PaymentRefundIdempotency:       repository.NewPaymentRefundIdempotencyRepository(db),
		ExchangeRate:                   repository.NewExchangeRateRepository(db),
		Shipping:                       repository.NewShippingRepository(db),
		Coupon:                         repository.NewCouponRepository(db),
		Loyalty:                        repository.NewLoyaltyRepository(db),
		LoyaltyProgram:                 repository.NewLoyaltyProgramRepository(db),
		Referral:                       repository.NewReferralRepository(db),
		ReferralProgram:                repository.NewReferralProgramRepository(db),
		Review:                         repository.NewReviewRepository(db),
		Ticket:                         repository.NewTicketRepository(db),
		CustomerServiceRetentionPolicy: repository.NewCustomerServiceRetentionPolicyRepository(db),
		Gallery:                        repository.NewGalleryRepository(db),
		Media:                          repository.NewMediaRepository(db),
		SiteLogo:                       repository.NewSiteLogoRepository(db),
		SiteFavicon:                    repository.NewSiteFaviconRepository(db),
		MediaDerivativePresets:         repository.NewMediaDerivativePresetRepository(db),
		MediaDerivativeRebuildJobs:     repository.NewMediaDerivativeRebuildJobRepository(db),
		StorefrontMarket:               repository.NewStorefrontMarketRepository(db),
		OpsDomainBinding:               repository.NewOpsDomainBindingRepository(db),
		OpsConnector:                   repository.NewOpsConnectorRepository(db),
		OpsConnectorOAuth:              repository.NewOpsConnectorOAuthRepository(db),
		OpsVPSBinding:                  repository.NewOpsVPSBindingRepository(db),
		OpsProjectBinding:              repository.NewOpsProjectBindingRepository(db),
		OpsNetworkRule:                 repository.NewOpsNetworkRuleRepository(db),
		GlobalIPBlockRule:              repository.NewGlobalIPBlockRuleRepository(db),
		OpsDeploymentWorkflow:          repository.NewOpsDeploymentWorkflowRepository(db),
		GoogleMerchant:                 repository.NewGoogleMerchantRepository(db),
		SocialOAuth:                    repository.NewSocialOAuthRepository(db),
		Warranty:                       repository.NewWarrantyRepository(db),
		ShipmentRecord:                 repository.NewShipmentRecordRepository(db),
		Audit:                          repository.NewAuditRepository(db),
		UGCShowcase:                    repository.NewUGCShowcaseRepository(db),
		HomeVisualTiles:                repository.NewHomeVisualTileRepository(db),
		Wishlist:                       repository.NewWishlistRepository(db),
		Feedback:                       repository.NewFeedbackRepository(db),
		SuggestionFeedback:             repository.NewSuggestionFeedbackRepository(db),
		Spoke:                          repository.NewSpokeRepository(db, fitmentHubSpecificationRepository),
		QuickBuy:                       repository.NewQuickBuyRepository(db),
		SelectionAssistant:             repository.NewSelectionAssistantRepository(db),
		SelectionConfigurationKey:      repository.NewSelectionConfigurationKeyRepository(db),
		WheelsetFitQuestionnaire:       repository.NewWheelsetFitQuestionnaireRepository(db),
		Subscription:                   repository.NewSubscriptionRepository(db),
		EmailChallenge:                 repository.NewEmailChallengeRepository(db),
		NotificationTemplates:          repository.NewNotificationTemplateRepository(db),
		EmailProviders:                 repository.NewEmailProviderRepository(db),
		VisitorProfile:                 repository.NewVisitorProfileRepository(db),
		RecommendationEvent:            repository.NewRecommendationEventRepository(db),
		VisitorRiskFact:                repository.NewVisitorRiskFactRepository(db),
		Outbox:                         repository.NewOutboxRepository(db),
	}

	repos.Spoke.ConfigureProductBrandRepository(repos.ProductBrand)
	repos.StorefrontRouteCatalog.ConfigureOutbox(repos.Outbox)
	if err := service.SeedDefaultMediaDerivativePresets(repos.MediaDerivativePresets); err != nil {
		return Repositories{}, fmt.Errorf("seed media derivative presets: %w", err)
	}

	return repos, nil
}
