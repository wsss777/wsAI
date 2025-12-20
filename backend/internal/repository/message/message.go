package message

import (
	"wsai/backend/internal/logger"
	"wsai/backend/internal/model"
	"wsai/backend/internal/mysql"

	"go.uber.org/zap"
)

func GetMessageBySessionID(sessionID string) ([]model.Message, error) {
	var messages []model.Message
	err := mysql.DB.
		Where("session_id = ?", sessionID).
		Order("created_at asc").
		Find(&messages).Error
	if err != nil {
		logger.L().Error("GetMessageBySessionID err",
			zap.Error(err),
			zap.String("session_id", sessionID))
		return nil, err
	}
	return messages, nil
}

func GetMessageBySessionIDs(sessionIDs []string) ([]model.Message, error) {
	var messages []model.Message
	if len(sessionIDs) == 0 {
		return messages, nil
	}
	err := mysql.DB.
		Where("session_id IN (?)", sessionIDs).
		Order("created_at asc").
		Find(&messages).Error
	if err != nil {
		logger.L().Error("GetMessageBySessionIDs err",
			zap.Error(err),
			zap.Strings("session_ids", sessionIDs))
		return nil, err
	}
	return messages, nil
}

func CreateMessage(message *model.Message) (*model.Message, error) {
	err := mysql.DB.Create(message).Error
	if err != nil {
		logger.L().Error("CreateMessage err",
			zap.Error(err),
			zap.Uint("message_id", message.ID),
			zap.String("session_id", message.SessionID),
			zap.String("username", message.UserName),
			zap.Bool("is_user", message.IsUser))
		return nil, err
	}
	return message, nil
}

func GetAllMessages() ([]model.Message, error) {
	var messages []model.Message
	err := mysql.DB.
		Order("created_at asc").
		Find(&messages).Error
	if err != nil {
		logger.L().Error("GetAllMessages err",
			zap.Error(err),
		)
		return nil, err
	}
	return messages, nil
}
