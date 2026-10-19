package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createStructuredTicketMediaBindingTestServices(t *testing.T) (*gorm.DB, *TicketService, *MediaService, ticket.Ticket, string) {
	t.Helper()
	db := newAutoReplyTestDB(t)
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &ticket.CustomerServiceInboxState{}))
	root := t.TempDir()
	assets := newMediaDeleteTestService(t, db, root)
	assets.siteURL = "https://shop.example.test"
	tickets := newAutoReplyTestTicketService(db)
	tickets.ConfigureMediaService(assets)
	agent := createAutoReplyTestUser(t, db, "structured-media-agent@example.test", "structured-media-agent")
	conversationID := "synthetic-structured-media-conversation"
	conversation := ticket.Ticket{UserID: agent.ID, AssignedTo: agent.ID, Subject: "Synthetic structured media", Category: customerServiceTicketCategory, Status: "open", ConversationID: &conversationID, VisitorSessionHash: strings.Repeat("c", 64)}
	require.NoError(t, db.Create(&conversation).Error)
	return db, tickets, assets, conversation, root
}

func createStructuredMessageOperationalImage(t *testing.T, db *gorm.DB, root, key string) *media.MediaAsset {
	t.Helper()
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, key)
	require.NoError(t, db.Model(asset).Updates(map[string]any{"upload_origin": media.MediaUploadOriginAdministratorShared, "classification_method": media.MediaClassificationRecordedUploadContext, "library_partition": media.MediaLibraryPartitionShared, "mime_type": "image/png"}).Error)
	objectPath := filepath.Join(root, filepath.FromSlash(key))
	require.NoError(t, os.MkdirAll(filepath.Dir(objectPath), 0o755))
	require.NoError(t, os.WriteFile(objectPath, []byte("synthetic media object"), 0o644))
	return asset
}

func structuredProductMetadataForMediaTest(t *testing.T, reference string) string {
	t.Helper()
	encoded, err := json.Marshal(map[string]any{"product_id": 9007199254740997, "url": "/products/synthetic", "thumbnail": reference})
	require.NoError(t, err)
	return string(encoded)
}

func structuredAutoReplyInputForMediaTest(metadata string) AutoReplyRuleInput {
	return AutoReplyRuleInput{Type: "welcome", ReplyMessage: "Synthetic structured reply", Locale: "en", MessageType: "product", Metadata: metadata, IsActive: true}
}

func TestStructuredTicketAndAutoReplyMetadataExactReferenceCoverageAndNavigationExclusion(t *testing.T) {
	for _, source := range []string{"message", "rule"} {
		for _, scenario := range []string{"original", "derivative", "trusted-alias", "escaped-slashes", "foreign-host", "longer-path", "navigation-only", "caption-only", "text-unused-thumbnail", "faq", "config", "image", "video-poster"} {
			t.Run(source+"/"+scenario, func(t *testing.T) {
				db, _, assets, conversation, root := createStructuredTicketMediaBindingTestServices(t)
				asset := createStructuredMessageOperationalImage(t, db, root, "content/metadata.png")
				derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, URL: assets.storage.GetURL("derivatives/metadata.webp"), StorageKey: "derivatives/metadata.webp", Preset: "card", PresetVersion: 1}
				require.NoError(t, db.Create(&derivative).Error)
				reference, messageType := asset.URL, "product"
				retains := true
				switch scenario {
				case "derivative":
					reference = derivative.URL + "?revision=1#preview"
				case "trusted-alias":
					reference = "https://shop.example.test/uploads/" + asset.StorageKey
				case "foreign-host":
					reference = "https://foreign.example.test/uploads/" + asset.StorageKey
					retains = false
				case "longer-path":
					reference += ".other.png"
					retains = false
				case "navigation-only", "caption-only", "text-unused-thumbnail":
					retains = false
				case "faq":
					messageType = "faq"
				case "config":
					messageType = "config_confirm"
				case "image":
					messageType = "image"
				case "video-poster":
					messageType = "video"
				}
				metadata := structuredProductMetadataForMediaTest(t, reference)
				switch scenario {
				case "navigation-only":
					metadata = fmt.Sprintf(`{"url":%q}`, reference)
				case "caption-only":
					metadata = fmt.Sprintf(`{"caption":%q}`, reference)
				case "text-unused-thumbnail":
					messageType = "text"
				case "faq":
					metadata = fmt.Sprintf(`{"faq":{"answer_image_url":%q}}`, reference)
				case "config":
					metadata = fmt.Sprintf(`{"product":{"thumbnail":%q}}`, reference)
				case "image":
					metadata = fmt.Sprintf(`{"url":%q}`, reference)
				case "video-poster":
					metadata = fmt.Sprintf(`{"thumbnail":%q}`, reference)
				case "escaped-slashes":
					metadata = strings.ReplaceAll(metadata, "/", `\/`)
				}
				if source == "message" {
					require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: conversation.ID, Content: "Synthetic history", MessageType: messageType, Metadata: metadata}).Error)
				} else {
					require.NoError(t, db.Create(&ticket.AutoReplyRule{Type: "welcome", ReplyMessage: "Synthetic rule", MessageType: messageType, Metadata: metadata}).Error)
				}
				report, err := assets.GetAssetReferences(asset.ID)
				require.NoError(t, err)
				if retains {
					require.Equal(t, 1, report.Total)
					require.Contains(t, report.References[0].Field, "metadata.")
					require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
				} else {
					require.Zero(t, report.Total)
					require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
				}
			})
		}
	}
}

