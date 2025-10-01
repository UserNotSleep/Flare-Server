package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Message struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	SenderName string `json:"senderName"`
	Timestamp  int64  `json:"timestamp"`
}

var (
	messages []Message
	mu       sync.RWMutex
)

func getMessages(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func postMessage(w http.ResponseWriter, r *http.Request) {
	var msg struct {
		Text       string `json:"text"`
		SenderName string `json:"senderName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	if msg.Text == "" || msg.SenderName == "" {
		http.Error(w, "Поля text и senderName обязательны", http.StatusBadRequest)
		return
	}

	newMsg := Message{
		ID:         time.Now().Format("20060102150405.999999999"),
		Text:       msg.Text,
		SenderName: msg.SenderName,
		Timestamp:  time.Now().UnixMilli(),
	}

	mu.Lock()
	messages = append(messages, newMsg)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newMsg)
}

func main() {
	http.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodGet {
			getMessages(w, r)
		} else if r.Method == http.MethodPost {
			postMessage(w, r)
		} else {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	port := ":3000"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	log.Printf("✅ Сервер запущен на порту %s", port[1:])
	log.Fatal(http.ListenAndServe(port, nil))
}
