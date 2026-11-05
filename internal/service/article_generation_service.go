package service

import (
	"context"
	"strconv"
	"strings"

	"fiberbackend/config"
	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/platform/openrouter"
)

const (
	// articleGenerationTemperature is fixed: the endpoint takes no model or
	// temperature from the client.
	articleGenerationTemperature = 0.7

	defaultArticleTone   = "informative"
	defaultArticleLength = "medium"

	articleSystemPrompt = `You are an expert blog writer for pilput, a publishing platform for writers. Write a complete, original, well-structured blog article based on the user's request.

Output rules:
- Respond with HTML only. Do not use Markdown and do not wrap the answer in code fences.
- Do not include <html>, <head>, <body>, <style> or <script> tags, images, or inline styles.
- Start with the article title in a single <h1> element.
- Write the body using only these tags: <h2>, <h3>, <p>, <ul>, <ol>, <li>, <strong>, <em>, <blockquote>, <code>, <pre>, <a href="...">.
- Write in the same language as the user's request.
- Do not add any commentary before or after the article.`
)

var articleToneDescriptions = map[string]string{
	"informative":  "informative and clear",
	"professional": "professional and authoritative",
	"casual":       "casual and friendly",
	"storytelling": "narrative, storytelling style",
}

var articleLengthWords = map[string]int{
	"short":  400,
	"medium": 800,
	"long":   1500,
}

// ArticleGenerationService streams an AI-written draft article for the
// create/edit-post page. It is stateless: nothing is read from or written to
// the database.
type ArticleGenerationService interface {
	// Available reports whether an article can be generated: an OpenRouter
	// client is wired, OPENROUTER_API_KEY is non-blank and
	// OPENROUTER_DEFAULT_MODEL is set.
	Available() bool
	// GenerateArticleStream starts the upstream stream. The chunk channel
	// yields text deltas; once it closes, exactly one of complete or errCh
	// carries a value. Cancelling ctx aborts the upstream request. It returns
	// ErrAIArticleGenerationUnavailable (and no channels) when !Available().
	GenerateArticleStream(ctx context.Context, req *dto.GenerateArticleRequest) (<-chan string, <-chan dto.GeneratedArticle, <-chan error, error)
}

type articleGenerationService struct {
	openRouter openrouter.Client
	config     *config.Config
}

func NewArticleGenerationService(openRouter openrouter.Client, cfg *config.Config) ArticleGenerationService {
	return &articleGenerationService{
		openRouter: openRouter,
		config:     cfg,
	}
}

func (s *articleGenerationService) Available() bool {
	return s.openRouter != nil && s.config != nil &&
		strings.TrimSpace(s.config.OpenRouter.APIKey) != "" &&
		s.defaultModel() != ""
}

func (s *articleGenerationService) defaultModel() string {
	if s.config == nil {
		return ""
	}
	return strings.TrimSpace(s.config.OpenRouter.DefaultModel)
}

func (s *articleGenerationService) GenerateArticleStream(ctx context.Context, req *dto.GenerateArticleRequest) (<-chan string, <-chan dto.GeneratedArticle, <-chan error, error) {
	if !s.Available() {
		return nil, nil, nil, apperrors.ErrAIArticleGenerationUnavailable
	}

	model := s.defaultModel()
	upstreamChunks, usageCh, upstreamErrCh := s.openRouter.GenerateStream(ctx, buildArticleMessages(req), &model, articleGenerationTemperature)
	outChunks := make(chan string)
	complete := make(chan dto.GeneratedArticle, 1)
	errCh := make(chan error, 1)

	go func() {
		defer close(outChunks)
		defer close(complete)
		defer close(errCh)

		var content strings.Builder
		for chunk := range upstreamChunks {
			content.WriteString(chunk)
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case outChunks <- chunk:
			}
		}

		if err, ok := <-upstreamErrCh; ok && err != nil {
			errCh <- err
			return
		}
		// Drain usage so the upstream goroutine never blocks on it; token
		// counts are not reported for generated articles.
		<-usageCh

		if content.Len() == 0 {
			errCh <- apperrors.ErrAIEmptyResponse
			return
		}
		complete <- dto.GeneratedArticle{Content: content.String(), Model: model}
	}()

	return outChunks, complete, errCh, nil
}

// buildArticleMessages returns the system + user messages sent to OpenRouter.
func buildArticleMessages(req *dto.GenerateArticleRequest) []openrouter.Message {
	return []openrouter.Message{
		{Role: "system", Content: articleSystemPrompt},
		{Role: "user", Content: buildArticleUserPrompt(req)},
	}
}

// buildArticleUserPrompt renders the user message. Blank or unknown tone and
// length fall back to the defaults; the working-title line is added only for
// a non-blank title.
func buildArticleUserPrompt(req *dto.GenerateArticleRequest) string {
	tone, ok := articleToneDescriptions[strings.TrimSpace(req.Tone)]
	if !ok {
		tone = articleToneDescriptions[defaultArticleTone]
	}
	words, ok := articleLengthWords[strings.TrimSpace(req.Length)]
	if !ok {
		words = articleLengthWords[defaultArticleLength]
	}

	lines := []string{
		"Write a blog article about: " + strings.TrimSpace(req.Prompt),
		"Tone: " + tone,
		"Length: about " + strconv.Itoa(words) + " words",
	}
	if title := strings.TrimSpace(req.Title); title != "" {
		lines = append(lines, "Working title: "+title)
	}
	return strings.Join(lines, "\n")
}
