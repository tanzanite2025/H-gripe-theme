package localbackup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"hash"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/api/publicuploadresponse"
	"commerce-platform/internal/api/restrictedmediaresponse"
	publicticket "commerce-platform/internal/api/v1/ticket"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/pkg/securecookie"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/visitorcookie"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

// These are URL identities only. They must match the running business boundary;
// roots and signing keys always come from the independent recovery configuration.
type StoppedProductLocalMediaVerificationConfiguration struct {
	SharedBaseURL        string `json:"shared_base_url"`
	SharedPrivateBaseURL string `json:"shared_private_base_url"`
	LogoBaseURL          string `json:"logo_base_url"`
	LogoPrivateBaseURL   string `json:"logo_private_base_url"`
}

type StoppedProductReadOnlyMediaHTTPRequest struct {
	ContractSHA256       string `json:"contract_sha256"`
	Operation            string `json:"operation"`
	ObjectKey            string `json:"object_key"`
	ConversationID       string `json:"conversation_id"`
	MessageID            string `json:"message_id"`
	AttachmentIndex      int    `json:"attachment_index"`
	Preset               string `json:"preset"`
	Range                string `json:"range"`
	VisitorCookie        string `json:"visitor_cookie"`
	StorefrontAuthCookie string `json:"storefront_auth_cookie"`
	AdminAuthCookie      string `json:"admin_auth_cookie"`
}

type StoppedProductReadOnlyMediaHTTPResult struct {
	capturedPublicImageBytes     []byte
	Version                      int                                    `json:"version"`
	Scope                        string                                 `json:"scope"`
	TaskID                       string                                 `json:"task_id"`
	WindowID                     string                                 `json:"window_id"`
	ContractSHA256               string                                 `json:"contract_sha256"`
	DatabaseOID                  int64                                  `json:"database_oid"`
	DetailCacheProgressSHA256    string                                 `json:"detail_cache_progress_sha256"`
	Target                       StoppedProductReadOnlyMediaHTTPRequest `json:"target"` // Cookie is always cleared.
	VisitorCredentialSupplied    bool                                   `json:"visitor_credential_supplied"`
	StorefrontCredentialSupplied bool                                   `json:"storefront_credential_supplied"`
	AdminCredentialSupplied      bool                                   `json:"admin_credential_supplied"`
	HTTPStatus                   int                                    `json:"http_status"`
	Headers                      map[string]string                      `json:"headers"`
	ResponseBytes                int64                                  `json:"response_bytes"`
	ResponseSHA256               string                                 `json:"response_sha256"`
	ExpectedPackageFile          *FileIdentity                          `json:"expected_package_file,omitempty"`
	MatchesPackageFile           *bool                                  `json:"matches_package_file,omitempty"`
	SQLReadOnly                  bool                                   `json:"sql_read_only"`
	ApplicationRemainsStopped    bool                                   `json:"application_remains_stopped"`
	OnlineOverwriteAvailable     bool                                   `json:"online_overwrite_available"`
	Verification                 string                                 `json:"verification"`
	RemainingChecks              []string                               `json:"remaining_checks"`
	CapturedAt                   time.Time                              `json:"captured_at"`
}