func TestStructuredMediaRuleCreateUpdateAndMessageSendRejectInvalidBindingsAtomically(t *testing.T) {
	for _, mutation := range []string{"create-rule", "update-rule", "agent-message", "public-message"} {
		for _, scenario := range []string{"unregistered", "retired", "private", "video", "dedicated", "duplicate"} {
			t.Run(mutation+"/"+scenario, func(t *testing.T) {
				db, tickets, assets, conversation, root := createStructuredTicketMediaBindingTestServices(t)
				asset := createStructuredMessageOperationalImage(t, db, root, "content/invalid.png")
				reference := asset.URL
				switch scenario {
				case "unregistered":
					reference = "/uploads/content/missing.png"
				case "retired":
					require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
				case "private":
					require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
				case "video":
					require.NoError(t, db.Model(asset).Update("media_type", "video").Error)
				case "dedicated":
					reference = "/uploads/site-logo/historical.png"
				case "duplicate":
					duplicate := createStructuredMessageOperationalImage(t, db, root, "content/duplicate.png")
					require.NoError(t, db.Model(duplicate).Update("storage_key", asset.StorageKey).Error)
				}
				input := structuredAutoReplyInputForMediaTest(structuredProductMetadataForMediaTest(t, reference))
				before := *conversation.ConversationID
				var previous *ticket.AutoReplyRule
				if mutation == "update-rule" {
					var err error
					previous, err = tickets.CreateAutoReplyRule(structuredAutoReplyInputForMediaTest(`{"url":"/products/previous"}`))
					require.NoError(t, err)
				}
				var err error
				switch mutation {
				case "create-rule":
					_, err = tickets.CreateAutoReplyRule(input)
				case "update-rule":
					_, err = tickets.UpdateAutoReplyRule(previous.ID, input)
				case "agent-message":
					err = tickets.AddCustomerServiceAgentMessage(&ticket.TicketMessage{TicketID: conversation.ID, Content: "Must not persist", MessageType: "product", Metadata: input.Metadata}, conversation.AssignedTo, false)
				case "public-message":
					_, _, err = tickets.AddPublicCustomerServiceMessage(before, CustomerServiceOwner{VisitorSessionHash: conversation.VisitorSessionHash}, "Must not persist", conversation.AssignedTo, "product", input.Metadata, "")
				}
				if strings.Contains(mutation, "rule") {
					require.ErrorIs(t, err, ErrInvalidAutoReplyRule)
				} else {
					require.ErrorIs(t, err, ErrCustomerServiceMessageMediaReferenceInvalid)
				}
				var messages, realtimeEvents int64
				require.NoError(t, db.Model(&ticket.TicketMessage{}).Count(&messages).Error)
				require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ?", outbox.EventTypeCustomerServiceRealtime).Count(&realtimeEvents).Error)
				require.Zero(t, messages)
				require.Zero(t, realtimeEvents)
				if previous != nil {
					stored, err := tickets.GetAutoReplyRule(previous.ID)
					require.NoError(t, err)
					require.Equal(t, previous.Metadata, stored.Metadata)
				}
			})
		}
	}
}

