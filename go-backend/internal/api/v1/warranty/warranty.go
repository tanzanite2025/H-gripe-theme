package warranty

import (
	"commerce-platform/internal/pkg/antibot"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/pkg/storage"
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
		OrderNumber    string `json:"order_number" binding:"required"`
		Email          string `json:"email" binding:"required,email"`
		SecondaryPhone string `json:"secondary_phone"`
		CaptchaToken   string `json:"captcha_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	if h.honeypotPolicy.ShouldDrop(req.SecondaryPhone, "warranty_verification", "secondary_phone", c.Request.URL.Path) {
		c.JSON(http.StatusAccepted, gin.H{
			"message": "If the order can be verified, a confirmation email has been sent.",
		})
		return
	}
	if !h.allowDelivery(c, req.Email, req.CaptchaToken) {
		return
	}

	if err := h.warrantySvc.RequestWarrantyOrderVerification(req.OrderNumber, req.Email); err != nil {
		apierror.RespondInternalError(c, err)
		return
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

	if err := h.warrantySvc.ValidateWarrantyOrderToken(token); err != nil {
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
	secondaryPhone := strings.TrimSpace(c.PostForm("secondary_phone"))
	captchaToken := strings.TrimSpace(c.PostForm("captcha_token"))
	if h.honeypotPolicy.ShouldDrop(secondaryPhone, "warranty_claim", "secondary_phone", c.Request.URL.Path) {
		response.Created(c, gin.H{
			"success": true,
			"message": "Claim submitted successfully",
		})
		return
	}
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

	claim, err := h.warrantySvc.CreateWarrantyClaimForOrder(service.WarrantyClaimByOrderInput{
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
		h.cleanupUploadedWarrantyFiles(c, imageURLs, videoURL)
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
	accessToken, accessTokenErr := h.warrantySvc.IssueWarrantyClaimAccessToken(claim)
	if accessTokenErr != nil {
		// The claim is already persisted. A missing access token must not turn a
		// successful submission into a client-visible failure; authenticated
		// customers can still access their claim and the token can be reissued by
		// a future verification flow.
		accessToken = ""
	}

	payload := gin.H{
		"success": true,
		"message": "Claim submitted successfully",
		"id":      claim.ID,
	}
	if accessToken != "" {
		payload["claim_access_token"] = accessToken
	}
	response.Created(c, payload)
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

func (h *Handler) GetWarrantyClaim(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	userID, _ := warrantyUserID(c)
	userEmail, _ := c.Get("email")
	viewerEmail, _ := userEmail.(string)
	accessToken := strings.TrimSpace(c.GetHeader("X-Warranty-Claim-Token"))
	if accessToken == "" {
		accessToken = strings.TrimSpace(c.Query("access_token"))
	}
	if userID == 0 && accessToken == "" {
		apierror.RespondUnauthorized(c)
		return
	}

	claim, err := h.warrantySvc.GetWarrantyClaimForViewer(uint(id), userID, viewerEmail, accessToken, false)
	if err != nil {
		respondWarrantyServiceError(c, err)
		return
	}

	response.Success(c, publicWarrantyClaimFromDomain(*claim, h.mediaResolver))
}

func respondWarrantyServiceError(c *gin.Context, err error) {
	switch {
	case service.IsRecordNotFound(err):
		apierror.RespondNotFound(c, "Resource")
	case err != nil && err.Error() == "unauthorized":
		apierror.RespondForbidden(c)
	default:
		apierror.RespondInternalError(c, err)
	}
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
	privateUploader, hasPrivateStorage := h.storageService.(storage.PrivateObjectUploader)
	if (len(imageFiles) > 0 || len(videoFiles) > 0) && !hasPrivateStorage {
		return nil, "", errWarrantyStorageUnavailable
	}
	if err := upload.ValidateSpecFiles(imageFiles, string(upload.SpecWarrantyEvidence)); err != nil {
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
		url, err := privateUploader.UploadWithPrefixPrivate(c.Request.Context(), file, "warranty")
		if err != nil {
			h.cleanupUploadedWarrantyFiles(c, imageURLs, "")
			return nil, "", err
		}
		imageURLs = append(imageURLs, url)
	}

	videoURL := ""
	if len(videoFiles) == 1 {
		url, err := privateUploader.UploadWithPrefixPrivate(c.Request.Context(), videoFiles[0], "warranty")
		if err != nil {
			h.cleanupUploadedWarrantyFiles(c, imageURLs, "")
			return nil, "", err
		}
		videoURL = url
	}
	return imageURLs, videoURL, nil
}

func (h *Handler) cleanupUploadedWarrantyFiles(c *gin.Context, imageURLs []string, videoURL string) {
	if h == nil || h.storageService == nil {
		return
	}
	for _, reference := range append(append([]string{}, imageURLs...), videoURL) {
		if strings.TrimSpace(reference) == "" {
			continue
		}
		_ = h.storageService.Delete(c.Request.Context(), reference)
	}
}
