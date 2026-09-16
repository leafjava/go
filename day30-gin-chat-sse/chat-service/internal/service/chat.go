package service

import (
	"context"
	"errors"

	"chat-service/internal/handler"
)

type ChatService struct {
	llm LLMClient
}

func NewChatService(llm LLMClient) *ChatService {
	return &ChatService{llm: llm}
}

func (s *ChatService) StreamChat(
	ctx context.Context,
	req ChatRequest,
	onChunk func(chunk string) error,
) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, errors.New("no messages")
	}

	if req.Model == "" {
		req.Model = "gpt-4o-mini"
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 1024
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	resp, err := s.llm.Stream(ctx, req, onChunk)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return resp, nil
		}
		return nil, handler.NewBizError(handler.CodeUpstreamFailed, "上游模型调用失败", err)
	}
	return resp, nil
}
