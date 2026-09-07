package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderEvidenceLifecycleUpdatesReadinessLocksAndRevises(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, true)
	repos := repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}
	evidenceService := NewOrderEvidenceService()
	attachmentService := NewOrderEvidenceAttachmentService()

	pkg, err := evidenceService.CreateInitialPackage(repos, snapshot)
	require.NoError(t, err)
	items, err := repos.OrderEvidence.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	require.Len(t, items, 5)

	byType := make(map[string]orderevidence.OrderEvidenceItem)
	for _, item := range items {
		byType[item.ItemType] = item
	}

	completeness, err := evidenceService.GetPackageCompleteness(repos, pkg.ID)
	require.NoError(t, err)
	assert.Equal(t, 20, completeness.Percent)
	assert.False(t, completeness.Ready)

	_, err = evidenceService.UpdateItem(repos, OrderEvidenceItemUpdateInput{
		ItemID:     byType[orderevidence.EvidenceItemTypeConfigurationConfirmation].ID,
		Status:     orderevidence.EvidenceItemStatusDraft,
		DataJSON:   []byte(`{"attempt":"must-not-change"}`),
		CapturedBy: 7,
	})
	require.ErrorContains(t, err, "configuration confirmation evidence item is immutable")

	attachment, err := attachmentService.RegisterReference(repos, OrderEvidenceAttachmentReferenceInput{
		EvidenceItemID:   byType[orderevidence.EvidenceItemTypeProductIdentity].ID,
		StorageKey:       "order-evidence/1/identity.jpg",
		OriginalFilename: "identity.jpg",
		MimeType:         "image/jpeg",
		SizeBytes:        12,
		SHA256:           "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		UploadedBy:       7,
	})
	require.NoError(t, err)
	require.NotNil(t, attachment)

	identityResult, err := evidenceService.UpdateItem(repos, OrderEvidenceItemUpdateInput{
		ItemID:     byType[orderevidence.EvidenceItemTypeProductIdentity].ID,
		Status:     orderevidence.EvidenceItemStatusComplete,
		DataJSON:   []byte(`{"serial_number":"SN-100"}`),
		CapturedBy: 7,
	})
	require.NoError(t, err)
	assert.Equal(t, 40, identityResult.Completeness.Percent)
	assert.False(t, identityResult.Completeness.Ready)

	for _, itemType := range []string{
		orderevidence.EvidenceItemTypeSpokeQCTension,
		orderevidence.EvidenceItemTypeOutboundWeightPackaging,
	} {
		item := byType[itemType]
		attachment, err := attachmentService.RegisterReference(repos, OrderEvidenceAttachmentReferenceInput{
			EvidenceItemID:   item.ID,
			StorageKey:       fmt.Sprintf("order-evidence/%d/%s.jpg", snapshot.OrderID, itemType),
			OriginalFilename: itemType + ".jpg",
			MimeType:         "image/jpeg",
			SizeBytes:        12,
			SHA256:           "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			UploadedBy:       7,
		})
		require.NoError(t, err)
		require.NotNil(t, attachment)
	}

	_, err = evidenceService.UpdateItem(repos, OrderEvidenceItemUpdateInput{
		ItemID: byType[orderevidence.EvidenceItemTypeSignedPOD].ID,
		Status: orderevidence.EvidenceItemStatusComplete,
		DataJSON: []byte(`{
			"schema_version": 1,
			"tracking_number": "TRACK-100",
			"delivered_at": "2026-09-05T10:30:00Z"
		}`),
		CapturedBy: 7,
	})
	require.ErrorIs(t, err, ErrOrderEvidenceAttachmentRequired)

	podAttachment, err := attachmentService.RegisterReference(repos, OrderEvidenceAttachmentReferenceInput{
		EvidenceItemID:   byType[orderevidence.EvidenceItemTypeSignedPOD].ID,
		StorageKey:       fmt.Sprintf("order-evidence/%d/signed-pod.jpg", snapshot.OrderID),
		OriginalFilename: "signed-pod.jpg",
		MimeType:         "image/jpeg",
		SizeBytes:        12,
		SHA256:           "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
		UploadedBy:       7,
	})
	require.NoError(t, err)
	require.NotNil(t, podAttachment)

	for _, itemType := range []string{
		orderevidence.EvidenceItemTypeSpokeQCTension,
		orderevidence.EvidenceItemTypeOutboundWeightPackaging,
		orderevidence.EvidenceItemTypeSignedPOD,
	} {
		dataJSON := []byte(`{"verified":true}`)
		if itemType == orderevidence.EvidenceItemTypeSpokeQCTension {
			dataJSON = []byte(`{
				"manual_conclusion": "fail",
				"measured_value": 140,
				"reference_range": {"minimum": 105, "maximum": 130}
			}`)
		} else if itemType == orderevidence.EvidenceItemTypeOutboundWeightPackaging {
			dataJSON = []byte(`{
				"schema_version": 1,
				"gross_weight_g": 12640,
				"package_count": 2,
				"packaging_method": "double wall carton"
			}`)
		} else if itemType == orderevidence.EvidenceItemTypeSignedPOD {
			dataJSON = []byte(`{
				"schema_version": 1,
				"tracking_number": "TRACK-100",
				"delivered_at": "2026-09-05T10:30:00Z"
			}`)
		}
		result, err := evidenceService.UpdateItem(repos, OrderEvidenceItemUpdateInput{
			ItemID:     byType[itemType].ID,
			Status:     orderevidence.EvidenceItemStatusComplete,
			DataJSON:   dataJSON,
			CapturedBy: 7,
		})
		require.NoError(t, err)
		if itemType == orderevidence.EvidenceItemTypeSignedPOD {
			assert.True(t, result.Completeness.Ready)
			assert.Equal(t, 100, result.Completeness.Percent)
			assert.Equal(t, orderevidence.PackageStatusReady, result.Package.Status)
		}
	}

	lockedAt := time.Date(2026, 9, 4, 13, 0, 0, 0, time.UTC)
	lockedPackage, completeness, err := evidenceService.LockPackage(repos, pkg.ID, lockedAt)
	require.NoError(t, err)
	assert.Equal(t, orderevidence.PackageStatusLocked, lockedPackage.Status)
	assert.Equal(t, lockedAt, *lockedPackage.LockedAt)
	assert.True(t, completeness.Ready)

	_, err = evidenceService.UpdateItem(repos, OrderEvidenceItemUpdateInput{
		ItemID:     byType[orderevidence.EvidenceItemTypeProductIdentity].ID,
		Status:     orderevidence.EvidenceItemStatusDraft,
		DataJSON:   []byte(`{"serial_number":"SN-101"}`),
		CapturedBy: 7,
	})
	require.ErrorIs(t, err, orderevidence.ErrOrderEvidencePackageLocked)

	_, err = attachmentService.RegisterReference(repos, OrderEvidenceAttachmentReferenceInput{
		EvidenceItemID:   byType[orderevidence.EvidenceItemTypeProductIdentity].ID,
		StorageKey:       "order-evidence/1/after-lock.jpg",
		OriginalFilename: "after-lock.jpg",
		MimeType:         "image/jpeg",
		SizeBytes:        12,
		SHA256:           "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		UploadedBy:       7,
	})
	require.ErrorIs(t, err, orderevidence.ErrOrderEvidencePackageLocked)

	revision, err := evidenceService.CreateRevision(repos, pkg.ID, 8)
	require.NoError(t, err)
	assert.Equal(t, 2, revision.PackageVersion)
	assert.Equal(t, orderevidence.PackageStatusReady, revision.Status)
	assert.Equal(t, pkg.ID, lockedPackage.ID)

	revisionItems, err := repos.OrderEvidence.ListItemsByPackageID(revision.ID)
	require.NoError(t, err)
	require.Len(t, revisionItems, len(items))
	revisionIdentity := findEvidenceItemByType(revisionItems, orderevidence.EvidenceItemTypeProductIdentity)
	revisionAttachments, err := repos.OrderEvidence.ListAttachmentsByItemID(revisionIdentity.ID)
	require.NoError(t, err)
	require.Len(t, revisionAttachments, 1)
	assert.Equal(t, attachment.StorageKey, revisionAttachments[0].StorageKey)
}

