package restrictedmediaresponse

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/service"
)

type syntheticRemotePrivateVideoByteRangeSource struct {
	metadata     storage.ObjectByteRangeMetadata
	opened       [][2]int64
	failed       bool
	closed       bool
	truncateBody bool
}

func (s *syntheticRemotePrivateVideoByteRangeSource) Metadata() storage.ObjectByteRangeMetadata {
	return s.metadata
}
func (s *syntheticRemotePrivateVideoByteRangeSource) Read([]byte) (int, error) {
	return 0, errors.New("HTTP handler must select a range before reading")
}
func (s *syntheticRemotePrivateVideoByteRangeSource) Close() error { s.closed = true; return nil }
func (s *syntheticRemotePrivateVideoByteRangeSource) OpenByteRange(ctx context.Context, start, length int64) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.opened = append(s.opened, [2]int64{start, length})
	if s.failed {
		return nil, storage.ErrRemoteObjectByteRangeUnavailable
	}
	if s.truncateBody {
		return io.NopCloser(strings.NewReader("2")), nil
	}
	return io.NopCloser(strings.NewReader("0123456789"[start : start+length])), nil
}

func TestRemotePrivateVideoRangeEmptyObjectsAvoidBodyReadsAndInterruptedBodiesReachClientsAsTruncated(t *testing.T) {
	for _, scenario := range []struct {
		name      string
		size      int64
		byteRange string
		status    int
	}{
		{"empty full object", 0, "", 200},
		{"empty object with range", 0, "bytes=0-", 416},
		{"interrupted body", 10, "bytes=2-5", 206},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			source := &syntheticRemotePrivateVideoByteRangeSource{metadata: storage.ObjectByteRangeMetadata{Size: scenario.size, ETag: `"version"`}, truncateBody: scenario.size > 0}
			router := gin.New()
			router.GET("/authorized", func(c *gin.Context) {
				StreamConversationAuthorizedMediaFile(c, &service.MediaAssetFile{ReadCloser: source, ByteRangeSource: source, MimeType: "video/mp4"})
			})
			server := httptest.NewServer(router)
			defer server.Close()
			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/authorized", nil)
			require.NoError(t, err)
			request.Header.Set("Range", scenario.byteRange)
			response, err := server.Client().Do(request)
			require.NoError(t, err)
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			require.Equal(t, scenario.status, response.StatusCode)
			require.Equal(t, "private, no-store", response.Header.Get("Cache-Control"))
			if scenario.size > 0 {
				require.ErrorIs(t, err, io.ErrUnexpectedEOF)
				require.Equal(t, "2", string(body))
				require.Equal(t, int64(4), response.ContentLength)
			} else {
				require.NoError(t, err)
				require.Empty(t, body)
			}
		})
	}
}

func TestRemotePrivateVideoRangeInt64MaximumObjectSizeDoesNotOverflowOffsets(t *testing.T) {
	const maximumSize int64 = 1<<63 - 1
	for _, scenario := range []struct {
		value         string
		start, length int64
	}{
		{"bytes=9223372036854775806-", maximumSize - 1, 1},
		{"bytes=0-9223372036854775807", 0, maximumSize},
		{"bytes=-9223372036854775807", 0, maximumSize},
		{"bytes=-1", maximumSize - 1, 1},
	} {
		t.Run(scenario.value, func(t *testing.T) {
			start, length, partial, unsatisfiable := parseConversationAuthorizedSingleVideoByteRange(scenario.value, maximumSize)
			require.Equal(t, scenario.start, start)
			require.Equal(t, scenario.length, length)
			require.True(t, partial)
			require.False(t, unsatisfiable)
		})
	}
}

