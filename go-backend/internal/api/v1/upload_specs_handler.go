package v1

import (
	"net/http"

	"commerce-platform/internal/pkg/upload"

	"github.com/gin-gonic/gin"
)

// GetUploadSpecs exposes the public upload contract so every client can show
// the same purpose-specific format, size, and dimension guidance.
func GetUploadSpecs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"specs": upload.ListUploadSpecs(),
	})
}
