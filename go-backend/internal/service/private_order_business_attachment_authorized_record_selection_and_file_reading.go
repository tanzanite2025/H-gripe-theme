package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/pkg/storage"
)

var ErrPrivateOrderBusinessAttachmentNotFound = errors.New("private order attachment not found")
var ErrPrivateOrderBusinessAttachmentUnavailable = errors.New("private order attachment storage unavailable")

// The selector is interpreted only after authorizing the owning business row.
// URLs and object keys supplied by the requester never participate in selection.
func (s *WarrantyService) OpenWarrantyClaimAttachmentForViewer(ctx context.Context, claimID, userID uint, email, token string, administrator bool, kind string, index int) (*storage.StoredObject, error) {
	claim, err := s.GetWarrantyClaimForViewer(claimID, userID, email, token, administrator)
	if err != nil {
		return nil, err
	}
	var reference string
	switch kind {
	case "images":
		var images []string
		if json.Unmarshal([]byte(claim.Images), &images) != nil || index < 0 || index >= len(images) {
			return nil, ErrPrivateOrderBusinessAttachmentNotFound
		}
		reference = images[index]
	case "video":
		if index != 0 {
			return nil, ErrPrivateOrderBusinessAttachmentNotFound
		}
		reference = claim.VideoURL
	default:
		return nil, ErrPrivateOrderBusinessAttachmentNotFound
	}
	if s.privateAttachmentUploads == nil {
		return nil, ErrPrivateOrderBusinessAttachmentUnavailable
	}
	return openPrivateOrderBusinessAttachmentFromOwningProvider(ctx, s.privateAttachmentUploads.storage, reference, "warranty", func(key string) bool {
		// Historical claims used warranty/<filename>; shipment evidence has a
		// different owner even when its basename matches a claim attachment.
		return !strings.HasPrefix(key, "warranty/shipment/")
	})
}

func (s *ShipmentRecordService) OpenShipmentImageForAdministrator(ctx context.Context, orderID uint, index int) (*storage.StoredObject, error) {
	record, err := s.GetAdmin(orderID)
	if err != nil {
		return nil, err
	}
	return s.openShipmentImageFromAuthorizedRecord(ctx, record.ShippingImages, record.OrderID, index)
}

func (s *ShipmentRecordService) OpenShipmentImageForCustomer(ctx context.Context, orderNumber string, userID uint, index int) (*storage.StoredObject, error) {
	if userID == 0 {
		return nil, ErrPrivateOrderBusinessAttachmentNotFound
	}
	record, err := s.GetForUser(orderNumber, userID)
	if err != nil {
		return nil, err
	}
	return s.openShipmentImageFromAuthorizedRecord(ctx, record.ShippingImages, record.OrderID, index)
}

func (s *ShipmentRecordService) openShipmentImageFromAuthorizedRecord(ctx context.Context, encodedImages []byte, orderID uint, index int) (*storage.StoredObject, error) {
	var images []string
	if json.Unmarshal(encodedImages, &images) != nil || index < 0 || index >= len(images) {
		return nil, ErrPrivateOrderBusinessAttachmentNotFound
	}
	if s.privateAttachmentUploads == nil {
		return nil, ErrPrivateOrderBusinessAttachmentUnavailable
	}
	return openPrivateOrderBusinessAttachmentFromOwningProvider(ctx, s.privateAttachmentUploads.storage, images[index], "warranty/shipment", func(key string) bool {
		return !strings.Contains(key, "/owned-upload-") || strings.HasPrefix(key, fmt.Sprintf("warranty/shipment/%d/owned-upload-", orderID))
	})
}

// This locator hides internal case/attachment IDs. It grants no access: every
// download independently checks the current order owner and attachment parent.
func CustomerAfterSalesAttachmentOpaqueLocator(attachment aftersales.AfterSalesCaseAttachment) string {
	digest := sha256.Sum256([]byte(fmt.Sprintf("after-sales-attachment:%d:%d:%s", attachment.CaseID, attachment.ID, attachment.StorageURL)))
	return hex.EncodeToString(digest[:])
}

func (s *AfterSalesService) OpenAfterSalesAttachmentForCustomer(ctx context.Context, orderNumber string, userID uint, locator string) (*AfterSalesAttachmentFile, error) {
	if s == nil || s.caseRepo == nil || userID == 0 || len(locator) != sha256.Size*2 {
		return nil, ErrAfterSalesAttachmentNotFound
	}
	attachments, err := s.caseRepo.FindAfterSalesAttachmentsForCurrentOrderOwner(ctx, orderNumber, userID)
	if err != nil {
		return nil, err
	}
	for _, attachment := range attachments {
		if CustomerAfterSalesAttachmentOpaqueLocator(attachment) == locator {
			return s.OpenAttachmentFile(ctx, attachment.CaseID, attachment.ID)
		}
	}
	return nil, ErrAfterSalesAttachmentNotFound
}

func openPrivateOrderBusinessAttachmentFromOwningProvider(ctx context.Context, provider storage.StorageService, reference, namespace string, allowedKey func(string) bool) (*storage.StoredObject, error) {
	key, err := storage.ResolvePrivateBusinessAttachmentObjectKeyForAuthorizedRead(provider, reference)
	if err != nil || !strings.HasPrefix(key, namespace+"/") || (allowedKey != nil && !allowedKey(key)) {
		return nil, ErrPrivateOrderBusinessAttachmentNotFound
	}
	opener, ok := provider.(storage.ObjectOpener)
	if !ok {
		return nil, ErrPrivateOrderBusinessAttachmentUnavailable
	}
	object, err := opener.Open(ctx, key)
	if err != nil {
		// Do not expose bucket names, filesystem paths or remote error bodies.
		return nil, ErrPrivateOrderBusinessAttachmentNotFound
	}
	if object == nil || object.ReadCloser == nil {
		return nil, ErrPrivateOrderBusinessAttachmentUnavailable
	}
	if object.Size < 0 {
		_ = object.ReadCloser.Close()
		return nil, ErrPrivateOrderBusinessAttachmentUnavailable
	}
	return object, nil
}

// Keep record-not-found recognizable to the existing API error adapters.
func IsPrivateOrderBusinessAttachmentNotFound(err error) bool {
	return errors.Is(err, ErrPrivateOrderBusinessAttachmentNotFound) || errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, ErrAfterSalesAttachmentNotFound)
}
