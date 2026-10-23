package localbackup

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	publicproduct "commerce-platform/internal/api/v1/product"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type StoppedProductReadOnlyPublicHTTPRequest struct {
	ContractSHA256 string `json:"contract_sha256"`
	Operation      string `json:"operation"`
	Locale         string `json:"locale"`
	Slug           string `json:"slug"`
	Keyword        string `json:"keyword"`
}

// This is an actual in-process invocation of public product handlers. It does
// not traverse a network, run Nuxt, authorize private attachments or test a CDN.
type StoppedProductReadOnlyPublicHTTPResult struct {
	Version                   int                                     `json:"version"`
	Scope                     string                                  `json:"scope"`
	TaskID                    string                                  `json:"task_id"`
	WindowID                  string                                  `json:"window_id"`
	ContractSHA256            string                                  `json:"contract_sha256"`
	DatabaseOID               int64                                   `json:"database_oid"`
	DetailCacheProgressSHA256 string                                  `json:"detail_cache_progress_sha256"`
	Request                   StoppedProductReadOnlyPublicHTTPRequest `json:"request"`
	HTTPStatus                int                                     `json:"http_status"`
	Response                  json.RawMessage                         `json:"response"`
	ResponseSHA256            string                                  `json:"response_sha256"`
	SQLReadOnly               bool                                    `json:"sql_read_only"`
	ApplicationRemainsStopped bool                                    `json:"application_remains_stopped"`
	OnlineOverwriteAvailable  bool                                    `json:"online_overwrite_available"`
	Verification              string                                  `json:"verification"`
	RemainingChecks           []string                                `json:"remaining_checks"`
	CapturedAt                time.Time                               `json:"captured_at"`
}

func validateStoppedProductReadOnlyPublicHTTPRequest(request StoppedProductReadOnlyPublicHTTPRequest) error {
	if !isValidExternalProjectionEvidenceSHA256(request.ContractSHA256) || len(request.Locale) < 1 || len(request.Locale) > 32 || len(request.Slug) > 512 || len(request.Keyword) > 200 {
		return errors.New("a current product projection contract and bounded explicit locale/query are required")
	}
	for _, value := range []string{request.Locale, request.Slug, request.Keyword} {
		if strings.TrimSpace(value) != value || strings.ContainsFunc(value, unicode.IsControl) {
			return errors.New("product query values must be trimmed and contain no control characters")
		}
	}
	for _, char := range request.Locale {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' && char != '_' {
			return errors.New("product query locale must be a simple explicit locale identity")
		}
	}
	switch request.Operation {
	case "detail":
		if request.Slug == "" || request.Keyword != "" {
			return errors.New("detail queries require only a slug and explicit locale")
		}
	case "list":
		if request.Slug != "" || request.Keyword != "" {
			return errors.New("list queries cannot carry detail or search arguments")
		}
	case "search":
		if request.Slug != "" || request.Keyword == "" {
			return errors.New("search queries require only a nonempty keyword and explicit locale")
		}
	default:
		return errors.New("only detail, first-page list and first-page search queries are available")
	}
	return nil
}

func (manager *Manager) ExecuteStoppedProductReadOnlyPublicHTTPQuery(ctx context.Context, id string, request StoppedProductReadOnlyPublicHTTPRequest) (*StoppedProductReadOnlyPublicHTTPResult, error) {
	if err := validateStoppedProductReadOnlyPublicHTTPRequest(request); err != nil {
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
		return nil, errors.New("the stopped product projection baseline changed; read it again before querying")
	}
	var status int
	var body []byte
	err = manager.withBoundStoppedProductReadOnlyTransaction(ctx, manager.Configuration.DatabaseName, task.ProductCatalogPreviewBinding.Candidate, func(tx *sql.Tx) error {
		var err error
		status, body, err = invokeStoppedProductPublicHandlerWithBoundReadOnlyTransaction(ctx, tx, request)
		return err
	})
	if err != nil {
		return nil, err
	}
	after, err := manager.inspectStoppedProductStorefrontProjectionBaselineWithHeldLease(ctx, task)
	if err != nil || after.ContractSHA256 != before.ContractSHA256 {
		return nil, errors.New("product SQL, media, cache or stopped boundary changed during the read-only query; result refused")
	}
	return &StoppedProductReadOnlyPublicHTTPResult{
		Version: 1, Scope: productCatalogBackupScope, TaskID: task.ID, WindowID: before.WindowID, ContractSHA256: before.ContractSHA256,
		DatabaseOID: before.DatabaseOID, DetailCacheProgressSHA256: before.DetailCacheProgressSHA256, Request: request,
		HTTPStatus: status, Response: body, ResponseSHA256: productCatalogBytesSHA256(body), SQLReadOnly: true, ApplicationRemainsStopped: true,
		Verification:    "in_process_public_product_http_handler_enforced_sql_read_only_no_network",
		RemainingChecks: []string{"all_product_routes_and_search_pages", "storefront_context_currency_shipping_reviews_and_breadcrumb_composition", "public_image_http_bytes_and_private_attachment_authorization", "actual_nuxt_html_route_index_sitemap_proxy_edge_and_browser", "final_business_ownership_acceptance_and_release"}, CapturedAt: time.Now().UTC(),
	}, nil
}

