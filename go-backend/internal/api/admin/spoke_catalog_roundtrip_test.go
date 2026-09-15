package admin_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	adminapi "commerce-platform/internal/api/admin"
	spokeapi "commerce-platform/internal/api/v1/spoke"
	productdomain "commerce-platform/internal/domain/product"
	spokedomain "commerce-platform/internal/domain/spoke"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSpokeCatalogHTTPRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("put save round-trip", func(t *testing.T) {
		runSpokeCatalogHTTPRoundTrip(t, func(payload spokedomain.ExportResponse) (*http.Request, error) {
			return newJSONCatalogRequest(http.MethodPut, "/api/admin/spoke-catalog", payload)
		})
	})

	t.Run("multipart import round-trip", func(t *testing.T) {
		runSpokeCatalogHTTPRoundTrip(t, func(payload spokedomain.ExportResponse) (*http.Request, error) {
			return newMultipartCatalogRequest(http.MethodPost, "/api/admin/spoke-catalog/import", payload)
		})
	})

	t.Run("xlsx preset template round-trip", func(t *testing.T) {
		runSpokePresetTemplateHTTPRoundTrip(t)
	})
}

func runSpokeCatalogHTTPRoundTrip(t *testing.T, buildRequest func(spokedomain.ExportResponse) (*http.Request, error)) {
	t.Helper()

	router := newSpokeCatalogRoundTripRouter(t)
	requestPayload, expectedExport := spokeCatalogRoundTripFixture()

	request, err := buildRequest(requestPayload)
	require.NoError(t, err)

	replaceExport := executeSpokeCatalogRequest(t, router, request)
	require.Equal(t, expectedExport, replaceExport)

	// The admin catalog endpoint is the authoritative round-trip projection.
	// The public spoke export intentionally redacts proprietary geometry and
	// verified build measurements, so it must not be used for this assertion.
	adminRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/admin/spoke-catalog", nil)
	adminExport := executeSpokeCatalogRequest(t, router, adminRequest)
	require.Equal(t, expectedExport, adminExport)

	publicRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/spoke/export", nil)
	publicExport := executeSpokeCatalogRequest(t, router, publicRequest)
	assertPublicSpokeExportRedacted(t, publicExport, expectedExport)
}

func newSpokeCatalogRoundTripRouter(t *testing.T) *gin.Engine {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&productdomain.ProductBrand{},
		&spokedomain.CatalogRimBrand{},
		&spokedomain.CatalogRimModel{},
		&spokedomain.CatalogHubBrand{},
		&spokedomain.CatalogHubModel{},
		&spokedomain.CatalogBuildPreset{},
	))

	require.NoError(t, db.Create(&productdomain.ProductBrand{
		Name:      "DT Swiss",
		Slug:      "dt_swiss",
		IsEnabled: true,
		SortOrder: 0,
	}).Error)

	spokeRepo := repository.NewSpokeRepository(db)
	spokeRepo.ConfigureProductBrandRepository(repository.NewProductBrandRepository(db))
	spokeService := service.NewSpokeService(spokeRepo)
	_, err = spokeService.ReplaceCatalog(spokeCatalogRoundTripSeed())
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/api/admin/spoke-catalog", adminapi.NewSpokeCatalogHandler(spokeService).Replace)
	router.POST("/api/admin/spoke-catalog/import", adminapi.NewSpokeCatalogHandler(spokeService).Import)
	router.GET("/api/admin/spoke-catalog", adminapi.NewSpokeCatalogHandler(spokeService).Get)
	router.GET("/api/admin/spoke-catalog/preset-template", adminapi.NewSpokeCatalogHandler(spokeService).DownloadPresetTemplate)
	router.POST("/api/admin/spoke-catalog/preset-template/import", adminapi.NewSpokeCatalogHandler(spokeService).ImportPresetTemplate)
	router.GET("/api/v1/spoke/export", spokeapi.NewHandler(spokeService).GetExport)
	return router
}

