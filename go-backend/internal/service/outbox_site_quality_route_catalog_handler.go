package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"commerce-platform/internal/domain/outbox"
)

type siteQualityRouteCatalogChangedPayload struct {
	RouteEntryID uint      `json:"route_entry_id"`
	Marker       string    `json:"marker,omitempty"`
	SeenAt       time.Time `json:"seen_at,omitempty"`
}

type SiteQualityRouteCatalogOutboxHandler struct {
	engine *SiteQualityEngineService
}

func NewSiteQualityRouteCatalogOutboxHandler(engine *SiteQualityEngineService) *SiteQualityRouteCatalogOutboxHandler {
	return &SiteQualityRouteCatalogOutboxHandler{engine: engine}
}

func (h *SiteQualityRouteCatalogOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || h.engine == nil {
		return errors.New("SiteQuality route catalog handler is unavailable")
	}
	var payload siteQualityRouteCatalogChangedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}
	if payload.RouteEntryID == 0 {
		return errors.New("SiteQuality route catalog event route entry ID is required")
	}
	if payload.SeenAt.IsZero() {
		return h.engine.SyncTargetFromRouteEntry(ctx, payload.RouteEntryID, payload.Marker)
	}
	return h.engine.SyncTargetFromRouteEntry(ctx, payload.RouteEntryID, payload.Marker, payload.SeenAt)
}
