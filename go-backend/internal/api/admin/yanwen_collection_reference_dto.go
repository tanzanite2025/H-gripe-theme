package admin

import "commerce-platform/internal/service"

// YanwenPublishedCollectionReferenceDTO is the narrow cross-domain contract
// consumed by shipping template and carrier-service management. Yanwen-only
// customs rules, package limits, volumetric settings, and operator notes stay
// inside the Yanwen domain API.
type YanwenPublishedCollectionReferenceDTO struct {
	ID          uint   `json:"id"`
	ProductCode string `json:"product_code"`
	DisplayName string `json:"display_name"`
	Countries   string `json:"countries"`
	Enabled     bool   `json:"enabled"`
}

func newYanwenPublishedCollectionReferenceDTO(
	reference service.YanwenPublishedCollectionReference,
) YanwenPublishedCollectionReferenceDTO {
	return YanwenPublishedCollectionReferenceDTO{
		ID:          reference.ID,
		ProductCode: reference.ProductCode,
		DisplayName: reference.DisplayName,
		Countries:   reference.Countries,
		Enabled:     reference.Enabled,
	}
}
