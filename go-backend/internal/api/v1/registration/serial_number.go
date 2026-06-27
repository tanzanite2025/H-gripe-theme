package registration

import (
	"strings"
	domainregistration "tanzanite/internal/domain/registration"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) VerifySerialNumber(c *gin.Context) {
	var req struct {
		SerialNumber string `json:"serial_number" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	reg, err := h.registrationSvc.VerifySerialNumber(req.SerialNumber)
	if err != nil {
		apierror.RespondNotFound(c, "Serial number")
		return
	}

	reg.User = nil
	response.Success(c, gin.H{
		"valid":            true,
		"registration":     reg,
		"warranty_expires": reg.WarrantyExpires,
	})
}

func (h *Handler) GetWarrantyStatus(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		apierror.RespondBadRequest(c, "Product code is required")
		return
	}

	reg, err := h.registrationSvc.VerifySerialNumber(code)
	if err != nil {
		apierror.RespondNotFound(c, "Product")
		return
	}

	response.Success(c, gin.H{
		"success": true,
		"data":    warrantyStatusResponse(reg),
	})
}

func warrantyStatusResponse(reg *domainregistration.ProductRegistration) gin.H {
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
