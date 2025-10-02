package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Message struct {
	ID         string `json:"id" firestore:"id"`
	Text       string `json:"text" firestore:"text"`
	SenderName string `json:"senderName" firestore:"senderName"`
	Timestamp  int64  `json:"timestamp" firestore:"timestamp"`
}

var (
	firestoreClient *firestore.Client
	appCtx          = context.Background()
)

func getMessages(w http.ResponseWriter, r *http.Request) {
	iter := firestoreClient.Collection("messages").OrderBy("timestamp", firestore.Asc).Documents(appCtx)
	defer iter.Stop()

	var messages []Message
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Ошибка чтения из Firestore: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		var msg Message
		if err := doc.DataTo(&msg); err != nil {
			log.Printf("Ошибка преобразования документа: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func postMessage(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Text       string `json:"text"`
		SenderName string `json:"senderName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	if input.Text == "" || input.SenderName == "" {
		http.Error(w, "Поля 'text' и 'senderName' обязательны", http.StatusBadRequest)
		return
	}

	newMsg := Message{
		ID:         time.Now().Format("20060102150405.999999999"),
		Text:       input.Text,
		SenderName: input.SenderName,
		Timestamp:  time.Now().UnixMilli(),
	}

	_, _, err := firestoreClient.Collection("messages").Add(appCtx, newMsg)
	if err != nil {
		log.Printf("Ошибка записи в Firestore: %v", err)
		http.Error(w, "Не удалось сохранить сообщение", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newMsg)
}

func corsMiddleware(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.URL.Path {
	case "/api/messages":
		switch r.Method {
		case http.MethodGet:
			getMessages(w, r)
		case http.MethodPost:
			postMessage(w, r)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}

func main() {

	keyPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT")
	if keyPath == "" {
		keyPath = "serviceAccountKey.json"
	}

	app, err := firebase.NewApp(appCtx, nil, option.WithCredentialsFile(keyPath))
	if err != nil {
		log.Fatalf("❌ Ошибка инициализации Firebase: %v", err)
	}

	firestoreClient, err = app.Firestore(appCtx)
	if err != nil {
		log.Fatalf("❌ Ошибка создания Firestore клиента: %v", err)
	}
	defer firestoreClient.Close()

	http.HandleFunc("/api/messages", corsMiddleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("✅ Сервер запущен на http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
