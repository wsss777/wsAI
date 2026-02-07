package session

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"wsai/backend/internal/ai"
	"wsai/backend/internal/common/code"
	"wsai/backend/internal/logger"
	"wsai/backend/internal/model"
	"wsai/backend/internal/repository/session"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ctx = context.Background()

//	func GetUserSessionByUsername(username string) ([]model.SessionInfo, error) {
//		manager := ai.GetGlobalManager()
//		Sessions := manager.GetUserSessions(username)
//		if len(Sessions) == 0 {
//			return []model.SessionInfo{}, nil
//		}
//		SessionInfos := make([]model.SessionInfo, 0, len(Sessions))
//
//		for _, sessionId := range Sessions {
//			title, err := session.GetTitleBySessionID(sessionId)
//			if err != nil {
//				logger.L().Warn("session.GetTitleBySessionID error",
//					zap.String("username", username),
//					zap.String("sessionId", sessionId),
//					zap.Error(err))
//				if title == "" {
//					title = "新会话"
//				}
//			}
//			SessionInfos = append(SessionInfos, model.SessionInfo{
//				SessionID: sessionId,
//				Title:     title,
//			})
//
//		}
//		return SessionInfos, nil
//
// }
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
	//确保writer 支持flush
	flusher, ok := writer.(http.Flusher)
	if !ok {
		logger.L().Warn("streamMessageToExistingSession http.Flusher error")
		return code.CodeServerBusy
	}

	manager := ai.GetGlobalManager()
	config := map[string]interface{}{
		"apiKey": "api-key",
	}
	helper, err := manager.GetOrCreateAIHelper(userName, sessionID, modelType, config)
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

	_, err_ := helper.StreamResponse(userName, ctx, cb, userQuestion)
	if err_ != nil {
		zap.L().Error("StreamMessageToExistingSession StreamResponse error",
			zap.String("username", userName),
			zap.String("sessionId", sessionID),
			zap.String("modelType", modelType),
			zap.Error(err_))
		return code.AIModelFail
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

func ChatStreamSend(userName string, sessionID string, userQuestion string, modelType string, writer http.ResponseWriter) code.Code {
	return StreamMessageToExistingSession(userName, sessionID, userQuestion, modelType, writer)
}

func GetChatHistory(userName string, sessionID string) ([]model.History, code.Code) {
	manager := ai.GetGlobalManager()
	helper, exists := manager.GetAIHelper(userName, sessionID)
	if !exists {
		return nil, code.CodeServerBusy
	}
	messages := helper.GetAllMessage()
	history := make([]model.History, 0, len(messages))

	for i, msg := range messages {
		isUser := i%2 == 0
		history = append(history, model.History{
			IsUser:  isUser,
			Content: msg.Content,
		})
	}
	return history, code.CodeSuccess
}
