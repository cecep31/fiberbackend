package handler

import (
	"context"
	"errors"
	"strings"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

// MsgAIArticleUnavailable is the 503 message when OpenRouter is not configured.
const MsgAIArticleUnavailable = "AI article generation is not available"

type AIHandler struct {
	articleService service.ArticleGenerationService
}

func NewAIHandler(articleService service.ArticleGenerationService) *AIHandler {
	return &AIHandler{articleService: articleService}
}

// GenerateArticle streams an AI-written draft article as SSE: `ai_chunk`
// events, then exactly one of `ai_complete` ({content, model}) or `error`,
// then `data: [DONE]`. Nothing is persisted.
func (h *AIHandler) GenerateArticle(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok || userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req dto.GenerateArticleRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Failed to generate article", err)
	}
	// The producer goroutine keeps these strings after the handler returns,
	// so detach them from fasthttp's recycled request buffer.
	req.Prompt = strings.Clone(strings.TrimSpace(req.Prompt))
	req.Tone = strings.Clone(strings.TrimSpace(req.Tone))
	req.Length = strings.Clone(strings.TrimSpace(req.Length))
	req.Title = strings.Clone(strings.TrimSpace(req.Title))
	if err := validate.Validate(&req); err != nil {
		return response.FromValidateError(c, err)
	}

	if !h.articleService.Available() {
		return response.ServiceUnavailable(c, MsgAIArticleUnavailable)
	}

	// fiber's c.Context() is never cancelled, so the stream owns a
	// cancellable ctx; streamAIEvents cancels it when the writer ends
	// (including on client disconnect) to stop the upstream request.
	ctx, cancel := context.WithCancel(c.Context())
	chunks, complete, errCh, err := h.articleService.GenerateArticleStream(ctx, &req)
	if err != nil {
		cancel()
		if errors.Is(err, apperrors.ErrAIArticleGenerationUnavailable) {
			return response.ServiceUnavailable(c, MsgAIArticleUnavailable)
		}
		return response.InternalServerError(c, "Failed to generate article", err)
	}
	return streamAIEvents(c, cancel, nil, chunks, complete, errCh, "article generation stream failed")
}
