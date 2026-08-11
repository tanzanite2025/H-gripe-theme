package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/merchant"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/secretbox"

	"gorm.io/gorm"
)

const (
	googleMerchantOAuthAuthorizationEndpoint = "https://accounts.google.com/o/oauth2/v2/auth"
	googleMerchantOAuthTokenEndpoint         = "https://oauth2.googleapis.com/token"
	googleMerchantUserInfoEndpoint           = "https://openidconnect.googleapis.com/v1/userinfo"
	googleMerchantProductsEndpoint           = "https://merchantapi.googleapis.com/products/v1"
	googleMerchantContentScope               = "https://www.googleapis.com/auth/content"
)

var (
	ErrGoogleMerchantConnectionNotFound = errors.New("google merchant connection not found")
	ErrGoogleMerchantOAuthNotConfigured = errors.New("google merchant oauth is not configured")
	ErrGoogleMerchantConnectionInvalid  = errors.New("google merchant connection settings invalid")
	ErrGoogleMerchantOAuthStateInvalid  = errors.New("google merchant oauth state invalid or expired")
	ErrGoogleMerchantOAuthExchange      = errors.New("google merchant oauth exchange failed")
	ErrGoogleMerchantRemoteAPI          = errors.New("google merchant remote api request failed")
)

type GoogleMerchantConnectionView struct {
	Configured                bool       `json:"configured"`
	OAuthConfigured           bool       `json:"oauth_configured"`
	TokenEncryptionConfigured bool       `json:"token_encryption_configured"`
	Connected                 bool       `json:"connected"`
	Status                    string     `json:"status"`
	GoogleAccountEmail        string     `json:"google_account_email"`
	MerchantAccountID         string     `json:"merchant_account_id"`
	DataSourceID              string     `json:"data_source_id"`
	StorefrontBaseURL         string     `json:"storefront_base_url"`
	LastConnectedAt           *time.Time `json:"last_connected_at,omitempty"`
	LastError                 string     `json:"last_error,omitempty"`
}

type GoogleMerchantConnectionInput struct {
	MerchantAccountID string `json:"merchant_account_id"`
	DataSourceID      string `json:"data_source_id"`
	StorefrontBaseURL string `json:"storefront_base_url"`
}

type GoogleMerchantRemotePrice struct {
	AmountMicros string `json:"amount_micros"`
	CurrencyCode string `json:"currency_code"`
}

type GoogleMerchantRemoteAttributes struct {
	Title                 string                     `json:"title"`
	Link                  string                     `json:"link"`
	ImageLink             string                     `json:"image_link"`
	Availability          string                     `json:"availability"`
	Brand                 string                     `json:"brand"`
	Condition             string                     `json:"condition"`
	GoogleProductCategory string                     `json:"google_product_category"`
	Price                 *GoogleMerchantRemotePrice `json:"price,omitempty"`
	SalePrice             *GoogleMerchantRemotePrice `json:"sale_price,omitempty"`
}

type GoogleMerchantRemoteDestinationStatus struct {
	ReportingContext     string   `json:"reporting_context"`
	ApprovedCountries    []string `json:"approved_countries,omitempty"`
	PendingCountries     []string `json:"pending_countries,omitempty"`
	DisapprovedCountries []string `json:"disapproved_countries,omitempty"`
}

type GoogleMerchantRemoteIssue struct {
	Code                string   `json:"code"`
	Severity            string   `json:"severity"`
	Resolution          string   `json:"resolution"`
	Attribute           string   `json:"attribute"`
	ReportingContext    string   `json:"reporting_context"`
	Description         string   `json:"description"`
	Detail              string   `json:"detail"`
	Documentation       string   `json:"documentation"`
	ApplicableCountries []string `json:"applicable_countries,omitempty"`
}

type GoogleMerchantRemoteProductStatus struct {
	DestinationStatuses  []GoogleMerchantRemoteDestinationStatus `json:"destination_statuses,omitempty"`
	ItemLevelIssues      []GoogleMerchantRemoteIssue             `json:"item_level_issues,omitempty"`
	CreationDate         string                                  `json:"creation_date"`
	LastUpdateDate       string                                  `json:"last_update_date"`
	GoogleExpirationDate string                                  `json:"google_expiration_date"`
}

