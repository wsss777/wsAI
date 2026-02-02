package mysql // Package mysql 包名 mysql

import (
	"fmt"
	"time"
	"wsai/backend/config" // 路径已调整
	"wsai/backend/internal/logger"
	"wsai/backend/internal/model" // 模型

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	return DB.AutoMigrate(
		new(model.User),
		new(model.Session),
		new(model.Message),
	)
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
