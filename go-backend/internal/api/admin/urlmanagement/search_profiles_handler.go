package urlmanagement

import (
	"errors"
	"net/http"
	"strconv"

	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SearchProfilesHandler struct {
	profiles *service.StorefrontURLSearchProfileService
}

func NewSearchProfilesHandler(profiles *service.StorefrontURLSearchProfileService) *SearchProfilesHandler {
	return &SearchProfilesHandler{profiles: profiles}
}

func (h *SearchProfilesHandler) List(c *gin.Context) {
	locale := c.Query("locale")
	profiles, err := h.profiles.List(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": profiles})
}

func (h *SearchProfilesHandler) Get(c *gin.Context) {
	routeEntryID, err := searchProfileRouteEntryID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.profiles.Get(routeEntryID)
	if err != nil {
		writeSearchProfileError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": profile})
}

func (h *SearchProfilesHandler) Upsert(c *gin.Context) {
	routeEntryID, err := searchProfileRouteEntryID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input urlmanagementdomain.StorefrontURLSearchProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.profiles.Upsert(routeEntryID, input)
	if err != nil {
		writeSearchProfileError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": profile})
}

func searchProfileRouteEntryID(value string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, errors.New("invalid route entry ID")
	}
	return uint(parsed), nil
}

func writeSearchProfileError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, gorm.ErrRecordNotFound) {
		status = http.StatusNotFound
	}
	c.JSON(status, gin.H{"error": err.Error()})
}