func (manager *Manager) withBoundStoppedProductReadOnlyTransaction(ctx context.Context, name string, expected productCatalogRecoveryDatabaseBinding, inspect func(*sql.Tx) error) (resultErr error) {
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	var oid int64
	var owner string
	var allowed bool
	if err := control.QueryRowContext(ctx, `SELECT oid::bigint,pg_get_userbyid(datdba),datallowconn FROM pg_database WHERE datname=$1`, name).Scan(&oid, &owner, &allowed); err != nil || oid != expected.OID || owner != expected.Owner || allowed {
		return errors.New("stopped product read-only database identity, ownership or sealed state changed")
	}
	if err := assertNoUnfencedDatabaseSessions(ctx, control, name); err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var actualOID int64
		if err := readDatabaseOID(cleanup, control, name, &actualOID); err != nil || actualOID != expected.OID {
			resultErr = errors.Join(resultErr, errors.New("stopped product read-only database cannot safely be resealed"))
			return
		}
		resultErr = errors.Join(resultErr, sealDatabase(cleanup, control, name))
	}()
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(name)+` ALLOW_CONNECTIONS true`); err != nil { // #nosec G701 -- Dynamic database, role, table and column identifiers use pq.QuoteIdentifier; data values remain bound parameters.
		return err
	}
	db, err := manager.Configuration.connectDatabase(ctx, name)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(db.Close)
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(tx.Rollback)
	var connectedOID int64
	var connectedOwner string
	if err := tx.QueryRowContext(ctx, `SELECT oid::bigint,pg_get_userbyid(datdba) FROM pg_database WHERE datname=current_database()`).Scan(&connectedOID, &connectedOwner); err != nil || connectedOID != expected.OID || connectedOwner != expected.Owner {
		return errors.New("read-only transaction connected to a changed product database identity")
	}
	if _, err := tx.ExecContext(ctx, `SET LOCAL statement_timeout='15s'`); err != nil {
		return err
	}
	if err := verifyStoppedProductSQLTransactionRemainsReadOnly(ctx, tx); err != nil {
		return err
	}
	if err := inspect(tx); err != nil {
		return err
	}
	return verifyStoppedProductSQLTransactionRemainsReadOnly(ctx, tx)
}

func verifyStoppedProductSQLTransactionRemainsReadOnly(ctx context.Context, tx *sql.Tx) error {
	var readOnly, isolation string
	if err := tx.QueryRowContext(ctx, `SELECT current_setting('transaction_read_only'),current_setting('transaction_isolation')`).Scan(&readOnly, &isolation); err != nil || readOnly != "on" || isolation != "repeatable read" {
		return errors.New("product HTTP queries require an active enforced read-only repeatable-read transaction")
	}
	return nil
}

// The response limit applies while the handler writes, before retaining its
// body. There is no caller-supplied route, method, SQL, DSN or network address.
type stoppedProductBoundedPublicHTTPResponse struct {
	header   http.Header
	status   int
	body     bytes.Buffer
	overflow bool
}

func (response *stoppedProductBoundedPublicHTTPResponse) Header() http.Header { return response.header }
func (response *stoppedProductBoundedPublicHTTPResponse) WriteHeader(status int) {
	if response.status == 0 {
		response.status = status
	}
}
func (response *stoppedProductBoundedPublicHTTPResponse) Write(data []byte) (int, error) {
	if response.status == 0 {
		response.status = http.StatusOK
	}
	if response.body.Len()+len(data) > 4<<20 {
		response.overflow = true
		return 0, errors.New("public product HTTP response exceeds 4 MiB")
	}
	return response.body.Write(data)
}

func invokeStoppedProductPublicHandlerWithBoundReadOnlyTransaction(ctx context.Context, tx *sql.Tx, request StoppedProductReadOnlyPublicHTTPRequest) (status int, body []byte, resultErr error) {
	defer func() {
		if recover() != nil {
			status, body, resultErr = 0, nil, errors.New("read-only product handler failed; result refused and database resealing required")
		}
	}()
	if err := verifyStoppedProductSQLTransactionRemainsReadOnly(ctx, tx); err != nil {
		return 0, nil, err
	}
	database, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return 0, nil, errors.New("cannot bind public product handler to the read-only transaction")
	}
	productService := service.NewProductServiceForReadOnlyRecoveryVerification(repository.NewProductRepository(database.WithContext(ctx)))
	handler := publicproduct.NewHandler(productService)
	response := &stoppedProductBoundedPublicHTTPResponse{header: make(http.Header)}
	query := url.Values{"page": {"1"}, "page_size": {"24"}}
	path := "/api/v1/products"
	if request.Operation == "search" {
		path = "/api/v1/customer-service/products"
		query.Set("keyword", request.Keyword)
	}
	if request.Operation == "detail" {
		path += "/" + url.PathEscape(request.Slug)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, path+"?"+query.Encode(), nil)
	if err != nil {
		return 0, nil, errors.New("cannot construct bounded public product HTTP request")
	}
	router := gin.New()
	router.UseRawPath = true
	router.Use(func(context *gin.Context) { context.Set("locale", request.Locale); context.Next() })
	router.GET("/api/v1/products", handler.ListProducts)
	router.GET("/api/v1/products/:id", handler.GetProduct)
	router.GET("/api/v1/customer-service/products", handler.ListPublicChatProducts)
	router.ServeHTTP(response, httpRequest)
	if response.overflow || !json.Valid(response.body.Bytes()) || response.status >= 500 {
		return 0, nil, errors.New("public product handler returned an oversized, invalid or failed response; no verification result accepted")
	}
	if err := verifyStoppedProductSQLTransactionRemainsReadOnly(ctx, tx); err != nil {
		return 0, nil, err
	}
	return response.status, append([]byte{}, response.body.Bytes()...), nil
}
