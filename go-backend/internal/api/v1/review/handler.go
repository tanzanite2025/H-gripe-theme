package review

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tanzanite/internal/domain/review"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	reviewService *service.ReviewService
}

func NewHandler(reviewService *service.ReviewService) *Handler {
	return &Handler{
		reviewService: reviewService,
	}
}

// CreateReviewRequest 创建评价请求
type CreateReviewRequest struct {
	ProductID uint     `json:"product_id" binding:"required"`
	Rating    int      `json:"rating" binding:"required,min=1,max=5"`
	Title     string   `json:"title" binding:"required"`
	Content   string   `json:"content" binding:"required"`
	Images    []string `json:"images"`
}

// CreateReview 创建评价
// @Summary 创建评价
// @Tags Reviews
// @Accept json
// @Produce json
// @Param review body CreateReviewRequest true "评价信息"
// @Success 201 {object} review.Review
// @Router /api/v1/reviews [post]
func (h *Handler) CreateReview(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	images, _ := json.Marshal(req.Images)

	r := &review.Review{
		ProductID: req.ProductID,
		UserID:    userID.(uint),
		Rating:    req.Rating,
		Title:     req.Title,
		Content:   req.Content,
		Images:    string(images),
	}

	if err := h.reviewService.CreateReview(r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, r)
}

// GetReview 获取评价详情
// @Summary 获取评价详情
// @Tags Reviews
// @Produce json
// @Param id path int true "评价ID"
// @Success 200 {object} review.Review
// @Router /api/v1/reviews/{id} [get]
func (h *Handler) GetReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	r, err := h.reviewService.GetReview(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, r)
}

// ListProductReviews 获取产品评价列表
// @Summary 获取产品评价列表
// @Tags Reviews
// @Produce json
// @Param product_id query int true "产品ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/reviews [get]
func (h *Handler) ListProductReviews(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Query("product_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	reviews, total, err := h.reviewService.GetProductReviews(uint(productID), page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reviews,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// ListUserReviews 获取用户评价列表
// @Summary 获取用户评价列表
// @Tags Reviews
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/reviews/my [get]
func (h *Handler) ListUserReviews(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	reviews, total, err := h.reviewService.GetUserReviews(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reviews,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetFeaturedReviews 获取精选评价
// @Summary 获取精选评价
// @Tags Reviews
// @Produce json
// @Param limit query int false "数量限制" default(10)
// @Success 200 {array} review.Review
// @Router /api/v1/reviews/featured [get]
func (h *Handler) GetFeaturedReviews(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	reviews, err := h.reviewService.GetFeaturedReviews(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": reviews})
}

// GetPendingReviews 获取待审核评价（管理员）
// @Summary 获取待审核评价
// @Tags Reviews
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
func (h *Handler) GetPendingReviews(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	reviews, total, err := h.reviewService.GetPendingReviews(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reviews,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// ApproveReview 审核通过评价（管理员）
// @Summary 审核通过评价
// @Tags Reviews
// @Produce json
// @Param id path int true "评价ID"
// @Success 200 {object} map[string]interface{}
func (h *Handler) ApproveReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	if err := h.reviewService.ApproveReview(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "review approved"})
}

// RejectReview 拒绝评价（管理员）
// @Summary 拒绝评价
// @Tags Reviews
// @Produce json
// @Param id path int true "评价ID"
// @Success 200 {object} map[string]interface{}
func (h *Handler) RejectReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	if err := h.reviewService.RejectReview(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "review rejected"})
}

// SetFeatured 设置精选评价（管理员）
// @Summary 设置精选评价
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path int true "评价ID"
// @Param request body map[string]bool true "精选状态"
// @Success 200 {object} map[string]interface{}
func (h *Handler) SetFeatured(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	var req struct {
		Featured bool `json:"featured"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.reviewService.SetFeatured(uint(id), req.Featured); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "featured status updated"})
}

// DeleteReview 删除评价
// @Summary 删除评价
// @Tags Reviews
// @Produce json
// @Param id path int true "评价ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/reviews/{id} [delete]
func (h *Handler) DeleteReview(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	if err := h.reviewService.DeleteReview(uint(id), userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "review deleted"})
}

// MarkHelpful 标记评价有用
// @Summary 标记评价有用
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path int true "评价ID"
// @Param request body map[string]bool true "是否有用"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/reviews/{id}/helpful [post]
func (h *Handler) MarkHelpful(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	var req struct {
		IsHelpful bool `json:"is_helpful"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.reviewService.MarkHelpful(uint(id), userID.(uint), req.IsHelpful); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "marked successfully"})
}

// GetReviewSummary 获取产品评价摘要
// @Summary 获取产品评价摘要
// @Tags Reviews
// @Produce json
// @Param product_id path int true "产品ID"
// @Success 200 {object} review.ReviewSummary
// @Router /api/v1/reviews/summary/{product_id} [get]
func (h *Handler) GetReviewSummary(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	summary, err := h.reviewService.GetReviewSummary(uint(productID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}
