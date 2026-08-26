package service

import (
	"testing"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestFrameFitmentEntryServicePersistsRearHubSpecificationsInOneDomainTransaction(t *testing.T) {
	db := newFitmentCatalogServiceTestDB(t)
	frameService, hubService := newFitmentCatalogServicesForTest(db)

	rear, err := hubService.Create(FitmentHubSpecificationInput{
		SpecCode:      "HUB-R-12X142",
		DisplayName:   "Rear 12x142 Thru Axle",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 142,
		IsEnabled:     true,
	})
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(rear.ID, true)
	require.NoError(t, err)

	frame, err := frameService.Create(FrameFitmentEntryInput{
		BrandName:           "Example",
		ModelName:           "Road X",
		YearMode:            fitmentcatalogdomain.YearModeAll,
		IsEnabled:           true,
		HubSpecificationIDs: []uint{rear.ID},
	})
	require.NoError(t, err)
	require.False(t, frame.IsEnabled)
	require.Len(t, frame.HubSpecifications, 1)
	require.Equal(t, rear.ID, frame.HubSpecifications[0].ID)

	var links []fitmentcatalogdomain.FrameHubSpecification
	require.NoError(t, db.Find(&links).Error)
	require.Len(t, links, 1)
	require.Equal(t, frame.ID, links[0].FrameEntryID)
	require.Equal(t, rear.ID, links[0].HubSpecificationID)

	enabledFrame, err := frameService.UpdateStatus(frame.ID, true)
	require.NoError(t, err)
	require.True(t, enabledFrame.IsEnabled)

	list, total, err := frameService.List(FrameFitmentEntryListInput{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	require.Equal(t, 1, list[0].HubSpecificationCount)
}

func TestFrameFitmentEntryServiceRejectsFrontOrDisabledHubSpecifications(t *testing.T) {
	db := newFitmentCatalogServiceTestDB(t)
	frameService, hubService := newFitmentCatalogServicesForTest(db)

	front, err := hubService.Create(FitmentHubSpecificationInput{
		SpecCode:      "HUB-F-100",
		DisplayName:   "Front 100 Quick Release",
		Position:      fitmentcatalogdomain.HubPositionFront,
		AxleType:      fitmentcatalogdomain.HubAxleTypeQuickRelease,
		AxleSpacingMM: 100,
		IsEnabled:     true,
	})
	require.NoError(t, err)

	disabledRear, err := hubService.Create(FitmentHubSpecificationInput{
		SpecCode:      "HUB-R-DISABLED",
		DisplayName:   "Disabled Rear",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 148,
		IsEnabled:     false,
	})
	require.NoError(t, err)

	_, err = frameService.Create(FrameFitmentEntryInput{
		BrandName:           "Example",
		ModelName:           "Front Match",
		YearMode:            fitmentcatalogdomain.YearModeUnknown,
		HubSpecificationIDs: []uint{front.ID},
	})
	require.ErrorIs(t, err, ErrFrameFitmentEntryInvalid)

	_, err = frameService.Create(FrameFitmentEntryInput{
		BrandName:           "Example",
		ModelName:           "Disabled Match",
		YearMode:            fitmentcatalogdomain.YearModeUnknown,
		HubSpecificationIDs: []uint{disabledRear.ID},
	})
	require.ErrorIs(t, err, ErrFrameFitmentEntryInvalid)
}

func TestFrameFitmentEntryServiceUpdateReplacesAssociationAtomically(t *testing.T) {
	db := newFitmentCatalogServiceTestDB(t)
	frameService, hubService := newFitmentCatalogServicesForTest(db)

	first, err := hubService.Create(FitmentHubSpecificationInput{
		SpecCode:      "HUB-R-12X142",
		DisplayName:   "Rear 12x142",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 142,
		IsEnabled:     true,
	})
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(first.ID, true)
	require.NoError(t, err)
	second, err := hubService.Create(FitmentHubSpecificationInput{
		SpecCode:      "HUB-R-12X148",
		DisplayName:   "Rear 12x148",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 148,
		IsEnabled:     true,
	})
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(second.ID, true)
	require.NoError(t, err)

	frame, err := frameService.Create(FrameFitmentEntryInput{
		BrandName: "Example",
		ModelName: "Road X",
		YearMode:  fitmentcatalogdomain.YearModeAll,
	})
	require.NoError(t, err)

	updated, err := frameService.Update(frame.ID, FrameFitmentEntryInput{
		BrandName:           "Example",
		ModelName:           "Road X",
		YearMode:            fitmentcatalogdomain.YearModeAll,
		HubSpecificationIDs: []uint{first.ID, second.ID},
	})
	require.NoError(t, err)
	require.Len(t, updated.HubSpecifications, 2)

	updated, err = frameService.Update(frame.ID, FrameFitmentEntryInput{
		BrandName:           "Example",
		ModelName:           "Road X",
		YearMode:            fitmentcatalogdomain.YearModeAll,
		HubSpecificationIDs: []uint{second.ID},
	})
	require.NoError(t, err)
	require.Len(t, updated.HubSpecifications, 1)
	require.Equal(t, second.ID, updated.HubSpecifications[0].ID)
}

func TestForkFitmentEntryServicePersistsFrontHubSpecificationsInOneDomainTransaction(t *testing.T) {
	db := newFitmentCatalogServiceTestDB(t)
	_, forkService, hubService := newFitmentCatalogServicesWithForkForTest(db)

	front, err := hubService.Create(FitmentHubSpecificationInput{
		SpecCode:      "HUB-F-12X100",
		DisplayName:   "Front 12x100 Thru Axle",
		Position:      fitmentcatalogdomain.HubPositionFront,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 100,
		IsEnabled:     true,
	})
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(front.ID, true)
	require.NoError(t, err)

	fork, err := forkService.Create(ForkFitmentEntryInput{
		BrandName:           "Fox",
		ModelName:           "32 Step-Cast",
		YearMode:            fitmentcatalogdomain.YearModeAll,
		IsEnabled:           true,
		HubSpecificationIDs: []uint{front.ID},
	})
	require.NoError(t, err)
	require.False(t, fork.IsEnabled)
	require.Len(t, fork.HubSpecifications, 1)
	require.Equal(t, front.ID, fork.HubSpecifications[0].ID)

	var links []fitmentcatalogdomain.ForkHubSpecification
	require.NoError(t, db.Find(&links).Error)
	require.Len(t, links, 1)
	require.Equal(t, fork.ID, links[0].ForkEntryID)
	require.Equal(t, front.ID, links[0].HubSpecificationID)

	enabledFork, err := forkService.UpdateStatus(fork.ID, true)
	require.NoError(t, err)
	require.True(t, enabledFork.IsEnabled)

	list, total, err := forkService.List(ForkFitmentEntryListInput{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	require.Equal(t, 1, list[0].HubSpecificationCount)

	hub, err := hubService.Get(front.ID)
	require.NoError(t, err)
	require.Equal(t, 1, hub.ForkReferenceCount)
}

func TestForkFitmentEntryServiceRejectsRearOrDisabledHubSpecifications(t *testing.T) {
	db := newFitmentCatalogServiceTestDB(t)
	_, forkService, hubService := newFitmentCatalogServicesWithForkForTest(db)

	rear, err := hubService.Create(FitmentHubSpecificationInput{
		SpecCode:      "HUB-R-142",
		DisplayName:   "Rear 142",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 142,
		IsEnabled:     true,
	})
	require.NoError(t, err)

	disabledFront, err := hubService.Create(FitmentHubSpecificationInput{
		SpecCode:      "HUB-F-DISABLED",
		DisplayName:   "Disabled Front",
		Position:      fitmentcatalogdomain.HubPositionFront,
		AxleType:      fitmentcatalogdomain.HubAxleTypeQuickRelease,
		AxleSpacingMM: 100,
		IsEnabled:     false,
	})
	require.NoError(t, err)

	_, err = forkService.Create(ForkFitmentEntryInput{
		BrandName:           "Fox",
		ModelName:           "Rear Match",
		YearMode:            fitmentcatalogdomain.YearModeUnknown,
		HubSpecificationIDs: []uint{rear.ID},
	})
	require.ErrorIs(t, err, ErrForkFitmentEntryInvalid)

	_, err = forkService.Create(ForkFitmentEntryInput{
		BrandName:           "Fox",
		ModelName:           "Disabled Match",
		YearMode:            fitmentcatalogdomain.YearModeUnknown,
		HubSpecificationIDs: []uint{disabledFront.ID},
	})
	require.ErrorIs(t, err, ErrForkFitmentEntryInvalid)
}

func TestFitmentHubSpecificationServicePreventsDeletingReferencedSpecification(t *testing.T) {
	db := newFitmentCatalogServiceTestDB(t)
	frameService, hubService := newFitmentCatalogServicesForTest(db)

	specification, err := hubService.Create(FitmentHubSpecificationInput{
		SpecCode:      "HUB-R-12X142",
		DisplayName:   "Rear 12x142",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 142,
		IsEnabled:     true,
	})
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(specification.ID, true)
	require.NoError(t, err)

	_, err = frameService.Create(FrameFitmentEntryInput{
		BrandName:           "Example",
		ModelName:           "Road X",
		YearMode:            fitmentcatalogdomain.YearModeAll,
		HubSpecificationIDs: []uint{specification.ID},
	})
	require.NoError(t, err)

	require.ErrorIs(t, hubService.Delete(specification.ID), ErrFitmentHubSpecificationInUse)
	require.NoError(t, func() error {
		_, statusErr := hubService.UpdateStatus(specification.ID, false)
		return statusErr
	}())
}

func TestFitmentHubSpecificationServicePreventsChangingPositionOfReferencedSpecification(t *testing.T) {
	db := newFitmentCatalogServiceTestDB(t)
	frameService, hubService := newFitmentCatalogServicesForTest(db)

	specification, err := hubService.Create(FitmentHubSpecificationInput{
		SpecCode:      "HUB-R-12X142",
		DisplayName:   "Rear 12x142",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 142,
	})
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(specification.ID, true)
	require.NoError(t, err)
	_, err = frameService.Create(FrameFitmentEntryInput{
		BrandName:           "Example",
		ModelName:           "Road X",
		YearMode:            fitmentcatalogdomain.YearModeAll,
		HubSpecificationIDs: []uint{specification.ID},
	})
	require.NoError(t, err)

	_, err = hubService.Update(specification.ID, FitmentHubSpecificationInput{
		SpecCode:      specification.SpecCode,
		DisplayName:   specification.DisplayName,
		Position:      fitmentcatalogdomain.HubPositionFront,
		AxleType:      specification.AxleType,
		AxleSpacingMM: specification.AxleSpacingMM,
	})
	require.ErrorIs(t, err, ErrFitmentHubSpecificationInvalid)
}

func TestFitmentHubSpecificationServicePreventsDeletingForkReferencedSpecification(t *testing.T) {
	db := newFitmentCatalogServiceTestDB(t)
	_, forkService, hubService := newFitmentCatalogServicesWithForkForTest(db)

	specification, err := hubService.Create(FitmentHubSpecificationInput{
		SpecCode:      "HUB-F-12X100",
		DisplayName:   "Front 12x100",
		Position:      fitmentcatalogdomain.HubPositionFront,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 100,
		IsEnabled:     true,
	})
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(specification.ID, true)
	require.NoError(t, err)

	_, err = forkService.Create(ForkFitmentEntryInput{
		BrandName:           "RockShox",
		ModelName:           "SID SL",
		YearMode:            fitmentcatalogdomain.YearModeAll,
		HubSpecificationIDs: []uint{specification.ID},
	})
	require.NoError(t, err)

	require.ErrorIs(t, hubService.Delete(specification.ID), ErrFitmentHubSpecificationInUse)
}

func newFitmentCatalogServicesForTest(
	db *gorm.DB,
) (*FrameFitmentEntryService, *FitmentHubSpecificationService) {
	frameService, _, hubService := newFitmentCatalogServicesWithForkForTest(db)
	return frameService, hubService
}

func newFitmentCatalogServicesWithForkForTest(
	db *gorm.DB,
) (*FrameFitmentEntryService, *ForkFitmentEntryService, *FitmentHubSpecificationService) {
	frameAssociationRepo := repository.NewFitmentFrameHubSpecificationRepository(db)
	forkAssociationRepo := repository.NewFitmentForkHubSpecificationRepository(db)
	hubRepo := repository.NewFitmentHubSpecificationRepository(db, frameAssociationRepo)
	hubRepo.ConfigureForkHubSpecificationRepository(forkAssociationRepo)
	hubService := NewFitmentHubSpecificationService(hubRepo, frameAssociationRepo)
	hubService.ConfigureForkHubSpecificationRepository(forkAssociationRepo)
	return NewFrameFitmentEntryService(
			repository.NewFrameFitmentEntryRepository(db),
			hubRepo,
			frameAssociationRepo,
		),
		NewForkFitmentEntryService(
			repository.NewForkFitmentEntryRepository(db),
			hubRepo,
			forkAssociationRepo,
		),
		hubService
}

func newFitmentCatalogServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&fitmentcatalogdomain.FrameFitmentEntry{},
		&fitmentcatalogdomain.ForkFitmentEntry{},
		&fitmentcatalogdomain.HubSpecification{},
		&fitmentcatalogdomain.FrameHubSpecification{},
		&fitmentcatalogdomain.ForkHubSpecification{},
	))
	return db
}
