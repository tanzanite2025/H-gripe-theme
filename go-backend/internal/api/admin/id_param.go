package admin

import (
	"commerce-platform/internal/pkg/apierror"

	"github.com/gin-gonic/gin"
)

func parsePositiveUintParam(c *gin.Context, name, message string) (uint, bool) {
	id, err := parseUintParam(c, name, message)
	if err != nil || id == 0 {
		if err == nil {
			apierror.RespondBadRequest(c, message)
		}
		return 0, false
	}
	return id, true
}