func TestStructuredAutoReplyCanonicalMediaDeliveryAndStaleRuleSnapshotsCannotRestoreDetachedMedia(t *testing.T) {
	db, tickets, assets, conversation, root := createStructuredTicketMediaBindingTestServices(t)
	asset := createStructuredMessageOperationalImage(t, db, root, "content/delivery.png")
	rule, err := tickets.CreateAutoReplyRule(structuredAutoReplyInputForMediaTest(structuredProductMetadataForMediaTest(t, "https://shop.example.test/uploads/"+asset.StorageKey)))
	require.NoError(t, err)
	require.Contains(t, rule.Metadata, asset.URL)
	_, _, message, err := tickets.deliverAutoReply(&conversation, *rule, "en", "welcome")
	require.NoError(t, err)
	require.NotNil(t, message)
	require.Contains(t, message.Metadata, `9007199254740997`)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 2, report.Total)
	_, err = tickets.UpdateAutoReplyRule(rule.ID, structuredAutoReplyInputForMediaTest(`{"url":"/products/detached"}`))
	require.NoError(t, err)
	_, _, stale, err := tickets.deliverAutoReply(&conversation, *rule, "en", "welcome")
	require.NoError(t, err)
	require.Nil(t, stale)
	require.NoError(t, tickets.DeleteAutoReplyRule(rule.ID))
	_, _, stale, err = tickets.deliverAutoReply(&conversation, *rule, "en", "welcome")
	require.NoError(t, err)
	require.Nil(t, stale)
	report, err = assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total, "sent message remains the authority after rule removal")
	_, err = tickets.UpdateAutoReplyRule(rule.ID, structuredAutoReplyInputForMediaTest(rule.Metadata))
	require.ErrorIs(t, err, ErrAutoReplyRuleNotFound)
}

func TestStructuredAutoReplyExactHistoricalFieldsMayRemainButCannotBeMovedRestoredOrCopiedToNewMessages(t *testing.T) {
	for _, reference := range []string{"/uploads/site-logo/legacy.png", "/uploads/content/unregistered-history.png"} {
		t.Run(reference, func(t *testing.T) {
			db, tickets, _, conversation, _ := createStructuredTicketMediaBindingTestServices(t)
			rule := ticket.AutoReplyRule{Type: "welcome", ReplyMessage: "Historical card", MessageType: "product", Metadata: structuredProductMetadataForMediaTest(t, reference), IsActive: true, Locale: "en"}
			require.NoError(t, db.Create(&rule).Error)
			input := structuredAutoReplyInputForMediaTest(rule.Metadata)
			input.ReplyMessage = "Edited caption"
			retained, err := tickets.UpdateAutoReplyRule(rule.ID, input)
			require.NoError(t, err)
			require.Equal(t, rule.Metadata, retained.Metadata)
			storedRule, err := tickets.GetAutoReplyRule(rule.ID)
			require.NoError(t, err)
			require.Equal(t, *storedRule, *retained)
			_, _, message, err := tickets.deliverAutoReply(&conversation, *retained, "en", "welcome")
			require.ErrorIs(t, err, ErrInvalidAutoReplyRule)
			require.Nil(t, message)
			moved := input
			moved.MessageType = "image"
			moved.Metadata = fmt.Sprintf(`{"url":%q}`, reference)
			_, err = tickets.UpdateAutoReplyRule(rule.ID, moved)
			require.ErrorIs(t, err, ErrInvalidAutoReplyRule)
			_, err = tickets.UpdateAutoReplyRule(rule.ID, structuredAutoReplyInputForMediaTest(`{"url":"/products/no-image"}`))
			require.NoError(t, err)
			_, err = tickets.UpdateAutoReplyRule(rule.ID, input)
			require.ErrorIs(t, err, ErrInvalidAutoReplyRule)
		})
	}
}

