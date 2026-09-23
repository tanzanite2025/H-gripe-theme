package admin

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRespondHistoricalRefundFXSnapshotErrorGivesRecoveryInstructions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	require.True(t, respondHistoricalRefundFXSnapshotError(context, service.ErrHistoricalRefundFXSnapshotMissing))
	require.Equal(t, http.StatusConflict, recorder.Code)
	require.Contains(t, recorder.Body.String(), "Historical FX Snapshot Backfill")
	require.Contains(t, recorder.Body.String(), "retry the refund")
	require.Contains(t, recorder.Body.String(), "Do not use the current exchange rate")
}

func TestRespondHistoricalRefundFXSnapshotErrorLeavesOtherErrorsAlone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	require.False(t, respondHistoricalRefundFXSnapshotError(context, errors.New("unrelated error")))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Empty(t, recorder.Body.String())
}
