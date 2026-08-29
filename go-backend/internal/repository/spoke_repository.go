package repository

import (
	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"
	"commerce-platform/internal/domain/spoke"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type SpokeRepository struct {
	db                      *gorm.DB
	fitmentHubSpecification *FitmentHubSpecificationRepository
}

const (
	fitmentSpokeHubBrandCode = "fitment_catalog"
	fitmentSpokeHubBrandName = "Fitment Catalog"
)

func NewSpokeRepository(
	db *gorm.DB,
	fitmentHubRepositories ...*FitmentHubSpecificationRepository,
) *SpokeRepository {
	var fitmentHubRepository *FitmentHubSpecificationRepository
	if len(fitmentHubRepositories) > 0 {
		fitmentHubRepository = fitmentHubRepositories[0]
	}
	return &SpokeRepository{
		db:                      db,
		fitmentHubSpecification: fitmentHubRepository,
	}
}

func (r *SpokeRepository) ConfigureFitmentHubSpecificationRepository(
	fitmentHubRepository *FitmentHubSpecificationRepository,
) {
	if r == nil {
		return
	}
	r.fitmentHubSpecification = fitmentHubRepository
}

func (r *SpokeRepository) UsesFitmentHubSpecifications() bool {
	return r != nil && r.fitmentHubSpecification != nil
}

func (r *SpokeRepository) CountFitmentHubSpecificationReferences(
	specificationID uint,
	specCode string,
) (int64, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("spoke repository is unavailable")
	}

	normalizedCode := normalizeSpokeCatalogID(specCode)
	query := r.db.
		Table("spoke_build_presets AS preset").
		Joins("JOIN spoke_hub_models AS hub_model ON hub_model.id = preset.hub_model_id").
		Where("preset.deleted_at IS NULL").
		Where("hub_model.deleted_at IS NULL")
	if specificationID > 0 {
		query = query.Where(
			"hub_model.fitment_hub_specification_id = ? OR LOWER(TRIM(hub_model.code)) = ?",
			specificationID,
			normalizedCode,
		)
	} else {
		query = query.Where("LOWER(TRIM(hub_model.code)) = ?", normalizedCode)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SpokeRepository) GetAuthoritativeHubBrands() ([]spoke.HubBrand, bool, error) {
	if !r.UsesFitmentHubSpecifications() {
		return nil, false, nil
	}

	hubSpecifications, err := r.fitmentHubSpecification.ListEnabledForSpokeCatalog()
	if err != nil {
		return nil, true, err
	}
	hubBrands, _, err := fitmentHubBrandsForSpokeCatalog(hubSpecifications)
	if err != nil {
		return nil, true, err
	}
	return hubBrands, true, nil
}

func (r *SpokeRepository) GetCatalogExport() (spoke.ExportResponse, bool, error) {
	if r == nil || r.db == nil {
		return spoke.ExportResponse{}, false, errors.New("spoke repository is unavailable")
	}

	var rimBrandCount int64
	if err := r.db.Model(&spoke.CatalogRimBrand{}).Count(&rimBrandCount).Error; err != nil {
		return spoke.ExportResponse{}, false, err
	}
	var presetCount int64
	if err := r.db.Model(&spoke.CatalogBuildPreset{}).Count(&presetCount).Error; err != nil {
		return spoke.ExportResponse{}, false, err
	}

	var authoritativeHubBrands []spoke.HubBrand
	activeFitmentHubModelIDs := make(map[string]struct{})
	activeFitmentHubSpecificationsByID := make(map[uint]fitmentcatalogdomain.HubSpecification)
	if r.UsesFitmentHubSpecifications() {
		hubSpecifications, err := r.fitmentHubSpecification.ListEnabledForSpokeCatalog()
		if err != nil {
			return spoke.ExportResponse{}, false, err
		}
		for _, specification := range hubSpecifications {
			activeFitmentHubSpecificationsByID[specification.ID] = specification
		}
		authoritativeHubBrands, activeFitmentHubModelIDs, err = fitmentHubBrandsForSpokeCatalog(hubSpecifications)
		if err != nil {
			return spoke.ExportResponse{}, false, err
		}
		if rimBrandCount == 0 && presetCount == 0 && len(authoritativeHubBrands) == 0 {
			return emptySpokeExport(), true, nil
		}
	} else {
		var hubBrandCount int64
		if err := r.db.Model(&spoke.CatalogHubBrand{}).Count(&hubBrandCount).Error; err != nil {
			return spoke.ExportResponse{}, false, err
		}
		if rimBrandCount == 0 && hubBrandCount == 0 && presetCount == 0 {
			return spoke.ExportResponse{}, false, nil
		}
	}

	var rimBrands []spoke.CatalogRimBrand
	if err := r.db.
		Where("is_enabled = ?", true).
		Order("sort_order ASC, name ASC, id ASC").
		Preload("Models", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_enabled = ?", true).Order("sort_order ASC, name ASC, id ASC")
		}).
		Find(&rimBrands).Error; err != nil {
		return spoke.ExportResponse{}, true, err
	}

	var hubBrands []spoke.CatalogHubBrand
	if !r.UsesFitmentHubSpecifications() {
		if err := r.db.
			Where("is_enabled = ?", true).
			Order("sort_order ASC, name ASC, id ASC").
			Preload("Models", func(db *gorm.DB) *gorm.DB {
				return db.Where("is_enabled = ?", true).Order("sort_order ASC, name ASC, id ASC")
			}).
			Find(&hubBrands).Error; err != nil {
			return spoke.ExportResponse{}, true, err
		}
	}

	export := spoke.ExportResponse{
		Options: spoke.DefaultOptions(),
		Rims:    make([]spoke.RimBrand, 0, len(rimBrands)),
		Hubs:    make([]spoke.HubBrand, 0, len(hubBrands)),
	}
	rimBrandCodesByID := make(map[uint]string)
	rimModelCodesByID := make(map[uint]string)
	hubBrandCodesByID := make(map[uint]string)
	hubModelCodesByID := make(map[uint]string)
	legacyHubModelsByID := make(map[uint]spoke.CatalogHubModel)

	for _, brand := range rimBrands {
		rimBrandCodesByID[brand.ID] = brand.Code
		items := make([]spoke.RimModel, 0, len(brand.Models))
		for _, model := range brand.Models {
			rimModelCodesByID[model.ID] = model.Code
			items = append(items, spoke.RimModel{
				ID:     model.Code,
				Name:   model.Name,
				ERD:    model.ERD,
				Weight: model.Weight,
			})
		}
		export.Rims = append(export.Rims, spoke.RimBrand{
			ID:    brand.Code,
			Name:  brand.Name,
			Items: items,
		})
	}

	if r.UsesFitmentHubSpecifications() {
		export.Hubs = authoritativeHubBrands
		if presetCount > 0 {
			var legacyHubModels []spoke.CatalogHubModel
			if err := r.db.Select("id, code, fitment_hub_specification_id").Find(&legacyHubModels).Error; err != nil {
				return spoke.ExportResponse{}, true, err
			}
			for _, model := range legacyHubModels {
				legacyHubModelsByID[model.ID] = model
				hubModelCodesByID[model.ID] = model.Code
			}
		}
	} else {
		for _, brand := range hubBrands {
			hubBrandCodesByID[brand.ID] = brand.Code
			items := make([]spoke.HubModel, 0, len(brand.Models))
			for _, model := range brand.Models {
				hubModelCodesByID[model.ID] = model.Code
				items = append(items, spoke.HubModel{
					ID:    model.Code,
					Name:  model.Name,
					Front: catalogHubGeometry(model.FrontLeftFlange, model.FrontRightFlange, model.FrontLeftFlangePCD, model.FrontRightFlangePCD, model.FrontSpokeHoleDiameter),
					Rear:  catalogHubGeometry(model.RearLeftFlange, model.RearRightFlange, model.RearLeftFlangePCD, model.RearRightFlangePCD, model.RearSpokeHoleDiameter),
				})
			}
			export.Hubs = append(export.Hubs, spoke.HubBrand{
				ID:    brand.Code,
				Name:  brand.Name,
				Items: items,
			})
		}
	}

	var presets []spoke.CatalogBuildPreset
	if err := r.db.
		Where("is_enabled = ?", true).
		Order("sort_order ASC, name ASC, id ASC").
		Find(&presets).Error; err != nil {
		return spoke.ExportResponse{}, true, err
	}

	export.Presets = make([]spoke.WheelBuildPreset, 0, len(presets))
	for _, preset := range presets {
		rimBrandCode, hasRimBrand := rimBrandCodesByID[preset.RimBrandID]
		rimModelCode, hasRimModel := rimModelCodesByID[preset.RimModelID]
		hubBrandCode, hasHubBrand := hubBrandCodesByID[preset.HubBrandID]
		hubModelCode, hasHubModel := hubModelCodesByID[preset.HubModelID]
		if !hasRimBrand || !hasRimModel || !hasHubModel {
			continue
		}
		if r.UsesFitmentHubSpecifications() {
			legacyHubModel := legacyHubModelsByID[preset.HubModelID]
			if legacyHubModel.FitmentHubSpecificationID != nil {
				specification, exists := activeFitmentHubSpecificationsByID[*legacyHubModel.FitmentHubSpecificationID]
				if !exists {
					continue
				}
				hubModelCode = normalizeSpokeCatalogID(specification.SpecCode)
			}
			hubModelCode = strings.ToLower(strings.TrimSpace(hubModelCode))
			if _, exists := activeFitmentHubModelIDs[hubModelCode]; !exists {
				continue
			}
			hubBrandCode = fitmentSpokeHubBrandCode
		} else if !hasHubBrand {
			continue
		}
		export.Presets = append(export.Presets, spoke.WheelBuildPreset{
			ID:            preset.Code,
			Name:          preset.Name,
			Keywords:      decodeKeywords(preset.KeywordsJSON),
			Description:   preset.Description,
			RimBrandID:    rimBrandCode,
			RimModelID:    rimModelCode,
			HubBrandID:    hubBrandCode,
			HubModelID:    hubModelCode,
			WheelPosition: preset.WheelPosition,
			SpokeCount:    preset.SpokeCount,
			Crossing:      preset.Crossing,
			NippleType:    preset.NippleType,
			NippleLength:  preset.NippleLength,
			ActualLengths: catalogActualLengths(preset),
		})
	}

	return export, true, nil
}

