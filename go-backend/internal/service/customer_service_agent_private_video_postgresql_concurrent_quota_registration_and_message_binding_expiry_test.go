package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
)

func TestPostgreSQLCustomerServiceAgentPrivateVideoBindingAndExpirySerializeOnAuthoritativeAssetOwnership(t *testing.T) {
	for _, firstOperation := range []string{"bind", "expire"} {
		t.Run(firstOperation+" first", func(t *testing.T) {
			env, retention, photo := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
			require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Where("id = ?", photo.PendingCustomerServiceUploadSessionID).Update("expires_at", time.Now().Add(7*24*time.Hour)).Error)
			uploads := NewCustomerServiceConversationAttachmentUploadService(env.access.tickets, env.mediaService)
			asset, err := uploads.UploadVerifiedAgentCustomerServiceConversationAttachment(t.Context(), env.conversation.ID, env.conversation.AssignedTo, false,
				multipartFileHeader(t, "racing-private-video.mp4", "video/mp4", syntheticCustomerServiceAgentVideoMP4HeaderBytes()))
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			firstLocked := make(chan int, 1)
			releaseFirst := make(chan struct{})
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
			t.Cleanup(release)
			pauseTransaction := func(tx *gorm.DB) {
				var pid int
				if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
					tx.AddError(err)
					return
				}
				firstLocked <- pid
				select {
				case <-releaseFirst:
				case <-ctx.Done():
					tx.AddError(ctx.Err())
				}
			}
			callback := "pause_agent_private_video_ownership_" + firstOperation
			if firstOperation == "bind" {
				require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "ticket_message_media_attachments" {
						pauseTransaction(tx)
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
			} else {
				require.NoError(t, env.db.Callback().Delete().After("gorm:delete").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "media_upload_sessions" {
						pauseTransaction(tx)
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Delete().Remove(callback) })
			}
			bindResult := make(chan error, 1)
			type expiryOutcome struct {
				result CustomerServicePendingAttachmentUploadExpiryBatchResult
				err    error
			}
			expiryResult := make(chan expiryOutcome, 1)
			bind := func() {
				message := ticket.TicketMessage{TicketID: env.conversation.ID, Content: "Racing video", MessageType: "video",
					RequestedMediaAttachmentBindings: []ticket.CustomerServiceMessageMediaAttachmentBindingRequest{{UploadSessionID: asset.PendingCustomerServiceUploadSessionID}}}
				bindResult <- env.access.tickets.AddCustomerServiceAgentMessage(&message, env.conversation.AssignedTo, false)
			}
			expire := func() {
				result, err := retention.ExpireUnboundCustomerServiceAttachmentUploads(ctx, 10)
				expiryResult <- expiryOutcome{result, err}
			}
			if firstOperation == "bind" {
				go bind()
			} else {
				go expire()
			}
			var blockerPID int
			select {
			case blockerPID = <-firstLocked:
			case <-ctx.Done():
				t.Fatal("first operation did not acquire ownership")
			}
			if firstOperation == "bind" {
				go expire()
			} else {
				go bind()
			}
			require.Eventually(t, func() bool {
				var waiting bool
				err := env.db.WithContext(ctx).Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)))", blockerPID).Scan(&waiting).Error
				return err == nil && waiting
			}, 5*time.Second, 20*time.Millisecond)
			release()
			var bindError error
			select {
			case bindError = <-bindResult:
			case <-ctx.Done():
				t.Fatal("binding did not finish")
			}
			var expired expiryOutcome
			select {
			case expired = <-expiryResult:
			case <-ctx.Done():
				t.Fatal("expiry did not finish")
			}
			require.NoError(t, expired.err)
			if firstOperation == "bind" {
				require.NoError(t, bindError)
				require.Zero(t, expired.result.Retired)
				var receipt media.MediaUploadSession
				require.NoError(t, env.db.First(&receipt, "id = ?", asset.PendingCustomerServiceUploadSessionID).Error)
				require.Equal(t, "bound", receipt.Status)
				_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(asset.StorageKey)))
				require.NoError(t, err)
			} else {
				require.Error(t, bindError)
				require.Equal(t, 1, expired.result.Retired)
				var event outbox.Event
				require.NoError(t, env.db.Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, strconv.FormatUint(uint64(asset.ID), 10)).First(&event).Error)
				require.NoError(t, NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil).Handle(t.Context(), event))
				_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(asset.StorageKey)))
				require.True(t, os.IsNotExist(err))
			}
		})
	}
}

