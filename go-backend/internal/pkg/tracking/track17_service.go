package tracking

import (
	"bytes"
	"context"
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
	track17GetTrackInfoEndpoint = "/track/v2.4/gettrackinfo"
	track17RegisterEndpoint     = "/track/v2.4/register"
)

// track17Service 17TRACK V2.4 服务实现。
type track17Service struct {
	config     *Config
	httpClient *http.Client
}

type track17RequestItem struct {
	Number  string          `json:"number"`
	Carrier json.RawMessage `json:"carrier,omitempty"`
}

type track17Response struct {
	Code int             `json:"code"`
	Data track17Data     `json:"data"`
	Msg  string          `json:"msg"`
	Raw  json.RawMessage `json:"-"`
}

type track17Data struct {
	Accepted []track17AcceptedItem `json:"accepted"`
	Rejected []track17RejectedItem `json:"rejected"`
}

type track17AcceptedItem struct {
	Number    string           `json:"number"`
	Carrier   json.RawMessage  `json:"carrier"`
	TrackInfo track17TrackInfo `json:"track_info"`
}

type track17RejectedItem struct {
	Number  string          `json:"number"`
	Carrier json.RawMessage `json:"carrier"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type track17TrackInfo struct {
	ShippingInfo track17ShippingInfo  `json:"shipping_info"`
	LatestStatus track17LatestStatus  `json:"latest_status"`
	LatestEvent  track17TrackingEvent `json:"latest_event"`
	Tracking     track17Tracking      `json:"tracking"`
}

type track17ShippingInfo struct {
	ShipperAddress   track17Address `json:"shipper_address"`
	RecipientAddress track17Address `json:"recipient_address"`
}

type track17Address struct {
	Country    string `json:"country"`
	State      string `json:"state"`
	City       string `json:"city"`
	Street     string `json:"street"`
	PostalCode string `json:"postal_code"`
}

type track17LatestStatus struct {
	Status    string `json:"status"`
	SubStatus string `json:"sub_status"`
}

type track17Tracking struct {
	Providers []track17Provider `json:"providers"`
}

type track17Provider struct {
	Provider track17ProviderInfo    `json:"provider"`
	Events   []track17TrackingEvent `json:"events"`
}

type track17ProviderInfo struct {
	Key  json.RawMessage `json:"key"`
	Name string          `json:"name"`
}

type track17TrackingEvent struct {
	TimeISO     string          `json:"time_iso"`
	TimeUTC     string          `json:"time_utc"`
	TimeRaw     json.RawMessage `json:"time_raw"`
	Status      string          `json:"status"`
	Stage       string          `json:"stage"`
	SubStatus   string          `json:"sub_status"`
	Location    string          `json:"location"`
	Description string          `json:"description"`
}

// NewTrackingService 创建物流追踪服务。
func NewTrackingService(config *Config) TrackingService {
	timeout := 15 * time.Second
	if config != nil && config.Timeout > 0 {
		timeout = config.Timeout
	}

	return &track17Service{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Track 查询单个物流。17TRACK V2.4 的查询接口实际支持批量，这里复用 BatchTrack 保持调用方简单。
func (s *track17Service) Track(ctx context.Context, trackingNumber, carrier string) (*TrackingInfo, error) {
	results, err := s.BatchTrack(ctx, []TrackingRequest{{TrackingNumber: trackingNumber, Carrier: carrier}})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 || results[0] == nil {
		return nil, fmt.Errorf("tracking number not found")
	}
	return results[0], nil
}

// BatchTrack 查询物流详情。17TRACK V2.4 /gettrackinfo 单次最多 40 个单号。
func (s *track17Service) BatchTrack(ctx context.Context, trackings []TrackingRequest) ([]*TrackingInfo, error) {
	if err := validateBatchTrackingRequestWithLimit(trackings, 40); err != nil {
		return nil, err
	}

	result, err := s.sendRequest(ctx, track17GetTrackInfoEndpoint, track17RequestItems(trackings))
	if err != nil {
		return nil, err
	}
	if err := resultError(result); err != nil {
		return nil, err
	}
	if len(result.Data.Accepted) == 0 {
		return nil, rejectedTrackingError(result.Data.Rejected)
	}

	results := make([]*TrackingInfo, 0, len(result.Data.Accepted))
	for _, item := range result.Data.Accepted {
		results = append(results, track17AcceptedToTrackingInfo(item))
	}
	return results, nil
}

// RegisterTrackings registers tracking numbers with 17TRACK so future tracking updates can be pushed by webhook.
func (s *track17Service) RegisterTrackings(ctx context.Context, trackings []TrackingRequest) error {
	if err := validateBatchTrackingRequestWithLimit(trackings, 40); err != nil {
		return err
	}

	result, err := s.sendRequest(ctx, track17RegisterEndpoint, track17RequestItems(trackings))
	if err != nil {
		return err
	}
	if err := resultError(result); err != nil {
		return err
	}
	if len(result.Data.Rejected) > 0 && len(result.Data.Accepted) == 0 {
		return rejectedTrackingError(result.Data.Rejected)
	}
	return nil
}

// DetectCarrier 17TRACK V2.4 查询链路依赖注册/查询时返回的 carrier，暂不单独暴露猜测结果。
func (s *track17Service) DetectCarrier(ctx context.Context, trackingNumber string) ([]string, error) {
	if err := validateTrackingNumber(trackingNumber); err != nil {
		return nil, err
	}
	return nil, errors.New("17TRACK V2.4 carrier detection is not exposed by this adapter")
}

func (s *track17Service) sendRequest(ctx context.Context, endpoint string, body interface{}) (*track17Response, error) {
	if s == nil || s.config == nil {
		return nil, errors.New("tracking service config is required")
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal 17TRACK request: %w", err)
	}

	baseURL := strings.TrimRight(strings.TrimSpace(s.config.BaseURL), "/")
	if baseURL == "" {
		return nil, errors.New("17TRACK base url is required")
	}
	requestURL, err := url.JoinPath(baseURL, strings.TrimLeft(endpoint, "/"))
	if err != nil {
		return nil, fmt.Errorf("failed to build 17TRACK request url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create 17TRACK request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("17token", strings.TrimSpace(s.config.APIKey))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send 17TRACK request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read 17TRACK response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("17TRACK returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result track17Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse 17TRACK response: %w", err)
	}
	result.Raw = respBody
	return &result, nil
}

func track17RequestItems(trackings []TrackingRequest) []track17RequestItem {
	items := make([]track17RequestItem, 0, len(trackings))
	for _, tracking := range trackings {
		item := track17RequestItem{Number: strings.TrimSpace(tracking.TrackingNumber)}
		if carrier := strings.TrimSpace(tracking.Carrier); carrier != "" {
			if _, err := strconv.Atoi(carrier); err == nil {
				item.Carrier = json.RawMessage(carrier)
			} else {
				item.Carrier = json.RawMessage(strconv.Quote(carrier))
			}
		}
		items = append(items, item)
	}
	return items
}

func resultError(result *track17Response) error {
	if result == nil {
		return errors.New("empty 17TRACK response")
	}
	if result.Code == 0 {
		return nil
	}
	message := strings.TrimSpace(result.Msg)
	if message == "" {
		message = "17TRACK API error"
	}
	return fmt.Errorf("%s: code %d", message, result.Code)
}

func rejectedTrackingError(rejected []track17RejectedItem) error {
	if len(rejected) == 0 {
		return fmt.Errorf("tracking number not found")
	}

	item := rejected[0]
	message := strings.TrimSpace(item.Error.Message)
	if message == "" {
		message = "tracking number rejected by 17TRACK"
	}
	if item.Error.Code != 0 {
		return fmt.Errorf("%s: code %d", message, item.Error.Code)
	}
	return errors.New(message)
}

func track17AcceptedToTrackingInfo(item track17AcceptedItem) *TrackingInfo {
	carrier := jsonValueString(item.Carrier)
	if carrier == "" {
		carrier = firstTrack17ProviderKey(item.TrackInfo.Tracking.Providers)
	}

	status := firstNonEmpty(
		item.TrackInfo.LatestStatus.SubStatus,
		item.TrackInfo.LatestStatus.Status,
		item.TrackInfo.LatestEvent.SubStatus,
		item.TrackInfo.LatestEvent.Stage,
		item.TrackInfo.LatestEvent.Status,
		item.TrackInfo.LatestEvent.Description,
	)

	info := &TrackingInfo{
		TrackingNumber: strings.TrimSpace(item.Number),
		Carrier:        carrier,
		Status:         status,
		StatusCode:     track17StatusCode(status),
		Events:         make([]TrackingEvent, 0),
		Origin:         track17AddressToLocation(item.TrackInfo.ShippingInfo.ShipperAddress),
		Destination:    track17AddressToLocation(item.TrackInfo.ShippingInfo.RecipientAddress),
	}

	for _, provider := range item.TrackInfo.Tracking.Providers {
		if info.Carrier == "" {
			info.Carrier = jsonValueString(provider.Provider.Key)
		}
		for _, event := range provider.Events {
			info.Events = append(info.Events, track17EventToTrackingEvent(event, status))
		}
	}

	if len(info.Events) == 0 && track17EventHasContent(item.TrackInfo.LatestEvent) {
		info.Events = append(info.Events, track17EventToTrackingEvent(item.TrackInfo.LatestEvent, status))
	}
	if len(info.Events) == 0 && status != "" {
		info.Events = append(info.Events, TrackingEvent{
			Time:        time.Now(),
			Status:      status,
			Description: status,
		})
	}

	info.UpdatedAt = latestEventTime(info.Events)
	if info.UpdatedAt.IsZero() {
		info.UpdatedAt = time.Now()
	}
	return info
}

func track17EventToTrackingEvent(event track17TrackingEvent, fallbackStatus string) TrackingEvent {
	eventTime, _ := parseTrack17EventTime(event)
	return TrackingEvent{
		Time:        eventTime,
		Status:      firstNonEmpty(event.SubStatus, event.Stage, event.Status, fallbackStatus),
		Description: strings.TrimSpace(event.Description),
		Location:    parseLocation(event.Location),
	}
}

func parseTrack17EventTime(event track17TrackingEvent) (time.Time, error) {
	if parsed, err := parseTrack17Time(firstNonEmpty(event.TimeISO, event.TimeUTC)); err == nil && !parsed.IsZero() {
		return parsed, nil
	}

	if len(event.TimeRaw) == 0 || string(event.TimeRaw) == "null" {
		return time.Time{}, nil
	}

	var rawString string
	if err := json.Unmarshal(event.TimeRaw, &rawString); err == nil {
		return parseTrack17Time(rawString)
	}

	var raw struct {
		Date     string `json:"date"`
		Time     string `json:"time"`
		Timezone string `json:"timezone"`
	}
	if err := json.Unmarshal(event.TimeRaw, &raw); err != nil {
		return time.Time{}, nil
	}
	return parseTrack17Time(strings.TrimSpace(strings.Join([]string{raw.Date, raw.Time, raw.Timezone}, " ")))
}

func parseTrack17Time(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	if unixTimestamp, err := strconv.ParseInt(value, 10, 64); err == nil {
		if unixTimestamp > 9999999999 {
			return time.UnixMilli(unixTimestamp).UTC(), nil
		}
		return time.Unix(unixTimestamp, 0).UTC(), nil
	}

	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05 -07:00",
		"2006-01-02 15:04 -07:00",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04 -0700",
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04 MST",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
	} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, nil
}

func track17AddressToLocation(address track17Address) *Location {
	parts := []string{address.City, address.State, address.Country, address.PostalCode}
	hasContent := false
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			hasContent = true
			break
		}
	}
	if !hasContent {
		return nil
	}
	return &Location{
		City:    strings.TrimSpace(address.City),
		State:   strings.TrimSpace(address.State),
		Country: strings.TrimSpace(address.Country),
		ZipCode: strings.TrimSpace(address.PostalCode),
	}
}

func parseLocation(locationStr string) *Location {
	locationStr = strings.TrimSpace(locationStr)
	if locationStr == "" {
		return nil
	}
	return &Location{City: locationStr}
}

func jsonValueString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	var stringValue string
	if err := json.Unmarshal(raw, &stringValue); err == nil {
		return strings.TrimSpace(stringValue)
	}

	var numberValue json.Number
	if err := json.Unmarshal(raw, &numberValue); err == nil {
		return strings.TrimSpace(numberValue.String())
	}

	return strings.Trim(strings.TrimSpace(string(raw)), `"`)
}

func firstTrack17ProviderKey(providers []track17Provider) string {
	for _, provider := range providers {
		if key := jsonValueString(provider.Provider.Key); key != "" {
			return key
		}
	}
	return ""
}

func track17EventHasContent(event track17TrackingEvent) bool {
	return firstNonEmpty(
		event.TimeISO,
		event.TimeUTC,
		string(event.TimeRaw),
		event.Status,
		event.Stage,
		event.SubStatus,
		event.Location,
		event.Description,
	) != ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" && trimmed != "null" {
			return trimmed
		}
	}
	return ""
}

func latestEventTime(events []TrackingEvent) time.Time {
	var latest time.Time
	for _, event := range events {
		if event.Time.After(latest) {
			latest = event.Time
		}
	}
	return latest
}

func track17StatusCode(status string) int {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "delivered":
		return 4
	case "exception", "expired", "deliveryfailure", "returned":
		return 3
	case "intransit", "transit", "departed", "arrival", "pickup", "pickedup":
		return 2
	case "inforeceived", "notfound":
		return 1
	default:
		return 0
	}
}
