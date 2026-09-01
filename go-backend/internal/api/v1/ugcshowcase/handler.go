package ugcshowcase

import (
	"context"
	"mime/multipart"

	ugcshowcasedomain "commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/service"
)

type showcaseUploadService interface {
	CountPendingSubmissions(uint) (int64, error)
	ValidateUploadOrder(context.Context, uint, uint) error
	UploadPhotos(context.Context, uint, uint, []*multipart.FileHeader, map[string]string) (*ugcshowcasedomain.UGCShowcase, error)
}

type showcaseUploadProtection interface {
	Evaluate(context.Context, service.UGCShowcaseUploadProtectionInput) (service.UGCShowcaseUploadProtectionDecision, error)
	RecordFailure(context.Context, service.UGCShowcaseUploadProtectionIdentity) error
}

type UGCShowcaseHandler struct {
	service           *service.UGCShowcaseService
	uploadService     showcaseUploadService
	uploadProtection  showcaseUploadProtection
	uploadEligibility *service.UGCShowcaseUploadEligibilityService
}

type ugcShowcaseHandler = UGCShowcaseHandler

func NewUGCShowcaseHandler(s *service.UGCShowcaseService) *UGCShowcaseHandler {
	handler := &UGCShowcaseHandler{service: s}
	if s != nil {
		handler.uploadService = s
	}
	return handler
}

func (h *UGCShowcaseHandler) ConfigureUploadProtection(protection *service.UGCShowcaseUploadProtectionService) {
	if h == nil {
		return
	}
	if protection == nil {
		h.uploadProtection = nil
		return
	}
	h.uploadProtection = protection
}

func (h *UGCShowcaseHandler) ConfigureUploadEligibility(eligibility *service.UGCShowcaseUploadEligibilityService) {
	if h != nil {
		h.uploadEligibility = eligibility
	}
}
