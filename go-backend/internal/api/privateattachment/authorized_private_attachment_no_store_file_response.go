package privateattachment

import (
	"mime"
	"net/http"
	"path"
	"strings"

	"commerce-platform/internal/pkg/storage"
	"github.com/gin-gonic/gin"
)

func SetPrivateAttachmentResponseHeaders(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
}

func ServeAuthorizedPrivateAttachmentFile(c *gin.Context, object *storage.StoredObject) {
	SetPrivateAttachmentResponseHeaders(c)
	defer func() { _ = object.ReadCloser.Close() }()
	contentType, _, err := mime.ParseMediaType(object.MimeType)
	if err != nil {
		contentType = "application/octet-stream"
	}
	disposition := "attachment"
	if c.Query("preview") == "1" {
		switch contentType {
		case "image/jpeg", "image/png", "image/webp", "image/gif", "image/avif", "video/mp4", "video/webm", "video/quicktime", "application/pdf":
			disposition = "inline"
		}
	}
	filename := path.Base(strings.ReplaceAll(object.Name, "\\", "/"))
	if filename == "." || filename == "/" || filename == "" {
		filename = "attachment"
	}
	c.Header("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": filename}))
	c.DataFromReader(http.StatusOK, object.Size, contentType, object.ReadCloser, nil)
}
