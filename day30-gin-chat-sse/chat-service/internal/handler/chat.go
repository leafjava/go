package handler

import (
	"chat-service/internal/service"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatSvc *service.ChatService
}

func NewChatHandler(chatSvc *service.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

type StreamChatRequest struct {
	Messages []service.Message `json:"messages" binding:"required,min=1"`
	Model    string            `json:"model"`
}

func (h *ChatHandler) StreamChat(c *gin.Context) {
	var req StreamChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, BadRequest("请求参数错误:"+err.Error()))
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(200)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		Fail(c, Internal(fmt.Errorf("response writer 不支持 flush")))
		return
	}

	onChunk := func(chunk string) error {
		// SSE 协议：每个 event 用 \n\n 结束，data 是 JSON
		payload, _ := json.Marshal(map[string]any{
			"type":    "chunk",
			"content": chunk,
		})
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", payload); err != nil {
			return err
		}
		flusher.Flush() // 关键：写完就 flush，否则会等缓冲
		return nil
	}

	chatReq := service.ChatRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   true,
	}

	resp, err := h.chatSvc.StreamChat(c.Request.Context(), chatReq, onChunk)
	if err != nil {
		errPayload, _ := json.Marshal(map[string]any{
			"type":    "error",
			"message": err.Error(),
		})
		fmt.Fprintf(c.Writer, "event: error\\ndata: %s\\n\\n", errPayload)
		flusher.Flush()
		return
	}

	donePayload, _ := json.Marshal(map[string]any{
		"type":  "done",
		"usage": resp.Usage,
	})
	fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", donePayload)
	flusher.Flush()
}

func (h *ChatHandler) NonStreamChat(c *gin.Context) {
	var req StreamChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, BadRequest("请求参数错误: "+err.Error()))
		return
	}

	resp, err := h.chatSvc.StreamChat(c.Request.Context(), service.ChatRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   false,
	}, nil)

	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, resp)

}

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "time": time.Now().Unix()})
}

var _ = io.Discard
