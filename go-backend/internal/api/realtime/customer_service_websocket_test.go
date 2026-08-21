package realtime

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/service"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomerServiceWebSocketOriginAllowed(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://admin.example.test/api/admin/customer-service/ws", nil)
	request.Host = "admin.example.test"

	tests := []struct {
		name           string
		origin         string
		forwardedProto string
		allowedOrigins []string
		want           bool
	}{
		{
			name: "rejects missing origin",
			want: false,
		},
		{
			name:   "accepts same http origin",
			origin: "http://admin.example.test",
			want:   true,
		},
		{
			name:           "accepts same forwarded https origin",
			origin:         "https://admin.example.test",
			forwardedProto: "https",
			want:           true,
		},
		{
			name:           "accepts configured origin",
			origin:         "https://admin-ui.example.test",
			forwardedProto: "https",
			allowedOrigins: []string{"https://admin-ui.example.test"},
			want:           true,
		},
		{
			name:           "rejects foreign origin",
			origin:         "https://attacker.example.test",
			forwardedProto: "https",
			allowedOrigins: []string{"https://admin-ui.example.test"},
			want:           false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := request.Clone(request.Context())
			candidate.Header.Set("Origin", test.origin)
			candidate.Header.Set("X-Forwarded-Proto", test.forwardedProto)

			assert.Equal(t, test.want, CustomerServiceWebSocketOriginAllowed(candidate, test.allowedOrigins))
		})
	}
}

func TestCustomerServiceWebSocketDeliversAllowedEventsAndAdvancesFilteredCursor(t *testing.T) {
	hub := service.NewCustomerServiceEventHub()
	server := newCustomerServiceWebSocketTestServer(t, hub, CustomerServiceWebSocketOptions{
		AllowEvent: func(event service.CustomerServiceRealtimeEvent) bool {
			return event.DeliversTo(service.CustomerServiceRealtimeAudiencePublic)
		},
	})
	connection := dialCustomerServiceWebSocket(t, server.URL)
	defer connection.Close()

	ready := readCustomerServiceWebSocketFrame(t, connection)
	assert.Equal(t, "ready", ready["type"])

	allowed := service.NewCustomerServiceRealtimeEventWithIDAndAudience(
		"public-message",
		service.CustomerServiceEventMessageCreated,
		123,
		"public-conversation-id",
		service.CustomerServiceRealtimeActor{Kind: "customer", Anonymous: true},
		service.CustomerServiceRealtimeAudienceBoth,
		map[string]uint{"message_id": 481},
	)
	allowed.StreamID = "1750000000000-0"
	hub.Publish(allowed)

	frame := readCustomerServiceWebSocketFrame(t, connection)
	assert.Equal(t, "event", frame["type"])
	assert.Equal(t, allowed.StreamID, frame["cursor"])
	event, ok := frame["event"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, allowed.EventID, event["event_id"])
	assert.Equal(t, "both", event["audience"])

	filtered := service.NewCustomerServiceRealtimeEventWithIDAndAudience(
		"backoffice-read",
		service.CustomerServiceEventMessagesRead,
		123,
		"public-conversation-id",
		service.CustomerServiceRealtimeActor{Kind: "agent"},
		service.CustomerServiceRealtimeAudienceBackoffice,
		map[string]uint{"message_id": 481},
	)
	filtered.StreamID = "1750000000000-1"
	hub.Publish(filtered)

	frame = readCustomerServiceWebSocketFrame(t, connection)
	assert.Equal(t, "cursor", frame["type"])
	assert.Equal(t, filtered.StreamID, frame["cursor"])
	assert.NotContains(t, frame, "event")
}

