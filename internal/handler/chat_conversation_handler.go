package handler

import (
	"fiberbackend/internal/model"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type ChatConversationHandler struct {
	chatConversationService service.ChatConversationService
}

func NewChatConversationHandler(chatConversationService service.ChatConversationService) *ChatConversationHandler {
	return &ChatConversationHandler{
		chatConversationService: chatConversationService,
	}
}

func (h *ChatConversationHandler) CreateConversation(c fiber.Ctx) error {
	var conversationReq model.CreateChatConversationDTO
	if err := c.Bind().Body(&conversationReq); err != nil {
		return response.HandleBindError(c, err)
	}

	// Get the user ID from the JWT token
	claims, _ := c.Locals("user").(jwt.MapClaims)
	userID := fmt.Sprintf("%v", claims["user_id"])

	newConversation, err := h.chatConversationService.CreateConversation(c.Context(), userID, &conversationReq)
	if err != nil {
		return response.InternalServerError(c, "Failed to create conversation", err)
	}

	return response.Created(c, "Successfully created conversation", newConversation)
}

func (h *ChatConversationHandler) GetConversation(c fiber.Ctx) error {
	id := c.Params("id")

	// Get the user ID from the JWT token
	claims, _ := c.Locals("user").(jwt.MapClaims)
	userID := fmt.Sprintf("%v", claims["user_id"])

	conversation, err := h.chatConversationService.GetConversationByID(c.Context(), id, userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to get conversation", err)
	}

	return response.Success(c, "Successfully retrieved conversation", conversation)
}

func (h *ChatConversationHandler) GetConversations(c fiber.Ctx) error {
	offset := c.Query("offset")
	limit := c.Query("limit")

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		offsetInt = 0 // Default offset if not provided or invalid
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10 // Default limit if not provided or invalid
	}

	// Get the user ID from the JWT token
	claims, _ := c.Locals("user").(jwt.MapClaims)
	userID := fmt.Sprintf("%v", claims["user_id"])

	conversations, total, err := h.chatConversationService.GetUserConversations(c.Context(), userID, offsetInt, limitInt)
	if err != nil {
		return response.InternalServerError(c, "Failed to get conversations", err)
	}

	meta := response.PaginationMeta{
		TotalItems: int(total),
		Offset:     offsetInt,
		Limit:      limitInt,
		TotalPages: int(total)/limitInt + 1,
	}
	if int(total)%limitInt == 0 {
		meta.TotalPages = int(total) / limitInt
	}

	return response.SuccessWithMeta(c, "Successfully retrieved conversations", conversations, meta)
}

func (h *ChatConversationHandler) UpdateConversation(c fiber.Ctx) error {
	id := c.Params("id")
	var updateDTO model.UpdateChatConversationDTO
	if err := c.Bind().Body(&updateDTO); err != nil {
		return response.HandleBindError(c, err)
	}

	// Get the user ID from the JWT token
	claims, _ := c.Locals("user").(jwt.MapClaims)
	userID := fmt.Sprintf("%v", claims["user_id"])

	updatedConversation, err := h.chatConversationService.UpdateConversation(c.Context(), id, userID, &updateDTO)
	if err != nil {
		return response.InternalServerError(c, "Failed to update conversation", err)
	}

	return response.Success(c, "Conversation updated successfully", updatedConversation)
}

func (h *ChatConversationHandler) DeleteConversation(c fiber.Ctx) error {
	id := c.Params("id")

	// Get the user ID from the JWT token
	claims, _ := c.Locals("user").(jwt.MapClaims)
	userID := fmt.Sprintf("%v", claims["user_id"])

	err := h.chatConversationService.DeleteConversation(c.Context(), id, userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to delete conversation", err)
	}

	return response.Success(c, "Successfully deleted conversation", nil)
}
