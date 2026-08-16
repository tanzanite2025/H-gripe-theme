package realtime

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	appLogger "commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/metrics"
	"commerce-platform/internal/service"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	customerServiceWebSocketReadBufferSize  = 1024
	customerServiceWebSocketWriteBufferSize = 1024
	customerServiceWebSocketMaxConnections  = 500
	customerServiceWebSocketMaxMessageSize  = 4 << 10
	customerServiceWebSocketWriteWait       = 10 * time.Second
	customerServiceWebSocketPongWait        = 60 * time.Second
	customerServiceWebSocketPingPeriod      = 50 * time.Second
	customerServiceWebSocketOutboundBuffer  = 256
)

var customerServiceWebSocketConnections atomic.Int64

// CustomerServiceWebSocketControl is deliberately limited to transient
// controls. Durable conversation mutations continue to use the HTTP commands.
type CustomerServiceWebSocketControl struct {
	Type     string `json:"type"`
	IsTyping *bool  `json:"is_typing,omitempty"`
}

type customerServiceWebSocketFrame struct {
	Type   string                                `json:"type"`
	Cursor string                                `json:"cursor,omitempty"`
	Event  *service.CustomerServiceRealtimeEvent `json:"event,omitempty"`
	Code   string                                `json:"code,omitempty"`
}

// CustomerServiceWebSocketOptions scopes one already-authorized connection.
// Subscription must be created before Serve so replay and live delivery cannot
// leave an intentional subscription gap.
type CustomerServiceWebSocketOptions struct {
	CheckOrigin   func(*http.Request) bool
	Subscription  *service.CustomerServiceEventSubscription
	Replay        []service.CustomerServiceRealtimeEvent
	AllowEvent    func(service.CustomerServiceRealtimeEvent) bool
	HandleControl func(CustomerServiceWebSocketControl)
}

// ServeCustomerServiceWebSocket owns all writes for one socket. A full outbound
// queue closes the connection so the client reconnects and reconciles through
// the authoritative HTTP APIs instead of ever blocking an event publisher.
func ServeCustomerServiceWebSocket(w http.ResponseWriter, r *http.Request, options CustomerServiceWebSocketOptions) {
	if options.Subscription == nil || options.Subscription.Events() == nil {
		http.Error(w, "customer service realtime unavailable", http.StatusServiceUnavailable)
		return
	}
	if !acquireCustomerServiceWebSocketConnection() {
		metrics.CustomerServiceRealtimeWebSocketConnectionAttempts.WithLabelValues("capacity_rejected").Inc()
		http.Error(w, "too many websocket connections", http.StatusServiceUnavailable)
		return
	}
	defer releaseCustomerServiceWebSocketConnection()
	defer options.Subscription.Cancel()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  customerServiceWebSocketReadBufferSize,
		WriteBufferSize: customerServiceWebSocketWriteBufferSize,
		CheckOrigin:     options.CheckOrigin,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		metrics.CustomerServiceRealtimeWebSocketConnectionAttempts.WithLabelValues("upgrade_failed").Inc()
		appLogger.Warn("customer-service websocket upgrade failed", zap.Error(err))
		return
	}
	metrics.CustomerServiceRealtimeWebSocketConnectionAttempts.WithLabelValues("accepted").Inc()

	done := make(chan struct{})
	var stopOnce sync.Once
	stop := func() {
		stopOnce.Do(func() {
			close(done)
		})
	}

	outbound := make(chan customerServiceWebSocketFrame, customerServiceWebSocketOutboundBuffer)
	enqueue := func(frame customerServiceWebSocketFrame) bool {
		select {
		case <-done:
			return false
		case outbound <- frame:
			return true
		default:
			metrics.CustomerServiceRealtimeWebSocketOutboundOverflows.Inc()
			appLogger.Warn("customer-service websocket outbound queue overflow")
			stop()
			return false
		}
	}

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		defer func() { _ = conn.Close() }()

		pingTicker := time.NewTicker(customerServiceWebSocketPingPeriod)
		defer pingTicker.Stop()

		for {
			select {
			case <-done:
				_ = conn.WriteControl(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing"),
					time.Now().Add(customerServiceWebSocketWriteWait),
				)
				return
			case frame := <-outbound:
				if err := conn.SetWriteDeadline(time.Now().Add(customerServiceWebSocketWriteWait)); err != nil {
					stop()
					return
				}
				if err := conn.WriteJSON(frame); err != nil {
					stop()
					return
				}
			case <-pingTicker.C:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(customerServiceWebSocketWriteWait)); err != nil {
					stop()
					return
				}
			}
		}
	}()

	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		defer stop()

		conn.SetReadLimit(customerServiceWebSocketMaxMessageSize)
		_ = conn.SetReadDeadline(time.Now().Add(customerServiceWebSocketPongWait))
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(customerServiceWebSocketPongWait))
		})

		for {
			messageType, payload, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					appLogger.Warn("customer-service websocket unexpected close", zap.Error(err))
				}
				return
			}
			if messageType != websocket.TextMessage {
				enqueue(customerServiceWebSocketFrame{Type: "error", Code: "text_frame_required"})
				continue
			}

			control, ok := decodeCustomerServiceWebSocketControl(payload)
			if !ok {
				enqueue(customerServiceWebSocketFrame{Type: "error", Code: "invalid_control"})
				continue
			}
			switch control.Type {
			case "ping":
				enqueue(customerServiceWebSocketFrame{Type: "pong"})
			case "typing":
				if options.HandleControl != nil {
					options.HandleControl(control)
				}
			default:
				enqueue(customerServiceWebSocketFrame{Type: "error", Code: "unsupported_control"})
			}
		}
	}()

	defer func() {
		stop()
		<-writerDone
		<-readerDone
	}()

	deduper := service.NewCustomerServiceRealtimeEventDeduper(2048)
	deliver := func(event service.CustomerServiceRealtimeEvent) {
		if options.AllowEvent != nil && !options.AllowEvent(event) {
			if event.StreamID != "" {
				enqueue(customerServiceWebSocketFrame{Type: "cursor", Cursor: event.StreamID})
			}
			return
		}
		if !deduper.First(event) {
			return
		}
		eventCopy := event
		enqueue(customerServiceWebSocketFrame{Type: "event", Cursor: event.StreamID, Event: &eventCopy})
	}

	if !enqueue(customerServiceWebSocketFrame{Type: "ready"}) {
		return
	}
	for _, event := range options.Replay {
		deliver(event)
	}

	for {
		select {
		case <-done:
			return
		case <-readerDone:
			return
		case event, ok := <-options.Subscription.Events():
			if !ok {
				return
			}
			deliver(event)
		}
	}
}

