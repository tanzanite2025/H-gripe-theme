package service

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	domainspoke "commerce-platform/internal/domain/spoke"

	"github.com/xuri/excelize/v2"
)

const spokePresetTemplateMaxRows = 5000

func (s *SpokeService) BuildPresetTemplate() ([]byte, error) {
	export, err := s.GetExport()
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, domainspoke.PresetTemplateSheet); err != nil {
		return nil, err
	}
	if _, err := file.NewSheet(domainspoke.PresetTemplateInstructionsSheet); err != nil {
		return nil, err
	}
	if _, err := file.NewSheet(domainspoke.PresetTemplateListsSheet); err != nil {
		return nil, err
	}
	if err := file.SetSheetVisible(domainspoke.PresetTemplateListsSheet, false, true); err != nil {
		return nil, err
	}

	headerStyle, err := file.NewStyle(&excelize.Style{
		Font:       &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:       excelize.Fill{Type: "pattern", Color: []string{"1F2937"}, Pattern: 1},
		Alignment:  &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Protection: &excelize.Protection{Locked: true},
	})
	if err != nil {
		return nil, err
	}
	bodyStyle, err := file.NewStyle(&excelize.Style{
		Alignment:  &excelize.Alignment{Vertical: "center", WrapText: true},
		Protection: &excelize.Protection{Locked: false},
	})
	if err != nil {
		return nil, err
	}
	instructionStyle, err := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
	})
	if err != nil {
		return nil, err
	}

	for columnIndex, column := range domainspoke.PresetTemplateColumns {
		cell, err := excelize.CoordinatesToCellName(columnIndex+1, 1)
		if err != nil {
			return nil, err
		}
		if err := file.SetCellValue(domainspoke.PresetTemplateSheet, cell, column); err != nil {
			return nil, err
		}
	}

	if err := file.SetCellStyle(domainspoke.PresetTemplateSheet, "A1", "R1", headerStyle); err != nil {
		return nil, err
	}
	if err := file.SetCellStyle(domainspoke.PresetTemplateSheet, "A2", "R5000", bodyStyle); err != nil {
		return nil, err
	}
	if err := file.SetRowHeight(domainspoke.PresetTemplateSheet, 1, 30); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(domainspoke.PresetTemplateSheet, "A", "A", 22); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(domainspoke.PresetTemplateSheet, "B", "C", 30); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(domainspoke.PresetTemplateSheet, "D", "G", 24); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(domainspoke.PresetTemplateSheet, "H", "K", 18); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(domainspoke.PresetTemplateSheet, "L", "Q", 16); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(domainspoke.PresetTemplateSheet, "R", "R", 30); err != nil {
		return nil, err
	}
	if err := file.AutoFilter(domainspoke.PresetTemplateSheet, "A1:R5000", nil); err != nil {
		return nil, err
	}
	if err := file.SetPanes(domainspoke.PresetTemplateSheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return nil, err
	}

	instructions := [][]interface{}{
		{"说明", "Presets 第 1 行表头已锁定，不能修改；第 2 行开始填写预设。"},
		{"系统字段", "轮圈、花鼓、轮位、辐条数、交叉数、辐条帽类型均使用当前系统字典下拉。"},
		{"导入规则", "模板只导入预设，不会覆盖轮圈、花鼓和计算器选项基础数据；重复 Code 会更新原预设。"},
		{"空值", "实际编制长度、描述、备注和关键词允许为空；不要删除表头或调整列顺序。"},
	}
	for rowIndex, row := range instructions {
		if err := file.SetSheetRow(domainspoke.PresetTemplateInstructionsSheet, fmt.Sprintf("A%d", rowIndex+1), &row); err != nil {
			return nil, err
		}
	}
	if err := file.SetCellStyle(domainspoke.PresetTemplateInstructionsSheet, "A1", "B4", instructionStyle); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(domainspoke.PresetTemplateInstructionsSheet, "A", "A", 14); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(domainspoke.PresetTemplateInstructionsSheet, "B", "B", 100); err != nil {
		return nil, err
	}

	listColumns := map[string][]string{
		"A": catalogOptionValues(export.Rims, func(brand domainspoke.RimBrand) (string, string) {
			return brand.ID, brand.Name
		}),
		"B": catalogModelOptionValues(export.Rims, func(brand domainspoke.RimBrand) []domainspoke.RimModel {
			return brand.Items
		}),
		"C": catalogOptionValues(export.Hubs, func(brand domainspoke.HubBrand) (string, string) {
			return brand.ID, brand.Name
		}),
		"D": catalogModelOptionValues(export.Hubs, func(brand domainspoke.HubBrand) []domainspoke.HubModel {
			return brand.Items
		}),
		"E": stringOptionValues(export.Options.WheelPositions),
		"F": intOptionValues(export.Options.SpokeCounts),
		"G": intOptionValues(export.Options.Crossings),
		"H": stringOptionValues(export.Options.NippleTypes),
	}
	for column, values := range listColumns {
		for rowIndex, value := range values {
			if err := file.SetCellValue(domainspoke.PresetTemplateListsSheet, fmt.Sprintf("%s%d", column, rowIndex+1), value); err != nil {
				return nil, err
			}
		}
		if len(values) == 0 {
			continue
		}
		dataValidation := excelize.NewDataValidation(true)
		dataValidation.Sqref = fmt.Sprintf("%s2:%s%d", templateColumnFor(column), templateColumnFor(column), spokePresetTemplateMaxRows)
		dataValidation.SetSqrefDropList(fmt.Sprintf("'%s'!$%s$1:$%s$%d", domainspoke.PresetTemplateListsSheet, column, column, len(values)))
		dataValidation.SetError(excelize.DataValidationErrorStyleStop, "只能选择系统值", "请从下拉列表中选择当前系统已有的值。")
		dataValidation.SetInput("系统字典", "此字段只能从系统当前数据中选择。")
		if err := file.AddDataValidation(domainspoke.PresetTemplateSheet, dataValidation); err != nil {
			return nil, err
		}
	}

	if err := file.ProtectSheet(domainspoke.PresetTemplateSheet, &excelize.SheetProtectionOptions{
		Password:            "spoke-template",
		AutoFilter:          true,
		SelectUnlockedCells: true,
	}); err != nil {
		return nil, err
	}
	for _, sheet := range []string{
		domainspoke.PresetTemplateInstructionsSheet,
		domainspoke.PresetTemplateListsSheet,
	} {
		if err := file.ProtectSheet(sheet, &excelize.SheetProtectionOptions{
			Password:          "spoke-template",
			SelectLockedCells: true,
		}); err != nil {
			return nil, err
		}
	}

	var output bytes.Buffer
	if err := file.Write(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func (s *SpokeService) ImportPresetTemplate(reader io.Reader) (domainspoke.ExportResponse, error) {
	file, err := excelize.OpenReader(reader)
	if err != nil {
		return domainspoke.ExportResponse{}, fmt.Errorf("%w: cannot open preset template: %v", ErrInvalidSpokeCatalog, err)
	}
	defer func() { _ = file.Close() }()

	rows, err := file.GetRows(domainspoke.PresetTemplateSheet)
	if err != nil {
		return domainspoke.ExportResponse{}, fmt.Errorf("%w: cannot read preset template: %v", ErrInvalidSpokeCatalog, err)
	}
	if len(rows) == 0 {
		return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset template is empty", ErrInvalidSpokeCatalog)
	}
	if err := validatePresetTemplateHeader(rows[0]); err != nil {
		return domainspoke.ExportResponse{}, err
	}

	current, err := s.GetExport()
	if err != nil {
		return domainspoke.ExportResponse{}, err
	}
	presetIndexes := make(map[string]int, len(current.Presets))
	for index, preset := range current.Presets {
		presetIndexes[normalizeCatalogID(preset.ID)] = index
	}

	for rowIndex, row := range rows[1:] {
		if templateRowIsEmpty(row) {
			continue
		}
		rowNumber := rowIndex + 2
		if rowNumber > spokePresetTemplateMaxRows {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: row %d is outside the supported template range", ErrInvalidSpokeCatalog, rowNumber)
		}
		if len(row) > len(domainspoke.PresetTemplateColumns) {
			for _, extra := range row[len(domainspoke.PresetTemplateColumns):] {
				if strings.TrimSpace(extra) != "" {
					return domainspoke.ExportResponse{}, fmt.Errorf("%w: row %d contains unexpected columns", ErrInvalidSpokeCatalog, rowNumber)
				}
			}
			row = row[:len(domainspoke.PresetTemplateColumns)]
		}
		preset, err := parsePresetTemplateRow(row, rowNumber)
		if err != nil {
			return domainspoke.ExportResponse{}, err
		}
		normalizedID := normalizeCatalogID(preset.ID)
		if existingIndex, exists := presetIndexes[normalizedID]; exists {
			current.Presets[existingIndex] = preset
			continue
		}
		presetIndexes[normalizedID] = len(current.Presets)
		current.Presets = append(current.Presets, preset)
	}

	return s.ReplaceCatalog(current)
}

func validatePresetTemplateHeader(row []string) error {
	if len(row) != len(domainspoke.PresetTemplateColumns) {
		return fmt.Errorf("%w: preset template header must contain exactly %d columns", ErrInvalidSpokeCatalog, len(domainspoke.PresetTemplateColumns))
	}
	for index, expected := range domainspoke.PresetTemplateColumns {
		if strings.TrimSpace(row[index]) != expected {
			return fmt.Errorf("%w: preset template column %d must be %q", ErrInvalidSpokeCatalog, index+1, expected)
		}
	}
	return nil
}

func parsePresetTemplateRow(row []string, rowNumber int) (domainspoke.WheelBuildPreset, error) {
	cell := func(index int) string {
		if index >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[index])
	}

	spokeCount, err := parseTemplateInt(cell(8), rowNumber, "spokeCount", true)
	if err != nil {
		return domainspoke.WheelBuildPreset{}, err
	}
	crossing, err := parseTemplateInt(cell(9), rowNumber, "crossing", true)
	if err != nil {
		return domainspoke.WheelBuildPreset{}, err
	}
	nippleLength, err := parseTemplateFloat(cell(11), rowNumber, "nippleLength")
	if err != nil {
		return domainspoke.WheelBuildPreset{}, err
	}
	actualFrontLeft, err := parseTemplateFloat(cell(12), rowNumber, "actualFrontLeft")
	if err != nil {
		return domainspoke.WheelBuildPreset{}, err
	}
	actualFrontRight, err := parseTemplateFloat(cell(13), rowNumber, "actualFrontRight")
	if err != nil {
		return domainspoke.WheelBuildPreset{}, err
	}
	actualRearLeft, err := parseTemplateFloat(cell(14), rowNumber, "actualRearLeft")
	if err != nil {
		return domainspoke.WheelBuildPreset{}, err
	}
	actualRearRight, err := parseTemplateFloat(cell(15), rowNumber, "actualRearRight")
	if err != nil {
		return domainspoke.WheelBuildPreset{}, err
	}

	preset := domainspoke.WheelBuildPreset{
		ID:            templateReferenceID(cell(0)),
		Name:          normalizeTemplateText(cell(1)),
		Description:   normalizeTemplateText(cell(2)),
		RimBrandID:    templateReferenceID(cell(3)),
		RimModelID:    templateReferenceID(cell(4)),
		HubBrandID:    templateReferenceID(cell(5)),
		HubModelID:    templateReferenceID(cell(6)),
		WheelPosition: templateReferenceID(cell(7)),
		SpokeCount:    spokeCount,
		Crossing:      crossing,
		NippleType:    templateReferenceID(cell(10)),
		NippleLength:  nippleLength,
		Keywords:      normalizeKeywords(splitTemplateKeywords(cell(17))),
	}
	if actualFrontLeft != nil || actualFrontRight != nil || actualRearLeft != nil || actualRearRight != nil || normalizeTemplateText(cell(16)) != "" {
		preset.ActualLengths = &domainspoke.WheelBuildActualLengths{
			FrontLeft:  actualFrontLeft,
			FrontRight: actualFrontRight,
			RearLeft:   actualRearLeft,
			RearRight:  actualRearRight,
			Notes:      normalizeTemplateText(cell(16)),
		}
	}
	return preset, nil
}