type GoogleMerchantRemoteProduct struct {
	Name              string                             `json:"name"`
	Base64EncodedName string                             `json:"base64_encoded_name"`
	LegacyLocal       bool                               `json:"legacy_local"`
	OfferID           string                             `json:"offer_id"`
	ContentLanguage   string                             `json:"content_language"`
	FeedLabel         string                             `json:"feed_label"`
	DataSource        string                             `json:"data_source"`
	ProductAttributes GoogleMerchantRemoteAttributes     `json:"product_attributes"`
	ProductStatus     *GoogleMerchantRemoteProductStatus `json:"product_status,omitempty"`
	Archived          bool                               `json:"archived"`
	VersionNumber     string                             `json:"version_number"`
}

type GoogleMerchantRemoteProductPage struct {
	Products      []GoogleMerchantRemoteProduct `json:"products"`
	NextPageToken string                        `json:"next_page_token,omitempty"`
}

type googleMerchantAPIPrice struct {
	AmountMicros string `json:"amountMicros"`
	CurrencyCode string `json:"currencyCode"`
}

type googleMerchantAPIAttributes struct {
	Title                 string                  `json:"title"`
	Link                  string                  `json:"link"`
	ImageLink             string                  `json:"imageLink"`
	Availability          string                  `json:"availability"`
	Brand                 string                  `json:"brand"`
	Condition             string                  `json:"condition"`
	GoogleProductCategory string                  `json:"googleProductCategory"`
	Price                 *googleMerchantAPIPrice `json:"price"`
	SalePrice             *googleMerchantAPIPrice `json:"salePrice"`
}

type googleMerchantAPIDestinationStatus struct {
	ReportingContext     string   `json:"reportingContext"`
	ApprovedCountries    []string `json:"approvedCountries"`
	PendingCountries     []string `json:"pendingCountries"`
	DisapprovedCountries []string `json:"disapprovedCountries"`
}

type googleMerchantAPIIssue struct {
	Code                string   `json:"code"`
	Severity            string   `json:"severity"`
	Resolution          string   `json:"resolution"`
	Attribute           string   `json:"attribute"`
	ReportingContext    string   `json:"reportingContext"`
	Description         string   `json:"description"`
	Detail              string   `json:"detail"`
	Documentation       string   `json:"documentation"`
	ApplicableCountries []string `json:"applicableCountries"`
}

type googleMerchantAPIProductStatus struct {
	DestinationStatuses  []googleMerchantAPIDestinationStatus `json:"destinationStatuses"`
	ItemLevelIssues      []googleMerchantAPIIssue             `json:"itemLevelIssues"`
	CreationDate         string                               `json:"creationDate"`
	LastUpdateDate       string                               `json:"lastUpdateDate"`
	GoogleExpirationDate string                               `json:"googleExpirationDate"`
}

type googleMerchantAPIProduct struct {
	Name              string                          `json:"name"`
	Base64EncodedName string                          `json:"base64EncodedName"`
	LegacyLocal       bool                            `json:"legacyLocal"`
	OfferID           string                          `json:"offerId"`
	ContentLanguage   string                          `json:"contentLanguage"`
	FeedLabel         string                          `json:"feedLabel"`
	DataSource        string                          `json:"dataSource"`
	ProductAttributes googleMerchantAPIAttributes     `json:"productAttributes"`
	ProductStatus     *googleMerchantAPIProductStatus `json:"productStatus"`
	Archived          bool                            `json:"archived"`
	VersionNumber     string                          `json:"versionNumber"`
}

type googleMerchantAPIProductPage struct {
	Products      []googleMerchantAPIProduct `json:"products"`
	NextPageToken string                     `json:"nextPageToken"`
}

type googleMerchantTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	Error        string `json:"error"`
	ErrorMessage string `json:"error_description"`
}

type googleMerchantUserInfo struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (s *GoogleMerchantService) GetConnection() (*GoogleMerchantConnectionView, error) {
	connection, err := s.offers.FindConnection()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.emptyConnectionView(), nil
	}
	if err != nil {
		return nil, err
	}
	return s.connectionView(connection), nil
}

