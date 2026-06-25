package ticket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

type Hub struct {
	sync.RWMutex
	clients map[*websocket.Conn]bool
}

var globalHub = &Hub{
	clients: make(map[*websocket.Conn]bool),
}

// ServeWS upgrades the HTTP connection to a WebSocket connection
func (h *Handler) ServeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("[CRITICAL] WebSocket upgrade failed:", err)
		return
	}

	globalHub.Lock()
	globalHub.clients[conn] = true
	globalHub.Unlock()

	defer func() {
		globalHub.Lock()
		delete(globalHub.clients, conn)
		globalHub.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[CRITICAL] WebSocket error: %v", err)
			}
			break
		}
	}
}
