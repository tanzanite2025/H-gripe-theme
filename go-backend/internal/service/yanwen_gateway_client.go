package service

import (
	"bytes"
	"context"
	"crypto/md5" // #nosec G501 -- Yanwen documents MD5 as its gateway signature.
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	yanwenCountryListMethod               = "common.country.getlist"
	yanwenWarehouseListMethod             = "common.warehouse.getlist"
	yanwenProductListMethod               = "express.channel.getlist"
	yanwenCreateOrderMethod               = "express.order.create"
	yanwenCancelOrderMethod               = "express.order.cancel"
	yanwenGetOrderMethod                  = "express.order.get"
	yanwenGetOrderListMethod              = "express.order.getlist"
	yanwenGetOrderLabelMethod             = "express.order.label.get"
	yanwenVerifyKoreaPCCCMethod           = "common.verify.kr.pccc"
	yanwenVerifyUnitedStatesAddressMethod = "common.verify.us.address"
	yanwenAPIVersion                      = "V1.0"
	yanwenDefaultEndpoint                 = "https://open.yw56.com.cn/api/order"
	yanwenFATEndpoint                     = "https://open-fat.yw56.com.cn/api/order"
	yanwenRequestTimeout                  = 20 * time.Second
	yanwenResponseLimit                   = 4 << 20
	yanwenGetOrderListMaximumIdentifiers  = 50
)

type yanwenGatewayCredentials struct {
	environment string
	endpoint    string
	userID      string
	apiToken    string
}

type YanwenGatewayClient struct {
	httpClient *http.Client
	now        func() time.Time
}

type YanwenCreateOrderRequest struct {
	ChannelID         string                   `json:"channelId"`
	OrderSource       string                   `json:"orderSource"`
	OrderNumber       string                   `json:"orderNumber"`
	TransactionNumber string                   `json:"transactionNumber,omitempty"`
	DateOfReceipt     string                   `json:"dateOfReceipt,omitempty"`
	CompanyCode       string                   `json:"companyCode"`
	ReceiverInfo      YanwenReceiverInfo       `json:"receiverInfo"`
	ParcelInfo        YanwenParcelInfo         `json:"parcelInfo"`
	ImportCustomsInfo *YanwenImportCustomsInfo `json:"importCustomsInfo,omitempty"`
	SalesPlatform     string                   `json:"salesPlatform,omitempty"`
}

type YanwenReceiverInfo struct {
	Name         string `json:"name"`
	Phone        string `json:"phone,omitempty"`
	Email        string `json:"email,omitempty"`
	Company      string `json:"company,omitempty"`
	Country      string `json:"country"`
	State        string `json:"state,omitempty"`
	City         string `json:"city,omitempty"`
	ZipCode      string `json:"zipCode,omitempty"`
	HouseNumber  string `json:"houseNumber,omitempty"`
	Address      string `json:"address"`
	TaxNumber    string `json:"taxNumber,omitempty"`
	ShortAddress string `json:"shortAddress,omitempty"`
}

type YanwenParcelInfo struct {
	HasBattery      int                    `json:"hasBattery"`
	Currency        string                 `json:"currency"`
	TotalQuantity   int                    `json:"totalQuantity"`
	TotalWeight     int                    `json:"totalWeight"`
	Height          int                    `json:"height,omitempty"`
	Width           int                    `json:"width,omitempty"`
	Length          int                    `json:"length,omitempty"`
	IOSS            string                 `json:"ioss,omitempty"`
	FreightCharges  interface{}            `json:"freightCharges,omitempty"`
	FreightCurrency string                 `json:"freightCurrency,omitempty"`
	ProductList     []YanwenProductDeclare `json:"productList"`
}

// YanwenImportCustomsInfo contains only the import customs fields documented
// by Yanwen's express.order.create method. EORI belongs directly under this
// object; it is not a receiver tax number and must not be moved to a generic
// customs record.
type YanwenImportCustomsInfo struct {
	EORI string `json:"eori,omitempty"`
}

type YanwenProductDeclare struct {
	GoodsNameChinese string      `json:"goodsNameCh"`
	GoodsNameEnglish string      `json:"goodsNameEn"`
	Price            interface{} `json:"price"`
	PriceExport      interface{} `json:"priceExport"`
	HSCode           string      `json:"hscode,omitempty"`
	Material         string      `json:"material,omitempty"`
	Quantity         int         `json:"quantity"`
	Weight           int         `json:"weight"`
	SKU              string      `json:"sku,omitempty"`
}

