package service

import (
	"bytes"
	"commerce-platform/internal/domain/shipping"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Yanwen official order, customs, cancellation, and label gateway methods.

func (c *YanwenGatewayClient) VerifyYanwenKoreaPersonalCustomsClearanceCode(
	ctx context.Context,
	credentials yanwenGatewayCredentials,
	request YanwenKoreaPersonalCustomsClearanceCodeVerificationRequest,
) (YanwenKoreaPersonalCustomsClearanceCodeVerificationResponse, error) {
	request.ReceiverInfo.Name = strings.TrimSpace(request.ReceiverInfo.Name)
	request.ReceiverInfo.Phone = strings.TrimSpace(request.ReceiverInfo.Phone)
	request.ReceiverInfo.TaxNumber = strings.TrimSpace(request.ReceiverInfo.TaxNumber)
	request.ReceiverInfo.ZipCode = strings.TrimSpace(request.ReceiverInfo.ZipCode)
	if request.ReceiverInfo.Name == "" || request.ReceiverInfo.Phone == "" || request.ReceiverInfo.TaxNumber == "" || request.ReceiverInfo.ZipCode == "" {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResponse{}, errors.New("Yanwen Korea PCCC verification requires name, phone, taxNumber and zipCode")
	}
	data, err := json.Marshal(request)
	if err != nil {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResponse{}, fmt.Errorf("encode Yanwen Korea PCCC verification request: %w", err)
	}
	responseBody, _, err := c.sendYanwenGatewayRequest(ctx, credentials, yanwenVerifyKoreaPCCCMethod, data)
	if err != nil {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResponse{}, err
	}
	response, err := decodeYanwenGatewayResponse(responseBody)
	if err != nil {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResponse{}, err
	}
	if err := validateYanwenGatewayResponse(response); err != nil {
		return YanwenKoreaPersonalCustomsClearanceCodeVerificationResponse{}, err
	}
	return YanwenKoreaPersonalCustomsClearanceCodeVerificationResponse{
		OfficialPassed: true,
		Code:           normalizeYanwenResponseCode(response.Code),
		Message:        strings.TrimSpace(response.Message),
		RawResponse:    append([]byte(nil), responseBody...),
	}, nil
}

func (c *YanwenGatewayClient) VerifyYanwenUnitedStatesAddress(
	ctx context.Context,
	credentials yanwenGatewayCredentials,
	request YanwenUnitedStatesAddressVerificationRequest,
) (YanwenUnitedStatesAddressVerificationResponse, error) {
	request.ReceiverInfo.Address = strings.TrimSpace(request.ReceiverInfo.Address)
	request.ReceiverInfo.ZipCode = strings.TrimSpace(request.ReceiverInfo.ZipCode)
	request.ReceiverInfo.City = strings.TrimSpace(request.ReceiverInfo.City)
	request.ReceiverInfo.State = strings.TrimSpace(request.ReceiverInfo.State)
	if request.ReceiverInfo.Address == "" || request.ReceiverInfo.ZipCode == "" || request.ReceiverInfo.City == "" || request.ReceiverInfo.State == "" {
		return YanwenUnitedStatesAddressVerificationResponse{}, errors.New("Yanwen US address verification requires address, zipCode, city and state")
	}
	data, err := json.Marshal(request)
	if err != nil {
		return YanwenUnitedStatesAddressVerificationResponse{}, fmt.Errorf("encode Yanwen US address verification request: %w", err)
	}
	responseBody, _, err := c.sendYanwenGatewayRequest(ctx, credentials, yanwenVerifyUnitedStatesAddressMethod, data)
	if err != nil {
		return YanwenUnitedStatesAddressVerificationResponse{}, err
	}
	response, err := decodeYanwenGatewayResponse(responseBody)
	if err != nil {
		return YanwenUnitedStatesAddressVerificationResponse{}, err
	}
	if err := validateYanwenGatewayResponse(response); err != nil {
		return YanwenUnitedStatesAddressVerificationResponse{}, err
	}
	if len(response.Data) == 0 || strings.EqualFold(strings.TrimSpace(string(response.Data)), "null") {
		return YanwenUnitedStatesAddressVerificationResponse{}, errors.New("Yanwen US address verification response is missing data")
	}
	var dataRecord struct {
		ReceiverInfo struct {
			Address  string `json:"address"`
			City     string `json:"city"`
			State    string `json:"state"`
			ZipCode4 string `json:"zipCode4"`
			ZipCode5 string `json:"zipCode5"`
		} `json:"receiverInfo"`
	}
	if err := json.Unmarshal(response.Data, &dataRecord); err != nil {
		return YanwenUnitedStatesAddressVerificationResponse{}, fmt.Errorf("decode Yanwen US address verification data: %w", err)
	}
	dataRecord.ReceiverInfo.Address = strings.TrimSpace(dataRecord.ReceiverInfo.Address)
	dataRecord.ReceiverInfo.City = strings.TrimSpace(dataRecord.ReceiverInfo.City)
	dataRecord.ReceiverInfo.State = strings.TrimSpace(dataRecord.ReceiverInfo.State)
	dataRecord.ReceiverInfo.ZipCode4 = strings.TrimSpace(dataRecord.ReceiverInfo.ZipCode4)
	dataRecord.ReceiverInfo.ZipCode5 = strings.TrimSpace(dataRecord.ReceiverInfo.ZipCode5)
	if dataRecord.ReceiverInfo.Address == "" || dataRecord.ReceiverInfo.City == "" || dataRecord.ReceiverInfo.State == "" || dataRecord.ReceiverInfo.ZipCode4 == "" || dataRecord.ReceiverInfo.ZipCode5 == "" {
		return YanwenUnitedStatesAddressVerificationResponse{}, errors.New("Yanwen US address verification response is missing standardized receiverInfo fields")
	}
	return YanwenUnitedStatesAddressVerificationResponse{
		OfficialPassed:     true,
		Code:               normalizeYanwenResponseCode(response.Code),
		Message:            strings.TrimSpace(response.Message),
		NormalizedAddress:  dataRecord.ReceiverInfo.Address,
		NormalizedCity:     dataRecord.ReceiverInfo.City,
		NormalizedState:    dataRecord.ReceiverInfo.State,
		NormalizedZipCode4: dataRecord.ReceiverInfo.ZipCode4,
		NormalizedZipCode5: dataRecord.ReceiverInfo.ZipCode5,
		RawResponse:        append([]byte(nil), responseBody...),
	}, nil
}

func (c *YanwenGatewayClient) CreateYanwenOrder(
	ctx context.Context,
	credentials yanwenGatewayCredentials,
	request YanwenCreateOrderRequest,
) (YanwenCreateOrderResponse, error) {
	if strings.TrimSpace(request.ChannelID) == "" || strings.TrimSpace(request.OrderSource) == "" || strings.TrimSpace(request.OrderNumber) == "" {
		return YanwenCreateOrderResponse{}, errors.New("Yanwen create order requires channelId, orderSource and orderNumber")
	}
	data, err := json.Marshal(request)
	if err != nil {
		return YanwenCreateOrderResponse{}, fmt.Errorf("encode Yanwen create order request: %w", err)
	}
	responseBody, _, err := c.sendYanwenGatewayRequest(ctx, credentials, yanwenCreateOrderMethod, data)
	if err != nil {
		return YanwenCreateOrderResponse{}, err
	}
	response, err := decodeYanwenGatewayResponse(responseBody)
	if err != nil {
		return YanwenCreateOrderResponse{}, err
	}
	if err := validateYanwenGatewayResponse(response); err != nil {
		return YanwenCreateOrderResponse{}, err
	}
	if len(response.Data) == 0 || strings.EqualFold(strings.TrimSpace(string(response.Data)), "null") {
		return YanwenCreateOrderResponse{}, errors.New("Yanwen create order response is missing data")
	}
	var dataRecord struct {
		WaybillNumber     string `json:"waybillNumber"`
		OrderNumber       string `json:"orderNumber"`
		YanwenOrderNumber string `json:"yanwenOrderNumber"`
	}
	if err := json.Unmarshal(response.Data, &dataRecord); err != nil {
		return YanwenCreateOrderResponse{}, fmt.Errorf("decode Yanwen create order data: %w", err)
	}
	dataRecord.WaybillNumber = strings.TrimSpace(dataRecord.WaybillNumber)
	dataRecord.OrderNumber = strings.TrimSpace(dataRecord.OrderNumber)
	dataRecord.YanwenOrderNumber = strings.TrimSpace(dataRecord.YanwenOrderNumber)
	if dataRecord.WaybillNumber == "" {
		return YanwenCreateOrderResponse{}, errors.New("Yanwen create order response is missing waybillNumber")
	}
	if dataRecord.OrderNumber != "" && dataRecord.OrderNumber != request.OrderNumber {
		return YanwenCreateOrderResponse{}, fmt.Errorf("Yanwen create order response order number %q does not match %q", dataRecord.OrderNumber, request.OrderNumber)
	}
	return YanwenCreateOrderResponse{
		WaybillNumber:     dataRecord.WaybillNumber,
		OrderNumber:       dataRecord.OrderNumber,
		YanwenOrderNumber: dataRecord.YanwenOrderNumber,
		RawResponse:       append([]byte(nil), responseBody...),
	}, nil
}

func (c *YanwenGatewayClient) CancelYanwenOrder(
	ctx context.Context,
	credentials yanwenGatewayCredentials,
	waybillNumber string,
	note string,
) (YanwenCancelOrderResponse, error) {
	waybillNumber = strings.TrimSpace(waybillNumber)
	if waybillNumber == "" {
		return YanwenCancelOrderResponse{}, errors.New("Yanwen cancel order requires waybillNumber")
	}
	requestPayload := struct {
		WaybillNumber string `json:"waybillNumber"`
		Note          string `json:"note,omitempty"`
	}{
		WaybillNumber: waybillNumber,
		Note:          strings.TrimSpace(note),
	}
	data, err := json.Marshal(requestPayload)
	if err != nil {
		return YanwenCancelOrderResponse{}, fmt.Errorf("encode Yanwen cancel order request: %w", err)
	}
	responseBody, _, err := c.sendYanwenGatewayRequest(ctx, credentials, yanwenCancelOrderMethod, data)
	if err != nil {
		return YanwenCancelOrderResponse{}, err
	}
	response, err := decodeYanwenGatewayResponse(responseBody)
	if err != nil {
		return YanwenCancelOrderResponse{}, err
	}
	if err := validateYanwenGatewayResponse(response); err != nil {
		return YanwenCancelOrderResponse{}, err
	}

	if len(response.Data) == 0 {
		return YanwenCancelOrderResponse{}, errors.New("Yanwen cancel order response is missing data")
	}
	// The documented success response contains data:null. If a future
	// response includes an object, only accept it when any returned waybill
	// number agrees with the request; never let a mismatched remote identity
	// be treated as a successful cancellation for this local record.
	if len(response.Data) > 0 && !strings.EqualFold(strings.TrimSpace(string(response.Data)), "null") {
		var dataRecord struct {
			WaybillNumber string `json:"waybillNumber"`
		}
		if err := json.Unmarshal(response.Data, &dataRecord); err != nil {
			return YanwenCancelOrderResponse{}, fmt.Errorf("decode Yanwen cancel order data: %w", err)
		}
		dataRecord.WaybillNumber = strings.TrimSpace(dataRecord.WaybillNumber)
		if dataRecord.WaybillNumber != "" && dataRecord.WaybillNumber != waybillNumber {
			return YanwenCancelOrderResponse{}, fmt.Errorf("Yanwen cancel order response waybill number %q does not match %q", dataRecord.WaybillNumber, waybillNumber)
		}
	}
	return YanwenCancelOrderResponse{
		WaybillNumber: waybillNumber,
		RawResponse:   append([]byte(nil), responseBody...),
	}, nil
}

func (c *YanwenGatewayClient) GetYanwenOrderDetails(
	ctx context.Context,
	credentials yanwenGatewayCredentials,
	waybillNumber string,
) (YanwenGetOrderResponse, error) {
	waybillNumber = strings.TrimSpace(waybillNumber)
	if waybillNumber == "" {
		return YanwenGetOrderResponse{}, errors.New("Yanwen get order requires waybillNumber")
	}
	data, err := json.Marshal(struct {
		WaybillNumber string `json:"waybillNumber"`
	}{WaybillNumber: waybillNumber})
	if err != nil {
		return YanwenGetOrderResponse{}, fmt.Errorf("encode Yanwen get order request: %w", err)
	}
	responseBody, _, err := c.sendYanwenGatewayRequest(ctx, credentials, yanwenGetOrderMethod, data)
	if err != nil {
		return YanwenGetOrderResponse{}, err
	}
	response, err := decodeYanwenGatewayResponse(responseBody)
	if err != nil {
		return YanwenGetOrderResponse{}, err
	}
	if err := validateYanwenGatewayResponse(response); err != nil {
		return YanwenGetOrderResponse{}, err
	}
	if len(response.Data) == 0 || strings.EqualFold(strings.TrimSpace(string(response.Data)), "null") {
		return YanwenGetOrderResponse{}, errors.New("Yanwen get order response is missing data")
	}
	var dataRecord yanwenGetOrderDataRecord
	if err := json.Unmarshal(response.Data, &dataRecord); err != nil {
		return YanwenGetOrderResponse{}, fmt.Errorf("decode Yanwen get order data: %w", err)
	}
	parsedRecord, err := parseYanwenGetOrderDataRecord(dataRecord)
	if err != nil {
		return YanwenGetOrderResponse{}, err
	}
	parsedRecord.RawResponse = append([]byte(nil), responseBody...)
	return parsedRecord, nil
}

func (c *YanwenGatewayClient) GetYanwenOrderDetailsBatch(
	ctx context.Context,
	credentials yanwenGatewayCredentials,
	identifiers []string,
) ([]YanwenGetOrderResponse, error) {
	if len(identifiers) == 0 {
		return nil, errors.New("Yanwen get order list requires at least one waybillNumber or orderNumber")
	}
	if len(identifiers) > yanwenGetOrderListMaximumIdentifiers {
		return nil, fmt.Errorf("Yanwen get order list supports at most %d identifiers", yanwenGetOrderListMaximumIdentifiers)
	}
	normalizedIdentifiers := make([]string, 0, len(identifiers))
	seenIdentifiers := make(map[string]struct{}, len(identifiers))
	for index, identifier := range identifiers {
		identifier = strings.TrimSpace(identifier)
		if identifier == "" {
			return nil, fmt.Errorf("Yanwen get order list identifier %d is empty", index)
		}
		if _, exists := seenIdentifiers[identifier]; exists {
			return nil, fmt.Errorf("Yanwen get order list contains duplicate identifier %q", identifier)
		}
		seenIdentifiers[identifier] = struct{}{}
		normalizedIdentifiers = append(normalizedIdentifiers, identifier)
	}
	data, err := json.Marshal(struct {
		ListNumber []string `json:"listNumber"`
	}{ListNumber: normalizedIdentifiers})
	if err != nil {
		return nil, fmt.Errorf("encode Yanwen get order list request: %w", err)
	}
	responseBody, _, err := c.sendYanwenGatewayRequest(ctx, credentials, yanwenGetOrderListMethod, data)
	if err != nil {
		return nil, err
	}
	response, err := decodeYanwenGatewayResponse(responseBody)
	if err != nil {
		return nil, err
	}
	if err := validateYanwenGatewayResponse(response); err != nil {
		return nil, err
	}
	if len(response.Data) == 0 || strings.EqualFold(strings.TrimSpace(string(response.Data)), "null") {
		return nil, errors.New("Yanwen get order list response is missing data")
	}
	var dataRecords []yanwenGetOrderDataRecord
	if err := json.Unmarshal(response.Data, &dataRecords); err != nil {
		return nil, fmt.Errorf("decode Yanwen get order list data: %w", err)
	}
	results := make([]YanwenGetOrderResponse, 0, len(dataRecords))
	seenWaybillNumbers := make(map[string]struct{}, len(dataRecords))
	for index, dataRecord := range dataRecords {
		parsedRecord, err := parseYanwenGetOrderDataRecord(dataRecord)
		if err != nil {
			return nil, fmt.Errorf("decode Yanwen get order list data record %d: %w", index, err)
		}
		if _, exists := seenWaybillNumbers[parsedRecord.WaybillNumber]; exists {
			return nil, fmt.Errorf("Yanwen get order list response contains duplicate waybillNumber %q", parsedRecord.WaybillNumber)
		}
		seenWaybillNumbers[parsedRecord.WaybillNumber] = struct{}{}
		parsedRecord.RawResponse = append([]byte(nil), responseBody...)
		results = append(results, parsedRecord)
	}
	return results, nil
}

func parseYanwenGetOrderDataRecord(dataRecord yanwenGetOrderDataRecord) (YanwenGetOrderResponse, error) {
	dataRecord.WaybillNumber = strings.TrimSpace(dataRecord.WaybillNumber)
	dataRecord.OrderNumber = strings.TrimSpace(dataRecord.OrderNumber)
	dataRecord.ReferenceNumber = strings.TrimSpace(dataRecord.ReferenceNumber)
	dataRecord.YanwenOrderNumber = strings.TrimSpace(dataRecord.YanwenOrderNumber)
	if dataRecord.WaybillNumber == "" {
		return YanwenGetOrderResponse{}, errors.New("Yanwen get order response is missing waybillNumber")
	}
	if dataRecord.OrderNumber == "" {
		return YanwenGetOrderResponse{}, errors.New("Yanwen get order response is missing orderNumber")
	}
	officialStatus, err := parseRequiredYanwenIntegerField(dataRecord.Status, "status")
	if err != nil {
		return YanwenGetOrderResponse{}, err
	}
	if !shipping.IsKnownYanwenOfficialWaybillStatus(officialStatus) {
		return YanwenGetOrderResponse{}, fmt.Errorf("Yanwen get order response contains unknown status %d", officialStatus)
	}
	isPrintedValue, err := parseRequiredYanwenIntegerField(dataRecord.IsPrint, "isPrint")
	if err != nil {
		return YanwenGetOrderResponse{}, err
	}
	if isPrintedValue != 0 && isPrintedValue != 1 {
		return YanwenGetOrderResponse{}, fmt.Errorf("Yanwen get order response contains invalid isPrint value %d", isPrintedValue)
	}
	return YanwenGetOrderResponse{
		WaybillNumber:     dataRecord.WaybillNumber,
		OrderNumber:       dataRecord.OrderNumber,
		ReferenceNumber:   dataRecord.ReferenceNumber,
		YanwenOrderNumber: dataRecord.YanwenOrderNumber,
		OfficialStatus:    officialStatus,
		IsPrinted:         isPrintedValue == 1,
	}, nil
}

func (c *YanwenGatewayClient) GetYanwenOrderLabel(
	ctx context.Context,
	credentials yanwenGatewayCredentials,
	waybillNumber string,
) (YanwenGetOrderLabelResponse, error) {
	waybillNumber = strings.TrimSpace(waybillNumber)
	if waybillNumber == "" {
		return YanwenGetOrderLabelResponse{}, errors.New("Yanwen get order label requires waybillNumber")
	}
	data, err := json.Marshal(struct {
		WaybillNumber string `json:"waybillNumber"`
	}{WaybillNumber: waybillNumber})
	if err != nil {
		return YanwenGetOrderLabelResponse{}, fmt.Errorf("encode Yanwen get order label request: %w", err)
	}
	responseBody, _, err := c.sendYanwenGatewayRequest(ctx, credentials, yanwenGetOrderLabelMethod, data)
	if err != nil {
		return YanwenGetOrderLabelResponse{}, err
	}
	response, err := decodeYanwenGatewayResponse(responseBody)
	if err != nil {
		return YanwenGetOrderLabelResponse{}, err
	}
	if err := validateYanwenGatewayResponse(response); err != nil {
		return YanwenGetOrderLabelResponse{}, err
	}
	if len(response.Data) == 0 || strings.EqualFold(strings.TrimSpace(string(response.Data)), "null") {
		return YanwenGetOrderLabelResponse{}, errors.New("Yanwen get order label response is missing data")
	}
	var dataRecord struct {
		WaybillNumber string          `json:"waybillNumber"`
		IsSuccessful  json.RawMessage `json:"isSuccess"`
		ErrorMessage  string          `json:"errorMsg"`
		Base64String  string          `json:"base64String"`
	}
	if err := json.Unmarshal(response.Data, &dataRecord); err != nil {
		return YanwenGetOrderLabelResponse{}, fmt.Errorf("decode Yanwen get order label data: %w", err)
	}
	dataRecord.WaybillNumber = strings.TrimSpace(dataRecord.WaybillNumber)
	dataRecord.ErrorMessage = strings.TrimSpace(dataRecord.ErrorMessage)
	dataRecord.Base64String = strings.TrimSpace(dataRecord.Base64String)
	if dataRecord.WaybillNumber == "" {
		return YanwenGetOrderLabelResponse{}, errors.New("Yanwen get order label response is missing waybillNumber")
	}
	if dataRecord.WaybillNumber != waybillNumber {
		return YanwenGetOrderLabelResponse{}, fmt.Errorf("Yanwen get order label response waybill number %q does not match %q", dataRecord.WaybillNumber, waybillNumber)
	}
	if len(dataRecord.IsSuccessful) == 0 || strings.EqualFold(strings.TrimSpace(string(dataRecord.IsSuccessful)), "null") {
		return YanwenGetOrderLabelResponse{}, errors.New("Yanwen get order label response is missing isSuccess")
	}
	var isSuccessful bool
	if err := json.Unmarshal(dataRecord.IsSuccessful, &isSuccessful); err != nil {
		return YanwenGetOrderLabelResponse{}, fmt.Errorf("Yanwen get order label response has invalid isSuccess: %w", err)
	}
	if !isSuccessful {
		if dataRecord.ErrorMessage == "" {
			dataRecord.ErrorMessage = "Yanwen did not generate a label"
		}
		return YanwenGetOrderLabelResponse{}, fmt.Errorf("Yanwen get order label failed: %s", dataRecord.ErrorMessage)
	}
	if dataRecord.Base64String == "" {
		return YanwenGetOrderLabelResponse{}, errors.New("Yanwen get order label response is missing base64String")
	}
	decodedLabel, err := base64.StdEncoding.DecodeString(dataRecord.Base64String)
	if err != nil {
		return YanwenGetOrderLabelResponse{}, fmt.Errorf("Yanwen get order label response has invalid base64String: %w", err)
	}
	if len(decodedLabel) == 0 || !bytes.HasPrefix(decodedLabel, []byte("%PDF")) {
		return YanwenGetOrderLabelResponse{}, errors.New("Yanwen get order label response is not a PDF file")
	}
	return YanwenGetOrderLabelResponse{
		WaybillNumber: dataRecord.WaybillNumber,
		IsSuccessful:  true,
		ErrorMessage:  dataRecord.ErrorMessage,
		Base64String:  dataRecord.Base64String,
		RawResponse:   append([]byte(nil), responseBody...),
	}, nil
}
