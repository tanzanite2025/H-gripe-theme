package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/merchant"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/resilience"

	"gorm.io/gorm"
)

type googleMerchantProductInput struct {
	OfferID           string                        `json:"offerId"`
	ContentLanguage   string                        `json:"contentLanguage"`
	FeedLabel         string                        `json:"feedLabel"`
	ProductAttributes googleMerchantWriteAttributes `json:"productAttributes"`
}

type googleMerchantWriteAttributes struct {
	Title                 string                  `json:"title"`
	Description           string                  `json:"description"`
	Link                  string                  `json:"link"`
	ImageLink             string                  `json:"imageLink"`
	Availability          string                  `json:"availability"`
	Brand                 string                  `json:"brand"`
	Condition             string                  `json:"condition"`
	GoogleProductCategory string                  `json:"googleProductCategory"`
	GTINs                 []string                `json:"gtins,omitempty"`
	MPN                   string                  `json:"mpn,omitempty"`
	IdentifierExists      bool                    `json:"identifierExists"`
	Price                 googleMerchantAPIPrice  `json:"price"`
	SalePrice             *googleMerchantAPIPrice `json:"salePrice,omitempty"`
}

const (
	googleMerchantAPIBreakerKey   = "google-merchant-api"
	googleMerchantOAuthBreakerKey = "google-oauth-api"
)

// SyncOffer inserts exactly one channel offer. The Merchant API operation is
// idempotent for the same offerId/contentLanguage/dataSource tuple, so the
// admin can safely retry the same record after a transient failure.
func (s *GoogleMerchantService) SyncOffer(ctx context.Context, id uint) (*merchant.GoogleMerchantOffer, error) {
	offer, err := s.offers.FindOfferByID(id)
	if err != nil {
		return nil, normalizeGoogleMerchantOfferError(err)
	}
	connection, err := s.offers.FindConnection()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.failOfferSync(offer, ErrGoogleMerchantConnectionNotFound, "sync_failed")
	}
	if err != nil {
		return s.failOfferSync(offer, err, "sync_failed")
	}
	accountID, err := normalizeGoogleMerchantID(connection.MerchantAccountID, "merchant account")
	if err != nil || accountID == "" {
		if err == nil {
			err = fmt.Errorf("%w: merchant account id is required", ErrGoogleMerchantConnectionInvalid)
		}
		return s.failOfferSync(offer, err, "sync_failed")
	}
	dataSourceID, err := normalizeGoogleMerchantID(connection.DataSourceID, "data source")
	if err != nil || dataSourceID == "" {
		if err == nil {
			err = fmt.Errorf("%w: data source id is required", ErrGoogleMerchantConnectionInvalid)
		}
		return s.failOfferSync(offer, err, "sync_failed")
	}
	if connection.Status != "connected" || connection.RefreshTokenEncrypted == "" {
		return s.failOfferSync(offer, ErrGoogleMerchantConnectionNotFound, "sync_failed")
	}
	storefrontBaseURL := s.effectiveStorefrontBaseURL(connection)
	if offer.PublicationStatus != "ready" {
		return s.failOfferSync(offer, fmt.Errorf("%w: offer must be validated before syncing", ErrGoogleMerchantOfferInvalid), "validation_failed")
	}
	if err := s.validateReadyOfferWithStorefrontURL(offer, storefrontBaseURL); err != nil {
		return s.failOfferSync(offer, err, "validation_failed")
	}

	input, err := s.buildGoogleMerchantProductInput(offer, storefrontBaseURL)
	if err != nil {
		return s.failOfferSync(offer, err, "validation_failed")
	}

	offer.SyncStatus = "syncing"
	offer.LastError = ""
	if err := s.offers.UpdateOfferSyncState(offer.ID, "syncing", nil, ""); err != nil {
		return nil, err
	}

	accessToken, err := s.AccessToken(ctx)
	if err != nil {
		return s.failOfferSync(offer, err, "sync_failed")
	}
	if err := insertGoogleMerchantProductInputWithClient(ctx, s.outboundHTTPClient(), accessToken, accountID, dataSourceID, input); err != nil {
		return s.failOfferSync(offer, err, "sync_failed")
	}

	now := time.Now().UTC()
	offer.SyncStatus = "synced"
	offer.LastSyncAt = &now
	offer.LastError = ""
	if err := s.offers.UpdateOfferSyncState(offer.ID, "synced", &now, ""); err != nil {
		return nil, err
	}
	return s.offers.FindOfferByID(id)
}

