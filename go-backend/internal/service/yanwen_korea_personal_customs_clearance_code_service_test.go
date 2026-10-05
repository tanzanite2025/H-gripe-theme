package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestYanwenKoreaPersonalCustomsClearanceCodeServiceCallsOfficialVerificationContract(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}))

	const userID = "stored-yanwen-user"
	const apiToken = "stored-yanwen-token"
	fixedTime := time.UnixMilli(1790000000123)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, yanwenVerifyKoreaPCCCMethod, request.URL.Query().Get("method"))
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		require.JSONEq(t, `{"receiverInfo":{"name":"김하성","phone":"01075654552","taxNumber":"P972150029777","zipCode":"03603"}}`, string(body))
		raw := userID + string(body) + "json" + yanwenVerifyKoreaPCCCMethod + request.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"操作成功","data":null}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepository := repository.NewYanwenAPIConfigRepository(db)
	apiService := NewYanwenAPIService(configRepository, nil)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	apiService.gateway.now = func() time.Time { return fixedTime }
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{
		Environment: "fat",
		Endpoint:    yanwenFATEndpoint,
		UserID:      userID,
		APIToken:    apiToken,
		Enabled:     true,
	}))

	customsService := NewYanwenKoreaPersonalCustomsClearanceCodeService(apiService)
	result, err := customsService.VerifyKoreaPersonalCustomsClearanceCode(context.Background(), YanwenKoreaPersonalCustomsClearanceCodeVerificationInput{
		Environment:   "fat",
		RecipientName: "김하성",
		Phone:         "01075654552",
		TaxNumber:     "P972150029777",
		PostalCode:    "03603",
	})
	require.NoError(t, err)
	require.Equal(t, YanwenKoreaPersonalCustomsClearanceCodeVerificationResult{
		Environment:    "fat",
		OfficialPassed: true,
		Code:           "0",
		Message:        "操作成功",
	}, result)
}

func TestYanwenKoreaPersonalCustomsClearanceCodeServiceRejectsNonFiveDigitPostalCodeBeforeGatewayCall(t *testing.T) {
	apiService := NewYanwenAPIService(nil, nil)
	customsService := NewYanwenKoreaPersonalCustomsClearanceCodeService(apiService)
	_, err := customsService.VerifyKoreaPersonalCustomsClearanceCode(context.Background(), YanwenKoreaPersonalCustomsClearanceCodeVerificationInput{
		Environment:   "fat",
		RecipientName: "김하성",
		Phone:         "01075654552",
		TaxNumber:     "P972150029777",
		PostalCode:    "0360A",
	})
	require.ErrorContains(t, err, "postal code must contain exactly 5 digits")
}

func TestYanwenGatewayKoreaPersonalCustomsClearanceCodeDoesNotAcceptOfficialRejectionAsPassed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"success":false,"code":"1001","message":"PCCC 校验未通过","data":null}`))
	}))
	defer server.Close()

	client := NewYanwenGatewayClient()
	client.httpClient = server.Client()
	_, err := client.VerifyYanwenKoreaPersonalCustomsClearanceCode(context.Background(), yanwenGatewayCredentials{
		environment: "fat",
		endpoint:    server.URL,
		userID:      "test-user",
		apiToken:    "test-token",
	}, YanwenKoreaPersonalCustomsClearanceCodeVerificationRequest{
		ReceiverInfo: YanwenKoreaPersonalCustomsClearanceCodeReceiverInfo{
			Name:      "김하성",
			Phone:     "01075654552",
			TaxNumber: "P972150029777",
			ZipCode:   "03603",
		},
	})
	require.ErrorContains(t, err, "PCCC 校验未通过")
}
