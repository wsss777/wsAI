package model

import "time"

const (
	AgentRunRunning   = "running"
	AgentRunCompleted = "completed"
	AgentRunFailed    = "failed"
	AgentRunCancelled = "cancelled"
)

// AgentRun 记录一次 Agent 任务的完整生命周期，正文只存用户输入和最终回答。
type AgentRun struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserName        string     `gorm:"index;not null;type:varchar(50)" json:"username"`
	SessionID       string     `gorm:"index;not null;type:varchar(36)" json:"session_id"`
	ModelType       string     `gorm:"type:varchar(32);not null" json:"model_type"`
	Status          string     `gorm:"index;not null;type:varchar(20)" json:"status"`
	InputMessageID  string     `gorm:"type:varchar(36);index" json:"input_message_id"`
	OutputMessageID string     `gorm:"type:varchar(36);index" json:"output_message_id"`
	Input           string     `gorm:"type:text;not null" json:"input"`
	FinalAnswer     string     `gorm:"type:longtext" json:"final_answer"`
	Error           string     `gorm:"type:text" json:"error,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}

// AgentStep 是用户可见执行轨迹的持久化摘要，不保存模型思维链或完整工具结果。
type AgentStep struct {
	ID         uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	RunID      uint       `gorm:"index;not null" json:"run_id"`
	Sequence   int        `gorm:"not null" json:"sequence"`
	Type       string     `gorm:"type:varchar(32);not null" json:"type"`
	Status     string     `gorm:"type:varchar(20);not null" json:"status"`
	Summary    string     `gorm:"type:text;not null" json:"summary"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type AgentToolCall struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	RunID         uint       `gorm:"index;not null" json:"run_id"`
	StepID        uint       `gorm:"index;not null" json:"step_id"`
	ToolName      string     `gorm:"type:varchar(64);not null" json:"tool_name"`
	ArgumentsJSON string     `gorm:"type:text;not null" json:"arguments_json"`
	ResultSummary string     `gorm:"type:text;not null" json:"result_summary"`
	ErrorMessage  string     `gorm:"type:text" json:"error_message,omitempty"`
	ArtifactID    string     `gorm:"type:varchar(36);index" json:"artifact_id,omitempty"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// AgentArtifact 预留完整 diff、受限工具结果和命令输出的不可变存储引用；正文不进入 messages。
type AgentArtifact struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	RunID       uint      `gorm:"index;not null" json:"run_id"`
	Kind        string    `gorm:"type:varchar(32);not null" json:"kind"`
	ContentHash string    `gorm:"type:varchar(64);not null" json:"content_hash"`
	StoragePath string    `gorm:"type:varchar(512)" json:"storage_path"`
	Content     string    `gorm:"type:longtext" json:"content,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// AgentFileChange 为第二期补丁审批预留变更证据，不由第一期只读 Agent 写入。
type AgentFileChange struct {
	ID                 uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	RunID              uint       `gorm:"index;not null" json:"run_id"`
	ToolCallID         uint       `gorm:"index" json:"tool_call_id"`
	Path               string     `gorm:"type:varchar(512);not null" json:"path"`
	Operation          string     `gorm:"type:varchar(16);not null" json:"operation"`
	BeforeHash         string     `gorm:"type:varchar(64)" json:"before_hash"`
	AfterHash          string     `gorm:"type:varchar(64)" json:"after_hash"`
	ProposalArtifactID string     `gorm:"type:varchar(36);index" json:"proposal_artifact_id"`
	AppliedArtifactID  string     `gorm:"type:varchar(36);index" json:"applied_artifact_id"`
	ApprovalID         string     `gorm:"type:varchar(36);index" json:"approval_id"`
	AppliedAt          *time.Time `json:"applied_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// AgentApproval 为补丁和受控命令的不可变审批快照预留；第一期不创建审批记录。
type AgentApproval struct {
	ID                 string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	RunID              uint       `gorm:"index;not null" json:"run_id"`
	ToolCallID         uint       `gorm:"index;not null" json:"tool_call_id"`
	Status             string     `gorm:"index;not null;type:varchar(20)" json:"status"`
	SnapshotArtifactID string     `gorm:"type:varchar(36);index;not null" json:"snapshot_artifact_id"`
	ApprovedBy         string     `gorm:"type:varchar(50)" json:"approved_by"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
