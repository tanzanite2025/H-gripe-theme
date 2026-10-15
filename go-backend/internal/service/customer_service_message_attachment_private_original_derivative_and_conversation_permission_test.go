package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

type customerServicePrivateAttachmentAccessTestEnvironment struct {
	db              *gorm.DB
	access          *CustomerServiceConversationMessageAttachmentAccessService
	mediaService    *MediaService
	storage         *customerServiceAttachmentAuthorizedObjectReadRecordingStorage
	conversation    ticket.Ticket
	message         ticket.TicketMessage
	asset           media.MediaAsset
	owner           CustomerServiceOwner
	publicRoot      string
	privateRoot     string
	originalPayload []byte
}

func createCustomerServicePrivateAttachmentAccessTestEnvironment(t *testing.T) customerServicePrivateAttachmentAccessTestEnvironment {
	t.Helper()
	db, tickets := newTestTicketBoundaryService(t)
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}))
	publicRoot, privateRoot := t.TempDir(), t.TempDir()
	store, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal,
		LocalPath: publicRoot, PrivateLocalPath: privateRoot, BaseURL: "https://shop.example.test", PrivateBaseURL: "https://private.example.test"})
	require.NoError(t, err)
	mediaService := NewMediaService(repository.NewMediaRepository(db), store, nil, "https://shop.example.test", 20<<30)
	payload := testOpaquePNG(t, 64, 32)
	file := multipartFileHeader(t, "customer-attachment.png", "image/png", payload)
	originalURL, err := store.(storage.PrivateObjectUploader).UploadWithPrefixPrivate(t.Context(), file, storage.CustomerServiceConversationAttachmentOriginalsPrefix)
	require.NoError(t, err)
	originalKey, err := store.ObjectKey(originalURL)
	require.NoError(t, err)
	owner := CustomerServiceOwner{VisitorSessionHash: strings.Repeat("a", 64)}
	asset := media.MediaAsset{Filename: filepath.Base(originalKey), OriginalFilename: file.Filename, URL: originalURL, StorageKey: originalKey,
		MediaType: "image", MimeType: "image/png", Size: int64(len(payload)), Width: 64, Height: 32, Status: "active", Visibility: "private",
		UploaderActorType: media.MediaUploadActorVisitor, UploaderVisitorSessionHash: owner.VisitorSessionHash}
	require.NoError(t, db.Create(&asset).Error)
	derivatives, err := mediaService.generateAssetDerivatives(t.Context(), &asset, file)
	require.NoError(t, err)
	require.Len(t, derivatives, 3)
	require.NoError(t, mediaService.repo.CreateAssetDerivatives(derivatives))
	asset.Derivatives = derivatives
	support := createTicketBoundaryUser(t, db, "attachment-support@example.test", "attachment-support", "support")
	conversationID := uuid.NewString()
	conversation := ticket.Ticket{TicketNumber: "attachment-" + uuid.NewString(), UserID: support.ID, AssignedTo: support.ID,
		ConversationID: &conversationID, VisitorSessionHash: owner.VisitorSessionHash, Category: "customer_service", Subject: "Attachment access", Status: "in_progress"}
	require.NoError(t, db.Create(&conversation).Error)
	attachmentJSON, err := json.Marshal([]string{originalURL})
	require.NoError(t, err)
	message := ticket.TicketMessage{TicketID: conversation.ID, Content: "Customer photo", MessageType: "image", Attachments: string(attachmentJSON)}
	require.NoError(t, db.Create(&message).Error)
	recordingStore := &customerServiceAttachmentAuthorizedObjectReadRecordingStorage{StorageService: store}
	mediaService.storage = recordingStore
	return customerServicePrivateAttachmentAccessTestEnvironment{db: db, access: NewCustomerServiceConversationMessageAttachmentAccessService(tickets, mediaService),
		mediaService: mediaService, storage: recordingStore, conversation: conversation, message: message, asset: asset, owner: owner,
		publicRoot: publicRoot, privateRoot: privateRoot, originalPayload: payload}
}

