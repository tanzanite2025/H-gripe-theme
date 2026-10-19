package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"
)

func createIsolatedPostgreSQLStructuredMessageMediaTestServices(ctx context.Context, t *testing.T) (customerServicePrivateAttachmentAccessTestEnvironment, *TicketService, *media.MediaAsset) {
	t.Helper()
	env, _, pending := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, env.db.AutoMigrate(&ticket.AutoReplyRule{}))
	env.db = env.db.WithContext(ctx)
	env.mediaService = NewMediaService(repository.NewMediaRepository(env.db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
	env.mediaService.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(env.db))
	tickets := NewTicketService(repository.NewTicketRepository(env.db), repository.NewUserRepository(env.db))
	tickets.ConfigureCustomerServiceRealtimeOutbox(repository.NewOutboxRepository(env.db))
	tickets.ConfigureMediaService(env.mediaService)
	return env, tickets, pending
}

type structuredMessageMediaTestAcquiredAssetLock struct {
	BackendPID int
	AssetID    uint
}

func createPostgreSQLStructuredMessageImageWithConfiguredStorageURL(t *testing.T, env customerServicePrivateAttachmentAccessTestEnvironment, key string) *media.MediaAsset {
	t.Helper()
	asset := createStructuredMessageOperationalImage(t, env.db, env.publicRoot, key)
	asset.URL = env.mediaService.storage.GetURL(key)
	require.NoError(t, env.db.Model(asset).Update("url", asset.URL).Error)
	return asset
}

func pauseFirstStructuredMessageMediaTestAssetRowLock(ctx context.Context, t *testing.T, db *gorm.DB) (<-chan structuredMessageMediaTestAcquiredAssetLock, func()) {
	t.Helper()
	locked := make(chan structuredMessageMediaTestAcquiredAssetLock, 1)
	resume := make(chan struct{})
	var pauseOnce, resumeOnce sync.Once
	release := func() { resumeOnce.Do(func() { close(resume) }) }
	t.Cleanup(release)
	name := "pause_first_structured_media_asset_lock_" + uuid.NewString()
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(name, func(tx *gorm.DB) {
		if tx.Error != nil || tx.Statement.Table != "media_assets" || !strings.Contains(tx.Statement.SQL.String(), "FOR UPDATE") {
			return
		}
		pauseOnce.Do(func() {
			asset, ok := tx.Statement.Dest.(*media.MediaAsset)
			if !ok {
				tx.AddError(fmt.Errorf("unexpected asset lock result type"))
				return
			}
			var pid int
			if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
				tx.AddError(err)
				return
			}
			locked <- structuredMessageMediaTestAcquiredAssetLock{BackendPID: pid, AssetID: asset.ID}
			select {
			case <-resume:
			case <-ctx.Done():
				tx.AddError(ctx.Err())
			}
		})
	}))
	t.Cleanup(func() { _ = db.Callback().Query().Remove(name) })
	return locked, release
}

func awaitStructuredMessageMediaTestAssetRowLock(ctx context.Context, t *testing.T, locked <-chan structuredMessageMediaTestAcquiredAssetLock) structuredMessageMediaTestAcquiredAssetLock {
	t.Helper()
	select {
	case result := <-locked:
		return result
	case <-ctx.Done():
		t.Fatal("first transaction did not acquire its media lock")
		return structuredMessageMediaTestAcquiredAssetLock{}
	}
}

func awaitStructuredMessageMediaTestMutationResult(ctx context.Context, t *testing.T, results <-chan error) error {
	t.Helper()
	select {
	case err := <-results:
		return err
	case <-ctx.Done():
		t.Fatal("structured media transaction did not finish")
		return ctx.Err()
	}
}

