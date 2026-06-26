package registration

import (
	"strconv"
	"tanzanite/internal/domain/registration"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

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
		apierror.RespondUnauthorized(c)
		return
	}

	var reg registration.ProductRegistration
	if err := c.ShouldBindJSON(&reg); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	// 设置用户ID
	reg.UserID = userID.(uint)
	reg.Status = "active"

	// 检查序列号是否已存在
	exists, err := h.registrationRepo.CheckSerialNumberExists(reg.SerialNumber)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if exists {
		apierror.RespondConflict(c, "Serial number already registered")
		return
	}

	if err := h.registrationRepo.CreateRegistration(&reg); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, reg)
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
		apierror.RespondUnauthorized(c)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid registration ID")
		return
	}

	reg, err := h.registrationRepo.FindRegistrationByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Registration")
		return
	}

	// 验证权限
	if reg.UserID != userID.(uint) {
		apierror.RespondForbidden(c)
		return
	}

	response.Success(c, reg)
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
		apierror.RespondUnauthorized(c)
		return
	}

	params := pagination.ParsePagination(c)

	registrations, total, err := h.registrationRepo.FindRegistrationsByUserID(userID.(uint), params.Page, params.PageSize)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, registrations, params.Page, params.PageSize, total)
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
	params := pagination.ParsePagination(c)
	status := c.Query("status")

	registrations, total, err := h.registrationRepo.FindAllRegistrations(params.Page, params.PageSize, status)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, registrations, params.Page, params.PageSize, total)
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
		apierror.RespondUnauthorized(c)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid registration ID")
		return
	}

	// 检查权限
	existing, err := h.registrationRepo.FindRegistrationByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Registration")
		return
	}

	if existing.UserID != userID.(uint) {
		apierror.RespondForbidden(c)
		return
	}

	var reg registration.ProductRegistration
	if err := c.ShouldBindJSON(&reg); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	reg.ID = uint(id)
	reg.UserID = userID.(uint)

	if err := h.registrationRepo.UpdateRegistration(&reg); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, reg)
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
		apierror.RespondBadRequest(c, "Invalid registration ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	if err := h.registrationRepo.UpdateRegistrationStatus(uint(id), req.Status); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Registration status updated", nil)
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
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, stats)
}
