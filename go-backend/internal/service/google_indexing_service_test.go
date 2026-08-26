package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/config"
	"github.com/alicebob/miniredis/v2"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type googleIndexingProductReaderStub struct {
	item *product.Product
	err  error
}

func (s googleIndexingProductReaderStub) GetAdminProduct(uint) (*product.Product, error) {
	return s.item, s.err
}

func TestGoogleIndexingServicePushProductUsesServiceAccountAndPublishesURL(t *testing.T) {
	privateKey := mustGoogleIndexingTestPrivateKey(t)
	var tokenRequests atomic.Int32
	var publishRequests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/token":
			tokenRequests.Add(1)
			require.Equal(t, http.MethodPost, request.Method)
			require.NoError(t, request.ParseForm())
			require.Equal(t, "urn:ietf:params:oauth:grant-type:jwt-bearer", request.PostForm.Get("grant_type"))
			require.NotEmpty(t, request.PostForm.Get("assertion"))
			writer.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(writer, `{"access_token":"test-access-token","expires_in":3600}`)
		case "/publish":
			publishRequests.Add(1)
			require.Equal(t, http.MethodPost, request.Method)
			require.Equal(t, "Bearer test-access-token", request.Header.Get("Authorization"))
			var payload map[string]string
			require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
			require.Equal(t, map[string]string{
				"url":  "https://store.example.test/products/carbon-wheel",
				"type": googleIndexingNotifyType,
			}, payload)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(writer, `{"urlNotificationMetadata":{"url":"https://store.example.test/products/carbon-wheel"}}`)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	redisServer, redisClient := newGoogleIndexingTestRedis(t)
	defer redisServer.Close()

	service := &GoogleIndexingService{
		products: googleIndexingProductReaderStub{
			item: &product.Product{
				ID:     7,
				Slug:   "carbon-wheel",
				Locale: "en",
				Status: "active",
			},
		},
		config: config.GoogleIndexingConfig{
			Enabled:               true,
			ServiceAccountJSON:    `{"client_email":"indexer@example.iam.gserviceaccount.com"}`,
			RequestTimeoutSeconds: 2,
		},
		storefrontURL: "https://store.example.test",
		credentials: googleIndexingCredentials{
			ClientEmail:  "indexer@example.iam.gserviceaccount.com",
			PrivateKey:   privateKey,
			PrivateKeyID: "test-key",
			TokenURI:     server.URL + "/token",
		},
		publishURL:  server.URL + "/publish",
		redisClient: redisClient,
		now:         func() time.Time { return time.Unix(1_700_000_000, 0).UTC() },
	}
	service.configureDefaultHTTPClients()

	result, err := service.PushProduct(context.Background(), 7)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, uint(7), result.ProductID)
	require.Equal(t, "https://store.example.test/products/carbon-wheel", result.URL)
	require.Equal(t, googleIndexingNotifyType, result.NotificationType)
	require.True(t, result.Accepted)
	require.Equal(t, http.StatusOK, result.HTTPStatus)
	require.Equal(t, int32(1), tokenRequests.Load())
	require.Equal(t, int32(1), publishRequests.Load())
	require.Equal(t, "https://store.example.test/products/carbon-wheel", result.Metadata["url"])

	_, err = service.PushProduct(context.Background(), 7)
	require.ErrorIs(t, err, ErrGoogleIndexingRecentlyNotified)
	require.Equal(t, int32(1), tokenRequests.Load(), "access token should be cached")
	require.Equal(t, int32(1), publishRequests.Load())

	redisServer.FastForward(googleIndexingCooldownTTL)
	_, err = service.PushProduct(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, int32(1), tokenRequests.Load(), "access token should be cached")
	require.Equal(t, int32(2), publishRequests.Load())
}

func TestGoogleIndexingServiceRejectsInactiveProduct(t *testing.T) {
	privateKey := mustGoogleIndexingTestPrivateKey(t)
	var upstreamRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		upstreamRequests.Add(1)
		http.Error(writer, "unexpected upstream request", http.StatusInternalServerError)
	}))
	defer server.Close()

	service := &GoogleIndexingService{
		products: googleIndexingProductReaderStub{
			item: &product.Product{ID: 3, Slug: "draft-wheel", Locale: "en", Status: "inactive"},
		},
		config: config.GoogleIndexingConfig{Enabled: true, ServiceAccountJSON: "configured"},
		credentials: googleIndexingCredentials{
			ClientEmail: "indexer@example.iam.gserviceaccount.com",
			PrivateKey:  privateKey,
		},
		publishURL: server.URL + "/publish",
	}
	service.configureDefaultHTTPClients()

	_, err := service.PushProduct(context.Background(), 3)

	require.ErrorIs(t, err, ErrGoogleIndexingProductNotPublic)
	require.Zero(t, upstreamRequests.Load(), "inactive products must not notify Google")
}

