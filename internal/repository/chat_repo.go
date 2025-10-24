package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"Flare-server/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type ChatRepo struct {
	client *firestore.Client
}

func NewChatRepo(client *firestore.Client) *ChatRepo {
	return &ChatRepo{client: client}
}

func (r *ChatRepo) CreateChat(ctx context.Context, chat models.Chat) (*models.Chat, error) {
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()
	
	docRef, _, err := r.client.Collection("chats").Add(ctx, chat)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}
	
	chat.ID = docRef.ID
	return &chat, nil
}

func (r *ChatRepo) GetChatByID(ctx context.Context, chatID string) (*models.Chat, error) {
	doc, err := r.client.Collection("chats").Doc(chatID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}
	
	var chat models.Chat
	if err := doc.DataTo(&chat); err != nil {
		return nil, fmt.Errorf("failed to decode chat: %w", err)
	}
	
	chat.ID = doc.Ref.ID
	return &chat, nil
}

func (r *ChatRepo) GetUserChats(ctx context.Context, userID string) ([]models.Chat, error) {
	iter := r.client.Collection("chat_members").Where("userId", "==", userID).Documents(ctx)
	defer iter.Stop()
	
	var chatIDs []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate chat members: %w", err)
		}
		
		var member models.ChatMember
		if err := doc.DataTo(&member); err != nil {
			continue
		}
		chatIDs = append(chatIDs, member.ChatID)
	}
	
	if len(chatIDs) == 0 {
		return []models.Chat{}, nil
	}
	
	var chats []models.Chat
	for _, chatID := range chatIDs {
		chat, err := r.GetChatByID(ctx, chatID)
		if err != nil {
			log.Printf("Failed to get chat %s: %v", chatID, err)
			continue
		}
		
		lastMessage, err := r.GetLastMessage(ctx, chatID)
		if err == nil {
			chat.LastMessage = lastMessage
		}
		
		chats = append(chats, *chat)
	}
	
	return chats, nil
}

func (r *ChatRepo) AddChatMember(ctx context.Context, member models.ChatMember) error {
	member.JoinedAt = time.Now()
	
	_, _, err := r.client.Collection("chat_members").Add(ctx, member)
	if err != nil {
		return fmt.Errorf("failed to add chat member: %w", err)
	}
	
	_, err = r.client.Collection("chats").Doc(member.ChatID).Update(ctx, []firestore.Update{
		{Path: "memberCount", Value: firestore.Increment(1)},
		{Path: "updatedAt", Value: time.Now()},
	})
	
	return err
}

func (r *ChatRepo) RemoveChatMember(ctx context.Context, chatID, userID string) error {
	iter := r.client.Collection("chat_members").
		Where("chatId", "==", chatID).
		Where("userId", "==", userID).
		Documents(ctx)
	defer iter.Stop()
	
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to find chat member: %w", err)
		}
		
		_, err = doc.Ref.Delete(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete chat member: %w", err)
		}
		
		_, err = r.client.Collection("chats").Doc(chatID).Update(ctx, []firestore.Update{
			{Path: "memberCount", Value: firestore.Increment(-1)},
			{Path: "updatedAt", Value: time.Now()},
		})
		
		return err
	}
	
	return fmt.Errorf("chat member not found")
}

func (r *ChatRepo) GetChatMembers(ctx context.Context, chatID string) ([]models.ChatMember, error) {
	iter := r.client.Collection("chat_members").Where("chatId", "==", chatID).Documents(ctx)
	defer iter.Stop()
	
	var members []models.ChatMember
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate chat members: %w", err)
		}
		
		var member models.ChatMember
		if err := doc.DataTo(&member); err != nil {
			continue
		}
		member.ID = doc.Ref.ID
		members = append(members, member)
	}
	
	return members, nil
}

func (r *ChatRepo) IsUserInChat(ctx context.Context, chatID, userID string) (bool, error) {
	iter := r.client.Collection("chat_members").
		Where("chatId", "==", chatID).
		Where("userId", "==", userID).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()
	
	_, err := iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check chat membership: %w", err)
	}
	
	return true, nil
}

