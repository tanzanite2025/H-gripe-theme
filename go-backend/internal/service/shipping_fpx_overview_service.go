package service

import "time"

type FpxOverview struct {
	GeneratedAt time.Time               `json:"generated_at"`
	Channels    FpxOverviewChannelStats `json:"channels"`
}

type FpxOverviewChannelStats struct {
	Total   int64 `json:"total"`
	Enabled int64 `json:"enabled"`
}

func (s *ShippingService) GetFpxOverview() (*FpxOverview, error) {
	counts, err := s.shippingRepo.GetFpxOverviewCounts()
	if err != nil {
		return nil, err
	}
	return &FpxOverview{
		GeneratedAt: time.Now().UTC(),
		Channels: FpxOverviewChannelStats{
			Total:   counts.ChannelTotal,
			Enabled: counts.EnabledChannelTotal,
		},
	}, nil
}
