package service

import (
	"context"
	"fmt"
	"strings"

	"Flare-server/internal/models"
	"Flare-server/internal/repository"
)

type ChatService struct {
	chatRepo *repository.ChatRepo
	userRepo *repository.UserRepo
}

func NewChatService(chatRepo *repository.ChatRepo, userRepo *repository.UserRepo) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		userRepo: userRepo,
	}
}

func (s *ChatService) CreateChat(ctx context.Context, req models.CreateChatRequest, creatorID, creatorUsername string) (*models.Chat, error) {
	if req.Type == models.ChatTypeGroup && strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("group chat name is required")
	}

	if req.Type == models.ChatTypePrivate && len(req.Members) != 1 {
		return nil, fmt.Errorf("private chat must have exactly one other member")
	}

	if req.Type == models.ChatTypePrivate {
		otherUsername := req.Members[0]
		otherUser, err := s.userRepo.GetUserByUsername(ctx, otherUsername)
		if err != nil {
			return nil, fmt.Errorf("user %s not found", otherUsername)
		}

		existingChat, err := s.chatRepo.FindPrivateChat(ctx, creatorID, otherUser.ID)
		if err == nil {
			return existingChat, nil
		}
	}

	chat := models.Chat{
		Name:        req.Name,
		Type:        req.Type,
		CreatedBy:   creatorID,
		Description: req.Description,
		MemberCount: 1,
	}

	if req.Type == models.ChatTypePrivate {
		chat.Name = ""
		chat.MemberCount = 2
	}

	createdChat, err := s.chatRepo.CreateChat(ctx, chat)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	creatorMember := models.ChatMember{
		ChatID:   createdChat.ID,
		UserID:   creatorID,
		Username: creatorUsername,
		Role:     models.RoleAdmin,
	}

	if err := s.chatRepo.AddChatMember(ctx, creatorMember); err != nil {
		return nil, fmt.Errorf("failed to add creator to chat: %w", err)
	}

	for _, username := range req.Members {
		user, err := s.userRepo.GetUserByUsername(ctx, username)
		if err != nil {
			continue
		}

		member := models.ChatMember{
			ChatID:   createdChat.ID,
			UserID:   user.ID,
			Username: user.Username,
			Role:     models.RoleMember,
		}

		if err := s.chatRepo.AddChatMember(ctx, member); err != nil {
			continue
		}

		if req.Type == models.ChatTypeGroup {
			systemMsg := models.Message{
				ChatID:   createdChat.ID,
				SenderID: "system",
				Username: "System",
				Text:     fmt.Sprintf("%s добавлен в чат", username),
				Type:     models.MessageTypeSystem,
			}
			s.chatRepo.SaveMessage(ctx, systemMsg)
		}
	}

	return createdChat, nil
}

func (s *ChatService) GetUserChats(ctx context.Context, userID string) ([]models.Chat, error) {
	return s.chatRepo.GetUserChats(ctx, userID)
}

func (s *ChatService) GetChatInfo(ctx context.Context, chatID, userID string) (*models.ChatInfo, error) {
	isMember, err := s.chatRepo.IsUserInChat(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check chat membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("access denied: user is not a member of this chat")
	}

	chat, err := s.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	members, err := s.chatRepo.GetChatMembers(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat members: %w", err)
	}

	return &models.ChatInfo{
		Chat:    *chat,
		Members: members,
	}, nil
}

func (s *ChatService) SendMessage(ctx context.Context, chatID, senderID, username string, req models.SendMessageRequest) (*models.Message, error) {
	isMember, err := s.chatRepo.IsUserInChat(ctx, chatID, senderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check chat membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("access denied: user is not a member of this chat")
	}

	if strings.TrimSpace(req.Text) == "" {
		return nil, fmt.Errorf("message text cannot be empty")
	}

	message := models.Message{
		ChatID:   chatID,
		SenderID: senderID,
		Username: username,
		Text:     strings.TrimSpace(req.Text),
		Type:     models.MessageTypeText,
		ReplyTo:  req.ReplyTo,
	}

	savedMessage, err := s.chatRepo.SaveMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	return savedMessage, nil
}

