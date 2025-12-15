package router

import (
	"github.com/gin-gonic/gin"
)

func AIRouter(r *gin.RouterGroup) {
	r.GET("/chat/sessions", session.GetUserSessionsByUsername)
	r.POST("/chat/send-new-session", session.CreateSessionAndSendMessage)
	r.POST("/chat/send")
	r.POST("/chat/history")
	r.POST("/chat/send-stream-new-session")
	r.POST("/chat/send-stream", session.ChatStreamSend)
}