func runSpokePresetTemplateHTTPRoundTrip(t *testing.T) {
	t.Helper()

	router := newSpokeCatalogRoundTripRouter(t)
	adminRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/admin/spoke-catalog", nil)
	adminExport := executeSpokeCatalogRequest(t, router, adminRequest)
	publicRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/spoke/export", nil)
	publicExport := executeSpokeCatalogRequest(t, router, publicRequest)

	templateRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/admin/spoke-catalog/preset-template", nil)
	templateRecorder := httptest.NewRecorder()
	router.ServeHTTP(templateRecorder, templateRequest)
	require.Equal(t, http.StatusOK, templateRecorder.Code, templateRecorder.Body.String())
	require.Contains(t, templateRecorder.Header().Get("Content-Disposition"), "spoke-preset-template.xlsx")

	workbook, err := excelize.OpenReader(bytes.NewReader(templateRecorder.Body.Bytes()))
	require.NoError(t, err)
	defer workbook.Close()
	templateRows, err := workbook.GetRows(spokedomain.PresetTemplateSheet)
	require.NoError(t, err)
	require.NotEmpty(t, templateRows)
	require.Equal(t, spokedomain.PresetTemplateColumns, templateRows[0])
	instructionsVisible, err := workbook.GetSheetVisible(spokedomain.PresetTemplateInstructionsSheet)
	require.NoError(t, err)
	require.True(t, instructionsVisible)
	listsVisible, err := workbook.GetSheetVisible(spokedomain.PresetTemplateListsSheet)
	require.NoError(t, err)
	require.False(t, listsVisible)

	archive, err := zip.NewReader(bytes.NewReader(templateRecorder.Body.Bytes()), int64(templateRecorder.Body.Len()))
	require.NoError(t, err)
	var presetSheetXML []byte
	protectedSheetCount := 0
	for _, entry := range archive.File {
		if !strings.HasPrefix(entry.Name, "xl/worksheets/sheet") || !strings.HasSuffix(entry.Name, ".xml") {
			continue
		}
		reader, openErr := entry.Open()
		require.NoError(t, openErr)
		sheetXML, readErr := io.ReadAll(reader)
		require.NoError(t, reader.Close())
		require.NoError(t, readErr)
		if strings.Contains(string(sheetXML), "sheetProtection") {
			protectedSheetCount++
		}
		if entry.Name == "xl/worksheets/sheet1.xml" {
			presetSheetXML = sheetXML
		}
	}
	require.Contains(t, string(presetSheetXML), "sheetProtection")
	require.Contains(t, string(presetSheetXML), "dataValidations")
	require.GreaterOrEqual(t, protectedSheetCount, 3)

	// Template selectors only need catalog identifiers. Use the public
	// projection here to verify that the browser-facing payload remains safe.
	input := publicExport
	rimBrand := input.Rims[0]
	rimModel := rimBrand.Items[0]
	hubBrand := input.Hubs[0]
	hubModel := hubBrand.Items[0]
	row := []interface{}{
		" build_template_1 ",
		" Verified template build ",
		" template description ",
		" " + rimBrand.ID + " | " + rimBrand.Name + " ",
		" " + rimModel.ID + " | " + rimModel.Name + " ",
		" " + hubBrand.ID + " | " + hubBrand.Name + " ",
		" " + hubModel.ID + " | " + hubModel.Name + " ",
		" ",
		" 24 ",
		" 2 ",
		" standard ",
		"",
		"282",
		"",
		"",
		"284",
		" verified from xlsx ",
		" DT Swiss 350 ; 350 ",
	}
	require.NoError(t, workbook.SetSheetRow(spokedomain.PresetTemplateSheet, "A2", &row))

	var filledTemplate bytes.Buffer
	require.NoError(t, workbook.Write(&filledTemplate))

	importRequest, err := newMultipartFileRequest(
		http.MethodPost,
		"/api/admin/spoke-catalog/preset-template/import",
		"spoke-preset-template.xlsx",
		filledTemplate.Bytes(),
	)
	require.NoError(t, err)

	importedExport := executeSpokeCatalogRequest(t, router, importRequest)
	require.Len(t, importedExport.Presets, len(input.Presets)+1)
	importedPreset := importedExport.Presets[len(importedExport.Presets)-1]
	require.Equal(t, "build_template_1", importedPreset.ID)
	require.Equal(t, "Verified template build", importedPreset.Name)
	require.Equal(t, "verified from xlsx", importedPreset.ActualLengths.Notes)
	require.Equal(t, []string{"DT Swiss 350", "350"}, importedPreset.Keywords)
	require.Equal(t, "auto", importedPreset.WheelPosition)

	adminRequest = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/admin/spoke-catalog", nil)
	adminExport = executeSpokeCatalogRequest(t, router, adminRequest)
	require.Equal(t, importedExport, adminExport)

	publicRequest = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/spoke/export", nil)
	publicExport = executeSpokeCatalogRequest(t, router, publicRequest)
	assertPublicSpokeExportRedacted(t, publicExport, importedExport)

	require.NoError(t, workbook.SetCellValue(spokedomain.PresetTemplateSheet, "A1", "wrong-header"))
	var invalidHeaderTemplate bytes.Buffer
	require.NoError(t, workbook.Write(&invalidHeaderTemplate))
	invalidHeaderRequest, err := newMultipartFileRequest(
		http.MethodPost,
		"/api/admin/spoke-catalog/preset-template/import",
		"invalid-header.xlsx",
		invalidHeaderTemplate.Bytes(),
	)
	require.NoError(t, err)
	invalidHeaderRecorder := httptest.NewRecorder()
	router.ServeHTTP(invalidHeaderRecorder, invalidHeaderRequest)
	require.Equal(t, http.StatusBadRequest, invalidHeaderRecorder.Code, invalidHeaderRecorder.Body.String())
	require.Contains(t, invalidHeaderRecorder.Body.String(), "preset template column")
}