func TestOrderEvidenceLifecycleRequiresPhotoAttachmentBeforePhysicalEvidenceCompletion(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, false)
	repos := repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}
	evidenceService := NewOrderEvidenceService()
	pkg, err := evidenceService.CreateInitialPackage(repos, snapshot)
	require.NoError(t, err)
	items, err := repos.OrderEvidence.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	identity := findEvidenceItemByType(items, orderevidence.EvidenceItemTypeProductIdentity)

	_, err = evidenceService.UpdateItem(repos, OrderEvidenceItemUpdateInput{
		ItemID:     identity.ID,
		Status:     orderevidence.EvidenceItemStatusComplete,
		DataJSON:   []byte(`{"manual_note":"captured at shipping"}`),
		CapturedBy: 7,
	})
	require.ErrorIs(t, err, ErrOrderEvidenceAttachmentRequired)
}

func TestOrderEvidenceLifecycleValidatesOutboundAndPODContractsOnly(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, false)
	repos := repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}
	evidenceService := NewOrderEvidenceService()
	attachmentService := NewOrderEvidenceAttachmentService()
	pkg, err := evidenceService.CreateInitialPackage(repos, snapshot)
	require.NoError(t, err)
	items, err := repos.OrderEvidence.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)

	outbound := findEvidenceItemByType(items, orderevidence.EvidenceItemTypeOutboundWeightPackaging)
	pod := findEvidenceItemByType(items, orderevidence.EvidenceItemTypeSignedPOD)

	for _, item := range []orderevidence.OrderEvidenceItem{outbound, pod} {
		_, err = attachmentService.RegisterReference(repos, OrderEvidenceAttachmentReferenceInput{
			EvidenceItemID:   item.ID,
			StorageKey:       fmt.Sprintf("order-evidence/%d/%d.jpg", snapshot.OrderID, item.ID),
			OriginalFilename: fmt.Sprintf("%d.jpg", item.ID),
			MimeType:         "image/jpeg",
			SizeBytes:        12,
			SHA256:           fmt.Sprintf("%064d", item.ID),
			UploadedBy:       7,
		})
		require.NoError(t, err)
	}

	_, err = evidenceService.UpdateItem(repos, OrderEvidenceItemUpdateInput{
		ItemID: outbound.ID,
		Status: orderevidence.EvidenceItemStatusComplete,
		DataJSON: []byte(`{
			"schema_version": 1,
			"package_count": 1,
			"packaging_method": "carton"
		}`),
		CapturedBy: 7,
	})
	require.ErrorContains(t, err, "gross_weight_g")

	_, err = evidenceService.UpdateItem(repos, OrderEvidenceItemUpdateInput{
		ItemID: pod.ID,
		Status: orderevidence.EvidenceItemStatusComplete,
		DataJSON: []byte(`{
			"schema_version": 1,
			"tracking_number": "TRACK-100"
		}`),
		CapturedBy: 7,
	})
	require.ErrorContains(t, err, "delivered_at")
}

