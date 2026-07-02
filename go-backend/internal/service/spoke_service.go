package service

import (
	"errors"
	"math"
	domainspoke "tanzanite/internal/domain/spoke"
	"tanzanite/internal/repository"
)

var (
	ErrSpokeGeometryNotFound    = errors.New("unknown rim or hub geometry")
	ErrSpokeHubGeometryMissing  = errors.New("hub geometry not available for requested position")
	ErrInvalidSpokeCalculation  = errors.New("invalid spoke calculation input")
	spokeCalculationFormulaName = "v1.0-go-backend"
)

type SpokeService struct {
	spokeRepo *repository.SpokeRepository
}

type SpokeCalculationInput struct {
	RimID         string
	HubID         string
	WheelPosition string
	SpokeCount    int
	Crossing      int
}

type SpokeCalculationResult struct {
	LeftLengthMM  float64               `json:"leftLengthMm"`
	RightLengthMM float64               `json:"rightLengthMm"`
	Debug         SpokeCalculationDebug `json:"debug"`
}

type SpokeCalculationDebug struct {
	Rim            *domainspoke.RimModel    `json:"rim"`
	Hub            *domainspoke.HubGeometry `json:"hub"`
	FormulaVersion string                   `json:"formulaVersion"`
}

func NewSpokeService(spokeRepo *repository.SpokeRepository) *SpokeService {
	return &SpokeService{spokeRepo: spokeRepo}
}

func (s *SpokeService) GetExport() domainspoke.ExportResponse {
	return domainspoke.DefaultExport()
}

func (s *SpokeService) ListHistory(search string, page, pageSize int) ([]domainspoke.History, int64, error) {
	return s.spokeRepo.ListHistory(search, page, pageSize)
}

func (s *SpokeService) Calculate(input SpokeCalculationInput) (*SpokeCalculationResult, error) {
	if input.SpokeCount <= 0 {
		return nil, ErrInvalidSpokeCalculation
	}

	export := s.GetExport()
	rim := findSpokeRim(export, input.RimID)
	hub := findSpokeHub(export, input.HubID)
	if rim == nil || hub == nil {
		return nil, ErrSpokeGeometryNotFound
	}

	hubGeo := hub.Rear
	if input.WheelPosition == "front" {
		hubGeo = hub.Front
	}
	if hubGeo == nil {
		return nil, ErrSpokeHubGeometryMissing
	}

	radius := rim.ERD / 2.0
	leftFlangeRadius := hubGeo.LeftFlangePCD / 2.0
	rightFlangeRadius := hubGeo.RightFlangePCD / 2.0
	angleRad := (720.0 * float64(input.Crossing) / float64(input.SpokeCount)) * math.Pi / 180.0

	left := math.Sqrt(radius*radius + leftFlangeRadius*leftFlangeRadius + hubGeo.LeftFlange*hubGeo.LeftFlange - 2*radius*leftFlangeRadius*math.Cos(angleRad))
	right := math.Sqrt(radius*radius + rightFlangeRadius*rightFlangeRadius + hubGeo.RightFlange*hubGeo.RightFlange - 2*radius*rightFlangeRadius*math.Cos(angleRad))

	return &SpokeCalculationResult{
		LeftLengthMM:  math.Round(left*10) / 10,
		RightLengthMM: math.Round(right*10) / 10,
		Debug: SpokeCalculationDebug{
			Rim:            rim,
			Hub:            hubGeo,
			FormulaVersion: spokeCalculationFormulaName,
		},
	}, nil
}

func findSpokeRim(export domainspoke.ExportResponse, rimID string) *domainspoke.RimModel {
	for _, brand := range export.Rims {
		for _, rim := range brand.Items {
			if rim.ID == rimID {
				foundRim := rim
				return &foundRim
			}
		}
	}
	return nil
}

func findSpokeHub(export domainspoke.ExportResponse, hubID string) *domainspoke.HubModel {
	for _, brand := range export.Hubs {
		for _, hub := range brand.Items {
			if hub.ID == hubID {
				foundHub := hub
				return &foundHub
			}
		}
	}
	return nil
}
