package handler

import (
	"bufio"
	"encoding/json"
	"errors"
	"strings"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/applog"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

var chatLog = applog.Component("chat")

type ChatConversationHandler struct {
	chatConversationService service.ChatConversationService
}

func NewChatConversationHandler(chatConversationService service.ChatConversationService) *ChatConversationHandler {
	return &ChatConversationHandler{
		chatConversationService: chatConversationService,
	}
}

func (h *ChatConversationHandler) CreateConversation(c fiber.Ctx) error {
	var conversationReq dto.CreateChatConversationRequest
	if err := c.Bind().Body(&conversationReq); err != nil {
		return response.BadRequest(c, "Failed to create conversation", err)
	}

	if err := bindValidate(c, &conversationReq); err != nil {
		return err
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	newConversation, err := h.chatConversationService.CreateConversation(c.Context(), userID, &conversationReq)
	if err != nil {
		return response.InternalServerError(c, "Failed to create conversation", err)
	}

	return response.Created(c, "Successfully created conversation", newConversation)
}

func (h *ChatConversationHandler) CreateConversationStream(c fiber.Ctx) error {
	var req dto.CreateChatConversationStreamRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Failed to create conversation stream", err)
	}
	if err := bindValidate(c, &req); err != nil {
		return err
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	result, chunks, complete, errCh, err := h.chatConversationService.CreateConversationStream(c.Context(), userID, &req)
	if err != nil {
		return h.respondChatError(c, "Failed to create conversation stream", err)
	}
	if chunks == nil {
		return response.Created(c, "Message created successfully", []any{result.UserMessage})
	}
	return streamChatEvents(c, map[string]any{
		"type": "conversation_created",
		"data": map[string]any{
			"conversation_id": result.ConversationID,
			"user_message":    result.UserMessage,
		},
	}, chunks, complete, errCh)
}

func (h *ChatConversationHandler) GetConversation(c fiber.Ctx) error {
	id := c.Params("id")

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	conversation, err := h.chatConversationService.GetConversationByID(c.Context(), id, userID)
	if err != nil {
		return h.respondChatError(c, "Failed to get conversation", err)
	}

	return response.Success(c, "Successfully retrieved conversation", conversation)
}

func (h *ChatConversationHandler) GetConversations(c fiber.Ctx) error {
	limit, offset := ParsePaginationParams(c, 10)

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	conversations, total, err := h.chatConversationService.GetUserConversations(c.Context(), userID, offset, limit)
	if err != nil {
		return h.respondChatError(c, "Failed to get conversations", err)
	}

	meta := response.CalculatePaginationMeta(total, offset, limit)
	return response.SuccessWithMeta(c, "Successfully retrieved conversations", conversations, meta)
}

func (h *ChatConversationHandler) UpdateConversation(c fiber.Ctx) error {
	id := c.Params("id")
	var updateDTO dto.UpdateChatConversationRequest
	if err := c.Bind().Body(&updateDTO); err != nil {
		return response.BadRequest(c, "Failed to update conversation", err)
	}

	if err := bindValidate(c, &updateDTO); err != nil {
		return err
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	updatedConversation, err := h.chatConversationService.UpdateConversation(c.Context(), id, userID, &updateDTO)
	if err != nil {
		return h.respondChatError(c, "Failed to update conversation", err)
	}

	return response.Success(c, "Conversation updated successfully", updatedConversation)
}

func (h *ChatConversationHandler) DeleteConversation(c fiber.Ctx) error {
	id := c.Params("id")

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	err := h.chatConversationService.DeleteConversation(c.Context(), id, userID)
	if err != nil {
		return h.respondChatError(c, "Failed to delete conversation", err)
	}

	return response.Success(c, "Successfully deleted conversation", nil)
}

func (h *ChatConversationHandler) CreateMessage(c fiber.Ctx) error {
	conversationID := c.Params("conversationId")
	if conversationID == "" {
		return response.BadRequest(c, "Conversation ID is required", nil)
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req dto.CreateChatMessageRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Failed to create message", err)
	}
	if err := bindValidate(c, &req); err != nil {
		return err
	}

	messages, err := h.chatConversationService.CreateMessage(c.Context(), userID, conversationID, &req)
	if err != nil {
		return h.respondChatError(c, "Failed to create message", err)
	}
	return response.Created(c, "Messages created successfully", messages)
}

func (h *ChatConversationHandler) CreateMessageStream(c fiber.Ctx) error {
	conversationID := c.Params("conversationId")
	if conversationID == "" {
		return response.BadRequest(c, "Conversation ID is required", nil)
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req dto.CreateChatMessageRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Failed to create message stream", err)
	}
	if err := bindValidate(c, &req); err != nil {
		return err
	}

	result, chunks, complete, errCh, err := h.chatConversationService.CreateStreamingMessage(c.Context(), userID, conversationID, &req)
	if err != nil {
		return h.respondChatError(c, "Failed to create message stream", err)
	}
	if chunks == nil {
		return response.Created(c, "Message created successfully", []any{result.UserMessage})
	}
	return streamChatEvents(c, map[string]any{
		"type": "user_message",
		"data": result.UserMessage,
	}, chunks, complete, errCh)
}

func (h *ChatConversationHandler) GetMessages(c fiber.Ctx) error {
	conversationID := c.Params("conversationId")
	if conversationID == "" {
		return response.BadRequest(c, "Conversation ID is required", nil)
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	messages, err := h.chatConversationService.GetMessages(c.Context(), conversationID, userID)
	if err != nil {
		return h.respondChatError(c, "Failed to get messages", err)
	}
	return response.Success(c, "Messages fetched successfully", messages)
}

func (h *ChatConversationHandler) GetMessage(c fiber.Ctx) error {
	id := c.Params("messageId")
	if id == "" {
		return response.BadRequest(c, "Message ID is required", nil)
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	message, err := h.chatConversationService.GetMessage(c.Context(), id, userID)
	if err != nil {
		return h.respondChatError(c, "Failed to get message", err)
	}
	return response.Success(c, "Message fetched successfully", message)
}

func (h *ChatConversationHandler) DeleteMessage(c fiber.Ctx) error {
	id := c.Params("messageId")
	if id == "" {
		return response.BadRequest(c, "Message ID is required", nil)
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	message, err := h.chatConversationService.DeleteMessage(c.Context(), id, userID)
	if err != nil {
		return h.respondChatError(c, "Failed to delete message", err)
	}
	return response.Success(c, "Message deleted successfully", message)
}

func (h *ChatConversationHandler) respondChatError(c fiber.Ctx, message string, err error) error {
	switch {
	case errors.Is(err, apperrors.ErrChatConversationNotFound), errors.Is(err, apperrors.ErrChatMessageNotFound):
		return response.NotFound(c, message, err)
	case errors.Is(err, apperrors.ErrConversationNotOwned):
		return response.Forbidden(c, err.Error())
	default:
		return response.InternalServerError(c, message, err)
	}
}

func streamChatEvents(c fiber.Ctx, initialEvent map[string]any, chunks <-chan string, complete <-chan dto.ChatMessageResponse, errCh <-chan error) error {
	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set("Connection", "keep-alive")

	c.RequestCtx().SetBodyStreamWriter(func(w *bufio.Writer) {
		if initialEvent != nil {
			if err := writeSSE(w, initialEvent); err != nil {
				return
			}
			_ = w.Flush()
		}

		for chunk := range chunks {
			if err := writeSSE(w, map[string]any{
				"type": "ai_chunk",
				"data": chunk,
			}); err != nil {
				return
			}
			_ = w.Flush()
		}

		if err, ok := <-errCh; ok && err != nil {
			chatLog.Error("chat stream failed", "error", err)
			_ = writeSSE(w, map[string]any{
				"type": "error",
				"data": streamErrorMessage(err),
			})
			_ = w.Flush()
			return
		}

		if msg, ok := <-complete; ok {
			if err := writeSSE(w, map[string]any{
				"type": "ai_complete",
				"data": msg,
			}); err != nil {
				return
			}
			_ = w.Flush()
		}

		_, _ = w.WriteString("data: [DONE]\n\n")
		_ = w.Flush()
	})
	return nil
}

func writeSSE(w *bufio.Writer, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = w.WriteString("data: " + string(data) + "\n\n")
	return err
}

func streamErrorMessage(err error) string {
	if err == nil {
		return "Failed to generate AI response"
	}
	msg := err.Error()
	if idx := strings.Index(msg, `{"error"`); idx >= 0 {
		var payload struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal([]byte(msg[idx:]), &payload) == nil && payload.Error.Message != "" {
			return payload.Error.Message
		}
	}
	return msg
}