func (s *GoogleMerchantService) RemoveRemoteOffer(ctx context.Context, id uint) (*merchant.GoogleMerchantOffer, error) {
	offer, err := s.offers.FindOfferByID(id)
	if err != nil {
		return nil, normalizeGoogleMerchantOfferError(err)
	}
	if offer.LastSyncAt == nil || offer.SyncStatus == "removed" {
		return nil, fmt.Errorf("%w: offer has not been submitted to Google", ErrGoogleMerchantOfferInvalid)
	}
	connection, err := s.offers.FindConnection()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.failOfferSync(offer, ErrGoogleMerchantConnectionNotFound, "sync_failed")
	}
	if err != nil {
		return s.failOfferSync(offer, err, "sync_failed")
	}
	accountID, err := normalizeGoogleMerchantID(connection.MerchantAccountID, "merchant account")
	if err != nil || accountID == "" {
		if err == nil {
			err = fmt.Errorf("%w: merchant account id is required", ErrGoogleMerchantConnectionInvalid)
		}
		return s.failOfferSync(offer, err, "sync_failed")
	}
	dataSourceID, err := normalizeGoogleMerchantID(connection.DataSourceID, "data source")
	if err != nil || dataSourceID == "" {
		if err == nil {
			err = fmt.Errorf("%w: data source id is required", ErrGoogleMerchantConnectionInvalid)
		}
		return s.failOfferSync(offer, err, "sync_failed")
	}
	if connection.Status != "connected" || connection.RefreshTokenEncrypted == "" {
		return s.failOfferSync(offer, ErrGoogleMerchantConnectionNotFound, "sync_failed")
	}
	accessToken, err := s.AccessToken(ctx)
	if err != nil {
		return s.failOfferSync(offer, err, "sync_failed")
	}
	if err := deleteGoogleMerchantProductInputWithClient(ctx, s.outboundHTTPClient(), accessToken, accountID, dataSourceID, offer.ContentLanguage, offer.FeedLabel, offer.OfferID); err != nil {
		return s.failOfferSync(offer, err, "sync_failed")
	}

	now := time.Now().UTC()
	offer.SyncStatus = "removed"
	offer.LastSyncAt = &now
	offer.LastError = ""
	if err := s.offers.UpdateOfferSyncState(offer.ID, "removed", &now, ""); err != nil {
		return nil, err
	}
	return s.offers.FindOfferByID(id)
}

func (s *GoogleMerchantService) buildGoogleMerchantProductInput(offer *merchant.GoogleMerchantOffer, storefrontBaseURL string) (*googleMerchantProductInput, error) {
	if offer == nil || offer.Product == nil || offer.Variant == nil {
		return nil, fmt.Errorf("%w: offer source SKU is unavailable", ErrGoogleMerchantOfferInvalid)
	}
	link, err := googleMerchantProductURL(storefrontBaseURL, offer.Product, offer.Variant)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleMerchantOfferInvalid, err)
	}
	imageLink, err := firstGoogleMerchantImage(storefrontBaseURL, offer.Product.Media, s.mediaResolver)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleMerchantOfferInvalid, err)
	}

	title := strings.TrimSpace(offer.Title)
	if title == "" {
		title = strings.TrimSpace(offer.Product.Name)
	}
	description := strings.TrimSpace(offer.Description)
	if description == "" {
		description = strings.TrimSpace(offer.Product.Description)
	}
	if title == "" || description == "" {
		return nil, fmt.Errorf("%w: title and description are required", ErrGoogleMerchantOfferInvalid)
	}

	price := offer.Variant.Price
	if offer.PriceOverride != nil {
		price = *offer.PriceOverride
	}
	priceMicros, err := googleMerchantPriceMicros(price)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleMerchantOfferInvalid, err)
	}

	availability := "OUT_OF_STOCK"
	if offer.Variant.IsActive && offer.Variant.Stock > 0 {
		availability = "IN_STOCK"
	}

	input := &googleMerchantProductInput{
		OfferID:         offer.OfferID,
		ContentLanguage: offer.ContentLanguage,
		FeedLabel:       offer.FeedLabel,
		ProductAttributes: googleMerchantWriteAttributes{
			Title:                 title,
			Description:           description,
			Link:                  link,
			ImageLink:             imageLink,
			Availability:          availability,
			Brand:                 offer.Brand,
			Condition:             strings.ToUpper(offer.Condition),
			GoogleProductCategory: offer.GoogleProductCategory,
			IdentifierExists:      *offer.IdentifierExists,
			Price: googleMerchantAPIPrice{
				AmountMicros: strconv.FormatInt(priceMicros, 10),
				CurrencyCode: offer.CurrencyCode,
			},
		},
	}
	if offer.GTIN != "" {
		input.ProductAttributes.GTINs = []string{offer.GTIN}
	}
	if offer.MPN != "" {
		input.ProductAttributes.MPN = offer.MPN
	}
	if offer.SalePriceOverride != nil {
		saleMicros, err := googleMerchantPriceMicros(*offer.SalePriceOverride)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrGoogleMerchantOfferInvalid, err)
		}
		input.ProductAttributes.SalePrice = &googleMerchantAPIPrice{
			AmountMicros: strconv.FormatInt(saleMicros, 10),
			CurrencyCode: offer.CurrencyCode,
		}
	} else if offer.Variant.SalePrice != nil {
		saleMicros, err := googleMerchantPriceMicros(*offer.Variant.SalePrice)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrGoogleMerchantOfferInvalid, err)
		}
		input.ProductAttributes.SalePrice = &googleMerchantAPIPrice{
			AmountMicros: strconv.FormatInt(saleMicros, 10),
			CurrencyCode: offer.CurrencyCode,
		}
	}
	return input, nil
}

