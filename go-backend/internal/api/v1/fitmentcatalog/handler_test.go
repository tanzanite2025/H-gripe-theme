package fitmentcatalog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPublicFitmentCatalogListsOnlyEnabledFrameEntriesAndHubSpecs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newPublicFitmentCatalogTestDB(t)
	frameService, forkService, hubService := newPublicFitmentCatalogServicesForTest(db)

	activeRear, err := hubService.Create(service.FitmentHubSpecificationInput{
		SpecCode:      "HUB-R-12X142",
		DisplayName:   "Rear 12x142 Thru Axle",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 142,
	})
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(activeRear.ID, true)
	require.NoError(t, err)

	staleRear, err := hubService.Create(service.FitmentHubSpecificationInput{
		SpecCode:      "HUB-R-12X148",
		DisplayName:   "Rear 12x148 Disabled",
		Position:      fitmentcatalogdomain.HubPositionRear,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 148,
	})
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(staleRear.ID, true)
	require.NoError(t, err)

	frame, err := frameService.Create(service.FrameFitmentEntryInput{
		BrandName:           "Example",
		ModelName:           "Road X",
		YearMode:            fitmentcatalogdomain.YearModeAll,
		HubSpecificationIDs: []uint{activeRear.ID, staleRear.ID},
	})
	require.NoError(t, err)
	_, err = frameService.UpdateStatus(frame.ID, true)
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(staleRear.ID, false)
	require.NoError(t, err)

	yearFrom := 2020
	yearTo := 2024
	rangeFrame, err := frameService.Create(service.FrameFitmentEntryInput{
		BrandName:           "Example",
		ModelName:           "Range Bike",
		YearMode:            fitmentcatalogdomain.YearModeRange,
		YearFrom:            &yearFrom,
		YearTo:              &yearTo,
		HubSpecificationIDs: []uint{activeRear.ID},
	})
	require.NoError(t, err)
	_, err = frameService.UpdateStatus(rangeFrame.ID, true)
	require.NoError(t, err)

	unknownFrame, err := frameService.Create(service.FrameFitmentEntryInput{
		BrandName:           "Example",
		ModelName:           "Range Unknown",
		YearMode:            fitmentcatalogdomain.YearModeUnknown,
		HubSpecificationIDs: []uint{activeRear.ID},
	})
	require.NoError(t, err)
	_, err = frameService.UpdateStatus(unknownFrame.ID, true)
	require.NoError(t, err)

	_, err = frameService.Create(service.FrameFitmentEntryInput{
		BrandName:           "Example",
		ModelName:           "Hidden Bike",
		YearMode:            fitmentcatalogdomain.YearModeAll,
		HubSpecificationIDs: []uint{activeRear.ID},
	})
	require.NoError(t, err)

	router := newPublicFitmentCatalogTestRouter(frameService, forkService, hubService)
	w := performPublicFitmentCatalogGET(router, "/api/v1/fitment-catalog/frame-entries?search=Example")

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"code":0`)
	require.Contains(t, w.Body.String(), `"model_name":"Road X"`)
	require.Contains(t, w.Body.String(), `"spec_code":"HUB-R-12X142"`)
	require.Contains(t, w.Body.String(), `"hub_specification_count":1`)
	require.NotContains(t, w.Body.String(), "Hidden Bike")
	require.NotContains(t, w.Body.String(), "HUB-R-12X148")
	require.NotContains(t, w.Body.String(), "is_enabled")

	w = performPublicFitmentCatalogGET(router, "/api/v1/fitment-catalog/frame-entries?search=Range&year=2022")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"model_name":"Range Bike"`)
	require.NotContains(t, w.Body.String(), "Range Unknown")

	w = performPublicFitmentCatalogGET(router, "/api/v1/fitment-catalog/frame-entries?search=Range&year=2025")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.NotContains(t, w.Body.String(), "Range Bike")
}

