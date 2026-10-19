package privateattachment

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/storage"
)

type privateAttachmentResponseClosureRecorder struct {
	io.Reader
	closed bool
}

func (r *privateAttachmentResponseClosureRecorder) Close() error { r.closed = true; return nil }

func TestAuthorizedPrivateAttachmentResponseNeverCachesAndOnlyPreviewsSupportedContentTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, fixture := range []struct{ name, contentType, query, disposition string }{
		{"download image", "image/png", "", "attachment"},
		{"preview image", "image/png", "?preview=1", "inline"},
		{"preview video", "video/mp4", "?preview=1", "inline"},
		{"preview PDF", "application/pdf", "?preview=1", "inline"},
		{"HTML forced download", "text/html", "?preview=1", "attachment"},
		{"SVG forced download", "image/svg+xml", "?preview=1", "attachment"},
		{"malformed MIME forced download", "invalid mime", "?preview=1", "attachment"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequestWithContext(t.Context(), "GET", "/file"+fixture.query, nil)
			reader := &privateAttachmentResponseClosureRecorder{Reader: strings.NewReader("private evidence")}
			ServeAuthorizedPrivateAttachmentFile(c, &storage.StoredObject{ReadCloser: reader, Name: "../../客户证据.png", MimeType: fixture.contentType, Size: 16})
			require.Equal(t, 200, recorder.Code)
			require.Equal(t, "private evidence", recorder.Body.String())
			require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
			require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
			require.Equal(t, "no-referrer", recorder.Header().Get("Referrer-Policy"))
			require.Contains(t, recorder.Header().Get("Content-Security-Policy"), "sandbox")
			require.True(t, strings.HasPrefix(recorder.Header().Get("Content-Disposition"), fixture.disposition+";"))
			require.NotContains(t, recorder.Header().Get("Content-Disposition"), "../")
			require.True(t, reader.closed)
		})
	}
}
