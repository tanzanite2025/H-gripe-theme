package service

import (
	"commerce-platform/internal/repository"
	"strings"
)

type WarrantyService struct {
	txManager       *repository.EmailChallengeTxManager
	warrantyRepo    *repository.WarrantyRepository
	orderRepo       *repository.OrderRepository
	challengeSecret string
	baseURL         string
	shipmentRepo    *repository.ShipmentRecordRepository
}

func NewWarrantyService(
	txManager *repository.EmailChallengeTxManager,
	warrantyRepo *repository.WarrantyRepository,
	orderRepo *repository.OrderRepository,
	shipmentRepos ...*repository.ShipmentRecordRepository,
) *WarrantyService {
	service := &WarrantyService{
		txManager:    txManager,
		warrantyRepo: warrantyRepo,
		orderRepo:    orderRepo,
	}
	if len(shipmentRepos) > 0 {
		service.shipmentRepo = shipmentRepos[0]
	} else if orderRepo != nil {
		service.shipmentRepo = orderRepo.ShipmentRecordRepository()
	}
	return service
}

func (s *WarrantyService) ConfigureShipmentRecordRepository(repo *repository.ShipmentRecordRepository) {
	if s != nil {
		s.shipmentRepo = repo
	}
}

func (s *WarrantyService) ConfigureEmailChallenges(secret string) {
	s.challengeSecret = secret
}

func (s *WarrantyService) ConfigureEmailBaseURL(baseURL string) {
	s.baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
}
