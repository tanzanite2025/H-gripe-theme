package service

import (
	"errors"
	"testing"

	"tanzanite/internal/domain/review"
)

func TestNormalizeReviewSubmissionStoresPlainText(t *testing.T) {
	item := &review.Review{
		Title:   `<strong>Great</strong><script>alert("x")</script>`,
		Content: `<p onclick="alert(1)">Solid <a href="javascript:alert(1)">wheel</a>.</p>`,
		Images:  `["/uploads/review-1.jpg","https://cdn.example.com/review-2.jpg"]`,
	}

	if err := normalizeReviewSubmission(item); err != nil {
		t.Fatalf("normalizeReviewSubmission() error = %v", err)
	}
	if item.Title != "Great" {
		t.Fatalf("title = %q, want %q", item.Title, "Great")
	}
	if item.Content != "Solid wheel." {
		t.Fatalf("content = %q, want %q", item.Content, "Solid wheel.")
	}
	if item.Images != `["/uploads/review-1.jpg","https://cdn.example.com/review-2.jpg"]` {
		t.Fatalf("images = %q", item.Images)
	}
}

func TestNormalizeReviewSubmissionRejectsEmptySanitizedContent(t *testing.T) {
	err := normalizeReviewSubmission(&review.Review{
		Title:   "Title",
		Content: `<script>alert("x")</script>`,
	})
	if !errors.Is(err, ErrReviewContentRequired) {
		t.Fatalf("normalizeReviewSubmission() error = %v, want ErrReviewContentRequired", err)
	}
}

func TestNormalizeReviewSubmissionRejectsUnsafeImageURL(t *testing.T) {
	err := normalizeReviewSubmission(&review.Review{
		Title:   "Title",
		Content: "Content",
		Images:  `["javascript:alert(1)"]`,
	})
	if !errors.Is(err, ErrReviewImagesInvalid) {
		t.Fatalf("normalizeReviewSubmission() error = %v, want ErrReviewImagesInvalid", err)
	}
}

func TestReviewServiceSanitizesLegacyPublicReviewContent(t *testing.T) {
	db, reviewService := newTestReviewService(t)
	item := review.Review{
		ProductID: 1,
		UserID:    10,
		Rating:    5,
		Title:     `<strong>Legacy</strong>`,
		Content:   `<p>Trusted <img src=x onerror="alert(1)">content</p><script>alert("x")</script>`,
		Status:    "approved",
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("create review: %v", err)
	}

	publicItem, err := reviewService.GetPublicReview(item.ID)
	if err != nil {
		t.Fatalf("GetPublicReview() error = %v", err)
	}
	if publicItem.Title != "Legacy" || publicItem.Content != "Trusted content" {
		t.Fatalf("public review = %#v", publicItem)
	}
}

func TestReviewServiceDropsUnsafeLegacyImageURLs(t *testing.T) {
	db, reviewService := newTestReviewService(t)
	item := review.Review{
		ProductID: 1,
		UserID:    10,
		Rating:    5,
		Title:     "Legacy",
		Content:   "Content",
		Images:    `["javascript:alert(1)"]`,
		Status:    "approved",
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("create review: %v", err)
	}

	publicItem, err := reviewService.GetPublicReview(item.ID)
	if err != nil {
		t.Fatalf("GetPublicReview() error = %v", err)
	}
	if publicItem.Images != "[]" {
		t.Fatalf("public image list = %q, want []", publicItem.Images)
	}
}
