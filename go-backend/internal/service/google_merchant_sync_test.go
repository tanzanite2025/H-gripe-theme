package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"tanzanite/internal/domain/merchant"
	"tanzanite/internal/domain/product"
)

type googleMerchantRoundTripper func(*http.Request) (*http.Response, error)

func (fn googleMerchantRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestBuildGoogleMerchantProductInputUsesOneSKUAndStorefrontFields(t *testing.T) {
	identifierExists := true
	salePrice := 1199.99
	offer := &merchant.GoogleMerchantOffer{
		OfferID:               "tz-wheel-700",
		Brand:                 "H-GRIPE",
		Condition:             "new",
		GoogleProductCategory: "Sporting Goods",
		GTIN:                  "0123456789012",
		MPN:                   "TZ-700",
		IdentifierExists:      &identifierExists,
		ContentLanguage:       "en",
		FeedLabel:             "US",
		CurrencyCode:          "USD",
		Product: &product.Product{
			Name:        "Carbon Wheelset",
			Slug:        "carbon-wheelset",
			Description: "A fast wheelset.",
			Media: []product.ProductMedia{{
				MediaType: "image",
				URL:       "https://cdn.tanzanite.site/wheelset.jpg",
				IsVisible: true,
			}},
		},
		Variant: &product.ProductVariant{
			ID:        700,
			SKU:       "TZ-700",
			Price:     1299.99,
			SalePrice: &salePrice,
			Stock:     3,
			IsActive:  true,
		},
	}

	input, err := (&GoogleMerchantService{}).buildGoogleMerchantProductInput(offer, "https://tanzanite.site")
	if err != nil {
		t.Fatalf("buildGoogleMerchantProductInput() error = %v", err)
	}

	if input.OfferID != "tz-wheel-700" || input.ContentLanguage != "en" || input.FeedLabel != "US" {
		t.Fatalf("unexpected immutable input fields: %#v", input)
	}
	attributes := input.ProductAttributes
	if attributes.Title != "Carbon Wheelset" || attributes.Description != "A fast wheelset." {
		t.Fatalf("unexpected title/description mapping: %#v", attributes)
	}
	if attributes.Link != "https://tanzanite.site/shop/carbon-wheelset?variant=700" || attributes.ImageLink != "https://cdn.tanzanite.site/wheelset.jpg" {
		t.Fatalf("unexpected URL mapping: %#v", attributes)
	}
	if attributes.Availability != "IN_STOCK" || attributes.Condition != "NEW" {
		t.Fatalf("unexpected availability/condition mapping: %#v", attributes)
	}
	if attributes.Price.AmountMicros != "1299990000" || attributes.Price.CurrencyCode != "USD" {
		t.Fatalf("unexpected price mapping: %#v", attributes.Price)
	}
	if attributes.SalePrice == nil || attributes.SalePrice.AmountMicros != "1199990000" {
		t.Fatalf("unexpected sale price mapping: %#v", attributes.SalePrice)
	}
	if len(attributes.GTINs) != 1 || attributes.GTINs[0] != "0123456789012" || attributes.MPN != "TZ-700" {
		t.Fatalf("unexpected identifier mapping: %#v", attributes)
	}
}

func TestBuildGoogleMerchantProductInputMapsOutOfStockSKU(t *testing.T) {
	identifierExists := false
	offer := &merchant.GoogleMerchantOffer{
		OfferID:               "tz-empty",
		Brand:                 "H-GRIPE",
		Condition:             "new",
		GoogleProductCategory: "Sporting Goods",
		IdentifierExists:      &identifierExists,
		ContentLanguage:       "en",
		FeedLabel:             "US",
		CurrencyCode:          "USD",
		Product: &product.Product{
			Name:        "Empty Wheelset",
			Slug:        "empty-wheelset",
			Description: "Temporarily unavailable.",
			Media: []product.ProductMedia{{
				MediaType: "image",
				URL:       "https://cdn.tanzanite.site/empty-wheelset.jpg",
				IsVisible: true,
			}},
		},
		Variant: &product.ProductVariant{
			Price:    99.99,
			Stock:    0,
			IsActive: true,
		},
	}

	input, err := (&GoogleMerchantService{}).buildGoogleMerchantProductInput(offer, "https://tanzanite.site")
	if err != nil {
		t.Fatalf("buildGoogleMerchantProductInput() error = %v", err)
	}
	if input.ProductAttributes.Availability != "OUT_OF_STOCK" {
		t.Fatalf("availability = %q, want OUT_OF_STOCK", input.ProductAttributes.Availability)
	}
}

func TestBuildGoogleMerchantProductInputResolvesRelativeImageAndLocaleURL(t *testing.T) {
	identifierExists := false
	offer := &merchant.GoogleMerchantOffer{
		OfferID:               "tz-fr",
		Brand:                 "H-GRIPE",
		Condition:             "new",
		GoogleProductCategory: "Sporting Goods",
		IdentifierExists:      &identifierExists,
		ContentLanguage:       "fr",
		FeedLabel:             "FR",
		CurrencyCode:          "EUR",
		Product: &product.Product{
			Name:        "Roue carbone",
			Slug:        "roue-carbone",
			Locale:      "fr",
			Description: "Roue carbone.",
			Media: []product.ProductMedia{{
				MediaType: "image",
				URL:       "/uploads/roue.jpg?width=1200",
				IsVisible: true,
			}},
		},
		Variant: &product.ProductVariant{
			ID:       701,
			Price:    999.99,
			Stock:    5,
			IsActive: true,
		},
	}

	input, err := (&GoogleMerchantService{}).buildGoogleMerchantProductInput(offer, "https://tanzanite.site")
	if err != nil {
		t.Fatalf("buildGoogleMerchantProductInput() error = %v", err)
	}
	if input.ProductAttributes.Link != "https://tanzanite.site/fr/shop/roue-carbone?variant=701" {
		t.Fatalf("link = %q", input.ProductAttributes.Link)
	}
	if input.ProductAttributes.ImageLink != "https://tanzanite.site/uploads/roue.jpg?width=1200" {
		t.Fatalf("image link = %q", input.ProductAttributes.ImageLink)
	}
}

func TestGoogleMerchantProductURLRejectsLocalOrInvalidBase(t *testing.T) {
	tests := []string{"", "not-a-url", "ftp://tanzanite.site", "https://localhost", "https://tanzanite.site?preview=true"}
	for _, input := range tests {
		if _, err := googleMerchantProductURL(input, &product.Product{Slug: "wheelset"}, nil); err == nil {
			t.Fatalf("googleMerchantProductURL(%q) error = nil, want validation error", input)
		}
	}
}

func TestInsertGoogleMerchantProductInputUsesV1Contract(t *testing.T) {
	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = googleMerchantRoundTripper(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", request.Method)
		}
		if request.URL.Scheme != "https" || request.URL.Host != "merchantapi.googleapis.com" {
			t.Errorf("unexpected endpoint %s", request.URL)
		}
		if request.URL.Path != "/products/v1/accounts/123/productInputs:insert" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if got := request.URL.Query().Get("dataSource"); got != "accounts/123/dataSources/456" {
			t.Errorf("dataSource = %q", got)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Errorf("authorization = %q", got)
		}
		if got := request.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("content type = %q", got)
		}

		var got googleMerchantProductInput
		if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if got.OfferID != "tz-wheel-700" || got.ProductAttributes.Availability != "IN_STOCK" {
			t.Errorf("unexpected request body: %#v", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
		}, nil
	})

	err := insertGoogleMerchantProductInput(context.Background(), "access-token", "123", "456", &googleMerchantProductInput{
		OfferID:         "tz-wheel-700",
		ContentLanguage: "en",
		FeedLabel:       "US",
		ProductAttributes: googleMerchantWriteAttributes{
			Availability: "IN_STOCK",
		},
	})
	if err != nil {
		t.Fatalf("insertGoogleMerchantProductInput() error = %v", err)
	}
}

