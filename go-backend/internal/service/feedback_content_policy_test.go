package service

import (
	"testing"

	"tanzanite/internal/domain/feedback"

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
