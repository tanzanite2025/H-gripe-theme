package service

import (
	"context"
	"errors"
	"fmt"

	"commerce-platform/internal/domain/shipping"
)

func (s *YanwenWaybillOperationsService) CancelYanwenWaybill(
	ctx context.Context,
	waybillID uint,
	note string,
) (YanwenWaybillCancellationResult, error) {
	if waybillID == 0 {
		return YanwenWaybillCancellationResult{}, errors.New("Yanwen waybill id is required")
	}
	if s == nil || s.waybills == nil {
		return YanwenWaybillCancellationResult{}, errors.New("Yanwen waybill repository is not configured")
	}
	if s.gateway == nil {
		return YanwenWaybillCancellationResult{}, errors.New("Yanwen gateway client is not configured")
	}
	waybill, err := s.waybills.FindYanwenWaybillByID(waybillID)
	if err != nil {
		return YanwenWaybillCancellationResult{}, fmt.Errorf("load Yanwen waybill %d: %w", waybillID, err)
	}
	credentials, err := s.resolveYanwenGatewayCredentials(YanwenAPIConfigInput{Environment: waybill.Environment})
	if err != nil {
		return YanwenWaybillCancellationResult{}, err
	}
	if _, err := s.gateway.CancelYanwenOrder(ctx, credentials, waybill.WaybillNumber, note); err != nil {
		return YanwenWaybillCancellationResult{}, err
	}

	result := YanwenWaybillCancellationResult{
		Waybill:              *waybill,
		CancellationAccepted: true,
		Message:              "燕文已接受取消请求，但官方状态同步失败，请稍后重新同步官方状态。",
	}
	officialWaybill, syncErr := s.SyncYanwenWaybillOfficialDetails(ctx, waybillID)
	if syncErr != nil {
		return result, nil
	}
	result.Waybill = *officialWaybill
	result.OfficialStatusSynced = true
	if officialWaybill.OfficialStatus == shipping.YanwenOfficialWaybillStatusCancelled {
		result.Message = "燕文已接受取消请求，官方状态已同步为已取消。"
	} else {
		result.Message = "燕文已接受取消请求，但官方状态尚未同步为已取消。"
	}
	return result, nil
}

func (s *YanwenWaybillOperationsService) CancelYanwenWaybills(
	ctx context.Context,
	waybillIDs []uint,
	note string,
) (YanwenBatchWaybillCancellationResult, error) {
	if len(waybillIDs) == 0 {
		return YanwenBatchWaybillCancellationResult{}, errors.New("at least one Yanwen waybill id is required")
	}
	if len(waybillIDs) > yanwenGetOrderListMaximumIdentifiers {
		return YanwenBatchWaybillCancellationResult{}, fmt.Errorf("Yanwen batch cancellation supports at most %d waybills", yanwenGetOrderListMaximumIdentifiers)
	}
	if s == nil || s.waybills == nil {
		return YanwenBatchWaybillCancellationResult{}, errors.New("Yanwen waybill repository is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	seenIDs := make(map[uint]struct{}, len(waybillIDs))
	loadedWaybills := make([]*shipping.YanwenWaybill, 0, len(waybillIDs))
	environment := ""
	for index, waybillID := range waybillIDs {
		if waybillID == 0 {
			return YanwenBatchWaybillCancellationResult{}, fmt.Errorf("Yanwen batch cancellation waybill id at index %d must be a positive integer", index)
		}
		if _, exists := seenIDs[waybillID]; exists {
			return YanwenBatchWaybillCancellationResult{}, fmt.Errorf("Yanwen batch cancellation contains duplicate waybill id %d", waybillID)
		}
		seenIDs[waybillID] = struct{}{}
		waybill, err := s.waybills.FindYanwenWaybillByID(waybillID)
		if err != nil {
			return YanwenBatchWaybillCancellationResult{}, fmt.Errorf("load Yanwen waybill %d: %w", waybillID, err)
		}
		waybillEnvironment := normalizeYanwenEnvironment(waybill.Environment)
		if !isSupportedYanwenEnvironment(waybillEnvironment) {
			return YanwenBatchWaybillCancellationResult{}, fmt.Errorf("Yanwen waybill %d has unsupported environment %q", waybillID, waybill.Environment)
		}
		if environment == "" {
			environment = waybillEnvironment
		} else if environment != waybillEnvironment {
			return YanwenBatchWaybillCancellationResult{}, errors.New("Yanwen batch cancellation only accepts waybills from one environment")
		}
		loadedWaybills = append(loadedWaybills, waybill)
	}

	result := YanwenBatchWaybillCancellationResult{
		Items: make([]YanwenBatchWaybillCancellationItem, 0, len(loadedWaybills)),
	}
	for _, waybill := range loadedWaybills {
		item := YanwenBatchWaybillCancellationItem{
			WaybillID:     waybill.ID,
			WaybillNumber: waybill.WaybillNumber,
		}
		if err := ctx.Err(); err != nil {
			item.Error = err.Error()
			result.Failed++
			result.Items = append(result.Items, item)
			continue
		}
		cancellation, err := s.CancelYanwenWaybill(ctx, waybill.ID, note)
		if err != nil {
			item.Error = err.Error()
			result.Failed++
			result.Items = append(result.Items, item)
			continue
		}
		item.WaybillNumber = cancellation.Waybill.WaybillNumber
		item.CancellationAccepted = cancellation.CancellationAccepted
		item.OfficialStatusSynced = cancellation.OfficialStatusSynced
		item.Message = cancellation.Message
		result.Succeeded++
		result.Items = append(result.Items, item)
	}
	return result, nil
}
