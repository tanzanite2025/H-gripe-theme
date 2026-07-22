package shipping

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

type trackingWebhookEnvelope struct {
	TrackingNumber      string                         `json:"tracking_number"`
	Number              string                         `json:"number"`
	ProviderCarrierCode string                         `json:"provider_carrier_code"`
	Carrier             string                         `json:"carrier"`
	Status              string                         `json:"status"`
	StatusCode          int                            `json:"status_code"`
	Events              []trackingWebhookEventEnvelope `json:"events"`
	Data                *trackingWebhookEnvelope       `json:"data"`
	Tracking            *trackingWebhookEnvelope       `json:"tracking"`
	Shipment            *trackingWebhookEnvelope       `json:"shipment"`
}

type trackingWebhookEventEnvelope struct {
	Status      string `json:"status"`
	Location    string `json:"location"`
	Description string `json:"description"`
	EventTime   string `json:"event_time"`
	Time        string `json:"time"`
}

type track17WebhookEnvelope struct {
	Event string             `json:"event"`
	Data  track17WebhookData `json:"data"`
}

type track17WebhookData struct {
	Number    string           `json:"number"`
	Carrier   string           `json:"carrier"`
	TrackInfo track17TrackInfo `json:"track_info"`
}

type track17TrackInfo struct {
	LatestStatus track17LatestStatus `json:"latest_status"`
	LatestEvent  track17Event        `json:"latest_event"`
	Tracking     track17Tracking     `json:"tracking"`
}

type track17LatestStatus struct {
	Status    string `json:"status"`
	SubStatus string `json:"sub_status"`
}

type track17Tracking struct {
	Providers []track17Provider `json:"providers"`
}

type track17Provider struct {
	Provider track17ProviderInfo `json:"provider"`
	Events   []track17Event      `json:"events"`
}