func parseTemplateInt(value string, rowNumber int, field string, required bool) (int, error) {
	if value == "" && !required {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%w: row %d %s must be an integer", ErrInvalidSpokeCatalog, rowNumber, field)
	}
	return parsed, nil
}

func parseTemplateFloat(value string, rowNumber int, field string) (*float64, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return nil, fmt.Errorf("%w: row %d %s must be a number", ErrInvalidSpokeCatalog, rowNumber, field)
	}
	return &parsed, nil
}

func normalizeTemplateText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func templateReferenceID(value string) string {
	value = normalizeTemplateText(value)
	if separator := strings.Index(value, "|"); separator >= 0 {
		value = value[:separator]
	}
	return normalizeCatalogID(value)
}

func splitTemplateKeywords(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n'
	})
}

func templateRowIsEmpty(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func templateColumnFor(listColumn string) string {
	return map[string]string{
		"A": "D",
		"B": "E",
		"C": "F",
		"D": "G",
		"E": "H",
		"F": "I",
		"G": "J",
		"H": "K",
	}[listColumn]
}

func catalogOptionValues[T any](brands []T, value func(T) (string, string)) []string {
	result := make([]string, 0, len(brands))
	for _, brand := range brands {
		id, name := value(brand)
		result = append(result, templateOptionValue(id, name))
	}
	return result
}

func catalogModelOptionValues[T any, M any](brands []T, models func(T) []M) []string {
	result := make([]string, 0)
	for _, brand := range brands {
		for _, model := range models(brand) {
			switch item := any(model).(type) {
			case domainspoke.RimModel:
				result = append(result, templateOptionValue(item.ID, item.Name))
			case domainspoke.HubModel:
				result = append(result, templateOptionValue(item.ID, item.Name))
			}
		}
	}
	return result
}

func stringOptionValues(options []domainspoke.StringOption) []string {
	result := make([]string, 0, len(options))
	for _, option := range options {
		result = append(result, templateOptionValue(option.Value, option.Label))
	}
	return result
}

func intOptionValues(options []domainspoke.IntOption) []string {
	result := make([]string, 0, len(options))
	for _, option := range options {
		result = append(result, strconv.Itoa(option.Value))
	}
	return result
}

func templateOptionValue(value, label string) string {
	value = normalizeTemplateText(value)
	label = normalizeTemplateText(label)
	if value == label || label == "" {
		return value
	}
	return value + " | " + label
}
