package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"commerce-platform/internal/domain/shipping"
)

func (s *YanwenWaybillOperationsService) ListYanwenWaybills(environment, keyword, status, warehouseCode string) ([]shipping.YanwenWaybill, error) {
	environment = normalizeYanwenEnvironment(environment)
	if environment != "" && !isSupportedYanwenEnvironment(environment) {
		return nil, errors.New("Yanwen environment must be fat or production")
	}
	if s == nil || s.waybills == nil {
		return nil, errors.New("Yanwen waybill repository is not configured")
	}
	return s.waybills.FindYanwenWaybills(environment, keyword, status, warehouseCode)
}

func (s *YanwenWaybillOperationsService) SyncYanwenWaybillOfficialDetails(ctx context.Context, waybillID uint) (*shipping.YanwenWaybill, error) {
	if waybillID == 0 {
		return nil, errors.New("Yanwen waybill id is required")
	}
	if s == nil || s.waybills == nil {
		return nil, errors.New("Yanwen waybill repository is not configured")
	}
	waybill, err := s.waybills.FindYanwenWaybillByID(waybillID)
	if err != nil {
		return nil, fmt.Errorf("load Yanwen waybill %d: %w", waybillID, err)
	}
	credentials, err := s.resolveYanwenGatewayCredentials(YanwenAPIConfigInput{Environment: waybill.Environment})
	if err != nil {
		return nil, err
	}
	officialDetails, err := s.gateway.GetYanwenOrderDetails(ctx, credentials, waybill.WaybillNumber)
	if err != nil {
		return nil, err
	}
	if officialDetails.WaybillNumber != waybill.WaybillNumber {
		return nil, fmt.Errorf("Yanwen get order response waybill number %q does not match local %q", officialDetails.WaybillNumber, waybill.WaybillNumber)
	}
	if officialDetails.OrderNumber != waybill.OrderNumber {
		return nil, fmt.Errorf("Yanwen get order response order number %q does not match local %q", officialDetails.OrderNumber, waybill.OrderNumber)
	}
	syncedAt := time.Now().UTC()
	if err := s.waybills.UpdateYanwenWaybillOfficialDetails(waybill.ID, shipping.YanwenOfficialWaybillDetails{
		ReferenceNumber:   officialDetails.ReferenceNumber,
		YanwenOrderNumber: officialDetails.YanwenOrderNumber,
		OfficialStatus:    officialDetails.OfficialStatus,
		IsPrinted:         officialDetails.IsPrinted,
		SyncedAt:          syncedAt,
		ResponseData:      officialDetails.RawResponse,
	}); err != nil {
		return nil, fmt.Errorf("save Yanwen official waybill details: %w", err)
	}
	waybill.ReferenceNumber = officialDetails.ReferenceNumber
	waybill.YanwenOrderNumber = officialDetails.YanwenOrderNumber
	waybill.OfficialStatus = officialDetails.OfficialStatus
	waybill.IsPrinted = officialDetails.IsPrinted
	waybill.LastOfficialSyncedAt = &syncedAt
	waybill.ResponseData = officialDetails.RawResponse
	return waybill, nil
}

