package repository

import (
	"wsai/backend/infra/logger"
	"wsai/backend/infra/mysql"
	"wsai/backend/model"

	"go.uber.org/zap"
)

func GetMessageBySessionID(sessionID string) ([]model.Message, error) {
	var messages []model.Message
	err := mysql.DB.
		Where("session_id = ?", sessionID).
		Order("created_at asc, id asc").
		Find(&messages).Error
	if err != nil {
		logger.L().Error("GetMessageBySessionID err",
			zap.Error(err),
			zap.String("session_id", sessionID))
		return nil, err
	}
	return messages, nil
}

// GetRecentMessagesBySessionID 查询会话最近的 limit 条消息，并按时间正序返回。
func GetRecentMessagesBySessionID(sessionID string, limit int) ([]model.Message, error) {
	if limit <= 0 {
		return []model.Message{}, nil
	}
	var messages []model.Message
	err := mysql.DB.
		Where("session_id = ?", sessionID).
		Order("created_at desc, id desc").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		logger.L().Error("GetRecentMessagesBySessionID err",
			zap.Error(err),
			zap.String("session_id", sessionID),
			zap.Int("limit", limit))
		return nil, err
	}
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
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
		Order("created_at asc, id asc").
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
		Order("created_at asc, id asc").
		Find(&messages).Error
	if err != nil {
		logger.L().Error("GetAllMessages err",
			zap.Error(err),
		)
		return nil, err
	}
	return messages, nil
}
