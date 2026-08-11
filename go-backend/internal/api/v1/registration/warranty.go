package registration

import (
	domainregistration "commerce-platform/internal/domain/registration"
	"commerce-platform/internal/pkg/antibot"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	warrantyClaimMaxRequestBytes        int64 = 82 << 20
	warrantyClaimMaxTotalAttachmentSize int64 = 80 << 20
)

var errWarrantyStorageUnavailable = errors.New("file storage is unavailable")

func (h *Handler) VerifyWarrantyOrder(c *gin.Context) {
	var req struct {
		OrderNumber  string `json:"order_number" binding:"required"`
		Email        string `json:"email" binding:"required,email"`
		CaptchaToken string `json:"captcha_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	if !h.allowDelivery(c, req.Email, req.CaptchaToken) {
		return
	}

	if err := h.registrationSvc.RequestWarrantyOrderVerification(req.OrderNumber, req.Email); err != nil {
		if h.antiBot != nil {
			h.antiBot.RecordDeliveryResult("email", false)
		}
		apierror.RespondInternalError(c, err)
		return
	}
	if h.antiBot != nil {
		h.antiBot.RecordDeliveryResult("email", true)
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "If the order can be verified, a confirmation email has been sent.",
	})
}

func (h *Handler) VerifyWarrantyOrderToken(c *gin.Context) {
	token := strings.TrimSpace(c.Param("token"))
	if token == "" {
		apierror.RespondBadRequest(c, "verification token is required")
		return
	}

	if err := h.registrationSvc.ValidateWarrantyOrderToken(token); err != nil {
		apierror.RespondBadRequest(c, "invalid or expired verification token")
		return
	}

	response.SuccessWithMessage(c, "Warranty verification is ready", gin.H{"verified": true})
}

func (h *Handler) SubmitWarrantyClaim(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, warrantyClaimMaxRequestBytes)
	if err := c.Request.ParseMultipartForm(16 << 20); err != nil {
		status := http.StatusBadRequest
		code := apierror.ErrCodeBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			status = http.StatusRequestEntityTooLarge
			code = upload.CodeFileTooLarge
		}
		apierror.RespondError(c, status, code, err.Error())
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}

	orderNumber := strings.TrimSpace(c.PostForm("order_number"))
	email := strings.TrimSpace(c.PostForm("email"))
	verificationToken := strings.TrimSpace(c.PostForm("verification_token"))
	captchaToken := strings.TrimSpace(c.PostForm("captcha_token"))
	if orderNumber == "" || email == "" {
		apierror.RespondBadRequest(c, "Order Number and Email are required")
		return
	}
	if verificationToken == "" {
		apierror.RespondUnauthorized(c)
		return
	}
	if !h.allowChallenge(c, captchaToken) {
		return
	}

	imageURLs, videoURL, err := h.uploadWarrantyClaimFiles(c)
	if err != nil {
		status := http.StatusBadRequest
		code := apierror.ErrCodeBadRequest
		if upload.ErrorCode(err) != "invalid_upload" {
			status = upload.HTTPStatus(err)
			code = upload.ErrorCode(err)
		}
		apierror.RespondError(c, status, code, err.Error())
		return
	}

	claim, err := h.registrationSvc.CreateWarrantyClaimForOrder(service.WarrantyClaimByOrderInput{
		OrderNumber:       orderNumber,
		Email:             email,
		VerificationToken: verificationToken,
		Description:       c.PostForm("issue_description"),
		TirePressure:      c.PostForm("tire_pressure"),
		IsTubeless:        c.PostForm("is_tubeless") == "yes",
		ImageURLs:         imageURLs,
		VideoURL:          videoURL,
	})
	if err != nil {
		if errors.Is(err, service.ErrWarrantyEmailMismatch) || service.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Order")
			return
		}
		if errors.Is(err, service.ErrWarrantyVerificationRequired) {
			apierror.RespondUnauthorized(c)
			return
		}
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, gin.H{
		"success": true,
		"message": "Claim submitted successfully",
		"id":      claim.ID,
	})
}

func (h *Handler) allowDelivery(c *gin.Context, destination, challengeToken string) bool {
	if h.antiBot == nil {
		return true
	}
	err := h.antiBot.Guard(c.Request.Context(), "email", destination, c.ClientIP(), challengeToken)
	switch {
	case err == nil:
		return true
	case errors.Is(err, antibot.ErrChallengeRequired), errors.Is(err, antibot.ErrChallengeInvalid):
		apierror.RespondError(c, http.StatusForbidden, "verification_required", "Verification challenge required")
	case errors.Is(err, antibot.ErrRateLimited):
		c.Header("Retry-After", "60")
		apierror.RespondError(c, http.StatusTooManyRequests, "verification_rate_limited", "Too many verification requests")
	case errors.Is(err, antibot.ErrBudgetExceeded), errors.Is(err, antibot.ErrCircuitOpen):
		c.Header("Retry-After", "300")
		apierror.RespondError(c, http.StatusServiceUnavailable, "verification_paused", "Verification delivery is temporarily paused")
	default:
		apierror.RespondError(c, http.StatusServiceUnavailable, "verification_unavailable", "Verification service is temporarily unavailable")
	}
	return false
}

func (h *Handler) allowChallenge(c *gin.Context, challengeToken string) bool {
	if h.antiBot == nil {
		return true
	}
	err := h.antiBot.VerifyChallenge(c.Request.Context(), challengeToken, c.ClientIP())
	switch {
	case err == nil:
		return true
	case errors.Is(err, antibot.ErrChallengeRequired), errors.Is(err, antibot.ErrChallengeInvalid):
		apierror.RespondError(c, http.StatusForbidden, "verification_required", "Verification challenge required")
	default:
		apierror.RespondError(c, http.StatusServiceUnavailable, "verification_unavailable", "Verification service is temporarily unavailable")
	}
	return false
}

func (h *Handler) CreateWarrantyClaim(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	var claim domainregistration.WarrantyClaim
	if err := c.ShouldBindJSON(&claim); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	if err := h.registrationSvc.CreateWarrantyClaim(&claim, userID.(uint)); err != nil {
		respondRegistrationServiceError(c, err)
		return
	}

	response.Created(c, claim)
}

func (h *Handler) GetWarrantyClaim(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	claim, err := h.registrationSvc.GetWarrantyClaim(uint(id), userID.(uint), false)
	if err != nil {
		respondRegistrationServiceError(c, err)
		return
	}

	response.Success(c, claim)
}

func (h *Handler) ListRegistrationClaims(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	registrationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid registration ID")
		return
	}

	claims, err := h.registrationSvc.GetRegistrationClaims(uint(registrationID), userID.(uint), false)
	if err != nil {
		respondRegistrationServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"data": claims})
}

func (h *Handler) uploadWarrantyClaimFiles(c *gin.Context) ([]string, string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return []string{}, "", nil
	}

	imageFiles := make([]*multipart.FileHeader, 0, len(form.File["images[]"])+len(form.File["images"]))
	imageFiles = append(imageFiles, form.File["images[]"]...)
	imageFiles = append(imageFiles, form.File["images"]...)
	videoFiles := form.File["video"]

	if (len(imageFiles) > 0 || len(videoFiles) > 0) && h.storageService == nil {
		return nil, "", errWarrantyStorageUnavailable
	}
	if err := upload.ValidateFiles(imageFiles, upload.WarrantyImageRule); err != nil {
		return nil, "", err
	}
	if len(videoFiles) > 1 {
		return nil, "", errors.New("too_many_files: maximum 1 video allowed")
	}
	if len(videoFiles) == 1 {
		if err := upload.ValidateFile(videoFiles[0], upload.WarrantyVideoRule); err != nil {
			return nil, "", err
		}
	}
	allFiles := make([]*multipart.FileHeader, 0, len(imageFiles)+len(videoFiles))
	allFiles = append(allFiles, imageFiles...)
	allFiles = append(allFiles, videoFiles...)
	if err := upload.ValidateTotalSize(allFiles, warrantyClaimMaxTotalAttachmentSize); err != nil {
		return nil, "", err
	}

	imageURLs := make([]string, 0, len(imageFiles))
	for _, file := range imageFiles {
		url, err := h.storageService.Upload(c.Request.Context(), file)
		if err != nil {
			return nil, "", err
		}
		imageURLs = append(imageURLs, url)
	}

	videoURL := ""
	if len(videoFiles) == 1 {
		url, err := h.storageService.Upload(c.Request.Context(), videoFiles[0])
		if err != nil {
			return nil, "", err
		}
		videoURL = url
	}
	return imageURLs, videoURL, nil
}
