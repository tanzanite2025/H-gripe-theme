package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderDisputeContactEmailEventKeyIsStableAndContentBound(t *testing.T) {
	first := orderDisputeContactEmailEventKey(12, "stripe", 34, "buyer@example.com", "Subject", "Body")
	retry := orderDisputeContactEmailEventKey(12, " stripe ", 34, " buyer@example.com ", " Subject ", " Body ")
	changedBody := orderDisputeContactEmailEventKey(12, "stripe", 34, "buyer@example.com", "Subject", "Changed")

	require.Equal(t, first, retry)
	require.NotEqual(t, first, changedBody)
	require.True(t, strings.HasPrefix(first, "order.dispute_contact_email:12:stripe:34:"))
	require.LessOrEqual(t, len(first), 160)
}
