package service

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/warranty"
	"commerce-platform/internal/pkg/storage"
)

func assertAuthorizedPrivateAttachmentHasReadableBytesAndClose(t *testing.T, object *storage.StoredObject, err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, object)
	defer object.ReadCloser.Close()
	data, err := io.ReadAll(object.ReadCloser)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.Equal(t, int64(len(data)), object.Size)
}

func TestPrivateOrderAttachmentWarrantyOwnerGuestTokenAndHistoricalReferenceReadAuthorization(t *testing.T) {
	env := createPrivateOrderBusinessAttachmentReceiptTestEnvironment(t)
	claim, err := env.warranty.CreateVerifiedWarrantyClaimWithPrivateAttachmentFiles(t.Context(), env.claimInput, PrivateOrderBusinessAttachmentUploadFiles{Images: []*multipart.FileHeader{newEvidencePNGFileHeader(t, "proof.png")}})
	require.NoError(t, err)
	object, err := env.warranty.OpenWarrantyClaimAttachmentForViewer(t.Context(), claim.ID, env.order.UserID, "", "", false, "images", 0)
	assertAuthorizedPrivateAttachmentHasReadableBytesAndClose(t, object, err)
	for _, selector := range []struct {
		kind  string
		index int
	}{{"images", -1}, {"images", 1}, {"video", 0}, {"object-key", 0}} {
		_, err := env.warranty.OpenWarrantyClaimAttachmentForViewer(t.Context(), claim.ID, env.order.UserID, "", "", false, selector.kind, selector.index)
		require.ErrorIs(t, err, ErrPrivateOrderBusinessAttachmentNotFound)
	}
	_, err = env.warranty.OpenWarrantyClaimAttachmentForViewer(t.Context(), claim.ID, env.order.UserID+1, claim.Email, "", false, "images", 0)
	require.EqualError(t, err, "unauthorized", "email must never bypass an account-owned claim")
	object, err = env.warranty.OpenWarrantyClaimAttachmentForViewer(t.Context(), claim.ID, 0, "", "", true, "images", 0)
	assertAuthorizedPrivateAttachmentHasReadableBytesAndClose(t, object, err)

	require.NoError(t, env.db.Model(claim).Update("user_id", 0).Error)
	claim.UserID = 0
	token, err := env.warranty.IssueWarrantyClaimAccessToken(claim)
	require.NoError(t, err)
	object, err = env.warranty.OpenWarrantyClaimAttachmentForViewer(t.Context(), claim.ID, 0, "", token, false, "images", 0)
	assertAuthorizedPrivateAttachmentHasReadableBytesAndClose(t, object, err)
	_, err = env.warranty.OpenWarrantyClaimAttachmentForViewer(t.Context(), claim.ID, 0, "", token+"corrupt", false, "images", 0)
	require.EqualError(t, err, "unauthorized")
	other := *claim
	other.ID = 0
	require.NoError(t, env.db.Omit("OrderItem", "ServiceRecords").Create(&other).Error)
	_, err = env.warranty.OpenWarrantyClaimAttachmentForViewer(t.Context(), other.ID, 0, "", token, false, "images", 0)
	require.EqualError(t, err, "unauthorized", "token must be bound to this exact claim")

	legacyKey := "warranty/historical-proof.png"
	_, err = env.provider.StorageService.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(t.Context(), strings.NewReader("historical private bytes"), legacyKey)
	require.NoError(t, err)
	encoded, _ := json.Marshal([]string{env.provider.GetURL(legacyKey)})
	require.NoError(t, env.db.Model(claim).Update("images", string(encoded)).Error)
	object, err = env.warranty.OpenWarrantyClaimAttachmentForViewer(t.Context(), claim.ID, 0, "", token, false, "images", 0)
	assertAuthorizedPrivateAttachmentHasReadableBytesAndClose(t, object, err)
	require.NoError(t, env.db.Delete(&warranty.WarrantyClaim{}, claim.ID).Error)
	_, err = env.warranty.OpenWarrantyClaimAttachmentForViewer(t.Context(), claim.ID, 0, "", token, false, "images", 0)
	require.True(t, IsPrivateOrderBusinessAttachmentNotFound(err), "retained deleted rows must not become public reads")
}

