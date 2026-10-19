package ticket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	ticketdomain "commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/visitorcookie"
)

func TestStructuredCustomerServiceInvalidManagedMediaReturnsHTTP400WithoutMessageOrOutboxWrites(t *testing.T) {
	db, handler, assets, router, _ := createCustomerServiceAttachmentIdentityTestEnvironment(t, 0)
	handler.ticketService.ConfigureMediaService(assets)
	visitorID := uuid.NewString()
	conversation := seedCustomerServiceAttachmentIdentityConversation(t, db, nil, visitorcookie.CustomerServiceVisitorHash(visitorID))
	for _, reference := range []string{"/uploads/site-logo/synthetic-private-detail.png", "/uploads/content/synthetic-unregistered.png"} {
		metadata, err := json.Marshal(map[string]any{"url": "/products/example", "thumbnail": reference})
		require.NoError(t, err)
		body, err := json.Marshal(map[string]any{"conversation_id": *conversation.ConversationID, "message": "Invalid card", "sender_name": "Synthetic visitor", "message_type": "product", "metadata": json.RawMessage(metadata)})
		require.NoError(t, err)
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/messages", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(&http.Cookie{Name: customerServiceVisitorCookie, Value: handler.signVisitorSessionID(visitorID)})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
		require.Contains(t, response.Body.String(), "message media is unavailable")
		require.NotContains(t, response.Body.String(), reference)
	}
	var messages, events int64
	require.NoError(t, db.Model(&ticketdomain.TicketMessage{}).Count(&messages).Error)
	require.NoError(t, db.Model(&outbox.Event{}).Count(&events).Error)
	require.Zero(t, messages)
	require.Zero(t, events)
}

func TestStructuredCustomerServiceHTTPMediaBindingCanonicalizesAliasAndPreservesLargeJSONIDs(t *testing.T) {
	db, handler, assets, router, _ := createCustomerServiceAttachmentIdentityTestEnvironment(t, 0)
	handler.ticketService.ConfigureMediaService(assets)
	visitorID := uuid.NewString()
	conversation := seedCustomerServiceAttachmentIdentityConversation(t, db, nil, visitorcookie.CustomerServiceVisitorHash(visitorID))
	asset := media.MediaAsset{Filename: "card.png", URL: "https://shop.example.test/uploads/content/card.png", StorageKey: "content/card.png", MediaType: "image", MimeType: "image/png", Status: "active", Visibility: "public", UploadOrigin: media.MediaUploadOriginAdministratorShared, ClassificationMethod: media.MediaClassificationRecordedUploadContext, LibraryPartition: media.MediaLibraryPartitionShared}
	require.NoError(t, db.Create(&asset).Error)
	metadata := json.RawMessage(`{"thumbnail":"/uploads/content/card.png","url":"/products/card","product_id":9007199254740997}`)
	body, err := json.Marshal(map[string]any{"conversation_id": *conversation.ConversationID, "message": "Valid card", "sender_name": "Synthetic visitor", "message_type": "product", "metadata": metadata})
	require.NoError(t, err)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/messages", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: customerServiceVisitorCookie, Value: handler.signVisitorSessionID(visitorID)})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), "9007199254740997")
	var message ticketdomain.TicketMessage
	require.NoError(t, db.First(&message, "ticket_id = ?", conversation.ID).Error)
	require.Contains(t, message.Metadata, asset.URL)
	require.Contains(t, message.Metadata, "9007199254740997")
}

func TestStructuredCustomerServiceMetadataOnlyPrivateMediaHistoryAndFileRoutesRecheckOwnerWithoutExtraAttachments(t *testing.T) {
	for _, member := range []bool{false, true} {
		t.Run(fmt.Sprint(member), func(t *testing.T) {
			env := createCustomerServicePrivateAttachmentHTTPTestEnvironment(t, member)
			metadata := fmt.Sprintf(`{"thumbnail":%q,"product_id":9007199254740997,"url":"/products/example"}`, env.asset.URL)
			require.NoError(t, env.db.Model(&env.message).Updates(map[string]any{"message_type": "product", "attachments": "[]", "metadata": metadata}).Error)
			historyRequest := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/customer-service/messages/"+*env.conversation.ConversationID, nil)
			historyRequest.AddCookie(env.cookie)
			if member {
				historyRequest.Header.Set("X-Test-Member-Identity", "owner")
			}
			history := httptest.NewRecorder()
			env.router.ServeHTTP(history, historyRequest)
			require.Equal(t, http.StatusOK, history.Code, history.Body.String())
			require.NotContains(t, history.Body.String(), env.asset.URL)
			require.Contains(t, history.Body.String(), "9007199254740997")
			var payload struct {
				Data []struct {
					Metadata    json.RawMessage
					Attachments []string
				}
			}
			require.NoError(t, json.Unmarshal(history.Body.Bytes(), &payload))
			require.Len(t, payload.Data, 1)
			require.Empty(t, payload.Data[0].Attachments)
			var card struct{ Thumbnail string }
			require.NoError(t, json.Unmarshal(payload.Data[0].Metadata, &card))
			parsed, err := url.Parse(card.Thumbnail)
			require.NoError(t, err)
			require.Contains(t, parsed.Path, fmt.Sprintf("/messages/%d/attachments/0/file", env.message.ID))
			for _, authorized := range []bool{true, false} {
				request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, parsed.RequestURI(), nil)
				if authorized {
					request.AddCookie(env.cookie)
					if member {
						request.Header.Set("X-Test-Member-Identity", "owner")
					}
				}
				response := httptest.NewRecorder()
				env.router.ServeHTTP(response, request)
				require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
				if authorized {
					require.Equal(t, http.StatusOK, response.Code)
					require.Equal(t, env.payload, response.Body.Bytes())
				} else {
					require.Equal(t, http.StatusNotFound, response.Code)
				}
			}
			// Previously issued links must stop working after this exact message field detaches.
			require.NoError(t, env.db.Model(&env.message).Update("metadata", "{}").Error)
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, parsed.RequestURI(), strings.NewReader(""))
			request.AddCookie(env.cookie)
			if member {
				request.Header.Set("X-Test-Member-Identity", "owner")
			}
			response := httptest.NewRecorder()
			env.router.ServeHTTP(response, request)
			require.Equal(t, http.StatusNotFound, response.Code)
		})
	}
}