func TestGoogleIndexingServiceRetainsCooldownAfterPublishFailure(t *testing.T) {
	privateKey := mustGoogleIndexingTestPrivateKey(t)
	var publishRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/token":
			writer.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(writer, `{"access_token":"test-access-token","expires_in":3600}`)
		case "/publish":
			publishRequests.Add(1)
			http.Error(writer, "temporary upstream failure", http.StatusBadGateway)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	_, redisClient := newGoogleIndexingTestRedis(t)

	indexingService := &GoogleIndexingService{
		products: googleIndexingProductReaderStub{
			item: &product.Product{ID: 8, Slug: "carbon-wheel", Locale: "en", Status: "active"},
		},
		config: config.GoogleIndexingConfig{
			Enabled:               true,
			ServiceAccountJSON:    `{"client_email":"indexer@example.iam.gserviceaccount.com"}`,
			RequestTimeoutSeconds: 2,
		},
		storefrontURL: "https://store.example.test",
		credentials: googleIndexingCredentials{
			ClientEmail: "indexer@example.iam.gserviceaccount.com",
			PrivateKey:  privateKey,
			TokenURI:    server.URL + "/token",
		},
		publishURL:  server.URL + "/publish",
		redisClient: redisClient,
	}
	indexingService.configureDefaultHTTPClients()

	_, err := indexingService.PushProduct(context.Background(), 8)
	require.ErrorIs(t, err, ErrGoogleIndexingUpstream)
	require.Equal(t, int32(1), publishRequests.Load())

	_, err = indexingService.PushProduct(context.Background(), 8)
	require.ErrorIs(t, err, ErrGoogleIndexingRecentlyNotified)
	require.Equal(t, int32(1), publishRequests.Load(), "a failed publish attempt must still block duplicate notification")
}

func TestGoogleIndexingServiceReleasesCooldownWhenTokenAcquisitionFails(t *testing.T) {
	privateKey := mustGoogleIndexingTestPrivateKey(t)
	var tokenRequests atomic.Int32
	var publishRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/token":
			tokenRequests.Add(1)
			http.Error(writer, "temporary token failure", http.StatusBadGateway)
		case "/publish":
			publishRequests.Add(1)
			http.Error(writer, "publish must not be reached", http.StatusInternalServerError)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	_, redisClient := newGoogleIndexingTestRedis(t)

	indexingService := &GoogleIndexingService{
		products: googleIndexingProductReaderStub{
			item: &product.Product{ID: 9, Slug: "carbon-wheel", Locale: "en", Status: "active"},
		},
		config: config.GoogleIndexingConfig{
			Enabled:               true,
			ServiceAccountJSON:    `{"client_email":"indexer@example.iam.gserviceaccount.com"}`,
			RequestTimeoutSeconds: 2,
		},
		storefrontURL: "https://store.example.test",
		credentials: googleIndexingCredentials{
			ClientEmail: "indexer@example.iam.gserviceaccount.com",
			PrivateKey:  privateKey,
			TokenURI:    server.URL + "/token",
		},
		publishURL:  server.URL + "/publish",
		redisClient: redisClient,
	}
	indexingService.configureDefaultHTTPClients()

	_, err := indexingService.PushProduct(context.Background(), 9)
	require.ErrorIs(t, err, ErrGoogleIndexingUpstream)

	_, err = indexingService.PushProduct(context.Background(), 9)
	require.ErrorIs(t, err, ErrGoogleIndexingUpstream)
	require.Equal(t, int32(2), tokenRequests.Load(), "token failures should be retryable")
	require.Zero(t, publishRequests.Load(), "token failures must not reach publish")
}

func TestGoogleIndexingServiceRequiresRedisDuplicateProtection(t *testing.T) {
	privateKey := mustGoogleIndexingTestPrivateKey(t)
	var upstreamRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		upstreamRequests.Add(1)
		http.Error(writer, "unexpected upstream request", http.StatusInternalServerError)
	}))
	defer server.Close()

	service := &GoogleIndexingService{
		products: googleIndexingProductReaderStub{
			item: &product.Product{ID: 4, Slug: "carbon-wheel", Locale: "en", Status: "active"},
		},
		config: config.GoogleIndexingConfig{Enabled: true, ServiceAccountJSON: "configured"},
		credentials: googleIndexingCredentials{
			ClientEmail: "indexer@example.iam.gserviceaccount.com",
			PrivateKey:  privateKey,
			TokenURI:    server.URL + "/token",
		},
		storefrontURL: "https://store.example.test",
		publishURL:    server.URL + "/publish",
	}
	service.configureDefaultHTTPClients()

	_, err := service.PushProduct(context.Background(), 4)

	require.ErrorIs(t, err, ErrGoogleIndexingProtection)
	require.Zero(t, upstreamRequests.Load(), "missing duplicate protection must not notify Google")
}

func TestLoadGoogleIndexingCredentialsParsesServiceAccountJSON(t *testing.T) {
	privateKey := mustGoogleIndexingTestPrivateKey(t)
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	raw := `{"type":"service_account","private_key_id":"key-1","private_key":` +
		strconvQuote(string(privatePEM)) +
		`,"client_email":"indexer@example.iam.gserviceaccount.com","token_uri":"https://oauth2.example.test/token"}`

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

func newGoogleIndexingTestRedis(t *testing.T) (*miniredis.Miniredis, redis.UniversalClient) {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})
	return server, client
}
