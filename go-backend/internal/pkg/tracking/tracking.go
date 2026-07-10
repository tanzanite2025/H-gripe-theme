package tracking

import (
	"context"
	"time"
)

// TrackingService 物流追踪服务接口
type TrackingService interface {
	Track(ctx context.Context, trackingNumber, carrier string) (*TrackingInfo, error)
	BatchTrack(ctx context.Context, trackings []TrackingRequest) ([]*TrackingInfo, error)
	DetectCarrier(ctx context.Context, trackingNumber string) ([]string, error)
}

// TrackingRequest 追踪请求
type TrackingRequest struct {
	TrackingNumber string `json:"tracking_number"`
	Carrier        string `json:"carrier,omitempty"`
}

// TrackingInfo 物流追踪信息
type TrackingInfo struct {
	TrackingNumber string          `json:"tracking_number"`
	Carrier        string          `json:"carrier"`
	Status         string          `json:"status"`
	StatusCode     int             `json:"status_code"`
	Events         []TrackingEvent `json:"events"`
	Origin         *Location       `json:"origin,omitempty"`
	Destination    *Location       `json:"destination,omitempty"`
	EstimatedTime  *time.Time      `json:"estimated_time,omitempty"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// TrackingEvent 物流事件
type TrackingEvent struct {
	Time        time.Time `json:"time"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Location    *Location `json:"location,omitempty"`
}

// Location 位置信息
type Location struct {
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
	Country string `json:"country,omitempty"`
	ZipCode string `json:"zip_code,omitempty"`
}
