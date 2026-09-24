package admin

import (
	"errors"
	"net/http"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func respondHistoricalRefundFXSnapshotError(c *gin.Context, err error) bool {
	if !errors.Is(err, service.ErrHistoricalRefundFXSnapshotMissing) {
		return false
	}
	apierror.RespondError(
		c,
		http.StatusConflict,
		"historical_fx_snapshot_missing",
		"Refund is blocked because this non-USD order has no valid historical FX snapshot. Open the order details and enter the finance-verified order-time rate in Historical FX Snapshot Backfill, then retry the refund. Do not use the current exchange rate.",
	)
	return true
}
