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

	ai.GET("/chatMessage/sessions", session.GetUserSessionsByUsername)                      // 获取当前用户的会话列表。
	ai.POST("/chatMessage/sessions/stream", session.CreateStreamSessionAndSendFirstMessage) // 创建会话并发送首条消息，以 SSE 持续返回模型生成内容。
	ai.POST("/chatMessage/sessions/:session_id/messages/stream", session.SendMessageStream) // 向指定会话追加消息，并以 SSE 持续返回模型生成内容。
	ai.GET("/chatMessage/sessions/:session_id/messages", session.GetMessageHistory)         // 查询指定会话的历史消息。

	images := api.Group("/image")
	images.Use(jwt.AuthMiddleware())
	images.POST("/recognize", image.RecognizeImage)

	ragGroup := api.Group("/rag")
	ragGroup.Use(jwt.AuthMiddleware())
	ragGroup.POST("/documents", rag.UploadDocument)                                     // 上传 PDF、Markdown 或文本文件，创建待解析的知识库文档。
	ragGroup.GET("/documents", rag.ListDocuments)                                       // 获取当前用户的知识库文档列表。
	ragGroup.GET("/documents/:document_id", rag.GetDocument)                            // 获取当前用户指定文档的详情和解析状态。
	ragGroup.DELETE("/documents/:document_id", rag.DeleteDocument)                      // 删除当前用户的指定文档、关联记录及上传源文件。
	ragGroup.POST("/sessions/:session_id/documents", rag.BindDocuments)                 // 为指定会话关联一个或多个当前用户的知识库文档。
	ragGroup.GET("/sessions/:session_id/documents", rag.ListSessionDocuments)           // 获取指定会话已关联的知识库文档。
	ragGroup.DELETE("/sessions/:session_id/documents/:document_id", rag.UnbindDocument) // 从指定会话移除一个知识库文档关联，不删除原文档。
	return r
}