func (s *YanwenWaybillOperationsService) SyncYanwenWaybillsOfficialDetails(ctx context.Context, waybillIDs []uint) ([]shipping.YanwenWaybill, error) {
	if len(waybillIDs) == 0 {
		return nil, errors.New("at least one Yanwen waybill id is required")
	}
	if len(waybillIDs) > yanwenGetOrderListMaximumIdentifiers {
		return nil, fmt.Errorf("Yanwen batch official sync supports at most %d waybills", yanwenGetOrderListMaximumIdentifiers)
	}
	if s == nil || s.waybills == nil {
		return nil, errors.New("Yanwen waybill repository is not configured")
	}

	loadedWaybills := make([]*shipping.YanwenWaybill, 0, len(waybillIDs))
	seenIDs := make(map[uint]struct{}, len(waybillIDs))
	groupedWaybills := make(map[string][]*shipping.YanwenWaybill)
	groupOrder := make([]string, 0, 2)
	for index, waybillID := range waybillIDs {
		if waybillID == 0 {
			return nil, fmt.Errorf("Yanwen waybill id at index %d must be a positive integer", index)
		}
		if _, exists := seenIDs[waybillID]; exists {
			return nil, fmt.Errorf("Yanwen batch official sync contains duplicate waybill id %d", waybillID)
		}
		seenIDs[waybillID] = struct{}{}
		waybill, err := s.waybills.FindYanwenWaybillByID(waybillID)
		if err != nil {
			return nil, fmt.Errorf("load Yanwen waybill %d: %w", waybillID, err)
		}
		if len(loadedWaybills) > 0 && loadedWaybills[0].Environment != waybill.Environment {
			return nil, errors.New("Yanwen batch official sync only accepts waybills from one environment")
		}
		loadedWaybills = append(loadedWaybills, waybill)
		if _, exists := groupedWaybills[waybill.Environment]; !exists {
			groupOrder = append(groupOrder, waybill.Environment)
		}
		groupedWaybills[waybill.Environment] = append(groupedWaybills[waybill.Environment], waybill)
	}

	type officialWaybillUpdate struct {
		waybill *shipping.YanwenWaybill
		details shipping.YanwenOfficialWaybillDetails
	}
	updates := make([]officialWaybillUpdate, 0, len(loadedWaybills))
	for _, environment := range groupOrder {
		group := groupedWaybills[environment]
		credentials, err := s.resolveYanwenGatewayCredentials(YanwenAPIConfigInput{Environment: environment})
		if err != nil {
			return nil, err
		}
		identifiers := make([]string, 0, len(group))
		localByWaybillNumber := make(map[string]*shipping.YanwenWaybill, len(group))
		for _, waybill := range group {
			identifiers = append(identifiers, waybill.WaybillNumber)
			localByWaybillNumber[waybill.WaybillNumber] = waybill
		}
		officialWaybills, err := s.gateway.GetYanwenOrderDetailsBatch(ctx, credentials, identifiers)
		if err != nil {
			return nil, err
		}
		seenOfficialWaybills := make(map[string]struct{}, len(officialWaybills))
		for _, officialWaybill := range officialWaybills {
			localWaybill, exists := localByWaybillNumber[officialWaybill.WaybillNumber]
			if !exists {
				return nil, fmt.Errorf("Yanwen get order list response waybill number %q was not requested", officialWaybill.WaybillNumber)
			}
			if _, exists := seenOfficialWaybills[officialWaybill.WaybillNumber]; exists {
				return nil, fmt.Errorf("Yanwen get order list response contains duplicate waybill number %q", officialWaybill.WaybillNumber)
			}
			seenOfficialWaybills[officialWaybill.WaybillNumber] = struct{}{}
			if officialWaybill.OrderNumber != localWaybill.OrderNumber {
				return nil, fmt.Errorf("Yanwen get order list response order number %q does not match local %q for waybill %q", officialWaybill.OrderNumber, localWaybill.OrderNumber, localWaybill.WaybillNumber)
			}
			updates = append(updates, officialWaybillUpdate{
				waybill: localWaybill,
				details: shipping.YanwenOfficialWaybillDetails{
					ReferenceNumber:   officialWaybill.ReferenceNumber,
					YanwenOrderNumber: officialWaybill.YanwenOrderNumber,
					OfficialStatus:    officialWaybill.OfficialStatus,
					IsPrinted:         officialWaybill.IsPrinted,
					SyncedAt:          time.Now().UTC(),
					ResponseData:      officialWaybill.RawResponse,
				},
			})
		}
		if len(seenOfficialWaybills) != len(group) {
			return nil, fmt.Errorf("Yanwen get order list response returned %d of %d requested waybills for %s", len(seenOfficialWaybills), len(group), environment)
		}
	}

	updatesByID := make(map[uint]shipping.YanwenOfficialWaybillDetails, len(updates))
	updatedByID := make(map[uint]*shipping.YanwenWaybill, len(updates))
	for _, update := range updates {
		updatesByID[update.waybill.ID] = update.details
		updatedByID[update.waybill.ID] = update.waybill
	}
	if err := s.waybills.UpdateYanwenWaybillOfficialDetailsBatch(updatesByID); err != nil {
		return nil, fmt.Errorf("save Yanwen batch official waybill details: %w", err)
	}
	for _, update := range updates {
		syncedAt := update.details.SyncedAt
		update.waybill.ReferenceNumber = update.details.ReferenceNumber
		update.waybill.YanwenOrderNumber = update.details.YanwenOrderNumber
		update.waybill.OfficialStatus = update.details.OfficialStatus
		update.waybill.IsPrinted = update.details.IsPrinted
		update.waybill.LastOfficialSyncedAt = &syncedAt
		update.waybill.ResponseData = update.details.ResponseData
	}

	updatedWaybills := make([]shipping.YanwenWaybill, 0, len(loadedWaybills))
	for _, waybill := range loadedWaybills {
		updatedWaybills = append(updatedWaybills, *updatedByID[waybill.ID])
	}
	return updatedWaybills, nil
}
