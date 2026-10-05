package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"commerce-platform/internal/domain/shipping"
)

func (s *YanwenWaybillOperationsService) DownloadYanwenWaybillLabel(ctx context.Context, waybillID uint) (YanwenWaybillLabelResult, error) {
	if waybillID == 0 {
		return YanwenWaybillLabelResult{}, errors.New("Yanwen waybill id is required")
	}
	if s == nil || s.waybills == nil {
		return YanwenWaybillLabelResult{}, errors.New("Yanwen waybill repository is not configured")
	}
	if s.gateway == nil {
		return YanwenWaybillLabelResult{}, errors.New("Yanwen gateway client is not configured")
	}
	waybill, err := s.waybills.FindYanwenWaybillByID(waybillID)
	if err != nil {
		return YanwenWaybillLabelResult{}, fmt.Errorf("load Yanwen waybill %d: %w", waybillID, err)
	}
	credentials, err := s.resolveYanwenGatewayCredentials(YanwenAPIConfigInput{Environment: waybill.Environment})
	if err != nil {
		return YanwenWaybillLabelResult{}, err
	}
	officialLabel, err := s.gateway.GetYanwenOrderLabel(ctx, credentials, waybill.WaybillNumber)
	if err != nil {
		return YanwenWaybillLabelResult{}, err
	}
	if officialLabel.WaybillNumber != waybill.WaybillNumber {
		return YanwenWaybillLabelResult{}, fmt.Errorf("Yanwen get order label response waybill number %q does not match local %q", officialLabel.WaybillNumber, waybill.WaybillNumber)
	}
	return YanwenWaybillLabelResult{
		WaybillID:     waybill.ID,
		WaybillNumber: waybill.WaybillNumber,
		FileName:      fmt.Sprintf("yanwen-%s.pdf", waybill.WaybillNumber),
		ContentType:   "application/pdf",
		Base64String:  officialLabel.Base64String,
	}, nil
}

