package session

import (
	"context"
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

func GetUserSessionByUsername(username string) ([]model.SessionInfo, error) {
	manager := ai.GetGlobalManager()
	Sessions := manager.GetUserSessions(username)
	if len(Sessions) == 0 {
		return []model.SessionInfo{}, nil
	}
	SessionInfos := make([]model.SessionInfo, 0, len(Sessions))

	for _, sessionId := range Sessions {
		title, err := session.GetTitleBySessionID(sessionId)
		if err != nil {
			logger.L().Warn("session.GetTitleBySessionID error",
				zap.String("username", username),
				zap.String("sessionId", sessionId),
				zap.Error(err))

			title = sessionId
			if len(title) > 12 {
				title = sessionId[:8] + "..."
			}

			if title == "" {
				title = "新会话"
			}
		}

		if len(title) > 12 {
			title = sessionId[:8] + "..."
		}
		SessionInfos = append(SessionInfos, model.SessionInfo{
			SessionID: sessionId,
			Title:     title,
		})

	}
	return SessionInfos, nil

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
