package handler

import (
	"bufio"
	"errors"
	"fmt"
	"time"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/applog"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

var guildChatLog = applog.Component("guild_chat")

const (
	// guildStreamHeartbeat keeps proxies from closing an idle stream and is
	// also how often a subscriber's membership is re-checked, so a member who
	// leaves stops receiving events within this interval.
	guildStreamHeartbeat = 25 * time.Second
	// guildStreamWriteTimeout bounds each write to the stream. fasthttp sets a
	// single absolute write deadline (the server-wide WriteTimeout) when the
	// response starts, so the stream refreshes it before every frame or it
	// would be cut WriteTimeout after it opened.
	guildStreamWriteTimeout = 10 * time.Second
)

type GuildChannelHandler struct {
	channelService service.GuildChannelService
	// streamHeartbeat is guildStreamHeartbeat; tests shorten it.
	streamHeartbeat time.Duration
}

func NewGuildChannelHandler(channelService service.GuildChannelService) *GuildChannelHandler {
	return &GuildChannelHandler{channelService: channelService, streamHeartbeat: guildStreamHeartbeat}
}

func (h *GuildChannelHandler) ListChannels(c fiber.Ctx) error {
	guildSlug := c.Params("slug")
	viewerID, _ := GetUserIDFromClaims(c)

	channels, err := h.channelService.ListChannels(c.Context(), guildSlug, viewerID)
	if err != nil {
		return handleGuildChannelError(c, "Failed to fetch channels", err)
	}
	return response.Success(c, "Channels fetched successfully", channels)
}

func (h *GuildChannelHandler) CreateChannel(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}

	var req dto.CreateGuildChannelRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", err)
	}
	if err := validate.Validate(&req); err != nil {
		return response.FromValidateError(c, err)
	}

	channel, err := h.channelService.CreateChannel(c.Context(), c.Params("slug"), userID, &req)
	if err != nil {
		return handleGuildChannelError(c, "Failed to create channel", err)
	}
	return response.Created(c, "Channel created successfully", channel)
}

func (h *GuildChannelHandler) UpdateChannel(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}

	var req dto.UpdateGuildChannelRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", err)
	}
	if err := validate.Validate(&req); err != nil {
		return response.FromValidateError(c, err)
	}

	channel, err := h.channelService.UpdateChannel(c.Context(), c.Params("slug"), c.Params("channelId"), userID, &req)
	if err != nil {
		return handleGuildChannelError(c, "Failed to update channel", err)
	}
	return response.Success(c, "Channel updated successfully", channel)
}

func (h *GuildChannelHandler) DeleteChannel(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}

	if err := h.channelService.DeleteChannel(c.Context(), c.Params("slug"), c.Params("channelId"), userID); err != nil {
		return handleGuildChannelError(c, "Failed to delete channel", err)
	}
	return response.Success(c, "Channel deleted successfully", nil)
}

// ListMessages returns channel history newest first, paginated backwards with
// the ?before=<message id> cursor rather than offset, so that messages
// arriving meanwhile do not shift the pages.
func (h *GuildChannelHandler) ListMessages(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}
	limit, _ := ParsePaginationParams(c, 50)

	messages, meta, err := h.channelService.ListMessages(
		c.Context(), c.Params("slug"), c.Params("channelId"), userID, c.Query("before"), limit)
	if err != nil {
		return handleGuildChannelError(c, "Failed to fetch messages", err)
	}
	return response.SuccessWithMeta(c, "Messages fetched successfully", messages, meta)
}

func (h *GuildChannelHandler) SendMessage(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}

	var req dto.CreateGuildMessageRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", err)
	}
	if err := validate.Validate(&req); err != nil {
		return response.FromValidateError(c, err)
	}

	message, err := h.channelService.SendMessage(c.Context(), c.Params("slug"), c.Params("channelId"), userID, &req)
	if err != nil {
		return handleGuildChannelError(c, "Failed to send message", err)
	}
	return response.Created(c, "Message sent successfully", message)
}

func (h *GuildChannelHandler) EditMessage(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}

	var req dto.UpdateGuildMessageRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", err)
	}
	if err := validate.Validate(&req); err != nil {
		return response.FromValidateError(c, err)
	}

	message, err := h.channelService.EditMessage(
		c.Context(), c.Params("slug"), c.Params("channelId"), c.Params("messageId"), userID, &req)
	if err != nil {
		return handleGuildChannelError(c, "Failed to edit message", err)
	}
	return response.Success(c, "Message updated successfully", message)
}

