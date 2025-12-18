package router

import (
	"wsai/backend/internal/handler/session"

	"github.com/gin-gonic/gin"
)

func AIRouter(r *gin.RouterGroup) {
	// 聊天会话资源集合：/chat/sessions
	sessions := r.Group("/chat/sessions")
	{
		// GET: 获取当前用户的所有会话列表（保留，因为不是发送接口）
		//1
		sessions.GET("", session.GetUserSessionsByUsername)

		// POST: 创建新会话并流式返回第一条 AI 回复
		// 请求体：{ "message": "用户想说的话" }
		// 响应：SSE 流式输出 AI 回复
		//2
		sessions.POST("/stream", session.CreateStreamSessionAndSendFirstMessage)
	}

	// 单个会话资源：/chat/sessions/:session_id
	sessionGroup := sessions.Group("/:session_id")
	{
		// POST: 向已有会话发送消息并流式返回 AI 回复
		// 请求体：{ "message": "用户想说的话" }
		// 响应：SSE 流式输出 AI 回复
		//3
		sessionGroup.POST("/messages/stream", session.SendMessageStream)

		// GET: 获取该会话的历史消息（保留，便于前端显示历史）
		// 支持查询参数分页，例如 ?limit=20&before_id=123
		//4
		sessionGroup.GET("/messages", session.GetMessageHistory)
	}
}