func (s *ChatService) GetChatMessages(ctx context.Context, chatID, userID string, limit int, lastMessageID string) (*models.ChatMessagesResponse, error) {
	isMember, err := s.chatRepo.IsUserInChat(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check chat membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("access denied: user is not a member of this chat")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	messages, err := s.chatRepo.GetChatMessages(ctx, chatID, limit+1, lastMessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	return &models.ChatMessagesResponse{
		Messages: messages,
		HasMore:  hasMore,
	}, nil
}

func (s *ChatService) AddMemberToChat(ctx context.Context, chatID, adminID string, req models.AddMemberRequest) error {
	isAdmin, err := s.isUserAdmin(ctx, chatID, adminID)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}
	if !isAdmin {
		return fmt.Errorf("access denied: only admins can add members")
	}

	chat, err := s.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	if chat.Type != models.ChatTypeGroup {
		return fmt.Errorf("can only add members to group chats")
	}

	user, err := s.userRepo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return fmt.Errorf("user %s not found", req.Username)
	}

	isMember, err := s.chatRepo.IsUserInChat(ctx, chatID, user.ID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return fmt.Errorf("user is already a member of this chat")
	}

	member := models.ChatMember{
		ChatID:   chatID,
		UserID:   user.ID,
		Username: user.Username,
		Role:     models.RoleMember,
	}

	if err := s.chatRepo.AddChatMember(ctx, member); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	systemMsg := models.Message{
		ChatID:   chatID,
		SenderID: "system",
		Username: "System",
		Text:     fmt.Sprintf("%s добавлен в чат", user.Username),
		Type:     models.MessageTypeSystem,
	}
	s.chatRepo.SaveMessage(ctx, systemMsg)

	return nil
}

func (s *ChatService) RemoveMemberFromChat(ctx context.Context, chatID, adminID, targetUserID string) error {
	isAdmin, err := s.isUserAdmin(ctx, chatID, adminID)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}
	if !isAdmin {
		return fmt.Errorf("access denied: only admins can remove members")
	}

	chat, err := s.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	if chat.Type != models.ChatTypeGroup {
		return fmt.Errorf("can only remove members from group chats")
	}

	if targetUserID == chat.CreatedBy {
		return fmt.Errorf("cannot remove chat creator")
	}

	members, err := s.chatRepo.GetChatMembers(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat members: %w", err)
	}

	var targetUsername string
	for _, member := range members {
		if member.UserID == targetUserID {
			targetUsername = member.Username
			break
		}
	}

	if targetUsername == "" {
		return fmt.Errorf("user is not a member of this chat")
	}

	if err := s.chatRepo.RemoveChatMember(ctx, chatID, targetUserID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	systemMsg := models.Message{
		ChatID:   chatID,
		SenderID: "system",
		Username: "System",
		Text:     fmt.Sprintf("%s удален из чата", targetUsername),
		Type:     models.MessageTypeSystem,
	}
	s.chatRepo.SaveMessage(ctx, systemMsg)

	return nil
}

func (s *ChatService) LeaveChat(ctx context.Context, chatID, userID string) error {
	chat, err := s.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	if chat.Type == models.ChatTypeGroup && userID == chat.CreatedBy {
		return fmt.Errorf("chat creator cannot leave the chat")
	}

	members, err := s.chatRepo.GetChatMembers(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat members: %w", err)
	}

	var username string
	for _, member := range members {
		if member.UserID == userID {
			username = member.Username
			break
		}
	}

	if username == "" {
		return fmt.Errorf("user is not a member of this chat")
	}

	if err := s.chatRepo.RemoveChatMember(ctx, chatID, userID); err != nil {
		return fmt.Errorf("failed to leave chat: %w", err)
	}

	if chat.Type == models.ChatTypeGroup {
		systemMsg := models.Message{
			ChatID:   chatID,
			SenderID: "system",
			Username: "System",
			Text:     fmt.Sprintf("%s покинул чат", username),
			Type:     models.MessageTypeSystem,
		}
		s.chatRepo.SaveMessage(ctx, systemMsg)
	}

	return nil
}

func (s *ChatService) UpdateChat(ctx context.Context, chatID, userID string, updates map[string]interface{}) error {
	isAdmin, err := s.isUserAdmin(ctx, chatID, userID)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}
	if !isAdmin {
		return fmt.Errorf("access denied: only admins can update chat")
	}

	allowedFields := map[string]bool{
		"name":        true,
		"description": true,
		"avatar":      true,
	}

	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	return s.chatRepo.UpdateChat(ctx, chatID, filteredUpdates)
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID, userID string) error {
	chat, err := s.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	if userID != chat.CreatedBy {
		return fmt.Errorf("access denied: only chat creator can delete the chat")
	}

	return s.chatRepo.DeleteChat(ctx, chatID)
}

func (s *ChatService) isUserAdmin(ctx context.Context, chatID, userID string) (bool, error) {
	members, err := s.chatRepo.GetChatMembers(ctx, chatID)
	if err != nil {
		return false, err
	}

	for _, member := range members {
		if member.UserID == userID && member.Role == models.RoleAdmin {
			return true, nil
		}
	}

	return false, nil
}

func (s *ChatService) IsUserInChat(ctx context.Context, chatID, userID string) (bool, error) {
	return s.chatRepo.IsUserInChat(ctx, chatID, userID)
}
