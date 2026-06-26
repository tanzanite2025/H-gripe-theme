package registration

import (
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/repository"
)

// Handler 产品注册处理器
type Handler struct {
	registrationRepo *repository.RegistrationRepository
	orderRepo        *repository.OrderRepository
	storageService   storage.StorageService
}

// NewHandler 创建产品注册处理器
func NewHandler(registrationRepo *repository.RegistrationRepository, orderRepo *repository.OrderRepository, storageService storage.StorageService) *Handler {
	return &Handler{
		registrationRepo: registrationRepo,
		orderRepo:        orderRepo,
		storageService:   storageService,
	}
}
