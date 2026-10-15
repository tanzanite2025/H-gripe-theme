package admin

import (
	"net/http"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/upload"

	"github.com/gin-gonic/gin"
)

const faqAnswerImageMaxRequestBytes = 4 << 20

// UploadAnswerImage 上传 FAQ 专用答案图片
// POST /api/admin/faqs/answer-image
func (h *FAQHandler) UploadAnswerImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, faqAnswerImageMaxRequestBytes)
	file, err := c.FormFile("file")
	if err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "FAQ answer image is too large", "code": upload.CodeFileTooLarge})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}
	if err := upload.ValidateSpecFile(file, string(upload.SpecFAQAnswerImage)); err != nil {
		c.JSON(upload.HTTPStatus(err), gin.H{
			"error": err.Error(),
			"code":  upload.ErrorCode(err),
		})
		return
	}

	url, err := h.faqService.UploadAnswerImage(c.Request.Context(), file, media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: currentUserID(c)}, c.PostForm("alt"))
	if err != nil {
		respondMediaError(c, err)
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
