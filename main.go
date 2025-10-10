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
	log.Printf("Using Firebase key: %s", cfg.FirebaseKey)
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(cfg.FirebaseKey))
	if err != nil {
		log.Fatalf("Ошибка иницализаций FireBase: %v", err)
	}

	firestoreClient, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("Ошибка создания клиента FireBase: %v", err)
	}
	defer firestoreClient.Close()

	repo := repository.NewFirestoreRepo(firestoreClient, cfg.Collection)
	svc := service.NewMessageService(repo)
	handler := handler.NewMessageHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetMessages(w, r)
		case http.MethodPost:
			handler.PostMessage(w, r)
		default:
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		}
	})

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: middleware.CORSMiddleware(mux),
	}

	log.Printf("Сервер запущен на порту %s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
