package repository

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type Message struct {
	ID         string `json:"id" firestore:"id"`
	Text       string `json:"text" firestore:"text"`
	SenderName string `json:"senderName" firestore:"senderName"`
	Timestamp  int64  `json:"timestamp" firestore:"timestamp"`
}

type FirestoreRepo struct {
	client     *firestore.Client
	collection string
}

func NewFirestoreRepo(client *firestore.Client, collection string) *FirestoreRepo {
	return &FirestoreRepo{client: client, collection: collection}
}

func (r *FirestoreRepo) GetMessages(ctx context.Context) ([]Message, error) {
	log.Printf("Querying Firestore collection: %s", r.collection) // ← Добавь
	iter := r.client.Collection(r.collection).OrderBy("timestamp", firestore.Asc).Documents(ctx)
	defer iter.Stop()

	var messages []Message
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Firestore iteration error: %v", err) // ← Добавь
			return nil, fmt.Errorf("failed to iterate messages: %w", err)
		}

		var msg Message
		if err := doc.DataTo(&msg); err != nil {
			log.Printf("Firestore decode error: %v", err) // ← Добавь
			return nil, fmt.Errorf("failed to decode message: %w", err)
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *FirestoreRepo) SaveMessage(ctx context.Context, msg Message) error {
	_, _, err := r.client.Collection(r.collection).Add(ctx, msg)
	if err != nil {
		return fmt.Errorf("Ошибка сохранения сообщения: %w", err)
	}
	return nil
}
