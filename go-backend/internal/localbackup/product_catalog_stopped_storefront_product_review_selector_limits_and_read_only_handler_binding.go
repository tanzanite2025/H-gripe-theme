package localbackup

import (
	"errors"

	publicreview "commerce-platform/internal/api/v1/review"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func isStoppedStorefrontProductReviewOperation(operation string) bool {
	return operation == "product_reviews" || operation == "product_review_summary"
}

func validateStoppedStorefrontProductReviewSelectorLimits(request StoppedProductReadOnlyStorefrontHTTPRequest) error {
	if !isStoppedStorefrontProductReviewOperation(request.Operation) {
		if request.ProductID != 0 || request.Page != 0 || request.PageSize != 0 {
			return errors.New("review selectors are only supported for product review operations")
		}
		return nil
	}
	if request.ProductID == 0 || uint64(request.ProductID) > 4294967295 {
		return errors.New("product reviews require a positive 32-bit product identity")
	}
	if request.Operation == "product_review_summary" {
		if request.Page != 0 || request.PageSize != 0 {
			return errors.New("review summaries do not accept pagination")
		}
		return nil
	}
	if request.Page < 1 || request.Page > 1000 || request.PageSize < 1 || request.PageSize > 20 {
		return errors.New("product review checks require page 1 through 1000 and page_size 1 through 20")
	}
	return nil
}

func bindStoppedStorefrontReadOnlyProductReviewHandlers(router *gin.Engine, database *gorm.DB, media *service.MediaService) {
	reviews := service.NewReviewServiceForReadOnlyRecoveryVerification(repository.NewReviewRepository(database))
	reviews.ConfigureMediaService(media)
	handler := publicreview.NewHandler(reviews)
	router.GET("/api/v1/reviews", handler.ListProductReviews)
	router.GET("/api/v1/reviews/summary/:product_id", handler.GetReviewSummary)
}