type YanwenCreateOrderResponse struct {
	WaybillNumber     string
	OrderNumber       string
	YanwenOrderNumber string
	RawResponse       []byte
}

// YanwenKoreaPersonalCustomsClearanceCodeVerificationRequest is the exact
// receiverInfo payload documented by Yanwen's common.verify.kr.pccc method.
type YanwenKoreaPersonalCustomsClearanceCodeVerificationRequest struct {
	ReceiverInfo YanwenKoreaPersonalCustomsClearanceCodeReceiverInfo `json:"receiverInfo"`
}

type YanwenKoreaPersonalCustomsClearanceCodeReceiverInfo struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	TaxNumber string `json:"taxNumber"`
	ZipCode   string `json:"zipCode"`
}

// YanwenKoreaPersonalCustomsClearanceCodeVerificationResponse contains only
// the official response facts needed by the Yanwen customs service.
type YanwenKoreaPersonalCustomsClearanceCodeVerificationResponse struct {
	OfficialPassed bool
	Code           string
	Message        string
	RawResponse    []byte
}

// YanwenUnitedStatesAddressVerificationRequest is the exact receiverInfo
// payload documented by Yanwen's common.verify.us.address method.
type YanwenUnitedStatesAddressVerificationRequest struct {
	ReceiverInfo YanwenUnitedStatesAddressReceiverInfo `json:"receiverInfo"`
}

type YanwenUnitedStatesAddressReceiverInfo struct {
	Address string `json:"address"`
	ZipCode string `json:"zipCode"`
	City    string `json:"city"`
	State   string `json:"state"`
}

// YanwenUnitedStatesAddressVerificationResponse contains the official
// standardized address returned by the Yanwen endpoint.
type YanwenUnitedStatesAddressVerificationResponse struct {
	OfficialPassed     bool
	Code               string
	Message            string
	NormalizedAddress  string
	NormalizedCity     string
	NormalizedState    string
	NormalizedZipCode4 string
	NormalizedZipCode5 string
	RawResponse        []byte
}

// YanwenCancelOrderResponse is the validated result of
// express.order.cancel. Yanwen normally returns data:null for a successful
// cancellation, so the requested waybill number is retained alongside the
// original gateway response.
type YanwenCancelOrderResponse struct {
	WaybillNumber string
	RawResponse   []byte
}

// YanwenGetOrderResponse is the validated result of express.order.get.
type YanwenGetOrderResponse struct {
	WaybillNumber     string
	OrderNumber       string
	ReferenceNumber   string
	YanwenOrderNumber string
	OfficialStatus    int
	IsPrinted         bool
	RawResponse       []byte
}

type yanwenGetOrderDataRecord struct {
	WaybillNumber     string          `json:"waybillNumber"`
	OrderNumber       string          `json:"orderNumber"`
	ReferenceNumber   string          `json:"referenceNumber"`
	YanwenOrderNumber string          `json:"yanwenOrderNumber"`
	Status            json.RawMessage `json:"status"`
	IsPrint           json.RawMessage `json:"isPrint"`
}

// YanwenGetOrderLabelResponse is the validated result of
// express.order.label.get.
type YanwenGetOrderLabelResponse struct {
	WaybillNumber string
	IsSuccessful  bool
	ErrorMessage  string
	Base64String  string
	RawResponse   []byte
}

func NewYanwenGatewayClient() *YanwenGatewayClient {
	return &YanwenGatewayClient{
		httpClient: &http.Client{Timeout: yanwenRequestTimeout},
		now:        time.Now,
	}
}

func (c *YanwenGatewayClient) PingYanwenGateway(ctx context.Context, credentials yanwenGatewayCredentials) (time.Duration, error) {
	responseBody, latency, err := c.sendYanwenGatewayRequest(ctx, credentials, yanwenCountryListMethod, []byte(`{}`))
	if err != nil {
		return 0, err
	}
	if err := parseYanwenGatewayResponse(responseBody); err != nil {
		return 0, err
	}
	return latency, nil
}

// VerifyKoreaPersonalCustomsClearanceCode calls Yanwen's official Korean
// PCCC verification method. A result is accepted only when the gateway
// envelope reports success=true and code=0.