func TestPostgreSQLStructuredMetadataRuleAndAgentOrPublicMessageBindingsSerializeWithRetirementInBothOrders(t *testing.T) {
	for _, mutation := range []string{"create-rule", "update-rule", "agent-message", "public-message", "legacy-public-attachment"} {
		for _, first := range []string{"binding", "retirement"} {
			t.Run(mutation+"/"+first+"-first", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				env, tickets, _ := createIsolatedPostgreSQLStructuredMessageMediaTestServices(ctx, t)
				asset := createPostgreSQLStructuredMessageImageWithConfiguredStorageURL(t, env, "content/structured-race.png")
				if mutation == "legacy-public-attachment" {
					require.NoError(t, env.db.Model(asset).Updates(map[string]any{"upload_origin": media.MediaUploadOriginCustomerServiceConversation, "uploader_actor_type": media.MediaUploadActorVisitor, "uploader_visitor_session_hash": env.owner.VisitorSessionHash, "library_partition": media.MediaLibraryPartitionCustomerSubmissions}).Error)
				}
				var existing *ticket.AutoReplyRule
				if mutation == "update-rule" {
					var err error
					existing, err = tickets.CreateAutoReplyRule(structuredAutoReplyInputForMediaTest(`{"url":"/products/before"}`))
					require.NoError(t, err)
				}
				bind := func() error {
					input := structuredAutoReplyInputForMediaTest(structuredProductMetadataForMediaTest(t, asset.URL))
					switch mutation {
					case "create-rule":
						_, err := tickets.CreateAutoReplyRule(input)
						return err
					case "update-rule":
						_, err := tickets.UpdateAutoReplyRule(existing.ID, input)
						return err
					case "agent-message":
						return tickets.AddCustomerServiceAgentMessage(&ticket.TicketMessage{TicketID: env.conversation.ID, Content: "Racing structured card", MessageType: "product", Metadata: input.Metadata}, env.conversation.AssignedTo, false)
					case "legacy-public-attachment":
						_, _, err := tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Racing structured card", 0, "image", fmt.Sprintf(`{"url":%q}`, asset.URL), fmt.Sprintf(`[%q]`, asset.URL))
						return err
					default:
						_, _, err := tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Racing structured card", env.conversation.AssignedTo, "product", input.Metadata, "")
						return err
					}
				}
				locked, release := pauseFirstStructuredMessageMediaTestAssetRowLock(ctx, t, env.db)
				bindResults, retireResults := make(chan error, 1), make(chan error, 1)
				startBinding := func() { go func() { bindResults <- bind() }() }
				startRetirement := func() {
					go func() {
						retireResults <- env.mediaService.DeleteAsset(ctx, asset.ID, MediaAssetDeleteConfirmation(asset.ID))
					}()
				}
				if first == "binding" {
					startBinding()
				} else {
					startRetirement()
				}
				acquired := awaitStructuredMessageMediaTestAssetRowLock(ctx, t, locked)
				require.Equal(t, asset.ID, acquired.AssetID)
				if first == "binding" {
					startRetirement()
				} else {
					startBinding()
				}
				waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, env.db, acquired.BackendPID, "media_assets")
				release()
				bindErr := awaitStructuredMessageMediaTestMutationResult(ctx, t, bindResults)
				retireErr := awaitStructuredMessageMediaTestMutationResult(ctx, t, retireResults)
				var cleanupEvents, messages, realtimeEvents int64
				require.NoError(t, env.db.Model(&outbox.Event{}).Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(asset.ID)).Count(&cleanupEvents).Error)
				require.NoError(t, env.db.Model(&ticket.TicketMessage{}).Where("content = ?", "Racing structured card").Count(&messages).Error)
				require.NoError(t, env.db.Model(&outbox.Event{}).Where("event_type = ?", outbox.EventTypeCustomerServiceRealtime).Count(&realtimeEvents).Error)
				if first == "binding" {
					require.NoError(t, bindErr)
					require.ErrorIs(t, retireErr, ErrMediaAssetInUse)
					require.Zero(t, cleanupEvents)
				} else {
					require.NoError(t, retireErr)
					if strings.Contains(mutation, "rule") {
						require.ErrorIs(t, bindErr, ErrInvalidAutoReplyRule)
					} else {
						require.ErrorIs(t, bindErr, ErrCustomerServiceMessageMediaReferenceInvalid)
					}
					require.EqualValues(t, 1, cleanupEvents)
					require.Zero(t, messages)
					require.Zero(t, realtimeEvents)
					require.Error(t, bind(), "physical cleanup must not make a stale managed URL reusable")
					if existing != nil {
						stored, err := tickets.GetAutoReplyRule(existing.ID)
						require.NoError(t, err)
						require.Equal(t, existing.Metadata, stored.Metadata)
					}
				}
			})
		}
	}
}

