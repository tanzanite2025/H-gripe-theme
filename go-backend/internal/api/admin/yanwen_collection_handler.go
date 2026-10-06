package admin

import (
	"errors"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// YanwenCollectionHandler exposes only the curated Yanwen collection boundary to the
// admin console and to the main shipping management domain.
type YanwenCollectionHandler struct {
	collectionService *service.YanwenPublishedCollectionService
}

func NewYanwenCollectionHandler(collectionService *service.YanwenPublishedCollectionService) *YanwenCollectionHandler {
	return &YanwenCollectionHandler{collectionService: collectionService}
}

func (h *YanwenCollectionHandler) ListYanwenPublishedChannels(c *gin.Context) {
	environment, err := shipping.NormalizeYanwenPublishedChannelEnvironment(c.Query("environment"))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	enabledOnly := c.Query("enabled") == "true"
	channels, err := h.collectionService.ListYanwenPublishedChannels(environment, enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": newYanwenPublishedChannelManagementDTOs(channels)})
}

func (h *YanwenCollectionHandler) ListYanwenPublishedCollectionReferences(c *gin.Context) {
	references, err := h.collectionService.ListProductionYanwenCollectionReferences()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	dtoReferences := make([]YanwenPublishedCollectionReferenceDTO, 0, len(references))
	for _, reference := range references {
		dtoReferences = append(dtoReferences, newYanwenPublishedCollectionReferenceDTO(reference))
	}
	response.Success(c, gin.H{"data": dtoReferences})
}

func (h *YanwenCollectionHandler) CreateYanwenPublishedChannel(c *gin.Context) {
	var request yanwenPublishedChannelRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	channel := request.newYanwenPublishedChannelDomain()
	if err := h.collectionService.CreateYanwenPublishedChannel(&channel); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Created(c, newYanwenPublishedChannelManagementDTO(channel))
}

func (h *YanwenCollectionHandler) UpdateYanwenPublishedChannel(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid Yanwen channel id")
	if err != nil {
		return
	}
	var request yanwenPublishedChannelRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	channel, err := h.collectionService.GetYanwenPublishedChannel(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierror.RespondNotFound(c, "Yanwen channel")
		} else {
			apierror.RespondInternalError(c, err)
		}
		return
	}
	if err := request.applyToYanwenPublishedChannel(channel); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if err := h.collectionService.UpdateYanwenPublishedChannel(channel); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, newYanwenPublishedChannelManagementDTO(*channel))
}

func (h *YanwenCollectionHandler) DeleteYanwenPublishedChannel(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid Yanwen channel id")
	if err != nil {
		return
	}
	if err := h.collectionService.DeleteYanwenPublishedChannel(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierror.RespondNotFound(c, "Yanwen channel")
			return
		}
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "Yanwen channel deleted", nil)
}