func TestCustomerServiceAttachmentAuthorizedReadStreamsPrivateOriginalAndDerivativesFromDedicatedRoot(t *testing.T) {
	env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
	file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 0, "")
	require.NoError(t, err)
	require.Empty(t, file.RedirectURL)
	payload, err := io.ReadAll(file.ReadCloser)
	require.NoError(t, err)
	require.NoError(t, file.ReadCloser.Close())
	require.Equal(t, env.originalPayload, payload)
	for _, derivative := range env.asset.Derivatives {
		require.True(t, storage.IsPrivateObjectKey(derivative.StorageKey))
		require.True(t, storage.IsCustomerServiceConversationAttachmentDerivativeObjectKey(derivative.StorageKey))
		_, err := os.Stat(filepath.Join(env.publicRoot, filepath.FromSlash(derivative.StorageKey)))
		require.True(t, os.IsNotExist(err))
		storedPayload, err := os.ReadFile(filepath.Join(env.privateRoot, filepath.FromSlash(derivative.StorageKey)))
		require.NoError(t, err)
		file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 0, derivative.Preset)
		require.NoError(t, err)
		payload, err := io.ReadAll(file.ReadCloser)
		require.NoError(t, err)
		require.NoError(t, file.ReadCloser.Close())
		require.Equal(t, storedPayload, payload)
		allowed, err := env.mediaService.CanServePublicUpload(derivative.StorageKey)
		require.NoError(t, err)
		require.False(t, allowed)
	}
	_, err = os.Stat(filepath.Join(env.publicRoot, filepath.FromSlash(env.asset.StorageKey)))
	require.True(t, os.IsNotExist(err))
	_, err = env.mediaService.OpenAssetFile(t.Context(), env.asset.ID)
	require.ErrorIs(t, err, ErrMediaAssetForbidden)
	_, _, err = env.mediaService.ExportCopyrightEvidence(t.Context(), env.asset.ID)
	require.ErrorIs(t, err, ErrMediaAssetForbidden)
}

func TestCustomerServiceAttachmentAuthorizedReadRejectsForeignOwnersMessagesAndUnsafeReferencesBeforeObjectOpen(t *testing.T) {
	for _, scenario := range []string{"missing owner", "other visitor", "member accessing guest", "foreign message", "internal message", "deleted conversation", "inactive asset", "deleted asset", "negative attachment index", "missing attachment index", "missing preset", "foreign derivative key", "foreign upload origin", "other private domain", "path traversal"} {
		t.Run(scenario, func(t *testing.T) {
			env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
			owner, messageID, index, preset := env.owner, env.message.ID, 0, ""
			switch scenario {
			case "missing owner":
				owner = CustomerServiceOwner{}
			case "other visitor":
				owner.VisitorSessionHash = strings.Repeat("b", 64)
			case "member accessing guest":
				userID := uint(999)
				owner.UserID = &userID
			case "foreign message":
				require.NoError(t, env.db.Model(&env.message).Update("ticket_id", env.conversation.ID+100).Error)
			case "internal message":
				require.NoError(t, env.db.Model(&env.message).Update("is_internal", true).Error)
			case "deleted conversation":
				require.NoError(t, env.db.Delete(&env.conversation).Error)
			case "inactive asset":
				require.NoError(t, env.db.Model(&env.asset).Update("status", "inactive").Error)
			case "deleted asset":
				require.NoError(t, env.db.Delete(&env.asset).Error)
			case "negative attachment index":
				index = -1
			case "missing attachment index":
				index = 1
			case "missing preset":
				preset = "not-generated"
			case "foreign derivative key":
				preset = env.asset.Derivatives[0].Preset
				require.NoError(t, env.db.Model(&env.asset.Derivatives[0]).Update("storage_key", "media-derivatives/foreign.jpg").Error)
			case "foreign upload origin":
				replaceCustomerServiceAttachmentAccessTestMessageReference(t, env, "https://untrusted.example/uploads/"+env.asset.StorageKey)
			case "other private domain":
				key := "warranty/customer-file.png"
				url := env.storage.GetURL(key)
				require.NoError(t, env.db.Model(&env.asset).Updates(map[string]interface{}{"storage_key": key, "url": url}).Error)
				replaceCustomerServiceAttachmentAccessTestMessageReference(t, env, url)
			case "path traversal":
				replaceCustomerServiceAttachmentAccessTestMessageReference(t, env, "/uploads/../"+env.asset.StorageKey)
			}
			file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, owner, messageID, index, preset)
			require.Error(t, err)
			require.Nil(t, file)
			require.Zero(t, env.storage.openCalls)
		})
	}
}