// VerifyUnitedStatesAddress calls Yanwen's official US address verification
// method and accepts a result only when success=true, code=0, and all
// documented standardized receiver fields are present.

// CancelYanwenOrder asks the official gateway to cancel one waybill. A
// successful cancellation response is accepted when Yanwen returns
// data:null; the caller must query express.order.get separately before
// presenting the waybill as officially cancelled.

// GetYanwenOrderDetails queries one official Yanwen waybill and rejects incomplete
// or unknown status responses before they reach the service layer.

// GetYanwenOrderDetailsBatch queries up to 50 official Yanwen waybills or order
// numbers with express.order.getlist. Each returned record is validated using
// the same status and print-flag rules as the single-waybill query.

// GetYanwenOrderLabel retrieves one official Yanwen PDF label as Base64 data.

func (c *YanwenGatewayClient) sendYanwenGatewayRequest(ctx context.Context, credentials yanwenGatewayCredentials, method string, data []byte) ([]byte, time.Duration, error) {
	if len(data) == 0 {
		data = []byte(`{}`)
	}
	timestamp := strconv.FormatInt(c.getYanwenCurrentTime().UnixMilli(), 10)
	format := "json"
	publicRaw := credentials.userID + string(data) + format + method + timestamp + yanwenAPIVersion
	signature := signYanwenRequest(credentials.apiToken, publicRaw)

	endpoint, err := url.Parse(credentials.endpoint)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid Yanwen endpoint: %w", err)
	}
	query := endpoint.Query()
	query.Set("user_id", credentials.userID)
	query.Set("method", method)
	query.Set("format", format)
	query.Set("timestamp", timestamp)
	query.Set("version", yanwenAPIVersion)
	query.Set("sign", signature)
	endpoint.RawQuery = query.Encode()

	startedAt := time.Now()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(data))
	if err != nil {
		return nil, 0, fmt.Errorf("create Yanwen request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json;charset=utf-8")
	request.Header.Set("Accept", "application/json;charset=utf-8")

	client := c.httpClient
	if client == nil {
		client = &http.Client{Timeout: yanwenRequestTimeout}
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, 0, errors.New("Yanwen gateway request failed; check network access and the official endpoint")
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, yanwenResponseLimit))
	if err != nil {
		return nil, 0, fmt.Errorf("read Yanwen response: %w", err)
	}
	latency := time.Since(startedAt)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, 0, fmt.Errorf("Yanwen gateway returned HTTP %d", response.StatusCode)
	}
	return responseBody, latency, nil
}

type yanwenGatewayResponseEnvelope struct {
	Success bool            `json:"success"`
	Code    json.RawMessage `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func decodeYanwenGatewayResponse(responseBody []byte) (yanwenGatewayResponseEnvelope, error) {
	var response yanwenGatewayResponseEnvelope
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return response, fmt.Errorf("decode Yanwen response: %w", err)
	}
	return response, nil
}

func validateYanwenGatewayResponse(response yanwenGatewayResponseEnvelope) error {
	code := normalizeYanwenResponseCode(response.Code)
	if !response.Success {
		if response.Message == "" {
			response.Message = "Yanwen gateway reported an unsuccessful response"
		}
		return fmt.Errorf("Yanwen API error (%s): %s", code, response.Message)
	}
	if code != "0" {
		return fmt.Errorf("Yanwen API returned unexpected success code %s: %s", code, response.Message)
	}
	return nil
}

func parseYanwenGatewayResponse(responseBody []byte) error {
	response, err := decodeYanwenGatewayResponse(responseBody)
	if err != nil {
		return err
	}
	return validateYanwenGatewayResponse(response)
}

func normalizeYanwenResponseCode(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var stringCode string
	if json.Unmarshal(raw, &stringCode) == nil {
		return strings.TrimSpace(stringCode)
	}
	var numericCode int
	if json.Unmarshal(raw, &numericCode) == nil {
		return strconv.Itoa(numericCode)
	}
	return string(raw)
}

func signYanwenRequest(apiToken, raw string) string {
	digest := md5.Sum([]byte(apiToken + raw + apiToken))
	return hex.EncodeToString(digest[:])
}

func (c *YanwenGatewayClient) getYanwenCurrentTime() time.Time {
	if c != nil && c.now != nil {
		return c.now()
	}
	return time.Now()
}