func executeSpokeCatalogRequest(t *testing.T, router *gin.Engine, request *http.Request) spokedomain.ExportResponse {
	t.Helper()

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	var export spokedomain.ExportResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &export))
	return export
}

func assertPublicSpokeExportRedacted(t *testing.T, public, full spokedomain.ExportResponse) {
	t.Helper()
	require.Equal(t, full.Options, public.Options)
	require.Equal(t, len(full.Rims), len(public.Rims))
	require.Equal(t, len(full.Hubs), len(public.Hubs))
	require.Equal(t, len(full.Presets), len(public.Presets))
	for i, brand := range public.Rims {
		require.Equal(t, full.Rims[i].ID, brand.ID)
		require.Equal(t, full.Rims[i].Name, brand.Name)
		require.Equal(t, len(full.Rims[i].Items), len(brand.Items))
		for j, model := range brand.Items {
			require.Equal(t, full.Rims[i].Items[j].ID, model.ID)
			require.Equal(t, full.Rims[i].Items[j].Name, model.Name)
			require.Nil(t, model.ERD)
			require.Nil(t, model.Weight)
		}
	}
	for i, brand := range public.Hubs {
		require.Equal(t, full.Hubs[i].ID, brand.ID)
		require.Equal(t, full.Hubs[i].Name, brand.Name)
		require.Equal(t, len(full.Hubs[i].Items), len(brand.Items))
		for j, model := range brand.Items {
			require.Equal(t, full.Hubs[i].Items[j].ID, model.ID)
			require.Equal(t, full.Hubs[i].Items[j].Name, model.Name)
			require.Nil(t, model.Front)
			require.Nil(t, model.Rear)
		}
	}
	for i, preset := range public.Presets {
		require.Equal(t, full.Presets[i].ID, preset.ID)
		require.Equal(t, full.Presets[i].Name, preset.Name)
		require.Equal(t, full.Presets[i].Keywords, preset.Keywords)
		require.Equal(t, full.Presets[i].Description, preset.Description)
		require.Equal(t, full.Presets[i].RimBrandID, preset.RimBrandID)
		require.Equal(t, full.Presets[i].RimModelID, preset.RimModelID)
		require.Equal(t, full.Presets[i].HubBrandID, preset.HubBrandID)
		require.Equal(t, full.Presets[i].HubModelID, preset.HubModelID)
		require.Equal(t, full.Presets[i].WheelPosition, preset.WheelPosition)
		require.Equal(t, full.Presets[i].SpokeCount, preset.SpokeCount)
		require.Equal(t, full.Presets[i].Crossing, preset.Crossing)
		require.Equal(t, full.Presets[i].NippleType, preset.NippleType)
		require.Nil(t, preset.NippleLength)
		require.Nil(t, preset.ActualLengths)
	}
}

