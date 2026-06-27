package ticket

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	websocketReadBufferSize  = 1024
	websocketWriteBufferSize = 1024
	websocketMaxClients      = 500
	websocketMaxMessageSize  = 4 << 10
	websocketWriteWait       = 10 * time.Second
	websocketPongWait        = 60 * time.Second
	websocketPingPeriod      = 50 * time.Second
)

type Hub struct {
	sync.RWMutex
	clients map[*websocket.Conn]bool
}

var globalHub = &Hub{
	clients: make(map[*websocket.Conn]bool),
}

// ServeWS upgrades the HTTP connection to a WebSocket connection
func (h *Handler) ServeWS(c *gin.Context) {
	if publicCustomerUserID(c) == nil {
		if _, ok := h.existingVisitorSessionHash(c); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "visitor_session_required", "message": "visitor session is required"})
			return
		}
	}

	if globalHub.count() >= websocketMaxClients {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "too_many_connections", "message": "too many websocket connections"})
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  websocketReadBufferSize,
		WriteBufferSize: websocketWriteBufferSize,
		CheckOrigin:     h.checkWSOrigin,
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("[CRITICAL] WebSocket upgrade failed:", err)
		return
	}

	if !globalHub.add(conn) {
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "too many websocket connections"), time.Now().Add(websocketWriteWait))
		_ = conn.Close()
		return
	}

	defer func() {
		globalHub.remove(conn)
		_ = conn.Close()
	}()

	conn.SetReadLimit(websocketMaxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(websocketPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(websocketPongWait))
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[CRITICAL] WebSocket error: %v", err)
				}
				return
			}
		}
	}()

	ticker := time.NewTicker(websocketPingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := conn.SetWriteDeadline(time.Now().Add(websocketWriteWait)); err != nil {
				return
			}
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Handler) checkWSOrigin(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}

	originURL, err := url.Parse(origin)
	if err != nil || originURL.Scheme == "" || originURL.Host == "" {
		return false
	}
	if sameHost(originURL.Host, r.Host) {
		return true
	}

	for _, allowedOrigin := range h.allowedOrigins {
		if originMatchesAllowed(originURL, allowedOrigin) {
			return true
		}
	}
	return false
}

func originMatchesAllowed(originURL *url.URL, allowedOrigin string) bool {
	allowedOrigin = strings.TrimSpace(allowedOrigin)
	if allowedOrigin == "" || allowedOrigin == "*" {
		return false
	}

	allowedURL, err := url.Parse(allowedOrigin)
	if err != nil || allowedURL.Scheme == "" || allowedURL.Host == "" {
		return false
	}
	return strings.EqualFold(originURL.Scheme, allowedURL.Scheme) && sameHost(originURL.Host, allowedURL.Host)
}

func sameHost(left string, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}

func (h *Hub) add(conn *websocket.Conn) bool {
	h.Lock()
	defer h.Unlock()
	if len(h.clients) >= websocketMaxClients {
		return false
	}
	h.clients[conn] = true
	return true
}

func (h *Hub) remove(conn *websocket.Conn) {
	h.Lock()
	delete(h.clients, conn)
	h.Unlock()
}

func (h *Hub) count() int {
	h.RLock()
	defer h.RUnlock()
	return len(h.clients)
}
