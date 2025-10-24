package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"Flare-server/internal/config"
	"Flare-server/internal/handler"
	"Flare-server/internal/middleware"
	"Flare-server/internal/repository"
	"Flare-server/internal/service"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func main() {
	cfg := config.Load()

	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(cfg.FirebaseKey))
	if err != nil {
		log.Fatalf("❌ Failed to initialize Firebase: %v", err)
	}

	firestoreClient, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	userRepo := repository.NewUserRepo(firestoreClient)
	messageRepo := repository.NewFirestoreRepo(firestoreClient, cfg.Collection)
	chatRepo := repository.NewChatRepo(firestoreClient)

	authService := service.NewAuthService(userRepo, []byte(cfg.JWTSecret))
	chatService := service.NewChatService(chatRepo, userRepo)
	messageHandler := handler.NewMessageHandler(messageRepo)

	authHandler := handler.NewAuthHandler(authService)
	chatHandler := handler.NewChatHandler(chatService)
	wsHandler := handler.NewWebSocketHandler(chatService, authService)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/register", authHandler.Register)
	mux.HandleFunc("/api/login", authHandler.Login)
	mux.HandleFunc("/api/logout", authHandler.Logout)

	protected := middleware.AuthMiddleware(authService)

	mux.Handle("/api/messages", protected(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			messageHandler.GetMessages(w, r)
		case http.MethodPost:
			messageHandler.PostMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/profile", protected(http.HandlerFunc(authHandler.Profile)))

	mux.Handle("/api/chats", protected(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			chatHandler.GetUserChats(w, r)
		case http.MethodPost:
			chatHandler.CreateChat(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/chats/", protected(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if r.URL.Path != "/api/chats/" && !strings.Contains(path[len("/api/chats/"):], "/") {
			switch r.Method {
			case http.MethodGet:
				chatHandler.GetChatInfo(w, r)
			case http.MethodPut:
				chatHandler.UpdateChat(w, r)
			case http.MethodDelete:
				chatHandler.DeleteChat(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/messages") {
			switch r.Method {
			case http.MethodGet:
				chatHandler.GetChatMessages(w, r)
			case http.MethodPost:
				chatHandler.SendMessage(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/members") {
			switch r.Method {
			case http.MethodPost:
				chatHandler.AddMember(w, r)
			case http.MethodDelete:
				chatHandler.RemoveMember(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/leave") {
			if r.Method == http.MethodPost {
				chatHandler.LeaveChat(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		http.Error(w, "Not found", http.StatusNotFound)
	})))

	mux.HandleFunc("/api/ws", wsHandler.HandleWebSocket)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: middleware.CORSMiddleware(mux),
	}

	log.Printf("✅ Server running on http://localhost:%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
