package service

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCartNotFound           = errors.New("cart not found")
	ErrCartMultipleCurrencies = errors.New("cart contains multiple price currencies")
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

func (s *CartService) FindCart(userID *uint, sessionID string) (*product.Cart, error) {
	sessionID = strings.TrimSpace(sessionID)
	if userID == nil && sessionID == "" {
		return nil, ErrCartNotFound
	}

	cart, err := s.findCartByIdentity(userID, sessionID)

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

	cart, err := s.findCartByIdentity(userID, sessionID)

	if repository.IsRecordNotFound(err) {
		cart = &product.Cart{
			UserID:    userID,
			SessionID: sessionID,
		}
		if createErr := s.cartRepo.Create(cart); createErr != nil {
			// The identity lookup and insert are intentionally separate. A
			// concurrent initializer may win between them; the unique index
			// turns that race into a safe read-after-conflict.
			if repository.IsDuplicatedKey(createErr) {
				cart, err = s.findCartByIdentity(userID, sessionID)
				if err == nil {
					return cart, nil
				}
				if !repository.IsRecordNotFound(err) {
					return nil, err
				}
			}
			return nil, createErr
		}
	} else if err != nil {
		return nil, err
	}

	return cart, nil
}

func (s *CartService) findCartByIdentity(userID *uint, sessionID string) (*product.Cart, error) {
	if userID != nil {
		return s.cartRepo.FindByUserID(*userID)
	}
	return s.cartRepo.FindBySessionID(sessionID)
}

func (s *CartService) ValidateAddToCart(productID uint, variantID *uint, quantity int) error {
	_, _, _, _, _, err := s.resolvePurchasableCartItemWithConfiguration(productID, variantID, quantity, nil)
	return err
}

func (s *CartService) ValidateAddToCartWithConfiguration(productID uint, variantID *uint, quantity int, selected []SelectedOption) error {
	_, _, _, _, _, err := s.resolvePurchasableCartItemWithConfiguration(productID, variantID, quantity, selected)
	return err
}

func (s *CartService) HasPurchasableSyncItems(items []SyncCartItemReq) bool {
	for _, item := range items {
		if _, _, _, _, _, err := s.resolvePurchasableCartItemWithConfiguration(item.ProductID, item.VariantID, item.Quantity, item.SelectedOptions); err == nil {
			return true
		}
	}
	return false
}

func (s *CartService) AddToCart(cartID, productID uint, variantID *uint, quantity int) error {
	return s.AddToCartWithConfiguration(cartID, productID, variantID, quantity, nil)
}

func (s *CartService) AddToCartWithConfiguration(cartID, productID uint, variantID *uint, quantity int, selected []SelectedOption) error {
	priceMoney, availableStock, resolvedVariantID, requiresStock, configuration, err := s.resolvePurchasableCartItemWithConfiguration(productID, variantID, quantity, selected)
	if err != nil {
		return err
	}
	itemCurrency := priceMoney.Currency().String()

	if err := s.ensureCartCurrency(cartID, itemCurrency); err != nil {
		return err
	}
	existingItem, err := s.cartRepo.FindItemWithConfiguration(cartID, productID, resolvedVariantID, configuration.Hash)
	if err == nil {
		if requiresStock && existingItem.Quantity+quantity > availableStock {
			return errors.New("insufficient stock")
		}
		if err := s.validateConfigurationInventory(configuration, existingItem.Quantity+quantity); err != nil {
			return err
		}
		existingItem.Quantity += quantity
		if err := existingItem.SetPriceMoney(priceMoney); err != nil {
			return err
		}
		existingItem.ConfigurationData = configuration.Data
		existingItem.ConfigurationHash = configuration.Hash
		return s.cartRepo.UpdateItem(existingItem)
	}
	if !repository.IsRecordNotFound(err) {
		return err
	}

	item := &product.CartItem{
		CartID:            cartID,
		ProductID:         productID,
		VariantID:         resolvedVariantID,
		Quantity:          quantity,
		ConfigurationData: configuration.Data,
		ConfigurationHash: configuration.Hash,
	}
	if err := item.SetPriceMoney(priceMoney); err != nil {
		return err
	}
	return s.cartRepo.AddItem(item)
}

