package admin

import (
	"errors"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	trackingProviderEnvironmentProduction = "production"
	trackingProviderEnvironmentSandbox    = "sandbox"
)

func (h *ShippingHandler) ListTrackingProviderConfigs(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	providers, err := h.shippingService.ListTrackingProviderConfigs(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": trackingProviderResponses(providers)})
}

func (h *ShippingHandler) GetTrackingProviderConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid tracking provider id")
	if err != nil {
		return
	}

	provider, err := h.shippingService.GetTrackingProviderConfig(id)
	if err != nil {
		apierror.RespondNotFound(c, "Tracking provider")
		return
	}

	response.Success(c, trackingProviderResponse(*provider))
}

func (h *ShippingHandler) CreateTrackingProviderConfig(c *gin.Context) {
	var req shippingTrackingProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	provider := req.toDomain()
	if err := validateTrackingProviderConfig(provider); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.CreateTrackingProviderConfig(&provider); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, trackingProviderResponse(provider))
}

func (h *ShippingHandler) UpdateTrackingProviderConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid tracking provider id")
	if err != nil {
		return
	}

	existing, err := h.shippingService.GetTrackingProviderConfig(id)
	if err != nil {
		apierror.RespondNotFound(c, "Tracking provider")
		return
	}

	var req shippingTrackingProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	provider := req.toDomain()
	provider.ID = id
	if provider.APIKey == "" {
		provider.APIKey = existing.APIKey
	}
	if provider.WebhookSecret == "" {
		provider.WebhookSecret = existing.WebhookSecret
	}
	if err := validateTrackingProviderConfig(provider); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.UpdateTrackingProviderConfig(&provider); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	updated, err := h.shippingService.GetTrackingProviderConfig(id)
	if err != nil {
		response.Success(c, trackingProviderResponse(provider))
		return
	}

	response.Success(c, trackingProviderResponse(*updated))
}

func (h *ShippingHandler) DeleteTrackingProviderConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid tracking provider id")
	if err != nil {
		return
	}

	if err := h.shippingService.DeleteTrackingProviderConfig(id); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "tracking provider deleted", nil)
}

func validateTrackingProviderConfig(provider shippingdomain.TrackingProviderConfig) error {
	if provider.ProviderCode == "" {
		return errors.New("provider code is required")
	}
	if provider.ProviderName == "" {
		return errors.New("provider name is required")
	}

	switch provider.Environment {
	case trackingProviderEnvironmentProduction, trackingProviderEnvironmentSandbox:
	default:
		return errors.New("tracking provider environment must be production or sandbox")
	}

	if provider.PollingIntervalMinutes < 0 ||
		provider.RequestTimeoutSeconds < 0 ||
		provider.SortOrder < 0 {
		return errors.New("tracking provider numeric fields cannot be negative")
	}

	return nil
}

func trackingProviderResponses(providers []shippingdomain.TrackingProviderConfig) []shippingTrackingProviderResponse {
	result := make([]shippingTrackingProviderResponse, 0, len(providers))
	for _, provider := range providers {
		result = append(result, trackingProviderResponse(provider))
	}
	return result
}

func trackingProviderResponse(provider shippingdomain.TrackingProviderConfig) shippingTrackingProviderResponse {
	return shippingTrackingProviderResponse{
		ID:                      provider.ID,
		ProviderCode:            provider.ProviderCode,
		ProviderName:            provider.ProviderName,
		Environment:             provider.Environment,
		BaseURL:                 provider.BaseURL,
		APIKeyConfigured:        provider.APIKey != "",
		WebhookSecretConfigured: provider.WebhookSecret != "",
		WebhookEnabled:          provider.WebhookEnabled,
		AutoRegister:            provider.AutoRegister,
		PollingEnabled:          provider.PollingEnabled,
		PollingIntervalMinutes:  provider.PollingIntervalMinutes,
		RequestTimeoutSeconds:   provider.RequestTimeoutSeconds,
		Enabled:                 provider.Enabled,
		SortOrder:               provider.SortOrder,
		Description:             provider.Description,
		CreatedAt:               provider.CreatedAt,
		UpdatedAt:               provider.UpdatedAt,
	}
}
