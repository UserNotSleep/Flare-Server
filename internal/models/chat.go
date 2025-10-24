package models

import "time"

type ChatType string

const (
	ChatTypePrivate ChatType = "private"
	ChatTypeGroup   ChatType = "group"
)

type Chat struct {
	ID          string    `json:"id" firestore:"id"`
	Name        string    `json:"name" firestore:"name"`
	Type        ChatType  `json:"type" firestore:"type"`
	CreatedBy   string    `json:"createdBy" firestore:"createdBy"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" firestore:"updatedAt"`
	LastMessage *Message  `json:"lastMessage,omitempty" firestore:"-"`
	MemberCount int       `json:"memberCount" firestore:"memberCount"`
	Avatar      string    `json:"avatar,omitempty" firestore:"avatar"`
	Description string    `json:"description,omitempty" firestore:"description"`
}

type ChatMember struct {
	ID       string    `json:"id" firestore:"id"`
	ChatID   string    `json:"chatId" firestore:"chatId"`
	UserID   string    `json:"userId" firestore:"userId"`
	Username string    `json:"username" firestore:"username"`
	Role     MemberRole `json:"role" firestore:"role"`
	JoinedAt time.Time `json:"joinedAt" firestore:"joinedAt"`
}

type MemberRole string

const (
	RoleAdmin  MemberRole = "admin"
	RoleMember MemberRole = "member"
)

type Message struct {
	ID        string    `json:"id" firestore:"id"`
	ChatID    string    `json:"chatId" firestore:"chatId"`
	SenderID  string    `json:"senderId" firestore:"senderId"`
	Username  string    `json:"username" firestore:"username"`
	Text      string    `json:"text" firestore:"text"`
	Type      MessageType `json:"type" firestore:"type"`
	Timestamp time.Time `json:"timestamp" firestore:"timestamp"`
	EditedAt  *time.Time `json:"editedAt,omitempty" firestore:"editedAt"`
	ReplyTo   string    `json:"replyTo,omitempty" firestore:"replyTo"`
}

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeSystem MessageType = "system"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
)

type CreateChatRequest struct {
	Name        string   `json:"name"`
	Type        ChatType `json:"type"`
	Members     []string `json:"members,omitempty"`
	Description string   `json:"description,omitempty"`
}

type SendMessageRequest struct {
	Text    string `json:"text"`
	ReplyTo string `json:"replyTo,omitempty"`
}

type ChatListResponse struct {
	Chats []Chat `json:"chats"`
}

type ChatMessagesResponse struct {
	Messages []Message `json:"messages"`
	HasMore  bool      `json:"hasMore"`
}

type AddMemberRequest struct {
	Username string `json:"username"`
}

type ChatInfo struct {
	Chat    Chat         `json:"chat"`
	Members []ChatMember `json:"members"`
}