func TestPostgreSQLStructuredMetadataAndLegacyAttachmentsUseOneAscendingLockOrderAcrossOppositeFieldOrders(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	env, tickets, _ := createIsolatedPostgreSQLStructuredMessageMediaTestServices(ctx, t)
	low := createPostgreSQLStructuredMessageImageWithConfiguredStorageURL(t, env, "content/structured-low.png")
	high := createPostgreSQLStructuredMessageImageWithConfiguredStorageURL(t, env, "content/structured-high.png")
	first := structuredAutoReplyInputForMediaTest(structuredProductMetadataForMediaTest(t, high.URL))
	first.Attachments = fmt.Sprintf(`[%q]`, low.URL)
	second := structuredAutoReplyInputForMediaTest(structuredProductMetadataForMediaTest(t, low.URL))
	second.Attachments = fmt.Sprintf(`[%q]`, high.URL)
	locked, release := pauseFirstStructuredMessageMediaTestAssetRowLock(ctx, t, env.db)
	results := make(chan error, 2)
	go func() { _, err := tickets.CreateAutoReplyRule(first); results <- err }()
	acquired := awaitStructuredMessageMediaTestAssetRowLock(ctx, t, locked)
	require.Equal(t, low.ID, acquired.AssetID)
	go func() { _, err := tickets.CreateAutoReplyRule(second); results <- err }()
	waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, env.db, acquired.BackendPID, "media_assets")
	release()
	require.NoError(t, awaitStructuredMessageMediaTestMutationResult(ctx, t, results))
	require.NoError(t, awaitStructuredMessageMediaTestMutationResult(ctx, t, results))
	for _, asset := range []*media.MediaAsset{low, high} {
		report, err := env.mediaService.GetAssetReferences(asset.ID)
		require.NoError(t, err)
		require.Equal(t, 2, report.Total)
	}
}

func TestPostgreSQLStructuredCardMediaAndPrivateUploadSessionShareAssetLockOrderWithoutConsumingReceiptEarly(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	env, tickets, pending := createIsolatedPostgreSQLStructuredMessageMediaTestServices(ctx, t)
	card := createPostgreSQLStructuredMessageImageWithConfiguredStorageURL(t, env, "content/card-with-private-attachment.png")
	require.Less(t, pending.ID, card.ID)
	blocker := env.db.Begin()
	require.NoError(t, blocker.Error)
	defer blocker.Rollback()
	var asset media.MediaAsset
	require.NoError(t, blocker.Clauses(clause.Locking{Strength: "UPDATE"}).First(&asset, pending.ID).Error)
	var pid int
	require.NoError(t, blocker.Raw("SELECT pg_backend_pid()").Scan(&pid).Error)
	results := make(chan error, 1)
	metadata := structuredProductMetadataForMediaTest(t, card.URL)
	go func() {
		_, _, err := tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Card and private photo", 0, "product", metadata, "", ticket.CustomerServiceMessageMediaAttachmentBindingRequest{UploadSessionID: pending.PendingCustomerServiceUploadSessionID})
		results <- err
	}()
	waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, env.db, pid, "media_assets")
	probe := env.db.Begin()
	require.NoError(t, probe.Error)
	var high media.MediaAsset
	require.NoError(t, probe.Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).First(&high, card.ID).Error, "writer must wait for the lower private asset before taking the card asset")
	require.NoError(t, probe.Rollback().Error)
	session, err := env.mediaService.repo.FindCustomerServiceMediaUploadSession(ctx, pending.PendingCustomerServiceUploadSessionID)
	require.NoError(t, err)
	require.Equal(t, "ready", session.Status)
	require.NoError(t, blocker.Commit().Error)
	require.NoError(t, awaitStructuredMessageMediaTestMutationResult(ctx, t, results))
	session, err = env.mediaService.repo.FindCustomerServiceMediaUploadSession(ctx, pending.PendingCustomerServiceUploadSessionID)
	require.NoError(t, err)
	require.Equal(t, "bound", session.Status)
	var message ticket.TicketMessage
	require.NoError(t, env.db.First(&message, "content = ?", "Card and private photo").Error)
	access := NewCustomerServiceConversationMessageAttachmentAccessService(tickets, env.mediaService)
	attachments, _ := access.DescribePublicCustomerServiceMessageAttachments(message, *env.conversation.ConversationID)
	require.Len(t, attachments, 1)
	for index := 0; index < 2; index++ {
		file, err := access.OpenPublicCustomerServiceMessageAttachment(ctx, *env.conversation.ConversationID, env.owner, message.ID, index, "")
		require.NoError(t, err)
		require.NoError(t, file.ReadCloser.Close())
	}
}

