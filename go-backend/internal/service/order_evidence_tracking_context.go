package service

import (
	"encoding/json"
	"strings"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/shipping"
)

// OrderEvidenceTrackingContext is read-only delivery context for the evidence
// workbench. It deliberately projects shipping data instead of exposing the
// provider aggregate, which may contain credentials and endpoint settings.
type OrderEvidenceTrackingContext = orderevidence.OrderEvidenceTrackingContext

type OrderEvidenceTrackingShipment = orderevidence.OrderEvidenceTrackingShipment

type OrderEvidenceDeliveryEvent = orderevidence.OrderEvidenceDeliveryEvent

type OrderEvidenceManualPOD = orderevidence.OrderEvidenceManualPOD

func buildOrderEvidenceTrackingContextFromSources(
	shipments []shipping.TrackingShipment,
	events []shipping.TrackingEvent,
	items []orderevidence.OrderEvidenceItem,
) *OrderEvidenceTrackingContext {
	trackingNumbers := make(map[string]struct{}, len(shipments))
	projectedShipments := make([]OrderEvidenceTrackingShipment, 0, len(shipments))
	for index := range shipments {
		trackingNumber := strings.TrimSpace(shipments[index].TrackingNumber)
		if trackingNumber == "" {
			continue
		}
		trackingNumbers[strings.ToLower(trackingNumber)] = struct{}{}
		projectedShipments = append(projectedShipments, *projectOrderEvidenceTrackingShipment(&shipments[index]))
	}
	events = filterTrackingEventsForShipments(events, trackingNumbers)
	result := &OrderEvidenceTrackingContext{
		Shipments: projectedShipments,
		ManualPOD: manualPODContext(items, shipments),
	}
	for _, shipment := range shipments {
		shipmentEvents := filterTrackingEventsForShipment(events, strings.TrimSpace(shipment.TrackingNumber))
		if latest := projectLatestDeliveryEvent(shipmentEvents); latest != nil {
			result.LatestDeliveryEvents = append(result.LatestDeliveryEvents, *latest)
		}
	}
	result.ProviderPODURLs = trackingProviderPODURLs(events)
	return result
}

func projectOrderEvidenceTrackingShipment(
	shipment *shipping.TrackingShipment,
) *OrderEvidenceTrackingShipment {
	if shipment == nil {
		return nil
	}

	result := &OrderEvidenceTrackingShipment{
		ID:                  shipment.ID,
		TrackingNumber:      strings.TrimSpace(shipment.TrackingNumber),
		ProviderCarrierCode: strings.TrimSpace(shipment.ProviderCarrierCode),
		RegistrationStatus:  strings.TrimSpace(shipment.RegistrationStatus),
		SyncStatus:          strings.TrimSpace(shipment.SyncStatus),
		EventCount:          shipment.EventCount,
		LastEventAt:         shipment.LastEventAt,
		LastSyncedAt:        shipment.LastSyncedAt,
		Enabled:             shipment.Enabled,
	}
	if shipment.Provider != nil {
		result.ProviderCode = strings.TrimSpace(shipment.Provider.ProviderCode)
		result.ProviderName = strings.TrimSpace(shipment.Provider.ProviderName)
	}
	if shipment.Carrier != nil {
		result.CarrierName = strings.TrimSpace(shipment.Carrier.Name)
	}
	if shipment.CarrierService != nil {
		result.CarrierServiceName = strings.TrimSpace(shipment.CarrierService.ServiceName)
	}
	return result
}

func projectLatestDeliveryEvent(events []shipping.TrackingEvent) *OrderEvidenceDeliveryEvent {
	index := latestDeliveryEventIndex(events)
	if index < 0 {
		return nil
	}
	projected := projectOrderEvidenceDeliveryEvent(events[index])
	return &projected
}

func projectOrderEvidenceDeliveryEvents(events []shipping.TrackingEvent) []OrderEvidenceDeliveryEvent {
	result := make([]OrderEvidenceDeliveryEvent, 0, len(events))
	for _, event := range events {
		result = append(result, projectOrderEvidenceDeliveryEvent(event))
	}
	return result
}

