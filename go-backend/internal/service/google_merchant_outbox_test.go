package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"tanzanite/internal/domain/merchant"
	"tanzanite/internal/domain/outbox"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/secretbox"

	"github.com/stretchr/testify/require"
)

func TestGoogleMerchantReconcileSkipsDraftOffersWithoutRemoteCalls(t *testing.T) {
	_, googleMerchantService, productRecord, variantRecord := newTestGoogleMerchantService(t)
	identifierExists := false

	offer, err := googleMerchantService.CreateOffer(GoogleMerchantOfferInput{
		ProductID:         productRecord.ID,
		VariantID:         variantRecord.ID,
		OfferID:           "tz-draft",
		Brand:             "H-GRIPE",
		Condition:         "new",
		IdentifierExists:  &identifierExists,
		TargetCountry:     "US",
		ContentLanguage:   "en",
		CurrencyCode:      "USD",
		FeedLabel:         "US",
		PublicationStatus: "draft",
	})
	require.NoError(t, err)
	require.NotNil(t, offer)

	result, err := googleMerchantService.ReconcileOffers(context.Background())

	require.NoError(t, err)
	require.Equal(t, 1, result.Considered)
	require.Equal(t, 0, result.Synced)
	require.Equal(t, 0, result.Withdrawn)
	require.Equal(t, 1, result.Skipped)
	require.Zero(t, result.Failed)
}

func TestGoogleMerchantOutboxHandlerWithdrawsOnlyRemoteSubmissions(t *testing.T) {
	_, googleMerchantService, productRecord, variantRecord := newTestGoogleMerchantService(t)
	identifierExists := false
	offer, err := googleMerchantService.CreateOffer(GoogleMerchantOfferInput{
		ProductID:         productRecord.ID,
		VariantID:         variantRecord.ID,
		OfferID:           "tz-not-submitted",
		Brand:             "H-GRIPE",
		Condition:         "new",
		IdentifierExists:  &identifierExists,
		TargetCountry:     "US",
		ContentLanguage:   "en",
		CurrencyCode:      "USD",
		FeedLabel:         "US",
		PublicationStatus: "ready",
	})
	require.NoError(t, err)
	require.NotNil(t, offer)

	productRecord.Status = "inactive"
	require.NoError(t, googleMerchantService.products.Update(&productRecord))

	handler := NewGoogleMerchantOutboxHandler(googleMerchantService)
	err = handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeMerchantProductWithdraw,
		Payload: mustJSONForMerchantOutboxTest(outbox.MerchantProductSyncPayload{
			ProductID: productRecord.ID,
		}),
	})

	require.NoError(t, err)
	saved, err := googleMerchantService.offers.FindOfferByID(offer.ID)
	require.NoError(t, err)
	require.Equal(t, "not_synced", saved.SyncStatus)
}

func TestGoogleMerchantOutboxHandlerDoesNotPromoteDraftOfferDuringRevalidation(t *testing.T) {
	_, googleMerchantService, productRecord, variantRecord := newTestGoogleMerchantService(t)
	identifierExists := false
	offer, err := googleMerchantService.CreateOffer(GoogleMerchantOfferInput{
		ProductID:         productRecord.ID,
		VariantID:         variantRecord.ID,
		OfferID:           "tz-draft-revalidate",
		Brand:             "H-GRIPE",
		Condition:         "new",
		IdentifierExists:  &identifierExists,
		TargetCountry:     "US",
		ContentLanguage:   "en",
		CurrencyCode:      "USD",
		FeedLabel:         "US",
		PublicationStatus: "draft",
	})
	require.NoError(t, err)

	handler := NewGoogleMerchantOutboxHandler(googleMerchantService)
	err = handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeMerchantOfferRevalidate,
		Payload: mustJSONForMerchantOutboxTest(outbox.MerchantOfferRevalidatePayload{
			OfferID: offer.ID,
		}),
	})

	require.NoError(t, err)
	saved, err := googleMerchantService.offers.FindOfferByID(offer.ID)
	require.NoError(t, err)
	require.Equal(t, "draft", saved.PublicationStatus)
	require.Equal(t, "not_synced", saved.SyncStatus)
}

func TestGoogleMerchantOutboxHandlerWithdrawsPausedRemoteOffer(t *testing.T) {
	_, googleMerchantService, productRecord, variantRecord := newTestGoogleMerchantService(t)
	offer := seedGoogleMerchantOffer(t, googleMerchantService, productRecord.ID, variantRecord.ID, "synced")
	updated, err := googleMerchantService.UpdateOffer(offer.ID, googleMerchantOfferInputForTest(productRecord.ID, variantRecord.ID, "paused"))
	require.NoError(t, err)
	require.Equal(t, "withdraw_pending", updated.SyncStatus)
	configureGoogleMerchantConnectionForRemoteTest(t, googleMerchantService)

	deleteCalls := 0
	withGoogleMerchantHTTPTransport(t, func(request *http.Request) (*http.Response, error) {
		switch request.URL.Host {
		case "oauth2.googleapis.com":
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString(`{"access_token":"access-token","expires_in":3600}`)),
			}, nil
		case "merchantapi.googleapis.com":
			deleteCalls++
			require.Equal(t, http.MethodDelete, request.Method)
			require.Contains(t, request.URL.Path, "/products/v1/accounts/123/productInputs/")
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		default:
			t.Fatalf("unexpected request host %s", request.URL.Host)
			return nil, nil
		}
	})

	handler := NewGoogleMerchantOutboxHandler(googleMerchantService)
	err = handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeMerchantOfferRevalidate,
		Payload: mustJSONForMerchantOutboxTest(outbox.MerchantOfferRevalidatePayload{
			OfferID: offer.ID,
		}),
	})

	require.NoError(t, err)
	require.Equal(t, 1, deleteCalls)
	saved, err := googleMerchantService.offers.FindOfferByID(offer.ID)
	require.NoError(t, err)
	require.Equal(t, "removed", saved.SyncStatus)
}

func TestMerchantOfferWithdrawalPredicateRequiresPublicSource(t *testing.T) {
	require.True(t, merchantOfferShouldBeWithdrawn(nil))
	require.True(t, merchantOfferShouldBeWithdrawn(&merchant.GoogleMerchantOffer{}))
}

func mustJSONForMerchantOutboxTest(value interface{}) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return payload
}

func configureGoogleMerchantConnectionForRemoteTest(t *testing.T, service *GoogleMerchantService) {
	t.Helper()

	service.googleConfig = config.GoogleMerchantConfig{
		ClientID:           "client-id",
		ClientSecret:       "client-secret",
		RedirectURL:        "https://admin.tanzanite.site/google-merchant/callback",
		TokenEncryptionKey: "google-merchant-test-master-key-32",
	}
	encrypted, err := secretbox.EncryptString("refresh-token", service.googleConfig.TokenEncryptionKey)
	require.NoError(t, err)
	require.NoError(t, service.offers.SaveConnection(&merchant.GoogleMerchantConnection{
		Provider:              merchant.GoogleMerchantProvider,
		MerchantAccountID:     "123",
		DataSourceID:          "456",
		StorefrontBaseURL:     "https://tanzanite.site",
		RefreshTokenEncrypted: encrypted,
		Status:                "connected",
	}))
}

func withGoogleMerchantHTTPTransport(t *testing.T, handler func(*http.Request) (*http.Response, error)) {
	t.Helper()

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})
	http.DefaultTransport = googleMerchantRoundTripper(handler)
}