func (r *SpokeRepository) ReplaceCatalog(export spoke.ExportResponse) error {
	if r == nil || r.db == nil {
		return errors.New("spoke repository is unavailable")
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		effectiveExport := export
		if r.UsesFitmentHubSpecifications() {
			fitmentRepository := r.fitmentHubSpecification.WithTx(tx)
			hubSpecifications, err := fitmentRepository.ListEnabledForSpokeCatalog()
			if err != nil {
				return err
			}
			effectiveExport.Hubs, _, err = fitmentHubBrandsForSpokeCatalog(hubSpecifications)
			if err != nil {
				return err
			}
		}

		if err := tx.Unscoped().Where("1 = 1").Delete(&spoke.CatalogBuildPreset{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("1 = 1").Delete(&spoke.CatalogRimModel{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("1 = 1").Delete(&spoke.CatalogRimBrand{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("1 = 1").Delete(&spoke.CatalogHubModel{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("1 = 1").Delete(&spoke.CatalogHubBrand{}).Error; err != nil {
			return err
		}

		rimBrandIDs := make(map[string]uint)
		rimModelIDs := make(map[string]uint)
		hubBrandIDs := make(map[string]uint)
		hubModelIDs := make(map[string]uint)

		for brandIndex, brand := range effectiveExport.Rims {
			record := spoke.CatalogRimBrand{
				Code:      brand.ID,
				Name:      brand.Name,
				SortOrder: brandIndex,
				IsEnabled: true,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			rimBrandIDs[brand.ID] = record.ID

			for modelIndex, model := range brand.Items {
				modelRecord := spoke.CatalogRimModel{
					BrandID:   record.ID,
					Code:      model.ID,
					Name:      model.Name,
					ERD:       model.ERD,
					Weight:    model.Weight,
					SortOrder: modelIndex,
					IsEnabled: true,
				}
				if err := tx.Create(&modelRecord).Error; err != nil {
					return err
				}
				rimModelIDs[model.ID] = modelRecord.ID
			}
		}

		for brandIndex, brand := range effectiveExport.Hubs {
			record := spoke.CatalogHubBrand{
				Code:      brand.ID,
				Name:      brand.Name,
				SortOrder: brandIndex,
				IsEnabled: true,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			hubBrandIDs[brand.ID] = record.ID

			for modelIndex, model := range brand.Items {
				modelRecord := spoke.CatalogHubModel{
					BrandID:                   record.ID,
					Code:                      model.ID,
					Name:                      model.Name,
					SortOrder:                 modelIndex,
					IsEnabled:                 true,
					FitmentHubSpecificationID: model.FitmentHubSpecificationID,
				}
				assignCatalogGeometry(&modelRecord, model.Front, "front")
				assignCatalogGeometry(&modelRecord, model.Rear, "rear")
				if err := tx.Create(&modelRecord).Error; err != nil {
					return err
				}
				hubModelIDs[model.ID] = modelRecord.ID
			}
		}

		for presetIndex, preset := range effectiveExport.Presets {
			keywords, err := json.Marshal(preset.Keywords)
			if err != nil {
				return err
			}
			rimBrandID, hasRimBrand := rimBrandIDs[preset.RimBrandID]
			rimModelID, hasRimModel := rimModelIDs[preset.RimModelID]
			hubBrandID, hasHubBrand := hubBrandIDs[preset.HubBrandID]
			hubModelID, hasHubModel := hubModelIDs[preset.HubModelID]
			if !hasRimBrand || !hasRimModel || !hasHubBrand || !hasHubModel {
				return fmt.Errorf("spoke preset %q references a catalog item that is not available", preset.ID)
			}
			record := spoke.CatalogBuildPreset{
				Code:          preset.ID,
				Name:          preset.Name,
				Description:   preset.Description,
				KeywordsJSON:  string(keywords),
				RimBrandID:    rimBrandID,
				RimModelID:    rimModelID,
				HubBrandID:    hubBrandID,
				HubModelID:    hubModelID,
				WheelPosition: preset.WheelPosition,
				SpokeCount:    preset.SpokeCount,
				Crossing:      preset.Crossing,
				NippleType:    preset.NippleType,
				NippleLength:  preset.NippleLength,
				SortOrder:     presetIndex,
				IsEnabled:     true,
			}
			assignCatalogActualLengths(&record, preset.ActualLengths)
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *SpokeRepository) ListHistory(search string, page, pageSize int) ([]spoke.History, int64, error) {
	return r.listHistory(search, page, pageSize, nil)
}

func (r *SpokeRepository) ListHistoryByUserID(userID uint, search string, page, pageSize int) ([]spoke.History, int64, error) {
	return r.listHistory(search, page, pageSize, &userID)
}

func (r *SpokeRepository) listHistory(search string, page, pageSize int, userID *uint) ([]spoke.History, int64, error) {
	var items []spoke.History
	var total int64

	query := r.db.Model(&spoke.History{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	search = strings.ToLower(strings.TrimSpace(search))
	if search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"LOWER(COALESCE(rim_brand, '')) LIKE ? OR LOWER(COALESCE(rim_model, '')) LIKE ? OR LOWER(COALESCE(hub_brand, '')) LIKE ? OR LOWER(COALESCE(hub_model, '')) LIKE ?",
			like,
			like,
			like,
			like,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func catalogHubGeometry(leftFlange, rightFlange, leftFlangePCD, rightFlangePCD, spokeHoleDiameter *float64) *spoke.HubGeometry {
	if leftFlange == nil && rightFlange == nil && leftFlangePCD == nil && rightFlangePCD == nil && spokeHoleDiameter == nil {
		return nil
	}
	return &spoke.HubGeometry{
		LeftFlange:        leftFlange,
		RightFlange:       rightFlange,
		LeftFlangePCD:     leftFlangePCD,
		RightFlangePCD:    rightFlangePCD,
		SpokeHoleDiameter: spokeHoleDiameter,
	}
}

func assignCatalogGeometry(model *spoke.CatalogHubModel, geometry *spoke.HubGeometry, position string) {
	if geometry == nil {
		return
	}

	switch position {
	case "front":
		model.FrontLeftFlange = geometry.LeftFlange
		model.FrontRightFlange = geometry.RightFlange
		model.FrontLeftFlangePCD = geometry.LeftFlangePCD
		model.FrontRightFlangePCD = geometry.RightFlangePCD
		model.FrontSpokeHoleDiameter = geometry.SpokeHoleDiameter
	case "rear":
		model.RearLeftFlange = geometry.LeftFlange
		model.RearRightFlange = geometry.RightFlange
		model.RearLeftFlangePCD = geometry.LeftFlangePCD
		model.RearRightFlangePCD = geometry.RightFlangePCD
		model.RearSpokeHoleDiameter = geometry.SpokeHoleDiameter
	}
}

func catalogActualLengths(preset spoke.CatalogBuildPreset) *spoke.WheelBuildActualLengths {
	notes := strings.TrimSpace(preset.ActualLengthNotes)
	if preset.ActualFrontLeftLength == nil &&
		preset.ActualFrontRightLength == nil &&
		preset.ActualRearLeftLength == nil &&
		preset.ActualRearRightLength == nil &&
		notes == "" {
		return nil
	}
	return &spoke.WheelBuildActualLengths{
		FrontLeft:  preset.ActualFrontLeftLength,
		FrontRight: preset.ActualFrontRightLength,
		RearLeft:   preset.ActualRearLeftLength,
		RearRight:  preset.ActualRearRightLength,
		Notes:      notes,
	}
}

func assignCatalogActualLengths(record *spoke.CatalogBuildPreset, actual *spoke.WheelBuildActualLengths) {
	if actual == nil {
		return
	}

	record.ActualFrontLeftLength = actual.FrontLeft
	record.ActualFrontRightLength = actual.FrontRight
	record.ActualRearLeftLength = actual.RearLeft
	record.ActualRearRightLength = actual.RearRight
	record.ActualLengthNotes = strings.TrimSpace(actual.Notes)
}

func decodeKeywords(value string) []string {
	var keywords []string
	if err := json.Unmarshal([]byte(value), &keywords); err != nil {
		return []string{}
	}
	return keywords
}

func fitmentHubBrandsForSpokeCatalog(
	specifications []fitmentcatalogdomain.HubSpecification,
) ([]spoke.HubBrand, map[string]struct{}, error) {
	activeModelIDs := make(map[string]struct{}, len(specifications))
	if len(specifications) == 0 {
		return []spoke.HubBrand{}, activeModelIDs, nil
	}

	brand := spoke.HubBrand{
		ID:    fitmentSpokeHubBrandCode,
		Name:  fitmentSpokeHubBrandName,
		Items: make([]spoke.HubModel, 0, len(specifications)),
	}
	for _, specification := range specifications {
		modelID := normalizeSpokeCatalogID(specification.SpecCode)
		if modelID == "" {
			return nil, nil, fmt.Errorf("fitment hub specification %d has an empty spoke catalog model id", specification.ID)
		}
		if !fitmentcatalogdomain.IsValidSpokeCatalogSpecCode(specification.SpecCode) {
			return nil, nil, fmt.Errorf("fitment hub specification %q has a code that is invalid for the spoke catalog", specification.SpecCode)
		}
		displayName := strings.TrimSpace(specification.DisplayName)
		if displayName == "" {
			return nil, nil, fmt.Errorf("fitment hub specification %q has an empty display name", modelID)
		}
		if _, exists := activeModelIDs[modelID]; exists {
			return nil, nil, fmt.Errorf("fitment hub specifications contain duplicate spoke catalog model id %q", modelID)
		}
		activeModelIDs[modelID] = struct{}{}

		sourceID := specification.ID
		model := spoke.HubModel{
			ID:                        modelID,
			Name:                      displayName,
			FitmentHubSpecificationID: &sourceID,
		}
		geometry := fitmentHubGeometry(specification)
		switch specification.Position {
		case fitmentcatalogdomain.HubPositionFront:
			model.Front = geometry
		case fitmentcatalogdomain.HubPositionRear:
			model.Rear = geometry
		default:
			return nil, nil, fmt.Errorf("fitment hub specification %q has unsupported position %q", modelID, specification.Position)
		}
		brand.Items = append(brand.Items, model)
	}

	return []spoke.HubBrand{brand}, activeModelIDs, nil
}

func normalizeSpokeCatalogID(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func fitmentHubGeometry(specification fitmentcatalogdomain.HubSpecification) *spoke.HubGeometry {
	if specification.WRMM == nil &&
		specification.WLMM == nil &&
		specification.PCDRMM == nil &&
		specification.PCDLMM == nil {
		return nil
	}
	return &spoke.HubGeometry{
		LeftFlange:     specification.WLMM,
		RightFlange:    specification.WRMM,
		LeftFlangePCD:  specification.PCDLMM,
		RightFlangePCD: specification.PCDRMM,
	}
}

func emptySpokeExport() spoke.ExportResponse {
	return spoke.ExportResponse{
		Options: spoke.DefaultOptions(),
		Rims:    []spoke.RimBrand{},
		Hubs:    []spoke.HubBrand{},
		Presets: []spoke.WheelBuildPreset{},
	}
}
