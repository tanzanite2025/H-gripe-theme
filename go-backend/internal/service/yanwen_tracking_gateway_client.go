package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"commerce-platform/internal/domain/shipping"
)

const (
	yanwenTrackingEndpoint           = "http://api.track.yw56.com.cn/api/tracking"
	yanwenTrackingMaximumIdentifiers = 30
	yanwenTrackingResponseLimit      = 8 << 20
)

// YanwenTrackingResult is one official result returned by the separate
// Yanwen tracking endpoint. Official fields remain intact; tracking_stage,
// tracking_stage_rank and tracking_has_exception are explicitly marked local
// display classifications and never replace the official status code.
type YanwenTrackingResult struct {
	TrackingNumber           string                       `json:"tracking_number"`
	WaybillNumber            string                       `json:"waybill_number"`
	ExchangeNumber           string                       `json:"exchange_number"`
	LastMileCarrier          string                       `json:"last_mile_carrier"`
	LastMileCarrierWebsite   string                       `json:"last_mile_carrier_website"`
	LastMileCarrierContact   string                       `json:"last_mile_carrier_contact_number"`
	Checkpoints              []YanwenTrackingCheckpoint   `json:"checkpoints"`
	TrackingStatus           string                       `json:"tracking_status"`
	TrackingStage            shipping.YanwenTrackingStage `json:"tracking_stage"`
	TrackingStageRank        int                          `json:"tracking_stage_rank"`
	TrackingHasException     bool                         `json:"tracking_has_exception"`
	TrackingStatusWaybill    YanwenTrackingStatusWaybill  `json:"tracking_status_waybill"`
	LastMileTrackingExpected bool                         `json:"last_mile_tracking_expected"`
	OriginCountry            string                       `json:"origin_country"`
	DestinationCountry       string                       `json:"destination_country"`
	RawResponse              []byte                       `json:"-"`
}

// YanwenTrackingCheckpoint is one official checkpoint. TimeStamp and
// TimeZone remain separate strings because the official API reports local
// time plus an offset such as "+08"; converting it would lose that contract.
type YanwenTrackingCheckpoint struct {
	TimeStamp            string                 `json:"time_stamp"`
	TimeZone             string                 `json:"time_zone"`
	TrackingStatus       string                 `json:"tracking_status"`
	Message              string                 `json:"message"`
	Location             string                 `json:"location"`
	IsLastMileCheckpoint bool                   `json:"is_last_mile_checkpoint"`
	ExtraProperties      map[string]interface{} `json:"extra_properties,omitempty"`
}

// YanwenTrackingStatusWaybill contains the official hierarchical status
// levels for a result. The values are strings in the documented response.
type YanwenTrackingStatusWaybill struct {
	Level1 string `json:"level1"`
	Level2 string `json:"level2"`
	Level3 string `json:"level3"`
}

