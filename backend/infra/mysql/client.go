package mysql

import (
	"fmt"
	"time"
	"wsai/backend/config" // 路径已调整
	"wsai/backend/infra/logger"
	"wsai/backend/model" // 模型

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormLogger "gorm.io/gorm/logger"
)

// DB 全局 GORM 实例（全项目共用）
var DB *gorm.DB

// Init 初始化 MySQL 连接和自动迁移（在 main.go 中调用）
func Init() error {
	host := config.C.MysqlConfig.Host
	port := config.C.MysqlConfig.Port
	user := config.C.MysqlConfig.User
	password := config.C.MysqlConfig.Password
	database := config.C.MysqlConfig.Database
	charset := config.C.MysqlConfig.Charset

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		user, password, host, port, database, charset)

	var gormlogger gormLogger.Interface
	if gin.Mode() == "debug" {
		gormlogger = gormLogger.Default.LogMode(gormLogger.Info)
	} else {
		gormlogger = gormLogger.Default.LogMode(gormLogger.Warn)
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: gormlogger,
	})
	if err != nil {
		logger.L().Error("MySQL 连接失败",
			zap.Error(err),
		)
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.L().Error("获取 MySQL 底层 DB 失败",
			zap.Error(err),
		)
		return err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db

	if err := autoMigrate(); err != nil {
		logger.L().Error("MySQL 自动迁移失败",
			zap.Error(err),
		)
		return err
	}

	logger.L().Info("MySQL 初始化成功",
		zap.String("database", database),
	)
	return nil
}

// autoMigrate 自动迁移表结构
func autoMigrate() error {
	if err := DB.AutoMigrate(
		new(model.User),
		new(model.Session),
		new(model.SessionMemory),
		new(model.LongTermMemory),
		new(model.Message),
		new(model.Document),
		new(model.DocumentChunk),
		new(model.SessionDocument),
		new(model.AgentRun),
		new(model.AgentStep),
		new(model.AgentToolCall),
		new(model.AgentArtifact),
		new(model.AgentFileChange),
		new(model.AgentApproval),
	); err != nil {
		return err
	}
	return backfillConversationSequences()
}

// backfillConversationSequences 为早于会话全局顺序号的旧消息补齐数据。
// 该操作具备幂等性：所有记录拥有正序号后，启动时只执行轻量的去重查询。
func backfillConversationSequences() error {
	var sessionIDs []string
	if err := DB.Model(&model.Message{}).
		Distinct("session_id").
		Where("sequence_no = 0").
		Pluck("session_id", &sessionIDs).Error; err != nil {
		return err
	}
	for _, sessionID := range sessionIDs {
		if err := DB.Transaction(func(tx *gorm.DB) error {
			var session model.Session
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&session, "id = ?", sessionID).Error; err != nil {
				return err
			}
			var items []model.Message
			if err := tx.Where("session_id = ?", sessionID).
				Order("created_at asc, id asc").Find(&items).Error; err != nil {
				return err
			}
			for index := range items {
				role := items[index].Role
				if role == "" {
					if items[index].IsUser {
						role = model.MessageRoleUser
					} else {
						role = model.MessageRoleAssistant
					}
				}
				visibility := items[index].Visibility
				if visibility == "" {
					visibility = model.MessageVisibilityUser
				}
				if err := tx.Model(&model.Message{}).Where("id = ?", items[index].ID).Updates(map[string]any{
					"sequence_no": uint64(index + 1),
					"role":        role,
					"visibility":  visibility,
				}).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&session).Update("memory_sequence", uint64(len(items))).Error; err != nil {
				return err
			}
			// 旧消息尚不可见时生成的摘要不可信，应当删除后重新生成。
			return tx.Where("session_id = ?", sessionID).Delete(&model.SessionMemory{}).Error
		}); err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
