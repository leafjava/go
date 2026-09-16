package service

import (
	"chat-service/internal/config"
	"context"
	"fmt"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type ChatResponse struct {
	Content string
	Usage   Usage
}

type LLMClient interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	Stream(ctx context.Context, req ChatRequest, onChunk func(chunk string) error) (*ChatResponse, error)
}

func NewLLMClient(cfg *config.Config) LLMClient {
	if cfg.LLM.MockMode {
		return NewMockClient(cfg.LLM)
	}
	return NewMockClient(cfg.LLM)
}

type mockClient struct {
	cfg config.LLMConfig
}

func (m *mockClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	fullText, err := m.simulateStream(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	return &ChatResponse{
		Content: fullText,
		Usage: Usage{
			PromptTokens:     estimateTextTokens(joinMessages(req.Messages)),
			CompletionTokens: estimateTextTokens(fullText),
			TotalTokens:      estimateTextTokens(fullText) * 2,
		},
	}, nil
}

func (m *mockClient) Stream(ctx context.Context,req ChatRequest,onChunk func(chunk string)
error) (*ChatResponse,error){
	fullText,err := m.simulateStream(ctx,req,onChunk)
	if err != nil {
		return nil, err
	}
	return &ChatResponse{
		Content: fullText,
		Usage: Usage{
			PromptTokens:100,
			CompletionTokens:50,
			TotalTokens:150
		}
	}, nil
}

func (m *mockClient) simulateStream(ctx context.Context,req ChatRequest,onChunk func(string) error) (string, error){
	reply := m.fakeAnswer(req)
	var sb strings.Builder

	runes := []rune(reply)
	for _, r := range runes {
		select {
		case <-ctx.Done():
			return sb.String(),ctx.Err()
		case <-time.After(m.cfg.StreamChunk):
		}

		sb.WriteRune(r)
		if onChunk != nil {
			if err := onChunk(string(r));err != nil{
				return sb.String(),err
			}
		}
	}
	return sb.String(), nil
}

func (m *mockClient) fakeAnswer(req ChatRequest) string {
	last := lastUserMessage(req.Messages)
	if last == "" {
		return "你好，我是 mock，可以问我任何问题。"
	}

	templates := []string{
		"你说的是：「%s」。这是个有趣的问题，让我想想。",
		"关于「%s」我理解你的意思，不过我只是个 mock，需要 Day 3 才接真模型。",
		"「%s」——好的，已收到。当前在 mock 模式，正式上线前请切换 LLM_MOCK_MODE=false。",
	}
	idx := len(last) % len(templates)
	return fmt.Sprint(templates[idx],truncate(last,50))
}

func lastUserMessage(msgs []Message) string {
	for i := len(msgs) -1; i>=0;i-- {
		if msgs[i].Role == "user" {
			return msgs[i].Content
		}
	}
	return ""
}

func joinMessages(msgs []Message) string {
	var sb strings.Builder
	for _, m := range msgs {
		sb.WriteString(m.Role)
		sb.WriteString(": ")
		sb.WriteString(m.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

func truncates(s string,n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

func estimateTextTokens(s string) int {
	if s== "" {
		return 0
	}
	return len([]rune(s)) * 2 / 3
}


