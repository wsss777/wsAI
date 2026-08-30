package session

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	ragapp "wsai/backend/application/rag"
	ai "wsai/backend/infra/llm"
	"wsai/backend/infra/logger"
	session "wsai/backend/infra/mysql/repository"
	"wsai/backend/model"
	"wsai/backend/response/code"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ctx = context.Background()

func GetUserSessionsByUsername(username string) ([]model.SessionInfo, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	sessions, err := session.FindUserSessions(username)
	if err != nil {
		return nil, err
	}
	// 没有数据时返回空数组
	if len(sessions) == 0 {
		return []model.SessionInfo{}, nil
	}
	infos := make([]model.SessionInfo, 0, len(sessions))

	for _, sess := range sessions {
		title := sess.Title
		infos = append(infos, model.SessionInfo{
			SessionID: sess.ID,
			Title:     title,
			UpdatedAt: sess.UpdatedAt,
		})
	}
	return infos, nil
}
func CreateStreamSessionOnly(username string, userQuestion string) (string, code.Code) {
	question := strings.TrimSpace(userQuestion)
	if question == "" {
		question = "新会话"
	}
	if len(question) > 80 {
		question = question[:77] + "..."
	}
	newSession := &model.Session{
		ID:       uuid.New().String(),
		UserName: username,
		Title:    question,
	}
	createdSession, err := session.CreateSession(newSession)
	if err != nil {
		logger.L().Warn("session.CreateSession error",
			zap.String("username", username),
			zap.String("question_preview", question[:min(50, len(question))]),
			zap.Error(err))
		return "", code.CodeServerBusy

	}
	return createdSession.ID, code.CodeSuccess
}

func StreamMessageToExistingSession(userName string, sessionID string, userQuestion string, modelType string, writer http.ResponseWriter) code.Code {
	if err := session.TouchSession(sessionID); err != nil {
		logger.L().Warn("session.TouchSession error", zap.String("session_id", sessionID), zap.Error(err))
	}
	// 确保响应写入器支持立即刷新。
	flusher, ok := writer.(http.Flusher)
	if !ok {
		logger.L().Warn("streamMessageToExistingSession http.Flusher error")
		return code.CodeServerBusy
	}

	manager := ai.GetGlobalManager()
	modelType, err := ai.NormalizeModelType(modelType)
	if err != nil {
		logger.L().Error("invalid chat model provider", zap.Error(err))
		return code.AIModelFail
	}
	helper, err := manager.GetOrCreateAIHelper(userName, sessionID, modelType, nil)
	if err != nil {
		logger.L().Error("manager.GetOrCreateAIHelper error , failed to create AI helper",
			zap.String("username", userName),
			zap.String("sessionId", sessionID),
			zap.String("modelType", modelType),
			zap.Error(err))
		return code.AIModelFail
	}

	cb := func(msg string) {
		zap.L().Debug("sending SSE chunk",
			zap.Int("length", len(msg)),
		)
		_, werr := writer.Write([]byte("data: " + msg + "\n\n"))
		if werr != nil {
			logger.L().Warn("SSE write error",
				zap.Error(werr))
		}
		return
	}
	flusher.Flush()
	zap.L().Debug("SSE message to existing session")

	retrieval, retrievalErr := ragapp.RetrieveContext(ctx, userName, sessionID, userQuestion, modelType)
	if retrievalErr != nil {
		logger.L().Warn("RAG 查询预处理或检索失败，继续使用可用结果", zap.Error(retrievalErr))
	}
	logger.L().Info("RAG 查询计划", zap.String("session_id", sessionID), zap.Bool("need_retrieval", retrieval.Plan.NeedRetrieval), zap.String("search_query", retrieval.Plan.SearchQuery), zap.String("planner", retrieval.Plan.Planner), zap.Int("citation_count", len(retrieval.Citations)))
	_, err_ := helper.StreamResponseWithContext(userName, ctx, cb, userQuestion, retrieval.Knowledge)
	if err_ != nil {
		zap.L().Error("StreamMessageToExistingSession StreamResponse error",
			zap.String("username", userName),
			zap.String("sessionId", sessionID),
			zap.String("modelType", modelType),
			zap.Error(err_))
		writeStreamError(writer, flusher, modelErrorMessage(err_))
		return code.AIModelFail
	}
	if len(retrieval.Citations) > 0 {
		data, _ := json.Marshal(retrieval.Citations)
		_, _ = writer.Write([]byte("event: citations\ndata: " + string(data) + "\n\n"))
	}

	_, err = writer.Write([]byte("data: [DONE]\n\n"))
	if err != nil {
		logger.L().Warn("StreamMessageToExistingSession write DONE error",
			zap.Error(err))
		return code.AIModelFail
	}

	flusher.Flush()

	return code.CodeSuccess

}

// writeStreamError 向前端发送可直接展示的流式错误消息。
func writeStreamError(writer http.ResponseWriter, flusher http.Flusher, message string) {
	payload, _ := json.Marshal(map[string]string{"message": message})
	_, _ = writer.Write([]byte("event: error\ndata: " + string(payload) + "\n\n"))
	flusher.Flush()
}

// modelErrorMessage 将第三方模型错误转换为用户可理解的提示。
func modelErrorMessage(err error) string {
	detail := strings.ToLower(err.Error())
	if strings.Contains(detail, "rate limit") || strings.Contains(detail, "429") || strings.Contains(detail, "1302") {
		return "当前模型请求过于频繁，已触发服务限流，请稍后再试。"
	}
	if strings.Contains(detail, "deadline exceeded") || strings.Contains(detail, "timeout") {
		return "模型服务响应超时，请稍后再试。"
	}
	return "模型服务请求失败，请检查模型配置或稍后再试。"
}

func ChatStreamSend(userName string, sessionID string, userQuestion string, modelType string, writer http.ResponseWriter) code.Code {
	return StreamMessageToExistingSession(userName, sessionID, userQuestion, modelType, writer)
}

func GetChatHistory(userName string, sessionID string) ([]model.History, code.Code) {
	sessionInfo, err := session.GetSessionByID(sessionID)
	if err != nil || sessionInfo.UserName != userName {
		return nil, code.CodeServerBusy
	}
	messages, err := session.GetMessageBySessionID(sessionID)
	if err != nil {
		logger.L().Error("GetChatHistory get messages error", zap.String("session_id", sessionID), zap.Error(err))
		return nil, code.CodeServerBusy
	}
	history := make([]model.History, 0, len(messages))

	for _, msg := range messages {
		history = append(history, model.History{
			IsUser:  msg.IsUser,
			Content: msg.Content,
		})
	}
	return history, code.CodeSuccess
}