type customerServiceAgentPrivateVideoConcurrentWriteBarrier struct {
	storage.StorageService
	written chan string
	release <-chan struct{}
}

func (s *customerServiceAgentPrivateVideoConcurrentWriteBarrier) VerifyStoredObjectExists(ctx context.Context, key string) error {
	return s.StorageService.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(ctx, key)
}

func (s *customerServiceAgentPrivateVideoConcurrentWriteBarrier) WritePrivateObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	url, err := s.StorageService.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(ctx, reader, key)
	if err != nil {
		return "", err
	}
	s.written <- key
	select {
	case <-s.release:
		return url, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func TestPostgreSQLCustomerServiceAgentPrivateVideoConcurrentQuotaRegistrationKeepsLosingWriteReceipt(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	payload := syntheticCustomerServiceAgentVideoMP4HeaderBytes()
	env.mediaService.accountStorageQuotaBytes = int64(len(payload))
	releaseWrites := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseWrites) }) }
	t.Cleanup(release)
	store := &customerServiceAgentPrivateVideoConcurrentWriteBarrier{StorageService: env.mediaService.storage, written: make(chan string, 2), release: releaseWrites}
	env.mediaService.storage = store
	uploads := NewCustomerServiceConversationAttachmentUploadService(env.access.tickets, env.mediaService)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	type uploadOutcome struct {
		asset *media.MediaAsset
		err   error
	}
	results := make(chan uploadOutcome, 2)
	for index := range 2 {
		file := multipartFileHeader(t, fmt.Sprintf("concurrent-agent-video-%d.mp4", index), "video/mp4", payload)
		go func() {
			asset, err := uploads.UploadVerifiedAgentCustomerServiceConversationAttachment(ctx, env.conversation.ID, env.conversation.AssignedTo, false, file)
			results <- uploadOutcome{asset, err}
		}()
	}
	for range 2 {
		select {
		case <-store.written:
		case <-ctx.Done():
			t.Fatal("both writes did not reach the registration barrier")
		}
	}
	release()
	succeeded, rejected := 0, 0
	for range 2 {
		var outcome uploadOutcome
		select {
		case outcome = <-results:
		case <-ctx.Done():
			t.Fatal("quota registration did not finish")
		}
		if outcome.err == nil {
			succeeded++
			require.NotNil(t, outcome.asset)
		} else {
			rejected++
			require.ErrorIs(t, outcome.err, ErrMediaAccountStorageQuotaExceeded)
			require.Nil(t, outcome.asset)
		}
	}
	require.Equal(t, 1, succeeded)
	require.Equal(t, 1, rejected)
	usage, err := env.mediaService.repo.AssetStorageUsageByUploadActor(media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: env.conversation.AssignedTo})
	require.NoError(t, err)
	require.Equal(t, int64(len(payload)), usage)
	var failed media.MediaUploadSession
	require.NoError(t, env.db.First(&failed, "status = ? AND upload_actor_account_id = ?", "failed", env.conversation.AssignedTo).Error)
	require.Nil(t, failed.MediaAssetID)
	require.Equal(t, int64(len(payload)), *failed.ExpectedUploadBytes)
	stored, err := os.ReadFile(filepath.Join(env.privateRoot, filepath.FromSlash(failed.ObjectKey)))
	require.NoError(t, err)
	require.Equal(t, payload, stored)
}
