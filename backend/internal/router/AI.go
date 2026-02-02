package router

import (
	"wsai/backend/internal/handler/session"

	"github.com/gin-gonic/gin"
)

func AIRouter(r *gin.RouterGroup) {
	// 聊天会话资源集合：/chatMessage/sessions
	sessions := r.Group("/chatMessage/sessions")
	{

		sessions.GET("", session.GetUserSessionsByUsername)

		sessions.POST("/stream", session.CreateStreamSessionAndSendFirstMessage)
	}
	// 单个会话资源：/chatMessage/sessions/:session_id
	sessionGroup := sessions.Group("/:session_id")
	{
		sessionGroup.POST("/messages/stream", session.SendMessageStream)

		sessionGroup.POST("/messages", session.GetMessageHistory)
	}
}
