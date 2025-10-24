package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"Flare-server/internal/models"
	"Flare-server/internal/service"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	ChatID  string      `json:"chatId,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type Client struct {
	ID       string
	UserID   string
	Username string
	Conn     *websocket.Conn
	Send     chan WebSocketMessage
	Hub      *Hub
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan WebSocketMessage
	register   chan *Client
	unregister chan *Client
	chatRooms  map[string]map[*Client]bool
	mutex      sync.RWMutex
}

type WebSocketHandler struct {
	hub         *Hub
	chatService *service.ChatService
	authService *service.AuthService
}

func NewWebSocketHandler(chatService *service.ChatService, authService *service.AuthService) *WebSocketHandler {
	hub := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan WebSocketMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		chatRooms:  make(map[string]map[*Client]bool),
	}

	handler := &WebSocketHandler{
		hub:         hub,
		chatService: chatService,
		authService: authService,
	}

	go hub.run()
	return handler
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("Authorization")
		if token != "" && len(token) > 7 {
			token = token[7:]
		}
	}

	if token == "" {
		http.Error(w, "Token required", http.StatusUnauthorized)
		return
	}

	userInfo, err := h.validateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:       generateClientID(),
		UserID:   userInfo.ID,
		Username: userInfo.Username,
		Conn:     conn,
		Send:     make(chan WebSocketMessage, 256),
		Hub:      h.hub,
	}

	h.hub.register <- client

	go client.writePump()
	go client.readPump(h)
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client %s connected", client.Username)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				
				for chatID, clients := range h.chatRooms {
					if _, exists := clients[client]; exists {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.chatRooms, chatID)
						}
					}
				}
			}
			h.mutex.Unlock()
			log.Printf("Client %s disconnected", client.Username)

		case message := <-h.broadcast:
			h.mutex.RLock()
			if message.ChatID != "" {
				if clients, exists := h.chatRooms[message.ChatID]; exists {
					for client := range clients {
						select {
						case client.Send <- message:
						default:
							close(client.Send)
							delete(h.clients, client)
							delete(clients, client)
						}
					}
				}
			} else {
				for client := range h.clients {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (c *Client) readPump(handler *WebSocketHandler) {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg WebSocketMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		handler.handleMessage(c, msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WebSocketHandler) handleMessage(client *Client, msg WebSocketMessage) {
	switch msg.Type {
	case "join_chat":
		h.handleJoinChat(client, msg)
	case "leave_chat":
		h.handleLeaveChat(client, msg)
	case "send_message":
		h.handleSendMessage(client, msg)
	case "typing":
		h.handleTyping(client, msg)
	default:
		client.Send <- WebSocketMessage{
			Type:  "error",
			Error: "Unknown message type",
		}
	}
}

func (h *WebSocketHandler) handleJoinChat(client *Client, msg WebSocketMessage) {
	chatID, ok := msg.Data.(string)
	if !ok {
		client.Send <- WebSocketMessage{
			Type:  "error",
			Error: "Invalid chat ID",
		}
		return
	}

	ctx := context.Background()
	isMember, err := h.chatService.IsUserInChat(ctx, chatID, client.UserID)
	if err != nil || !isMember {
		client.Send <- WebSocketMessage{
			Type:  "error",
			Error: "Access denied",
		}
		return
	}

	h.hub.mutex.Lock()
	if h.hub.chatRooms[chatID] == nil {
		h.hub.chatRooms[chatID] = make(map[*Client]bool)
	}
	h.hub.chatRooms[chatID][client] = true
	h.hub.mutex.Unlock()

	client.Send <- WebSocketMessage{
		Type:   "joined_chat",
		ChatID: chatID,
	}
}

func (h *WebSocketHandler) handleLeaveChat(client *Client, msg WebSocketMessage) {
	chatID, ok := msg.Data.(string)
	if !ok {
		return
	}

	h.hub.mutex.Lock()
	if clients, exists := h.hub.chatRooms[chatID]; exists {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.hub.chatRooms, chatID)
		}
	}
	h.hub.mutex.Unlock()

	client.Send <- WebSocketMessage{
		Type:   "left_chat",
		ChatID: chatID,
	}
}

func (h *WebSocketHandler) handleSendMessage(client *Client, msg WebSocketMessage) {
	var messageData struct {
		ChatID string `json:"chatId"`
		Text   string `json:"text"`
	}

	dataBytes, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(dataBytes, &messageData); err != nil {
		client.Send <- WebSocketMessage{
			Type:  "error",
			Error: "Invalid message data",
		}
		return
	}

	ctx := context.Background()
	req := models.SendMessageRequest{Text: messageData.Text}
	message, err := h.chatService.SendMessage(ctx, messageData.ChatID, client.UserID, client.Username, req)
	if err != nil {
		client.Send <- WebSocketMessage{
			Type:  "error",
			Error: err.Error(),
		}
		return
	}

	h.hub.broadcast <- WebSocketMessage{
		Type:   "new_message",
		ChatID: messageData.ChatID,
		Data:   message,
	}
}

func (h *WebSocketHandler) handleTyping(client *Client, msg WebSocketMessage) {
	chatID, ok := msg.Data.(string)
	if !ok {
		return
	}

	h.hub.broadcast <- WebSocketMessage{
		Type:   "user_typing",
		ChatID: chatID,
		Data: map[string]interface{}{
			"userId":   client.UserID,
			"username": client.Username,
		},
	}
}

func (h *WebSocketHandler) BroadcastMessage(chatID string, messageType string, data interface{}) {
	h.hub.broadcast <- WebSocketMessage{
		Type:   messageType,
		ChatID: chatID,
		Data:   data,
	}
}


func (h *WebSocketHandler) validateToken(token string) (*UserInfo, error) {
	return &UserInfo{
		ID:       "user123",
		Username: "testuser",
	}, nil
}

func generateClientID() string {
	return time.Now().Format("20060102150405.999999999")
}