func (r *ChatRepo) SaveMessage(ctx context.Context, message models.Message) (*models.Message, error) {
	message.Timestamp = time.Now()
	
	docRef, _, err := r.client.Collection("messages").Add(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}
	
	message.ID = docRef.ID
	
	_, err = r.client.Collection("chats").Doc(message.ChatID).Update(ctx, []firestore.Update{
		{Path: "updatedAt", Value: time.Now()},
	})
	if err != nil {
		log.Printf("Failed to update chat updatedAt: %v", err)
	}
	
	return &message, nil
}

func (r *ChatRepo) GetChatMessages(ctx context.Context, chatID string, limit int, lastMessageID string) ([]models.Message, error) {
	query := r.client.Collection("messages").
		Where("chatId", "==", chatID).
		OrderBy("timestamp", firestore.Desc).
		Limit(limit)
	
	if lastMessageID != "" {
		lastDoc, err := r.client.Collection("messages").Doc(lastMessageID).Get(ctx)
		if err == nil {
			query = query.StartAfter(lastDoc)
		}
	}
	
	iter := query.Documents(ctx)
	defer iter.Stop()
	
	var messages []models.Message
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate messages: %w", err)
		}
		
		var message models.Message
		if err := doc.DataTo(&message); err != nil {
			continue
		}
		message.ID = doc.Ref.ID
		messages = append(messages, message)
	}
	
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	return messages, nil
}

func (r *ChatRepo) GetLastMessage(ctx context.Context, chatID string) (*models.Message, error) {
	iter := r.client.Collection("messages").
		Where("chatId", "==", chatID).
		OrderBy("timestamp", firestore.Desc).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()
	
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("no messages found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last message: %w", err)
	}
	
	var message models.Message
	if err := doc.DataTo(&message); err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}
	message.ID = doc.Ref.ID
	
	return &message, nil
}

func (r *ChatRepo) FindPrivateChat(ctx context.Context, user1ID, user2ID string) (*models.Chat, error) {
	iter := r.client.Collection("chat_members").Where("userId", "==", user1ID).Documents(ctx)
	defer iter.Stop()
	
	var user1ChatIDs []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate user1 chats: %w", err)
		}
		
		var member models.ChatMember
		if err := doc.DataTo(&member); err != nil {
			continue
		}
		user1ChatIDs = append(user1ChatIDs, member.ChatID)
	}
	
	for _, chatID := range user1ChatIDs {
		chat, err := r.GetChatByID(ctx, chatID)
		if err != nil || chat.Type != models.ChatTypePrivate {
			continue
		}
		
		isUser2InChat, err := r.IsUserInChat(ctx, chatID, user2ID)
		if err != nil {
			continue
		}
		
		if isUser2InChat {
			return chat, nil
		}
	}
	
	return nil, fmt.Errorf("private chat not found")
}

func (r *ChatRepo) UpdateChat(ctx context.Context, chatID string, updates map[string]interface{}) error {
	updates["updatedAt"] = time.Now()
	
	var firestoreUpdates []firestore.Update
	for key, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}
	
	_, err := r.client.Collection("chats").Doc(chatID).Update(ctx, firestoreUpdates)
	return err
}

func (r *ChatRepo) DeleteChat(ctx context.Context, chatID string) error {
	batch := r.client.Batch()
	
	chatRef := r.client.Collection("chats").Doc(chatID)
	batch.Delete(chatRef)
	
	membersIter := r.client.Collection("chat_members").Where("chatId", "==", chatID).Documents(ctx)
	defer membersIter.Stop()
	
	for {
		doc, err := membersIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate members for deletion: %w", err)
		}
		batch.Delete(doc.Ref)
	}
	
	messagesIter := r.client.Collection("messages").Where("chatId", "==", chatID).Documents(ctx)
	defer messagesIter.Stop()
	
	for {
		doc, err := messagesIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate messages for deletion: %w", err)
		}
		batch.Delete(doc.Ref)
	}
	
	_, err := batch.Commit(ctx)
	return err
}
