package admin

import (
	"net/http"
	"strings"
	"tanzanite/internal/pkg/upload"

	"github.com/gin-gonic/gin"
)

// UploadAnswerImage 上传 FAQ 专用答案图片
// POST /api/admin/faqs/answer-image
func (h *FAQHandler) UploadAnswerImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if err := upload.ValidateFile(file, upload.FAQAnswerImageRule); err != nil {
		c.JSON(upload.HTTPStatus(err), gin.H{
			"error": err.Error(),
			"code":  upload.ErrorCode(err),
		})
		return
	}
	if err := upload.ValidateWebPDimensions(file, 800, 800); err != nil {
		c.JSON(upload.HTTPStatus(err), gin.H{
			"error": err.Error(),
			"code":  upload.ErrorCode(err),
		})
		return
	}

	url, err := h.faqService.UploadAnswerImage(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload FAQ answer image"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"image": gin.H{
			"url":    url,
			"alt":    strings.TrimSpace(c.PostForm("alt")),
			"width":  800,
			"height": 800,
		},
	})
}
