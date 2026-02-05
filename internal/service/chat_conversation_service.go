package service

import (
	"context"
	"errors"
	"fmt"

	"fiberbackend/internal/model"
	"fiberbackend/internal/repository"
)

type ChatConversationService interface {
	CreateConversation(ctx context.Context, userID string, conversation *model.CreateChatConversationDTO) (*model.ChatConversationResponse, error)
	GetConversationByID(ctx context.Context, id string, userID string) (*model.ChatConversationResponse, error)
	GetUserConversations(ctx context.Context, userID string, offset int, limit int) ([]*model.ChatConversationResponse, int64, error)
	UpdateConversation(ctx context.Context, id string, userID string, conversation *model.UpdateChatConversationDTO) (*model.ChatConversationResponse, error)
	DeleteConversation(ctx context.Context, id string, userID string) error
}

type chatConversationService struct {
	conversationRepo repository.ChatConversationRepository
}

func NewChatConversationService(conversationRepo repository.ChatConversationRepository) ChatConversationService {
	return &chatConversationService{
		conversationRepo: conversationRepo,
	}
}

func (s *chatConversationService) CreateConversation(ctx context.Context, userID string, conversation *model.CreateChatConversationDTO) (*model.ChatConversationResponse, error) {
	// Create the conversation entity
	chatConversation := &model.ChatConversation{
		Title:  conversation.Title,
		UserID: userID,
	}

	createdConversation, err := s.conversationRepo.CreateConversation(ctx, chatConversation)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return createdConversation.ToResponse(), nil
}

func (s *chatConversationService) GetConversationByID(ctx context.Context, id string, userID string) (*model.ChatConversationResponse, error) {
	conversation, err := s.conversationRepo.GetConversationByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation by id: %w", err)
	}

	// Check if the conversation belongs to the user
	if conversation.UserID != userID {
		return nil, errors.New("access denied: conversation does not belong to user")
	}

	return conversation.ToResponse(), nil
}

func (s *chatConversationService) GetUserConversations(ctx context.Context, userID string, offset int, limit int) ([]*model.ChatConversationResponse, int64, error) {
	conversations, total, err := s.conversationRepo.GetUserConversations(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user conversations: %w", err)
	}

	// Convert to response format
	var responses []*model.ChatConversationResponse
	for _, conversation := range conversations {
		responses = append(responses, conversation.ToResponse())
	}

	return responses, total, nil
}

func (s *chatConversationService) UpdateConversation(ctx context.Context, id string, userID string, conversation *model.UpdateChatConversationDTO) (*model.ChatConversationResponse, error) {
	// First, verify that the conversation belongs to the user
	existingConversation, err := s.conversationRepo.GetConversationByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation for update: %w", err)
	}

	if existingConversation.UserID != userID {
		return nil, errors.New("access denied: conversation does not belong to user")
	}

	updatedConversation, err := s.conversationRepo.UpdateConversation(ctx, id, conversation)
	if err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	return updatedConversation.ToResponse(), nil
}

func (s *chatConversationService) DeleteConversation(ctx context.Context, id string, userID string) error {
	// First, verify that the conversation belongs to the user
	existingConversation, err := s.conversationRepo.GetConversationByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get conversation for deletion: %w", err)
	}

	if existingConversation.UserID != userID {
		return errors.New("access denied: conversation does not belong to user")
	}

	if err := s.conversationRepo.DeleteConversation(ctx, id); err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}
	return nil
}