func TestStructuredMetadataDedicatedHistoryRetainsOriginalCleanupEventWithoutGrantingAnonymousAccess(t *testing.T) {
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		for _, source := range []string{"ticket_messages", "ticket_auto_replies"} {
			t.Run(namespace+"/"+source, func(t *testing.T) {
				db, provider, handler, event, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, namespace)
				require.NoError(t, db.Exec("CREATE TABLE "+source+" (id INTEGER PRIMARY KEY, metadata TEXT, message_type TEXT)").Error)
				metadata := structuredProductMetadataForMediaTest(t, provider.GetURL(key)+"?revision=old")
				require.NoError(t, db.Table(source).Create(map[string]any{"id": 1, "metadata": metadata, "message_type": "product"}).Error)
				access := NewPublicUploadAccessService(nil, nil)
				access.ConfigureSiteLogoService(handler.siteLogo)
				access.ConfigureSiteFaviconService(handler.siteFavicon)
				access.ConfigureHomeVisualTileService(handler.homeVisual)
				allowed, err := access.CanServePublicUpload(t.Context(), key)
				require.NoError(t, err)
				require.False(t, allowed)
				require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
				requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, provider, key)
				require.NoError(t, db.Table(source).Where("id = ?", 1).Update("metadata", "{}").Error)
				require.NoError(t, handler.Handle(t.Context(), event))
				require.NoError(t, handler.Handle(t.Context(), event))
				_, err = provider.(storage.ObjectOpener).Open(t.Context(), key)
				require.Error(t, err)
				var count int64
				require.NoError(t, db.Model(&outbox.Event{}).Count(&count).Error)
				require.EqualValues(t, 1, count)
			})
		}
	}
}

func TestStructuredMetadataAndAttachmentReferencesMergeByBusinessRowAndScanBeyondFirstPage(t *testing.T) {
	db, _, assets, conversation, root := createStructuredTicketMediaBindingTestServices(t)
	asset := createStructuredMessageOperationalImage(t, db, root, "content/combined.png")
	attachments, err := json.Marshal([]string{asset.URL})
	require.NoError(t, err)
	message := ticket.TicketMessage{TicketID: conversation.ID, Content: "Combined", MessageType: "product", Metadata: structuredProductMetadataForMediaTest(t, asset.URL), Attachments: string(attachments)}
	require.NoError(t, db.Create(&message).Error)
	for index := 0; index < 256; index++ {
		require.NoError(t, db.Create(&ticket.AutoReplyRule{Type: "welcome", MessageType: "text", Metadata: "not parsed as image JSON"}).Error)
	}
	rule := ticket.AutoReplyRule{Type: "welcome", MessageType: "product", Metadata: message.Metadata}
	require.NoError(t, db.Create(&rule).Error)
	require.NoError(t, db.Model(&rule).Update("is_active", false).Error)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 2, report.Total)
	for _, reference := range report.References {
		if reference.ResourceType == "ticket_message" {
			require.Contains(t, reference.Field, "attachments")
			require.Contains(t, reference.Field, "metadata.thumbnail")
		} else {
			require.Equal(t, "disabled", reference.UsageState)
		}
	}
}

func TestStructuredMetadataMinimalHistoricalSchemaPreservesKnownFieldsWithoutGuessingNavigationURL(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/minimal.png")
	require.NoError(t, db.Exec("CREATE TABLE ticket_messages (id INTEGER PRIMARY KEY, metadata TEXT)").Error)
	require.NoError(t, db.Table("ticket_messages").Create(map[string]any{"id": 1, "metadata": fmt.Sprintf(`{"url":%q}`, asset.URL)}).Error)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Zero(t, report.Total)
	require.NoError(t, db.Table("ticket_messages").Where("id = ?", 1).Update("metadata", structuredProductMetadataForMediaTest(t, asset.URL)).Error)
	report, err = assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
}

