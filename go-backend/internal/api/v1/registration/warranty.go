package registration

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	orderdomain "tanzanite/internal/domain/order"
	"tanzanite/internal/domain/registration"

	"github.com/gin-gonic/gin"
)

var (
	errWarrantyEmailMismatch      = errors.New("Email does not match order record.")
	errWarrantyStorageUnavailable = errors.New("file storage is unavailable")
)

// VerifyWarrantyOrder 验证保修订单
// @Summary 验证保修订单
// @Tags Registration
// @Accept json
// @Produce json
// @Param request body map[string]string true "订单号和邮箱"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/registrations/warranty/verify-order [post]
func (h *Handler) VerifyWarrantyOrder(c *gin.Context) {
	var req struct {
		OrderNumber string `json:"order_number" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_params", "message": "Order Number and Email are required."})
		return
	}

	if _, err := h.findVerifiedWarrantyOrder(req.OrderNumber, req.Email); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_order", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Order verified successfully."})
}

// SubmitWarrantyClaim 提交保修申请
// @Summary 提交保修申请
// @Tags Registration
// @Accept multipart/form-data
// @Produce json
// @Param order_number formData string true "订单号"
// @Param email formData string true "邮箱"
// @Param issue_description formData string true "问题描述"
// @Param tire_pressure formData string false "胎压"
// @Param is_tubeless formData string false "是否无内胎"
// @Param images formData file false "图片"
// @Param video formData file false "视频"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/registrations/warranty/claim [post]
func (h *Handler) SubmitWarrantyClaim(c *gin.Context) {
	orderNumber := strings.TrimSpace(c.PostForm("order_number"))
	email := strings.TrimSpace(c.PostForm("email"))
	if orderNumber == "" || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_params", "message": "Order Number and Email are required."})
		return
	}

	order, err := h.findVerifiedWarrantyOrder(orderNumber, email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_order", "message": err.Error()})
		return
	}

	imageURLs, videoURL, err := h.uploadWarrantyClaimFiles(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "upload_failed", "message": err.Error()})
		return
	}
	imagesJSON, err := json.Marshal(imageURLs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "encode_failed", "message": err.Error()})
		return
	}

	claim := registration.WarrantyClaim{
		UserID:       order.UserID,
		IssueType:    "warranty",
		Description:  strings.TrimSpace(c.PostForm("issue_description")),
		Images:       string(imagesJSON),
		OrderNumber:  orderNumber,
		Email:        email,
		TirePressure: strings.TrimSpace(c.PostForm("tire_pressure")),
		IsTubeless:   c.PostForm("is_tubeless") == "yes",
		VideoURL:     videoURL,
		Status:       "pending",
	}

	if err := h.registrationRepo.CreateWarrantyClaim(&claim); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "db_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Claim submitted successfully.",
		"id":      claim.ID,
	})
}

// GetExpiringWarranties 获取即将过期的保修（管理员）
// @Summary 获取即将过期的保修
// @Tags Registration
// @Produce json
// @Param days query int false "天数" default(30)
// @Success 200 {array} registration.ProductRegistration
// @Router /api/v1/admin/registrations/expiring [get]
func (h *Handler) GetExpiringWarranties(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	registrations, err := h.registrationRepo.FindExpiringWarranties(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": registrations})
}

// CreateWarrantyClaim 创建保修申请
// @Summary 创建保修申请
// @Tags Registration
// @Accept json
// @Produce json
// @Param claim body registration.WarrantyClaim true "保修申请信息"
// @Success 201 {object} registration.WarrantyClaim
// @Router /api/v1/registrations/warranty-claims [post]
func (h *Handler) CreateWarrantyClaim(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var claim registration.WarrantyClaim
	if err := c.ShouldBindJSON(&claim); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证注册记录是否属于当前用户
	reg, err := h.registrationRepo.FindRegistrationByID(claim.RegistrationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "registration not found"})
		return
	}

	if reg.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// 设置默认状态
	claim.Status = "pending"

	if err := h.registrationRepo.CreateWarrantyClaim(&claim); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, claim)
}

// GetWarrantyClaim 获取保修申请详情
// @Summary 获取保修申请详情
// @Tags Registration
// @Produce json
// @Param id path int true "保修申请ID"
// @Success 200 {object} registration.WarrantyClaim
// @Router /api/v1/registrations/warranty-claims/{id} [get]
func (h *Handler) GetWarrantyClaim(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid claim id"})
		return
	}

	claim, err := h.registrationRepo.FindWarrantyClaimByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 验证权限
	if claim.RegistrationID != 0 {
		reg, err := h.registrationRepo.FindRegistrationByID(claim.RegistrationID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "registration not found"})
			return
		}
		if reg.UserID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	} else {
		if claim.UserID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, claim)
}

// ListRegistrationClaims 获取注册的保修申请列表
// @Summary 获取注册的保修申请列表
// @Tags Registration
// @Produce json
// @Param registration_id path int true "注册ID"
// @Success 200 {array} registration.WarrantyClaim
// @Router /api/v1/registrations/{registration_id}/warranty-claims [get]
func (h *Handler) ListRegistrationClaims(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	registrationID, err := strconv.ParseUint(c.Param("registration_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid registration id"})
		return
	}

	// 验证权限
	reg, err := h.registrationRepo.FindRegistrationByID(uint(registrationID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "registration not found"})
		return
	}

	if reg.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	claims, err := h.registrationRepo.FindWarrantyClaimsByRegistrationID(uint(registrationID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": claims})
}

// ListAllWarrantyClaims 获取所有保修申请（管理员）
// @Summary 获取所有保修申请
// @Tags Registration
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/registrations/warranty-claims [get]
func (h *Handler) ListAllWarrantyClaims(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	claims, total, err := h.registrationRepo.FindAllWarrantyClaims(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": claims,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateWarrantyClaimStatus 更新保修申请状态（管理员）
// @Summary 更新保修申请状态
// @Tags Registration
// @Accept json
// @Produce json
// @Param id path int true "保修申请ID"
// @Param request body map[string]string true "状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/registrations/warranty-claims/{id}/status [put]
func (h *Handler) UpdateWarrantyClaimStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid claim id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.registrationRepo.UpdateWarrantyClaimStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "warranty claim status updated"})
}

// 私有辅助方法

// findVerifiedWarrantyOrder 查找并验证保修订单
func (h *Handler) findVerifiedWarrantyOrder(orderNumber, email string) (*orderdomain.Order, error) {
	orderNumber = strings.TrimSpace(orderNumber)
	email = strings.ToLower(strings.TrimSpace(email))
	o, err := h.orderRepo.FindByOrderNumberForVerification(orderNumber)
	if err != nil {
		return nil, err
	}

	shippingEmail := strings.ToLower(strings.TrimSpace(o.ShippingAddress.Email))
	billingEmail := strings.ToLower(strings.TrimSpace(o.BillingAddress.Email))
	if email == "" || (email != shippingEmail && email != billingEmail) {
		return nil, errWarrantyEmailMismatch
	}
	return o, nil
}

// uploadWarrantyClaimFiles 上传保修申请文件
func (h *Handler) uploadWarrantyClaimFiles(c *gin.Context) ([]string, string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return []string{}, "", nil
	}

	imageFiles := append(form.File["images[]"], form.File["images"]...)
	videoFiles := form.File["video"]
	hasFiles := len(imageFiles) > 0 || len(videoFiles) > 0
	if hasFiles && h.storageService == nil {
		return nil, "", errWarrantyStorageUnavailable
	}

	imageURLs := make([]string, 0, len(imageFiles))
	for _, file := range imageFiles {
		url, err := h.storageService.Upload(c.Request.Context(), file)
		if err != nil {
			return nil, "", err
		}
		imageURLs = append(imageURLs, url)
	}

	videoURL := ""
	if len(videoFiles) > 0 {
		url, err := h.storageService.Upload(c.Request.Context(), videoFiles[0])
		if err != nil {
			return nil, "", err
		}
		videoURL = url
	}
	return imageURLs, videoURL, nil
}
