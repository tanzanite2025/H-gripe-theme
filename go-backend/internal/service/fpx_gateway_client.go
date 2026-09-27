package service

import (
	"bytes"
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type FpxGatewayClient struct {
	httpClient *http.Client
	now        func() time.Time
}

type FpxGatewayFetchResult struct {
	Channels []shipping.FpxChannel
	Scanned  int
	Latency  time.Duration
}

func NewFpxGatewayClient() *FpxGatewayClient {
	return &FpxGatewayClient{
		httpClient: &http.Client{Timeout: fpxRequestTimeout},
		now:        time.Now,
	}
}

// FetchChannels retrieves and parses the official channel directory.
func (c *FpxGatewayClient) FetchChannels(ctx context.Context, credentials fpxGatewayCredentials) (FpxGatewayFetchResult, error) {
	responseBody, latency, err := c.call(ctx, credentials)
	if err != nil {
		return FpxGatewayFetchResult{}, err
	}
	channels, scanned, err := parseFpxChannelResponseWithCount(responseBody)
	if err != nil {
		return FpxGatewayFetchResult{}, errors.New(redactFpxCredentials(err.Error(), credentials))
	}
	return FpxGatewayFetchResult{Channels: channels, Scanned: scanned, Latency: latency}, nil
}

// Ping shares the signed gateway request with channel sync but only validates
// the gateway result; it never requires or parses a channel directory payload.
func (c *FpxGatewayClient) Ping(ctx context.Context, credentials fpxGatewayCredentials) (time.Duration, error) {
	responseBody, latency, err := c.call(ctx, credentials)
	if err != nil {
		return 0, err
	}
	if _, err := parseFpxGatewayResponse(responseBody); err != nil {
		return 0, errors.New(redactFpxCredentials(err.Error(), credentials))
	}
	return latency, nil
}

// call signs and sends the same compact JSON bytes as the request body.
func (c *FpxGatewayClient) call(ctx context.Context, credentials fpxGatewayCredentials) ([]byte, time.Duration, error) {
	startedAt := time.Now()
	bodyBytes, err := json.Marshal(map[string]string{"transport_mode": "1"})
	if err != nil {
		return nil, 0, fmt.Errorf("encode 4PX request: %w", err)
	}

	timestamp := strconv.FormatInt(c.currentTime().UnixMilli(), 10)
	publicParams := map[string]string{
		"app_key":   credentials.appKey,
		"format":    "json",
		"method":    fpxChannelMethod,
		"timestamp": timestamp,
		"v":         fpxAPIVersion,
	}
	publicParams["sign"] = signFpxRequest(publicParams, bodyBytes, credentials.appSecret)

	endpoint, err := url.Parse(credentials.endpoint)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid 4PX endpoint: %w", err)
	}
	query := endpoint.Query()
	for key, value := range publicParams {
		query.Set(key, value)
	}
	if credentials.accessToken != "" {
		query.Set("access_token", credentials.accessToken)
	}
	endpoint.RawQuery = query.Encode()

	// Never log the constructed URL: it contains the signature and optional
	// Access Token. Credential values are represented only by fixed masks.
	logger.Debug("4PX gateway request",
		zap.String("environment", credentials.environment),
		zap.String("endpoint", credentials.endpoint),
		zap.String("method", fpxChannelMethod),
		zap.String("app_key", maskFpxCredential(credentials.appKey)),
		zap.String("app_secret", maskFpxCredential(credentials.appSecret)),
		zap.String("access_token", maskFpxCredential(credentials.accessToken)),
	)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("create 4PX request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json;charset=utf-8")
	request.Header.Set("Accept", "application/json;charset=utf-8")

	client := c.httpClient
	if client == nil {
		client = &http.Client{Timeout: fpxRequestTimeout}
	}
	response, err := client.Do(request)
	if err != nil {
		// net/http error strings can contain the URL and its signed query.
		logger.Warn("4PX gateway request failed",
			zap.String("environment", credentials.environment),
			zap.String("endpoint", credentials.endpoint),
			zap.Duration("latency", time.Since(startedAt)),
		)
		return nil, 0, errors.New("4PX gateway request failed; check network access and the official endpoint")
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, fpxResponseLimit))
	if err != nil {
		logger.Warn("4PX gateway response read failed",
			zap.String("environment", credentials.environment),
			zap.Int("http_status", response.StatusCode),
			zap.Duration("latency", time.Since(startedAt)),
		)
		return nil, 0, fmt.Errorf("read 4PX response: %w", err)
	}

	message, gatewayErrors := fpxResponseLogDetails(responseBody)
	message = redactFpxCredentials(message, credentials)
	gatewayErrors = redactFpxCredentials(gatewayErrors, credentials)
	latency := time.Since(startedAt)
	logger.Info("4PX gateway response",
		zap.String("environment", credentials.environment),
		zap.Int("http_status", response.StatusCode),
		zap.Duration("latency", latency),
		zap.String("msg", message),
		zap.String("errors", gatewayErrors),
	)

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, 0, fmt.Errorf("4PX gateway returned HTTP %d%s", response.StatusCode, fpxErrorSuffix(message, gatewayErrors))
	}
	return responseBody, latency, nil
}

func (c *FpxGatewayClient) currentTime() time.Time {
	if c != nil && c.now != nil {
		return c.now()
	}
	return time.Now()
}

func maskFpxCredential(value string) string {
	if value == "" {
		return ""
	}
	return "****"
}

func fpxResponseLogDetails(responseBody []byte) (message string, gatewayErrors string) {
	var response fpxChannelResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", "invalid JSON response"
	}
	return fpxResponseMessage(response), fpxResponseError(response)
}

func fpxErrorSuffix(message string, gatewayErrors string) string {
	if gatewayErrors != "" {
		return ": " + gatewayErrors
	}
	if message != "" {
		return ": " + message
	}
	return ""
}
