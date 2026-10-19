package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/aftersales"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/domain/warranty"
)

func (s *WarrantyService) ConfigurePrivateOrderBusinessAttachmentUploads(uploads *PrivateOrderBusinessAttachmentUploadService) {
	s.privateAttachmentUploads = uploads
}
func (s *AfterSalesService) ConfigurePrivateOrderBusinessAttachmentUploads(uploads *PrivateOrderBusinessAttachmentUploadService) {
	s.privateAttachmentUploads = uploads
}
func (s *ShipmentRecordService) ConfigurePrivateOrderBusinessAttachmentUploads(uploads *PrivateOrderBusinessAttachmentUploadService) {
	s.privateAttachmentUploads = uploads
}

func (s *WarrantyService) CreateWarrantyClaimForOrder(input WarrantyClaimByOrderInput) (*warranty.WarrantyClaim, error) {
	if s == nil || s.txManager == nil {
		return nil, ErrEmailChallengeUnavailable
	}
	if !s.acceptPreallocatedPrivateAttachmentBinding && privateOrderBusinessReferencesContainPreallocatedUpload(append(append([]string{}, input.ImageURLs...), input.VideoURL)) {
		return nil, fmt.Errorf("new private warranty attachments require the verified upload workflow")
	}
	var result *warranty.WarrantyClaim
	err := s.txManager.WithinPrivateOrderBusinessAttachmentMutationLock(func(tx *gorm.DB) error {
		clonedValue := *s
		clonedValue.txManager = s.txManager.WithPrivateOrderBusinessAttachmentTransaction(tx)
		clonedValue.warrantyRepo = s.warrantyRepo.WithTx(tx)
		clonedValue.orderRepo = s.orderRepo.WithTx(tx)
		if s.shipmentRepo != nil {
			clonedValue.shipmentRepo = s.shipmentRepo.WithTx(tx)
		}
		var err error
		result, err = clonedValue.createWarrantyClaimAfterPrivateAttachmentNamespaceLock(input)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *WarrantyService) CreateVerifiedWarrantyClaimWithPrivateAttachmentFiles(ctx context.Context, input WarrantyClaimByOrderInput, files PrivateOrderBusinessAttachmentUploadFiles) (*warranty.WarrantyClaim, error) {
	if s == nil {
		return nil, ErrEmailChallengeUnavailable
	}
	input.ImageURLs, input.VideoURL = nil, ""
	if len(files.Images) == 0 && files.Video == nil {
		return s.CreateWarrantyClaimForOrder(input)
	}
	var result *warranty.WarrantyClaim
	err := s.privateAttachmentUploads.uploadPrivateOrderBusinessAttachmentBatchAndBindOwner(ctx, "warranty/claims", files, func(tx *gorm.DB, prepared []preparedPrivateOrderBusinessAttachment) error {
		clonedValue := *s
		clonedValue.acceptPreallocatedPrivateAttachmentBinding = true
		clonedValue.txManager = s.txManager.WithPrivateOrderBusinessAttachmentTransaction(tx)
		clonedValue.warrantyRepo = s.warrantyRepo.WithTx(tx)
		clonedValue.orderRepo = s.orderRepo.WithTx(tx)
		if s.shipmentRepo != nil {
			clonedValue.shipmentRepo = s.shipmentRepo.WithTx(tx)
		}
		input.ImageURLs, input.VideoURL = privateOrderBusinessAttachmentURLs(prepared)
		var err error
		result, err = clonedValue.createWarrantyClaimAfterPrivateAttachmentNamespaceLock(input)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AfterSalesService) CreateCustomerAfterSalesRequestWithPrivateAttachmentFiles(ctx context.Context, input CreateCustomerAfterSalesRequestInput, files PrivateOrderBusinessAttachmentUploadFiles) (*aftersales.AfterSalesCase, error) {
	if s == nil {
		return nil, ErrAfterSalesAttachmentStorageUnavailable
	}
	input.Attachments = nil
	if len(files.Images) == 0 && files.Video == nil {
		return s.CreateCustomerRequest(input)
	}
	var result *aftersales.AfterSalesCase
	err := s.privateAttachmentUploads.uploadPrivateOrderBusinessAttachmentBatchAndBindOwner(ctx, "after-sales/pending", files, func(tx *gorm.DB, prepared []preparedPrivateOrderBusinessAttachment) error {
		clonedValue := s.afterSalesServiceWithPrivateAttachmentTransaction(tx)
		clonedValue.acceptPreallocatedPrivateAttachmentBinding = true
		owner, err := clonedValue.orderRepo.FindByIDForUpdateWithItems(input.OrderID)
		if err != nil {
			return err
		}
		if input.CreatedBy == 0 || owner.UserID != input.CreatedBy {
			return ErrAfterSalesOrderNotFound
		}
		input.Attachments = privateOrderBusinessAfterSalesAttachments(prepared)
		result, err = clonedValue.createCustomerRequestAfterPrivateAttachmentNamespaceLock(input)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AfterSalesService) afterSalesServiceWithPrivateAttachmentTransaction(tx *gorm.DB) *AfterSalesService {
	clonedValue := *s
	clonedValue.txManager = s.txManager.WithPrivateOrderBusinessAttachmentTransaction(tx)
	clonedValue.caseRepo = s.caseRepo.WithTx(tx)
	clonedValue.orderRepo = s.orderRepo.WithTx(tx)
	if s.userRepo != nil {
		clonedValue.userRepo = s.userRepo.WithTx(tx)
	}
	if s.refundReviewRepo != nil {
		clonedValue.refundReviewRepo = s.refundReviewRepo.WithTx(tx)
	}
	return &clonedValue
}

func (s *AfterSalesService) CreateCustomerRequest(input CreateCustomerAfterSalesRequestInput) (*aftersales.AfterSalesCase, error) {
	if s == nil || s.txManager == nil || s.caseRepo == nil || s.orderRepo == nil {
		return nil, ErrAfterSalesAttachmentStorageUnavailable
	}
	if !s.acceptPreallocatedPrivateAttachmentBinding {
		for _, attachment := range input.Attachments {
			if privateOrderBusinessReferencesContainPreallocatedUpload([]string{attachment.StorageURL}) {
				return nil, fmt.Errorf("new private after-sales attachments require the authenticated upload workflow")
			}
		}
	}
	var result *aftersales.AfterSalesCase
	err := s.txManager.WithinPrivateOrderBusinessAttachmentMutationLock(func(tx *gorm.DB) error {
		var err error
		result, err = s.afterSalesServiceWithPrivateAttachmentTransaction(tx).createCustomerRequestAfterPrivateAttachmentNamespaceLock(input)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ShipmentRecordService) AddPrivateShipmentImageFilesWithAtomicOwnerBinding(ctx context.Context, orderID uint, images []*multipart.FileHeader) (*shippingdomain.ShipmentRecord, []string, error) {
	if s == nil || s.repo == nil {
		return nil, nil, ErrShipmentRecordUnavailable
	}
	var result *shippingdomain.ShipmentRecord
	var urls []string
	err := s.privateAttachmentUploads.uploadPrivateOrderBusinessAttachmentBatchAndBindOwner(ctx, fmt.Sprintf("warranty/shipment/%d", orderID), PrivateOrderBusinessAttachmentUploadFiles{Images: images}, func(tx *gorm.DB, prepared []preparedPrivateOrderBusinessAttachment) error {
		clonedValue := *s
		clonedValue.repo = s.repo.WithTx(tx)
		clonedValue.acceptPreallocatedPrivateAttachmentBinding = true
		urls, _ = privateOrderBusinessAttachmentURLs(prepared)
		var err error
		result, err = clonedValue.addShipmentImagesAfterPrivateAttachmentNamespaceLock(orderID, urls)
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return result, urls, nil
}

func validateNewPreallocatedPrivateShipmentImageReferences(current, requested []string, allowed bool) error {
	if allowed {
		return nil
	}
	retained := make(map[string]bool, len(current))
	for _, reference := range current {
		retained[reference] = true
	}
	for _, reference := range requested {
		if privateOrderBusinessReferencesContainPreallocatedUpload([]string{reference}) && !retained[reference] {
			return fmt.Errorf("new private shipment images must be uploaded to this order")
		}
	}
	return nil
}

func privateOrderBusinessReferencesContainPreallocatedUpload(references []string) bool {
	for _, reference := range references {
		if decoded, err := url.PathUnescape(reference); err == nil {
			reference = decoded
		}
		reference = strings.ReplaceAll(reference, "\\", "/")
		if strings.Contains(reference, "/owned-upload-") {
			return true
		}
	}
	return false
}
