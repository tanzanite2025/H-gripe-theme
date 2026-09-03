package content

import "commerce-platform/internal/service"

type Handler struct {
	postService                     *service.PostService
	faqService                      *service.FAQService
	mediaService                    *service.MediaService
	refundCancellationPolicyService *service.RefundCancellationPolicyService
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

func (h *Handler) ConfigureRefundCancellationPolicyService(policyService *service.RefundCancellationPolicyService) {
	if h == nil {
		return
	}
	h.refundCancellationPolicyService = policyService
}