type yanwenTrackingResponseEnvelope struct {
	Code    json.RawMessage `json:"code"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

type yanwenTrackingResultRecord struct {
	TrackingNumber           string          `json:"tracking_number"`
	WaybillNumber            string          `json:"waybill_number"`
	ExchangeNumber           string          `json:"exchange_number"`
	LastMileCarrier          string          `json:"last_mile_carrier"`
	LastMileCarrierWebsite   string          `json:"last_mile_carrier_website"`
	LastMileCarrierContact   string          `json:"last_mile_carrier_contact_number"`
	Checkpoints              json.RawMessage `json:"checkpoints"`
	TrackingStatus           string          `json:"tracking_status"`
	TrackingStatusWaybill    json.RawMessage `json:"tracking_status_waybill"`
	LastMileTrackingExpected json.RawMessage `json:"last_mile_tracking_expected"`
	OriginCountry            string          `json:"origin_country"`
	DestinationCountry       string          `json:"destination_country"`
}

type yanwenTrackingCheckpointRecord struct {
	TimeStamp            string          `json:"time_stamp"`
	TimeZone             string          `json:"time_zone"`
	TrackingStatus       string          `json:"tracking_status"`
	Message              string          `json:"message"`
	Location             string          `json:"location"`
	IsLastMileCheckpoint json.RawMessage `json:"is_last_mile_checkpoint"`
	ExtraProperties      json.RawMessage `json:"extraProperties"`
}

type yanwenTrackingStatusWaybillRecord struct {
	Level1 json.RawMessage `json:"level1"`
	Level2 json.RawMessage `json:"level2"`
	Level3 json.RawMessage `json:"level3"`
}

// QueryYanwenTracking calls the documented production-only tracking endpoint.
// The endpoint authenticates with the merchant or order-creation account, not
// the signed order API token.
func (c *YanwenGatewayClient) QueryYanwenTracking(ctx context.Context, authorization string, trackingNumbers []string) ([]YanwenTrackingResult, error) {
	normalizedNumbers, err := normalizeYanwenTrackingNumbers(trackingNumbers)
	if err != nil {
		return nil, err
	}
	authorization = strings.TrimSpace(authorization)
	if authorization == "" {
		return nil, errors.New("Yanwen tracking authorization is required")
	}

	trackingEndpoint, err := url.Parse(yanwenTrackingEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid Yanwen tracking endpoint: %w", err)
	}
	query := trackingEndpoint.Query()
	query.Set("nums", strings.Join(normalizedNumbers, ","))
	trackingEndpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, trackingEndpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create Yanwen tracking request: %w", err)
	}
	request.Header.Set("Authorization", authorization)
	request.Header.Set("Accept", "application/json")

	client := c.httpClient
	if client == nil {
		client = &http.Client{Timeout: yanwenRequestTimeout}
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, errors.New("Yanwen tracking request failed; check network access to the official tracking endpoint")
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, yanwenTrackingResponseLimit))
	if err != nil {
		return nil, fmt.Errorf("read Yanwen tracking response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Yanwen tracking endpoint returned HTTP %d", response.StatusCode)
	}
	results, err := parseYanwenTrackingResponse(responseBody)
	if err != nil {
		return nil, err
	}
	requestedNumbers := make(map[string]struct{}, len(normalizedNumbers))
	for _, number := range normalizedNumbers {
		requestedNumbers[number] = struct{}{}
	}
	for _, result := range results {
		if _, requested := requestedNumbers[result.TrackingNumber]; !requested {
			return nil, fmt.Errorf("Yanwen tracking response returned unrequested tracking number %q", result.TrackingNumber)
		}
	}
	for index := range results {
		results[index].RawResponse = append([]byte(nil), responseBody...)
	}
	return results, nil
}

func normalizeYanwenTrackingNumbers(trackingNumbers []string) ([]string, error) {
	seen := make(map[string]struct{}, len(trackingNumbers))
	normalized := make([]string, 0, len(trackingNumbers))
	for _, rawNumber := range trackingNumbers {
		for _, number := range strings.FieldsFunc(rawNumber, func(r rune) bool {
			return r == ',' || r == '，' || unicode.IsSpace(r)
		}) {
			number = strings.TrimSpace(number)
			if number == "" {
				continue
			}
			if _, exists := seen[number]; exists {
				continue
			}
			seen[number] = struct{}{}
			normalized = append(normalized, number)
		}
	}
	if len(normalized) == 0 {
		return nil, errors.New("at least one Yanwen tracking number is required")
	}
	if len(normalized) > yanwenTrackingMaximumIdentifiers {
		return nil, fmt.Errorf("Yanwen tracking supports at most %d tracking numbers", yanwenTrackingMaximumIdentifiers)
	}
	return normalized, nil
}

func parseYanwenTrackingResponse(responseBody []byte) ([]YanwenTrackingResult, error) {
	var response yanwenTrackingResponseEnvelope
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("decode Yanwen tracking response: %w", err)
	}
	if len(response.Code) == 0 || strings.EqualFold(strings.TrimSpace(string(response.Code)), "null") {
		return nil, errors.New("Yanwen tracking response is missing code")
	}
	if normalizeYanwenResponseCode(response.Code) != "0" {
		message := strings.TrimSpace(response.Message)
		if message == "" {
			message = "Yanwen tracking endpoint reported an unsuccessful response"
		}
		return nil, fmt.Errorf("Yanwen tracking API error (%s): %s", normalizeYanwenResponseCode(response.Code), message)
	}
	if len(response.Result) == 0 || strings.EqualFold(strings.TrimSpace(string(response.Result)), "null") {
		return nil, errors.New("Yanwen tracking response is missing result")
	}
	var records []yanwenTrackingResultRecord
	if err := json.Unmarshal(response.Result, &records); err != nil {
		return nil, fmt.Errorf("decode Yanwen tracking result: %w", err)
	}
	results := make([]YanwenTrackingResult, 0, len(records))
	seenTrackingNumbers := make(map[string]struct{}, len(records))
	for index, record := range records {
		parsed, err := parseYanwenTrackingResultRecord(record)
		if err != nil {
			return nil, fmt.Errorf("decode Yanwen tracking result record %d: %w", index, err)
		}
		if _, exists := seenTrackingNumbers[parsed.TrackingNumber]; exists {
			return nil, fmt.Errorf("Yanwen tracking response contains duplicate tracking number %q", parsed.TrackingNumber)
		}
		seenTrackingNumbers[parsed.TrackingNumber] = struct{}{}
		results = append(results, parsed)
	}
	return results, nil
}

func parseYanwenTrackingResultRecord(record yanwenTrackingResultRecord) (YanwenTrackingResult, error) {
	record.TrackingNumber = strings.TrimSpace(record.TrackingNumber)
	record.WaybillNumber = strings.TrimSpace(record.WaybillNumber)
	record.ExchangeNumber = strings.TrimSpace(record.ExchangeNumber)
	record.LastMileCarrier = strings.TrimSpace(record.LastMileCarrier)
	record.LastMileCarrierWebsite = strings.TrimSpace(record.LastMileCarrierWebsite)
	record.LastMileCarrierContact = strings.TrimSpace(record.LastMileCarrierContact)
	record.TrackingStatus = strings.TrimSpace(record.TrackingStatus)
	record.OriginCountry = strings.ToUpper(strings.TrimSpace(record.OriginCountry))
	record.DestinationCountry = strings.ToUpper(strings.TrimSpace(record.DestinationCountry))
	if record.TrackingNumber == "" {
		return YanwenTrackingResult{}, errors.New("tracking_number is required")
	}
	if record.WaybillNumber == "" {
		return YanwenTrackingResult{}, errors.New("waybill_number is required")
	}
	if record.TrackingStatus == "" {
		return YanwenTrackingResult{}, errors.New("tracking_status is required")
	}
	if len(record.Checkpoints) == 0 || strings.EqualFold(strings.TrimSpace(string(record.Checkpoints)), "null") {
		return YanwenTrackingResult{}, errors.New("checkpoints is required")
	}
	var checkpointRecords []yanwenTrackingCheckpointRecord
	if err := json.Unmarshal(record.Checkpoints, &checkpointRecords); err != nil {
		return YanwenTrackingResult{}, fmt.Errorf("decode checkpoints: %w", err)
	}
	checkpoints := make([]YanwenTrackingCheckpoint, 0, len(checkpointRecords))
	for index, checkpoint := range checkpointRecords {
		parsedCheckpoint, err := parseYanwenTrackingCheckpoint(checkpoint)
		if err != nil {
			return YanwenTrackingResult{}, fmt.Errorf("decode checkpoint %d: %w", index, err)
		}
		checkpoints = append(checkpoints, parsedCheckpoint)
	}
	statusWaybill, err := parseYanwenTrackingStatusWaybill(record.TrackingStatusWaybill)
	if err != nil {
		return YanwenTrackingResult{}, err
	}
	checkpointStatuses := make([]string, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		checkpointStatuses = append(checkpointStatuses, checkpoint.TrackingStatus)
	}
	stage := shipping.ClassifyYanwenTrackingProgress(record.TrackingStatus, checkpointStatuses)
	lastMileTrackingExpected, err := parseYanwenTrackingBoolean(record.LastMileTrackingExpected, "last_mile_tracking_expected", false)
	if err != nil {
		return YanwenTrackingResult{}, err
	}
	return YanwenTrackingResult{
		TrackingNumber:           record.TrackingNumber,
		WaybillNumber:            record.WaybillNumber,
		ExchangeNumber:           record.ExchangeNumber,
		LastMileCarrier:          record.LastMileCarrier,
		LastMileCarrierWebsite:   record.LastMileCarrierWebsite,
		LastMileCarrierContact:   record.LastMileCarrierContact,
		Checkpoints:              checkpoints,
		TrackingStatus:           record.TrackingStatus,
		TrackingStage:            stage.Stage,
		TrackingStageRank:        stage.StageRank,
		TrackingHasException:     stage.IsException,
		TrackingStatusWaybill:    statusWaybill,
		LastMileTrackingExpected: lastMileTrackingExpected,
		OriginCountry:            record.OriginCountry,
		DestinationCountry:       record.DestinationCountry,
	}, nil
}

func parseYanwenTrackingCheckpoint(record yanwenTrackingCheckpointRecord) (YanwenTrackingCheckpoint, error) {
	record.TimeStamp = strings.TrimSpace(record.TimeStamp)
	record.TimeZone = strings.TrimSpace(record.TimeZone)
	record.TrackingStatus = strings.TrimSpace(record.TrackingStatus)
	record.Message = strings.TrimSpace(record.Message)
	record.Location = strings.TrimSpace(record.Location)
	if record.TimeStamp == "" || record.TimeZone == "" {
		return YanwenTrackingCheckpoint{}, errors.New("time_stamp and time_zone are required")
	}
	if record.TrackingStatus == "" {
		return YanwenTrackingCheckpoint{}, errors.New("tracking_status is required")
	}
	if record.Message == "" {
		return YanwenTrackingCheckpoint{}, errors.New("message is required")
	}
	isLastMileCheckpoint, err := parseYanwenTrackingBoolean(record.IsLastMileCheckpoint, "is_last_mile_checkpoint", true)
	if err != nil {
		return YanwenTrackingCheckpoint{}, err
	}
	extraProperties, err := parseYanwenTrackingExtraProperties(record.ExtraProperties)
	if err != nil {
		return YanwenTrackingCheckpoint{}, err
	}
	return YanwenTrackingCheckpoint{
		TimeStamp:            record.TimeStamp,
		TimeZone:             record.TimeZone,
		TrackingStatus:       record.TrackingStatus,
		Message:              record.Message,
		Location:             record.Location,
		IsLastMileCheckpoint: isLastMileCheckpoint,
		ExtraProperties:      extraProperties,
	}, nil
}

func parseYanwenTrackingStatusWaybill(raw json.RawMessage) (YanwenTrackingStatusWaybill, error) {
	if len(raw) == 0 || strings.EqualFold(strings.TrimSpace(string(raw)), "null") {
		return YanwenTrackingStatusWaybill{}, errors.New("tracking_status_waybill is required")
	}
	var record yanwenTrackingStatusWaybillRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		return YanwenTrackingStatusWaybill{}, fmt.Errorf("decode tracking_status_waybill: %w", err)
	}
	level1, err := parseYanwenTrackingText(record.Level1, "tracking_status_waybill.level1", true)
	if err != nil {
		return YanwenTrackingStatusWaybill{}, err
	}
	level2, err := parseYanwenTrackingText(record.Level2, "tracking_status_waybill.level2", true)
	if err != nil {
		return YanwenTrackingStatusWaybill{}, err
	}
	level3, err := parseYanwenTrackingText(record.Level3, "tracking_status_waybill.level3", true)
	if err != nil {
		return YanwenTrackingStatusWaybill{}, err
	}
	return YanwenTrackingStatusWaybill{Level1: level1, Level2: level2, Level3: level3}, nil
}

func parseYanwenTrackingText(raw json.RawMessage, fieldName string, required bool) (string, error) {
	if len(raw) == 0 || strings.EqualFold(strings.TrimSpace(string(raw)), "null") {
		if required {
			return "", fmt.Errorf("%s is required", fieldName)
		}
		return "", nil
	}
	var stringValue string
	if json.Unmarshal(raw, &stringValue) == nil {
		stringValue = strings.TrimSpace(stringValue)
		if required && stringValue == "" {
			return "", fmt.Errorf("%s is required", fieldName)
		}
		return stringValue, nil
	}
	var numberValue json.Number
	if json.Unmarshal(raw, &numberValue) == nil {
		return numberValue.String(), nil
	}
	return "", fmt.Errorf("%s must be a string or number", fieldName)
}

func parseYanwenTrackingBoolean(raw json.RawMessage, fieldName string, required bool) (bool, error) {
	if len(raw) == 0 || strings.EqualFold(strings.TrimSpace(string(raw)), "null") {
		if required {
			return false, fmt.Errorf("%s is required", fieldName)
		}
		return false, nil
	}
	var boolValue bool
	if json.Unmarshal(raw, &boolValue) == nil {
		return boolValue, nil
	}
	var stringValue string
	if json.Unmarshal(raw, &stringValue) == nil {
		switch strings.TrimSpace(strings.ToLower(stringValue)) {
		case "1", "true":
			return true, nil
		case "0", "false":
			return false, nil
		}
	}
	var numberValue int
	if json.Unmarshal(raw, &numberValue) == nil {
		if numberValue == 1 {
			return true, nil
		}
		if numberValue == 0 {
			return false, nil
		}
	}
	return false, fmt.Errorf("%s must be a boolean or 0/1", fieldName)
}

func parseYanwenTrackingExtraProperties(raw json.RawMessage) (map[string]interface{}, error) {
	if len(raw) == 0 || strings.EqualFold(strings.TrimSpace(string(raw)), "null") {
		return nil, nil
	}
	var properties map[string]interface{}
	if err := json.Unmarshal(raw, &properties); err != nil {
		return nil, fmt.Errorf("extraProperties must be an object: %w", err)
	}
	return properties, nil
}
