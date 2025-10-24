package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"Flare-server/internal/models"
	"Flare-server/internal/service"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userInfo := getUserFromContext(r.Context())
	if userInfo == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req models.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Type != models.ChatTypePrivate && req.Type != models.ChatTypeGroup {
		http.Error(w, "Invalid chat type", http.StatusBadRequest)
		return
	}

	chat, err := h.chatService.CreateChat(r.Context(), req, userInfo.ID, userInfo.Username)
	if err != nil {
		log.Printf("❌ Error creating chat: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chat)
}

func (h *ChatHandler) GetUserChats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userInfo := getUserFromContext(r.Context())
	if userInfo == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	chats, err := h.chatService.GetUserChats(r.Context(), userInfo.ID)
	if err != nil {
		log.Printf("❌ Error getting user chats: %v", err)
		http.Error(w, "Failed to get chats", http.StatusInternalServerError)
		return
	}

	response := models.ChatListResponse{Chats: chats}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ChatHandler) GetChatInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatID := extractChatID(r.URL.Path)
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	userInfo := getUserFromContext(r.Context())
	if userInfo == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	chatInfo, err := h.chatService.GetChatInfo(r.Context(), chatID, userInfo.ID)
	if err != nil {
		log.Printf("❌ Error getting chat info: %v", err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatInfo)
}

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatID := extractChatID(r.URL.Path)
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	userInfo := getUserFromContext(r.Context())
	if userInfo == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	message, err := h.chatService.SendMessage(r.Context(), chatID, userInfo.ID, userInfo.Username, req)
	if err != nil {
		log.Printf("❌ Error sending message: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

func (h *ChatHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatID := extractChatID(r.URL.Path)
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	userInfo := getUserFromContext(r.Context())
	if userInfo == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	limit := 50
	if l := query.Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	lastMessageID := query.Get("lastMessageId")

	messages, err := h.chatService.GetChatMessages(r.Context(), chatID, userInfo.ID, limit, lastMessageID)
	if err != nil {
		log.Printf("❌ Error getting chat messages: %v", err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *ChatHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatID := extractChatID(r.URL.Path)
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	userInfo := getUserFromContext(r.Context())
	if userInfo == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req models.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.chatService.AddMemberToChat(r.Context(), chatID, userInfo.ID, req); err != nil {
		log.Printf("❌ Error adding member: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Member added successfully"})
}

func (h *ChatHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatID := extractChatID(r.URL.Path)
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	targetUserID := r.URL.Query().Get("userId")
	if targetUserID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userInfo := getUserFromContext(r.Context())
	if userInfo == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if err := h.chatService.RemoveMemberFromChat(r.Context(), chatID, userInfo.ID, targetUserID); err != nil {
		log.Printf("❌ Error removing member: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Member removed successfully"})
}

func (h *ChatHandler) LeaveChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatID := extractChatID(r.URL.Path)
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	userInfo := getUserFromContext(r.Context())
	if userInfo == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if err := h.chatService.LeaveChat(r.Context(), chatID, userInfo.ID); err != nil {
		log.Printf("❌ Error leaving chat: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Left chat successfully"})
}

func (h *ChatHandler) UpdateChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatID := extractChatID(r.URL.Path)
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	userInfo := getUserFromContext(r.Context())
	if userInfo == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.chatService.UpdateChat(r.Context(), chatID, userInfo.ID, updates); err != nil {
		log.Printf("❌ Error updating chat: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Chat updated successfully"})
}

func (h *ChatHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatID := extractChatID(r.URL.Path)
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	userInfo := getUserFromContext(r.Context())
	if userInfo == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if err := h.chatService.DeleteChat(r.Context(), chatID, userInfo.ID); err != nil {
		log.Printf("❌ Error deleting chat: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Chat deleted successfully"})
}


type UserInfo struct {
	ID       string
	Username string
}

func getUserFromContext(ctx context.Context) *UserInfo {
	userInfo := ctx.Value("user")
	if userInfo == nil {
		return nil
	}
	
	userMap, ok := userInfo.(map[string]interface{})
	if !ok {
		return nil
	}
	
	userID, hasID := userMap["userID"].(string)
	username, hasUsername := userMap["username"].(string)
	
	if !hasID || !hasUsername {
		return nil
	}
	
	return &UserInfo{
		ID:       userID,
		Username: username,
	}
}

func extractChatID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "chats" {
		return parts[2]
	}
	return ""
}
