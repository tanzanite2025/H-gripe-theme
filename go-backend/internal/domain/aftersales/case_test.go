package aftersales

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAfterSalesCaseTransitions(t *testing.T) {
	record := &AfterSalesCase{Status: StatusRequested}

	assert.True(t, record.CanTransitionTo(StatusReviewing))
	assert.True(t, record.CanTransitionTo(StatusCancelled))
	assert.False(t, record.CanTransitionTo(StatusApproved))

	record.Status = StatusReviewing
	assert.True(t, record.CanTransitionTo(StatusApproved))
	assert.True(t, record.CanTransitionTo(StatusRejected))
	assert.False(t, record.CanTransitionTo(StatusCompleted))

	record.Status = StatusCompleted
	assert.False(t, record.CanTransitionTo(StatusReviewing))
}
