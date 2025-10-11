package main

import (
	"context"
	"log"
	"net/http"

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

	// Repos
	userRepo := repository.NewUserRepo(firestoreClient)
	messageRepo := repository.NewFirestoreRepo(firestoreClient, cfg.Collection)

	// Services
	authService := service.NewAuthService(userRepo, []byte(cfg.JWTSecret))
	messageHandler := handler.NewMessageHandler(messageRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/register", authHandler.Register)
	mux.HandleFunc("/api/login", authHandler.Login)
	mux.HandleFunc("/api/logout", authHandler.Logout)

	// Защищённые роуты
	protected := middleware.AuthMiddleware(authService)

	// Сообщения
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

	// Профиль
	mux.Handle("/api/profile", protected(http.HandlerFunc(authHandler.Profile)))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: middleware.CORSMiddleware(mux),
	}

	log.Printf("✅ Server running on http://localhost:%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