func (s *CartService) UpdateCartItem(cartID, productID uint, variantID *uint, quantity int) error {
	return s.UpdateCartItemWithConfiguration(cartID, productID, variantID, quantity, nil)
}

func (s *CartService) UpdateCartItemWithConfiguration(cartID, productID uint, variantID *uint, quantity int, selected []SelectedOption) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	productRecord, variant, err := s.productRepo.FindPurchasableVariant(productID, variantID)
	if err != nil || variant == nil {
		return errors.New("product not found")
	}
	configuration, err := ResolveProductConfiguration(productRecord, variant, selected)
	if err != nil {
		return err
	}
	item, err := s.cartRepo.FindItemWithConfiguration(cartID, productID, &variant.ID, configuration.Hash)
	if err != nil {
		return errors.New("item not found in cart")
	}

	priceMoney, availableStock, _, err := purchasablePriceStock(variant)
	if err != nil {
		return err
	}
	if productRequiresStock(productRecord) && availableStock < quantity {
		return errors.New("insufficient stock")
	}
	if err := s.validateConfigurationInventory(configuration, quantity); err != nil {
		return err
	}
	priceMoney, err = priceMoney.Add(configuration.Delta)
	if err != nil {
		return fmt.Errorf("calculate configured price: %w", err)
	}

	item.Quantity = quantity
	if err := item.SetPriceMoney(priceMoney); err != nil {
		return err
	}
	item.ConfigurationData = configuration.Data
	item.ConfigurationHash = configuration.Hash
	return s.cartRepo.UpdateItem(item)
}

func (s *CartService) RemoveFromCart(cartID, productID uint, variantID *uint) error {
	return s.RemoveFromCartWithConfiguration(cartID, productID, variantID, product.DefaultConfigurationHash)
}

func (s *CartService) RemoveFromCartWithConfiguration(cartID, productID uint, variantID *uint, configurationHash string) error {
	item, err := s.cartRepo.FindItemWithConfiguration(cartID, productID, variantID, configurationHash)
	if err != nil {
		return nil
	}
	return s.cartRepo.RemoveItem(item.ID)
}

type SyncCartItemReq struct {
	ProductID       uint             `json:"product_id"`
	VariantID       *uint            `json:"variant_id"`
	Quantity        int              `json:"quantity"`
	SelectedOptions []SelectedOption `json:"selected_options,omitempty"`
}

func (s *CartService) SyncCart(cartID uint, items []SyncCartItemReq) error {
	if len(items) == 0 {
		return nil
	}

	var cartItems []product.CartItem
	currencySet := make(map[string]struct{})
	for _, req := range items {
		priceMoney, _, resolvedVariantID, _, configuration, err := s.resolvePurchasableCartItemWithConfiguration(req.ProductID, req.VariantID, req.Quantity, req.SelectedOptions)
		if err != nil {
			continue
		}
		itemCurrency := priceMoney.Currency().String()
		currencySet[itemCurrency] = struct{}{}

		item := product.CartItem{
			CartID:            cartID,
			ProductID:         req.ProductID,
			VariantID:         resolvedVariantID,
			Quantity:          req.Quantity,
			ConfigurationData: configuration.Data,
			ConfigurationHash: configuration.Hash,
		}
		if err := item.SetPriceMoney(priceMoney); err != nil {
			continue
		}
		cartItems = append(cartItems, item)
	}

	if len(currencySet) > 1 {
		return ErrCartMultipleCurrencies
	}
	if len(currencySet) == 1 {
		for itemCurrency := range currencySet {
			if err := s.ensureCartCurrency(cartID, itemCurrency); err != nil {
				return err
			}
		}
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
		ItemCount:  0,
		TotalMoney: domainmoney.MustNew(0, product.DefaultPriceCurrency),
		Items:      []product.CartItem{},
	}
}

func (s *CartService) ClearCart(cartID uint) error {
	return s.cartRepo.ClearCart(cartID)
}

// MergeGuestCartOnLogin transfers the cart identified by the browser session
// to the user's persistent cart. It is safe to call when no guest cart exists.
func (s *CartService) MergeGuestCartOnLogin(userID uint, sessionID string) error {
	if s == nil || s.cartRepo == nil {
		return nil
	}
	return s.cartRepo.MergeGuestCartOnLogin(userID, sessionID)
}