func TestCustomerServiceAttachmentAgentReadTracksAssignmentAndAdministratorScope(t *testing.T) {
	env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
	for _, testCase := range []struct {
		agentID    uint
		canViewAll bool
		allowed    bool
	}{
		{agentID: env.conversation.AssignedTo, allowed: true},
		{agentID: env.conversation.AssignedTo + 100},
		{agentID: env.conversation.AssignedTo + 100, canViewAll: true, allowed: true},
		{canViewAll: true},
	} {
		file, err := env.access.OpenAgentCustomerServiceMessageAttachment(t.Context(), env.conversation.ID, testCase.agentID, testCase.canViewAll, env.message.ID, 0, "thumbnail")
		if testCase.allowed {
			require.NoError(t, err)
			require.NoError(t, file.ReadCloser.Close())
		} else {
			require.Error(t, err)
			require.Nil(t, file)
		}
	}
	previousAgent := env.conversation.AssignedTo
	require.NoError(t, env.db.Model(&env.conversation).Update("assigned_to", previousAgent+100).Error)
	file, err := env.access.OpenAgentCustomerServiceMessageAttachment(t.Context(), env.conversation.ID, previousAgent, false, env.message.ID, 0, "")
	require.ErrorIs(t, err, ErrCustomerServiceAgentAccessDenied)
	require.Nil(t, file)
	file, err = env.access.OpenAgentCustomerServiceMessageAttachment(t.Context(), env.conversation.ID, previousAgent+100, false, env.message.ID, 0, "")
	require.NoError(t, err)
	require.NoError(t, file.ReadCloser.Close())
}

func TestCustomerServiceAttachmentDescriptionsReplacePrivateURLsAndPreserveHistoricalPublicURLs(t *testing.T) {
	env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
	urls, descriptions := env.access.DescribePublicCustomerServiceMessageAttachments(env.message, *env.conversation.ConversationID)
	require.Len(t, descriptions, 1)
	require.Equal(t, env.asset.ID, descriptions[0].AssetID)
	require.Len(t, descriptions[0].Derivatives, 3)
	require.Equal(t, []string{descriptions[0].AccessURL}, urls)
	require.Contains(t, urls[0], fmt.Sprintf("/messages/%d/attachments/0/file", env.message.ID))
	encoded, err := json.Marshal(descriptions)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), env.asset.StorageKey)
	require.NotContains(t, string(encoded), "private.example.test")
	publicURLs, _ := env.access.DescribePublicCustomerServiceMessageAttachments(ticket.TicketMessage{Attachments: `["http://localhost:9199/uploads/2026/08/20/legacy.webp"]`}, "legacy")
	require.Equal(t, []string{"https://shop.example.test/uploads/2026/08/20/legacy.webp"}, publicURLs)
	agentURLs, _ := env.access.DescribeAgentCustomerServiceMessageAttachments(env.message)
	require.Contains(t, agentURLs[0], fmt.Sprintf("/api/admin/customer-service/conversations/%d/", env.conversation.ID))
	legacyAgentURLs, _ := env.access.DescribeAgentCustomerServiceMessageAttachments(ticket.TicketMessage{Attachments: `["http://localhost:9199/uploads/2026/08/20/legacy.webp"]`})
	require.Equal(t, []string{"http://localhost:9199/uploads/2026/08/20/legacy.webp"}, legacyAgentURLs)
}

func replaceCustomerServiceAttachmentAccessTestMessageReference(t *testing.T, env customerServicePrivateAttachmentAccessTestEnvironment, reference string) {
	t.Helper()
	payload, err := json.Marshal([]string{reference})
	require.NoError(t, err)
	require.NoError(t, env.db.Model(&env.message).Update("attachments", string(payload)).Error)
}

