package service

import (
	"encoding/json"
	"errors"
	"strings"

	domainchat "tanzanite/internal/domain/chat"
	"tanzanite/internal/repository"
)

var (
	ErrChatAuthenticationRequired = errors.New("authentication required")
	ErrChatSessionAccessDenied    = errors.New("chat session access denied")
	ErrChatMessageAlreadyExists   = errors.New("chat message already exists")
	ErrChatInvalidSession         = errors.New("session_id is required")
	ErrChatInvalidMessage         = errors.New("invalid chat message")
)

type ChatService struct {
	chatRepo *repository.ChatRepository
}

type ChatMessageInput struct {
	SessionID string
	MessageID string
	Content   string
	Role      string
	Timestamp int64
	AgentID   string
	Metadata  map[string]interface{}
}

func NewChatService(chatRepo *repository.ChatRepository) *ChatService {
	return &ChatService{chatRepo: chatRepo}
}

func (s *ChatService) SaveUserMessage(userID uint, input ChatMessageInput) (*domainchat.ChatMessage, error) {
	input.SessionID = strings.TrimSpace(input.SessionID)
	if input.SessionID == "" {
		return nil, ErrChatInvalidSession
	}
	if input.MessageID == "" || input.Content == "" || input.Timestamp == 0 || !isValidChatRole(input.Role) {
		return nil, ErrChatInvalidMessage
	}

	if err := s.EnsureSessionWritable(input.SessionID, &userID); err != nil {
		return nil, err
	}

	metadata, err := marshalChatMetadata(input.Metadata)
	if err != nil {
		return nil, ErrChatInvalidMessage
	}

	message := &domainchat.ChatMessage{
		SessionID: input.SessionID,
		MessageID: input.MessageID,
		Content:   input.Content,
		Role:      input.Role,
		Timestamp: input.Timestamp,
		UserID:    &userID,
		AgentID:   input.AgentID,
		Metadata:  metadata,
	}

	if err := s.chatRepo.SaveMessage(message); err != nil {
		if isDuplicateChatMessageError(err) {
			return nil, ErrChatMessageAlreadyExists
		}
		return nil, err
	}

	return message, nil
}

func (s *ChatService) GetMessages(sessionID string, userID uint, limit int) ([]domainchat.ChatMessage, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, ErrChatInvalidSession
	}

	belongs, err := s.SessionBelongsToUser(sessionID, userID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, ErrChatSessionAccessDenied
	}

	return s.chatRepo.GetMessages(sessionID, limit)
}

func (s *ChatService) GetUserMessages(userID uint, limit int) ([]domainchat.ChatMessage, error) {
	return s.chatRepo.GetMessagesByUser(userID, limit)
}

func (s *ChatService) EnsureSessionWritable(sessionID string, userID *uint) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return ErrChatInvalidSession
	}

	exists, err := s.chatRepo.SessionExists(sessionID)
	if err != nil {
		return err
	}
	if !exists {
		if userID != nil {
			return nil
		}
		return ErrChatAuthenticationRequired
	}

	if userID == nil {
		return ErrChatSessionAccessDenied
	}

	belongs, err := s.SessionBelongsToUser(sessionID, *userID)
	if err != nil {
		return err
	}
	if !belongs {
		return ErrChatSessionAccessDenied
	}

	return nil
}

func (s *ChatService) SessionBelongsToUser(sessionID string, userID uint) (bool, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return false, ErrChatInvalidSession
	}
	return s.chatRepo.SessionBelongsToUser(sessionID, userID)
}

func marshalChatMetadata(metadata map[string]interface{}) (string, error) {
	if len(metadata) == 0 {
		return "{}", nil
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func isValidChatRole(role string) bool {
	switch role {
	case "user", "agent", "system":
		return true
	default:
		return false
	}
}

func isDuplicateChatMessageError(err error) bool {
	if repository.IsDuplicatedKey(err) {
		return true
	}

	message := err.Error()
	return strings.Contains(message, "UNIQUE constraint failed: chat_messages.message_id") ||
		strings.Contains(message, "duplicate key value violates unique constraint") ||
		(strings.Contains(message, "Duplicate entry") && strings.Contains(message, "message_id"))
}
