package model

import (
	"time"
)

type Message struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID string    `gorm:"type:varchar(36);uniqueIndex:uk_messages_message_id" json:"message_id"`
	SessionID string    `gorm:"index;index:idx_message_session_created,priority:1;not null;type:varchar(36)" json:"session_id"`
	UserName  string    `gorm:"type:varchar(20)" json:"username"`
	Content   string    `gorm:"type:text" json:"content"`
	IsUser    bool      `gorm:"not null;" json:"is_user"`
	CreatedAt time.Time `gorm:"index:idx_message_session_created,priority:2" json:"created_at"`
}

type History struct {
	IsUser  bool   `json:"is_user"`
	Content string `json:"content"`
}
