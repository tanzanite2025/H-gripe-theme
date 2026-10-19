package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
	"testing"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/domain/verification"
	"commerce-platform/internal/domain/warranty"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type privateOrderBusinessAttachmentReceiptTestProvider struct {
	storage.StorageService
	mode        string
	keys        []string
	deleteCalls int
	beforeWrite func(context.Context, string)
	afterWrite  func(context.Context, string)
}

func (p *privateOrderBusinessAttachmentReceiptTestProvider) WritePrivateObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	p.keys = append(p.keys, key)
	if p.beforeWrite != nil {
		p.beforeWrite(ctx, key)
	}
	if p.mode == "corrupt bytes" {
		reader = bytes.NewReader([]byte("corrupt private evidence"))
	}
	reference, err := p.StorageService.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(ctx, reader, key)
	if err != nil {
		return "", err
	}
	if p.afterWrite != nil {
		p.afterWrite(ctx, key)
	}
	switch p.mode {
	case "lost response":
		return "", errors.New("provider accepted write but response was lost")
	case "second lost response":
		if len(p.keys) == 2 {
			return "", errors.New("second write accepted without response")
		}
	case "empty response":
		return "", nil
	case "foreign response":
		return "https://untrusted.example.test/uploads/" + key, nil
	}
	return reference, nil
}

func (p *privateOrderBusinessAttachmentReceiptTestProvider) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return p.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func (p *privateOrderBusinessAttachmentReceiptTestProvider) Delete(ctx context.Context, key string) error {
	p.deleteCalls++
	return p.StorageService.Delete(ctx, key)
}

type privateOrderBusinessAttachmentReceiptTestEnvironment struct {
	db              *gorm.DB
	provider        *privateOrderBusinessAttachmentReceiptTestProvider
	uploads         *PrivateOrderBusinessAttachmentUploadService
	warranty        *WarrantyService
	afterSales      *AfterSalesService
	shipment        *ShipmentRecordService
	order           *order.Order
	claimInput      WarrantyClaimByOrderInput
	afterSalesInput CreateCustomerAfterSalesRequestInput
}

func configurePrivateOrderBusinessAttachmentReceiptTestEnvironment(t *testing.T, db *gorm.DB, afterSales *AfterSalesService) privateOrderBusinessAttachmentReceiptTestEnvironment {
	t.Helper()
	fixtureDDL := db.Session(&gorm.Session{})
	fixtureDDL.Config.DisableForeignKeyConstraintWhenMigrating = true
	fixtureDDL.Config.IgnoreRelationshipsWhenMigrating = true
	for _, model := range []any{&warranty.WarrantyClaim{}, &warranty.WarrantyServiceRecord{}, &verification.EmailChallenge{}, &shippingdomain.ShipmentRecord{}, &shippingdomain.TrackingProviderConfig{}, &shippingdomain.TrackingShipment{}, &outbox.Event{}} {
		if !db.Migrator().HasTable(model) {
			require.NoError(t, fixtureDDL.Migrator().CreateTable(model), "create private attachment fixture model %T", model)
		}
	}
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://private-business.example.test/uploads"})
	require.NoError(t, err)
	controlled := &privateOrderBusinessAttachmentReceiptTestProvider{StorageService: provider}
	uploads := NewPrivateOrderBusinessAttachmentUploadService(repository.NewOutboxRepository(db), controlled, "https://shop.example.test")
	warrantyService := NewWarrantyService(newTestEmailChallengeTxManager(db), repository.NewWarrantyRepository(db), repository.NewOrderRepository(db), repository.NewShipmentRecordRepository(db))
	warrantyService.ConfigureEmailChallenges("test-email-secret")
	warrantyService.ConfigureEmailBaseURL("https://shop.example.test")
	warrantyService.ConfigurePrivateOrderBusinessAttachmentUploads(uploads)
	afterSales.ConfigurePrivateOrderBusinessAttachmentUploads(uploads)
	shipmentService := NewShipmentRecordService(repository.NewShipmentRecordRepository(db))
	shipmentService.ConfigurePrivateOrderBusinessAttachmentUploads(uploads)
	orderRecord := seedAfterSalesOrder(t, db, 2)
	orderRecord.OrderNumber = "PRIVATE-ATTACHMENT-" + uuid.NewString()
	orderRecord.ShippingAddress.Email = "verified-private@example.test"
	require.NoError(t, db.Save(orderRecord).Error)
	require.NoError(t, warrantyService.RequestWarrantyOrderVerification(orderRecord.OrderNumber, orderRecord.ShippingAddress.Email))
	sender := &recordingEmailSender{}
	processEmailChallengeDelivery(t, db, sender)
	require.Len(t, sender.bodies, 1)
	verificationURL, err := url.Parse(strings.TrimSpace(strings.Split(sender.bodies[0], "\n\n")[1]))
	require.NoError(t, err)
	return privateOrderBusinessAttachmentReceiptTestEnvironment{db: db, provider: controlled, uploads: uploads, warranty: warrantyService, afterSales: afterSales, shipment: shipmentService, order: orderRecord,
		claimInput:      WarrantyClaimByOrderInput{OrderNumber: orderRecord.OrderNumber, Email: orderRecord.ShippingAddress.Email, VerificationToken: verificationURL.Query().Get("verification_token"), Description: "private proof"},
		afterSalesInput: CreateCustomerAfterSalesRequestInput{OrderID: orderRecord.ID, Reason: "damage", Description: "private proof", CreatedBy: orderRecord.UserID, Items: []AfterSalesCaseItemInput{{OrderItemID: orderRecord.Items[0].ID, Quantity: 1}}}}
}

