package service

import (
	"commerce-platform/internal/domain/shipping"
)

const yanwenBatchWaybillCreationMaximumRequests = 50

type YanwenCreateWaybillInput struct {
	Environment       string `json:"environment"`
	OrderID           uint   `json:"order_id"`
	ChannelID         uint   `json:"channel_id"`
	WarehouseCode     string `json:"warehouse_code"`
	OrderSource       string `json:"order_source"`
	HasBattery        *bool  `json:"has_battery"`
	ReceiverTaxNumber string `json:"receiver_tax_number"`
	IOSS              string `json:"ioss"`
	EORI              string `json:"eori"`
}

// YanwenBatchWaybillCreationInput carries independently validated Yanwen
// create requests that should be submitted in one operator action.
type YanwenBatchWaybillCreationInput struct {
	Requests []YanwenCreateWaybillInput `json:"requests"`
}

// YanwenBatchWaybillCreationItem reports one real Yanwen create attempt. A
// failed item never fabricates a local waybill number or lifecycle status.
type YanwenBatchWaybillCreationItem struct {
	RequestIndex  int    `json:"request_index"`
	Environment   string `json:"environment"`
	OrderID       uint   `json:"order_id"`
	ChannelID     uint   `json:"channel_id"`
	WaybillID     uint   `json:"waybill_id,omitempty"`
	OrderNumber   string `json:"order_number,omitempty"`
	WaybillNumber string `json:"waybill_number,omitempty"`
	Error         string `json:"error,omitempty"`
}

// YanwenBatchWaybillCreationResult contains per-request outcomes from a
// bounded, same-environment batch push.
type YanwenBatchWaybillCreationResult struct {
	Items     []YanwenBatchWaybillCreationItem `json:"items"`
	Succeeded int                              `json:"succeeded"`
	Failed    int                              `json:"failed"`
}

// YanwenCancelWaybillInput carries the optional operator note accepted by
// express.order.cancel.
type YanwenCancelWaybillInput struct {
	Note string `json:"note"`
}

// YanwenBatchWaybillCancellationInput identifies persisted Yanwen waybills
// that should receive the same operator cancellation note.
type YanwenBatchWaybillCancellationInput struct {
	WaybillIDs []uint `json:"waybill_ids"`
	Note       string `json:"note"`
}

// YanwenBatchWaybillCancellationItem keeps one official cancellation outcome
// separate from the other items so a failed remote request is never presented
// as a successful batch operation.
type YanwenBatchWaybillCancellationItem struct {
	WaybillID            uint   `json:"waybill_id"`
	WaybillNumber        string `json:"waybill_number"`
	CancellationAccepted bool   `json:"cancellation_accepted"`
	OfficialStatusSynced bool   `json:"official_status_synced"`
	Message              string `json:"message"`
	Error                string `json:"error,omitempty"`
}

// YanwenBatchWaybillCancellationResult reports every requested Yanwen
// cancellation, including partial remote failures.
type YanwenBatchWaybillCancellationResult struct {
	Items     []YanwenBatchWaybillCancellationItem `json:"items"`
	Succeeded int                                  `json:"succeeded"`
	Failed    int                                  `json:"failed"`
}

// YanwenBatchWaybillOfficialSyncInput identifies the locally persisted
// waybills to refresh through express.order.getlist.
type YanwenBatchWaybillOfficialSyncInput struct {
	WaybillIDs []uint `json:"waybill_ids"`
}

// YanwenBatchWaybillLabelInput identifies persisted waybills whose official
// PDF labels should be included in one download archive.
type YanwenBatchWaybillLabelInput struct {
	WaybillIDs []uint `json:"waybill_ids"`
}

// YanwenWaybillLabelResult contains one official PDF label for an admin download.
type YanwenWaybillLabelResult struct {
	WaybillID     uint   `json:"waybill_id"`
	WaybillNumber string `json:"waybill_number"`
	FileName      string `json:"file_name"`
	ContentType   string `json:"content_type"`
	Base64String  string `json:"base64_string"`
}

// YanwenBatchWaybillLabelItem records one PDF generation outcome. The PDF
// bytes stay inside the archive and are never duplicated in the JSON manifest.
type YanwenBatchWaybillLabelItem struct {
	WaybillID     uint   `json:"waybill_id"`
	WaybillNumber string `json:"waybill_number"`
	FileName      string `json:"file_name"`
	Downloaded    bool   `json:"downloaded"`
	Error         string `json:"error,omitempty"`
}

// YanwenBatchWaybillLabelResult is written into the downloaded archive so
// operators can see partial official failures without treating them as hidden.
type YanwenBatchWaybillLabelResult struct {
	Items     []YanwenBatchWaybillLabelItem `json:"items"`
	Succeeded int                           `json:"succeeded"`
	Failed    int                           `json:"failed"`
}

// YanwenBatchWaybillLabelArchive is the response artifact for a bounded
// batch-label request. The archive contains successful PDFs and a manifest.
type YanwenBatchWaybillLabelArchive struct {
	FileName    string
	ContentType string
	Data        []byte
	Summary     YanwenBatchWaybillLabelResult
}

// YanwenWaybillCancellationResult reports the remote cancellation acceptance
// separately from the latest official status.
type YanwenWaybillCancellationResult struct {
	Waybill              shipping.YanwenWaybill `json:"waybill"`
	CancellationAccepted bool                   `json:"cancellation_accepted"`
	OfficialStatusSynced bool                   `json:"official_status_synced"`
	Message              string                 `json:"message"`
}

// YanwenOrderFactReader is the only generic order dependency required by the
// Yanwen waybill workflow. It returns a narrow read-only projection instead of
// exposing the complete order aggregate or the generic order repository API.
type YanwenOrderFactReader interface {
	FindYanwenOrderFactsByID(id uint) (*shipping.YanwenOrderFacts, error)
}
