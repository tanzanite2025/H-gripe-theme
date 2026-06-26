package registration

import (
	"strings"
	"tanzanite/internal/domain/registration"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

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
		apierror.RespondValidationError(c, err.Error())
		return
	}

	reg, err := h.registrationRepo.FindRegistrationBySerialNumber(req.SerialNumber)
	if err != nil {
		apierror.RespondNotFound(c, "Serial number")
		return
	}

	// 清除可能导致隐私数据泄漏的用户信息
	reg.User = nil

	response.Success(c, gin.H{
		"valid":            true,
		"registration":     reg,
		"warranty_expires": reg.WarrantyExpires,
	})
}

// GetWarrantyStatus 获取保修状态
// @Summary 获取保修状态
// @Tags Registration
// @Produce json
// @Param code path string true "产品代码"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/registrations/warranty/{code} [get]
func (h *Handler) GetWarrantyStatus(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		apierror.RespondBadRequest(c, "Product code is required")
		return
	}

	reg, err := h.registrationRepo.FindRegistrationBySerialNumber(code)
	if err != nil {
		apierror.RespondNotFound(c, "Product")
		return
	}

	response.Success(c, gin.H{
		"success": true,
		"data":    warrantyStatusResponse(reg),
	})
}

// warrantyStatusResponse 格式化保修状态响应
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