func TestPostgreSQLStructuredAutoReplyRechecksLockedRuleAfterConcurrentDetachmentDeletionOrDisable(t *testing.T) {
	for _, mutation := range []string{"detach", "delete", "disable"} {
		t.Run(mutation, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			env, tickets, _ := createIsolatedPostgreSQLStructuredMessageMediaTestServices(ctx, t)
			asset := createPostgreSQLStructuredMessageImageWithConfiguredStorageURL(t, env, "content/stale-rule.png")
			rule, err := tickets.CreateAutoReplyRule(structuredAutoReplyInputForMediaTest(structuredProductMetadataForMediaTest(t, asset.URL)))
			require.NoError(t, err)
			blocker := env.db.Begin()
			require.NoError(t, blocker.Error)
			defer blocker.Rollback()
			var current ticket.AutoReplyRule
			require.NoError(t, blocker.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, rule.ID).Error)
			var pid int
			require.NoError(t, blocker.Raw("SELECT pg_backend_pid()").Scan(&pid).Error)
			switch mutation {
			case "detach":
				require.NoError(t, blocker.Model(&current).Update("metadata", `{"url":"/products/detached"}`).Error)
			case "delete":
				require.NoError(t, blocker.Delete(&current).Error)
			default:
				require.NoError(t, blocker.Model(&current).Update("is_active", false).Error)
			}
			results := make(chan error, 1)
			go func() {
				_, _, message, err := tickets.deliverAutoReply(&env.conversation, *rule, "en", "welcome")
				if err == nil && message != nil {
					err = fmt.Errorf("stale rule created a new message")
				}
				results <- err
			}()
			waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, env.db, pid, "ticket_auto_replies")
			require.NoError(t, blocker.Commit().Error)
			require.NoError(t, awaitStructuredMessageMediaTestMutationResult(ctx, t, results))
			var events, replies int64
			require.NoError(t, env.db.Model(&outbox.Event{}).Where("event_type = ?", outbox.EventTypeCustomerServiceRealtime).Count(&events).Error)
			require.NoError(t, env.db.Model(&ticket.TicketMessage{}).Where("is_staff = ?", true).Count(&replies).Error)
			require.Zero(t, events)
			require.Zero(t, replies)
		})
	}
}

func TestPostgreSQLStructuredAutoReplyFreshCreateAndUpdateSnapshotsRespectDatabaseTimestampPrecision(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	env, tickets, _ := createIsolatedPostgreSQLStructuredMessageMediaTestServices(ctx, t)
	asset := createPostgreSQLStructuredMessageImageWithConfiguredStorageURL(t, env, "content/fresh-rule.png")
	for _, mutation := range []string{"create", "update"} {
		input := structuredAutoReplyInputForMediaTest(structuredProductMetadataForMediaTest(t, asset.URL))
		rule, err := tickets.CreateAutoReplyRule(input)
		require.NoError(t, err)
		if mutation == "update" {
			input.ReplyMessage = "Updated fresh rule"
			rule, err = tickets.UpdateAutoReplyRule(rule.ID, input)
			require.NoError(t, err)
		}
		_, _, message, err := tickets.deliverAutoReply(&env.conversation, *rule, "en", "welcome")
		require.NoError(t, err)
		require.NotNil(t, message, "a current rule must send successfully with database timestamp precision")
		require.Contains(t, message.Metadata, "9007199254740997")
	}
}
