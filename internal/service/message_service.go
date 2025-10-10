package service

import (
	"context"
	"fmt"
	"time"

	"Flare-server/internal/repository"
)

type MessageService struct {
	repo *repository.FirestoreRepo
}

func NewMessageService(repo *repository.FirestoreRepo) *MessageService {
	return &MessageService{repo: repo}
}

func (s *MessageService) GetMessages(ctx context.Context) ([]repository.Message, error) {
	return s.repo.GetMessages(ctx)
}

func (s *MessageService) CreateMessage(ctx context.Context, input CreateMessageInput) (*repository.Message, error) {
	if input.Text == "" || input.SenderName == "" {
		return nil, fmt.Errorf("text and senderName are required")
	}

	msg := repository.Message{
		ID:         time.Now().Format("20060102150405.999999999"),
		Text:       input.Text,
		SenderName: input.SenderName,
		Timestamp:  time.Now(),
	}

	err := s.repo.SaveMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

type CreateMessageInput struct {
	Text       string `json:"text" validate:"required,max=1000"`
	SenderName string `json:"senderName" validate:"required,max=100"`
}