func TestStructuredMetadataPrivateFieldsKeepIndependentOriginalDerivativeIndicesAndLargeIDs(t *testing.T) {
	env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
	metadata := fmt.Sprintf(`{"url":%q,"thumbnail":%q,"product_id":9007199254740997}`, env.asset.URL, env.asset.Derivatives[0].URL+"?revision=1#card")
	require.NoError(t, env.db.Model(&env.message).Update("metadata", metadata).Error)
	env.message.Metadata = metadata
	attachments, descriptions := env.access.DescribePublicCustomerServiceMessageAttachments(env.message, *env.conversation.ConversationID)
	require.Len(t, attachments, 1, "card media must not become extra attachments")
	projected := env.access.MetadataWithAuthorizedProtectedMediaURLsForPublicMessage(env.message, descriptions, *env.conversation.ConversationID)
	require.Contains(t, projected, "9007199254740997")
	require.NotContains(t, projected, env.asset.URL)
	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(projected), &fields))
	require.Contains(t, string(fields["url"]), "/attachments/1/file")
	require.Contains(t, string(fields["thumbnail"]), "/attachments/2/file")
	for index, path := range []string{env.asset.StorageKey, env.asset.StorageKey, env.asset.Derivatives[0].StorageKey} {
		file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, index, "")
		require.NoError(t, err)
		actual, err := io.ReadAll(file.ReadCloser)
		require.NoError(t, err)
		require.NoError(t, file.ReadCloser.Close())
		expected, err := os.ReadFile(filepath.Join(env.privateRoot, filepath.FromSlash(path)))
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}
	// A damaged card must not break an independently valid original attachment.
	require.NoError(t, env.db.Model(&env.message).Update("metadata", `{"url":`).Error)
	file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 0, "")
	require.NoError(t, err)
	require.NoError(t, file.ReadCloser.Close())
	_, err = env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 1, "")
	require.ErrorIs(t, err, ErrCustomerServiceMessageAttachmentUnavailable)
}

func TestStructuredMetadataReadOnlyInventoryReportsTrueMediaFieldsAndMalformedPayloadWithoutExposingContent(t *testing.T) {
	db, inventory := createHistoricalMediaReadOnlyInventoryRegressionEnvironment(t)
	require.NoError(t, db.Exec("CREATE TABLE ticket_messages (id INTEGER PRIMARY KEY, metadata TEXT, message_type TEXT)").Error)
	require.NoError(t, db.Exec("CREATE TABLE ticket_auto_replies (id INTEGER PRIMARY KEY, metadata TEXT, message_type TEXT)").Error)
	for _, source := range []string{"ticket_messages", "ticket_auto_replies"} {
		require.NoError(t, db.Table(source).Create([]map[string]any{
			{"id": 1, "message_type": "product", "metadata": `{"thumbnail":"/uploads/content/sensitive-image.png?token=sensitive-secret","url":"/uploads/navigation-only.png","caption":"sensitive-customer-text"}`},
			{"id": 2, "message_type": "product", "metadata": `{"thumbnail":`},
		}).Error)
	}
	require.NoError(t, db.Exec("PRAGMA query_only = ON").Error)
	defer db.Exec("PRAGMA query_only = OFF")
	report, err := inventory.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), defaultHistoricalMediaReadOnlyInventoryRegressionOptions())
	require.NoError(t, err)
	require.Equal(t, 2, report.IssueCounts["business_managed_url_not_registered"])
	require.Equal(t, 2, report.IssueCounts["business_structured_media_could_not_be_parsed"])
	encoded, err := json.Marshal(report)
	require.NoError(t, err)
	for _, secret := range []string{"sensitive-image", "sensitive-secret", "sensitive-customer", "navigation-only"} {
		require.NotContains(t, string(encoded), secret)
	}
}

