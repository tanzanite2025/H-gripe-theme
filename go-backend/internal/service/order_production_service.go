package service

import (
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/repository"
	"time"
)

func orderRequiresProduction(record *order.Order) bool {
	if record == nil {
		return false
	}
	mode := order.NormalizeFulfillmentMode(record.FulfillmentMode)
	return mode == order.FulfillmentModeMadeToOrder || mode == order.FulfillmentModeMixed
}

func (s *OrderService) StartProduction(id uint) (*order.Order, error) {
	if s == nil || s.txManager == nil {
		return nil, ErrOrderProductionTransactionNeeded
	}

	startedAt := time.Now().UTC()
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Order == nil {
			return ErrOrderProductionTransactionNeeded
		}

		record, err := repos.Order.FindByIDForUpdateWithItems(id)
		if err != nil {
			return normalizeOrderError(err)
		}
		if !orderRequiresProduction(record) {
			return ErrOrderProductionNotRequired
		}
		if record.PaymentStatus != "paid" {
			return ErrOrderProductionPaymentRequired
		}
		if record.Status != "paid" && record.Status != "processing" {
			return ErrOrderProductionNotAllowed
		}

		switch order.NormalizeProductionStatus(record.ProductionStatus) {
		case order.ProductionStatusStarted, order.ProductionStatusCompleted:
			return ErrOrderProductionAlreadyStarted
		case order.ProductionStatusNotStarted:
		default:
			return ErrOrderProductionNotAllowed
		}

		updated, err := repos.Order.MarkProductionStarted(id, startedAt)
		if err != nil {
			return err
		}
		if !updated {
			return ErrOrderProductionAlreadyStarted
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetAdminOrder(id)
}

func (s *OrderService) CompleteProduction(id uint) (*order.Order, error) {
	if s == nil || s.txManager == nil {
		return nil, ErrOrderProductionTransactionNeeded
	}

	completedAt := time.Now().UTC()
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Order == nil {
			return ErrOrderProductionTransactionNeeded
		}

		record, err := repos.Order.FindByIDForUpdateWithItems(id)
		if err != nil {
			return normalizeOrderError(err)
		}
		if !orderRequiresProduction(record) {
			return ErrOrderProductionNotRequired
		}
		if record.PaymentStatus != "paid" {
			return ErrOrderProductionPaymentRequired
		}
		if record.Status != "paid" && record.Status != "processing" {
			return ErrOrderProductionNotAllowed
		}

		switch order.NormalizeProductionStatus(record.ProductionStatus) {
		case order.ProductionStatusNotStarted:
			return ErrOrderProductionNotStarted
		case order.ProductionStatusCompleted:
			return ErrOrderProductionAlreadyStarted
		case order.ProductionStatusStarted:
		default:
			return ErrOrderProductionNotAllowed
		}

		updated, err := repos.Order.MarkProductionCompleted(id, completedAt)
		if err != nil {
			return err
		}
		if !updated {
			return ErrOrderProductionNotAllowed
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetAdminOrder(id)
}
