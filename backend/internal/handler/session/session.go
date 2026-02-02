package session

import (
	"fmt"
	"net/http"
	"wsai/backend/internal/common"
	"wsai/backend/internal/common/code"
	"wsai/backend/internal/model"
	"wsai/backend/internal/service/session"

	"github.com/gin-gonic/gin"
)

type (
	GetUserSessionsResponse struct {
		Sessions []model.SessionInfo `json:"sessions,omitempty"`
		common.Response
	}
	CreateSessionAndSendFirstMessageRequest struct {
		UserQuestion string `json:"question" binding:"required"`
		ModelType    string `json:"modelType" binding:"required"`
	}
	CreateSessionAndSendFirstMessageResponse struct {
		AiInformation string `json:"Information,omitempty"` // AI回答
		SessionID     string `json:"sessionId,omitempty"`   // 当前会话ID
		common.Response
	}
	SendMessageStreamRequest struct {
		UserQuestion string `json:"question" binding:"required"`
		ModelType    string `json:"modelType" binding:"required"`
		SessionID    string `json:"sessionId,omitempty" binding:"required"`
	}
	SendMessageStreamResponse struct {
		AiInformation string `json:"Information,omitempty"`
		common.Response
	}
	GetMeaasgeHistoryRequest struct {
		SessionID string `json:"sessionId,omitempty" binding:"required"`
	}
	GetMeaasgeHistoryResponse struct {
		history []model.History `json:"history"`
		common.Response
	}
)

// GetUserSessions 获取当前用户的所有会话列表
// GET /chatMessage/sessions
func GetUserSessionsByUsername(c *gin.Context) {
	res := new(GetUserSessionsResponse)
	username_ := c.GetString("username")

	userSessions, err := session.GetUserSessionByUsername(username_)
	if err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
	res.Success()
	res.Sessions = userSessions
	c.JSON(http.StatusOK, res)
}

// CreateStreamSessionAndSendFirstMessage 创建新会话并流式输出第一条 AI 回复
// POST /chatMessage/sessions/stream
func CreateStreamSessionAndSendFirstMessage(c *gin.Context) {
	req := new(CreateSessionAndSendFirstMessageRequest)
	userName := c.GetString("username")
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid request"})
		return
	}

	setSSEHeaders(c)

	//创建会话，获取session ID
	sessionID, code_ := session.CreateStreamSessionOnly(userName, req.UserQuestion)
	if code_ != code.CodeSuccess {
		c.SSEvent("error", gin.H{
			"message": "failed to create session",
		})
		return
	}
	//// 立即把 sessionID 返回给前端（用于前端立即显示新会话标签）
	c.Writer.WriteString(fmt.Sprintf("data: {\"sessionId\": \"%s\"}\n\n", sessionID))
	c.Writer.Flush()

	// 然后开始把本次回答进行流式发送（包含最后的 [DONE]）
	code_ = session.StreamMessageToExistingSession(userName, sessionID,
		req.UserQuestion, req.ModelType, http.ResponseWriter(c.Writer))
	if code_ != code.CodeSuccess {
		c.SSEvent("error", gin.H{
			"message": "failed to create session",
		})
		return
	}

}

// SendMessageStream 向已有会话发送消息并流式返回
// POST /chatMessage/sessions/:session_id/messages/stream
func SendMessageStream(c *gin.Context) {
	req := new(SendMessageStreamRequest)
	userName := c.GetString("username")
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid request"})
		return
	}
	setSSEHeaders(c)

	code_ := session.ChatStreamSend(userName, req.SessionID, req.UserQuestion, req.ModelType, http.ResponseWriter(c.Writer))
	if code_ != code.CodeSuccess {
		c.SSEvent("error", gin.H{
			"message": "failed to send message stream",
		})
		return
	}
}

// GetMessageHistory 获取指定会话的历史消息
// GET /chatMessage/sessions/:session_id/messages
func GetMessageHistory(c *gin.Context) {
	req := new(GetMeaasgeHistoryRequest)
	res := new(GetMeaasgeHistoryResponse)
	userName := c.GetString("username")
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	history, code_ := session.GetChatHistory(userName, req.SessionID)
	if code_ != code.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code_))
		return
	}
	res.Success()
	res.history = history
	c.JSON(http.StatusOK, res)
}

func setSSEHeaders(c *gin.Context) {
	c.Header("content-type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no")
}