func TestStructuredMetadataHistoryRetainsOriginalCleanupEventAndMalformedFieldsFailClosed(t *testing.T) {
	db, _, assets, conversation, root := createStructuredTicketMediaBindingTestServices(t)
	asset := createStructuredMessageOperationalImage(t, db, root, "content/replay.png")
	provider := &mediaCleanupStorageWithRetainedOriginalForDelayedConflictTest{StorageService: assets.storage, deleteError: errors.New("synthetic storage outage")}
	assets.storage = provider
	require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	var event outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&event).Error)
	message := ticket.TicketMessage{TicketID: conversation.ID, Content: "Historical metadata", MessageType: "product", Metadata: structuredProductMetadataForMediaTest(t, asset.URL)}
	require.NoError(t, db.Create(&message).Error)
	require.NoError(t, db.Delete(&conversation).Error)
	provider.deleteError = nil
	handler := NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil)
	require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
	require.NoError(t, db.Model(&message).Update("metadata", `{"thumbnail":`).Error)
	err := handler.Handle(t.Context(), event)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ticket_messages.metadata")
	require.NotContains(t, err.Error(), asset.URL)
	_, err = os.Stat(filepath.Join(root, filepath.FromSlash(asset.StorageKey)))
	require.NoError(t, err)
	require.NoError(t, db.Model(&message).Update("metadata", `{"url":"/products/no-media"}`).Error)
	require.NoError(t, handler.Handle(t.Context(), event))
	require.NoError(t, handler.Handle(t.Context(), event))
	_, err = os.Stat(filepath.Join(root, filepath.FromSlash(asset.StorageKey)))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestStructuredMetadataOnlyPrivateOriginalAndDerivativeRemainAvailableOnlyToConversationOwner(t *testing.T) {
	for _, useDerivative := range []bool{false, true} {
		t.Run(fmt.Sprint(useDerivative), func(t *testing.T) {
			env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
			reference := env.asset.URL
			if useDerivative {
				reference = env.asset.Derivatives[0].URL
			}
			require.NoError(t, env.db.Model(&env.message).Updates(map[string]any{"attachments": "[]", "message_type": "product", "metadata": structuredProductMetadataForMediaTest(t, reference)}).Error)
			require.NoError(t, env.db.First(&env.message, env.message.ID).Error)
			metadata := env.access.MetadataWithAuthorizedProtectedMediaURLsForPublicMessage(env.message, nil, *env.conversation.ConversationID)
			require.NotContains(t, metadata, reference)
			require.Contains(t, metadata, fmt.Sprintf("/messages/%d/attachments/0/file", env.message.ID))
			file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 0, "")
			require.NoError(t, err)
			payload, err := io.ReadAll(file.ReadCloser)
			require.NoError(t, err)
			require.NoError(t, file.ReadCloser.Close())
			require.NotEmpty(t, payload)
			_, err = env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, CustomerServiceOwner{VisitorSessionHash: strings.Repeat("d", 64)}, env.message.ID, 0, "")
			require.ErrorIs(t, err, ErrCustomerServiceConversationAccessDenied)
		})
	}
}

func TestStructuredMetadataOnlyHistoricalUnclassifiedMediaWithholdsAnonymousAccessAndKeepsAuthorizedRead(t *testing.T) {
	for _, derivative := range []bool{false, true} {
		t.Run(fmt.Sprint(derivative), func(t *testing.T) {
			env := createHistoricalUnclassifiedPublicTicketAttachmentAccessTestEnvironment(t)
			reference := env.asset.URL
			if derivative {
				reference = env.asset.Derivatives[0].URL
			}
			env.message.Attachments = "[]"
			env.message.MessageType = "product"
			env.message.Metadata = structuredProductMetadataForMediaTest(t, reference+"?revision=old")
			require.NoError(t, env.db.Save(&env.message).Error)
			for _, key := range []string{env.asset.StorageKey, env.asset.Derivatives[0].StorageKey} {
				access, err := env.mediaService.PublicUploadAssetAccess(key)
				require.NoError(t, err)
				require.False(t, access.Allowed)
			}
			projected := env.access.MetadataWithAuthorizedProtectedMediaURLsForPublicMessage(env.message, nil, *env.conversation.ConversationID)
			require.NotContains(t, projected, "localhost:9200")
			require.Contains(t, projected, "/attachments/0/file")
			file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 0, "")
			require.NoError(t, err)
			require.NoError(t, file.ReadCloser.Close())
			var unchanged media.MediaAsset
			require.NoError(t, env.db.First(&unchanged, env.asset.ID).Error)
			require.Equal(t, env.asset.Visibility, unchanged.Visibility)
			require.Equal(t, env.asset.ClassificationRevision, unchanged.ClassificationRevision)
		})
	}
}

