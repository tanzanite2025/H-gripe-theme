package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/config"

	"github.com/stretchr/testify/require"
)

type googleIndexingProductReaderFunc func(uint) (*product.Product, error)

func (f googleIndexingProductReaderFunc) GetAdminProduct(id uint) (*product.Product, error) {
	return f(id)
}

func TestGoogleIndexingServiceRejectsProductPagesBeforeAnySideEffect(t *testing.T) {
	var productReads atomic.Int32
	var upstreamRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		upstreamRequests.Add(1)
		http.Error(writer, "unexpected upstream request", http.StatusInternalServerError)
	}))
	defer server.Close()

	service := &GoogleIndexingService{
		products: googleIndexingProductReaderFunc(func(uint) (*product.Product, error) {
			productReads.Add(1)
			return &product.Product{ID: 7, Slug: "carbon-wheel", Locale: "en", Status: "active"}, nil
		}),
		config: config.GoogleIndexingConfig{
			Enabled: true, ServiceAccountJSON: "configured",
		},
		storefrontURL: "https://store.example.test",
		publishURL:    server.URL + "/publish",
	}
	service.configureDefaultHTTPClients()

	result, err := service.PushProduct(context.Background(), 7)

	require.Nil(t, result)
	require.ErrorIs(t, err, ErrGoogleIndexingProductUnsupported)
	require.Zero(t, productReads.Load(), "unsupported product pages must not be loaded")
	require.Zero(t, upstreamRequests.Load(), "unsupported product pages must not notify Google")
}

func TestGoogleIndexingServiceStatusExplainsProductRestriction(t *testing.T) {
	service := &GoogleIndexingService{
		config: config.GoogleIndexingConfig{Enabled: true, ServiceAccountJSON: "configured"},
	}

	status := service.Status()

	require.False(t, status.Enabled)
	require.False(t, status.Configured)
	require.False(t, status.Ready)
	require.Contains(t, status.Message, "not supported for product pages")
}

func TestLoadGoogleIndexingCredentialsParsesServiceAccountJSON(t *testing.T) {
	privateKey := mustGoogleIndexingTestPrivateKey(t)
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	raw := "{\"type\":\"service_account\",\"private_key_id\":\"key-1\",\"private_key\":" +
		strconvQuote(string(privatePEM)) +
		",\"client_email\":\"indexer@example.iam.gserviceaccount.com\",\"token_uri\":\"https://oauth2.example.test/token\"}"

	credentials, err := loadGoogleIndexingCredentials(config.GoogleIndexingConfig{
		Enabled:               true,
		ServiceAccountJSON:    raw,
		RequestTimeoutSeconds: 15,
	})

	require.NoError(t, err)
	require.Equal(t, "indexer@example.iam.gserviceaccount.com", credentials.ClientEmail)
	require.Equal(t, "key-1", credentials.PrivateKeyID)
	require.Equal(t, "https://oauth2.example.test/token", credentials.TokenURI)
	require.Equal(t, privateKey.PublicKey.N, credentials.PrivateKey.PublicKey.N)
}

func mustGoogleIndexingTestPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func strconvQuote(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}