var stoppedMediaMessageIDPattern = regexp.MustCompile(`^[1-9][0-9]{0,17}$`)
var stoppedMediaRangePattern = regexp.MustCompile(`^bytes=[0-9]{1,18}-[0-9]{0,18}$`)
var stoppedMediaSegmentPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func validateStoppedProductReadOnlyMediaHTTPRequest(request StoppedProductReadOnlyMediaHTTPRequest) error {
	if !isValidExternalProjectionEvidenceSHA256(request.ContractSHA256) || len(request.ObjectKey) > 1024 || len(request.ConversationID) > 128 || len(request.Preset) > 64 || len(request.VisitorCookie) > 512 || len(request.StorefrontAuthCookie) > 4096 || len(request.AdminAuthCookie) > 4096 || request.AttachmentIndex < 0 || request.AttachmentIndex > 999 || (request.Range != "" && !stoppedMediaRangePattern.MatchString(request.Range)) {
		return errors.New("a current contract and bounded media target are required")
	}
	for _, value := range []string{request.ObjectKey, request.ConversationID, request.MessageID, request.Preset, request.Range, request.VisitorCookie, request.StorefrontAuthCookie, request.AdminAuthCookie} {
		if strings.TrimSpace(value) != value || strings.ContainsFunc(value, unicode.IsControl) {
			return errors.New("media query fields must be trimmed and contain no control characters")
		}
	}
	for _, credential := range []string{request.VisitorCookie, request.StorefrontAuthCookie, request.AdminAuthCookie} {
		if strings.ContainsAny(credential, " ;,\"\\") || (&http.Cookie{Name: "business_cookie", Value: credential}).Valid() != nil { // #nosec G124 -- This is a request cookie or value validation, not a Set-Cookie response; browser response flags do not apply.
			return errors.New("only individual existing business cookie values are accepted")
		}
	}
	switch request.Operation {
	case "public_upload":
		key, ok := storage.NormalizeObjectKey(request.ObjectKey)
		if !ok || key != request.ObjectKey || strings.ContainsAny(key, "%?#") || request.ConversationID != "" || request.MessageID != "" || request.AttachmentIndex != 0 || request.Preset != "" || request.VisitorCookie != "" || request.StorefrontAuthCookie != "" || request.AdminAuthCookie != "" {
			return errors.New("public upload checks accept only an exact object key and optional byte range")
		}
	case "conversation_attachment", "agent_conversation_attachment":
		if request.ObjectKey != "" || !stoppedMediaSegmentPattern.MatchString(request.ConversationID) || !stoppedMediaMessageIDPattern.MatchString(request.MessageID) || (request.Preset != "" && !stoppedMediaSegmentPattern.MatchString(request.Preset)) {
			return errors.New("conversation checks require exact conversation/message/index identities")
		}
		if request.Operation == "conversation_attachment" && request.AdminAuthCookie != "" {
			return errors.New("customer attachment checks accept storefront and visitor cookies only")
		}
		if request.Operation == "agent_conversation_attachment" && (!stoppedMediaMessageIDPattern.MatchString(request.ConversationID) || request.VisitorCookie != "" || request.StorefrontAuthCookie != "") {
			return errors.New("agent checks require the numeric internal conversation ID and an optional existing admin cookie only")
		}
	default:
		return errors.New("only public_upload, conversation_attachment and agent_conversation_attachment checks are supported")
	}
	return nil
}

func (configuration *StoppedProductLocalMediaVerificationConfiguration) verifyRegisteredMediaURLIdentity(expected string) error {
	if configuration == nil {
		return errors.New("configure stopped_product_local_media_verification using the actual registered local media URL identities first")
	}
	for index, value := range []string{configuration.SharedBaseURL, configuration.SharedPrivateBaseURL, configuration.LogoBaseURL, configuration.LogoPrivateBaseURL} {
		if value == "" && (index == 1 || index == 3) {
			continue
		}
		parsed, err := url.Parse(value)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return errors.New("media verification URL identities must be explicit HTTP origins/paths without credentials or queries")
		}
	}
	actual := maintenancegate.FingerprintExternalProjectionConfiguration([]string{configuration.SharedBaseURL, configuration.SharedPrivateBaseURL, configuration.LogoBaseURL, configuration.LogoPrivateBaseURL})
	if actual != expected {
		return errors.New("local verification media URL identities differ from the registered business boundary")
	}
	return nil
}

func (manager *Manager) ExecuteStoppedProductReadOnlyMediaHTTPQuery(ctx context.Context, id string, request StoppedProductReadOnlyMediaHTTPRequest) (*StoppedProductReadOnlyMediaHTTPResult, error) {
	return manager.executeStoppedProductMediaHTTPQueryWithOptionalPublicImageCapture(ctx, id, request, false)
}

