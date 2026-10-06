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

func TestYanwenUnitedStatesAddressVerificationServiceCallsOfficialVerificationContract(t *testing.T) {
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
		require.Equal(t, yanwenVerifyUnitedStatesAddressMethod, request.URL.Query().Get("method"))
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		require.JSONEq(t, `{"receiverInfo":{"address":"25815 North Cabernet Lane","zipCode":"86334","city":"Paulden","state":"AZ"}}`, string(body))
		raw := userID + string(body) + "json" + yanwenVerifyUnitedStatesAddressMethod + request.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"操作成功","data":{"receiverInfo":{"address":"25815 N CABERNET LN","city":"PAULDEN","state":"AZ","zipCode4":"3364","zipCode5":"86334"}}}`))
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

	addressVerificationService := NewYanwenUnitedStatesAddressVerificationService(apiService)
	result, err := addressVerificationService.VerifyUnitedStatesAddress(context.Background(), YanwenUnitedStatesAddressVerificationInput{
		Environment: "fat",
		Address:     "25815 North Cabernet Lane",
		ZipCode:     "86334",
		City:        "Paulden",
		State:       "AZ",
	})
	require.NoError(t, err)
	require.Equal(t, YanwenUnitedStatesAddressVerificationResult{
		Environment:        "fat",
		OfficialPassed:     true,
		Code:               "0",
		Message:            "操作成功",
		NormalizedAddress:  "25815 N CABERNET LN",
		NormalizedCity:     "PAULDEN",
		NormalizedState:    "AZ",
		NormalizedZipCode4: "3364",
		NormalizedZipCode5: "86334",
	}, result)
}

func TestYanwenUnitedStatesAddressVerificationRejectsSuccessfulResponseWithoutStandardizedFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"操作成功","data":{"receiverInfo":{"address":"25815 N CABERNET LN"}}}`))
	}))
	defer server.Close()

	client := NewYanwenGatewayClient()
	client.httpClient = server.Client()
	_, err := client.VerifyYanwenUnitedStatesAddress(context.Background(), yanwenGatewayCredentials{
		environment: "fat",
		endpoint:    server.URL,
		userID:      "test-user",
		apiToken:    "test-token",
	}, YanwenUnitedStatesAddressVerificationRequest{
		ReceiverInfo: YanwenUnitedStatesAddressReceiverInfo{
			Address: "25815 North Cabernet Lane",
			ZipCode: "86334",
			City:    "Paulden",
			State:   "AZ",
		},
	})
	require.ErrorContains(t, err, "missing standardized receiverInfo fields")
}
