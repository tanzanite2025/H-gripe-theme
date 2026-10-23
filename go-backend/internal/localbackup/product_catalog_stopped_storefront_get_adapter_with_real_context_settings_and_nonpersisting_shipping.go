package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	publicproduct "commerce-platform/internal/api/v1/product"
	publicsettings "commerce-platform/internal/api/v1/settings"
	publicstorefront "commerce-platform/internal/api/v1/storefront"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type StoppedProductReadOnlyStorefrontHTTPRequest struct {
	ContractSHA256 string `json:"contract_sha256"`
	Operation      string `json:"operation"`
	Locale         string `json:"locale"`
	Country        string `json:"country"`
	Currency       string `json:"currency"`
	Slug           string `json:"slug"`
	Keyword        string `json:"keyword"`
	ProductID      uint   `json:"product_id,omitempty"`
	Page           int    `json:"page,omitempty"`
	PageSize       int    `json:"page_size,omitempty"`
}

// The original bytes are transported as base64 so a loopback bridge can check
// their digest without reserializing JSON numbers, escapes or property order.
type StoppedProductReadOnlyStorefrontHTTPResult struct {
	Version                   int                                         `json:"version"`
	Scope                     string                                      `json:"scope"`
	TaskID                    string                                      `json:"task_id"`
	WindowID                  string                                      `json:"window_id"`
	ContractSHA256            string                                      `json:"contract_sha256"`
	DatabaseOID               int64                                       `json:"database_oid"`
	DetailCacheProgressSHA256 string                                      `json:"detail_cache_progress_sha256"`
	Request                   StoppedProductReadOnlyStorefrontHTTPRequest `json:"request"`
	HTTPStatus                int                                         `json:"http_status"`
	ResponseBytesBase64       []byte                                      `json:"response_bytes_base64"`
	ResponseSHA256            string                                      `json:"response_sha256"`
	SQLReadOnly               bool                                        `json:"sql_read_only"`
	ApplicationRemainsStopped bool                                        `json:"application_remains_stopped"`
	OnlineOverwriteAvailable  bool                                        `json:"online_overwrite_available"`
	Verification              string                                      `json:"verification"`
	RemainingChecks           []string                                    `json:"remaining_checks"`
	CapturedAt                time.Time                                   `json:"captured_at"`
}

var stoppedStorefrontExplicitCountryPattern = regexp.MustCompile(`^[A-Z]{2}$`)
var stoppedStorefrontExplicitCurrencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

func validateStoppedProductReadOnlyStorefrontHTTPRequest(request StoppedProductReadOnlyStorefrontHTTPRequest) error {
	base := StoppedProductReadOnlyPublicHTTPRequest{ContractSHA256: request.ContractSHA256, Operation: request.Operation, Locale: request.Locale, Slug: request.Slug, Keyword: request.Keyword}
	switch request.Operation {
	case "context", "site_settings", "categories":
		base.Operation = "list"
	}
	if _, supported := stoppedStorefrontPageShellReadOnlyGETPaths[request.Operation]; supported {
		base.Operation = "list"
	}
	if isStoppedStorefrontProductReviewOperation(request.Operation) {
		base.Operation = "list"
	}
	if err := validateStoppedProductReadOnlyPublicHTTPRequest(base); err != nil {
		return err
	}
	if !stoppedStorefrontExplicitCountryPattern.MatchString(request.Country) || !stoppedStorefrontExplicitCurrencyPattern.MatchString(request.Currency) {
		return errors.New("storefront checks require an explicit uppercase country and display currency")
	}
	return validateStoppedStorefrontProductReviewSelectorLimits(request)
}

