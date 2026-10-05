package shipping

import (
	"errors"
	"strings"
	"time"

	"gorm.io/datatypes"
)

const YanwenWaybillStatusCreated = "created"

const (
	YanwenOfficialWaybillStatusCreated           = 0
	YanwenOfficialWaybillStatusConfirmed         = 1
	YanwenOfficialWaybillStatusReceived          = 2
	YanwenOfficialWaybillStatusInTransit         = 3
	YanwenOfficialWaybillStatusDelivered         = 4
	YanwenOfficialWaybillStatusCancelled         = 5
	YanwenOfficialWaybillStatusIntercepted       = 6
	YanwenOfficialWaybillStatusDeliveryFailed    = 7
	YanwenOfficialWaybillStatusWarehouseError    = 8
	YanwenOfficialWaybillStatusWarehouseReturned = 9
	YanwenOfficialWaybillStatusReturnReceived    = 10
	YanwenOfficialWaybillStatusTransferError     = 11
	YanwenOfficialWaybillStatusOutForDelivery    = 12
	YanwenOfficialWaybillStatusAwaitingPickup    = 13
	YanwenOfficialWaybillStatusTrackingClosed    = 15
)

func IsKnownYanwenOfficialWaybillStatus(status int) bool {
	switch status {
	case YanwenOfficialWaybillStatusCreated,
		YanwenOfficialWaybillStatusConfirmed,
		YanwenOfficialWaybillStatusReceived,
		YanwenOfficialWaybillStatusInTransit,
		YanwenOfficialWaybillStatusDelivered,
		YanwenOfficialWaybillStatusCancelled,
		YanwenOfficialWaybillStatusIntercepted,
		YanwenOfficialWaybillStatusDeliveryFailed,
		YanwenOfficialWaybillStatusWarehouseError,
		YanwenOfficialWaybillStatusWarehouseReturned,
		YanwenOfficialWaybillStatusReturnReceived,
		YanwenOfficialWaybillStatusTransferError,
		YanwenOfficialWaybillStatusOutForDelivery,
		YanwenOfficialWaybillStatusAwaitingPickup,
		YanwenOfficialWaybillStatusTrackingClosed:
		return true
	default:
		return false
	}
}

// YanwenOfficialWaybillDetails contains the fields returned by
// express.order.get that are safe to persist on an existing local waybill.
type YanwenOfficialWaybillDetails struct {
	ReferenceNumber   string
	YanwenOrderNumber string
	OfficialStatus    int
	IsPrinted         bool
	SyncedAt          time.Time
	ResponseData      datatypes.JSON
}

// YanwenWaybill stores the successful result of a real express.order.create
// call together with its request snapshot and the most recent official
// response snapshot. A record is created only after Yanwen returns a waybill number.
type YanwenWaybill struct {
	ID                   uint           `gorm:"primarykey" json:"id"`
	Environment          string         `gorm:"size:16;not null;index;uniqueIndex:idx_shipping_yanwen_waybill_environment_order_product,priority:1" json:"environment"`
	OrderID              uint           `gorm:"not null;index;uniqueIndex:idx_shipping_yanwen_waybill_environment_order_product,priority:2" json:"order_id"`
	OrderNumber          string         `gorm:"size:80;not null;index" json:"order_number"`
	ProductCode          string         `gorm:"size:80;not null;uniqueIndex:idx_shipping_yanwen_waybill_environment_order_product,priority:3" json:"product_code"`
	ChannelName          string         `gorm:"size:160;not null;default:''" json:"channel_name"`
	WarehouseCode        string         `gorm:"size:80;not null" json:"warehouse_code"`
	DestinationCountry   string         `gorm:"size:16;not null" json:"destination_country"`
	ConsigneeName        string         `gorm:"size:200;not null" json:"consignee_name"`
	DeclaredDescription  string         `gorm:"size:255;not null" json:"declared_description"`
	TotalQuantity        int            `gorm:"not null" json:"total_quantity"`
	TotalWeightGrams     int            `gorm:"not null" json:"total_weight_grams"`
	WaybillNumber        string         `gorm:"size:100;not null;uniqueIndex" json:"waybill_number"`
	YanwenOrderNumber    string         `gorm:"size:100;not null;default:''" json:"yanwen_order_number"`
	ReferenceNumber      string         `gorm:"size:100;not null;default:''" json:"reference_number"`
	Status               string         `gorm:"size:32;not null;default:'created';index" json:"status"`
	OfficialStatus       int            `gorm:"column:official_status;not null;default:0;index" json:"official_status"`
	IsPrinted            bool           `gorm:"column:is_printed;not null;default:false" json:"is_printed"`
	LastOfficialSyncedAt *time.Time     `gorm:"column:last_official_synced_at" json:"last_official_synced_at,omitempty"`
	RequestData          datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"-"`
	ResponseData         datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"-"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

func (YanwenWaybill) TableName() string { return "shipping_yanwen_waybills" }

func (w *YanwenWaybill) Validate() error {
	if w == nil {
		return errors.New("Yanwen waybill is required")
	}
	w.Environment = strings.ToLower(strings.TrimSpace(w.Environment))
	w.OrderNumber = strings.TrimSpace(w.OrderNumber)
	w.ProductCode = strings.TrimSpace(w.ProductCode)
	w.ChannelName = strings.TrimSpace(w.ChannelName)
	w.WarehouseCode = strings.TrimSpace(w.WarehouseCode)
	w.DestinationCountry = strings.ToUpper(strings.TrimSpace(w.DestinationCountry))
	w.ConsigneeName = strings.TrimSpace(w.ConsigneeName)
	w.DeclaredDescription = strings.TrimSpace(w.DeclaredDescription)
	w.WaybillNumber = strings.TrimSpace(w.WaybillNumber)
	w.YanwenOrderNumber = strings.TrimSpace(w.YanwenOrderNumber)
	w.ReferenceNumber = strings.TrimSpace(w.ReferenceNumber)
	w.Status = strings.ToLower(strings.TrimSpace(w.Status))
	if w.Environment != "fat" && w.Environment != "production" {
		return errors.New("Yanwen waybill environment must be fat or production")
	}
	if w.OrderID == 0 || w.OrderNumber == "" {
		return errors.New("Yanwen waybill order is required")
	}
	if w.ProductCode == "" || w.WarehouseCode == "" {
		return errors.New("Yanwen waybill product and warehouse are required")
	}
	if w.DestinationCountry == "" || w.ConsigneeName == "" || w.DeclaredDescription == "" {
		return errors.New("Yanwen waybill consignee and declaration are required")
	}
	if w.TotalQuantity <= 0 || w.TotalWeightGrams <= 0 {
		return errors.New("Yanwen waybill quantity and weight must be positive")
	}
	if w.WaybillNumber == "" {
		return errors.New("Yanwen waybill number is required")
	}
	if w.Status == "" {
		w.Status = YanwenWaybillStatusCreated
	}
	if !IsKnownYanwenOfficialWaybillStatus(w.OfficialStatus) {
		return errors.New("Yanwen official waybill status is invalid")
	}
	if len(w.RequestData) == 0 {
		w.RequestData = datatypes.JSON([]byte("{}"))
	}
	if len(w.ResponseData) == 0 {
		w.ResponseData = datatypes.JSON([]byte("{}"))
	}
	return nil
}
