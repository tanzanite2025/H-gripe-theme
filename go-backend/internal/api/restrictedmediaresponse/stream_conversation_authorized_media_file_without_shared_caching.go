package restrictedmediaresponse

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"commerce-platform/internal/service"
)

func SetConversationAttachmentPrivateResponseHeaders(c *gin.Context) {
	setConversationAttachmentPrivateHTTPHeaders(c.Writer.Header())
}

func setConversationAttachmentPrivateHTTPHeaders(headers http.Header) {
	headers.Set("Cache-Control", "private, no-store")
	headers.Set("X-Content-Type-Options", "nosniff")
	headers.Set("Referrer-Policy", "no-referrer")
	for _, value := range headers.Values("Vary") {
		for _, field := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(field), "Cookie") {
				return
			}
		}
	}
	headers.Add("Vary", "Cookie")
}

type conversationAttachmentPrivateHTTPResponseWriter struct{ http.ResponseWriter }

func (w conversationAttachmentPrivateHTTPResponseWriter) WriteHeader(statusCode int) {
	// ServeContent clears cache headers on errors such as an invalid Range.
	// Reapply the conversation policy immediately before committing any status.
	setConversationAttachmentPrivateHTTPHeaders(w.Header())
	w.ResponseWriter.WriteHeader(statusCode)
}

func StreamConversationAuthorizedMediaFile(c *gin.Context, file *service.MediaAssetFile) {
	SetConversationAttachmentPrivateResponseHeaders(c)
	if file == nil || file.ReadCloser == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	defer func() { _ = file.ReadCloser.Close() }()
	mimeType := strings.TrimSpace(file.MimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", file.Filename))
	if file.ByteRangeSource != nil {
		streamConversationAuthorizedRemoteVideoByteRange(c, file.ByteRangeSource)
		return
	}
	if seekable, ok := file.ReadCloser.(io.ReadSeeker); ok {
		http.ServeContent(conversationAttachmentPrivateHTTPResponseWriter{c.Writer}, c.Request, file.Filename, file.ModTime, seekable)
		return
	}
	c.DataFromReader(http.StatusOK, file.Size, mimeType, file.ReadCloser, nil)
}
