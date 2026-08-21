package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	// DocumentParseStatusPending 表示文档已上传，等待异步解析。
	DocumentParseStatusPending = "pending"
	// DocumentParseStatusProcessing 表示文档正在解析、切片和向量化。
	DocumentParseStatusProcessing = "processing"
	// DocumentParseStatusReady 表示文档解析完成，可以参与检索。
	DocumentParseStatusReady = "ready"
	// DocumentParseStatusFailed 表示文档解析失败，需要查看错误原因。
	DocumentParseStatusFailed = "failed"
)

type Document struct {
	// ID 是数据库内部主键
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	DocumentID     string         `gorm:"type:varchar(64);not null;uniqueIndex" json:"document_id"`
	UserName       string         `gorm:"type:varchar(50);not null;index" json:"username"`
	FileName       string         `gorm:"type:varchar(255);not null;" json:"file_name"`
	FileType       string         `gorm:"type:varchar(20);not null;" json:"file_type"`
	FileSize       int64          `gorm:"not null;" json:"file_size"`
	StoragePath    string         `gorm:"type:varchar(500);not null;" json:"storage_path"`
	ContentHash    string         `gorm:"type:varchar(64);not null;" json:"content_hash"`
	ParseStatus    string         `gorm:"type:varchar(20);not null;default:'pending';index" json:"parse_status"`
	ChunkCount     int64          `gorm:"not null;default:0" json:"chunk_count"`
	EmbeddingModel string         `gorm:"type:varchar(100);not null;default:''" json:"embedding_model"`
	ErrorMessage   string         `gorm:"type:text" json:"error_message"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type DocumentChunk struct {
	ID      int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	ChunkID string `gorm:"type:varchar(64);not null;uniqueIndex" json:"chunk_id"`
	//DocumentPK 是所属文档的数据库主键，数据库内部关联使用它。
	DocumentPK int64  `gorm:"not null;index" json:"document_pk"`
	DocumentID string `gorm:"type:varchar(64);not null;index" json:"document_id"`
	ChunkIndex int64  `gorm:"not null;index" json:"chunk_index"`
	// ChunkIndex 是文档内的切片序号，用于保持稳定顺序。
	PageNumber int64 `gorm:"not null;default:0" json:"page_number"`
	// PageNumber 表示切片来源页码，txt 和 md 可默认写 0。
	SectionTitle string `gorm:"type:varchar(255);not null;default:''" json:"section_title"`
	// SectionTitle 表示切片所属标题，没有时保持空字符串。
	Content string `gorm:"type:longtext;not null" json:"content"`
	// Content 是切片正文
	TokenCount int64 `gorm:"not null;default:0" json:"token_count"`
	// TokenCount 表示切片词元数，用于后续检索和提示词控制。
	EmbeddingModel string `gorm:"type:varchar(100);not null;default:''" json:"embedding_model"`
	VectorPointID  string `gorm:"type:varchar(64);not null;index" json:"vector_point_id"`
	//VectorPointID 是该切片在 Qdrant 中对应的点位 ID。
	MetaJSON datatypes.JSON `gorm:"type:json" json:"meta_json"`
	// MetaJSON 预留给页码、标题层级等附加元数据。
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type SessionDocument struct {
	ID         int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID  string         `gorm:"type:varchar(36);not null;uniqueIndex:uk_session_document;index" json:"session_id"`
	DocumentPK int64          `gorm:"not null;uniqueIndex:uk_session_document;index" json:"document_pk"`
	DocumentID string         `gorm:"type:varchar(64);not null;index" json:"document_id"`
	CreatedAt  time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