func (s *GoogleMerchantService) UpdateConnection(input GoogleMerchantConnectionInput) (*GoogleMerchantConnectionView, error) {
	merchantAccountID, err := normalizeGoogleMerchantID(input.MerchantAccountID, "merchant account")
	if err != nil {
		return nil, err
	}
	dataSourceID, err := normalizeGoogleMerchantID(input.DataSourceID, "data source")
	if err != nil {
		return nil, err
	}
	storefrontBaseURL, err := normalizeGoogleMerchantStorefrontBaseURL(input.StorefrontBaseURL)
	if err != nil {
		return nil, err
	}

	connection, err := s.offers.FindConnection()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		connection = &merchant.GoogleMerchantConnection{
			Provider: merchant.GoogleMerchantProvider,
			Status:   "disconnected",
		}
	} else if err != nil {
		return nil, err
	}
	if googleMerchantConnectionSyncBoundaryChanged(connection, merchantAccountID, dataSourceID, storefrontBaseURL) {
		count, err := s.offers.CountOffersWithRemoteSync()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("%w: remove synced Google offers before changing Merchant Account, Data Source, or Storefront Base URL", ErrGoogleMerchantConnectionInvalid)
		}
	}

	connection.MerchantAccountID = merchantAccountID
	connection.DataSourceID = dataSourceID
	connection.StorefrontBaseURL = storefrontBaseURL
	if connection.Status == "" {
		connection.Status = "disconnected"
	}
	if err := s.offers.SaveConnection(connection); err != nil {
		return nil, err
	}
	return s.connectionView(connection), nil
}

func googleMerchantConnectionSyncBoundaryChanged(connection *merchant.GoogleMerchantConnection, merchantAccountID, dataSourceID, storefrontBaseURL string) bool {
	if connection == nil || connection.ID == 0 {
		return false
	}
	return strings.TrimSpace(connection.MerchantAccountID) != merchantAccountID ||
		strings.TrimSpace(connection.DataSourceID) != dataSourceID ||
		strings.TrimRight(strings.TrimSpace(connection.StorefrontBaseURL), "/") != storefrontBaseURL
}

func (s *GoogleMerchantService) StartOAuth(userID uint) (string, error) {
	if !s.oauthConfigured() || strings.TrimSpace(s.googleConfig.TokenEncryptionKey) == "" {
		return "", ErrGoogleMerchantOAuthNotConfigured
	}

	state, err := newGoogleMerchantOAuthState()
	if err != nil {
		return "", err
	}
	stateHash := hashGoogleMerchantOAuthState(state)
	ttl := time.Duration(s.googleConfig.StateTTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	expiresAt := time.Now().UTC().Add(ttl)

	connection, err := s.offers.FindConnection()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		connection = &merchant.GoogleMerchantConnection{
			Provider: merchant.GoogleMerchantProvider,
			Status:   "disconnected",
		}
	} else if err != nil {
		return "", err
	}

	connection.OAuthStateHash = stateHash
	connection.OAuthStateExpiresAt = &expiresAt
	connection.OAuthInitiatedByUserID = &userID
	connection.LastError = ""
	if err := s.offers.SaveConnection(connection); err != nil {
		return "", err
	}

	query := url.Values{}
	query.Set("client_id", strings.TrimSpace(s.googleConfig.ClientID))
	query.Set("redirect_uri", strings.TrimSpace(s.googleConfig.RedirectURL))
	query.Set("response_type", "code")
	query.Set("scope", "openid email profile "+googleMerchantContentScope)
	query.Set("access_type", "offline")
	query.Set("prompt", "consent")
	query.Set("include_granted_scopes", "true")
	query.Set("state", state)
	return googleMerchantOAuthAuthorizationEndpoint + "?" + query.Encode(), nil
}

