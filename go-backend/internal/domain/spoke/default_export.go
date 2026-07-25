package spoke

func DefaultExport() ExportResponse {
	return ExportResponse{
		Rims: []RimBrand{
			{
				ID:   "dt_swiss",
				Name: "DT Swiss",
				Items: []RimModel{
					{ID: "rr411_db", Name: "RR 411 db", ERD: 598},
					{ID: "rr511_db", Name: "RR 511 db", ERD: 581},
					{ID: "rr421_db", Name: "RR 421 db", ERD: 594},
					{ID: "r460_db", Name: "R 460 db", ERD: 596},
					{ID: "gr531_db", Name: "GR 531 db", ERD: 597},
					{ID: "g540_db", Name: "G 540 db", ERD: 592},
				},
			},
			{
				ID:   "mavic",
				Name: "Mavic",
				Items: []RimModel{
					{ID: "open_pro_ust_disc", Name: "Open Pro UST Disc", ERD: 589},
					{ID: "open_pro_ust", Name: "Open Pro UST", ERD: 589},
					{ID: "a_1028", Name: "A 1028", ERD: 614},
				},
			},
			{
				ID:   "kinlin",
				Name: "Kinlin",
				Items: []RimModel{
					{ID: "xr26t", Name: "XR-26T", ERD: 592},
					{ID: "xr31t", Name: "XR-31T", ERD: 580},
				},
			},
		},
		Hubs: []HubBrand{
			{
				ID:   "dt_swiss",
				Name: "DT Swiss",
				Items: []HubModel{
					{ID: "180_road_db_cl", Name: "180 Road db CL", Front: &HubGeometry{LeftFlange: 22.5, RightFlange: 35.6, LeftFlangePCD: 44, RightFlangePCD: 42}, Rear: &HubGeometry{LeftFlange: 33, RightFlange: 20.2, LeftFlangePCD: 46, RightFlangePCD: 46}},
					{ID: "240_road_db_cl", Name: "240 EXP Road db CL", Front: &HubGeometry{LeftFlange: 22.5, RightFlange: 35.6, LeftFlangePCD: 44, RightFlangePCD: 42}, Rear: &HubGeometry{LeftFlange: 33, RightFlange: 20.2, LeftFlangePCD: 46, RightFlangePCD: 46}},
					{ID: "350_road_db_cl", Name: "350 Road db CL", Front: &HubGeometry{LeftFlange: 22.5, RightFlange: 35.6, LeftFlangePCD: 44, RightFlangePCD: 42}, Rear: &HubGeometry{LeftFlange: 33, RightFlange: 20.2, LeftFlangePCD: 46, RightFlangePCD: 46}},
					{ID: "350_classic_db_is", Name: "350 Classic db IS (6-bolt)", Front: &HubGeometry{LeftFlange: 22.5, RightFlange: 35.6, LeftFlangePCD: 58, RightFlangePCD: 45}, Rear: &HubGeometry{LeftFlange: 35.5, RightFlange: 21.2, LeftFlangePCD: 58, RightFlangePCD: 52}},
				},
			},
			{
				ID:   "shimano",
				Name: "Shimano",
				Items: []HubModel{
					{ID: "hb_r7070", Name: "105 HB-R7070", Front: &HubGeometry{LeftFlange: 22, RightFlange: 35.6, LeftFlangePCD: 44, RightFlangePCD: 44}, Rear: &HubGeometry{LeftFlange: 36.5, RightFlange: 21.6, LeftFlangePCD: 45, RightFlangePCD: 45}},
				},
			},
			{
				ID:   "novatec",
				Name: "Novatec",
				Items: []HubModel{
					{ID: "d791sb_d792sb", Name: "D791SB / D792SB", Front: &HubGeometry{LeftFlange: 27, RightFlange: 32, LeftFlangePCD: 58, RightFlangePCD: 45}, Rear: &HubGeometry{LeftFlange: 35, RightFlange: 21, LeftFlangePCD: 58, RightFlangePCD: 49}},
				},
			},
		},
	}
}