func (manager *Manager) executeStoppedProductMediaHTTPQueryWithOptionalPublicImageCapture(ctx context.Context, id string, request StoppedProductReadOnlyMediaHTTPRequest, capturePublicImage bool) (*StoppedProductReadOnlyMediaHTTPResult, error) {
	if err := validateStoppedProductReadOnlyMediaHTTPRequest(request); err != nil {
		return nil, err
	}
	if capturePublicImage && (request.Operation != "public_upload" || request.Range != "") {
		return nil, errors.New("image bytes are available only for complete shared public uploads")
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
		return nil, errors.New("read the current stopped product projection baseline before checking media")
	}
	if err := manager.Configuration.StoppedProductLocalMediaVerification.verifyRegisteredMediaURLIdentity(before.MediaDeliveryConfigurationSHA256); err != nil {
		return nil, err
	}
	var result *StoppedProductReadOnlyMediaHTTPResult
	err = manager.withBoundStoppedProductReadOnlyTransaction(ctx, manager.Configuration.DatabaseName, task.ProductCatalogPreviewBinding.Candidate, func(tx *sql.Tx) error {
		var invokeErr error
		result, invokeErr = manager.invokeStoppedProductMediaHandlersWithReadOnlyTransactionAndOptionalPublicImageCapture(ctx, tx, request, capturePublicImage)
		return invokeErr
	})
	if err != nil {
		return nil, err
	}
	after, err := manager.inspectStoppedProductStorefrontProjectionBaselineWithHeldLease(ctx, task)
	if err != nil || after.ContractSHA256 != before.ContractSHA256 {
		return nil, errors.New("SQL, media, cache or stopped boundary changed during media verification; result refused")
	}
	result.Version = 2
	result.Scope = productCatalogBackupScope
	result.TaskID = id
	result.WindowID = before.WindowID
	result.ContractSHA256 = before.ContractSHA256
	result.DatabaseOID = before.DatabaseOID
	result.DetailCacheProgressSHA256 = before.DetailCacheProgressSHA256
	result.VisitorCredentialSupplied = request.VisitorCookie != ""
	result.StorefrontCredentialSupplied = request.StorefrontAuthCookie != ""
	result.AdminCredentialSupplied = request.AdminAuthCookie != ""
	request.VisitorCookie = ""
	request.StorefrontAuthCookie = ""
	request.AdminAuthCookie = ""
	result.Target = request
	result.SQLReadOnly = true
	result.ApplicationRemainsStopped = true
	result.Verification = "in_process_business_local_media_http_handlers_active_account_role_assignment_read_only_no_network"
	result.CapturedAt = time.Now().UTC()
	result.RemainingChecks = []string{"target_environment_member_agent_sessions_and_all_attachment_permissions", "all_media_originals_derivatives_and_other_modules", "actual_network_origin_proxy_cdn_and_browser", "actual_nuxt_pages_route_index_sitemap", "final_business_ownership_acceptance_and_release"}
	if request.Operation == "public_upload" && !service.IsSiteLogoStorageKey(request.ObjectKey) {
		for _, file := range task.Manifest.Files {
			if file.Path == "media/uploads/"+request.ObjectKey {
				identity := file
				result.ExpectedPackageFile = &identity
				if result.HTTPStatus == http.StatusOK && request.Range == "" {
					matches := file.Size == result.ResponseBytes && file.SHA256 == result.ResponseSHA256
					result.MatchesPackageFile = &matches
				}
				break
			}
		}
	}
	return result, nil
}

// Hash complete handler output without retaining or returning attachment bytes.
type stoppedProductMediaHashingHTTPResponse struct {
	header          http.Header
	status          int
	bytes           int64
	digest          hash.Hash
	ctx             context.Context
	writeErr        error
	publicImageBody *bytes.Buffer
}

func (response *stoppedProductMediaHashingHTTPResponse) Header() http.Header { return response.header }
func (response *stoppedProductMediaHashingHTTPResponse) WriteHeader(status int) {
	if response.status == 0 {
		response.status = status
	}
}
func (response *stoppedProductMediaHashingHTTPResponse) Write(data []byte) (int, error) {
	if response.status == 0 {
		response.status = http.StatusOK
	}
	if response.writeErr != nil {
		return 0, response.writeErr
	}
	if err := response.ctx.Err(); err != nil {
		response.writeErr = err
		return 0, err
	}
	if response.bytes+int64(len(data)) > 64<<20 {
		response.writeErr = errors.New("media HTTP verification response exceeds 64 MiB")
		return 0, response.writeErr
	}
	if response.publicImageBody != nil && response.bytes+int64(len(data)) > stoppedProductPublicImageMaximumBytes {
		response.writeErr = errors.New("shared public image exceeds the 4 MiB preview limit")
		return 0, response.writeErr
	}
	n, err := response.digest.Write(data)
	if response.publicImageBody != nil {
		_, _ = response.publicImageBody.Write(data[:n])
	}
	response.bytes += int64(n)
	return n, err
}

