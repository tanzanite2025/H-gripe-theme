package admin

import (
	"encoding/json"
	"time"

	"commerce-platform/internal/service"
	"gorm.io/datatypes"
)

type orderEvidenceItemUpdateRequest struct {
	Status     string          `json:"status" binding:"required,oneof=missing draft complete waived"`
	DataJSON   json.RawMessage `json:"data_json"`
	CapturedAt *time.Time      `json:"captured_at"`
}

func (r orderEvidenceItemUpdateRequest) toServiceInput(itemID, actorID uint) service.OrderEvidenceAdminItemUpdateInput {
	return service.OrderEvidenceAdminItemUpdateInput{
		ItemID:     itemID,
		Status:     r.Status,
		DataJSON:   datatypes.JSON(r.DataJSON),
		CapturedAt: r.CapturedAt,
		CapturedBy: actorID,
	}
}