func (s *GoogleMerchantService) CompleteOAuth(ctx context.Context, state, code, providerError string) error {
	state = strings.TrimSpace(state)
	code = strings.TrimSpace(code)
	if state == "" {
		return ErrGoogleMerchantOAuthStateInvalid
	}

	connection, err := s.offers.ConsumeOAuthState(hashGoogleMerchantOAuthState(state), time.Now().UTC())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGoogleMerchantOAuthStateInvalid
		}
		return err
	}
	if providerError != "" {
		err := fmt.Errorf("%w: %s", ErrGoogleMerchantOAuthExchange, providerError)
		connection.Status = "error"
		connection.LastError = err.Error()
		_ = s.offers.SaveConnection(connection)
		return err
	}
	if code == "" {
		err := fmt.Errorf("%w: authorization code missing", ErrGoogleMerchantOAuthExchange)
		connection.Status = "error"
		connection.LastError = err.Error()
		_ = s.offers.SaveConnection(connection)
		return err
	}

	token, err := exchangeGoogleMerchantCode(ctx, s.googleConfig, code)
	if err != nil {
		connection.Status = "error"
		connection.LastError = err.Error()
		_ = s.offers.SaveConnection(connection)
		return err
	}
	if token.AccessToken == "" {
		err := fmt.Errorf("%w: access token missing", ErrGoogleMerchantOAuthExchange)
		connection.Status = "error"
		connection.LastError = err.Error()
		_ = s.offers.SaveConnection(connection)
		return err
	}

	userInfo, err := fetchGoogleMerchantUserInfo(ctx, token.AccessToken)
	if err != nil {
		connection.Status = "error"
		connection.LastError = err.Error()
		_ = s.offers.SaveConnection(connection)
		return err
	}
	if userInfo.Subject == "" || userInfo.Email == "" || !userInfo.EmailVerified {
		err := fmt.Errorf("%w: verified Google account identity is required", ErrGoogleMerchantOAuthExchange)
		connection.Status = "error"
		connection.LastError = err.Error()
		_ = s.offers.SaveConnection(connection)
		return err
	}

	refreshToken := strings.TrimSpace(token.RefreshToken)
	if refreshToken == "" {
		if connection.RefreshTokenEncrypted == "" {
			err := fmt.Errorf("%w: refresh token missing; reconnect with consent", ErrGoogleMerchantOAuthExchange)
			connection.Status = "error"
			connection.LastError = err.Error()
			_ = s.offers.SaveConnection(connection)
			return err
		}
	} else {
		encrypted, encryptErr := secretbox.EncryptString(refreshToken, s.googleConfig.TokenEncryptionKey)
		if encryptErr != nil {
			err := fmt.Errorf("%w: token encryption unavailable", ErrGoogleMerchantOAuthExchange)
			connection.Status = "error"
			connection.LastError = err.Error()
			_ = s.offers.SaveConnection(connection)
			return err
		}
		connection.RefreshTokenEncrypted = encrypted
	}

	now := time.Now().UTC()
	connection.GoogleSubject = strings.TrimSpace(userInfo.Subject)
	connection.GoogleAccountEmail = strings.ToLower(strings.TrimSpace(userInfo.Email))
	connection.GrantedScopes = strings.TrimSpace(token.Scope)
	connection.Status = "connected"
	connection.LastConnectedAt = &now
	connection.LastError = ""
	if token.ExpiresIn > 0 {
		expiresAt := now.Add(time.Duration(token.ExpiresIn) * time.Second)
		connection.TokenExpiresAt = &expiresAt
	}
	if err := s.offers.SaveConnection(connection); err != nil {
		return err
	}
	return nil
}

func (s *GoogleMerchantService) Disconnect() error {
	connection, err := s.offers.FindConnection()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	connection.GoogleSubject = ""
	connection.GoogleAccountEmail = ""
	connection.RefreshTokenEncrypted = ""
	connection.GrantedScopes = ""
	connection.TokenExpiresAt = nil
	connection.Status = "disconnected"
	connection.LastConnectedAt = nil
	connection.LastError = ""
	connection.OAuthStateHash = ""
	connection.OAuthStateExpiresAt = nil
	connection.OAuthInitiatedByUserID = nil
	return s.offers.SaveConnection(connection)
}

func (s *GoogleMerchantService) AccessToken(ctx context.Context) (string, error) {
	connection, err := s.offers.FindConnection()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrGoogleMerchantConnectionNotFound
	}
	if err != nil {
		return "", err
	}
	if connection.RefreshTokenEncrypted == "" {
		return "", ErrGoogleMerchantConnectionNotFound
	}

	refreshToken, err := secretbox.DecryptString(connection.RefreshTokenEncrypted, s.googleConfig.TokenEncryptionKey)
	if err != nil {
		return "", fmt.Errorf("%w: stored refresh token cannot be decrypted", ErrGoogleMerchantOAuthExchange)
	}
	token, err := refreshGoogleMerchantAccessToken(ctx, s.googleConfig, refreshToken)
	if err != nil {
		return "", err
	}
	if token.ExpiresIn > 0 {
		expiresAt := time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second)
		connection.TokenExpiresAt = &expiresAt
		_ = s.offers.SaveConnection(connection)
	}
	return token.AccessToken, nil
}

