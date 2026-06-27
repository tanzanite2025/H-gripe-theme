package registration

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	orderdomain "tanzanite/internal/domain/order"
	domainregistration "tanzanite/internal/domain/registration"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/pkg/upload"

	"github.com/gin-gonic/gin"
)

var (
	errWarrantyEmailMismatch      = errors.New("email does not match order record")
	errWarrantyStorageUnavailable = errors.New("file storage is unavailable")

	warrantyClaimMaxRequestBytes int64 = 135 << 20
)

func (h *Handler) VerifyWarrantyOrder(c *gin.Context) {
	var req struct {
		OrderNumber string `json:"order_number" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	if _, err := h.findVerifiedWarrantyOrder(req.OrderNumber, req.Email); err != nil {
		apierror.RespondNotFound(c, "Order")
		return
	}

	response.SuccessWithMessage(c, "Order verified successfully", nil)
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

	orderNumber := strings.TrimSpace(c.PostForm("order_number"))
	email := strings.TrimSpace(c.PostForm("email"))
	if orderNumber == "" || email == "" {
		apierror.RespondBadRequest(c, "Order Number and Email are required")
		return
	}

	order, err := h.findVerifiedWarrantyOrder(orderNumber, email)
	if err != nil {
		apierror.RespondNotFound(c, "Order")
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

	imagesJSON, err := json.Marshal(imageURLs)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	claim := domainregistration.WarrantyClaim{
		UserID:       order.UserID,
		IssueType:    "warranty",
		Description:  strings.TrimSpace(c.PostForm("issue_description")),
		Images:       string(imagesJSON),
		OrderNumber:  orderNumber,
		Email:        email,
		TirePressure: strings.TrimSpace(c.PostForm("tire_pressure")),
		IsTubeless:   c.PostForm("is_tubeless") == "yes",
		VideoURL:     videoURL,
		Status:       "pending",
	}

	if err := h.registrationRepo.CreateWarrantyClaim(&claim); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, gin.H{
		"success": true,
		"message": "Claim submitted successfully",
		"id":      claim.ID,
	})
}

func (h *Handler) GetExpiringWarranties(c *gin.Context) {
	days := pagination.ParseLimit(c)
	if days > 365 {
		days = 30
	}

	registrations, err := h.registrationSvc.GetExpiringWarranties(days)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": registrations})
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

	reg, err := h.registrationRepo.FindRegistrationByID(claim.RegistrationID)
	if err != nil {
		apierror.RespondNotFound(c, "Registration")
		return
	}
	if reg.UserID != userID.(uint) {
		apierror.RespondForbidden(c)
		return
	}

	claim.Status = "pending"
	if err := h.registrationRepo.CreateWarrantyClaim(&claim); err != nil {
		apierror.RespondBadRequest(c, err.Error())
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

	claim, err := h.registrationRepo.FindWarrantyClaimByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Warranty claim")
		return
	}

	if claim.RegistrationID != 0 {
		reg, err := h.registrationRepo.FindRegistrationByID(claim.RegistrationID)
		if err != nil {
			apierror.RespondNotFound(c, "Registration")
			return
		}
		if reg.UserID != userID.(uint) {
			apierror.RespondForbidden(c)
			return
		}
	} else if claim.UserID != userID.(uint) {
		apierror.RespondForbidden(c)
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

	registrationID, err := strconv.ParseUint(c.Param("registration_id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid registration ID")
		return
	}

	reg, err := h.registrationRepo.FindRegistrationByID(uint(registrationID))
	if err != nil {
		apierror.RespondNotFound(c, "Registration")
		return
	}
	if reg.UserID != userID.(uint) {
		apierror.RespondForbidden(c)
		return
	}

	claims, err := h.registrationRepo.FindWarrantyClaimsByRegistrationID(uint(registrationID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": claims})
}

func (h *Handler) ListAllWarrantyClaims(c *gin.Context) {
	params := pagination.ParsePagination(c)
	status := c.Query("status")

	claims, total, err := h.registrationRepo.FindAllWarrantyClaims(params.Page, params.PageSize, status)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, claims, params.Page, params.PageSize, total)
}

func (h *Handler) UpdateWarrantyClaimStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	if err := h.registrationRepo.UpdateWarrantyClaimStatus(uint(id), req.Status); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Warranty claim status updated", nil)
}

func (h *Handler) findVerifiedWarrantyOrder(orderNumber, email string) (*orderdomain.Order, error) {
	orderNumber = strings.TrimSpace(orderNumber)
	email = strings.ToLower(strings.TrimSpace(email))
	order, err := h.orderRepo.FindByOrderNumberForVerification(orderNumber)
	if err != nil {
		return nil, err
	}

	shippingEmail := strings.ToLower(strings.TrimSpace(order.ShippingAddress.Email))
	billingEmail := strings.ToLower(strings.TrimSpace(order.BillingAddress.Email))
	if email == "" || (email != shippingEmail && email != billingEmail) {
		return nil, errWarrantyEmailMismatch
	}
	return order, nil
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
