package session

import (
	"errors"
	"wsai/backend/internal/common/mysql"
	"wsai/backend/internal/model"

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
