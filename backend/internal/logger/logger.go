package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger

func L() *zap.Logger {
	return logger
}

func S() *zap.SugaredLogger {
	return sugar
}

func Init(prod bool) error {
	var cfg zap.Config
	if prod {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}
	l, err := cfg.Build()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create zap logger: %v\n", err)
		return err
	}
	logger = l
	sugar = l.Sugar()
	return nil
}

//// 推荐（性能好、类型安全）
//logger.L().Info("用户登录成功",
//zap.String("user", username),
//zap.String("ip", ip),
//)
//
//// 错误日志
//logger.L().Error("数据库查询失败",
//zap.Error(err),           // 自动包含 error 和 stack
//zap.String("sql", sqlStr),
//zap.Int("user_id", userID),
//)
//
//// 致命错误
//logger.L().Fatal("数据库连接失败，无法启动",
//zap.Error(err),
//)