func (s *GoogleMerchantService) failOfferSync(offer *merchant.GoogleMerchantOffer, syncErr error, status string) (*merchant.GoogleMerchantOffer, error) {
	if offer == nil {
		return nil, syncErr
	}
	offer.SyncStatus = status
	offer.LastError = syncErr.Error()
	if err := s.offers.UpdateOfferSyncState(offer.ID, status, nil, syncErr.Error()); err != nil {
		return nil, fmt.Errorf("%w; failed to record offer error: %v", syncErr, err)
	}
	return nil, syncErr
}

func insertGoogleMerchantProductInput(ctx context.Context, accessToken, accountID, dataSourceID string, input *googleMerchantProductInput) error {
	return insertGoogleMerchantProductInputWithClient(
		ctx,
		googleMerchantHTTPClient(),
		accessToken,
		accountID,
		dataSourceID,
		input,
	)
}

func insertGoogleMerchantProductInputWithClient(
	ctx context.Context,
	client *resilience.HTTPClient,
	accessToken, accountID, dataSourceID string,
	input *googleMerchantProductInput,
) error {
	endpoint := fmt.Sprintf("%s/accounts/%s/productInputs:insert", googleMerchantProductsEndpoint, url.PathEscape(accountID))
	query := url.Values{}
	query.Set("dataSource", fmt.Sprintf("accounts/%s/dataSources/%s", accountID, dataSourceID))
	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("%w: encode product input", ErrGoogleMerchantRemoteAPI)
	}

	resp, err := client.Do(ctx, func() (*http.Request, error) {
		clonedReq, createErr := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?"+query.Encode(), bytes.NewReader(body))
		if createErr != nil {
			return nil, fmt.Errorf("%w: create product input request: %v", ErrGoogleMerchantRemoteAPI, createErr)
		}
		clonedReq.Header.Set("Authorization", "Bearer "+accessToken)
		clonedReq.Header.Set("Content-Type", "application/json")
		return clonedReq, nil
	})
	if err != nil {
		if resp == nil || resp.Body == nil {
			return fmt.Errorf("%w: %v", ErrGoogleMerchantRemoteAPI, err)
		}
		defer resp.Body.Close()
	} else {
		defer resp.Body.Close()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiResponse struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&apiResponse)
		message := strings.TrimSpace(apiResponse.Error.Message)
		if message == "" {
			message = "Google Merchant rejected the product input"
		}
		if err != nil {
			message = fmt.Sprintf("%s: %v", message, err)
		}
		return fmt.Errorf("%w: %s", ErrGoogleMerchantRemoteAPI, message)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return nil
}

func deleteGoogleMerchantProductInput(ctx context.Context, accessToken, accountID, dataSourceID, contentLanguage, feedLabel, offerID string) error {
	return deleteGoogleMerchantProductInputWithClient(
		ctx,
		googleMerchantHTTPClient(),
		accessToken,
		accountID,
		dataSourceID,
		contentLanguage,
		feedLabel,
		offerID,
	)
}