func (s *GoogleMerchantService) ListRemoteProducts(ctx context.Context, pageSize int, pageToken string) (*GoogleMerchantRemoteProductPage, error) {
	connection, err := s.offers.FindConnection()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrGoogleMerchantConnectionNotFound
	}
	if err != nil {
		return nil, err
	}
	accountID, err := normalizeGoogleMerchantID(connection.MerchantAccountID, "merchant account")
	if err != nil || accountID == "" {
		return nil, fmt.Errorf("%w: merchant account id is required", ErrGoogleMerchantConnectionInvalid)
	}
	if connection.Status != "connected" || connection.RefreshTokenEncrypted == "" {
		return nil, ErrGoogleMerchantConnectionNotFound
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 1000 {
		pageSize = 1000
	}
	pageToken = strings.TrimSpace(pageToken)
	if len(pageToken) > 2048 {
		return nil, fmt.Errorf("%w: page token is too long", ErrGoogleMerchantConnectionInvalid)
	}

	accessToken, err := s.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	return fetchGoogleMerchantProducts(ctx, accessToken, accountID, pageSize, pageToken)
}

func (s *GoogleMerchantService) connectionView(connection *merchant.GoogleMerchantConnection) *GoogleMerchantConnectionView {
	return &GoogleMerchantConnectionView{
		Configured:                connection.MerchantAccountID != "" && connection.DataSourceID != "" && connection.StorefrontBaseURL != "",
		OAuthConfigured:           s.oauthConfigured(),
		TokenEncryptionConfigured: strings.TrimSpace(s.googleConfig.TokenEncryptionKey) != "",
		Connected:                 connection.RefreshTokenEncrypted != "" && connection.Status == "connected",
		Status:                    connection.Status,
		GoogleAccountEmail:        connection.GoogleAccountEmail,
		MerchantAccountID:         connection.MerchantAccountID,
		DataSourceID:              connection.DataSourceID,
		StorefrontBaseURL:         connection.StorefrontBaseURL,
		LastConnectedAt:           connection.LastConnectedAt,
		LastError:                 connection.LastError,
	}
}

func (s *GoogleMerchantService) emptyConnectionView() *GoogleMerchantConnectionView {
	return &GoogleMerchantConnectionView{
		OAuthConfigured:           s.oauthConfigured(),
		TokenEncryptionConfigured: strings.TrimSpace(s.googleConfig.TokenEncryptionKey) != "",
		Status:                    "disconnected",
	}
}

func (s *GoogleMerchantService) effectiveStorefrontBaseURL(connection *merchant.GoogleMerchantConnection) string {
	if connection != nil {
		if value := strings.TrimRight(strings.TrimSpace(connection.StorefrontBaseURL), "/"); value != "" {
			return value
		}
	}
	return strings.TrimRight(strings.TrimSpace(s.storefrontURL), "/")
}

func (s *GoogleMerchantService) oauthConfigured() bool {
	return strings.TrimSpace(s.googleConfig.ClientID) != "" &&
		strings.TrimSpace(s.googleConfig.ClientSecret) != "" &&
		strings.TrimSpace(s.googleConfig.RedirectURL) != ""
}

func exchangeGoogleMerchantCode(ctx context.Context, cfg config.GoogleMerchantConfig, code string) (*googleMerchantTokenResponse, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", strings.TrimSpace(cfg.ClientID))
	form.Set("client_secret", strings.TrimSpace(cfg.ClientSecret))
	form.Set("redirect_uri", strings.TrimSpace(cfg.RedirectURL))
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleMerchantOAuthTokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: create token request", ErrGoogleMerchantOAuthExchange)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleMerchantOAuthExchange, err)
	}
	defer resp.Body.Close()

	var token googleMerchantTokenResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&token); err != nil {
		return nil, fmt.Errorf("%w: invalid token response", ErrGoogleMerchantOAuthExchange)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || token.Error != "" {
		message := strings.TrimSpace(token.ErrorMessage)
		if message == "" {
			message = strings.TrimSpace(token.Error)
		}
		if message == "" {
			message = "Google rejected the authorization code"
		}
		return nil, fmt.Errorf("%w: %s", ErrGoogleMerchantOAuthExchange, message)
	}
	return &token, nil
}

