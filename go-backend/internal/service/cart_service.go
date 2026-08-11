package service

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"errors"
	"strings"
)

var ErrCartNotFound = errors.New("cart not found")

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

func (s *CartService) FindCart(userID *uint, sessionID string) (*product.Cart, error) {
	sessionID = strings.TrimSpace(sessionID)
	if userID == nil && sessionID == "" {
		return nil, ErrCartNotFound
	}

	var cart *product.Cart
	var err error
	if userID != nil {
		cart, err = s.cartRepo.FindByUserID(*userID)
	} else {
		cart, err = s.cartRepo.FindBySessionID(sessionID)
	}

	if repository.IsRecordNotFound(err) {
		return nil, ErrCartNotFound
	}
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) GetOrCreateCart(userID *uint, sessionID string) (*product.Cart, error) {
	sessionID = strings.TrimSpace(sessionID)
	if userID == nil && sessionID == "" {
		return nil, ErrCartNotFound
	}

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

func (s *CartService) ValidateAddToCart(productID uint, variantID *uint, quantity int) error {
	_, _, _, _, err := s.resolvePurchasableCartItem(productID, variantID, quantity)
	return err
}

func (s *CartService) HasPurchasableSyncItems(items []SyncCartItemReq) bool {
	for _, item := range items {
		if _, _, _, _, err := s.resolvePurchasableCartItem(item.ProductID, item.VariantID, item.Quantity); err == nil {
			return true
		}
	}
	return false
}

func (s *CartService) AddToCart(cartID, productID uint, variantID *uint, quantity int) error {
	price, itemCurrency, availableStock, resolvedVariantID, err := s.resolvePurchasableCartItem(productID, variantID, quantity)
	if err != nil {
		return err
	}

	existingItem, err := s.cartRepo.FindItem(cartID, productID, resolvedVariantID)
	if err == nil {
		if existingItem.Quantity+quantity > availableStock {
			return errors.New("insufficient stock")
		}
		existingItem.Quantity += quantity
		existingItem.Price = price
		existingItem.Currency = itemCurrency
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
		Currency:  itemCurrency,
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

	price, itemCurrency, availableStock, _ := purchasablePriceStock(variant)
	if availableStock < quantity {
		return errors.New("insufficient stock")
	}

	item.Quantity = quantity
	item.Price = price
	item.Currency = itemCurrency
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
		price, itemCurrency, _, resolvedVariantID, err := s.resolvePurchasableCartItem(req.ProductID, req.VariantID, req.Quantity)
		if err != nil {
			continue
		}

		cartItems = append(cartItems, product.CartItem{
			CartID:    cartID,
			ProductID: req.ProductID,
			VariantID: resolvedVariantID,
			Quantity:  req.Quantity,
			Price:     price,
			Currency:  itemCurrency,
		})
	}

	return s.cartRepo.BulkUpsertItems(cartItems)
}

func (s *CartService) GetCartSummary(userID *uint, sessionID string) (*product.CartSummary, error) {
	cart, err := s.FindCart(userID, sessionID)
	if errors.Is(err, ErrCartNotFound) {
		return emptyCartSummary(), nil
	}
	if err != nil {
		return nil, err
	}

	return s.cartRepo.GetSummary(cart.ID)
}

func emptyCartSummary() *product.CartSummary {
	return &product.CartSummary{
		ItemCount: 0,
		Total:     0,
		Items:     []product.CartItem{},
	}
}

func (s *CartService) ClearCart(cartID uint) error {
	return s.cartRepo.ClearCart(cartID)
}

func (s *CartService) resolvePurchasableCartItem(productID uint, variantID *uint, quantity int) (float64, string, int, *uint, error) {
	if quantity <= 0 {
		return 0, "", 0, nil, errors.New("quantity must be greater than 0")
	}

	_, variant, err := s.productRepo.FindPurchasableVariant(productID, variantID)
	if err != nil || variant == nil {
		return 0, "", 0, nil, errors.New("product not found")
	}

	price, itemCurrency, availableStock, resolvedVariantID := purchasablePriceStock(variant)
	if availableStock < quantity {
		return 0, "", 0, nil, errors.New("insufficient stock")
	}
	return price, itemCurrency, availableStock, resolvedVariantID, nil
}

func purchasablePriceStock(variant *product.ProductVariant) (float64, string, int, *uint) {
	variantID := variant.ID
	itemCurrency := currency.NormalizeCode(variant.Currency)
	if itemCurrency == "" {
		itemCurrency = product.DefaultPriceCurrency
	}
	if !currency.IsValidCode(itemCurrency) || !currency.IsCatalogCode(itemCurrency) {
		itemCurrency = product.DefaultPriceCurrency
	}
	return variant.EffectivePrice(), itemCurrency, variant.Stock, &variantID
}
