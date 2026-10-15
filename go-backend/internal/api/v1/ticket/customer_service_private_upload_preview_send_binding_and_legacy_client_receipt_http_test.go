package ticket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	ticketdomain "commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/visitorcookie"
	"commerce-platform/internal/repository"
)

func TestPublicCustomerServicePrivateUploadPreviewAndLegacyURLSendProduceAuthorizedAssetBindings(t *testing.T) {
	db, handler, mediaService, router, _ := createCustomerServiceAttachmentIdentityTestEnvironment(t, 0)
	require.NoError(t, db.AutoMigrate(&media.MediaUploadSession{}, &ticketdomain.TicketMessageMediaAttachment{}))
	mediaService.ConfigureCustomerServicePrivateAttachmentUploads(true)
	router.GET("/api/v1/customer-service/conversations/:conversation_id/attachment-uploads/:upload_session_id/file", handler.ServePublicCustomerServicePendingAttachmentUploadPreview)
	router.GET("/api/v1/customer-service/conversations/:conversation_id/messages/:message_id/attachments/:attachment_index/file", handler.ServePublicCustomerServiceMessageAttachmentFile)
	visitorID := uuid.NewString()
	ownerHash := visitorcookie.CustomerServiceVisitorHash(visitorID)
	conversation := seedCustomerServiceAttachmentIdentityConversation(t, db, nil, ownerHash)
	cookie := &http.Cookie{Name: customerServiceVisitorCookie, Value: handler.signVisitorSessionID(visitorID)}
	uploadResponse := httptest.NewRecorder()
	router.ServeHTTP(uploadResponse, createCustomerServiceAttachmentMultipartRequest(t, *conversation.ConversationID, cookie))
	require.Equal(t, http.StatusCreated, uploadResponse.Code, uploadResponse.Body.String())
	require.Equal(t, "private, no-store", uploadResponse.Header().Get("Cache-Control"))
	var uploaded struct {
		Asset struct {
			ID              uint
			URL             string
			UploadSessionID string `json:"upload_session_id"`
		}
	}
	require.NoError(t, json.Unmarshal(uploadResponse.Body.Bytes(), &uploaded))
	require.NotEmpty(t, uploaded.Asset.UploadSessionID)
	require.Contains(t, uploaded.Asset.URL, "/attachment-uploads/"+uploaded.Asset.UploadSessionID+"/file")
	require.NotContains(t, uploadResponse.Body.String(), "customer-service/attachments/originals/")
	require.NotContains(t, uploadResponse.Body.String(), ownerHash)
	previewRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, uploaded.Asset.URL+"?preset=thumbnail", nil)
	previewRequest.AddCookie(cookie)
	previewResponse := httptest.NewRecorder()
	router.ServeHTTP(previewResponse, previewRequest)
	require.Equal(t, http.StatusOK, previewResponse.Code, previewResponse.Body.String())
	require.NotEmpty(t, previewResponse.Body.Bytes())
	require.Empty(t, previewResponse.Header().Get("Location"))

	for _, badCookie := range []*http.Cookie{nil, {Name: customerServiceVisitorCookie, Value: handler.signVisitorSessionID(uuid.NewString())}} {
		request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, uploaded.Asset.URL, nil)
		if badCookie != nil {
			request.AddCookie(badCookie)
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
	}
	payload, err := json.Marshal(map[string]interface{}{"conversation_id": *conversation.ConversationID, "message": "Photo", "sender_name": "Visitor",
		"message_type": "image", "attachment_url": uploaded.Asset.URL, "attachments": []string{uploaded.Asset.URL}})
	require.NoError(t, err)
	sendRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/messages", bytes.NewReader(payload))
	sendRequest.Header.Set("Content-Type", "application/json")
	sendRequest.AddCookie(cookie)
	sendResponse := httptest.NewRecorder()
	router.ServeHTTP(sendResponse, sendRequest)
	require.Equal(t, http.StatusOK, sendResponse.Code, sendResponse.Body.String())
	require.NotContains(t, sendResponse.Body.String(), "/attachment-uploads/")
	var message ticketdomain.TicketMessage
	require.NoError(t, db.Where("ticket_id = ? AND content = ?", conversation.ID, "Photo").First(&message).Error)
	require.Equal(t, ticketdomain.TicketMessageMediaAssetBindings, message.AttachmentReferenceFormat)
	require.Empty(t, message.Attachments)
	var binding ticketdomain.TicketMessageMediaAttachment
	require.NoError(t, db.First(&binding, "ticket_message_id = ?", message.ID).Error)
	require.Equal(t, uploaded.Asset.ID, binding.MediaAssetID)
	var session media.MediaUploadSession
	require.NoError(t, db.First(&session, "id = ?", uploaded.Asset.UploadSessionID).Error)
	require.Equal(t, "bound", session.Status)
	attachmentRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("/api/v1/customer-service/conversations/%s/messages/%d/attachments/0/file", *conversation.ConversationID, message.ID), nil)
	attachmentRequest.AddCookie(cookie)
	attachmentResponse := httptest.NewRecorder()
	router.ServeHTTP(attachmentResponse, attachmentRequest)
	require.Equal(t, http.StatusOK, attachmentResponse.Code)
	require.NotEmpty(t, attachmentResponse.Body.Bytes())
	var beforeMessages, beforeEvents int64
	require.NoError(t, db.Model(&ticketdomain.TicketMessage{}).Count(&beforeMessages).Error)
	require.NoError(t, db.Model(&outbox.Event{}).Count(&beforeEvents).Error)
	retryRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/messages", bytes.NewReader(payload))
	retryRequest.Header.Set("Content-Type", "application/json")
	retryRequest.AddCookie(cookie)
	retryResponse := httptest.NewRecorder()
	router.ServeHTTP(retryResponse, retryRequest)
	require.Equal(t, http.StatusBadRequest, retryResponse.Code, retryResponse.Body.String())
	var afterMessages, afterEvents int64
	require.NoError(t, db.Model(&ticketdomain.TicketMessage{}).Count(&afterMessages).Error)
	require.NoError(t, db.Model(&outbox.Event{}).Count(&afterEvents).Error)
	require.Equal(t, beforeMessages, afterMessages)
	require.Equal(t, beforeEvents, afterEvents)
	asset, err := repository.NewMediaRepository(db).FindAssetByID(uploaded.Asset.ID)
	require.NoError(t, err)
	require.Equal(t, "private", asset.Visibility)
	require.True(t, strings.HasPrefix(asset.StorageKey, "customer-service/attachments/originals/"))
}