func createPrivateOrderBusinessAttachmentReceiptTestEnvironment(t *testing.T) privateOrderBusinessAttachmentReceiptTestEnvironment {
	t.Helper()
	db, afterSales := newAfterSalesService(t)
	return configurePrivateOrderBusinessAttachmentReceiptTestEnvironment(t, db, afterSales)
}

func (e privateOrderBusinessAttachmentReceiptTestEnvironment) uploadTwoPrivateBusinessAttachmentFiles(t *testing.T, ctx context.Context, owner string) error {
	t.Helper()
	files := PrivateOrderBusinessAttachmentUploadFiles{Images: []*multipart.FileHeader{newEvidencePNGFileHeader(t, "first.png"), newEvidencePNGFileHeader(t, "second.png")}}
	switch owner {
	case "warranty":
		_, err := e.warranty.CreateVerifiedWarrantyClaimWithPrivateAttachmentFiles(ctx, e.claimInput, files)
		return err
	case "after-sales":
		_, err := e.afterSales.CreateCustomerAfterSalesRequestWithPrivateAttachmentFiles(ctx, e.afterSalesInput, files)
		return err
	case "shipment":
		_, _, err := e.shipment.AddPrivateShipmentImageFilesWithAtomicOwnerBinding(ctx, e.order.ID, files.Images)
		return err
	}
	return fmt.Errorf("unsupported private attachment test owner")
}

func loadPrivateOrderBusinessAttachmentReceipt(t *testing.T, db *gorm.DB) (outbox.Event, outbox.ObjectStorageCleanupPayload) {
	t.Helper()
	var events []outbox.Event
	require.NoError(t, db.Where("event_key LIKE ?", repository.PrivateOrderBusinessAttachmentReceiptEventKeyPrefix+"%").Find(&events).Error)
	require.Len(t, events, 1)
	var payload outbox.ObjectStorageCleanupPayload
	require.NoError(t, json.Unmarshal(events[0].Payload, &payload))
	return events[0], payload
}

