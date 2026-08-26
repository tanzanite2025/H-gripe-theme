package service

import (
	"commerce-platform/internal/repository"
	"strings"
)

type WarrantyService struct {
	warrantyRepo    *repository.WarrantyRepository
	orderRepo       *repository.OrderRepository
	challengeRepo   *repository.EmailChallengeRepository
	challengeSecret string
	emailSender     EmailChallengeSender
	baseURL         string
}

func NewWarrantyService(
	warrantyRepo *repository.WarrantyRepository,
	orderRepo *repository.OrderRepository,
) *WarrantyService {
	return &WarrantyService{
		warrantyRepo: warrantyRepo,
		orderRepo:    orderRepo,
	}
}

func (s *WarrantyService) ConfigureEmailChallenges(
	challengeRepo *repository.EmailChallengeRepository,
	secret string,
	senders ...EmailChallengeSender,
) {
	s.challengeRepo = challengeRepo
	s.challengeSecret = secret
	if len(senders) > 0 {
		s.emailSender = senders[0]
	}
}

func (s *WarrantyService) ConfigureEmailBaseURL(baseURL string) {
	s.baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
}