func TestGoogleMerchantProductInputNameUsesEncodedOfferIdentity(t *testing.T) {
	name, err := googleMerchantProductInputName("123", "en", "US", "sku/123")
	if err != nil {
		t.Fatalf("googleMerchantProductInputName() error = %v", err)
	}
	if name != "accounts/123/productInputs/ZW5-VVN-c2t1LzEyMw" {
		t.Fatalf("googleMerchantProductInputName() = %q", name)
	}
}

func TestDeleteGoogleMerchantProductInputUsesV1Contract(t *testing.T) {
	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = googleMerchantRoundTripper(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", request.Method)
		}
		if request.URL.Scheme != "https" || request.URL.Host != "merchantapi.googleapis.com" {
			t.Errorf("unexpected endpoint %s", request.URL)
		}
		if request.URL.Path != "/products/v1/accounts/123/productInputs/ZW5-VVN-c2t1LzEyMw" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if got := request.URL.Query().Get("dataSource"); got != "accounts/123/dataSources/456" {
			t.Errorf("dataSource = %q", got)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Errorf("authorization = %q", got)
		}

		return &http.Response{
			StatusCode: http.StatusNoContent,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(``)),
		}, nil
	})

	if err := deleteGoogleMerchantProductInput(context.Background(), "access-token", "123", "456", "en", "US", "sku/123"); err != nil {
		t.Fatalf("deleteGoogleMerchantProductInput() error = %v", err)
	}
}

func TestDeleteGoogleMerchantProductInputTreatsNotFoundAsSuccess(t *testing.T) {
	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = googleMerchantRoundTripper(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(`{"error":{"message":"not found"}}`)),
		}, nil
	})

	if err := deleteGoogleMerchantProductInput(context.Background(), "access-token", "123", "456", "en", "US", "missing"); err != nil {
		t.Fatalf("deleteGoogleMerchantProductInput() error = %v", err)
	}
}

func TestGoogleMerchantFeedLabelValidation(t *testing.T) {
	for _, label := range []string{"US", "US_EN", "EU-1"} {
		if !isGoogleMerchantFeedLabel(label) {
			t.Fatalf("isGoogleMerchantFeedLabel(%q) = false, want true", label)
		}
	}
	for _, label := range []string{"", "us", "US label", strings.Repeat("A", 21)} {
		if isGoogleMerchantFeedLabel(label) {
			t.Fatalf("isGoogleMerchantFeedLabel(%q) = true, want false", label)
		}
	}
}
