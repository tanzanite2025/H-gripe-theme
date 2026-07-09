package service

import (
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/repository"
	"time"
)

type GiftCardCreateInput struct {
	Code           string
	InitialValue   float64
	Currency       string
	RecipientEmail string
	RecipientName  string
	SenderName     string
	Message        string
	CoverImage     string
	ExpiresAt      *time.Time
}

type GiftCardDetail struct {
	GiftCard     *coupon.GiftCard
	Transactions []coupon.GiftCardTransaction
}

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

func (s *MarketingService) CreateGiftCardAdmin(input GiftCardCreateInput) (*coupon.GiftCard, error) {
	if input.Currency == "" {
		input.Currency = "USD"
	}

	var card coupon.GiftCard
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if err := ensureGiftCardCodeAvailable(repos.Coupon, input.Code, 0); err != nil {
			return err
		}

		card = coupon.GiftCard{
			Code:           input.Code,
			InitialValue:   input.InitialValue,
			Balance:        input.InitialValue,
			Currency:       input.Currency,
			Status:         "active",
			RecipientEmail: input.RecipientEmail,
			RecipientName:  input.RecipientName,
			SenderName:     input.SenderName,
			Message:        input.Message,
			CoverImage:     input.CoverImage,
			ExpiresAt:      input.ExpiresAt,
		}
		if err := repos.Coupon.CreateGiftCard(&card); err != nil {
			return err
		}

		transaction := &coupon.GiftCardTransaction{
			GiftCardID: card.ID,
			Type:       "issue",
			Amount:     input.InitialValue,
			Balance:    input.InitialValue,
			Note:       "Admin issued gift card",
		}
		return repos.Coupon.CreateGiftCardTransaction(transaction)
	})
	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (s *MarketingService) UpdateGiftCardStatus(id uint, status string) (*coupon.GiftCard, error) {
	card, err := s.couponRepo.FindGiftCardByID(id)
	if err != nil {
		return nil, normalizeMarketingError(err)
	}

	card.Status = status
	if err := s.couponRepo.UpdateGiftCard(card); err != nil {
		return nil, err
	}

	return card, nil
}

func ensureGiftCardCodeAvailable(repo *repository.CouponRepository, code string, excludeID uint) error {
	existing, err := repo.FindGiftCardByCode(code)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil
		}
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return ErrGiftCardCodeExists
	}
	return nil
}