func refreshGoogleMerchantAccessToken(ctx context.Context, cfg config.GoogleMerchantConfig, refreshToken string) (*googleMerchantTokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", strings.TrimSpace(cfg.ClientID))
	form.Set("client_secret", strings.TrimSpace(cfg.ClientSecret))
	form.Set("refresh_token", refreshToken)
	form.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleMerchantOAuthTokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: create refresh request", ErrGoogleMerchantOAuthExchange)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleMerchantOAuthExchange, err)
	}
	defer resp.Body.Close()

	var token googleMerchantTokenResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&token); err != nil {
		return nil, fmt.Errorf("%w: invalid refresh response", ErrGoogleMerchantOAuthExchange)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || token.Error != "" || token.AccessToken == "" {
		message := strings.TrimSpace(token.ErrorMessage)
		if message == "" {
			message = strings.TrimSpace(token.Error)
		}
		if message == "" {
			message = "Google rejected the refresh token"
		}
		return nil, fmt.Errorf("%w: %s", ErrGoogleMerchantOAuthExchange, message)
	}
	return &token, nil
}

func fetchGoogleMerchantProducts(ctx context.Context, accessToken, accountID string, pageSize int, pageToken string) (*GoogleMerchantRemoteProductPage, error) {
	endpoint := fmt.Sprintf("%s/accounts/%s/products", googleMerchantProductsEndpoint, url.PathEscape(accountID))
	query := url.Values{}
	query.Set("pageSize", strconv.Itoa(pageSize))
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: create products request", ErrGoogleMerchantRemoteAPI)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleMerchantRemoteAPI, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiResponse struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&apiResponse)
		message := strings.TrimSpace(apiResponse.Error.Message)
		if message == "" {
			message = "Google Merchant rejected the products request"
		}
		return nil, fmt.Errorf("%w: %s", ErrGoogleMerchantRemoteAPI, message)
	}

	var apiPage googleMerchantAPIProductPage
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&apiPage); err != nil {
		return nil, fmt.Errorf("%w: invalid products response", ErrGoogleMerchantRemoteAPI)
	}
	page := toGoogleMerchantRemoteProductPage(apiPage)
	if page.Products == nil {
		page.Products = []GoogleMerchantRemoteProduct{}
	}
	return page, nil
}

