package showcase

import (
	"context"
	"mime/multipart"

	showcasedomain "commerce-platform/internal/domain/showcase"
	"commerce-platform/internal/service"
)

type showcaseUploadService interface {
	CountPendingSubmissions(uint) (int64, error)
	ValidateUploadOrder(context.Context, uint, uint) error
	UploadPhotos(context.Context, uint, uint, []*multipart.FileHeader, map[string]string) (*showcasedomain.Showcase, error)
}

type showcaseUploadProtection interface {
	Evaluate(context.Context, service.ShowcaseUploadProtectionInput) (service.ShowcaseUploadProtectionDecision, error)
	RecordFailure(context.Context, service.ShowcaseUploadProtectionIdentity) error
}

type ShowcaseHandler struct {
	service           *service.ShowcaseService
	uploadService     showcaseUploadService
	uploadProtection  showcaseUploadProtection
	uploadEligibility *service.ShowcaseUploadEligibilityService
}

func NewShowcaseHandler(s *service.ShowcaseService) *ShowcaseHandler {
	handler := &ShowcaseHandler{service: s}
	if s != nil {
		handler.uploadService = s
	}
	return handler
}

func (h *ShowcaseHandler) ConfigureUploadProtection(protection *service.ShowcaseUploadProtectionService) {
	if h == nil {
		return
	}
	if protection == nil {
		h.uploadProtection = nil
		return
	}
	h.uploadProtection = protection
}

func (h *ShowcaseHandler) ConfigureUploadEligibility(eligibility *service.ShowcaseUploadEligibilityService) {
	if h != nil {
		h.uploadEligibility = eligibility
	}
}
