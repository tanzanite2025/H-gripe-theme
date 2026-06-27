package registration

import (
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/service"
)

type Handler struct {
	registrationSvc *service.RegistrationService
	storageService  storage.StorageService
}

func NewHandler(registrationSvc *service.RegistrationService, storageService storage.StorageService) *Handler {
	return &Handler{
		registrationSvc: registrationSvc,
		storageService:  storageService,
	}
}