func TestCustomerServiceAttachmentDescriptionsAndReadTrackAnUpgradedDerivativePresetVersion(t *testing.T) {
	env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
	payload := testOpaquePNG(t, 80, 40)
	reference, err := env.storage.StorageService.(storage.PrivateCacheControlledReaderUploader).UploadPrivateObjectFromReaderWithPrefixAndCacheControl(
		t.Context(), bytes.NewReader(payload), "thumbnail-v2.png", fmt.Sprintf("%s/%d/thumbnail/v2", storage.CustomerServiceConversationAttachmentDerivativesPrefix, env.asset.ID), "private, no-store")
	require.NoError(t, err)
	key, err := env.storage.ObjectKey(reference)
	require.NoError(t, err)
	latest := media.MediaAssetDerivative{MediaAssetID: env.asset.ID, Preset: "thumbnail", PresetVersion: 2,
		URL: reference, StorageKey: key, MimeType: "image/png", Size: int64(len(payload)), Width: 80, Height: 40}
	// Existing preset rows are replaced in place; the repository's unique
	// asset/preset constraint prevents competing active version records.
	require.NoError(t, env.db.Model(&media.MediaAssetDerivative{}).
		Where("media_asset_id = ? AND preset = ?", env.asset.ID, latest.Preset).
		Updates(map[string]interface{}{"preset_version": latest.PresetVersion, "url": latest.URL, "storage_key": latest.StorageKey,
			"mime_type": latest.MimeType, "size": latest.Size, "width": latest.Width, "height": latest.Height}).Error)
	_, descriptions := env.access.DescribePublicCustomerServiceMessageAttachments(env.message, *env.conversation.ConversationID)
	require.Len(t, descriptions, 1)
	require.Len(t, descriptions[0].Derivatives, 3)
	thumbnailCount := 0
	for _, description := range descriptions[0].Derivatives {
		if description.Preset == "thumbnail" {
			thumbnailCount++
			require.Equal(t, latest.Width, description.Width)
			require.Equal(t, latest.Height, description.Height)
		}
	}
	require.Equal(t, 1, thumbnailCount)
	file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 0, "thumbnail")
	require.NoError(t, err)
	content, err := io.ReadAll(file.ReadCloser)
	require.NoError(t, err)
	require.NoError(t, file.ReadCloser.Close())
	require.Equal(t, payload, content)
}

type customerServiceAttachmentAuthorizedObjectReadRecordingStorage struct {
	storage.StorageService
	openCalls int
}

func (s *customerServiceAttachmentAuthorizedObjectReadRecordingStorage) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	s.openCalls++
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func TestCustomerServiceAttachmentPrivateDerivativeStoragePolicyNeverFallsBackToPublicCache(t *testing.T) {
	env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
	_, _, err := mediaAssetDerivativeStoragePrefixAndCachePolicy(&media.MediaAsset{ID: 1, Visibility: "private", StorageKey: "warranty/file.png"}, "thumbnail", 1)
	require.ErrorIs(t, err, ErrMediaDerivativeGenerationFailed)
	_, err = uploadMediaDerivativeObjectUsingAssetAccessPolicy(t.Context(), env.storage, &env.asset, bytes.NewReader([]byte("derivative")), "file.png", "thumbnail", 1)
	require.ErrorIs(t, err, ErrMediaStorageUnavailable)
	genericReaderStorage := &customerServiceAttachmentGenericReaderUploaderWithoutPrivateAccessPolicy{StorageService: env.storage.StorageService}
	_, err = uploadMediaDerivativeObjectUsingAssetAccessPolicy(t.Context(), genericReaderStorage, &env.asset, bytes.NewReader([]byte("derivative")), "file.png", "thumbnail", 1)
	require.ErrorIs(t, err, ErrMediaStorageUnavailable)
	_, cacheControl, err := mediaAssetDerivativeStoragePrefixAndCachePolicy(&env.asset, "thumbnail", 1)
	require.NoError(t, err)
	require.Equal(t, "private, no-store", cacheControl)
}

type customerServiceAttachmentGenericReaderUploaderWithoutPrivateAccessPolicy struct {
	storage.StorageService
}

func (s *customerServiceAttachmentGenericReaderUploaderWithoutPrivateAccessPolicy) UploadFromReaderWithPrefixAndCacheControl(ctx context.Context, reader io.Reader, filename, prefix, cacheControl string) (string, error) {
	return s.StorageService.(storage.CacheControlledReaderUploader).UploadFromReaderWithPrefixAndCacheControl(ctx, reader, filename, prefix, cacheControl)
}
