package registration

import (
	"tanzanite/internal/pkg/antibot"
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/service"
)

type Handler struct {
	registrationSvc *service.RegistrationService
	storageService  storage.StorageService
	antiBot         *antibot.Service
}

func NewHandler(registrationSvc *service.RegistrationService, storageService storage.StorageService, antiBotServices ...*antibot.Service) *Handler {
	var antiBot *antibot.Service
	if len(antiBotServices) > 0 {
		antiBot = antiBotServices[0]
	}
	return &Handler{
		registrationSvc: registrationSvc,
		storageService:  storageService,
		antiBot:         antiBot,
	}
}
