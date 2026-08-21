package handler

import (
	"wsai/backend/handler/image"
	"wsai/backend/handler/middleware/cors"
	"wsai/backend/handler/middleware/jwt"
	"wsai/backend/handler/rag"
	"wsai/backend/handler/session"
	"wsai/backend/handler/user"

	"github.com/gin-gonic/gin"
)

// InitRouter 在传输层组装 HTTP 路由。
func InitRouter() *gin.Engine {
	r := gin.Default()
	_ = r.SetTrustedProxies(nil)
	r.Use(cors.Middleware())
	api := r.Group("/api/v1")
	api.POST("/user/users", user.Register)
	api.POST("/user/login", user.Login)
	api.POST("/user/email-login", user.EmailLogin)
	api.POST("/user/captcha", user.HandleCaptcha)

	ai := api.Group("/AI")
	ai.Use(jwt.AuthMiddleware())
	ai.GET("/chatMessage/sessions", session.GetUserSessionsByUsername)
	ai.POST("/chatMessage/sessions/stream", session.CreateStreamSessionAndSendFirstMessage)
	ai.POST("/chatMessage/sessions/:session_id/messages/stream", session.SendMessageStream)
	ai.GET("/chatMessage/sessions/:session_id/messages", session.GetMessageHistory)

	images := api.Group("/image")
	images.Use(jwt.AuthMiddleware())
	images.POST("/recognize", image.RecognizeImage)

	ragGroup := api.Group("/rag")
	ragGroup.Use(jwt.AuthMiddleware())
	ragGroup.POST("/documents", rag.UploadDocument)
	ragGroup.GET("/documents", rag.ListDocuments)
	ragGroup.GET("/documents/:document_id", rag.GetDocument)
	ragGroup.DELETE("/documents/:document_id", rag.DeleteDocument)
	ragGroup.POST("/sessions/:session_id/documents", rag.BindDocuments)
	ragGroup.GET("/sessions/:session_id/documents", rag.ListSessionDocuments)
	ragGroup.DELETE("/sessions/:session_id/documents/:document_id", rag.UnbindDocument)
	return r
}