func (manager *Manager) ExecuteStoppedProductReadOnlyStorefrontHTTPQuery(ctx context.Context, id string, request StoppedProductReadOnlyStorefrontHTTPRequest) (*StoppedProductReadOnlyStorefrontHTTPResult, error) {
	if err := validateStoppedProductReadOnlyStorefrontHTTPRequest(request); err != nil {
		return nil, err
	}
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("another maintenance operation is running")
	}
	defer manager.operationMutex.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	before, err := manager.inspectStoppedProductStorefrontProjectionBaselineWithHeldLease(ctx, task)
	if err != nil || before.ContractSHA256 != request.ContractSHA256 {
		return nil, errors.New("read the current stopped product baseline before querying the storefront")
	}
	if err := manager.Configuration.StoppedProductLocalMediaVerification.verifyRegisteredMediaURLIdentity(before.MediaDeliveryConfigurationSHA256); err != nil {
		return nil, err
	}
	var status int
	var body []byte
	err = manager.withBoundStoppedProductReadOnlyTransaction(ctx, manager.Configuration.DatabaseName, task.ProductCatalogPreviewBinding.Candidate, func(tx *sql.Tx) error {
		var invokeError error
		status, body, invokeError = manager.invokeStoppedProductStorefrontGETWithBoundReadOnlyTransaction(ctx, tx, request)
		return invokeError
	})
	if err != nil {
		return nil, err
	}
	after, err := manager.inspectStoppedProductStorefrontProjectionBaselineWithHeldLease(ctx, task)
	if err != nil || after.ContractSHA256 != before.ContractSHA256 {
		return nil, errors.New("storefront query changed or lost the stopped SQL/media/cache boundary; result refused")
	}
	return &StoppedProductReadOnlyStorefrontHTTPResult{
		Version: 1, Scope: productCatalogBackupScope, TaskID: task.ID, WindowID: before.WindowID, ContractSHA256: before.ContractSHA256,
		DatabaseOID: before.DatabaseOID, DetailCacheProgressSHA256: before.DetailCacheProgressSHA256, Request: request,
		HTTPStatus: status, ResponseBytesBase64: body, ResponseSHA256: productCatalogBytesSHA256(body), SQLReadOnly: true, ApplicationRemainsStopped: true,
		Verification:    "real_storefront_get_handlers_bound_read_only_sql_without_checkout_snapshots_or_network",
		RemainingChecks: []string{"remaining_page_api_dependencies_and_all_routes", "actual_nuxt_build_html_hydration_and_media_delivery", "actual_proxy_cdn_old_cache_and_new_browser", "final_business_ownership_acceptance_and_release"}, CapturedAt: time.Now().UTC(),
	}, nil
}

