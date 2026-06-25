package registration

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	orderdomain "tanzanite/internal/domain/order"
	"tanzanite/internal/domain/registration"
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	registrationRepo *repository.RegistrationRepository
	orderRepo        *repository.OrderRepository
	storageService   storage.StorageService
}

func NewHandler(registrationRepo *repository.RegistrationRepository, orderRepo *repository.OrderRepository, storageService storage.StorageService) *Handler {
	return &Handler{
		registrationRepo: registrationRepo,
		orderRepo:        orderRepo,
		storageService:   storageService,
	}
}

// Product Registration 相关接口

// CreateRegistration 创建产品注册
// @Summary 创建产品注册
// @Tags Registration
// @Accept json
// @Produce json
// @Param registration body registration.ProductRegistration true "注册信息"
// @Success 201 {object} registration.ProductRegistration
// @Router /api/v1/registrations [post]
func (h *Handler) CreateRegistration(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var reg registration.ProductRegistration
	if err := c.ShouldBindJSON(&reg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置用户ID
	reg.UserID = userID.(uint)
	reg.Status = "active"

	// 检查序列号是否已存在
	exists, err := h.registrationRepo.CheckSerialNumberExists(reg.SerialNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "serial number already registered"})
		return
	}

	if err := h.registrationRepo.CreateRegistration(&reg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reg)
}

// GetRegistration 获取注册详情
// @Summary 获取注册详情
// @Tags Registration
// @Produce json
// @Param id path int true "注册ID"
// @Success 200 {object} registration.ProductRegistration
// @Router /api/v1/registrations/{id} [get]
func (h *Handler) GetRegistration(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid registration id"})
		return
	}

	reg, err := h.registrationRepo.FindRegistrationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 验证权限
	if reg.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, reg)
}

// ListUserRegistrations 获取用户的注册列表
// @Summary 获取用户的注册列表
// @Tags Registration
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/registrations [get]
func (h *Handler) ListUserRegistrations(c *gin.Context) {
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

	registrations, total, err := h.registrationRepo.FindRegistrationsByUserID(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": registrations,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// ListAllRegistrations 获取所有注册（管理员）
// @Summary 获取所有注册
// @Tags Registration
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/registrations [get]
func (h *Handler) ListAllRegistrations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	registrations, total, err := h.registrationRepo.FindAllRegistrations(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": registrations,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateRegistration 更新注册信息
// @Summary 更新注册信息
// @Tags Registration
// @Accept json
// @Produce json
// @Param id path int true "注册ID"
// @Param registration body registration.ProductRegistration true "注册信息"
// @Success 200 {object} registration.ProductRegistration
// @Router /api/v1/registrations/{id} [put]
func (h *Handler) UpdateRegistration(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid registration id"})
		return
	}

	// 检查权限
	existing, err := h.registrationRepo.FindRegistrationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if existing.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var reg registration.ProductRegistration
	if err := c.ShouldBindJSON(&reg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reg.ID = uint(id)
	reg.UserID = userID.(uint)

	if err := h.registrationRepo.UpdateRegistration(&reg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reg)
}

// UpdateRegistrationStatus 更新注册状态（管理员）
// @Summary 更新注册状态
// @Tags Registration
// @Accept json
// @Produce json
// @Param id path int true "注册ID"
// @Param request body map[string]string true "状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/registrations/{id}/status [put]
func (h *Handler) UpdateRegistrationStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid registration id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.registrationRepo.UpdateRegistrationStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "registration status updated"})
}

// VerifySerialNumber 验证序列号
// @Summary 验证序列号
// @Tags Registration
// @Accept json
// @Produce json
// @Param request body map[string]string true "序列号"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/registrations/verify [post]
func (h *Handler) VerifySerialNumber(c *gin.Context) {
	var req struct {
		SerialNumber string `json:"serial_number" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reg, err := h.registrationRepo.FindRegistrationBySerialNumber(req.SerialNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"valid":   false,
			"message": "serial number not found",
		})
		return
	}

	// 清除可能导致隐私数据泄漏的用户信息
	reg.User = nil

	c.JSON(http.StatusOK, gin.H{
		"valid":            true,
		"registration":     reg,
		"warranty_expires": reg.WarrantyExpires,
	})
}

func (h *Handler) GetWarrantyStatus(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_code", "message": "Product code is required."})
		return
	}

	reg, err := h.registrationRepo.FindRegistrationBySerialNumber(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Product not found."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    warrantyStatusResponse(reg),
	})
}

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

// GetRegistrationStats 获取注册统计（管理员）
// @Summary 获取注册统计
// @Tags Registration
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/registrations/stats [get]
func (h *Handler) GetRegistrationStats(c *gin.Context) {
	stats, err := h.registrationRepo.GetRegistrationStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Warranty Claim 相关接口

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

var (
	errWarrantyEmailMismatch      = errors.New("Email does not match order record.")
	errWarrantyStorageUnavailable = errors.New("file storage is unavailable")
)

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

func warrantyStatusResponse(reg *registration.ProductRegistration) gin.H {
	now := time.Now()
	status := "expired"
	if reg.WarrantyExpires.After(now) && reg.Status != "expired" {
		status = "valid"
	}

	remaining := gin.H{"months": 0, "days": 0, "total_days": 0}
	if status == "valid" {
		days := int(reg.WarrantyExpires.Sub(now).Hours() / 24)
		remaining["months"] = days / 30
		remaining["days"] = days % 30
		remaining["total_days"] = days
	} else {
		days := int(now.Sub(reg.WarrantyExpires).Hours() / 24)
		if days < 0 {
			days = 0
		}
		remaining["expired_days"] = days
	}

	productTypeCode := ""
	productTypeName := ""
	productName := ""
	if reg.Product != nil {
		productTypeCode = reg.Product.SKU
		productTypeName = reg.Product.Name
		productName = reg.Product.Name
	}

	return gin.H{
		"product_code":    reg.SerialNumber,
		"product_type":    gin.H{"code": productTypeCode, "name": productTypeName, "name_zh": productTypeName},
		"product_name":    productName,
		"ship_date":       reg.PurchaseDate.Format("2006-01"),
		"warranty_months": reg.WarrantyPeriod,
		"warranty_end":    reg.WarrantyExpires.Format("2006-01"),
		"status":          status,
		"remaining":       remaining,
		"records":         []gin.H{},
	}
}
