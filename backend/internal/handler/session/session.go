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
		History []model.History `json:"history"`
		common.Response
	}
)

// GetUserSessions 获取当前用户的所有会话列表
// @Summary      获取当前用户的所有会话列表
// @Description  返回当前登录用户创建的所有聊天会话（通常按创建时间降序）
// @Tags         会话管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200   {object}  session.GetUserSessionsResponse
// @Failure      200   {object}  session.GetUserSessionsResponse   // 注意：这里写 200，因为统一返回 200
// @Router       /api/v1/AI/chatMessage/sessions [get]
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

// CreateStreamSessionAndSendFirstMessage 创建新会话并流式输出 AI 的第一条回复
// @Summary      创建新会话 + 发送第一个问题（SSE 流式返回）
// @Description  创建一个新会话，同时把用户的第一个问题发给 AI，使用 SSE 流式返回
// @Tags         会话管理
// @Accept       json
// @Produce      text/event-stream
// @Param        body  body  session.CreateSessionAndSendFirstMessageRequest  true  "请求参数"
// @Security     ApiKeyAuth
// @Success      200   {string}  string   "SSE 事件流"
// @Failure      200   {string}  string   "SSE error 事件"
// @Router       /api/v1/AI/chatMessage/sessions/stream [post]
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

// SendMessageStream 向已有会话发送消息并流式返回 AI 回复
// @Summary      向已有会话追加消息（SSE 流式返回）
// @Description  向指定的会话 ID 发送一条新消息，并使用 SSE 流式返回 AI 回答
// @Tags         会话管理
// @Accept       json
// @Produce      text/event-stream
// @Param        session_id  path   string  true   "会话ID"
// @Param        body        body   session.SendMessageStreamRequest  true  "请求参数"
// @Security     ApiKeyAuth
// @Success      200   {string}  string   "SSE 事件流"
// @Failure      200   {string}  string   "SSE error 事件"
// @Router       /api/v1/AI/chatMessage/sessions/{session_id}/messages/stream [post]
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

// // GetMessageHistory 获取指定会话的历史消息
// // @Summary      获取指定会话的聊天历史
// // @Description  返回指定会话ID下的所有用户和AI消息历史
// // @Tags         ChatSession
// // @Accept       json
// // @Produce      json
// // @Param        session_id  path   string  true  "会话ID"
// // @Security     ApiKeyAuth
// // @Success      200   {object}  session.GetMeaasgeHistoryResponse
// // @Failure      500   {object}  common.Response
// // @Router       /chatMessage/sessions/:session_id/messages [get]
//
//	func GetMessageHistory(c *gin.Context) {
//		req := new(GetMeaasgeHistoryRequest)
//		res := new(GetMeaasgeHistoryResponse)
//		userName := c.GetString("username")
//
//		if err := c.ShouldBindJSON(req); err != nil {
//			c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
//			return
//		}
//		history, code_ := session.GetChatHistory(userName, req.SessionID)
//		if code_ != code.CodeSuccess {
//			c.JSON(http.StatusOK, res.CodeOf(code_))
//			return
//		}
//		res.Success()
//		res.history = history
//		c.JSON(http.StatusOK, res)
//	}
//
// GetMessageHistory 获取指定会话的聊天历史
// @Summary      获取指定会话的全部历史消息
// @Description  返回指定会话 ID 下的所有消息记录（按时间升序）
// @Tags         会话管理
// @Accept       json
// @Produce      json
// @Param        session_id  path   string  true   "会话ID"
// @Security     ApiKeyAuth
// @Success      200   {object}  session.GetMeaasgeHistoryResponse
// @Failure      200   {object}  session.GetMeaasgeHistoryResponse
// @Router       /api/v1/AI/chatMessage/sessions/{session_id}/messages [get]
func GetMessageHistory(c *gin.Context) {
	userName := c.GetString("username")
	sessionID := c.Param("session_id") // 直接从路径参数取 session_id

	// 简单校验 sessionID 是否为空
	if sessionID == "" {
		res := new(GetMeaasgeHistoryResponse)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	history, code_ := session.GetChatHistory(userName, sessionID)
	if code_ != code.CodeSuccess {
		res := new(GetMeaasgeHistoryResponse)
		c.JSON(http.StatusOK, res.CodeOf(code_)) // 保持你原有的风格：200 + error code
		return
	}

	res := new(GetMeaasgeHistoryResponse)
	res.Success()
	res.History = history
	c.JSON(http.StatusOK, res)
}

func setSSEHeaders(c *gin.Context) {
	c.Header("content-type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no")
}
