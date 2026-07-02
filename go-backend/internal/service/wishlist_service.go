package service

import (
	"errors"
	"tanzanite/internal/domain/wishlist"
	"tanzanite/internal/repository"
)

var (
	ErrWishlistProductNotFound = errors.New("product not found")
	ErrWishlistItemNotFound    = errors.New("wishlist item not found")
	ErrWishlistForbidden       = errors.New("wishlist item forbidden")
)

type WishlistService struct {
	wishlistRepo *repository.WishlistRepository
	productRepo  *repository.ProductRepository
}

func NewWishlistService(wishlistRepo *repository.WishlistRepository, productRepo *repository.ProductRepository) *WishlistService {
	return &WishlistService{
		wishlistRepo: wishlistRepo,
		productRepo:  productRepo,
	}
}

func (s *WishlistService) List(userID uint) ([]wishlist.Item, error) {
	return s.wishlistRepo.ListByUserID(userID)
}

func (s *WishlistService) Add(userID, productID uint) (*wishlist.Item, error) {
	if _, err := s.productRepo.FindByID(productID); err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrWishlistProductNotFound
		}
		return nil, err
	}

	existing, err := s.wishlistRepo.FindByUserAndProduct(userID, productID)
	if err == nil {
		return existing, nil
	}
	if !repository.IsRecordNotFound(err) {
		return nil, err
	}

	item := &wishlist.Item{
		UserID:    userID,
		ProductID: productID,
	}
	if err := s.wishlistRepo.Create(item); err != nil {
		return nil, err
	}

	return s.wishlistRepo.FindByID(item.ID)
}

func (s *WishlistService) Remove(userID, itemID uint) error {
	item, err := s.wishlistRepo.FindByID(itemID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return ErrWishlistItemNotFound
		}
		return err
	}
	if item.UserID != userID {
		return ErrWishlistForbidden
	}
	return s.wishlistRepo.Delete(itemID)
}