func TestPublicFitmentCatalogListsForkEntriesAndFiltersHubSpecifications(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newPublicFitmentCatalogTestDB(t)
	frameService, forkService, hubService := newPublicFitmentCatalogServicesForTest(db)

	front, err := hubService.Create(service.FitmentHubSpecificationInput{
		SpecCode:      "HUB-F-12X100",
		DisplayName:   "Front 12x100 Thru Axle",
		Position:      fitmentcatalogdomain.HubPositionFront,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 100,
	})
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(front.ID, true)
	require.NoError(t, err)

	disabledFront, err := hubService.Create(service.FitmentHubSpecificationInput{
		SpecCode:      "HUB-F-DISABLED",
		DisplayName:   "Disabled Front",
		Position:      fitmentcatalogdomain.HubPositionFront,
		AxleType:      fitmentcatalogdomain.HubAxleTypeThruAxle,
		AxleSpacingMM: 110,
	})
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(disabledFront.ID, true)
	require.NoError(t, err)

	fork, err := forkService.Create(service.ForkFitmentEntryInput{
		BrandName:           "RockShox",
		ModelName:           "SID SL",
		YearMode:            fitmentcatalogdomain.YearModeAll,
		HubSpecificationIDs: []uint{front.ID, disabledFront.ID},
	})
	require.NoError(t, err)
	_, err = forkService.UpdateStatus(fork.ID, true)
	require.NoError(t, err)
	_, err = hubService.UpdateStatus(disabledFront.ID, false)
	require.NoError(t, err)

	router := newPublicFitmentCatalogTestRouter(frameService, forkService, hubService)
	w := performPublicFitmentCatalogGET(router, "/api/v1/fitment-catalog/fork-entries?search=SID")

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"model_name":"SID SL"`)
	require.Contains(t, w.Body.String(), `"spec_code":"HUB-F-12X100"`)
	require.NotContains(t, w.Body.String(), "HUB-F-DISABLED")

	w = performPublicFitmentCatalogGET(router, "/api/v1/fitment-catalog/hub-specifications?position=front&axle_type=thru_axle")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"spec_code":"HUB-F-12X100"`)
	require.NotContains(t, w.Body.String(), "HUB-F-DISABLED")
	require.NotContains(t, w.Body.String(), "is_enabled")
}

func TestPublicFitmentCatalogRejectsInvalidHubSpecificationFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newPublicFitmentCatalogTestDB(t)
	frameService, forkService, hubService := newPublicFitmentCatalogServicesForTest(db)
	router := newPublicFitmentCatalogTestRouter(frameService, forkService, hubService)

	w := performPublicFitmentCatalogGET(router, "/api/v1/fitment-catalog/hub-specifications?position=center")

	require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), "unsupported position")

	w = performPublicFitmentCatalogGET(router, "/api/v1/fitment-catalog/frame-entries?year=not-a-year")

	require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), "year must be an integer")
}

func newPublicFitmentCatalogTestRouter(
	frameService *service.FrameFitmentEntryService,
	forkService *service.ForkFitmentEntryService,
	hubService *service.FitmentHubSpecificationService,
) *gin.Engine {
	router := gin.New()
	handler := NewHandler(frameService, forkService, hubService)
	handler.RegisterRoutes(router.Group("/api/v1/fitment-catalog"))
	return router
}

func performPublicFitmentCatalogGET(router *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	router.ServeHTTP(w, req)
	return w
}

func newPublicFitmentCatalogServicesForTest(
	db *gorm.DB,
) (*service.FrameFitmentEntryService, *service.ForkFitmentEntryService, *service.FitmentHubSpecificationService) {
	frameAssociationRepo := repository.NewFitmentFrameHubSpecificationRepository(db)
	forkAssociationRepo := repository.NewFitmentForkHubSpecificationRepository(db)
	hubRepo := repository.NewFitmentHubSpecificationRepository(db, frameAssociationRepo)
	hubRepo.ConfigureForkHubSpecificationRepository(forkAssociationRepo)
	hubService := service.NewFitmentHubSpecificationService(hubRepo, frameAssociationRepo)
	hubService.ConfigureForkHubSpecificationRepository(forkAssociationRepo)

	return service.NewFrameFitmentEntryService(
			repository.NewFrameFitmentEntryRepository(db),
			hubRepo,
			frameAssociationRepo,
		),
		service.NewForkFitmentEntryService(
			repository.NewForkFitmentEntryRepository(db),
			hubRepo,
			forkAssociationRepo,
		),
		hubService
}

func newPublicFitmentCatalogTestDB(t *testing.T) *gorm.DB {
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
