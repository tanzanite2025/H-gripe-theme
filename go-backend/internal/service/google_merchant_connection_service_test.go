package service

import (
	"strings"
	"testing"
	"context"
	"io"
	"net/http"
	"sync/atomic"

	"commerce-platform/internal/pkg/config"
)

func TestNormalizeGoogleMerchantID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain account id", input: "123456789", want: "123456789"},
		{name: "merchant resource name", input: "accounts/123456789", want: "123456789"},
		{name: "data source resource name", input: "dataSources/987654321", want: "987654321"},
		{name: "empty", input: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeGoogleMerchantID(test.input, "merchant account")
			if err != nil {
				t.Fatalf("normalizeGoogleMerchantID() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("normalizeGoogleMerchantID() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeGoogleMerchantIDRejectsNonNumericValue(t *testing.T) {
	_, err := normalizeGoogleMerchantID("accounts/not-a-number", "merchant account")
	if err == nil || !strings.Contains(err.Error(), "digits only") {
		t.Fatalf("normalizeGoogleMerchantID() error = %v, want numeric validation error", err)
	}
}

func TestNormalizeGoogleMerchantStorefrontBaseURL(t *testing.T) {
	got, err := normalizeGoogleMerchantStorefrontBaseURL("https://example.com/")
	if err != nil {
		t.Fatalf("normalizeGoogleMerchantStorefrontBaseURL() error = %v", err)
	}
	if got != "https://example.com" {
		t.Fatalf("normalizeGoogleMerchantStorefrontBaseURL() = %q", got)
	}
}

func TestNormalizeGoogleMerchantStorefrontBaseURLRejectsLocalhost(t *testing.T) {
	_, err := normalizeGoogleMerchantStorefrontBaseURL("http://localhost:9100")
	if err == nil || !strings.Contains(err.Error(), "public host") {
		t.Fatalf("normalizeGoogleMerchantStorefrontBaseURL() error = %v, want public host validation", err)
	}
}

func TestGoogleMerchantOAuthRequiresEveryServerSecret(t *testing.T) {
	service := &GoogleMerchantService{
		googleConfig: config.GoogleMerchantConfig{
			ClientID:    "client-id",
			RedirectURL: "https://admin.example.test/api/admin/google-merchant/oauth/callback",
		},
	}
	if service.oauthConfigured() {
		t.Fatal("oauthConfigured() should require client secret")
	}

	service.googleConfig.ClientSecret = "client-secret"
	if !service.oauthConfigured() {
		t.Fatal("oauthConfigured() should accept complete OAuth credentials")
	}
}

func TestGoogleMerchantRemoteProductMapping(t *testing.T) {
	page := toGoogleMerchantRemoteProductPage(googleMerchantAPIProductPage{
		NextPageToken: "next-token",
		Products: []googleMerchantAPIProduct{
			{
				Name:            "accounts/123/products/en~US~offer-1",
				OfferID:         "offer-1",
				ContentLanguage: "en",
				FeedLabel:       "US",
				ProductAttributes: googleMerchantAPIAttributes{
					Title:        "Wheelset",
					Availability: "in_stock",
					Price: &googleMerchantAPIPrice{
						AmountMicros: "1299000000",
						CurrencyCode: "USD",
					},
				},
				ProductStatus: &googleMerchantAPIProductStatus{
					DestinationStatuses: []googleMerchantAPIDestinationStatus{{
						ApprovedCountries: []string{"US"},
					}},
					ItemLevelIssues: []googleMerchantAPIIssue{{
						Code:     "image_link_pending",
						Severity: "DEMOTION",
					}},
				},
			},
		},
	})

	if page.NextPageToken != "next-token" || len(page.Products) != 1 {
		t.Fatalf("unexpected page mapping: %#v", page)
	}
	product := page.Products[0]
	if product.OfferID != "offer-1" || product.ProductAttributes.Title != "Wheelset" {
		t.Fatalf("unexpected product mapping: %#v", product)
	}
	if product.ProductAttributes.Price == nil || product.ProductAttributes.Price.AmountMicros != "1299000000" {
		t.Fatalf("unexpected price mapping: %#v", product.ProductAttributes.Price)
	}
	if product.ProductStatus == nil || len(product.ProductStatus.ItemLevelIssues) != 1 {
		t.Fatalf("unexpected status mapping: %#v", product.ProductStatus)
	}
}

func TestGoogleMerchantOAuthTokenExchangeDoesNotRetryUnsafePOST(t *testing.T) {
	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	var requests atomic.Int32
	http.DefaultTransport = googleMerchantRoundTripper(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"temporary"}}`)),
		}, nil
	})

	_, err := exchangeGoogleMerchantCode(context.Background(), config.GoogleMerchantConfig{
		ClientID:    "client-id",
		ClientSecret: "client-secret",
		RedirectURL: "https://admin.example.test/api/admin/google-merchant/oauth/callback",
	}, "authorization-code")
	if err == nil {
		t.Fatal("exchangeGoogleMerchantCode() error = nil, want failure")
	}
	if requests.Load() != 1 {
		t.Fatalf("requests = %d, want 1 for unsafe POST", requests.Load())
	}
}

func TestGoogleMerchantUserInfoRetriesGetRequests(t *testing.T) {
	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	var requests atomic.Int32
	http.DefaultTransport = googleMerchantRoundTripper(func(*http.Request) (*http.Response, error) {
		count := requests.Add(1)
		if count < 3 {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"temporary"}}`)),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"sub":"123","email":"buyer@example.test","email_verified":true}`)),
		}, nil
	})

	userInfo, err := fetchGoogleMerchantUserInfo(context.Background(), "access-token")
	if err != nil {
		t.Fatalf("fetchGoogleMerchantUserInfo() error = %v", err)
	}
	if userInfo.Email != "buyer@example.test" {
		t.Fatalf("userInfo.Email = %q", userInfo.Email)
	}
	if requests.Load() != 3 {
		t.Fatalf("requests = %d, want 3 for retried GET", requests.Load())
	}
}
