package model

import (
	"time"
)

const (
	MessageRoleUser           = "user"
	MessageRoleAssistant      = "assistant"
	MessageRoleTool           = "tool"
	MessageVisibilityUser     = "user"
	MessageVisibilityTrace    = "trace"
	MessageVisibilityInternal = "internal"
)

type Message struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID string `gorm:"type:varchar(36);uniqueIndex:uk_messages_message_id" json:"message_id"`
	SessionID string `gorm:"index;index:idx_message_session_created,priority:1;index:idx_message_session_sequence,priority:1;not null;type:varchar(36)" json:"session_id"`
	UserName  string `gorm:"type:varchar(20)" json:"username"`
	Content   string `gorm:"type:text" json:"content"`
	IsUser    bool   `gorm:"not null;" json:"is_user"`
	// Role 是标准会话事件角色；IsUser 仅为兼容旧客户端而保留。
	Role          string    `gorm:"type:varchar(16);index" json:"role"`
	SequenceNo    uint64    `gorm:"index:idx_message_session_sequence,priority:2" json:"sequence_no"`
	ToolCallsJSON string    `gorm:"type:longtext" json:"tool_calls_json,omitempty"`
	ToolCallID    string    `gorm:"type:varchar(128);index" json:"tool_call_id,omitempty"`
	ToolName      string    `gorm:"type:varchar(64)" json:"tool_name,omitempty"`
	AgentRunID    uint      `gorm:"index" json:"agent_run_id,omitempty"`
	ArtifactID    string    `gorm:"type:varchar(36);index" json:"artifact_id,omitempty"`
	Visibility    string    `gorm:"type:varchar(16);index" json:"visibility"`
	CreatedAt     time.Time `gorm:"index:idx_message_session_created,priority:2" json:"created_at"`
}

type History struct {
	IsUser     bool   `json:"is_user"`
	Content    string `json:"content"`
	Role       string `json:"role,omitempty"`
	ToolName   string `json:"tool_name,omitempty"`
	ArtifactID string `json:"artifact_id,omitempty"`
	Visibility string `json:"visibility,omitempty"`
}
