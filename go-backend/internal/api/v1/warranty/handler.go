package warranty

import (
	"commerce-platform/internal/pkg/antibot"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/service"
)

type Handler struct {
	warrantySvc    *service.WarrantyService
	shipmentSvc    *service.ShipmentRecordService
	storageService storage.StorageService
	antiBot        *antibot.Service
	mediaResolver  service.PublicMediaURLResolver
}

func NewHandler(warrantySvc *service.WarrantyService, storageService storage.StorageService, antiBotServices ...*antibot.Service) *Handler {
	var antiBot *antibot.Service
	if len(antiBotServices) > 0 {
		antiBot = antiBotServices[0]
	}
	return &Handler{
		warrantySvc:    warrantySvc,
		storageService: storageService,
		antiBot:        antiBot,
	}
}

func (h *Handler) ConfigureMediaService(resolver service.PublicMediaURLResolver) {
	if h == nil {
		return
	}
	h.mediaResolver = resolver
}

func (h *Handler) ConfigureShipmentRecordService(shipmentSvc *service.ShipmentRecordService) {
	if h == nil {
		return
	}
	h.shipmentSvc = shipmentSvc
}
