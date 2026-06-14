package wishlist

import (
	"errors"
	"net/http"
	"strconv"
	"tanzanite/internal/domain/wishlist"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	wishlistService *service.WishlistService
}

func NewHandler(wishlistService *service.WishlistService) *Handler {
	return &Handler{
		wishlistService: wishlistService,
	}
}

type AddWishlistItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

func (h *Handler) ListItems(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		respondWishlistError(c, http.StatusUnauthorized, "not_logged_in", "Please log in to view your wishlist.")
		return
	}

	items, err := h.wishlistService.List(userID)
	if err != nil {
		respondWishlistError(c, http.StatusInternalServerError, "wishlist_list_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": wishlistResponses(items),
		"meta":  gin.H{"total": len(items)},
	})
}

func (h *Handler) CreateItem(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		respondWishlistError(c, http.StatusUnauthorized, "not_logged_in", "Please log in to use wishlist.")
		return
	}

	var req AddWishlistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWishlistError(c, http.StatusBadRequest, "invalid_product_id", "Invalid product_id.")
		return
	}

	item, err := h.wishlistService.Add(userID, req.ProductID)
	if err != nil {
		if errors.Is(err, service.ErrWishlistProductNotFound) {
			respondWishlistError(c, http.StatusNotFound, "product_not_found", "Product not found.")
			return
		}
		respondWishlistError(c, http.StatusInternalServerError, "failed_add_wishlist", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"item": item.ToResponse()})
}

func (h *Handler) DeleteItem(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		respondWishlistError(c, http.StatusUnauthorized, "not_logged_in", "Please log in to use wishlist.")
		return
	}

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || itemID == 0 {
		respondWishlistError(c, http.StatusBadRequest, "invalid_id", "Invalid wishlist item id.")
		return
	}

	if err := h.wishlistService.Remove(userID, uint(itemID)); err != nil {
		switch {
		case errors.Is(err, service.ErrWishlistItemNotFound):
			respondWishlistError(c, http.StatusNotFound, "wishlist_not_found", "Wishlist item not found.")
		case errors.Is(err, service.ErrWishlistForbidden):
			respondWishlistError(c, http.StatusForbidden, "forbidden", "You cannot modify this wishlist item.")
		default:
			respondWishlistError(c, http.StatusInternalServerError, "failed_delete_wishlist", err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted": true,
		"id":      uint(itemID),
	})
}

func currentUserID(c *gin.Context) (uint, bool) {
	uid, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	userID, ok := uid.(uint)
	return userID, ok && userID > 0
}

func wishlistResponses(items []wishlist.Item) []*wishlist.ItemResponse {
	responses := make([]*wishlist.ItemResponse, 0, len(items))
	for i := range items {
		responses = append(responses, items[i].ToResponse())
	}
	return responses
}

func respondWishlistError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error":   code,
		"message": message,
	})
}
