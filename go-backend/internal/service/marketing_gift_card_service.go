package service

import (
	"errors"
	"tanzanite/internal/domain/coupon"
)

type GiftCardDetail struct {
	GiftCard     *coupon.GiftCard
	Transactions []coupon.GiftCardTransaction
}

var ErrInvalidGiftCardStatusTransition = errors.New("invalid gift card status transition")

func (s *MarketingService) ListGiftCardsAdmin(page, pageSize int, status string) ([]coupon.GiftCard, int64, error) {
	return s.couponRepo.FindAllGiftCards(page, pageSize, status)
}

func (s *MarketingService) GetGiftCard(id uint) (*GiftCardDetail, error) {
	card, err := s.couponRepo.FindGiftCardByID(id)
	if err != nil {
		return nil, normalizeMarketingError(err)
	}

	transactions, err := s.couponRepo.FindGiftCardTransactionsByCardID(id)
	if err != nil {
		return nil, err
	}

	return &GiftCardDetail{
		GiftCard:     card,
		Transactions: transactions,
	}, nil
}

func (s *MarketingService) UpdateGiftCardStatus(id uint, status string) (*coupon.GiftCard, error) {
	card, err := s.couponRepo.FindGiftCardByID(id)
	if err != nil {
		return nil, normalizeMarketingError(err)
	}

	if !canTransitionGiftCardStatus(card, status) {
		return nil, ErrInvalidGiftCardStatusTransition
	}

	card.Status = status
	if err := s.couponRepo.UpdateGiftCard(card); err != nil {
		return nil, err
	}

	return card, nil
}

func canTransitionGiftCardStatus(card *coupon.GiftCard, next string) bool {
	if card == nil || card.Status == next {
		return true
	}

	switch card.Status {
	case "active":
		if next == "used" {
			return card.BalanceCents <= 0
		}
		return next == "expired" || next == "cancelled"
	case "used", "expired", "cancelled":
		return false
	default:
		return false
	}
}
