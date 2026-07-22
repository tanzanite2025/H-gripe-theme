package tracking

import (
	"context"
	"time"
)

// MockTrackingService 模拟追踪服务 (用于测试)
type MockTrackingService struct{}

func NewMockTrackingService() TrackingService {
	return &MockTrackingService{}
}

func (s *MockTrackingService) Track(ctx context.Context, trackingNumber, carrier string) (*TrackingInfo, error) {
	return &TrackingInfo{
		TrackingNumber: trackingNumber,
		Carrier:        carrier,
		Status:         "In Transit",
		StatusCode:     2,
		Events: []TrackingEvent{
			{
				Time:        time.Now().Add(-48 * time.Hour),
				Status:      "Picked Up",
				Description: "Package picked up by carrier",
				Location:    &Location{City: "Shanghai", Country: "China"},
			},
			{
				Time:        time.Now().Add(-24 * time.Hour),
				Status:      "In Transit",
				Description: "Package in transit",
				Location:    &Location{City: "Hong Kong", Country: "China"},
			},
		},
		UpdatedAt: time.Now(),
	}, nil
}

func (s *MockTrackingService) BatchTrack(ctx context.Context, trackings []TrackingRequest) ([]*TrackingInfo, error) {
	if err := validateBatchTrackingRequest(trackings); err != nil {
		return nil, err
	}

	results := make([]*TrackingInfo, len(trackings))
	for i, t := range trackings {
		info, _ := s.Track(ctx, t.TrackingNumber, t.Carrier)
		results[i] = info
	}
	return results, nil
}

func (s *MockTrackingService) DetectCarrier(ctx context.Context, trackingNumber string) ([]string, error) {
	return []string{"ups", "fedex", "dhl"}, nil
}

func (s *MockTrackingService) RegisterTrackings(ctx context.Context, trackings []TrackingRequest) error {
	return validateBatchTrackingRequest(trackings)
}
