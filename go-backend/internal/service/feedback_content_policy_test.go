package service

import (
	"errors"
	"strings"
	"testing"

	"commerce-platform/internal/domain/feedback"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeedbackServiceCreateNormalizesPublicContent(t *testing.T) {
	db, feedbackService := newTestFeedbackService(t)
	item := &feedback.Feedback{
		ThreadKey: "product:1",
		UserID:    1,
		Name:      `<strong>Buyer</strong>`,
		Content:   `<p onclick="alert(1)">Works <a href="javascript:alert(1)">well</a>.</p><script>alert("x")</script>`,
	}

	require.NoError(t, feedbackService.Create(item))
	assert.Equal(t, "Buyer", item.Name)
	assert.Equal(t, "Works well.", item.Content)
	assert.Equal(t, "pending", item.Status)

	var saved feedback.Feedback
	require.NoError(t, db.First(&saved, item.ID).Error)
	assert.Equal(t, "Works well.", saved.Content)
}

func TestFeedbackServiceListPublicSanitizesLegacyContent(t *testing.T) {
	db, feedbackService := newTestFeedbackService(t)
	require.NoError(t, db.Create(&feedback.Feedback{
		ThreadKey: "product:1",
		UserID:    1,
		Name:      `<strong>Buyer</strong>`,
		Content:   `<p>Old <a href="javascript:alert(1)">feedback</a></p><script>alert("x")</script>`,
		Status:    "approved",
	}).Error)

	items, total, err := feedbackService.ListPublic("product:1", "", 1, 20)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, "Buyer", items[0].Name)
	assert.Equal(t, "Old feedback", items[0].Content)
}

func TestFeedbackServiceCreateRejectsInvalidThreadKey(t *testing.T) {
	_, feedbackService := newTestFeedbackService(t)
	item := &feedback.Feedback{
		ThreadKey: "support payment<script>",
		UserID:    1,
		Content:   "Works well.",
	}

	err := feedbackService.Create(item)
	if !errors.Is(err, ErrFeedbackInvalidThread) {
		t.Fatalf("Create() error = %v, want ErrFeedbackInvalidThread", err)
	}
}

func TestFeedbackServiceCreateBoundsFeedbackMetadata(t *testing.T) {
	_, feedbackService := newTestFeedbackService(t)
	item := &feedback.Feedback{
		ThreadKey: "support-payment",
		UserID:    1,
		Name:      "Buyer",
		Content:   "Works well.",
		PagePath:  "/" + strings.Repeat("a", maxFeedbackPagePathRunes+1),
	}

	err := feedbackService.Create(item)
	if !errors.Is(err, ErrFeedbackPagePathTooLong) {
		t.Fatalf("Create() error = %v, want ErrFeedbackPagePathTooLong", err)
	}
}

func TestFeedbackServiceCreateNormalizesPageMetadata(t *testing.T) {
	_, feedbackService := newTestFeedbackService(t)
	item := &feedback.Feedback{
		ThreadKey: "support-payment",
		UserID:    1,
		Content:   "Works well.",
		Locale:    "zh_cn",
		PagePath:  `<p>/support/payment</p><script>alert(1)</script>`,
		PageTitle: `<strong>Payment FAQ</strong><script>alert(1)</script>`,
	}

	require.NoError(t, feedbackService.Create(item))
	assert.Equal(t, "/support/payment", item.PagePath)
	assert.Equal(t, "Payment FAQ", item.PageTitle)
}