func projectOrderEvidenceDeliveryEvent(event shipping.TrackingEvent) OrderEvidenceDeliveryEvent {
	return OrderEvidenceDeliveryEvent{
		ID:                     event.ID,
		TrackingNumber:         strings.TrimSpace(event.TrackingNumber),
		Status:                 strings.TrimSpace(event.Status),
		Location:               strings.TrimSpace(event.Location),
		Description:            strings.TrimSpace(event.Description),
		RecipientSignatureName: strings.TrimSpace(event.RecipientSignatureName),
		ProofOfDeliveryURL:     strings.TrimSpace(event.ProofOfDeliveryURL),
		EventTime:              event.EventTime,
	}
}

func latestDeliveryEventIndex(events []shipping.TrackingEvent) int {
	latestIndex := -1
	for index := range events {
		if !trackingStatusTextIndicatesDelivery(events[index].Status) {
			continue
		}
		if latestIndex < 0 ||
			events[index].EventTime.After(events[latestIndex].EventTime) ||
			(events[index].EventTime.Equal(events[latestIndex].EventTime) && events[index].ID > events[latestIndex].ID) {
			latestIndex = index
		}
	}
	return latestIndex
}

func latestProviderPODURL(events []shipping.TrackingEvent) string {
	latestIndex := -1
	for index := range events {
		if strings.TrimSpace(events[index].ProofOfDeliveryURL) == "" {
			continue
		}
		if latestIndex < 0 ||
			events[index].EventTime.After(events[latestIndex].EventTime) ||
			(events[index].EventTime.Equal(events[latestIndex].EventTime) && events[index].ID > events[latestIndex].ID) {
			latestIndex = index
		}
	}
	if latestIndex < 0 {
		return ""
	}
	return strings.TrimSpace(events[latestIndex].ProofOfDeliveryURL)
}

func manualPODContext(
	items []orderevidence.OrderEvidenceItem,
	shipments []shipping.TrackingShipment,
) *OrderEvidenceManualPOD {
	for _, item := range items {
		if item.ItemType != orderevidence.EvidenceItemTypeSignedPOD {
			continue
		}
		manualTrackingNumber := manualPODTrackingNumber(item.DataJSON)
		associationStatus := "not_recorded"
		if len(shipments) == 0 {
			associationStatus = "shipment_unavailable"
		} else if manualTrackingNumber != "" && trackingNumberMatchesShipment(manualTrackingNumber, shipments) {
			associationStatus = "matched"
		} else if manualTrackingNumber != "" {
			associationStatus = "mismatch"
		}
		return &OrderEvidenceManualPOD{
			ItemID:                    item.ID,
			Status:                    strings.TrimSpace(item.Status),
			TrackingNumber:            manualTrackingNumber,
			TrackingAssociationStatus: associationStatus,
			CapturedAt:                item.CapturedAt,
			CapturedBy:                item.CapturedBy,
			AttachmentCount:           len(item.Attachments),
		}
	}
	return nil
}

func filterTrackingEventsForShipment(
	events []shipping.TrackingEvent,
	trackingNumber string,
) []shipping.TrackingEvent {
	trackingNumber = strings.TrimSpace(trackingNumber)
	if trackingNumber == "" {
		return []shipping.TrackingEvent{}
	}

	filtered := make([]shipping.TrackingEvent, 0, len(events))
	for _, event := range events {
		if strings.EqualFold(strings.TrimSpace(event.TrackingNumber), trackingNumber) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func filterTrackingEventsForShipments(events []shipping.TrackingEvent, trackingNumbers map[string]struct{}) []shipping.TrackingEvent {
	filtered := make([]shipping.TrackingEvent, 0, len(events))
	for _, event := range events {
		if _, ok := trackingNumbers[strings.ToLower(strings.TrimSpace(event.TrackingNumber))]; ok {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func trackingNumberMatchesShipment(trackingNumber string, shipments []shipping.TrackingShipment) bool {
	for _, shipment := range shipments {
		if strings.EqualFold(strings.TrimSpace(trackingNumber), strings.TrimSpace(shipment.TrackingNumber)) {
			return true
		}
	}
	return false
}

func trackingProviderPODURLs(events []shipping.TrackingEvent) []string {
	seen := make(map[string]struct{})
	urls := make([]string, 0)
	for _, event := range events {
		url := strings.TrimSpace(event.ProofOfDeliveryURL)
		if url == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		urls = append(urls, url)
	}
	return urls
}

func manualPODTrackingNumber(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	var payload struct {
		TrackingNumber string `json:"tracking_number"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.TrackingNumber)
}
