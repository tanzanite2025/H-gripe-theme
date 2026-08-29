package spoke

type ExportResponse struct {
	Options CatalogOptions     `json:"options"`
	Rims    []RimBrand         `json:"rims"`
	Hubs    []HubBrand         `json:"hubs"`
	Presets []WheelBuildPreset `json:"presets"`
}

type CatalogOptions struct {
	SpokeCounts    []IntOption    `json:"spokeCounts"`
	Crossings      []IntOption    `json:"crossings"`
	NippleTypes    []StringOption `json:"nippleTypes"`
	WheelPositions []StringOption `json:"wheelPositions"`
}

type IntOption struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

type StringOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type RimBrand struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Items []RimModel `json:"items"`
}

type RimModel struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	ERD    *float64 `json:"erd"`
	Weight *float64 `json:"weight,omitempty"`
}

type HubBrand struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Items []HubModel `json:"items"`
}

type HubModel struct {
	ID                        string       `json:"id"`
	Name                      string       `json:"name"`
	FitmentHubSpecificationID *uint        `json:"-"`
	Front                     *HubGeometry `json:"front,omitempty"`
	Rear                      *HubGeometry `json:"rear,omitempty"`
}

type HubGeometry struct {
	LeftFlange        *float64 `json:"leftFlange"`
	RightFlange       *float64 `json:"rightFlange"`
	LeftFlangePCD     *float64 `json:"leftFlangePcd"`
	RightFlangePCD    *float64 `json:"rightFlangePcd"`
	SpokeHoleDiameter *float64 `json:"spokeHoleDiameter,omitempty"`
}

type WheelBuildPreset struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Keywords      []string                 `json:"keywords"`
	Description   string                   `json:"description,omitempty"`
	RimBrandID    string                   `json:"rimBrandId"`
	RimModelID    string                   `json:"rimModelId"`
	HubBrandID    string                   `json:"hubBrandId"`
	HubModelID    string                   `json:"hubModelId"`
	WheelPosition string                   `json:"wheelPosition,omitempty"`
	SpokeCount    int                      `json:"spokeCount"`
	Crossing      int                      `json:"crossing"`
	NippleType    string                   `json:"nippleType"`
	NippleLength  *float64                 `json:"nippleLength"`
	ActualLengths *WheelBuildActualLengths `json:"actualLengths,omitempty"`
}

type WheelBuildActualLengths struct {
	FrontLeft  *float64 `json:"frontLeft"`
	FrontRight *float64 `json:"frontRight"`
	RearLeft   *float64 `json:"rearLeft"`
	RearRight  *float64 `json:"rearRight"`
	Notes      string   `json:"notes,omitempty"`
}