func newJSONCatalogRequest(method, path string, payload spokedomain.ExportResponse) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	request := httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	return request, nil
}

func newMultipartCatalogRequest(method, path string, payload spokedomain.ExportResponse) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "spoke-catalog.json")
	if err != nil {
		return nil, err
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(raw); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	request := httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewReader(body.Bytes()))
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request, nil
}

func newMultipartFileRequest(method, path, filename string, raw []byte) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(raw); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	request := httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewReader(body.Bytes()))
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request, nil
}

func spokeCatalogRoundTripFixture() (spokedomain.ExportResponse, spokedomain.ExportResponse) {
	input := spokeCatalogRoundTripSeed()
	frontLeft := 282.0
	rearRight := 284.0

	input.Presets[0].Description = " verified bench build "
	input.Presets[0].Keywords = []string{" 350 ", "350", "DT Swiss", "dt swiss"}
	input.Presets[0].ActualLengths = &spokedomain.WheelBuildActualLengths{
		FrontLeft: &frontLeft,
		RearRight: &rearRight,
		Notes:     " verified bench build ",
	}

	expected := input
	expected.Presets = append([]spokedomain.WheelBuildPreset(nil), input.Presets...)
	expected.Presets[0].Description = "verified bench build"
	expected.Presets[0].Keywords = []string{"350", "DT Swiss"}
	expected.Presets[0].ActualLengths = &spokedomain.WheelBuildActualLengths{
		FrontLeft: &frontLeft,
		RearRight: &rearRight,
		Notes:     "verified bench build",
	}
	for i := range expected.Presets {
		expected.Presets[i].WheelPosition = "auto"
	}

	return input, expected
}

func spokeCatalogRoundTripSeed() spokedomain.ExportResponse {
	erd := 598.0
	left := 22.5
	right := 35.6
	pcd := 44.0
	nippleLength := 14.0

	return spokedomain.ExportResponse{
		Options: spokedomain.DefaultOptions(),
		Rims: []spokedomain.RimBrand{
			{
				ID:   "dt_swiss",
				Name: "DT Swiss",
				Items: []spokedomain.RimModel{
					{ID: "rr411_db", Name: "RR 411 db", ERD: &erd},
				},
			},
		},
		Hubs: []spokedomain.HubBrand{
			{
				ID:   "dt_swiss",
				Name: "DT Swiss",
				Items: []spokedomain.HubModel{
					{
						ID:   "350_road_db_cl",
						Name: "350 Road db CL",
						Front: &spokedomain.HubGeometry{
							LeftFlange:     &left,
							RightFlange:    &right,
							LeftFlangePCD:  &pcd,
							RightFlangePCD: &pcd,
						},
					},
				},
			},
		},
		Presets: []spokedomain.WheelBuildPreset{
			{
				ID:           "dt350_rr411",
				Name:         "DT Swiss RR 411 + 350",
				Description:  "Reference bench build.",
				Keywords:     []string{"350", "DT Swiss"},
				RimBrandID:   "dt_swiss",
				RimModelID:   "rr411_db",
				HubBrandID:   "dt_swiss",
				HubModelID:   "350_road_db_cl",
				SpokeCount:   24,
				Crossing:     2,
				NippleType:   "standard",
				NippleLength: &nippleLength,
			},
		},
	}
}