func decodeCustomerServiceWebSocketControl(payload []byte) (CustomerServiceWebSocketControl, bool) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()

	var control CustomerServiceWebSocketControl
	if err := decoder.Decode(&control); err != nil {
		return CustomerServiceWebSocketControl{}, false
	}
	var trailing interface{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		return CustomerServiceWebSocketControl{}, false
	}

	control.Type = strings.ToLower(strings.TrimSpace(control.Type))
	return control, control.Type != ""
}

// CustomerServiceWebSocketOriginAllowed requires an explicit same-origin or
// configured allowed-origin browser upgrade. Native browser sockets always send
// Origin, so accepting a missing value would weaken the cookie-auth boundary.
func CustomerServiceWebSocketOriginAllowed(r *http.Request, allowedOrigins []string) bool {
	if r == nil {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}

	originURL, err := url.Parse(origin)
	if err != nil || originURL.Scheme == "" || originURL.Host == "" {
		return false
	}
	if sameCustomerServiceWebSocketOrigin(originURL, requestCustomerServiceWebSocketOrigin(r)) {
		return true
	}

	for _, allowedOrigin := range allowedOrigins {
		allowedURL, err := url.Parse(strings.TrimSpace(allowedOrigin))
		if err != nil || allowedURL.Scheme == "" || allowedURL.Host == "" {
			continue
		}
		if sameCustomerServiceWebSocketOrigin(originURL, allowedURL) {
			return true
		}
	}
	return false
}

func requestCustomerServiceWebSocketOrigin(r *http.Request) *url.URL {
	scheme := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if comma := strings.IndexByte(scheme, ','); comma >= 0 {
		scheme = strings.TrimSpace(scheme[:comma])
	}
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return &url.URL{Scheme: scheme, Host: r.Host}
}

func sameCustomerServiceWebSocketOrigin(left, right *url.URL) bool {
	if left == nil || right == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(left.Scheme), strings.TrimSpace(right.Scheme)) &&
		strings.EqualFold(strings.TrimSpace(left.Host), strings.TrimSpace(right.Host))
}

func acquireCustomerServiceWebSocketConnection() bool {
	if customerServiceWebSocketConnections.Add(1) <= customerServiceWebSocketMaxConnections {
		metrics.CustomerServiceRealtimeWebSocketConnections.Inc()
		return true
	}
	customerServiceWebSocketConnections.Add(-1)
	return false
}

func releaseCustomerServiceWebSocketConnection() {
	customerServiceWebSocketConnections.Add(-1)
	metrics.CustomerServiceRealtimeWebSocketConnections.Dec()
}
