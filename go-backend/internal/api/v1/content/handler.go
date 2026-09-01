package content

import "commerce-platform/internal/service"

type Handler struct {
	postService               *service.PostService
	faqService                *service.FAQService
	mediaService              *service.MediaService
	refundReturnPolicyService *service.RefundReturnPolicyService
}

func NewHandler(postService *service.PostService, faqService *service.FAQService, mediaServices ...*service.MediaService) *Handler {
	var mediaService *service.MediaService
	if len(mediaServices) > 0 {
		mediaService = mediaServices[0]
	}
	return &Handler{
		postService:  postService,
		faqService:   faqService,
		mediaService: mediaService,
	}
}

func (h *Handler) ConfigureRefundReturnPolicyService(policyService *service.RefundReturnPolicyService) {
	if h == nil {
		return
	}
	h.refundReturnPolicyService = policyService
}