func (manager *Manager) invokeStoppedProductStorefrontGETWithBoundReadOnlyTransaction(ctx context.Context, tx *sql.Tx, request StoppedProductReadOnlyStorefrontHTTPRequest) (status int, body []byte, resultError error) {
	defer func() {
		if recover() != nil {
			status, body, resultError = 0, nil, errors.New("stopped storefront handler failed; result refused")
		}
	}()
	if err := verifyStoppedProductSQLTransactionRemainsReadOnly(ctx, tx); err != nil {
		return 0, nil, err
	}
	database, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return 0, nil, errors.New("cannot bind storefront handlers to read-only SQL")
	}
	database = database.WithContext(ctx)
	// Some normal handlers omit optional fields on query failure. Such a failure
	// must not become accepted recovery evidence. RecordNotFound still retains
	// normal business behavior (missing product, absent optional settings/rates).
	var observedDatabaseError error
	if err := database.Callback().Query().After("gorm:query").Register("stopped_storefront_reject_hidden_query_errors", func(query *gorm.DB) {
		if query.Error != nil && !errors.Is(query.Error, gorm.ErrRecordNotFound) {
			observedDatabaseError = query.Error
		}
	}); err != nil {
		return 0, nil, err
	}
	configuration := manager.Configuration.StoppedProductLocalMediaVerification
	shared, err := storage.NewReadOnlyLocalRecoveryObjectStorageWithoutDirectoryCreation(storage.Config{Type: storage.StorageTypeLocal, LocalPath: manager.Configuration.StorageRoots["uploads"], PrivateLocalPath: manager.Configuration.StorageRoots["uploads_private"], BaseURL: configuration.SharedBaseURL, PrivateBaseURL: configuration.SharedPrivateBaseURL})
	if err != nil {
		return 0, nil, err
	}
	logos, err := storage.NewReadOnlyLocalRecoveryObjectStorageWithoutDirectoryCreation(storage.Config{Type: storage.StorageTypeLocal, LocalPath: manager.Configuration.StorageRoots["site_logo_uploads"], PrivateLocalPath: manager.Configuration.StorageRoots["uploads_private"], BaseURL: configuration.LogoBaseURL, PrivateBaseURL: configuration.LogoPrivateBaseURL})
	if err != nil {
		return 0, nil, err
	}
	settingsRepository := repository.NewSettingRepository(database)
	settingsService := service.NewSettingService(settingsRepository, nil, 0)
	mediaService := service.NewMediaService(repository.NewMediaRepository(database), shared, nil, configuration.SharedBaseURL, 0)
	policy := service.NewCurrencyPolicyService(settingsRepository)
	storefrontContext := service.NewStorefrontContextServiceWithMarkets(policy, service.NewStorefrontMarketService(repository.NewStorefrontMarketRepository(database)))
	// Resolve explicitly first: the normal product handler tolerates context
	// failures, which would otherwise hide a broken market/currency dependency.
	resolved, err := storefrontContext.Resolve(service.StorefrontContextInput{Country: request.Country, CountrySource: "request", RequestedLocale: request.Locale, RequestedCurrency: request.Currency})
	if err != nil || resolved == nil {
		return 0, nil, errors.New("real storefront market/currency context could not be resolved")
	}
	productRepository := repository.NewProductRepository(database)
	products := service.NewProductServiceForReadOnlyRecoveryVerification(productRepository)
	products.ConfigureCurrencyPolicy(policy)
	shipping := service.NewShippingServiceForReadOnlyRecoveryVerification(repository.NewShippingRepository(database), productRepository)
	shipping.ConfigureCurrencyPolicy(policy)
	shipping.ConfigureExchangeRateService(service.NewExchangeRateService(repository.NewExchangeRateRepository(database), settingsRepository))
	productHandler := publicproduct.NewHandler(products)
	productHandler.ConfigureStorefrontContext(storefrontContext)
	productHandler.ConfigureMediaService(mediaService)
	productHandler.ConfigureProductCategoryService(service.NewProductCategoryService(repository.NewProductCategoryRepository(database)))
	productHandler.ConfigureReviewService(service.NewReviewServiceForReadOnlyRecoveryVerification(repository.NewReviewRepository(database)))
	productHandler.ConfigureShippingService(shipping)
	settingsHandler := publicsettings.NewHandler(settingsService, service.NewWebsiteProfileService(settingsService))
	settingsHandler.ConfigureMediaService(mediaService)
	settingsHandler.ConfigureSiteLogoService(service.NewSiteLogoService(repository.NewSiteLogoRepository(database), logos, ""))
	favicon := service.NewSiteFaviconService(repository.NewSiteFaviconRepository(database), shared, "")
	favicon.ConfigureSettingService(settingsService)
	settingsHandler.ConfigureSiteFaviconService(favicon)
	router := gin.New()
	router.RedirectTrailingSlash = false
	router.UseRawPath = true
	router.Use(func(c *gin.Context) { c.Set("locale", request.Locale); c.Next() })
	router.GET("/api/v1/storefront/context", publicstorefront.NewContextHandler(storefrontContext).GetContext)
	router.GET("/api/v1/settings/site", settingsHandler.GetSiteSettings)
	router.GET("/api/v1/products/categories", productHandler.ListCategories)
	router.GET("/api/v1/products", productHandler.ListProducts)
	router.GET("/api/v1/products/:id", productHandler.GetProduct)
	router.GET("/api/v1/customer-service/products", productHandler.ListPublicChatProducts)
	bindStoppedProductReadOnlyStorefrontPageShellDatabaseHandlers(router, database, settingsService, settingsHandler, productHandler, policy, mediaService)
	bindStoppedStorefrontReadOnlyProductReviewHandlers(router, database, mediaService)
	path := map[string]string{"context": "/api/v1/storefront/context", "site_settings": "/api/v1/settings/site", "categories": "/api/v1/products/categories", "list": "/api/v1/products", "search": "/api/v1/customer-service/products"}[request.Operation]
	if dependencyPath, supported := stoppedStorefrontPageShellReadOnlyGETPaths[request.Operation]; supported {
		path = dependencyPath
	}
	query := url.Values{"locale": {request.Locale}, "country": {request.Country}, "currency": {request.Currency}, "page": {"1"}, "page_size": {"24"}}
	if request.Operation == "detail" {
		path = "/api/v1/products/" + url.PathEscape(request.Slug)
	}
	if request.Operation == "search" {
		query.Set("keyword", request.Keyword)
	}
	if request.Operation == "product_page_faq" {
		query.Set("route_path", "/products/:slug")
	}
	if request.Operation == "product_reviews" {
		path = "/api/v1/reviews"
		query.Set("product_id", strconv.FormatUint(uint64(request.ProductID), 10))
		query.Set("page", strconv.Itoa(request.Page))
		query.Set("page_size", strconv.Itoa(request.PageSize))
	}
	if request.Operation == "product_review_summary" {
		path = "/api/v1/reviews/summary/" + strconv.FormatUint(uint64(request.ProductID), 10)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, path+"?"+query.Encode(), nil)
	if err != nil {
		return 0, nil, err
	}
	response := &stoppedProductBoundedPublicHTTPResponse{header: make(http.Header)}
	router.ServeHTTP(response, httpRequest)
	if observedDatabaseError != nil || response.overflow || !json.Valid(response.body.Bytes()) || (response.status != http.StatusOK && response.status != http.StatusNotFound) || response.header.Get("Set-Cookie") != "" || response.header.Get("Location") != "" {
		return 0, nil, errors.New("real storefront handler failed, hid a SQL error or returned an invalid response; no result accepted")
	}
	if err := verifyStoppedProductSQLTransactionRemainsReadOnly(ctx, tx); err != nil {
		return 0, nil, err
	}
	return response.status, append([]byte(nil), response.body.Bytes()...), nil
}
