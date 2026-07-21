package admin

import (
	"errors"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	shippingBindingScopeDefault     = "default"
	shippingBindingScopeProductType = "product_type"
	shippingBindingScopeProduct     = "product"
	shippingBindingScopeVariant     = "variant"
)

func (h *ShippingHandler) ListTemplateBindings(c *gin.Context) {
	bindings, err := h.shippingService.ListTemplateBindings()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": bindings})
}

func (h *ShippingHandler) GetTemplateBinding(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid binding id")
	if err != nil {
		return
	}

	binding, err := h.shippingService.GetTemplateBinding(id)
	if err != nil {
		apierror.RespondNotFound(c, "Shipping template binding")
		return
	}

	response.Success(c, binding)
}

func (h *ShippingHandler) CreateTemplateBinding(c *gin.Context) {
	var req shippingTemplateBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	binding := req.toDomain()
	normalizeBindingScope(&binding)
	if err := validateTemplateBinding(binding); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.CreateTemplateBinding(&binding); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, binding)
}

func (h *ShippingHandler) UpdateTemplateBinding(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid binding id")
	if err != nil {
		return
	}

	var req shippingTemplateBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	binding := req.toDomain()
	binding.ID = id
	normalizeBindingScope(&binding)
	if err := validateTemplateBinding(binding); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.UpdateTemplateBinding(&binding); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, binding)
}

func (h *ShippingHandler) DeleteTemplateBinding(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid binding id")
	if err != nil {
		return
	}

	if err := h.shippingService.DeleteTemplateBinding(id); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "shipping template binding deleted", nil)
}

func normalizeBindingScope(binding *shippingdomain.ShippingTemplateBinding) {
	switch binding.Scope {
	case shippingBindingScopeDefault:
		binding.ProductTypeID = nil
		binding.ProductID = nil
		binding.VariantID = nil
	case shippingBindingScopeProductType:
		binding.ProductID = nil
		binding.VariantID = nil
	case shippingBindingScopeProduct:
		binding.ProductTypeID = nil
		binding.VariantID = nil
	case shippingBindingScopeVariant:
		binding.ProductTypeID = nil
		binding.ProductID = nil
	}
}

func validateTemplateBinding(binding shippingdomain.ShippingTemplateBinding) error {
	if binding.TemplateID == 0 {
		return errors.New("template id is required")
	}

	switch binding.Scope {
	case shippingBindingScopeDefault:
		return nil
	case shippingBindingScopeProductType:
		if binding.ProductTypeID == nil || *binding.ProductTypeID == 0 {
			return errors.New("product type id is required")
		}
	case shippingBindingScopeProduct:
		if binding.ProductID == nil || *binding.ProductID == 0 {
			return errors.New("product id is required")
		}
	case shippingBindingScopeVariant:
		if binding.VariantID == nil || *binding.VariantID == 0 {
			return errors.New("variant id is required")
		}
	default:
		return errors.New("scope must be default, product_type, product or variant")
	}

	return nil
}