func TestRemotePrivateVideoRangeHTTPHandlesExactSingleRangesAndPrivateFailuresWithoutBuffering(t *testing.T) {
	for _, scenario := range []struct {
		name, method, byteRange, ifRange string
		status                           int
		expected                         string
		start, length                    int64
		noRead                           bool
	}{
		{"full", "GET", "", "", 200, "0123456789", 0, 10, false},
		{"closed", "GET", "bytes=2-5", "", 206, "2345", 2, 4, false},
		{"open", "GET", "bytes=7-", "", 206, "789", 7, 3, false},
		{"suffix", "GET", "bytes=-3", "", 206, "789", 7, 3, false},
		{"oversized suffix", "GET", "bytes=-184467440737095516160000", "", 206, "0123456789", 0, 10, false},
		{"oversized end", "GET", "bytes=7-184467440737095516160000", "", 206, "789", 7, 3, false},
		{"matching If-Range", "GET", "bytes=2-5", `"version"`, 206, "2345", 2, 4, false},
		{"matching date", "GET", "bytes=2-5", "Mon, 19 Oct 2026 00:00:00 GMT", 206, "2345", 2, 4, false},
		{"stale If-Range", "GET", "bytes=2-5", `"old-version"`, 200, "0123456789", 0, 10, false},
		{"weak If-Range", "GET", "bytes=2-5", `W/"version"`, 200, "0123456789", 0, 10, false},
		{"multipart ignored", "GET", "bytes=0-1,3-4", "", 200, "0123456789", 0, 10, false},
		{"unknown range unit ignored", "GET", "items=0-1", "", 200, "0123456789", 0, 10, false},
		{"unsatisfiable", "GET", "bytes=10-", "", 416, "", 0, 0, true},
		{"overflow start", "GET", "bytes=184467440737095516160000-", "", 416, "", 0, 0, true},
		{"zero suffix", "GET", "bytes=-0", "", 416, "", 0, 0, true},
		{"backward range", "GET", "bytes=5-2", "", 416, "", 0, 0, true},
		{"invalid range", "GET", "bytes=abc-def", "", 416, "", 0, 0, true},
		{"head", "HEAD", "bytes=2-5", "", 200, "", 0, 0, true},
		{"provider failure", "GET", "bytes=2-5", "", 503, "", 2, 4, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			source := &syntheticRemotePrivateVideoByteRangeSource{metadata: storage.ObjectByteRangeMetadata{Size: 10, ETag: `"version"`, ModTime: time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)}, failed: scenario.name == "provider failure"}
			router := gin.New()
			router.Handle(scenario.method, "/authorized", func(c *gin.Context) {
				c.Header("Vary", "Origin")
				StreamConversationAuthorizedMediaFile(c, &service.MediaAssetFile{ReadCloser: source, ByteRangeSource: source, MimeType: "video/mp4", Filename: "video.mp4", Size: 10})
			})
			request := httptest.NewRequestWithContext(t.Context(), scenario.method, "/authorized", nil)
			request.Header.Set("Range", scenario.byteRange)
			request.Header.Set("If-Range", scenario.ifRange)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, scenario.status, response.Code)
			require.Equal(t, scenario.expected, response.Body.String())
			require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
			require.Contains(t, response.Header().Values("Vary"), "Cookie")
			require.Contains(t, response.Header().Values("Vary"), "Origin")
			require.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
			require.Empty(t, response.Header().Get("Location"))
			require.True(t, source.closed)
			if scenario.noRead {
				require.Empty(t, source.opened)
			} else {
				require.Equal(t, [][2]int64{{scenario.start, scenario.length}}, source.opened)
			}
			if scenario.status == 416 {
				require.Equal(t, "bytes */10", response.Header().Get("Content-Range"))
			}
			if scenario.status == 206 {
				require.Equal(t, "bytes", response.Header().Get("Accept-Ranges"))
			}
		})
	}
}

func TestRemotePrivateVideoRangeConditionalRequestsEvaluateOnlyAfterAuthorizationWithoutOpeningBodies(t *testing.T) {
	for _, scenario := range []struct {
		header, value string
		status        int
	}{
		{"If-None-Match", `W/"version"`, 304}, {"If-None-Match", `"other", "version"`, 304},
		{"If-Modified-Since", "Mon, 19 Oct 2026 00:00:00 GMT", 304},
		{"If-Match", `"old-version"`, 412}, {"If-Match", `W/"version"`, 412},
		{"If-Unmodified-Since", "Sun, 18 Oct 2026 00:00:00 GMT", 412},
	} {
		t.Run(scenario.header+scenario.value, func(t *testing.T) {
			source := &syntheticRemotePrivateVideoByteRangeSource{metadata: storage.ObjectByteRangeMetadata{Size: 10, ETag: `"version"`, ModTime: time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)}}
			router := gin.New()
			router.GET("/authorized", func(c *gin.Context) {
				StreamConversationAuthorizedMediaFile(c, &service.MediaAssetFile{ReadCloser: source, ByteRangeSource: source, MimeType: "video/mp4"})
			})
			request := httptest.NewRequestWithContext(t.Context(), "GET", "/authorized", nil)
			request.Header.Set(scenario.header, scenario.value)
			request.Header.Set("Range", "bytes=2-5")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, scenario.status, response.Code)
			require.Empty(t, response.Body.String())
			require.Empty(t, source.opened)
			require.True(t, source.closed)
			require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
		})
	}
}
