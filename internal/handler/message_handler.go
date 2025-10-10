package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"Flare-server/internal/service"
)

type MessageHandler struct {
	service *service.MessageService
}

func NewMessageHandler(svc *service.MessageService) *MessageHandler {
	return &MessageHandler{service: svc}
}

func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := h.service.GetMessages(r.Context())
	if err != nil {
		log.Printf("❌ Error retrieving messages: %v", err) // ← Добавь это
		http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *MessageHandler) PostMessage(w http.ResponseWriter, r *http.Request) {
	var input service.CreateMessageInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	msg, err := h.service.CreateMessage(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}