func toGoogleMerchantRemoteProductPage(apiPage googleMerchantAPIProductPage) *GoogleMerchantRemoteProductPage {
	page := &GoogleMerchantRemoteProductPage{
		Products:      make([]GoogleMerchantRemoteProduct, 0, len(apiPage.Products)),
		NextPageToken: apiPage.NextPageToken,
	}
	for _, apiProduct := range apiPage.Products {
		product := GoogleMerchantRemoteProduct{
			Name:              apiProduct.Name,
			Base64EncodedName: apiProduct.Base64EncodedName,
			LegacyLocal:       apiProduct.LegacyLocal,
			OfferID:           apiProduct.OfferID,
			ContentLanguage:   apiProduct.ContentLanguage,
			FeedLabel:         apiProduct.FeedLabel,
			DataSource:        apiProduct.DataSource,
			ProductAttributes: GoogleMerchantRemoteAttributes{
				Title:                 apiProduct.ProductAttributes.Title,
				Link:                  apiProduct.ProductAttributes.Link,
				ImageLink:             apiProduct.ProductAttributes.ImageLink,
				Availability:          apiProduct.ProductAttributes.Availability,
				Brand:                 apiProduct.ProductAttributes.Brand,
				Condition:             apiProduct.ProductAttributes.Condition,
				GoogleProductCategory: apiProduct.ProductAttributes.GoogleProductCategory,
				Price:                 toGoogleMerchantRemotePrice(apiProduct.ProductAttributes.Price),
				SalePrice:             toGoogleMerchantRemotePrice(apiProduct.ProductAttributes.SalePrice),
			},
			Archived:      apiProduct.Archived,
			VersionNumber: apiProduct.VersionNumber,
		}
		if apiProduct.ProductStatus != nil {
			status := &GoogleMerchantRemoteProductStatus{
				DestinationStatuses:  make([]GoogleMerchantRemoteDestinationStatus, 0, len(apiProduct.ProductStatus.DestinationStatuses)),
				ItemLevelIssues:      make([]GoogleMerchantRemoteIssue, 0, len(apiProduct.ProductStatus.ItemLevelIssues)),
				CreationDate:         apiProduct.ProductStatus.CreationDate,
				LastUpdateDate:       apiProduct.ProductStatus.LastUpdateDate,
				GoogleExpirationDate: apiProduct.ProductStatus.GoogleExpirationDate,
			}
			for _, destination := range apiProduct.ProductStatus.DestinationStatuses {
				status.DestinationStatuses = append(status.DestinationStatuses, GoogleMerchantRemoteDestinationStatus{
					ReportingContext:     destination.ReportingContext,
					ApprovedCountries:    destination.ApprovedCountries,
					PendingCountries:     destination.PendingCountries,
					DisapprovedCountries: destination.DisapprovedCountries,
				})
			}
			for _, issue := range apiProduct.ProductStatus.ItemLevelIssues {
				status.ItemLevelIssues = append(status.ItemLevelIssues, GoogleMerchantRemoteIssue{
					Code:                issue.Code,
					Severity:            issue.Severity,
					Resolution:          issue.Resolution,
					Attribute:           issue.Attribute,
					ReportingContext:    issue.ReportingContext,
					Description:         issue.Description,
					Detail:              issue.Detail,
					Documentation:       issue.Documentation,
					ApplicableCountries: issue.ApplicableCountries,
				})
			}
			product.ProductStatus = status
		}
		page.Products = append(page.Products, product)
	}
	return page
}

func toGoogleMerchantRemotePrice(price *googleMerchantAPIPrice) *GoogleMerchantRemotePrice {
	if price == nil {
		return nil
	}
	return &GoogleMerchantRemotePrice{
		AmountMicros: price.AmountMicros,
		CurrencyCode: price.CurrencyCode,
	}
}

func fetchGoogleMerchantUserInfo(ctx context.Context, accessToken string) (*googleMerchantUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleMerchantUserInfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: create userinfo request", ErrGoogleMerchantOAuthExchange)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleMerchantOAuthExchange, err)
	}
	defer resp.Body.Close()

	var userInfo googleMerchantUserInfo
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("%w: invalid Google account response", ErrGoogleMerchantOAuthExchange)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: Google account identity request rejected", ErrGoogleMerchantOAuthExchange)
	}
	return &userInfo, nil
}

func newGoogleMerchantOAuthState() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashGoogleMerchantOAuthState(state string) string {
	hash := sha256.Sum256([]byte(state))
	return fmt.Sprintf("%x", hash[:])
}

var googleMerchantIDPattern = regexp.MustCompile(`^[0-9]+$`)

func normalizeGoogleMerchantID(value, label string) (string, error) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "accounts/")
	value = strings.TrimPrefix(value, "dataSources/")
	if value == "" {
		return "", nil
	}
	if !googleMerchantIDPattern.MatchString(value) {
		return "", fmt.Errorf("%w: %s id must contain digits only", ErrGoogleMerchantConnectionInvalid, label)
	}
	if _, err := strconv.ParseUint(value, 10, 64); err != nil {
		return "", fmt.Errorf("%w: %s id is too large", ErrGoogleMerchantConnectionInvalid, label)
	}
	return value, nil
}

func normalizeGoogleMerchantStorefrontBaseURL(value string) (string, error) {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return "", nil
	}
	normalized, err := googleMerchantPublicBaseURL(value, "storefront base URL")
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGoogleMerchantConnectionInvalid, err)
	}
	if len(normalized) > 255 {
		return "", fmt.Errorf("%w: storefront base URL must not exceed 255 characters", ErrGoogleMerchantConnectionInvalid)
	}
	return strings.TrimRight(normalized, "/"), nil
}
