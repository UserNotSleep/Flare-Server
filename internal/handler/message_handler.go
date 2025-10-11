package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"Flare-server/internal/repository"
)

type Message struct {
	ID         string `json:"id" firestore:"id"`
	Text       string `json:"text" firestore:"text"`
	SenderName string `json:"senderName" firestore:"senderName"`
	Timestamp  int64  `json:"timestamp" firestore:"timestamp"`
}

type MessageHandler struct {
	repo *repository.FirestoreRepo
}

func NewMessageHandler(repo *repository.FirestoreRepo) *MessageHandler {
	return &MessageHandler{repo: repo}
}

func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := h.repo.GetMessages(r.Context())
	if err != nil {
		log.Printf("❌ Error retrieving messages: %v", err)
		http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *MessageHandler) PostMessage(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Text string `json:"text"`
	}

	username := r.Context().Value("user")
	if username == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	msg := repository.Message{
		ID:         time.Now().Format("20060102150405.999999999"),
		Text:       input.Text,
		SenderName: username.(string),
		Timestamp:  time.Now().UnixMilli(),
	}

	err := h.repo.SaveMessage(r.Context(), msg)
	if err != nil {
		log.Printf("❌ Error saving message: %v", err)
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}
