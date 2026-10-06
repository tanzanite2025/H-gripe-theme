package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

func (s *YanwenWaybillOperationsService) CreateYanwenWaybills(
	ctx context.Context,
	requests []YanwenCreateWaybillInput,
) (YanwenBatchWaybillCreationResult, error) {
	normalizedRequests, environment, err := validateYanwenBatchWaybillCreationRequests(requests)
	if err != nil {
		return YanwenBatchWaybillCreationResult{}, err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	result := YanwenBatchWaybillCreationResult{
		Items: make([]YanwenBatchWaybillCreationItem, 0, len(normalizedRequests)),
	}
	for index, request := range normalizedRequests {
		item := YanwenBatchWaybillCreationItem{
			RequestIndex: index,
			Environment:  environment,
			OrderID:      request.OrderID,
			ChannelID:    request.ChannelID,
		}
		if err := ctx.Err(); err != nil {
			item.Error = err.Error()
			result.Failed++
			result.Items = append(result.Items, item)
			continue
		}
		waybill, err := s.CreateYanwenWaybill(ctx, request)
		if err != nil {
			item.Error = err.Error()
			result.Failed++
			result.Items = append(result.Items, item)
			continue
		}
		item.WaybillID = waybill.ID
		item.OrderNumber = waybill.OrderNumber
		item.WaybillNumber = waybill.WaybillNumber
		result.Succeeded++
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func validateYanwenBatchWaybillCreationRequests(
	requests []YanwenCreateWaybillInput,
) ([]YanwenCreateWaybillInput, string, error) {
	if len(requests) == 0 {
		return nil, "", errors.New("at least one Yanwen waybill creation request is required")
	}
	if len(requests) > yanwenBatchWaybillCreationMaximumRequests {
		return nil, "", fmt.Errorf("Yanwen batch waybill creation supports at most %d requests", yanwenBatchWaybillCreationMaximumRequests)
	}

	normalizedRequests := make([]YanwenCreateWaybillInput, len(requests))
	seenOrderChannels := make(map[string]struct{}, len(requests))
	environment := ""
	for index, request := range requests {
		request.Environment = normalizeYanwenEnvironment(request.Environment)
		if !isSupportedYanwenEnvironment(request.Environment) {
			return nil, "", fmt.Errorf("Yanwen batch waybill creation request at index %d must use fat or production environment", index)
		}
		if environment == "" {
			environment = request.Environment
		} else if environment != request.Environment {
			return nil, "", errors.New("Yanwen batch waybill creation only accepts requests from one environment")
		}
		if request.OrderID == 0 || request.ChannelID == 0 {
			return nil, "", fmt.Errorf("Yanwen batch waybill creation request at index %d requires order_id and channel_id", index)
		}
		if request.HasBattery == nil {
			return nil, "", fmt.Errorf("Yanwen batch waybill creation request at index %d requires has_battery", index)
		}
		if strings.TrimSpace(request.WarehouseCode) == "" {
			return nil, "", fmt.Errorf("Yanwen batch waybill creation request at index %d requires warehouse_code", index)
		}
		if len(strings.TrimSpace(request.OrderSource)) > 20 {
			return nil, "", fmt.Errorf("Yanwen batch waybill creation request at index %d has an order_source longer than 20 characters", index)
		}
		if len(strings.TrimSpace(request.ReceiverTaxNumber)) > maximumYanwenReceiverTaxNumberLength {
			return nil, "", fmt.Errorf("Yanwen batch waybill creation request at index %d has a receiver tax number longer than %d characters", index, maximumYanwenReceiverTaxNumberLength)
		}
		if len(strings.TrimSpace(request.IOSS)) > maximumYanwenIOSSLength {
			return nil, "", fmt.Errorf("Yanwen batch waybill creation request at index %d has an IOSS longer than %d characters", index, maximumYanwenIOSSLength)
		}
		if len(strings.TrimSpace(request.EORI)) > maximumYanwenEORILength {
			return nil, "", fmt.Errorf("Yanwen batch waybill creation request at index %d has an EORI longer than %d characters", index, maximumYanwenEORILength)
		}
		key := fmt.Sprintf("%d:%d", request.OrderID, request.ChannelID)
		if _, exists := seenOrderChannels[key]; exists {
			return nil, "", fmt.Errorf("Yanwen batch waybill creation contains duplicate order_id %d and channel_id %d", request.OrderID, request.ChannelID)
		}
		seenOrderChannels[key] = struct{}{}
		normalizedRequests[index] = request
	}
	return normalizedRequests, environment, nil
}

func (s *YanwenWaybillOperationsService) CreateYanwenWaybill(ctx context.Context, input YanwenCreateWaybillInput) (*shipping.YanwenWaybill, error) {
	input.Environment = normalizeYanwenEnvironment(input.Environment)
	if !isSupportedYanwenEnvironment(input.Environment) {
		return nil, errors.New("Yanwen environment must be fat or production")
	}
	if input.OrderID == 0 || input.ChannelID == 0 {
		return nil, errors.New("Yanwen order and enabled channel are required")
	}
	if input.HasBattery == nil {
		return nil, errors.New("has_battery must be explicitly provided")
	}
	if s == nil || s.orders == nil || s.publishedChannels == nil || s.waybills == nil || s.countries == nil || s.warehouses == nil {
		return nil, errors.New("Yanwen waybill dependencies are not configured")
	}
	if s.gateway == nil {
		return nil, errors.New("Yanwen gateway client is not configured")
	}
	if s.products == nil {
		return nil, errors.New("Yanwen product catalog repository is not configured")
	}
	var channel *shipping.YanwenPublishedChannel
	var err error
	channel, err = s.publishedChannels.FindYanwenPublishedChannelByID(input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("load Yanwen published channel: %w", err)
	}
	if !channel.Enabled {
		return nil, errors.New("Yanwen channel is disabled in the curated service collection")
	}
	if err := channel.Validate(); err != nil {
		return nil, err
	}
	if channel.Environment != input.Environment {
		return nil, fmt.Errorf("Yanwen channel %q belongs to the %s environment", channel.ProductCode, channel.Environment)
	}
	if err := validateYanwenChannelCustomsDeclarationRequirements(*channel, YanwenCustomsDeclarationInput{
		ReceiverTaxNumber: input.ReceiverTaxNumber,
		IOSS:              input.IOSS,
		EORI:              input.EORI,
	}); err != nil {
		return nil, err
	}
	products, err := s.products.FindYanwenProductCatalogEntriesByEnvironment(input.Environment)
	if err != nil {
		return nil, fmt.Errorf("load Yanwen products: %w", err)
	}
	productFound := false
	for _, product := range products {
		if strings.EqualFold(strings.TrimSpace(product.ProductID), channel.ProductCode) {
			productFound = true
			break
		}
	}
	if !productFound {
		return nil, fmt.Errorf("Yanwen product %q is not present in the official %s catalog", channel.ProductCode, input.Environment)
	}
	warehouseCode := strings.TrimSpace(input.WarehouseCode)
	if warehouseCode == "" {
		return nil, errors.New("Yanwen warehouse code is required")
	}
	warehouses, err := s.warehouses.FindYanwenWarehouseCatalogEntriesByEnvironment(input.Environment)
	if err != nil {
		return nil, fmt.Errorf("load Yanwen warehouses: %w", err)
	}
	warehouseFound := false
	for _, warehouse := range warehouses {
		if strings.EqualFold(strings.TrimSpace(warehouse.WarehouseCode), warehouseCode) {
			warehouseFound = true
			warehouseCode = warehouse.WarehouseCode
			break
		}
	}
	if !warehouseFound {
		return nil, fmt.Errorf("Yanwen warehouse %q is not present in the official %s catalog", warehouseCode, input.Environment)
	}
	orderFacts, err := s.orders.FindYanwenOrderFactsByID(input.OrderID)
	if err != nil {
		return nil, fmt.Errorf("load order %d: %w", input.OrderID, err)
	}
	if orderFacts == nil {
		return nil, fmt.Errorf("load order %d: Yanwen order facts are empty", input.OrderID)
	}
	if orderFacts.PaymentStatus != "paid" {
		return nil, errors.New("only paid orders can create a Yanwen waybill")
	}
	if orderFacts.Status != "paid" && orderFacts.Status != "processing" {
		return nil, fmt.Errorf("order status %q cannot create a Yanwen waybill", orderFacts.Status)
	}
	if existing, existingErr := s.waybills.FindYanwenWaybillByEnvironmentAndOrderIDAndProductCode(input.Environment, orderFacts.ID, channel.ProductCode); existingErr == nil {
		if input.Environment == "production" && s.trackingSnapshots != nil {
			if err := s.trackingSnapshots.RegisterProductionTrackingNumbers(
				ctx,
				[]string{existing.WaybillNumber},
				time.Now().UTC(),
			); err != nil {
				return nil, fmt.Errorf("register existing Yanwen tracking polling target: %w", err)
			}
		}
		return existing, nil
	} else if !repository.IsRecordNotFound(existingErr) {
		return nil, fmt.Errorf("check existing Yanwen waybill: %w", existingErr)
	}
	countryCode, err := s.resolveYanwenOrderCountry(input.Environment, orderFacts.ShippingAddress.Country)
	if err != nil {
		return nil, err
	}
	if err := validateYanwenChannelCountryScope(channel.Countries, countryCode, orderFacts.ShippingAddress.Country); err != nil {
		return nil, err
	}
	request, description, totalQuantity, totalWeight, err := buildYanwenCreateOrderRequestFromFacts(
		orderFacts,
		channel.ProductCode,
		warehouseCode,
		input.OrderSource,
		*input.HasBattery,
		countryCode,
		YanwenCustomsDeclarationInput{
			ReceiverTaxNumber: input.ReceiverTaxNumber,
			IOSS:              input.IOSS,
			EORI:              input.EORI,
		},
	)
	if err != nil {
		return nil, err
	}
	credentials, err := s.resolveYanwenGatewayCredentials(YanwenAPIConfigInput{Environment: input.Environment})
	if err != nil {
		return nil, err
	}
	if err := s.verifyYanwenWaybillCustomsPreflight(ctx, credentials, countryCode, request); err != nil {
		return nil, err
	}
	remoteResult, err := s.gateway.CreateYanwenOrder(ctx, credentials, request)
	if err != nil {
		return nil, err
	}
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encode Yanwen request snapshot: %w", err)
	}
	waybill := &shipping.YanwenWaybill{
		Environment:         input.Environment,
		OrderID:             orderFacts.ID,
		OrderNumber:         orderFacts.OrderNumber,
		ProductCode:         channel.ProductCode,
		ChannelName:         channel.DisplayName,
		WarehouseCode:       warehouseCode,
		DestinationCountry:  countryCode,
		ConsigneeName:       strings.TrimSpace(orderFacts.ShippingAddress.FirstName + " " + orderFacts.ShippingAddress.LastName),
		DeclaredDescription: description,
		TotalQuantity:       totalQuantity,
		TotalWeightGrams:    totalWeight,
		WaybillNumber:       remoteResult.WaybillNumber,
		YanwenOrderNumber:   remoteResult.YanwenOrderNumber,
		Status:              shipping.YanwenWaybillStatusCreated,
		RequestData:         requestData,
		ResponseData:        remoteResult.RawResponse,
	}
	if err := s.waybills.CreateYanwenWaybill(waybill); err != nil {
		return nil, fmt.Errorf("save Yanwen waybill after official success: %w", err)
	}
	if input.Environment == "production" && s.trackingSnapshots != nil {
		if err := s.trackingSnapshots.RegisterProductionTrackingNumbers(
			ctx,
			[]string{waybill.WaybillNumber},
			time.Now().UTC(),
		); err != nil {
			return nil, fmt.Errorf("register Yanwen tracking polling target: %w", err)
		}
	}
	return waybill, nil
}

func (s *YanwenWaybillOperationsService) verifyYanwenWaybillCustomsPreflight(
	ctx context.Context,
	credentials yanwenGatewayCredentials,
	countryCode string,
	request YanwenCreateOrderRequest,
) error {
	switch strings.ToUpper(strings.TrimSpace(countryCode)) {
	case "KR":
		if s == nil || s.gateway == nil {
			return errors.New("Yanwen customs preflight gateway is not configured")
		}
		receiver := request.ReceiverInfo
		if strings.TrimSpace(receiver.TaxNumber) == "" {
			return errors.New("Korea PCCC tax number is required before Yanwen waybill creation")
		}
		if !isFiveDigitYanwenPostalCode(strings.TrimSpace(receiver.ZipCode)) {
			return errors.New("Korea PCCC postal code must contain exactly 5 digits before Yanwen waybill creation")
		}
		if _, err := s.gateway.VerifyYanwenKoreaPersonalCustomsClearanceCode(ctx, credentials, YanwenKoreaPersonalCustomsClearanceCodeVerificationRequest{
			ReceiverInfo: YanwenKoreaPersonalCustomsClearanceCodeReceiverInfo{
				Name:      receiver.Name,
				Phone:     receiver.Phone,
				TaxNumber: receiver.TaxNumber,
				ZipCode:   receiver.ZipCode,
			},
		}); err != nil {
			return fmt.Errorf("verify Yanwen Korea PCCC before waybill creation: %w", err)
		}
	case "US":
		if s == nil || s.gateway == nil {
			return errors.New("Yanwen customs preflight gateway is not configured")
		}
		receiver := request.ReceiverInfo
		if _, err := s.gateway.VerifyYanwenUnitedStatesAddress(ctx, credentials, YanwenUnitedStatesAddressVerificationRequest{
			ReceiverInfo: YanwenUnitedStatesAddressReceiverInfo{
				Address: receiver.Address,
				ZipCode: receiver.ZipCode,
				City:    receiver.City,
				State:   receiver.State,
			},
		}); err != nil {
			return fmt.Errorf("verify Yanwen US address before waybill creation: %w", err)
		}
	}

	return nil
}

func (s *YanwenWaybillOperationsService) resolveYanwenOrderCountry(environment, value string) (string, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "", errors.New("order shipping country is required")
	}
	countries, err := s.countries.FindYanwenCountryCatalogEntriesByEnvironment(environment)
	if err != nil {
		return "", fmt.Errorf("load Yanwen countries: %w", err)
	}
	for _, country := range countries {
		if strings.EqualFold(country.CountryCode, value) || strings.EqualFold(country.CountryID, value) {
			return strings.ToUpper(country.CountryCode), nil
		}
	}
	return "", fmt.Errorf("order shipping country %q is not present in the official %s catalog", value, environment)
}
