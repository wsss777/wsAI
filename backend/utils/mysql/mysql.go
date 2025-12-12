package mysql

import (
	"fmt"
	"time"
	"wsai/backend/config"
	"wsai/backend/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitMysql() error {
	host := config.C.MysqlConfig.Host
	port := config.C.MysqlConfig.Port
	user := config.C.MysqlConfig.User
	password := config.C.MysqlConfig.Password
	database := config.C.MysqlConfig.Database
	charset := config.C.MysqlConfig.Charset

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local", user, password, host, port, database, charset)

	var log logger.Interface
	if gin.Mode() == "debug" {
		log = logger.Default.LogMode(logger.Info)
	} else {
		log = logger.Default.LogMode(logger.Warn)
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: log,
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	DB = db
	return migration()
}

func migration() error {
	return DB.AutoMigrate(
		new(model.User),
		new(model.Session),
		new(model.Message),
	)
}

func InsertUser(user *model.User) (*model.User, error) {
	err := DB.Create(user).Error
	return user, err
}

func GetUserByUsername(username string) (*model.User, error) {
	user := &model.User{}
	err := DB.Where("username = ?", username).First(user).Error
	return user, err
}
func GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := DB.Where("email = ?", email).First(user).Error
	return user, err
}