func TestOrderEvidenceLifecyclePreservesManualTensionConclusionAndValues(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, true)
	repos := repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}
	evidenceService := NewOrderEvidenceService()
	attachmentService := NewOrderEvidenceAttachmentService()
	pkg, err := evidenceService.CreateInitialPackage(repos, snapshot)
	require.NoError(t, err)
	items, err := repos.OrderEvidence.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	tension := findEvidenceItemByType(items, orderevidence.EvidenceItemTypeSpokeQCTension)

	_, err = attachmentService.RegisterReference(repos, OrderEvidenceAttachmentReferenceInput{
		EvidenceItemID:   tension.ID,
		StorageKey:       fmt.Sprintf("order-evidence/%d/tension.jpg", snapshot.OrderID),
		OriginalFilename: "tension.jpg",
		MimeType:         "image/jpeg",
		SizeBytes:        12,
		SHA256:           "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		UploadedBy:       7,
	})
	require.NoError(t, err)

	input := []byte(`{
		"manual_conclusion": "fail",
		"measured_value": 140,
		"reference_range": {"minimum": 105, "maximum": 130}
	}`)
	result, err := evidenceService.UpdateItem(repos, OrderEvidenceItemUpdateInput{
		ItemID:     tension.ID,
		Status:     orderevidence.EvidenceItemStatusComplete,
		DataJSON:   input,
		CapturedBy: 7,
	})
	require.NoError(t, err)
	assert.JSONEq(t, string(input), string(result.Item.DataJSON))
	assert.Equal(t, orderevidence.EvidenceItemStatusComplete, result.Item.Status)
	assert.NotContains(t, string(result.Item.DataJSON), "system_conclusion")
	assert.NotContains(t, string(result.Item.DataJSON), "quality_status")
}

