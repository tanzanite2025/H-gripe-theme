package admin

import (
	"strconv"

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

func parseIntParam(c *gin.Context, name, message string) (int, error) {
	value, err := strconv.Atoi(c.Param(name))
	if err != nil || value <= 0 {
		apierror.RespondBadRequest(c, message)
		if err == nil {
			err = strconv.ErrSyntax
		}
		return 0, err
	}
	return value, nil
}
