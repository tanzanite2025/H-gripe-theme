package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderEvidencePackageAssemblerIncludesAttachmentHashAndCurrentTracking(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&shipping.TrackingProviderConfig{},
		&shipping.TrackingShipment{},
		&shipping.TrackingEvent{},
	))

	snapshot := servicePlanSnapshot(t, db, false)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	pkg, err := NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)

	items, err := evidenceRepo.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	identity := findEvidenceItemByType(items, orderevidence.EvidenceItemTypeProductIdentity)
	require.NotZero(t, identity.ID)

	const attachmentSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	_, err = NewOrderEvidenceAttachmentService().RegisterReference(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		OrderEvidenceAttachmentReferenceInput{
			EvidenceItemID:   identity.ID,
			StorageKey:       "order-evidence/1/product-identity.jpg",
			OriginalFilename: "product-identity.jpg",
			MimeType:         "image/jpeg",
			SizeBytes:        123,
			SHA256:           attachmentSHA256,
			UploadedBy:       7,
		},
	)
	require.NoError(t, err)

	provider := &shipping.TrackingProviderConfig{
		ProviderCode: "mock",
		ProviderName: "Mock Tracking",
		APIKey:       "must-not-leak",
		Enabled:      true,
	}
	require.NoError(t, db.Create(provider).Error)
	shipment := &shipping.TrackingShipment{
		OrderID:             snapshot.OrderID,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "CURRENT-TRACKING",
		ProviderCarrierCode: "mock-carrier",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		Enabled:             true,
	}
	require.NoError(t, db.Create(shipment).Error)

	now := time.Now().UTC()
	require.NoError(t, db.Create(&shipping.TrackingEvent{
		OrderID:             snapshot.OrderID,
		TrackingNumber:      "OLD-TRACKING",
		ProviderCarrierCode: "mock-carrier",
		Status:              "delivered",
		Description:         "old shipment",
		EventTime:           now.Add(-2 * time.Hour),
	}).Error)
	require.NoError(t, db.Create(&shipping.TrackingEvent{
		OrderID:             snapshot.OrderID,
		TrackingNumber:      shipment.TrackingNumber,
		ProviderCarrierCode: shipment.ProviderCarrierCode,
		Status:              "delivered",
		Description:         "current shipment",
		ProofOfDeliveryURL:  "https://carrier.example.test/current-pod.pdf",
		EventTime:           now.Add(-time.Hour),
	}).Error)

	assembly, err := NewOrderEvidencePackageAssembler(
		repository.NewOrderRepository(db),
		evidenceRepo,
		repository.NewShippingRepository(db),
	).Assemble(snapshot.OrderID)
	require.NoError(t, err)
	require.NotNil(t, assembly)
	require.NotNil(t, assembly.Package)
	require.Len(t, assembly.Shipments, 1)
	require.NotNil(t, assembly.TrackingContext)
	require.Len(t, assembly.TrackingEvents, 1)

	assert.Equal(t, "CURRENT-TRACKING", assembly.Shipments[0].TrackingNumber)
	assert.Equal(t, "current shipment", assembly.TrackingEvents[0].Description)
	assert.Equal(t, "CURRENT-TRACKING", assembly.TrackingContext.LatestDeliveryEvents[0].TrackingNumber)
	assert.Equal(t, []string{"https://carrier.example.test/current-pod.pdf"}, assembly.TrackingContext.ProviderPODURLs)
	assert.Contains(t, sourceHashes(assembly.Sources), attachmentSHA256)

	payload, err := json.Marshal(assembly)
	require.NoError(t, err)
	assert.NotContains(t, string(payload), "must-not-leak")
}

func TestOrderEvidencePackageAssemblerDoesNotUseTrackingEventsWithoutShipment(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&shipping.TrackingEvent{}))
	require.NoError(t, db.AutoMigrate(
		&shipping.TrackingProviderConfig{},
		&shipping.TrackingShipment{},
	))

	snapshot := servicePlanSnapshot(t, db, false)
	require.NoError(t, db.Create(&shipping.TrackingEvent{
		OrderID:             snapshot.OrderID,
		TrackingNumber:      "HISTORICAL-TRACKING",
		ProviderCarrierCode: "legacy-carrier",
		Status:              "delivered",
		Description:         "historical delivery",
		EventTime:           time.Now().UTC(),
	}).Error)

	assembly, err := NewOrderEvidencePackageAssembler(
		repository.NewOrderRepository(db),
		repository.NewOrderEvidenceRepository(db),
		repository.NewShippingRepository(db),
	).Assemble(snapshot.OrderID)
	require.NoError(t, err)
	require.NotNil(t, assembly)
	require.NotNil(t, assembly.TrackingContext)

	assert.Empty(t, assembly.TrackingEvents)
	assert.Empty(t, assembly.TrackingContext.LatestDeliveryEvents)
	assert.Empty(t, assembly.TrackingContext.ProviderPODURLs)
	assert.Contains(t, strings.Join(assembly.Warnings, "\n"), "no current tracking shipment")
}

func sourceHashes(sources []OrderEvidenceSourceReference) []string {
	hashes := make([]string, 0, len(sources))
	for _, source := range sources {
		if source.ContentSHA256 != "" {
			hashes = append(hashes, source.ContentSHA256)
		}
	}
	return hashes
}
