package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	wheelsetlacingdomain "commerce-platform/internal/domain/wheelsetlacing"
)

const WheelsetLacingContractVersion = "v1.2"

// WheelsetLacingService is a deterministic, read-only service for the
// independent wheelset lacing reference page. It has no repository, product,
// spoke-calculator, or physical-mechanics dependency.
type WheelsetLacingService struct {
	catalog *wheelsetlacingdomain.Catalog
	etag    string
}

func NewWheelsetLacingService() *WheelsetLacingService {
	return NewWheelsetLacingServiceWithCatalog(wheelsetlacingdomain.NewDefaultCatalog())
}

func NewWheelsetLacingServiceWithCatalog(catalog *wheelsetlacingdomain.Catalog) *WheelsetLacingService {
	if catalog == nil {
		catalog = wheelsetlacingdomain.NewDefaultCatalog()
	}
	payload, err := json.Marshal(catalog.List())
	if err != nil {
		// Topology contains only JSON-safe scalar values and slices. Reaching this
		// branch would mean the checked-in contract changed in an invalid way.
		panic(fmt.Errorf("wheelset lacing catalog cannot be serialized: %w", err))
	}
	digest := sha256.Sum256(payload)
	return &WheelsetLacingService{
		catalog: catalog,
		etag:    `"` + hex.EncodeToString(digest[:]) + `"`,
	}
}

func (s *WheelsetLacingService) List() []wheelsetlacingdomain.Topology {
	if s == nil || s.catalog == nil {
		return nil
	}
	return s.catalog.List()
}

func (s *WheelsetLacingService) Validate(request wheelsetlacingdomain.ValidateRequest) (wheelsetlacingdomain.Topology, error) {
	if s == nil || s.catalog == nil {
		return wheelsetlacingdomain.Topology{}, fmt.Errorf("%w: service is nil", wheelsetlacingdomain.ErrInvalidTopology)
	}
	return s.catalog.Validate(request)
}

func (s *WheelsetLacingService) ETag() string {
	if s == nil {
		return ""
	}
	return s.etag
}
