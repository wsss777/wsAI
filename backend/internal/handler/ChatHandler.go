package handler

import (
	"io"
	"net/http"
	"wsai/backend/internal/model"
	"wsai/backend/utils/common"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

func ChatHandler(c *gin.Context) {
	userMsg := c.Query("message")
	if userMsg == "" {
		common.Error(c, 400, "message不能为空")
		return
	}

	stream, err := model.Get().Stream(c.Request.Context(), []*schema.Message{
		{Role: "user", Content: userMsg},
	})
	if err != nil {
		common.Error(c, 500, err.Error())
		return
	}
	defer stream.Close()

	// SSE响应头
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no")

	c.Stream(func(w io.Writer) bool {
		flusher, ok := w.(http.Flusher)
		if !ok {
			// 理论上不会走到这里，除非底层实现变更
			common.Error(c, 500, "响应Writer不支持流式刷新")
			return false
		}

		msg, err := stream.Recv()
		if err != nil {
			return false
		}
		if msg.Content != "" {
			_, writeErr := w.Write([]byte("data: " + msg.Content + "\n\n"))
			if writeErr != nil {
				return false
			}
			flusher.Flush()
		}
		return true
	})

	_, _ = c.Writer.Write([]byte("data: [DONE]\n\n"))
	c.Writer.(http.Flusher).Flush()
}
