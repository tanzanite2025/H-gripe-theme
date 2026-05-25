package service

import (
	"errors"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/repository"

	"gorm.io/gorm"
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

// GetOrCreateCart 获取或创建购物车
func (s *CartService) GetOrCreateCart(userID *uint, sessionID string) (*product.Cart, error) {
	var cart *product.Cart
	var err error

	if userID != nil {
		cart, err = s.cartRepo.FindByUserID(*userID)
	} else {
		cart, err = s.cartRepo.FindBySessionID(sessionID)
	}

	if err == gorm.ErrRecordNotFound {
		// 创建新购物车
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

// AddToCart 添加商品到购物车
func (s *CartService) AddToCart(cartID, productID uint, quantity int) error {
	// 检查产品是否存在
	prod, err := s.productRepo.FindByID(productID)
	if err != nil {
		return errors.New("product not found")
	}

	// 检查库存
	if prod.Stock < quantity {
		return errors.New("insufficient stock")
	}

	// 检查是否已存在该商品
	existingItem, err := s.cartRepo.FindItem(cartID, productID)
	if err == nil {
		// 更新数量
		existingItem.Quantity += quantity
		return s.cartRepo.UpdateItem(existingItem)
	}

	// 添加新商品
	item := &product.CartItem{
		CartID:    cartID,
		ProductID: productID,
		Quantity:  quantity,
		Price:     prod.Price,
	}

	if prod.SalePrice != nil {
		item.Price = *prod.SalePrice
	}

	return s.cartRepo.AddItem(item)
}

// UpdateCartItem 更新购物车项目数量
func (s *CartService) UpdateCartItem(cartID, productID uint, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	item, err := s.cartRepo.FindItem(cartID, productID)
	if err != nil {
		return errors.New("item not found in cart")
	}

	// 检查库存
	prod, err := s.productRepo.FindByID(productID)
	if err != nil {
		return err
	}

	if prod.Stock < quantity {
		return errors.New("insufficient stock")
	}

	item.Quantity = quantity
	return s.cartRepo.UpdateItem(item)
}

// RemoveFromCart 从购物车移除商品
func (s *CartService) RemoveFromCart(itemID uint) error {
	return s.cartRepo.RemoveItem(itemID)
}

// GetCartSummary 获取购物车摘要
func (s *CartService) GetCartSummary(userID *uint, sessionID string) (*product.CartSummary, error) {
	cart, err := s.GetOrCreateCart(userID, sessionID)
	if err != nil {
		return nil, err
	}

	return s.cartRepo.GetSummary(cart.ID)
}

// ClearCart 清空购物车
func (s *CartService) ClearCart(cartID uint) error {
	return s.cartRepo.ClearCart(cartID)
}