func TestOrderEvidenceLifecycleRequiresCompletePackageBeforeLockAndLockedBeforeRevision(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, false)
	repos := repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}
	evidenceService := NewOrderEvidenceService()
	pkg, err := evidenceService.CreateInitialPackage(repos, snapshot)
	require.NoError(t, err)

	_, _, err = evidenceService.LockPackage(repos, pkg.ID, time.Now().UTC())
	require.ErrorIs(t, err, orderevidence.ErrOrderEvidencePackageIncomplete)

	_, err = evidenceService.CreateRevision(repos, pkg.ID, 8)
	require.ErrorIs(t, err, orderevidence.ErrOrderEvidenceRevisionSource)

	var missingRevisionSource bool
	if errors.Is(err, orderevidence.ErrOrderEvidenceRevisionSource) {
		missingRevisionSource = true
	}
	assert.True(t, missingRevisionSource)
}

func TestOrderEvidenceLifecycleRequiresReasonForWaivedItem(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	snapshot := servicePlanSnapshot(t, db, false)
	repos := repository.TxRepositories{
		OrderEvidence: repository.NewOrderEvidenceRepository(db),
	}
	evidenceService := NewOrderEvidenceService()
	pkg, err := evidenceService.CreateInitialPackage(repos, snapshot)
	require.NoError(t, err)
	items, err := repos.OrderEvidence.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)

	identity := findEvidenceItemByType(items, orderevidence.EvidenceItemTypeProductIdentity)
	_, err = evidenceService.UpdateItem(repos, OrderEvidenceItemUpdateInput{
		ItemID:     identity.ID,
		Status:     orderevidence.EvidenceItemStatusWaived,
		DataJSON:   []byte(`{}`),
		CapturedBy: 7,
	})
	require.ErrorContains(t, err, "waiver_reason is required")

	result, err := evidenceService.UpdateItem(repos, OrderEvidenceItemUpdateInput{
		ItemID:     identity.ID,
		Status:     orderevidence.EvidenceItemStatusWaived,
		DataJSON:   []byte(`{"waiver_reason":"supplier batch unavailable"}`),
		CapturedBy: 7,
	})
	require.NoError(t, err)
	assert.Equal(t, orderevidence.EvidenceItemStatusWaived, result.Item.Status)
	assert.Equal(t, 50, result.Completeness.Percent)
}

func findEvidenceItemByType(items []orderevidence.OrderEvidenceItem, itemType string) orderevidence.OrderEvidenceItem {
	for _, item := range items {
		if item.ItemType == itemType {
			return item
		}
	}
	return orderevidence.OrderEvidenceItem{}
}