func (s *CartService) ensureCartCurrency(cartID uint, itemCurrency string) error {
	summary, err := s.cartRepo.GetSummary(cartID)
	if err != nil {
		return err
	}
	for _, item := range summary.Items {
		existingCurrency := currency.NormalizeCode(item.Currency)
		if existingCurrency != "" && existingCurrency != itemCurrency {
			return ErrCartMultipleCurrencies
		}
	}
	return nil
}

func (s *CartService) resolvePurchasableCartItem(productID uint, variantID *uint, quantity int) (domainmoney.Money, int, *uint, bool, error) {
	price, stock, resolved, requires, _, err := s.resolvePurchasableCartItemWithConfiguration(productID, variantID, quantity, nil)
	return price, stock, resolved, requires, err
}

func (s *CartService) resolvePurchasableCartItemWithConfiguration(productID uint, variantID *uint, quantity int, selected []SelectedOption) (domainmoney.Money, int, *uint, bool, ProductConfigurationResult, error) {
	if quantity <= 0 {
		return domainmoney.Money{}, 0, nil, false, ProductConfigurationResult{}, errors.New("quantity must be greater than 0")
	}

	productRecord, variant, err := s.productRepo.FindPurchasableVariant(productID, variantID)
	if err != nil || variant == nil {
		return domainmoney.Money{}, 0, nil, false, ProductConfigurationResult{}, errors.New("product not found")
	}

	price, availableStock, resolvedVariantID, err := purchasablePriceStock(variant)
	if err != nil {
		return domainmoney.Money{}, 0, nil, false, ProductConfigurationResult{}, err
	}
	configuration, err := ResolveProductConfiguration(productRecord, variant, selected)
	if err != nil {
		return domainmoney.Money{}, 0, nil, false, ProductConfigurationResult{}, err
	}
	if err := s.validateConfigurationInventory(configuration, quantity); err != nil {
		return domainmoney.Money{}, 0, nil, false, ProductConfigurationResult{}, err
	}
	price, err = price.Add(configuration.Delta)
	if err != nil {
		return domainmoney.Money{}, 0, nil, false, ProductConfigurationResult{}, fmt.Errorf("calculate configured price: %w", err)
	}
	requiresStock := productRequiresStock(productRecord)
	if requiresStock && availableStock < quantity {
		return domainmoney.Money{}, 0, nil, requiresStock, ProductConfigurationResult{}, errors.New("insufficient stock")
	}
	return price, availableStock, resolvedVariantID, requiresStock, configuration, nil
}

func (s *CartService) validateConfigurationInventory(configuration ProductConfigurationResult, quantity int) error {
	if quantity <= 0 || len(configuration.InventoryAllocations) == 0 {
		return nil
	}
	items, err := ConfigurationInventoryQuantities(configuration, quantity)
	if err != nil {
		return err
	}
	if err := s.productRepo.ValidateVariantStocks(items); err != nil {
		return fmt.Errorf("insufficient component stock: %w", err)
	}
	return nil
}

func productRequiresStock(item *product.Product) bool {
	return item == nil || product.NormalizeFulfillmentMode(item.FulfillmentMode) == product.FulfillmentModeStock
}

func purchasablePriceStock(variant *product.ProductVariant) (domainmoney.Money, int, *uint, error) {
	if variant == nil {
		return domainmoney.Money{}, 0, nil, errors.New("product variant is required")
	}
	variantID := variant.ID
	itemCurrency := currency.NormalizeCode(variant.Currency)
	if itemCurrency == "" {
		itemCurrency = product.DefaultPriceCurrency
	}
	if !currency.IsValidCode(itemCurrency) || !currency.IsCatalogCode(itemCurrency) {
		return domainmoney.Money{}, 0, nil, errors.New("product price currency is invalid")
	}
	price, err := variant.EffectivePriceMoney()
	if err != nil {
		return domainmoney.Money{}, 0, nil, err
	}
	if price.Currency().String() != itemCurrency {
		return domainmoney.Money{}, 0, nil, errors.New("product price currency mismatch")
	}
	return price, variant.Stock, &variantID, nil
}
