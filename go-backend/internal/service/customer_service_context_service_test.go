package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/ticket"

	"github.com/stretchr/testify/assert"
)

func TestCustomerServiceReplyMetricsPairsLatestCustomerTurn(t *testing.T) {
	start := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	messages := []ticket.TicketMessage{
		{CreatedAt: start.Add(1 * time.Minute)},
		{CreatedAt: start.Add(2 * time.Minute)},
		{IsStaff: true, CreatedAt: start.Add(5 * time.Minute)},
		{CreatedAt: start.Add(10 * time.Minute)},
		{IsInternal: true, CreatedAt: start.Add(11 * time.Minute)},
		{IsStaff: true, CreatedAt: end},
	}

	totalSeconds, replyCount, unansweredTurns := customerServiceReplyMetrics(messages, start, end)

	assert.Equal(t, 180.0, totalSeconds)
	assert.Equal(t, 1, replyCount)
	assert.Equal(t, 1, unansweredTurns)
}

func TestCustomerServiceReplyMetricsIncludesPendingTurnBeforeWindow(t *testing.T) {
	start := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	messages := []ticket.TicketMessage{
		{CreatedAt: start.Add(-30 * time.Minute)},
		{IsStaff: true, CreatedAt: start.Add(30 * time.Minute)},
	}

	totalSeconds, replyCount, unansweredTurns := customerServiceReplyMetrics(messages, start, end)

	assert.Equal(t, 3600.0, totalSeconds)
	assert.Equal(t, 1, replyCount)
	assert.Equal(t, 0, unansweredTurns)
}