type track17ProviderInfo struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type track17Event struct {
	TimeISO     string `json:"time_iso"`
	TimeUTC     string `json:"time_utc"`
	TimeRaw     string `json:"time_raw"`
	Status      string `json:"status"`
	Stage       string `json:"stage"`
	SubStatus   string `json:"sub_status"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

func (h *Handler) HandleTrackingWebhook(c *gin.Context) {
	startedAt := time.Now()
	providerCode := strings.TrimSpace(c.Param("provider"))
	webhookState := service.TrackingWebhookRunState{
		LastReceivedAt:   &startedAt,
		LastProviderCode: providerCode,
	}
	finish := func(status int, accepted bool, errMessage string) {
		finishedAt := time.Now()
		webhookState.LastFinishedAt = &finishedAt
		webhookState.LastDurationMs = finishedAt.Sub(startedAt).Milliseconds()
		webhookState.LastHTTPStatus = status
		webhookState.LastAccepted = accepted
		webhookState.LastError = errMessage
		h.shippingService.RecordTrackingWebhookRun(webhookState)
	}

	if providerCode == "" {
		finish(400, false, "tracking provider is required")
		apierror.RespondBadRequest(c, "tracking provider is required")
		return
	}

	provider, err := h.shippingService.GetTrackingProviderConfigByCode(providerCode)
	if err != nil {
		finish(404, false, "tracking provider not found")
		apierror.RespondNotFound(c, "Tracking provider")
		return
	}
	webhookState.LastProviderID = provider.ID
	webhookState.LastProviderCode = provider.ProviderCode
	if !provider.Enabled || !provider.WebhookEnabled {
		finish(400, false, "tracking provider webhook is disabled")
		apierror.RespondBadRequest(c, "tracking provider webhook is disabled")
		return
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		finish(400, false, "failed to read request body")
		apierror.RespondBadRequest(c, "failed to read request body")
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(payload))

	webhookState.LastSignatureChecked = true
	if !verifyTrackingWebhook(c, strings.TrimSpace(provider.WebhookSecret), payload) {
		finish(401, false, "invalid webhook signature")
		apierror.RespondUnauthorized(c)
		return
	}
	webhookState.LastSignatureValid = true

	input, err := parseTrackingWebhookInput(payload)
	if err != nil {
		finish(400, false, err.Error())
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	input.ProviderID = provider.ID
	webhookState.LastTrackingNumber = input.TrackingNumber
	webhookState.LastCarrierCode = input.ProviderCarrierCode
	webhookState.LastEventCount = len(input.Events)

	result, err := h.shippingService.ApplyTrackingWebhook(input)
	if err != nil {
		finish(400, false, err.Error())
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if result != nil {
		webhookState.LastEventCount = len(result.Events)
		if result.Shipment != nil {
			webhookState.LastOrderID = result.Shipment.OrderID
			webhookState.LastTrackingNumber = result.Shipment.TrackingNumber
			webhookState.LastCarrierCode = result.Shipment.ProviderCarrierCode
		}
	}

	finish(200, true, "")
	response.SuccessWithMessage(c, "tracking webhook processed", result)
}

func verifyTrackingWebhook(c *gin.Context, secret string, payload []byte) bool {
	if secret == "" {
		return false
	}

	plainSecret := strings.TrimSpace(c.GetHeader("X-Tanzanite-Webhook-Secret"))
	if plainSecret != "" && hmac.Equal([]byte(plainSecret), []byte(secret)) {
		return true
	}

	track17Signature := normalizeWebhookSignature(c.GetHeader("sign"))
	if track17Signature != "" {
		hash := sha256.New()
		hash.Write(payload)
		hash.Write([]byte("/"))
		hash.Write([]byte(secret))
		expected := hex.EncodeToString(hash.Sum(nil))
		if hmac.Equal([]byte(track17Signature), []byte(expected)) {
			return true
		}
	}

	signature := strings.TrimSpace(c.GetHeader("X-Tanzanite-Signature"))
	if signature == "" {
		signature = strings.TrimSpace(c.GetHeader("X-17Track-Signature"))
	}
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	signature = normalizeWebhookSignature(signature)
	return hmac.Equal([]byte(signature), []byte(expected))
}

func parseTrackingWebhookInput(payload []byte) (service.TrackingWebhookInput, error) {
	if input, ok, err := parseTrack17WebhookInput(payload); err != nil {
		return service.TrackingWebhookInput{}, err
	} else if ok {
		return input, nil
	}

	var envelope trackingWebhookEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return service.TrackingWebhookInput{}, err
	}

	normalized := unwrapTrackingWebhookEnvelope(envelope)
	trackingNumber := firstNonEmpty(normalized.TrackingNumber, normalized.Number)
	providerCarrierCode := firstNonEmpty(normalized.ProviderCarrierCode, normalized.Carrier)

	input := service.TrackingWebhookInput{
		TrackingNumber:      trackingNumber,
		ProviderCarrierCode: providerCarrierCode,
		Status:              strings.TrimSpace(normalized.Status),
		StatusCode:          normalized.StatusCode,
		Events:              make([]service.TrackingWebhookEventInput, 0, len(normalized.Events)),
	}

	for _, event := range normalized.Events {
		eventTime, _ := parseWebhookTime(firstNonEmpty(event.EventTime, event.Time))
		input.Events = append(input.Events, service.TrackingWebhookEventInput{
			Status:      strings.TrimSpace(event.Status),
			Location:    strings.TrimSpace(event.Location),
			Description: strings.TrimSpace(event.Description),
			EventTime:   eventTime,
		})
	}

	return input, nil
}

func parseTrack17WebhookInput(payload []byte) (service.TrackingWebhookInput, bool, error) {
	var envelope track17WebhookEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return service.TrackingWebhookInput{}, false, err
	}

	if strings.ToUpper(strings.TrimSpace(envelope.Event)) != "TRACKING_UPDATED" {
		return service.TrackingWebhookInput{}, false, nil
	}

	trackingNumber := strings.TrimSpace(envelope.Data.Number)
	if trackingNumber == "" {
		return service.TrackingWebhookInput{}, false, nil
	}

	providerCarrierCode := strings.TrimSpace(envelope.Data.Carrier)
	if providerCarrierCode == "" {
		providerCarrierCode = firstTrack17ProviderKey(envelope.Data.TrackInfo.Tracking.Providers)
	}

	latestStatus := firstNonEmpty(
		envelope.Data.TrackInfo.LatestStatus.SubStatus,
		envelope.Data.TrackInfo.LatestStatus.Status,
		envelope.Data.TrackInfo.LatestEvent.SubStatus,
		envelope.Data.TrackInfo.LatestEvent.Stage,
		envelope.Data.TrackInfo.LatestEvent.Status,
		envelope.Data.TrackInfo.LatestEvent.Description,
	)

	input := service.TrackingWebhookInput{
		TrackingNumber:      trackingNumber,
		ProviderCarrierCode: providerCarrierCode,
		Status:              latestStatus,
		Events:              make([]service.TrackingWebhookEventInput, 0),
	}

	for _, provider := range envelope.Data.TrackInfo.Tracking.Providers {
		if input.ProviderCarrierCode == "" {
			input.ProviderCarrierCode = strings.TrimSpace(provider.Provider.Key)
		}
		for _, event := range provider.Events {
			eventTime, _ := parseWebhookTime(firstNonEmpty(event.TimeISO, event.TimeUTC, event.TimeRaw))
			input.Events = append(input.Events, service.TrackingWebhookEventInput{
				Status:      firstNonEmpty(event.SubStatus, event.Stage, event.Status, latestStatus),
				Location:    strings.TrimSpace(event.Location),
				Description: strings.TrimSpace(event.Description),
				EventTime:   eventTime,
			})
		}
	}

	if len(input.Events) == 0 && (track17EventHasContent(envelope.Data.TrackInfo.LatestEvent) || latestStatus != "") {
		latestEvent := envelope.Data.TrackInfo.LatestEvent
		eventTime, _ := parseWebhookTime(firstNonEmpty(latestEvent.TimeISO, latestEvent.TimeUTC, latestEvent.TimeRaw))
		description := firstNonEmpty(latestEvent.Description, latestStatus)
		input.Events = append(input.Events, service.TrackingWebhookEventInput{
			Status:      latestStatus,
			Location:    strings.TrimSpace(latestEvent.Location),
			Description: description,
			EventTime:   eventTime,
		})
	}

	return input, true, nil
}

func unwrapTrackingWebhookEnvelope(envelope trackingWebhookEnvelope) trackingWebhookEnvelope {
	switch {
	case envelope.Data != nil:
		return unwrapTrackingWebhookEnvelope(*envelope.Data)
	case envelope.Tracking != nil:
		return unwrapTrackingWebhookEnvelope(*envelope.Tracking)
	case envelope.Shipment != nil:
		return unwrapTrackingWebhookEnvelope(*envelope.Shipment)
	default:
		return envelope
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizeWebhookSignature(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "sha256=")
	return strings.TrimSpace(value)
}

func firstTrack17ProviderKey(providers []track17Provider) string {
	for _, provider := range providers {
		if key := strings.TrimSpace(provider.Provider.Key); key != "" {
			return key
		}
	}
	return ""
}

func track17EventHasContent(event track17Event) bool {
	return firstNonEmpty(
		event.TimeISO,
		event.TimeUTC,
		event.TimeRaw,
		event.Status,
		event.Stage,
		event.SubStatus,
		event.Location,
		event.Description,
	) != ""
}

func parseWebhookTime(value string) (time.Time, error) {
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
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04 MST",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04 -0700",
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
