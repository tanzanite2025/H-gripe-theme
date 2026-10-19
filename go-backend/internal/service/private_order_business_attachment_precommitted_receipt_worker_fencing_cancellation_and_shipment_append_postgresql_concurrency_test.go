package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
)

func createIsolatedPostgreSQLPrivateOrderBusinessAttachmentReceiptEnvironment(t *testing.T) privateOrderBusinessAttachmentReceiptTestEnvironment {
	t.Helper()
	base, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, base.db.AutoMigrate(&order.Order{}, &order.OrderItem{}, &aftersales.AfterSalesCase{}, &aftersales.AfterSalesCaseItem{}, &aftersales.AfterSalesCaseEvent{}, &aftersales.AfterSalesCaseAttachment{}, &aftersales.AfterSalesReturnShipment{}, &aftersales.AfterSalesRefundReview{}))
	_, afterSales := newAfterSalesService(t)
	afterSales = afterSales.afterSalesServiceWithPrivateAttachmentTransaction(base.db)
	return configurePrivateOrderBusinessAttachmentReceiptTestEnvironment(t, base.db, afterSales)
}

func capturePrivateOrderBusinessAttachmentWriterPostgreSQLBackendPID(t *testing.T, db *gorm.DB) <-chan int {
	t.Helper()
	result := make(chan int, 1)
	var once sync.Once
	name := "capture_private_order_attachment_writer_" + uuid.NewString()
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(name, func(tx *gorm.DB) {
		if tx.Statement.Table == "outbox_events" {
			if _, active := tx.Statement.ConnPool.(gorm.TxCommitter); active {
				once.Do(func() {
					var pid int
					if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
						tx.AddError(err)
					} else {
						result <- pid
					}
				})
			}
		}
	}))
	t.Cleanup(func() { _ = db.Callback().Query().Remove(name) })
	return result
}

func TestPostgreSQLPrivateOrderBusinessAttachmentReceiptVisibleBeforeIOWorkerSkipsAndCancellationRetainsLock(t *testing.T) {
	for _, owner := range []string{"warranty", "after-sales", "shipment"} {
		t.Run(owner, func(t *testing.T) {
			env := createIsolatedPostgreSQLPrivateOrderBusinessAttachmentReceiptEnvironment(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			entered, released := make(chan struct{}), make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(released) }) }
			defer release()
			captured := capturePrivateOrderBusinessAttachmentWriterPostgreSQLBackendPID(t, env.db)
			env.provider.afterWrite = func(context.Context, string) { close(entered); <-released }
			writer := make(chan error, 1)
			go func() { writer <- env.uploadTwoPrivateBusinessAttachmentFiles(ctx, t, owner) }()
			select {
			case <-entered:
			case err := <-writer:
				t.Fatalf("private attachment writer failed before provider IO: %v", err)
			case <-time.After(10 * time.Second):
				t.Fatal("private attachment provider did not accept bytes")
			}
			pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t.Context(), t, captured)
			receipt, payload := loadPrivateOrderBusinessAttachmentReceipt(t, env.db)
			require.Contains(t, payload.ObjectKeys, env.provider.keys[0])
			claimed, err := repository.NewOutboxRepository(env.db).ClaimReadyEvents(time.Now().Add(6*time.Minute), "private-business-receipt-worker", 100, time.Minute)
			require.NoError(t, err)
			require.Empty(t, claimed)
			cancel()
			cleanup := make(chan error, 1)
			handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
			handler.ConfigurePrivateOrderBusinessAttachmentUploads(env.uploads)
			go func() { cleanup <- handler.Handle(t.Context(), receipt) }()
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t.Context(), t, env.db, pid)
			release()
			select {
			case err := <-writer:
				require.ErrorIs(t, err, context.Canceled)
			case <-time.After(10 * time.Second):
				t.Fatal("canceled attachment writer stuck")
			}
			select {
			case err := <-cleanup:
				require.NoError(t, err)
			case <-time.After(10 * time.Second):
				t.Fatal("attachment cleanup stuck")
			}
			for _, key := range payload.ObjectKeys {
				_, err := env.provider.Open(t.Context(), key)
				require.Error(t, err)
			}
			for _, table := range []string{"warranty_claims", "after_sales_cases", "shipment_records"} {
				var count int64
				require.NoError(t, env.db.Table(table).Count(&count).Error)
				require.Zero(t, count)
			}
		})
	}
}