func TestPrivateOrderBusinessAttachmentBatchDurableReceiptRollbackAndRetryForAllThreeOwners(t *testing.T) {
	for _, owner := range []string{"warranty", "after-sales", "shipment"} {
		for _, failure := range []string{"lost response", "second lost response", "empty response", "foreign response", "corrupt bytes", "receipt exchange rollback"} {
			t.Run(owner+"/"+failure, func(t *testing.T) {
				env := createPrivateOrderBusinessAttachmentReceiptTestEnvironment(t)
				env.provider.mode = failure
				if failure == "receipt exchange rollback" {
					require.NoError(t, env.db.Callback().Delete().Before("gorm:delete").Register("fail_private_attachment_receipt_exchange", func(tx *gorm.DB) {
						if tx.Statement.Table == "outbox_events" {
							tx.AddError(errors.New("receipt exchange failed"))
						}
					}))
					defer env.db.Callback().Delete().Remove("fail_private_attachment_receipt_exchange")
				}
				require.Error(t, env.uploadTwoPrivateBusinessAttachmentFiles(t, t.Context(), owner))
				require.NotEmpty(t, env.provider.keys, "the provider must actually have accepted bytes in this failure scenario")
				receipt, payload := loadPrivateOrderBusinessAttachmentReceipt(t, env.db)
				require.Len(t, payload.ObjectKeys, 2)
				require.Zero(t, env.provider.deleteCalls)
				for _, table := range []string{"warranty_claims", "after_sales_cases", "after_sales_case_attachments", "shipment_records"} {
					var count int64
					require.NoError(t, env.db.Table(table).Count(&count).Error)
					require.Zero(t, count, table)
				}
				if owner == "warranty" {
					require.NoError(t, env.warranty.ValidateWarrantyOrderToken(env.claimInput.VerificationToken))
				}
				handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
				handler.ConfigurePrivateOrderBusinessAttachmentUploads(env.uploads)
				require.NoError(t, handler.Handle(t.Context(), receipt))
				for _, key := range payload.ObjectKeys {
					_, err := env.provider.Open(t.Context(), key)
					require.Error(t, err)
				}
				if failure == "receipt exchange rollback" {
					require.NoError(t, env.db.Callback().Delete().Remove("fail_private_attachment_receipt_exchange"))
				}
				env.provider.mode = ""
				require.NoError(t, env.uploadTwoPrivateBusinessAttachmentFiles(t, t.Context(), owner))
				retained, _ := loadPrivateOrderBusinessAttachmentReceipt(t, env.db)
				require.Equal(t, receipt.ID, retained.ID)
			})
		}
	}
}

func TestPrivateOrderBusinessAttachmentCleanupRetainsActiveAndHistoricalBindingsUntilFinalDetach(t *testing.T) {
	for _, owner := range []string{"warranty", "after-sales", "shipment"} {
		t.Run(owner, func(t *testing.T) {
			env := createPrivateOrderBusinessAttachmentReceiptTestEnvironment(t)
			env.provider.mode = "lost response"
			require.Error(t, env.uploadTwoPrivateBusinessAttachmentFiles(t, t.Context(), owner))
			receipt, payload := loadPrivateOrderBusinessAttachmentReceipt(t, env.db)
			reference := env.provider.GetURL(payload.ObjectKeys[0])
			encoded, err := json.Marshal([]string{reference})
			require.NoError(t, err)
			var remove func()
			switch owner {
			case "warranty":
				claim := warranty.WarrantyClaim{OrderNumber: env.order.OrderNumber, Email: env.claimInput.Email, Images: string(encoded)}
				require.NoError(t, env.db.Create(&claim).Error)
				require.NoError(t, env.db.Delete(&claim).Error)
				remove = func() { require.NoError(t, env.db.Unscoped().Delete(&claim).Error) }
			case "after-sales":
				attachment := aftersales.AfterSalesCaseAttachment{CaseID: 991, Kind: aftersales.AttachmentKindImage, StorageURL: reference}
				require.NoError(t, env.db.Create(&attachment).Error)
				remove = func() { require.NoError(t, env.db.Delete(&attachment).Error) }
			case "shipment":
				record := shippingdomain.ShipmentRecord{OrderID: env.order.ID, OrderNumber: env.order.OrderNumber, ShippingImages: encoded}
				require.NoError(t, env.db.Create(&record).Error)
				remove = func() { require.NoError(t, env.db.Delete(&record).Error) }
			}
			handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
			handler.ConfigurePrivateOrderBusinessAttachmentUploads(env.uploads)
			require.Error(t, handler.Handle(t.Context(), receipt))
			require.Zero(t, env.provider.deleteCalls)
			remove()
			require.NoError(t, handler.Handle(t.Context(), receipt))
			retained, _ := loadPrivateOrderBusinessAttachmentReceipt(t, env.db)
			require.Equal(t, receipt.ID, retained.ID)
		})
	}
}

