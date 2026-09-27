package repository

import (
	"time"

	"wsai/backend/infra/mysql"
	"wsai/backend/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetSessionMemory(sessionID string) (*model.SessionMemory, error) {
	var value model.SessionMemory
	err := mysql.DB.First(&value, "session_id = ?", sessionID).Error
	if err == gorm.ErrRecordNotFound {
		return &model.SessionMemory{SessionID: sessionID}, nil
	}
	return &value, err
}

func SaveSessionMemory(value *model.SessionMemory) error {
	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		var session model.Session
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&session, "id = ?", value.SessionID).Error; err != nil {
			return err
		}
		var current model.SessionMemory
		err := tx.First(&current, "session_id = ?", value.SessionID).Error
		if err == nil && current.CoveredSequenceNo >= value.CoveredSequenceNo {
			return nil
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		return tx.Save(value).Error
	})
}

func ListMessagesAfter(sessionID string, sequence uint64) ([]model.Message, error) {
	var items []model.Message
	err := mysql.DB.Where("session_id = ? AND sequence_no > ?", sessionID, sequence).
		Order("sequence_no asc, id asc").Find(&items).Error
	return items, err
}

func CreateLongTermMemory(value *model.LongTermMemory) error { return mysql.DB.Create(value).Error }

func SupersedeLongTermMemories(username, category, memoryKey string) error {
	if memoryKey == "" {
		return nil
	}
	return mysql.DB.Model(&model.LongTermMemory{}).
		Where("user_name = ? AND category = ? AND memory_key = ? AND status = ?", username, category, memoryKey, model.LongTermMemoryActive).
		Updates(map[string]any{"status": model.LongTermMemoryDeleted, "updated_at": time.Now()}).Error
}

func ListActiveLongTermMemories(username string) ([]model.LongTermMemory, error) {
	var items []model.LongTermMemory
	err := mysql.DB.Where("user_name = ? AND status = ?", username, model.LongTermMemoryActive).
		Order("updated_at desc").Limit(100).Find(&items).Error
	return items, err
}

func DeleteLongTermMemories(username, contains string) (int64, error) {
	result := mysql.DB.Model(&model.LongTermMemory{}).
		Where("user_name = ? AND status = ? AND content LIKE ?", username, model.LongTermMemoryActive, "%"+contains+"%").
		Updates(map[string]any{"status": model.LongTermMemoryDeleted, "updated_at": time.Now()})
	return result.RowsAffected, result.Error
}

func DeleteLongTermMemoryByID(username, id string) (bool, error) {
	result := mysql.DB.Model(&model.LongTermMemory{}).
		Where("id = ? AND user_name = ? AND status = ?", id, username, model.LongTermMemoryActive).
		Updates(map[string]any{"status": model.LongTermMemoryDeleted, "updated_at": time.Now()})
	return result.RowsAffected == 1, result.Error
}

func TouchLongTermMemories(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return mysql.DB.Model(&model.LongTermMemory{}).Where("id IN ?", ids).Update("last_used_at", time.Now()).Error
}