func (s *YanwenWaybillOperationsService) DownloadYanwenWaybillLabelsArchive(
	ctx context.Context,
	waybillIDs []uint,
) (YanwenBatchWaybillLabelArchive, error) {
	if len(waybillIDs) == 0 {
		return YanwenBatchWaybillLabelArchive{}, errors.New("at least one Yanwen waybill id is required")
	}
	if len(waybillIDs) > yanwenGetOrderListMaximumIdentifiers {
		return YanwenBatchWaybillLabelArchive{}, fmt.Errorf("Yanwen batch label download supports at most %d waybills", yanwenGetOrderListMaximumIdentifiers)
	}
	if s == nil || s.waybills == nil {
		return YanwenBatchWaybillLabelArchive{}, errors.New("Yanwen waybill repository is not configured")
	}
	if s.gateway == nil {
		return YanwenBatchWaybillLabelArchive{}, errors.New("Yanwen gateway client is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	seenIDs := make(map[uint]struct{}, len(waybillIDs))
	loadedWaybills := make([]*shipping.YanwenWaybill, 0, len(waybillIDs))
	environment := ""
	for index, waybillID := range waybillIDs {
		if waybillID == 0 {
			return YanwenBatchWaybillLabelArchive{}, fmt.Errorf("Yanwen batch label waybill id at index %d must be a positive integer", index)
		}
		if _, exists := seenIDs[waybillID]; exists {
			return YanwenBatchWaybillLabelArchive{}, fmt.Errorf("Yanwen batch label download contains duplicate waybill id %d", waybillID)
		}
		seenIDs[waybillID] = struct{}{}
		waybill, err := s.waybills.FindYanwenWaybillByID(waybillID)
		if err != nil {
			return YanwenBatchWaybillLabelArchive{}, fmt.Errorf("load Yanwen waybill %d: %w", waybillID, err)
		}
		waybillEnvironment := normalizeYanwenEnvironment(waybill.Environment)
		if !isSupportedYanwenEnvironment(waybillEnvironment) {
			return YanwenBatchWaybillLabelArchive{}, fmt.Errorf("Yanwen waybill %d has unsupported environment %q", waybillID, waybill.Environment)
		}
		if environment == "" {
			environment = waybillEnvironment
		} else if environment != waybillEnvironment {
			return YanwenBatchWaybillLabelArchive{}, errors.New("Yanwen batch label download only accepts waybills from one environment")
		}
		loadedWaybills = append(loadedWaybills, waybill)
	}

	credentials, err := s.resolveYanwenGatewayCredentials(YanwenAPIConfigInput{Environment: environment})
	if err != nil {
		return YanwenBatchWaybillLabelArchive{}, err
	}
	summary := YanwenBatchWaybillLabelResult{
		Items: make([]YanwenBatchWaybillLabelItem, 0, len(loadedWaybills)),
	}
	pdfFiles := make([]struct {
		name string
		data []byte
	}, 0, len(loadedWaybills))
	for _, waybill := range loadedWaybills {
		item := YanwenBatchWaybillLabelItem{
			WaybillID:     waybill.ID,
			WaybillNumber: waybill.WaybillNumber,
			FileName:      yanwenBatchLabelFileName(waybill.ID, waybill.WaybillNumber),
		}
		if err := ctx.Err(); err != nil {
			item.Error = err.Error()
			summary.Failed++
			summary.Items = append(summary.Items, item)
			continue
		}
		officialLabel, err := s.gateway.GetYanwenOrderLabel(ctx, credentials, waybill.WaybillNumber)
		if err != nil {
			item.Error = err.Error()
			summary.Failed++
			summary.Items = append(summary.Items, item)
			continue
		}
		decodedLabel, err := decodeYanwenPDFBase64(officialLabel.Base64String)
		if err != nil {
			item.Error = err.Error()
			summary.Failed++
			summary.Items = append(summary.Items, item)
			continue
		}
		item.Downloaded = true
		summary.Succeeded++
		summary.Items = append(summary.Items, item)
		pdfFiles = append(pdfFiles, struct {
			name string
			data []byte
		}{name: item.FileName, data: decodedLabel})
	}

	archiveBuffer := &bytes.Buffer{}
	archiveWriter := zip.NewWriter(archiveBuffer)
	for _, pdfFile := range pdfFiles {
		fileWriter, err := archiveWriter.Create(pdfFile.name)
		if err != nil {
			_ = archiveWriter.Close()
			return YanwenBatchWaybillLabelArchive{}, fmt.Errorf("create Yanwen label archive entry: %w", err)
		}
		if _, err := fileWriter.Write(pdfFile.data); err != nil {
			_ = archiveWriter.Close()
			return YanwenBatchWaybillLabelArchive{}, fmt.Errorf("write Yanwen label archive entry: %w", err)
		}
	}
	manifest, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		_ = archiveWriter.Close()
		return YanwenBatchWaybillLabelArchive{}, fmt.Errorf("encode Yanwen label archive manifest: %w", err)
	}
	manifestWriter, err := archiveWriter.Create("yanwen-label-manifest.json")
	if err != nil {
		_ = archiveWriter.Close()
		return YanwenBatchWaybillLabelArchive{}, fmt.Errorf("create Yanwen label archive manifest: %w", err)
	}
	if _, err := manifestWriter.Write(manifest); err != nil {
		_ = archiveWriter.Close()
		return YanwenBatchWaybillLabelArchive{}, fmt.Errorf("write Yanwen label archive manifest: %w", err)
	}
	if err := archiveWriter.Close(); err != nil {
		return YanwenBatchWaybillLabelArchive{}, fmt.Errorf("close Yanwen label archive: %w", err)
	}
	return YanwenBatchWaybillLabelArchive{
		FileName:    "yanwen-waybill-labels.zip",
		ContentType: "application/zip",
		Data:        archiveBuffer.Bytes(),
		Summary:     summary,
	}, nil
}

func yanwenBatchLabelFileName(waybillID uint, waybillNumber string) string {
	safeWaybillNumber := strings.Map(func(character rune) rune {
		if unicode.IsLetter(character) || unicode.IsNumber(character) || character == '-' || character == '_' || character == '.' {
			return character
		}
		return '_'
	}, strings.TrimSpace(waybillNumber))
	if safeWaybillNumber == "" {
		safeWaybillNumber = "unknown"
	}
	return fmt.Sprintf("yanwen-waybill-%d-%s.pdf", waybillID, safeWaybillNumber)
}

func decodeYanwenPDFBase64(base64String string) ([]byte, error) {
	decodedLabel, err := base64.StdEncoding.DecodeString(strings.TrimSpace(base64String))
	if err != nil {
		return nil, fmt.Errorf("Yanwen label base64 is invalid: %w", err)
	}
	if len(decodedLabel) == 0 || !bytes.HasPrefix(decodedLabel, []byte("%PDF")) {
		return nil, errors.New("Yanwen label content is not a PDF file")
	}
	return decodedLabel, nil
}
