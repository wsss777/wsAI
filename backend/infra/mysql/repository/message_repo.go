package repository

import (
	"errors"
	"wsai/backend/infra/logger"
	"wsai/backend/infra/mysql"
	"wsai/backend/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AppendConversationItem 是会话事件唯一的有序写入入口。
// 锁定会话行可为并发的 Agent 和工具写入分配稳定顺序号。
func AppendConversationItem(item *model.Message) (*model.Message, error) {
	if item.SessionID == "" || item.MessageID == "" {
		return nil, errors.New("session_id and message_id are required")
	}
	if item.Role == "" {
		if item.IsUser {
			item.Role = model.MessageRoleUser
		} else {
			item.Role = model.MessageRoleAssistant
		}
	}
	if item.Visibility == "" {
		item.Visibility = model.MessageVisibilityUser
	}
	item.IsUser = item.Role == model.MessageRoleUser
	err := mysql.DB.Transaction(func(tx *gorm.DB) error {
		var session model.Session
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&session, "id = ?", item.SessionID).Error; err != nil {
			return err
		}
		session.MemorySequence++
		item.SequenceNo = session.MemorySequence
		result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "message_id"}}, DoNothing: true}).Create(item)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&session).Update("memory_sequence", session.MemorySequence).Error
	})
	return item, err
}

func GetMessageBySessionID(sessionID string) ([]model.Message, error) {
	var messages []model.Message
	err := mysql.DB.
		Where("session_id = ?", sessionID).
		Order("sequence_no asc, created_at asc, id asc").
		Find(&messages).Error
	if err != nil {
		logger.L().Error("GetMessageBySessionID err",
			zap.Error(err),
			zap.String("session_id", sessionID))
		return nil, err
	}
	return messages, nil
}