func TestPrivateOrderBusinessAttachmentReceiptPersistenceAndBusinessValidationFailBeforeProviderIO(t *testing.T) {
	for _, owner := range []string{"warranty", "after-sales", "shipment"} {
		for _, failure := range []string{"receipt persistence", "business authorization"} {
			t.Run(owner+"/"+failure, func(t *testing.T) {
				env := createPrivateOrderBusinessAttachmentReceiptTestEnvironment(t)
				if failure == "receipt persistence" {
					require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register("fail_private_attachment_receipt_insert", func(tx *gorm.DB) {
						if tx.Statement.Table == "outbox_events" {
							tx.AddError(errors.New("receipt persistence unavailable"))
						}
					}))
					defer env.db.Callback().Create().Remove("fail_private_attachment_receipt_insert")
				} else {
					env.claimInput.VerificationToken = "invalid-token"
					env.afterSalesInput.OrderID = 99999
					if owner == "shipment" {
						env.order.ID = 99999
					}
				}
				require.Error(t, env.uploadTwoPrivateBusinessAttachmentFiles(t, t.Context(), owner))
				require.Empty(t, env.provider.keys)
				require.Zero(t, env.provider.deleteCalls)
			})
		}
	}
}

func TestPrivateOrderBusinessAttachmentRejectsCopiedPreallocatedReferencesAndOtherCustomerUploads(t *testing.T) {
	env := createPrivateOrderBusinessAttachmentReceiptTestEnvironment(t)
	require.NoError(t, env.uploadTwoPrivateBusinessAttachmentFiles(t, t.Context(), "shipment"))
	record, err := env.shipment.GetAdmin(env.order.ID)
	require.NoError(t, err)
	images, err := decodeShipmentStringList(record.ShippingImages)
	require.NoError(t, err)
	_, err = env.shipment.Update(env.order.ID, ShipmentRecordUpdateInput{ShippingImages: images})
	require.NoError(t, err, "already bound exact references remain editable")
	reference := images[0]
	_, err = env.shipment.Update(env.order.ID, ShipmentRecordUpdateInput{ShippingImages: []string{strings.Replace(reference, "owned-upload-", "owned%2dupload-", 1)}})
	require.Error(t, err)
	forgedClaim := env.claimInput
	forgedClaim.ImageURLs = images
	_, err = env.warranty.CreateWarrantyClaimForOrder(forgedClaim)
	require.Error(t, err)
	forgedRequest := env.afterSalesInput
	forgedRequest.Attachments = []aftersales.AfterSalesCaseAttachment{{Kind: aftersales.AttachmentKindImage, StorageURL: reference}}
	_, err = env.afterSales.CreateCustomerRequest(forgedRequest)
	require.Error(t, err)
	forgedRequest.CreatedBy = env.order.UserID + 100
	_, err = env.afterSales.CreateCustomerAfterSalesRequestWithPrivateAttachmentFiles(t.Context(), forgedRequest, PrivateOrderBusinessAttachmentUploadFiles{Images: []*multipart.FileHeader{newEvidencePNGFileHeader(t, "unowned.png")}})
	require.Error(t, err)
	require.Len(t, env.provider.keys, 2, "another customer must not write any more objects")
}

func TestPrivateOrderBusinessAttachmentWarrantyAndAfterSalesVideoOnlyUploadsPreserveDetectedFormats(t *testing.T) {
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
				if owner == "warranty" {
					claim, err := env.warranty.CreateVerifiedWarrantyClaimWithPrivateAttachmentFiles(t.Context(), env.claimInput, files)
					require.NoError(t, err)
					require.NotEmpty(t, claim.VideoURL)
				} else {
					record, err := env.afterSales.CreateCustomerAfterSalesRequestWithPrivateAttachmentFiles(t.Context(), env.afterSalesInput, files)
					require.NoError(t, err)
					require.Len(t, record.Attachments, 1)
					require.Equal(t, video.mime, record.Attachments[0].ContentType)
					require.Equal(t, aftersales.AttachmentKindVideo, record.Attachments[0].Kind)
				}
				var receipts int64
				require.NoError(t, env.db.Model(&outbox.Event{}).Where("event_key LIKE ?", repository.PrivateOrderBusinessAttachmentReceiptEventKeyPrefix+"%").Count(&receipts).Error)
				require.Zero(t, receipts)
			})
		}
	}
}