func TestCustomerServiceWebSocketControlsAndReplayDeduplication(t *testing.T) {
	hub := service.NewCustomerServiceEventHub()
	typingControls := make(chan CustomerServiceWebSocketControl, 1)
	replayed := service.NewCustomerServiceRealtimeEventWithID(
		"replayed-message",
		service.CustomerServiceEventMessageCreated,
		123,
		"public-conversation-id",
		service.CustomerServiceRealtimeActor{Kind: "customer", Anonymous: true},
		map[string]uint{"message_id": 481},
	)
	replayed.StreamID = "1750000000000-2"
	server := newCustomerServiceWebSocketTestServer(t, hub, CustomerServiceWebSocketOptions{
		Replay: []service.CustomerServiceRealtimeEvent{replayed},
		HandleControl: func(control CustomerServiceWebSocketControl) {
			typingControls <- control
		},
	})
	connection := dialCustomerServiceWebSocket(t, server.URL)
	defer connection.Close()

	assert.Equal(t, "ready", readCustomerServiceWebSocketFrame(t, connection)["type"])
	frame := readCustomerServiceWebSocketFrame(t, connection)
	assert.Equal(t, "event", frame["type"])
	assert.Equal(t, replayed.EventID, frame["event"].(map[string]interface{})["event_id"])

	require.NoError(t, connection.WriteJSON(map[string]interface{}{"type": "ping"}))
	assert.Equal(t, "pong", readCustomerServiceWebSocketFrame(t, connection)["type"])

	require.NoError(t, connection.WriteJSON(map[string]interface{}{"type": "typing", "is_typing": false}))
	select {
	case control := <-typingControls:
		require.NotNil(t, control.IsTyping)
		assert.False(t, *control.IsTyping)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for typing control")
	}

	require.NoError(t, connection.WriteJSON(map[string]interface{}{"type": "typing", "display_name": "not accepted"}))
	frame = readCustomerServiceWebSocketFrame(t, connection)
	assert.Equal(t, "error", frame["type"])
	assert.Equal(t, "invalid_control", frame["code"])

	hub.Publish(replayed)
	assertNoCustomerServiceWebSocketFrame(t, connection)
}

func TestCustomerServiceWebSocketReleasesConnectionSlotAfterDisconnect(t *testing.T) {
	require.Eventually(t, func() bool {
		return customerServiceWebSocketConnections.Load() == 0
	}, 5*time.Second, 10*time.Millisecond)

	hub := service.NewCustomerServiceEventHub()
	server := newCustomerServiceWebSocketTestServer(t, hub, CustomerServiceWebSocketOptions{})
	connection := dialCustomerServiceWebSocket(t, server.URL)
	assert.Equal(t, "ready", readCustomerServiceWebSocketFrame(t, connection)["type"])
	require.NoError(t, connection.Close())

	require.Eventually(t, func() bool {
		return customerServiceWebSocketConnections.Load() == 0
	}, 5*time.Second, 10*time.Millisecond)
}

func newCustomerServiceWebSocketTestServer(t *testing.T, hub *service.CustomerServiceEventHub, options CustomerServiceWebSocketOptions) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subscription := hub.SubscribeConversation(123)
		options.Subscription = subscription
		options.CheckOrigin = func(*http.Request) bool { return true }
		ServeCustomerServiceWebSocket(w, r, options)
	}))
	t.Cleanup(server.Close)
	return server
}

func dialCustomerServiceWebSocket(t *testing.T, serverURL string) *websocket.Conn {
	t.Helper()

	webSocketURL, err := url.Parse(serverURL)
	require.NoError(t, err)
	webSocketURL.Scheme = "ws"
	webSocketURL.Path = "/customer-service/ws"

	connection, response, err := websocket.DefaultDialer.Dial(webSocketURL.String(), http.Header{
		"Origin": []string{serverURL},
	})
	if response != nil {
		defer response.Body.Close()
	}
	require.NoError(t, err)
	return connection
}

func readCustomerServiceWebSocketFrame(t *testing.T, connection *websocket.Conn) map[string]interface{} {
	t.Helper()

	require.NoError(t, connection.SetReadDeadline(time.Now().Add(time.Second)))
	var frame map[string]interface{}
	require.NoError(t, connection.ReadJSON(&frame))
	return frame
}

func assertNoCustomerServiceWebSocketFrame(t *testing.T, connection *websocket.Conn) {
	t.Helper()

	require.NoError(t, connection.SetReadDeadline(time.Now().Add(100*time.Millisecond)))
	_, _, err := connection.ReadMessage()
	if err == nil {
		t.Fatal("unexpected customer-service websocket frame")
	}
	if websocket.IsUnexpectedCloseError(err) {
		t.Fatalf("unexpected websocket close: %v", err)
	}
	if !strings.Contains(err.Error(), "i/o timeout") {
		t.Fatalf("expected no websocket frame, got: %v", err)
	}
}
