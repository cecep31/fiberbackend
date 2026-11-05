package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"fiberbackend/config"
	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/platform/openrouter"
)

func articleTestConfig() *config.Config {
	return &config.Config{OpenRouter: config.OpenRouterConfig{APIKey: "test-key", DefaultModel: " test-model "}}
}

// fakeStream mimics openrouter.Client.GenerateStream: it delivers chunks,
// then err (if any) or a usage value, and closes every channel.
func fakeStream(chunks []string, streamErr error) (<-chan string, <-chan openrouter.Usage, <-chan error) {
	chunkCh := make(chan string, len(chunks))
	for _, c := range chunks {
		chunkCh <- c
	}
	close(chunkCh)
	usageCh := make(chan openrouter.Usage, 1)
	errCh := make(chan error, 1)
	if streamErr != nil {
		errCh <- streamErr
	} else {
		usageCh <- openrouter.Usage{TotalTokens: 3}
	}
	close(usageCh)
	close(errCh)
	return chunkCh, usageCh, errCh
}

// drainArticle collects a GenerateArticleStream result with a timeout.
func drainArticle(t *testing.T, chunks <-chan string, complete <-chan dto.GeneratedArticle, errCh <-chan error) ([]string, *dto.GeneratedArticle, error) {
	t.Helper()
	done := make(chan struct{})
	var got []string
	var article *dto.GeneratedArticle
	var streamErr error
	go func() {
		defer close(done)
		for c := range chunks {
			got = append(got, c)
		}
		if err, ok := <-errCh; ok {
			streamErr = err
		}
		if a, ok := <-complete; ok {
			article = &a
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("article stream did not finish")
	}
	return got, article, streamErr
}

func TestBuildArticleUserPrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  dto.GenerateArticleRequest
		want string
	}{
		{
			name: "defaults without title",
			req:  dto.GenerateArticleRequest{Prompt: "  Go generics  "},
			want: "Write a blog article about: Go generics\nTone: informative and clear\nLength: about 800 words",
		},
		{
			name: "blank title is ignored",
			req:  dto.GenerateArticleRequest{Prompt: "Go generics", Tone: "casual", Length: "short", Title: "   "},
			want: "Write a blog article about: Go generics\nTone: casual and friendly\nLength: about 400 words",
		},
		{
			name: "with title",
			req:  dto.GenerateArticleRequest{Prompt: "Go generics", Tone: "storytelling", Length: "long", Title: " My Title "},
			want: "Write a blog article about: Go generics\nTone: narrative, storytelling style\nLength: about 1500 words\nWorking title: My Title",
		},
		{
			name: "professional medium",
			req:  dto.GenerateArticleRequest{Prompt: "Tax tips", Tone: "professional", Length: "medium"},
			want: "Write a blog article about: Tax tips\nTone: professional and authoritative\nLength: about 800 words",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := buildArticleUserPrompt(&tt.req); got != tt.want {
				t.Fatalf("got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestBuildArticleMessages(t *testing.T) {
	t.Parallel()

	msgs := buildArticleMessages(&dto.GenerateArticleRequest{Prompt: "x y z"})
	if len(msgs) != 2 || msgs[0].Role != "system" || msgs[1].Role != "user" {
		t.Fatalf("messages = %+v", msgs)
	}
	if !strings.HasPrefix(msgs[0].Content, "You are an expert blog writer for pilput") ||
		!strings.HasSuffix(msgs[0].Content, "- Do not add any commentary before or after the article.") ||
		!strings.Contains(msgs[0].Content, "request.\n\nOutput rules:\n- Respond with HTML only.") {
		t.Fatalf("unexpected system prompt:\n%s", msgs[0].Content)
	}
}

func TestArticleGenerationAvailable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		client openrouter.Client
		cfg    *config.Config
		want   bool
	}{
		{"configured", &mockOpenRouterService{}, articleTestConfig(), true},
		{"nil client", nil, articleTestConfig(), false},
		{"nil config", &mockOpenRouterService{}, nil, false},
		{"blank key", &mockOpenRouterService{}, &config.Config{OpenRouter: config.OpenRouterConfig{APIKey: "  ", DefaultModel: "m"}}, false},
		{"no default model", &mockOpenRouterService{}, &config.Config{OpenRouter: config.OpenRouterConfig{APIKey: "k", DefaultModel: " "}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewArticleGenerationService(tt.client, tt.cfg).Available(); got != tt.want {
				t.Fatalf("Available() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateArticleStreamUnavailableSkipsOpenRouter(t *testing.T) {
	t.Parallel()

	openRouter := &mockOpenRouterService{
		generateStreamFn: func(context.Context, []openrouter.Message, *string, float64) (<-chan string, <-chan openrouter.Usage, <-chan error) {
			t.Fatal("OpenRouter must not be called when AI is unavailable")
			return nil, nil, nil
		},
	}
	svc := NewArticleGenerationService(openRouter, &config.Config{OpenRouter: config.OpenRouterConfig{DefaultModel: "m"}})
	chunks, complete, errCh, err := svc.GenerateArticleStream(context.Background(), &dto.GenerateArticleRequest{Prompt: "abc"})
	if !errors.Is(err, apperrors.ErrAIArticleGenerationUnavailable) || chunks != nil || complete != nil || errCh != nil {
		t.Fatalf("got err=%v, want ErrAIArticleGenerationUnavailable and no channels", err)
	}
}

func TestGenerateArticleStreamCompletes(t *testing.T) {
	t.Parallel()

	var gotMessages []openrouter.Message
	var gotModel string
	var gotTemp float64
	openRouter := &mockOpenRouterService{
		generateStreamFn: func(_ context.Context, messages []openrouter.Message, model *string, temperature float64) (<-chan string, <-chan openrouter.Usage, <-chan error) {
			gotMessages = messages
			gotModel = *model
			gotTemp = temperature
			return fakeStream([]string{"<h1>Hi</h1>", "<p>there</p>"}, nil)
		},
	}
	svc := NewArticleGenerationService(openRouter, articleTestConfig())
	chunks, complete, errCh, err := svc.GenerateArticleStream(context.Background(), &dto.GenerateArticleRequest{Prompt: "greetings", Title: "Hello"})
	if err != nil {
		t.Fatalf("GenerateArticleStream: %v", err)
	}
	got, article, streamErr := drainArticle(t, chunks, complete, errCh)
	if streamErr != nil {
		t.Fatalf("stream error: %v", streamErr)
	}
	if strings.Join(got, "|") != "<h1>Hi</h1>|<p>there</p>" {
		t.Fatalf("chunks = %q", got)
	}
	if article == nil || article.Content != "<h1>Hi</h1><p>there</p>" || article.Model != "test-model" {
		t.Fatalf("article = %+v", article)
	}
	if gotModel != "test-model" || gotTemp != 0.7 {
		t.Fatalf("model=%q temperature=%v, want test-model 0.7", gotModel, gotTemp)
	}
	if len(gotMessages) != 2 || gotMessages[0].Content != articleSystemPrompt ||
		gotMessages[1].Content != "Write a blog article about: greetings\nTone: informative and clear\nLength: about 800 words\nWorking title: Hello" {
		t.Fatalf("messages = %+v", gotMessages)
	}
}

func TestGenerateArticleStreamEmptyResponse(t *testing.T) {
	t.Parallel()

	openRouter := &mockOpenRouterService{
		generateStreamFn: func(context.Context, []openrouter.Message, *string, float64) (<-chan string, <-chan openrouter.Usage, <-chan error) {
			return fakeStream(nil, nil)
		},
	}
	svc := NewArticleGenerationService(openRouter, articleTestConfig())
	chunks, complete, errCh, err := svc.GenerateArticleStream(context.Background(), &dto.GenerateArticleRequest{Prompt: "abc"})
	if err != nil {
		t.Fatalf("GenerateArticleStream: %v", err)
	}
	_, article, streamErr := drainArticle(t, chunks, complete, errCh)
	if !errors.Is(streamErr, apperrors.ErrAIEmptyResponse) || article != nil {
		t.Fatalf("err=%v article=%+v, want ErrAIEmptyResponse", streamErr, article)
	}
	if streamErr.Error() != "OpenRouter returned an empty response" {
		t.Fatalf("message = %q", streamErr.Error())
	}
}

func TestGenerateArticleStreamForwardsUpstreamError(t *testing.T) {
	t.Parallel()

	upstream := errors.New("OpenRouter API error: 429 Too Many Requests")
	openRouter := &mockOpenRouterService{
		generateStreamFn: func(context.Context, []openrouter.Message, *string, float64) (<-chan string, <-chan openrouter.Usage, <-chan error) {
			return fakeStream([]string{"partial"}, upstream)
		},
	}
	svc := NewArticleGenerationService(openRouter, articleTestConfig())
	chunks, complete, errCh, err := svc.GenerateArticleStream(context.Background(), &dto.GenerateArticleRequest{Prompt: "abc"})
	if err != nil {
		t.Fatalf("GenerateArticleStream: %v", err)
	}
	got, article, streamErr := drainArticle(t, chunks, complete, errCh)
	if !errors.Is(streamErr, upstream) || article != nil || len(got) != 1 {
		t.Fatalf("chunks=%q err=%v article=%+v", got, streamErr, article)
	}
}

func TestGenerateArticleStreamStopsOnCancel(t *testing.T) {
	t.Parallel()

	openRouter := &mockOpenRouterService{
		generateStreamFn: func(context.Context, []openrouter.Message, *string, float64) (<-chan string, <-chan openrouter.Usage, <-chan error) {
			return fakeStream([]string{"a", "b"}, nil)
		},
	}
	svc := NewArticleGenerationService(openRouter, articleTestConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // nobody reads chunks; the producer must still exit
	_, complete, errCh, err := svc.GenerateArticleStream(ctx, &dto.GenerateArticleRequest{Prompt: "abc"})
	if err != nil {
		t.Fatalf("GenerateArticleStream: %v", err)
	}
	select {
	case streamErr := <-errCh:
		if !errors.Is(streamErr, context.Canceled) {
			t.Fatalf("err = %v, want context.Canceled", streamErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("producer did not stop after cancel")
	}
	if _, ok := <-complete; ok {
		t.Fatal("complete must not carry a value after cancel")
	}
}
