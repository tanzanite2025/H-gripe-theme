package service

import (
	"errors"

	"commerce-platform/internal/repository"
)

func (s *ProductService) Delete(id uint) error {
	return s.deleteProductByID(id, true)
}

func (s *ProductService) deleteProductByID(id uint, shouldInvalidateHTML bool) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if s.txManager != nil {
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.Product == nil {
				return errors.New("transactional product repository is not configured")
			}
			if err := tx.Product.Delete(id); err != nil {
				return err
			}
			cacheEvents, merchantEvents, err := s.transactionalProductPublishers(tx.Outbox)
			if err != nil {
				return err
			}
			if err := enqueueProductCacheInvalidationByIDsWithPublisher(cacheEvents, []uint{existingProduct.ID}, "admin product delete"); err != nil {
				return err
			}
			if err := enqueueMerchantProductWithdrawWithPublisher(merchantEvents, existingProduct, "product_deleted"); err != nil {
				return requireMerchantEvent(err, existingProduct, "product_deleted")
			}
			return nil
		})
		if err != nil {
			return err
		}
		s.clearProductCache(existingProduct)
		if shouldInvalidateHTML {
			s.invalidateStorefrontHTMLCache("admin product delete")
		}
		return nil
	}

	if err := s.productRepo.Delete(id); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)
	if err := s.enqueueProductCacheInvalidationByIDs([]uint{existingProduct.ID}, "admin product delete"); err != nil {
		return err
	}
	if shouldInvalidateHTML {
		s.invalidateStorefrontHTMLCache("admin product delete")
	}

	if err := s.enqueueMerchantProductWithdraw(existingProduct, "product_deleted"); err != nil {
		return requireMerchantEvent(err, existingProduct, "product_deleted")
	}
	return nil
}

func (s *ProductService) UpdateStatus(id uint, status string) error {
	return s.updateProductStatusByID(id, status, true)
}

func (s *ProductService) updateProductStatusByID(id uint, status string, shouldInvalidateHTML bool) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if s.txManager != nil {
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.Product == nil {
				return errors.New("transactional product repository is not configured")
			}
			if err := tx.Product.UpdateStatus(id, status); err != nil {
				return err
			}
			updatedProduct, err := tx.Product.FindByID(id)
			if err != nil {
				return err
			}
			cacheEvents, merchantEvents, err := s.transactionalProductPublishers(tx.Outbox)
			if err != nil {
				return err
			}
			if err := enqueueProductCacheInvalidationByIDsWithPublisher(cacheEvents, []uint{id}, "admin product status update"); err != nil {
				return err
			}
			if err := enqueueMerchantProductChangeWithPublisher(merchantEvents, updatedProduct, "product_status_changed"); err != nil {
				return requireMerchantEvent(err, updatedProduct, "product_status_changed")
			}
			return nil
		})
		if err != nil {
			return err
		}
		s.clearProductCache(existingProduct)
		if shouldInvalidateHTML {
			s.invalidateStorefrontHTMLCache("admin product status update")
		}
		return nil
	}

	if err := s.productRepo.UpdateStatus(id, status); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)
	if err := s.enqueueProductCacheInvalidationByIDs([]uint{existingProduct.ID}, "admin product status update"); err != nil {
		return err
	}
	if shouldInvalidateHTML {
		s.invalidateStorefrontHTMLCache("admin product status update")
	}

	existingProduct.Status = status
	if err := s.enqueueMerchantProductChange(existingProduct, "product_status_changed"); err != nil {
		return requireMerchantEvent(err, existingProduct, "product_status_changed")
	}
	return nil
}

func (s *ProductService) BatchUpdateStatus(ids []uint, status string) (int, error) {
	updated := 0
	for _, id := range ids {
		if err := s.updateProductStatusByID(id, status, false); err != nil {
			if errors.Is(err, ErrProductNotFound) {
				continue
			}
			return updated, err
		}
		updated++
	}
	if updated > 0 {
		s.invalidateStorefrontHTMLCache("admin product batch status update")
	}

	return updated, nil
}

func (s *ProductService) BatchDelete(ids []uint) (int, error) {
	deleted := 0
	for _, id := range ids {
		if err := s.deleteProductByID(id, false); err != nil {
			if errors.Is(err, ErrProductNotFound) {
				continue
			}
			return deleted, err
		}
		deleted++
	}
	if deleted > 0 {
		s.invalidateStorefrontHTMLCache("admin product batch delete")
	}

	return deleted, nil
}