func deleteGoogleMerchantProductInputWithClient(
	ctx context.Context,
	client *resilience.HTTPClient,
	accessToken, accountID, dataSourceID, contentLanguage, feedLabel, offerID string,
) error {
	name, err := googleMerchantProductInputName(accountID, contentLanguage, feedLabel, offerID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrGoogleMerchantOfferInvalid, err)
	}
	endpoint := googleMerchantProductsEndpoint + "/" + name
	query := url.Values{}
	query.Set("dataSource", fmt.Sprintf("accounts/%s/dataSources/%s", accountID, dataSourceID))

	resp, err := client.Do(ctx, func() (*http.Request, error) {
		clonedReq, createErr := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint+"?"+query.Encode(), nil)
		if createErr != nil {
			return nil, fmt.Errorf("%w: create product input delete request: %v", ErrGoogleMerchantRemoteAPI, createErr)
		}
		clonedReq.Header.Set("Authorization", "Bearer "+accessToken)
		return clonedReq, nil
	})
	if err != nil {
		if resp == nil || resp.Body == nil {
			return fmt.Errorf("%w: %v", ErrGoogleMerchantRemoteAPI, err)
		}
		defer resp.Body.Close()
	} else {
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusNotFound {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiResponse struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&apiResponse)
		message := strings.TrimSpace(apiResponse.Error.Message)
		if message == "" {
			message = "Google Merchant rejected the product input deletion"
		}
		if err != nil {
			message = fmt.Sprintf("%s: %v", message, err)
		}
		return fmt.Errorf("%w: %s", ErrGoogleMerchantRemoteAPI, message)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return nil
}

func googleMerchantHTTPClient() *resilience.HTTPClient {
	return newGoogleMerchantHTTPClient(resilience.HTTPRetryPolicy{
		MaxAttempts: 3,
		Backoff: resilience.BackoffPolicy{
			BaseDelay: 300 * time.Millisecond,
			MaxDelay:  3 * time.Second,
			Jitter:    300 * time.Millisecond,
		},
		RetryUnsafeMethods: true,
	}, resilience.SharedCircuitBreaker(googleMerchantAPIBreakerKey, resilience.CircuitBreakerConfig{
		FailureThreshold: 4,
		FailureWindow:    60 * time.Second,
		OpenDuration:     30 * time.Second,
	}))
}

func newGoogleMerchantHTTPClient(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) *resilience.HTTPClient {
	return newGoogleMerchantHTTPClientWithOptions(googleMerchantAPIBreakerKey, 20*time.Second, retry, breaker)
}

func googleMerchantOAuthHTTPClient() *resilience.HTTPClient {
	const breakerKey = googleMerchantOAuthBreakerKey
	return newGoogleMerchantOAuthHTTPClient(resilience.HTTPRetryPolicy{
		MaxAttempts: 3,
		Backoff: resilience.BackoffPolicy{
			BaseDelay: 300 * time.Millisecond,
			MaxDelay:  3 * time.Second,
			Jitter:    300 * time.Millisecond,
		},
		RetryUnsafeMethods: false,
	}, resilience.SharedCircuitBreaker(breakerKey, resilience.CircuitBreakerConfig{
		FailureThreshold: 4,
		FailureWindow:    60 * time.Second,
		OpenDuration:     30 * time.Second,
	}))
}

func newGoogleMerchantOAuthHTTPClient(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) *resilience.HTTPClient {
	retry.RetryUnsafeMethods = false
	return newGoogleMerchantHTTPClientWithOptions(googleMerchantOAuthBreakerKey, 10*time.Second, retry, breaker)
}

func newGoogleMerchantHTTPClientWithOptions(
	breakerKey string,
	timeout time.Duration,
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) *resilience.HTTPClient {
	return resilience.NewHTTPClient(
		&http.Client{Timeout: timeout},
		retry,
		breaker,
		breakerKey,
	)
}

func (s *GoogleMerchantService) outboundHTTPClient() *resilience.HTTPClient {
	if s != nil && s.httpClient != nil {
		return s.httpClient
	}
	return googleMerchantHTTPClient()
}

func (s *GoogleMerchantService) oauthOutboundHTTPClient() *resilience.HTTPClient {
	if s != nil && s.oauthHTTPClient != nil {
		return s.oauthHTTPClient
	}
	return googleMerchantOAuthHTTPClient()
}

