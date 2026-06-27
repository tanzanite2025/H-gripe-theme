package registration

import (
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"
)

type Handler struct {
	registrationRepo *repository.RegistrationRepository
	registrationSvc  *service.RegistrationService
	orderRepo        *repository.OrderRepository
	storageService   storage.StorageService
}

func NewHandler(registrationRepo *repository.RegistrationRepository, registrationSvc *service.RegistrationService, orderRepo *repository.OrderRepository, storageService storage.StorageService) *Handler {
	return &Handler{
		registrationRepo: registrationRepo,
		registrationSvc:  registrationSvc,
		orderRepo:        orderRepo,
		storageService:   storageService,
	}
}