func TestStructuredMetadataAuthorizedFileLookupRejectsForeignIdentityAndDuplicateRegistrationBeforeStorageRead(t *testing.T) {
	for _, scenario := range []string{"foreign-host", "duplicate-owner", "retired-asset", "navigation-only"} {
		t.Run(scenario, func(t *testing.T) {
			env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
			reference := env.asset.URL
			switch scenario {
			case "foreign-host":
				reference = "https://foreign.example.test/uploads/" + env.asset.StorageKey
			case "duplicate-owner":
				duplicate := env.asset
				duplicate.ID = 0
				duplicate.Derivatives = nil
				duplicate.URL = "https://foreign.example.test/duplicate.png"
				require.NoError(t, env.db.Create(&duplicate).Error)
			case "retired-asset":
				require.NoError(t, env.db.Delete(&env.asset).Error)
			}
			metadata := structuredProductMetadataForMediaTest(t, reference)
			if scenario == "navigation-only" {
				metadata = fmt.Sprintf(`{"url":%q}`, reference)
			}
			require.NoError(t, env.db.Model(&env.message).Updates(map[string]any{"attachments": "[]", "message_type": "product", "metadata": metadata}).Error)
			_, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 0, "")
			require.Error(t, err)
			require.Zero(t, env.storage.openCalls)
		})
	}
}

func TestStructuredMessageLegacyPublicConversationUploadRemainsSendableOnlyByItsVerifiedUploaderAsAnAttachment(t *testing.T) {
	for _, scenario := range []string{"owner-image-with-metadata-copy", "different-uploader", "metadata-only-card", "product-card-with-attachment", "auto-reply"} {
		t.Run(scenario, func(t *testing.T) {
			db, tickets, _, conversation, root := createStructuredTicketMediaBindingTestServices(t)
			asset := createStructuredMessageOperationalImage(t, db, root, "content/legacy-customer-upload.png")
			hash := conversation.VisitorSessionHash
			if scenario == "different-uploader" {
				hash = strings.Repeat("e", 64)
			}
			require.NoError(t, db.Model(asset).Updates(map[string]any{"upload_origin": media.MediaUploadOriginCustomerServiceConversation, "uploader_actor_type": media.MediaUploadActorVisitor, "uploader_visitor_session_hash": hash, "library_partition": media.MediaLibraryPartitionCustomerSubmissions}).Error)
			alias := "/uploads/" + asset.StorageKey
			metadata := fmt.Sprintf(`{"url":%q,"product_id":9007199254740997}`, alias)
			messageType, attachments := "image", fmt.Sprintf(`[%q]`, alias)
			if scenario == "metadata-only-card" || scenario == "product-card-with-attachment" || scenario == "auto-reply" {
				messageType = "product"
				metadata = structuredProductMetadataForMediaTest(t, alias)
			}
			if scenario == "metadata-only-card" {
				attachments = ""
			}
			if scenario == "auto-reply" {
				_, err := tickets.CreateAutoReplyRule(structuredAutoReplyInputForMediaTest(metadata))
				require.ErrorIs(t, err, ErrInvalidAutoReplyRule)
				return
			}
			_, message, err := tickets.AddPublicCustomerServiceMessage(*conversation.ConversationID, CustomerServiceOwner{VisitorSessionHash: conversation.VisitorSessionHash}, "Legacy photo", 0, messageType, metadata, attachments)
			if scenario == "owner-image-with-metadata-copy" {
				require.NoError(t, err)
				require.NotNil(t, message)
				require.Contains(t, message.Metadata, "9007199254740997")
				require.Contains(t, message.Metadata, asset.URL)
				require.Contains(t, message.Attachments, asset.URL)
			} else {
				require.ErrorIs(t, err, ErrCustomerServiceMessageMediaReferenceInvalid)
			}
		})
	}
}
