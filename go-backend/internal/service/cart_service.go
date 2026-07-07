package service

import (
	"errors"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/repository"
)

type CartService struct {
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewCartService(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *CartService) GetOrCreateCart(userID *uint, sessionID string) (*product.Cart, error) {
	var cart *product.Cart
	var err error

	if userID != nil {
		cart, err = s.cartRepo.FindByUserID(*userID)
	} else {
		cart, err = s.cartRepo.FindBySessionID(sessionID)
	}

	if repository.IsRecordNotFound(err) {
		cart = &product.Cart{
			UserID:    userID,
			SessionID: sessionID,
		}
		if err := s.cartRepo.Create(cart); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return cart, nil
}

func (s *CartService) AddToCart(cartID, productID uint, variantID *uint, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	_, variant, err := s.productRepo.FindPurchasableVariant(productID, variantID)
	if err != nil || variant == nil {
		return errors.New("product not found")
	}

	price, availableStock, resolvedVariantID := purchasablePriceStock(variant)
	if availableStock < quantity {
		return errors.New("insufficient stock")
	}

	existingItem, err := s.cartRepo.FindItem(cartID, productID, resolvedVariantID)
	if err == nil {
		if existingItem.Quantity+quantity > availableStock {
			return errors.New("insufficient stock")
		}
		existingItem.Quantity += quantity
		existingItem.Price = price
		return s.cartRepo.UpdateItem(existingItem)
	}
	if !repository.IsRecordNotFound(err) {
		return err
	}

	return s.cartRepo.AddItem(&product.CartItem{
		CartID:    cartID,
		ProductID: productID,
		VariantID: resolvedVariantID,
		Quantity:  quantity,
		Price:     price,
	})
}

func (s *CartService) UpdateCartItem(cartID, productID uint, variantID *uint, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	item, err := s.cartRepo.FindItem(cartID, productID, variantID)
	if err != nil {
		return errors.New("item not found in cart")
	}

	_, variant, err := s.productRepo.FindPurchasableVariant(productID, item.VariantID)
	if err != nil || variant == nil {
		return errors.New("product not found")
	}

	price, availableStock, _ := purchasablePriceStock(variant)
	if availableStock < quantity {
		return errors.New("insufficient stock")
	}

	item.Quantity = quantity
	item.Price = price
	return s.cartRepo.UpdateItem(item)
}

func (s *CartService) RemoveFromCart(cartID, productID uint, variantID *uint) error {
	item, err := s.cartRepo.FindItem(cartID, productID, variantID)
	if err != nil {
		return nil
	}
	return s.cartRepo.RemoveItem(item.ID)
}

type SyncCartItemReq struct {
	ProductID uint  `json:"product_id"`
	VariantID *uint `json:"variant_id"`
	Quantity  int   `json:"quantity"`
}

func (s *CartService) SyncCart(cartID uint, items []SyncCartItemReq) error {
	if len(items) == 0 {
		return nil
	}

	var cartItems []product.CartItem
	for _, req := range items {
		if req.Quantity <= 0 {
			continue
		}

		_, variant, err := s.productRepo.FindPurchasableVariant(req.ProductID, req.VariantID)
		if err != nil || variant == nil {
			continue
		}

		price, availableStock, resolvedVariantID := purchasablePriceStock(variant)
		if availableStock < req.Quantity {
			continue
		}

		cartItems = append(cartItems, product.CartItem{
			CartID:    cartID,
			ProductID: req.ProductID,
			VariantID: resolvedVariantID,
			Quantity:  req.Quantity,
			Price:     price,
		})
	}

	return s.cartRepo.BulkUpsertItems(cartItems)
}

func (s *CartService) GetCartSummary(userID *uint, sessionID string) (*product.CartSummary, error) {
	cart, err := s.GetOrCreateCart(userID, sessionID)
	if err != nil {
		return nil, err
	}

	return s.cartRepo.GetSummary(cart.ID)
}

func (s *CartService) ClearCart(cartID uint) error {
	return s.cartRepo.ClearCart(cartID)
}

func purchasablePriceStock(variant *product.ProductVariant) (float64, int, *uint) {
	variantID := variant.ID
	return variant.EffectivePrice(), variant.Stock, &variantID
}