func TestPrivateOrderAttachmentShipmentCurrentOrderOwnershipExactIndexAndForeignIdentityRejection(t *testing.T) {
	env := createPrivateOrderBusinessAttachmentReceiptTestEnvironment(t)
	record, _, err := env.shipment.AddPrivateShipmentImageFilesWithAtomicOwnerBinding(t.Context(), env.order.ID, []*multipart.FileHeader{newEvidencePNGFileHeader(t, "proof.png")})
	require.NoError(t, err)
	object, err := env.shipment.OpenShipmentImageForCustomer(t.Context(), env.order.OrderNumber, env.order.UserID, 0)
	assertAuthorizedPrivateAttachmentHasReadableBytesAndClose(t, object, err)
	for _, userID := range []uint{0, env.order.UserID + 1} {
		_, err := env.shipment.OpenShipmentImageForCustomer(t.Context(), env.order.OrderNumber, userID, 0)
		require.True(t, IsPrivateOrderBusinessAttachmentNotFound(err))
	}
	_, err = env.shipment.OpenShipmentImageForCustomer(t.Context(), env.order.OrderNumber+"-wrong", env.order.UserID, 0)
	require.True(t, IsPrivateOrderBusinessAttachmentNotFound(err))
	_, err = env.shipment.OpenShipmentImageForAdministrator(t.Context(), env.order.ID, 1)
	require.ErrorIs(t, err, ErrPrivateOrderBusinessAttachmentNotFound)
	var refs []string
	require.NoError(t, json.Unmarshal(record.ShippingImages, &refs))
	for _, reference := range []string{"https://foreign.example.test/uploads/warranty/shipment/proof.png", env.provider.GetURL("warranty/claims/proof.png"), env.provider.GetURL("warranty/shipment/999/owned-upload-wrong/proof.png")} {
		encoded, _ := json.Marshal([]string{reference})
		require.NoError(t, env.db.Model(record).Update("shipping_images", string(encoded)).Error)
		_, err := env.shipment.OpenShipmentImageForAdministrator(t.Context(), env.order.ID, 0)
		require.ErrorIs(t, err, ErrPrivateOrderBusinessAttachmentNotFound)
	}
	encoded, _ := json.Marshal(refs)
	require.NoError(t, env.db.Model(record).Update("shipping_images", string(encoded)).Error)
	require.NoError(t, env.db.Model(env.order).Update("user_id", env.order.UserID+1).Error)
	_, err = env.shipment.OpenShipmentImageForCustomer(t.Context(), env.order.OrderNumber, 1, 0)
	require.True(t, IsPrivateOrderBusinessAttachmentNotFound(err), "attachment snapshot user must not authorize a previous order owner")
}