func googleMerchantProductInputName(accountID, contentLanguage, feedLabel, offerID string) (string, error) {
	accountID = strings.TrimSpace(accountID)
	contentLanguage = strings.TrimSpace(contentLanguage)
	feedLabel = strings.TrimSpace(feedLabel)
	offerID = strings.TrimSpace(offerID)
	if accountID == "" || contentLanguage == "" || feedLabel == "" || offerID == "" {
		return "", errors.New("account, content language, feed label, and offer id are required")
	}
	identity := contentLanguage + "~" + feedLabel + "~" + offerID
	encodedIdentity := base64.RawURLEncoding.EncodeToString([]byte(identity))
	return fmt.Sprintf("accounts/%s/productInputs/%s", url.PathEscape(accountID), encodedIdentity), nil
}

func googleMerchantProductURL(baseURL string, item *product.Product, variant *product.ProductVariant) (string, error) {
	baseURL, err := googleMerchantPublicBaseURL(baseURL, "storefront base URL")
	if err != nil {
		return "", err
	}
	if item == nil {
		return "", errors.New("product is required for storefront URL")
	}
	slug := strings.TrimSpace(item.Slug)
	if strings.TrimSpace(slug) == "" {
		return "", errors.New("product slug is required for storefront URL")
	}
	path := "/shop/" + url.PathEscape(strings.Trim(slug, "/"))
	if locale := strings.ToLower(strings.TrimSpace(item.Locale)); locale != "" && locale != "en" {
		path = "/" + url.PathEscape(locale) + path
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	parsed.RawQuery = ""
	if variant != nil && variant.ID != 0 {
		query := parsed.Query()
		query.Set("variant", strconv.FormatUint(uint64(variant.ID), 10))
		parsed.RawQuery = query.Encode()
	}
	return parsed.String(), nil
}

func googleMerchantPublicHTTPURL(value, label string) (string, error) {
	return googleMerchantPublicURL(value, label, false, false)
}

func googleMerchantPublicBaseURL(value, label string) (string, error) {
	return googleMerchantPublicURL(value, label, false, false)
}

func googleMerchantPublicResourceURL(value, label string) (string, error) {
	return googleMerchantPublicURL(value, label, true, false)
}

func googleMerchantPublicURL(value, label string, allowQuery bool, allowFragment bool) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is not configured", label)
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("%s must be an absolute HTTP(S) URL", label)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%s must use HTTP or HTTPS", label)
	}
	if !allowQuery && parsed.RawQuery != "" {
		return "", fmt.Errorf("%s must not include a query string", label)
	}
	if !allowFragment && parsed.Fragment != "" {
		return "", fmt.Errorf("%s must not include a fragment", label)
	}
	if parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "::1" {
		return "", fmt.Errorf("%s must use a public host", label)
	}
	return parsed.String(), nil
}

func googleMerchantResolvePublicResourceURL(baseURL, value, label string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is not configured", label)
	}
	parsed, err := url.Parse(value)
	if err == nil && parsed.Scheme != "" {
		return googleMerchantPublicResourceURL(value, label)
	}
	base, err := googleMerchantPublicBaseURL(baseURL, "storefront base URL")
	if err != nil {
		return "", err
	}
	baseParsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	relative, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("%s must be a valid URL path", label)
	}
	return googleMerchantPublicResourceURL(baseParsed.ResolveReference(relative).String(), label)
}

func firstGoogleMerchantImage(baseURL string, media []product.ProductMedia, resolver PublicMediaURLResolver) (string, error) {
	var fallbackErr error
	for _, item := range media {
		if !item.IsVisible || item.MediaType != "image" {
			continue
		}
		candidate := canonicalPublicMediaURL(resolver, item.URL)
		if candidate == "" {
			continue
		}
		normalized, err := googleMerchantResolvePublicResourceURL(baseURL, candidate, "source product image")
		if err == nil {
			return normalized, nil
		}
		if fallbackErr == nil {
			fallbackErr = err
		}
	}
	if fallbackErr != nil {
		return "", fallbackErr
	}
	return "", errors.New("source product requires a visible image")
}

func googleMerchantPriceMicros(value float64) (int64, error) {
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, errors.New("price must be greater than zero")
	}
	micros := math.Round(value * 1_000_000)
	if micros > math.MaxInt64 {
		return 0, errors.New("price is too large")
	}
	return int64(micros), nil
}