func (manager *Manager) invokeStoppedProductMediaHandlersWithReadOnlyTransaction(ctx context.Context, tx *sql.Tx, request StoppedProductReadOnlyMediaHTTPRequest) (result *StoppedProductReadOnlyMediaHTTPResult, resultErr error) {
	return manager.invokeStoppedProductMediaHandlersWithReadOnlyTransactionAndOptionalPublicImageCapture(ctx, tx, request, false)
}

func (manager *Manager) invokeStoppedProductMediaHandlersWithReadOnlyTransactionAndOptionalPublicImageCapture(ctx context.Context, tx *sql.Tx, request StoppedProductReadOnlyMediaHTTPRequest, capturePublicImage bool) (result *StoppedProductReadOnlyMediaHTTPResult, resultErr error) {
	defer func() {
		if recover() != nil {
			result = nil
			resultErr = errors.New("read-only media handler failed; result refused")
		}
	}()
	if err := verifyStoppedProductSQLTransactionRemainsReadOnly(ctx, tx); err != nil {
		return nil, err
	}
	database, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, errors.New("cannot bind media checks to read-only SQL")
	}
	database = database.WithContext(ctx)
	configuration := manager.Configuration.StoppedProductLocalMediaVerification
	shared, err := storage.NewReadOnlyLocalRecoveryObjectStorageWithoutDirectoryCreation(storage.Config{Type: storage.StorageTypeLocal, LocalPath: manager.Configuration.StorageRoots["uploads"], PrivateLocalPath: manager.Configuration.StorageRoots["uploads_private"], BaseURL: configuration.SharedBaseURL, PrivateBaseURL: configuration.SharedPrivateBaseURL})
	if err != nil {
		return nil, err
	}
	mediaService := service.NewMediaService(repository.NewMediaRepository(database), shared, nil, configuration.SharedBaseURL, 0)
	router := gin.New()
	router.RedirectTrailingSlash = false
	target := "/uploads/" + request.ObjectKey
	if request.Operation == "public_upload" {
		access := service.NewPublicUploadAccessService(mediaService, nil)
		// The byte-delivery contract additionally binds the two site-identity
		// owners. Other dedicated owners, and the legacy digest-only contract,
		// retain their explicit unsupported boundary.
		siteIdentityImage := service.IsSiteLogoStorageKey(request.ObjectKey) || service.IsSiteFaviconStorageKey(request.ObjectKey)
		if (storage.DedicatedPublicObjectNamespaceForStorageKey(request.ObjectKey) != "" || service.IsCustomerServiceAvatarStorageKey(request.ObjectKey) || siteIdentityImage || service.IsHomeVisualTileStorageKey(request.ObjectKey)) && (!capturePublicImage || !siteIdentityImage) {
			return nil, errors.New("this check supports shared registered uploads; dedicated owner namespaces require their own verification")
		}
		var logoFileServer http.Handler
		if capturePublicImage {
			logoFileServer, err = manager.bindStoppedProductPublicImageSiteIdentityOwnersAndRestrictedLogoFileServer(database, shared, access)
			if err != nil {
				return nil, err
			}
		}
		router.GET("/uploads/*key", publicuploadresponse.NewRegisteredLocalUploadHTTPHandler(access, publicuploadresponse.NewRestrictedLocalUploadFileServer(manager.Configuration.StorageRoots["uploads"]), logoFileServer))
	} else {
		users := repository.NewUserRepository(database)
		tickets := service.NewTicketService(repository.NewTicketRepository(database), users)
		authentication := service.NewAuthService(users, config.JWTConfig{Secret: manager.Configuration.RecoverySecrets["JWT_SECRET"]})
		if request.Operation == "agent_conversation_attachment" {
			access := service.NewCustomerServiceConversationMessageAttachmentAccessService(tickets, mediaService)
			router.GET("/api/admin/customer-service/conversations/:id/messages/:message_id/attachments/:attachment_index/file",
				middleware.AuthMiddleware(authentication), middleware.RequireBackofficeAccess(), middleware.RequirePermission(auth.PermTicketView),
				restrictedmediaresponse.NewAgentCustomerServiceMessageAttachmentAuthorizedHTTPHandler(access))
			target = "/api/admin/customer-service/conversations/" + request.ConversationID
		} else {
			handler := publicticket.NewHandler(tickets, publicticket.Options{MediaService: mediaService, VisitorSecret: manager.Configuration.RecoverySecrets["JWT_SECRET"]})
			router.GET("/api/v1/customer-service/conversations/:conversation_id/messages/:message_id/attachments/:attachment_index/file", middleware.OptionalAuthMiddleware(authentication), handler.ServePublicCustomerServiceMessageAttachmentFile)
			target = "/api/v1/customer-service/conversations/" + request.ConversationID
		}
		target += "/messages/" + request.MessageID + "/attachments/" + strconv.Itoa(request.AttachmentIndex) + "/file"
		if request.Preset != "" {
			target += "?preset=" + url.QueryEscape(request.Preset)
		}
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, errors.New("cannot construct bounded media request")
	}
	if request.VisitorCookie != "" {
		httpRequest.AddCookie(&http.Cookie{Name: visitorcookie.CustomerServiceVisitorCookie, Value: request.VisitorCookie}) // #nosec G124 -- This is a request cookie or value validation, not a Set-Cookie response; browser response flags do not apply.
	}
	if request.StorefrontAuthCookie != "" {
		httpRequest.AddCookie(&http.Cookie{Name: securecookie.StorefrontAuthTokenCookie, Value: request.StorefrontAuthCookie}) // #nosec G124 -- This is a request cookie or value validation, not a Set-Cookie response; browser response flags do not apply.
	}
	if request.AdminAuthCookie != "" {
		httpRequest.AddCookie(&http.Cookie{Name: securecookie.AdminAuthTokenCookie, Value: request.AdminAuthCookie}) // #nosec G124 -- This is a request cookie or value validation, not a Set-Cookie response; browser response flags do not apply.
	}
	if request.Range != "" {
		httpRequest.Header.Set("Range", request.Range)
	}
	response := &stoppedProductMediaHashingHTTPResponse{header: make(http.Header), digest: sha256.New(), ctx: ctx}
	if capturePublicImage {
		if request.Operation != "public_upload" || request.Range != "" {
			return nil, errors.New("private or ranged media must never capture image bytes")
		}
		response.publicImageBody = &bytes.Buffer{}
	}
	router.ServeHTTP(response, httpRequest)
	if response.writeErr != nil {
		return nil, response.writeErr
	}
	if length := response.header.Get("Content-Length"); length != "" {
		expectedBytes, err := strconv.ParseInt(length, 10, 64)
		if err != nil || expectedBytes != response.bytes {
			return nil, errors.New("media handler returned an incomplete or invalid content length; result refused")
		}
	}
	if response.status >= 500 || response.status >= 300 && response.status < 400 || response.header.Get("Location") != "" || response.header.Get("Set-Cookie") != "" {
		return nil, errors.New("media handler failed, redirected or tried to set a cookie; result refused")
	}
	if err := verifyStoppedProductSQLTransactionRemainsReadOnly(ctx, tx); err != nil {
		return nil, err
	}
	headers := map[string]string{}
	for _, name := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "Cache-Control", "Vary", "X-Content-Type-Options", "Referrer-Policy"} {
		headers[name] = strings.Join(response.header.Values(name), ", ")
	}
	result = &StoppedProductReadOnlyMediaHTTPResult{HTTPStatus: response.status, Headers: headers, ResponseBytes: response.bytes, ResponseSHA256: hex.EncodeToString(response.digest.Sum(nil))}
	if capturePublicImage {
		if err := validateStoppedProductCapturedPublicImageResponse(response.status, headers, response.publicImageBody.Bytes()); err != nil {
			return nil, err
		}
		result.capturedPublicImageBytes = response.publicImageBody.Bytes()
	}
	return result, nil
}