func TestPrivateOrderAttachmentAfterSalesOpaqueLocatorOrderParentAndCurrentOwnership(t *testing.T) {
	env := createPrivateOrderBusinessAttachmentReceiptTestEnvironment(t)
	env.afterSales.ConfigureAttachmentStorage(env.provider)
	record, err := env.afterSales.CreateCustomerAfterSalesRequestWithPrivateAttachmentFiles(t.Context(), env.afterSalesInput, PrivateOrderBusinessAttachmentUploadFiles{Images: []*multipart.FileHeader{newEvidencePNGFileHeader(t, "proof.png")}})
	require.NoError(t, err)
	require.Len(t, record.Attachments, 1)
	attachment := record.Attachments[0]
	locator := CustomerAfterSalesAttachmentOpaqueLocator(attachment)
	require.Len(t, locator, 64)
	file, err := env.afterSales.OpenAfterSalesAttachmentForCustomer(t.Context(), env.order.OrderNumber, env.order.UserID, locator)
	require.NoError(t, err)
	data, err := io.ReadAll(file.ReadCloser)
	require.NoError(t, err)
	require.NoError(t, file.ReadCloser.Close())
	require.NotEmpty(t, data)
	for _, request := range []struct {
		number  string
		userID  uint
		locator string
	}{
		{env.order.OrderNumber, 0, locator}, {env.order.OrderNumber, env.order.UserID + 1, locator},
		{env.order.OrderNumber + "-wrong", env.order.UserID, locator}, {env.order.OrderNumber, env.order.UserID, strings.Repeat("0", 64)},
	} {
		_, err := env.afterSales.OpenAfterSalesAttachmentForCustomer(t.Context(), request.number, request.userID, request.locator)
		require.ErrorIs(t, err, ErrAfterSalesAttachmentNotFound)
	}
	_, err = env.afterSales.OpenAttachmentFile(t.Context(), record.ID+1, attachment.ID)
	require.ErrorIs(t, err, ErrAfterSalesAttachmentNotFound)
	for _, reference := range []string{"https://foreign.example.test/uploads/after-sales/pending/proof.png", env.provider.GetURL("warranty/claims/proof.png")} {
		require.NoError(t, env.db.Model(&aftersales.AfterSalesCaseAttachment{}).Where("id = ?", attachment.ID).Update("storage_url", reference).Error)
		_, err := env.afterSales.OpenAttachmentFile(t.Context(), record.ID, attachment.ID)
		require.ErrorIs(t, err, ErrPrivateOrderBusinessAttachmentNotFound)
	}
	require.NoError(t, env.db.Model(&aftersales.AfterSalesCaseAttachment{}).Where("id = ?", attachment.ID).Update("storage_url", attachment.StorageURL).Error)
	require.NoError(t, env.db.Model(env.order).Update("user_id", env.order.UserID+1).Error)
	_, err = env.afterSales.OpenAfterSalesAttachmentForCustomer(t.Context(), env.order.OrderNumber, 1, locator)
	require.ErrorIs(t, err, ErrAfterSalesAttachmentNotFound)
	require.NoError(t, env.db.Delete(&order.Order{}, env.order.ID).Error)
	_, err = env.afterSales.OpenAfterSalesAttachmentForCustomer(t.Context(), env.order.OrderNumber, env.order.UserID, locator)
	require.ErrorIs(t, err, ErrAfterSalesAttachmentNotFound)
}

func TestPrivateOrderAttachmentAuthorizedVideoReadPreservesMP4MOVAndWebMTypes(t *testing.T) {
	for _, owner := range []string{"warranty", "after-sales"} {
		for _, video := range []struct {
			name, mime string
			content    []byte
		}{
			{"proof.mov", "video/quicktime", []byte{0, 0, 0, 20, 'f', 't', 'y', 'p', 'q', 't', ' ', ' ', 0, 0, 0, 0}},
			{"proof.mp4", "video/mp4", []byte{0, 0, 0, 20, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 0, 0, 0, 0}},
			{"proof.webm", "video/webm", []byte{0x1a, 0x45, 0xdf, 0xa3, 0, 0, 'w', 'e', 'b', 'm'}},
		} {
			t.Run(owner+"/"+video.name, func(t *testing.T) {
				env := createPrivateOrderBusinessAttachmentReceiptTestEnvironment(t)
				files := PrivateOrderBusinessAttachmentUploadFiles{Video: multipartFileHeader(t, video.name, video.mime, video.content)}
				var reader io.ReadCloser
				var contentType string
				if owner == "warranty" {
					claim, err := env.warranty.CreateVerifiedWarrantyClaimWithPrivateAttachmentFiles(t.Context(), env.claimInput, files)
					require.NoError(t, err)
					object, err := env.warranty.OpenWarrantyClaimAttachmentForViewer(t.Context(), claim.ID, env.order.UserID, "", "", false, "video", 0)
					require.NoError(t, err)
					reader, contentType = object.ReadCloser, object.MimeType
				} else {
					env.afterSales.ConfigureAttachmentStorage(env.provider)
					record, err := env.afterSales.CreateCustomerAfterSalesRequestWithPrivateAttachmentFiles(t.Context(), env.afterSalesInput, files)
					require.NoError(t, err)
					file, err := env.afterSales.OpenAfterSalesAttachmentForCustomer(t.Context(), env.order.OrderNumber, env.order.UserID, CustomerAfterSalesAttachmentOpaqueLocator(record.Attachments[0]))
					require.NoError(t, err)
					reader, contentType = file.ReadCloser, file.MimeType
				}
				defer reader.Close()
				content, err := io.ReadAll(reader)
				require.NoError(t, err)
				require.Equal(t, video.content, content)
				require.Equal(t, video.mime, contentType)
			})
		}
	}
}
