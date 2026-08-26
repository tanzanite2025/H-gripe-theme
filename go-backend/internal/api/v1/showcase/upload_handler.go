package showcase

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const showcaseMaxRequestBytes = 55 << 20

const (
	showcaseMaxRegion      = 100
	showcaseMaxLocation    = 100
	showcaseMaxNickname    = 100
	showcaseMaxNotes       = 2000
	showcaseUploadAccepted = "submitted_for_review"
)

func (h *ShowcaseHandler) Upload(c *gin.Context) {
	userID, ok := showcaseAuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
		return
	}
	if h == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Upload service is temporarily unavailable"})
		return
	}
	identity := service.ShowcaseUploadProtectionIdentity{
		UserID:    userID,
		IPAddress: c.ClientIP(),
	}
	if h.uploadProtection == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Upload protection is temporarily unavailable"})
		return
	}
	uploadService := h.uploadService
	if uploadService == nil && h.service != nil {
		uploadService = h.service
	}
	if uploadService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Upload service is temporarily unavailable"})
		return
	}
	recordFailure := func() {
		_ = h.uploadProtection.RecordFailure(c.Request.Context(), identity)
	}

	if c.Request.ContentLength > showcaseMaxRequestBytes {
		recordFailure()
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error":   "Request entity too large",
			"message": "Upload request is too large",
			"code":    upload.CodeFileTooLarge,
		})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, showcaseMaxRequestBytes)

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		recordFailure()
		status := http.StatusBadRequest
		code := "tpg_invalid_form"
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			status = http.StatusRequestEntityTooLarge
			code = upload.CodeFileTooLarge
		}
		c.JSON(status, gin.H{"error": "Failed to parse form", "message": err.Error(), "code": code})
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}

	form := c.Request.MultipartForm
	files := form.File["file[]"]
	if len(files) == 0 {
		recordFailure()
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files provided", "message": "No files provided", "code": "tpg_missing_files"})
		return
	}
	if err := upload.ValidateSpecFiles(files, string(upload.SpecUserShowcaseImage)); err != nil {
		recordFailure()
		c.JSON(upload.HTTPStatus(err), gin.H{"error": err.Error(), "message": err.Error(), "code": upload.ErrorCode(err)})
		return
	}

	params, fieldCode, fieldMessage := readShowcaseUploadParams(c)
	if fieldCode != "" {
		recordFailure()
		c.JSON(http.StatusBadRequest, gin.H{"error": fieldMessage, "message": fieldMessage, "code": fieldCode})
		return
	}

	orderID, orderCode, orderMessage := readShowcaseUploadOrderID(c)
	if orderCode != "" {
		recordFailure()
		c.JSON(http.StatusBadRequest, gin.H{"error": orderMessage, "message": orderMessage, "code": orderCode})
		return
	}

	if err := uploadService.ValidateUploadOrder(c.Request.Context(), userID, orderID); err != nil {
		switch {
		case errors.Is(err, service.ErrShowcaseUploadOrderRequired):
			recordFailure()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": err.Error(), "code": "showcase_upload_order_required"})
		case errors.Is(err, service.ErrShowcaseUploadOrderNotEligible):
			recordFailure()
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "message": "A completed order belonging to your account is required.", "code": "showcase_upload_order_not_eligible"})
		case errors.Is(err, service.ErrShowcaseUploadEligibilityUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Upload eligibility is temporarily unavailable"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to validate upload order"})
		}
		return
	}

	pendingSubmissions, err := uploadService.CountPendingSubmissions(userID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Upload service is temporarily unavailable"})
		return
	}

	decision, err := h.uploadProtection.Evaluate(c.Request.Context(), service.ShowcaseUploadProtectionInput{
		Identity:           identity,
		UploadBytes:        showcaseUploadBudgetBytes(c.Request.ContentLength),
		PendingSubmissions: pendingSubmissions,
	})
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Upload protection is temporarily unavailable"})
		return
	}
	if !decision.Allowed {
		if decision.RetryAfter > 0 {
			c.Header("Retry-After", strconv.FormatInt(int64(decision.RetryAfter.Seconds()), 10))
		}
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":     "Too many uploads",
			"message":   "Upload limit exceeded. Please try again later.",
			"code":      "showcase_upload_rate_limited",
			"dimension": decision.Dimension,
		})
		return
	}

	item, err := uploadService.UploadPhotos(c.Request.Context(), userID, orderID, files, params)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrShowcaseUploadOrderRequired):
			recordFailure()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": err.Error(), "code": "showcase_upload_order_required"})
		case errors.Is(err, service.ErrShowcaseUploadOrderNotEligible):
			recordFailure()
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "message": "A completed order belonging to your account is required.", "code": "showcase_upload_order_not_eligible"})
		case errors.Is(err, service.ErrShowcaseUploadEligibilityUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Upload eligibility is temporarily unavailable"})
		case errors.Is(err, service.ErrShowcaseUploadPendingLimitExceeded):
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "Too many uploads",
				"message":   "Upload limit exceeded. Please try again later.",
				"code":      "showcase_upload_rate_limited",
				"dimension": "pending_submissions",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      item.ID,
		"status":  item.Status,
		"message": showcaseUploadAccepted,
	})
}

func showcaseAuthenticatedUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	userID, ok := value.(uint)
	return userID, ok && userID > 0
}

func readShowcaseUploadParams(c *gin.Context) (map[string]string, string, string) {
	params := map[string]string{
		"region":   strings.TrimSpace(c.PostForm("region")),
		"location": strings.TrimSpace(c.PostForm("location")),
		"nickname": strings.TrimSpace(c.PostForm("nickname")),
		"notes":    strings.TrimSpace(c.PostForm("notes")),
	}
	if params["region"] == "" {
		return nil, "tpg_missing_region", "missing_region: Region is required"
	}

	fieldLimits := map[string]int{
		"region":   showcaseMaxRegion,
		"location": showcaseMaxLocation,
		"nickname": showcaseMaxNickname,
		"notes":    showcaseMaxNotes,
	}
	for field, limit := range fieldLimits {
		if utf8.RuneCountInString(params[field]) > limit {
			message := fmt.Sprintf("field_too_long: %s must be at most %d characters", field, limit)
			return nil, "tpg_field_too_long", message
		}
	}
	return params, "", ""
}

func readShowcaseUploadOrderID(c *gin.Context) (uint, string, string) {
	rawOrderID := strings.TrimSpace(c.PostForm("order_id"))
	if rawOrderID == "" {
		return 0, "showcase_upload_order_required", "showcase_upload_order_required: A completed order is required"
	}

	orderID, err := strconv.ParseUint(rawOrderID, 10, 63)
	if err != nil || orderID == 0 || uint64(uint(orderID)) != orderID {
		return 0, "showcase_upload_order_invalid", "showcase_upload_order_invalid: Invalid order"
	}
	return uint(orderID), "", ""
}

func showcaseUploadBudgetBytes(contentLength int64) int64 {
	if contentLength <= 0 {
		return showcaseMaxRequestBytes
	}
	return contentLength
}
