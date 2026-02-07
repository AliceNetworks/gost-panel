package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// allowedWSOrigins 存储允许的 WebSocket 来源
var allowedWSOrigins []string
var wsDebugMode bool

// SetWSOrigins 设置允许的 WebSocket 来源
func SetWSOrigins(origins []string, debug bool) {
	allowedWSOrigins = origins
	wsDebugMode = debug
}

// checkWSOrigin 验证 WebSocket 连接来源
func checkWSOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	// 如果没有 Origin 头 (同源请求)，允许
	if origin == "" {
		return true
	}

	// 调试模式下允许 localhost
	if wsDebugMode {
		if strings.HasPrefix(origin, "http://localhost") ||
			strings.HasPrefix(origin, "http://127.0.0.1") ||
			strings.HasPrefix(origin, "https://localhost") ||
			strings.HasPrefix(origin, "https://127.0.0.1") {
			return true
		}
	}

	// 如果未配置允许来源列表，使用同源策略
	if len(allowedWSOrigins) == 0 {
		// 检查 Origin 是否与 Host 匹配
		host := r.Host
		// 从 origin 提取 host 部分
		originHost := extractHost(origin)
		return originHost == host
	}

	// 检查是否在允许列表中
	for _, allowed := range allowedWSOrigins {
		if allowed == "*" {
			return true
		}
		if origin == allowed {
			return true
		}
		// 支持通配符前缀匹配
		if strings.HasPrefix(allowed, "*.") {
			suffix := allowed[1:] // ".example.com"
			originHost := extractHost(origin)
			if strings.HasSuffix(originHost, suffix) || originHost == allowed[2:] {
				return true
			}
		}
	}

	log.Printf("WebSocket: rejected origin %s (allowed: %v)", origin, allowedWSOrigins)
	return false
}

// extractHost 从 URL 提取 host
func extractHost(urlStr string) string {
	// 简单解析: http://example.com:8080/path -> example.com:8080
	s := urlStr
	if idx := strings.Index(s, "://"); idx >= 0 {
		s = s[idx+3:]
	}
	if idx := strings.Index(s, "/"); idx >= 0 {
		s = s[:idx]
	}
	return s
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkWSOrigin,
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WSClient represents a WebSocket client connection
type WSClient struct {
	hub    *WSHub
	conn   *websocket.Conn
	send   chan []byte
	userID uint
}

// WSHub maintains active WebSocket connections
type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.RWMutex
}

// NewWSHub creates a new WebSocket hub
func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

// Run starts the hub's main loop
func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected, total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected, total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *WSHub) Broadcast(msgType string, data interface{}) {
	msg := WSMessage{
		Type: msgType,
		Data: data,
	}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}

	select {
	case h.broadcast <- jsonData:
	default:
		log.Println("WebSocket broadcast channel full, dropping message")
	}
}

// ClientCount returns the number of connected clients
func (h *WSHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived, websocket.CloseNormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}

	client := &WSClient{
		hub:  s.wsHub,
		conn: conn,
		send: make(chan []byte, 256),
	}

	s.wsHub.register <- client

	go client.writePump()
	go client.readPump()
}

// BroadcastNodeStatus broadcasts node status update
func (s *Server) BroadcastNodeStatus(nodeID uint, status string, connections int, trafficIn, trafficOut int64) {
	if s.wsHub == nil {
		return
	}

	s.wsHub.Broadcast("node_status", map[string]interface{}{
		"node_id":     nodeID,
		"status":      status,
		"connections": connections,
		"traffic_in":  trafficIn,
		"traffic_out": trafficOut,
		"timestamp":   time.Now().Unix(),
	})
}

// BroadcastStats broadcasts dashboard stats update
func (s *Server) BroadcastStats(stats interface{}) {
	if s.wsHub == nil {
		return
	}
	s.wsHub.Broadcast("stats", stats)
}
