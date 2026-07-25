package spoke

type ExportResponse struct {
	Rims []RimBrand `json:"rims"`
	Hubs []HubBrand `json:"hubs"`
}

type RimBrand struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Items []RimModel `json:"items"`
}

type RimModel struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	ERD    float64  `json:"erd"`
	Weight *float64 `json:"weight,omitempty"`
}

type HubBrand struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Items []HubModel `json:"items"`
}

type HubModel struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Front *HubGeometry `json:"front,omitempty"`
	Rear  *HubGeometry `json:"rear,omitempty"`
}

type HubGeometry struct {
	LeftFlange        float64  `json:"leftFlange"`
	RightFlange       float64  `json:"rightFlange"`
	LeftFlangePCD     float64  `json:"leftFlangePcd"`
	RightFlangePCD    float64  `json:"rightFlangePcd"`
	SpokeHoleDiameter *float64 `json:"spokeHoleDiameter,omitempty"`
}
