package repository

import (
	"errors"
	"time"
	"wsai/backend/infra/mysql"
	"wsai/backend/model"

	"gorm.io/gorm"
)

func GetTitleBySessionID(sessionID string) (string, error) {
	var sess model.Session
	err := mysql.DB.Select("title").
		Where("id = ? AND deleted_at IS NULL", sessionID).
		First(&sess).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return sess.Title, nil
}
func CreateSession(session *model.Session) (*model.Session, error) {
	err := mysql.DB.Create(session).Error
	return session, err
}
func GetSessionByID(sessionID string) (*model.Session, error) {
	var session model.Session
	err := mysql.DB.Where("id = ?", sessionID).First(&session).Error
	return &session, err
}
func FindUserSessions(userID string) ([]*model.Session, error) {
	var sessions []*model.Session
	err := mysql.DB.Select("id", "title", "updated_at").
		Where("user_name = ? AND deleted_at IS NULL", userID).
		Order("updated_at desc").
		Find(&sessions).Error
	return sessions, err
}

// TouchSession 在用户发送消息时立即记录活跃时间，
// 使当前会话在历史列表中排在最前。
func TouchSession(sessionID string) error {
	return mysql.DB.Model(&model.Session{}).
		Where("id = ? AND deleted_at IS NULL", sessionID).
		Update("updated_at", time.Now()).Error
}
