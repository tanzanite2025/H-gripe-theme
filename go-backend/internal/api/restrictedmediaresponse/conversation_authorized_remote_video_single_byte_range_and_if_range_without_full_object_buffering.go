package restrictedmediaresponse

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/pkg/storage"
	"github.com/gin-gonic/gin"
)

func parseConversationAuthorizedSingleVideoByteRange(value string, size int64) (start, length int64, partial, unsatisfiable bool) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "bytes=") || strings.Contains(value, ",") {
		return 0, size, false, false
	}
	first, last, found := strings.Cut(strings.TrimSpace(strings.TrimPrefix(value, "bytes=")), "-")
	if !found || size <= 0 {
		return 0, 0, false, true
	}
	parseDigits := func(value string) (uint64, bool, bool) {
		if value == "" {
			return 0, false, false
		}
		for _, digit := range value {
			if digit < '0' || digit > '9' {
				return 0, false, false
			}
		}
		parsed, err := strconv.ParseUint(value, 10, 64)
		return parsed, true, err != nil
	}
	if first == "" {
		suffix, valid, overflow := parseDigits(last)
		if !valid || (!overflow && suffix == 0) {
			return 0, 0, false, true
		}
		if overflow || suffix > uint64(size) {
			suffix = uint64(size)
		}
		return size - int64(suffix), int64(suffix), true, false
	}
	position, valid, overflow := parseDigits(first)
	if !valid || overflow || position >= uint64(size) {
		return 0, 0, false, true
	}
	end := uint64(size - 1)
	if last != "" {
		requestedEnd, valid, overflow := parseDigits(last)
		if !valid || (!overflow && requestedEnd < position) {
			return 0, 0, false, true
		}
		if !overflow && requestedEnd < end {
			end = requestedEnd
		}
	}
	return int64(position), int64(end - position + 1), true, false
}

func conversationAuthorizedVideoIfRangeMatches(value string, metadata storage.ObjectByteRangeMetadata) bool {
	if value == "" {
		return true
	}
	if strings.HasPrefix(value, "\"") {
		return value == metadata.ETag
	}
	date, err := http.ParseTime(value)
	return err == nil && !metadata.ModTime.IsZero() && metadata.ModTime.Truncate(time.Second).Equal(date)
}

func conversationAuthorizedVideoETagConditionContainsMatch(value, etag string, allowWeak bool) bool {
	value = strings.TrimSpace(value)
	if value == "*" {
		return true
	}
	for value != "" {
		weak := strings.HasPrefix(value, "W/")
		if weak {
			value = strings.TrimPrefix(value, "W/")
		}
		if !strings.HasPrefix(value, "\"") {
			return false
		}
		end := strings.IndexByte(value[1:], '"')
		if end < 0 {
			return false
		}
		candidate := value[:end+2]
		if candidate == etag && (allowWeak || !weak) {
			return true
		}
		value = strings.TrimSpace(value[end+2:])
		if value == "" {
			return false
		}
		if !strings.HasPrefix(value, ",") {
			return false
		}
		value = strings.TrimSpace(value[1:])
	}
	return false
}

func conversationAuthorizedVideoConditionalResponseStatus(request *http.Request, metadata storage.ObjectByteRangeMetadata) int {
	if condition := request.Header.Get("If-Match"); condition != "" {
		if !conversationAuthorizedVideoETagConditionContainsMatch(condition, metadata.ETag, false) {
			return http.StatusPreconditionFailed
		}
	} else if date, err := http.ParseTime(request.Header.Get("If-Unmodified-Since")); err == nil && !metadata.ModTime.IsZero() && metadata.ModTime.Truncate(time.Second).After(date) {
		return http.StatusPreconditionFailed
	}
	if condition := request.Header.Get("If-None-Match"); condition != "" {
		if conversationAuthorizedVideoETagConditionContainsMatch(condition, metadata.ETag, true) {
			return http.StatusNotModified
		}
	} else if date, err := http.ParseTime(request.Header.Get("If-Modified-Since")); err == nil && !metadata.ModTime.IsZero() && !metadata.ModTime.Truncate(time.Second).After(date) {
		return http.StatusNotModified
	}
	return 0
}

func streamConversationAuthorizedRemoteVideoByteRange(c *gin.Context, source storage.ObjectByteRangeSource) {
	metadata := source.Metadata()
	if metadata.Size < 0 {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	c.Header("ETag", metadata.ETag)
	if !metadata.ModTime.IsZero() {
		c.Header("Last-Modified", metadata.ModTime.UTC().Format(http.TimeFormat))
	}
	if status := conversationAuthorizedVideoConditionalResponseStatus(c.Request, metadata); status != 0 {
		c.Status(status)
		return
	}
	rangeValue := c.GetHeader("Range")
	if c.Request.Method != http.MethodGet || !conversationAuthorizedVideoIfRangeMatches(c.GetHeader("If-Range"), metadata) {
		rangeValue = ""
	}
	start, length, partial, unsatisfiable := parseConversationAuthorizedSingleVideoByteRange(rangeValue, metadata.Size)
	if unsatisfiable {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", metadata.Size))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	// Validate the provider response before committing 200/206. An ignored
	// range, changed ETag or wrong offset must never be forwarded as video bytes.
	var body io.ReadCloser
	if length > 0 && c.Request.Method != http.MethodHead {
		var err error
		body, err = source.OpenByteRange(c.Request.Context(), start, length)
		if err != nil || body == nil {
			c.Status(http.StatusServiceUnavailable)
			return
		}
		defer func() { _ = body.Close() }()
	}
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Length", strconv.FormatInt(length, 10))
	status := http.StatusOK
	if partial {
		status = http.StatusPartialContent
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, start+length-1, metadata.Size))
	}
	c.Status(status)
	if body != nil {
		if _, err := io.CopyN(c.Writer, body, length); err != nil {
			_ = c.Error(storage.ErrRemoteObjectByteRangeUnavailable)
			c.Abort()
		}
	}
}
