package spoke

func DefaultOptions() CatalogOptions {
	return CatalogOptions{
		SpokeCounts: []IntOption{
			{Value: 16, Label: "16"},
			{Value: 18, Label: "18"},
			{Value: 20, Label: "20"},
			{Value: 24, Label: "24"},
			{Value: 28, Label: "28"},
			{Value: 32, Label: "32"},
			{Value: 36, Label: "36"},
		},
		Crossings: []IntOption{
			{Value: 0, Label: "0-cross (Radial)"},
			{Value: 1, Label: "1-cross"},
			{Value: 2, Label: "2-cross"},
			{Value: 3, Label: "3-cross"},
			{Value: 4, Label: "4-cross"},
		},
		NippleTypes: []StringOption{
			{Value: "standard", Label: "Standard external"},
			{Value: "hidden", Label: "Hidden / aero"},
		},
		WheelPositions: []StringOption{
			{Value: "auto", Label: "Auto"},
			{Value: "front", Label: "Front"},
			{Value: "rear", Label: "Rear"},
		},
	}
}

func DefaultExport() ExportResponse {
	return ExportResponse{
		Options: DefaultOptions(),
		Rims: []RimBrand{
			{
				ID:   "dt_swiss",
				Name: "DT Swiss",
				Items: []RimModel{
					{ID: "rr411_db", Name: "RR 411 db", ERD: floatPtr(598)},
					{ID: "rr511_db", Name: "RR 511 db", ERD: floatPtr(581)},
					{ID: "rr421_db", Name: "RR 421 db", ERD: floatPtr(594)},
					{ID: "r460_db", Name: "R 460 db", ERD: floatPtr(596)},
					{ID: "gr531_db", Name: "GR 531 db", ERD: floatPtr(597)},
					{ID: "g540_db", Name: "G 540 db", ERD: floatPtr(592)},
				},
			},
			{
				ID:   "mavic",
				Name: "Mavic",
				Items: []RimModel{
					{ID: "open_pro_ust_disc", Name: "Open Pro UST Disc", ERD: floatPtr(589)},
					{ID: "open_pro_ust", Name: "Open Pro UST", ERD: floatPtr(589)},
					{ID: "a_1028", Name: "A 1028", ERD: floatPtr(614)},
				},
			},
			{
				ID:   "kinlin",
				Name: "Kinlin",
				Items: []RimModel{
					{ID: "xr26t", Name: "XR-26T", ERD: floatPtr(592)},
					{ID: "xr31t", Name: "XR-31T", ERD: floatPtr(580)},
				},
			},
		},
		Hubs: []HubBrand{
			{
				ID:   "dt_swiss",
				Name: "DT Swiss",
				Items: []HubModel{
					{ID: "180_road_db_cl", Name: "180 Road db CL", Front: hubGeometry(22.5, 35.6, 44, 42), Rear: hubGeometry(33, 20.2, 46, 46)},
					{ID: "240_road_db_cl", Name: "240 EXP Road db CL", Front: hubGeometry(22.5, 35.6, 44, 42), Rear: hubGeometry(33, 20.2, 46, 46)},
					{ID: "350_road_db_cl", Name: "350 Road db CL", Front: hubGeometry(22.5, 35.6, 44, 42), Rear: hubGeometry(33, 20.2, 46, 46)},
					{ID: "350_classic_db_is", Name: "350 Classic db IS (6-bolt)", Front: hubGeometry(22.5, 35.6, 58, 45), Rear: hubGeometry(35.5, 21.2, 58, 52)},
				},
			},
			{
				ID:   "shimano",
				Name: "Shimano",
				Items: []HubModel{
					{ID: "hb_r7070", Name: "105 HB-R7070", Front: hubGeometry(22, 35.6, 44, 44), Rear: hubGeometry(36.5, 21.6, 45, 45)},
				},
			},
			{
				ID:   "novatec",
				Name: "Novatec",
				Items: []HubModel{
					{ID: "d791sb_d792sb", Name: "D791SB / D792SB", Front: hubGeometry(27, 32, 58, 45), Rear: hubGeometry(35, 21, 58, 49)},
				},
			},
		},
		Presets: []WheelBuildPreset{
			{
				ID:           "tz_ar45_dt350_fr",
				Name:         "Tanzanite AR 45 Disc + DT Swiss 350",
				Description:  "Popular all-rounder build. Reliable ratchet hub with aero rim.",
				Keywords:     []string{"350", "240", "dt swiss", "45mm", "ar45", "disc", "road"},
				RimBrandID:   "dt_swiss",
				RimModelID:   "rr411_db",
				HubBrandID:   "dt_swiss",
				HubModelID:   "350_road_db_cl",
				SpokeCount:   24,
				Crossing:     2,
				NippleType:   "standard",
				NippleLength: floatPtr(14),
			},
			{
				ID:           "tz_ar50_dt240_fr",
				Name:         "Tanzanite AR 50 Disc + DT Swiss 240 EXP",
				Description:  "Lightweight racing build. Top-tier hub performance.",
				Keywords:     []string{"240", "dt swiss", "50mm", "ar50", "exp", "racing"},
				RimBrandID:   "dt_swiss",
				RimModelID:   "rr511_db",
				HubBrandID:   "dt_swiss",
				HubModelID:   "240_road_db_cl",
				SpokeCount:   24,
				Crossing:     2,
				NippleType:   "hidden",
				NippleLength: floatPtr(12),
			},
			{
				ID:           "mavic_open_dt350",
				Name:         "Mavic Open Pro UST + DT Swiss 350",
				Description:  "Classic training wheelset. Bombproof reliability.",
				Keywords:     []string{"mavic", "open pro", "350", "training"},
				RimBrandID:   "mavic",
				RimModelID:   "open_pro_ust_disc",
				HubBrandID:   "dt_swiss",
				HubModelID:   "350_road_db_cl",
				SpokeCount:   28,
				Crossing:     3,
				NippleType:   "standard",
				NippleLength: floatPtr(12),
			},
		},
	}
}

func floatPtr(value float64) *float64 {
	return &value
}

func hubGeometry(leftFlange, rightFlange, leftFlangePCD, rightFlangePCD float64) *HubGeometry {
	return &HubGeometry{
		LeftFlange:     floatPtr(leftFlange),
		RightFlange:    floatPtr(rightFlange),
		LeftFlangePCD:  floatPtr(leftFlangePCD),
		RightFlangePCD: floatPtr(rightFlangePCD),
	}
}