func (h *GuildChannelHandler) DeleteMessage(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}

	err := h.channelService.DeleteMessage(
		c.Context(), c.Params("slug"), c.Params("channelId"), c.Params("messageId"), userID)
	if err != nil {
		return handleGuildChannelError(c, "Failed to delete message", err)
	}
	return response.Success(c, "Message deleted successfully", nil)
}

// StreamEvents holds a Server-Sent Events stream open and forwards every
// realtime event of the guild (all channels) to a member. Each event is one
// `data:` line holding a dto.GuildEvent. The stream ends when the client
// disconnects, the member leaves the guild, the subscriber falls too far
// behind, or the server shuts down; clients should reconnect and refetch
// recent history to fill any gap.
func (h *GuildChannelHandler) StreamEvents(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}
	ctx := c.Context()

	guildID, sub, err := h.channelService.Subscribe(ctx, c.Params("slug"), userID)
	if err != nil {
		return handleGuildChannelError(c, "Failed to open event stream", err)
	}

	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set("Connection", "keep-alive")
	// Stop nginx-style proxies from buffering the stream.
	c.Set("X-Accel-Buffering", "no")

	// fasthttp sets one absolute write deadline (now + WriteTimeout) before it
	// writes the response and never refreshes it while copying the body
	// stream to the socket, so without intervention the stream would be cut
	// WriteTimeout (60s) after it opened. Like echobackend, every frame first
	// pushes the connection's write deadline guildStreamWriteTimeout ahead:
	// the writer below fills an in-memory pipe that the server goroutine
	// copies to the conn right after each Flush, so that copy runs under the
	// refreshed deadline. (The very first frame may race fasthttp's own
	// deadline, which still covers it.) The fiber ctx must not be used inside
	// the writer: the handler returns as soon as the writer is registered and
	// the ctx is recycled, which is also why the subscription is closed from
	// inside the writer rather than via defer in the handler.
	conn := c.RequestCtx().Conn()
	heartbeatEvery := h.streamHeartbeat
	if heartbeatEvery <= 0 {
		heartbeatEvery = guildStreamHeartbeat
	}
	c.RequestCtx().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer sub.Close()

		write := func(frame string) error {
			if conn != nil {
				if err := conn.SetWriteDeadline(time.Now().Add(guildStreamWriteTimeout)); err != nil {
					return err
				}
			}
			if _, err := w.WriteString(frame); err != nil {
				return err
			}
			return w.Flush()
		}

		if err := write(": connected\n\n"); err != nil {
			return
		}

		heartbeat := time.NewTicker(heartbeatEvery)
		defer heartbeat.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case payload, ok := <-sub.C:
				if !ok {
					return
				}
				if err := write(fmt.Sprintf("data: %s\n\n", payload)); err != nil {
					return
				}
			case <-heartbeat.C:
				member, err := h.channelService.IsMember(ctx, guildID, userID)
				if err != nil {
					guildChatLog.Warn("guild chat: membership re-check failed", "guild_id", guildID, "error", err)
				} else if !member {
					return
				}
				if err := write(": ping\n\n"); err != nil {
					return
				}
			}
		}
	})
	return nil
}

func handleGuildChannelError(c fiber.Ctx, message string, err error) error {
	switch {
	case errors.Is(err, apperrors.ErrGuildNotFound):
		return response.NotFound(c, "Guild not found", err)
	case errors.Is(err, apperrors.ErrGuildNotOwned):
		return response.Forbidden(c, "Access forbidden")
	case errors.Is(err, apperrors.ErrGuildChannelNotFound):
		return response.NotFound(c, "Channel not found", err)
	case errors.Is(err, apperrors.ErrGuildMessageNotFound):
		return response.NotFound(c, "Message not found", err)
	case errors.Is(err, apperrors.ErrNotGuildMember),
		errors.Is(err, apperrors.ErrGuildMessageNotOwned):
		return response.Forbidden(c, "Access forbidden")
	case errors.Is(err, apperrors.ErrGuildChannelNameExists):
		return response.Conflict(c, "Channel name already taken", err.Error())
	case errors.Is(err, apperrors.ErrGuildChannelNameInvalid),
		errors.Is(err, apperrors.ErrGuildMessageEmpty),
		errors.Is(err, apperrors.ErrGuildMessageReplyNotFound),
		errors.Is(err, apperrors.ErrInvalidMessageCursor):
		return response.BadRequest(c, err.Error(), err)
	default:
		return response.InternalServerError(c, message, err)
	}
}
