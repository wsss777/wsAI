package model

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID             string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserName       string         `gorm:"index;not null" json:"username"`
	Title          string         `gorm:"type:varchar(100)" json:"title"`
	MemorySequence uint64         `gorm:"not null;default:0" json:"-"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type SessionInfo struct {
	SessionID string    `json:"sessionId"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SessionMemory 保存截至 CoveredSequenceNo 的会话事件滚动摘要。
type SessionMemory struct {
	SessionID         string    `gorm:"primaryKey;type:varchar(36)" json:"session_id"`
	Summary           string    `gorm:"type:longtext;not null" json:"summary"`
	CoveredSequenceNo uint64    `gorm:"not null;default:0" json:"covered_sequence_no"`
	UpdatedAt         time.Time `json:"updated_at"`
}

const (
	LongTermMemoryActive  = "active"
	LongTermMemoryDeleted = "deleted"
)

// LongTermMemory 仅保存用户明确要求保留的跨会话事实或偏好。
type LongTermMemory struct {
	ID              string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserName        string     `gorm:"index;not null;type:varchar(50)" json:"username"`
	Category        string     `gorm:"index;not null;type:varchar(32)" json:"category"`
	MemoryKey       string     `gorm:"index;type:varchar(100)" json:"memory_key,omitempty"`
	Content         string     `gorm:"type:text;not null" json:"content"`
	Status          string     `gorm:"index;not null;type:varchar(16)" json:"status"`
	SourceSessionID string     `gorm:"index;type:varchar(36)" json:"source_session_id"`
	SourceMessageID string     `gorm:"index;type:varchar(36)" json:"source_message_id"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