func TestPostgreSQLPrivateOrderBusinessAttachmentWaitingWriterRejectsClaimedCompletedRetriedExpiredMissingAndChangedReceipts(t *testing.T) {
	for _, fence := range []string{"claimed", "completed", "retried", "expired", "missing", "changed"} {
		t.Run(fence, func(t *testing.T) {
			env := createIsolatedPostgreSQLPrivateOrderBusinessAttachmentReceiptEnvironment(t)
			locked, released := make(chan int, 1), make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(released) }) }
			defer release()
			holder := make(chan error, 1)
			go func() {
				holder <- env.uploads.outbox.WithPrivateOrderBusinessAttachmentCleanupLockUntilProviderReturns(t.Context(), func(tx *gorm.DB) error {
					var pid int
					if err := tx.Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
						return err
					}
					locked <- pid
					<-released
					return nil
				})
			}()
			pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t.Context(), t, locked)
			writer := make(chan error, 1)
			go func() { writer <- env.uploadTwoPrivateBusinessAttachmentFiles(t.Context(), t, "shipment") }()
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t.Context(), t, env.db, pid)
			receipt, _ := loadPrivateOrderBusinessAttachmentReceipt(t, env.db)
			query := env.db.Model(&outbox.Event{}).Where("id = ?", receipt.ID)
			switch fence {
			case "claimed":
				require.NoError(t, query.Updates(map[string]any{"status": outbox.EventStatusProcessing, "attempts": 1}).Error)
			case "completed":
				require.NoError(t, query.Update("status", outbox.EventStatusProcessed).Error)
			case "retried":
				require.NoError(t, query.Updates(map[string]any{"attempts": 0, "last_attempt_at": time.Now()}).Error)
			case "expired":
				require.NoError(t, query.Update("available_at", time.Now().Add(-time.Minute)).Error)
			case "missing":
				require.NoError(t, env.db.Delete(&receipt).Error)
			case "changed":
				require.NoError(t, query.Update("payload", []byte(`{"resource_type":"different"}`)).Error)
			}
			release()
			require.NoError(t, <-holder)
			require.ErrorIs(t, <-writer, repository.ErrPrivateOrderBusinessAttachmentReceiptUnavailable)
			require.Empty(t, env.provider.keys)
		})
	}
}

func TestPostgreSQLPrivateShipmentConcurrentAppendPreservesBothSuccessfulAttachmentBatches(t *testing.T) {
	env := createIsolatedPostgreSQLPrivateOrderBusinessAttachmentReceiptEnvironment(t)
	entered, released := make(chan struct{}), make(chan struct{})
	var releaseOnce, blockOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(released) }) }
	defer release()
	captured := capturePrivateOrderBusinessAttachmentWriterPostgreSQLBackendPID(t, env.db)
	env.provider.beforeWrite = func(context.Context, string) { blockOnce.Do(func() { close(entered); <-released }) }
	first, second := make(chan error, 1), make(chan error, 1)
	go func() { first <- env.uploadTwoPrivateBusinessAttachmentFiles(t.Context(), t, "shipment") }()
	select {
	case <-entered:
	case <-time.After(10 * time.Second):
		t.Fatal("first shipment append did not reach provider")
	}
	pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t.Context(), t, captured)
	go func() { second <- env.uploadTwoPrivateBusinessAttachmentFiles(t.Context(), t, "shipment") }()
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t.Context(), t, env.db, pid)
	release()
	require.NoError(t, <-first)
	require.NoError(t, <-second)
	record, err := env.shipment.GetAdmin(env.order.ID)
	require.NoError(t, err)
	images, err := decodeShipmentStringList(record.ShippingImages)
	require.NoError(t, err)
	require.Len(t, images, 4)
	var count int64
	require.NoError(t, env.db.Model(&outbox.Event{}).Where("event_key LIKE ?", repository.PrivateOrderBusinessAttachmentReceiptEventKeyPrefix+"%").Count(&count).Error)
	require.Zero(t, count)
}
