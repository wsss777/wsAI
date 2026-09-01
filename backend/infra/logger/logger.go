package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const logFilePath = "logs/wsai.log"

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

func L() *zap.Logger {
	return logger
}

func S() *zap.SugaredLogger {
	return sugar
}

func Init(prod bool) error {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		return err
	}

	var cfg zap.Config
	if prod {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	// wsai.log 始终是当前正在写入的固定日志文件；旧日志按大小轮转，
	// 避免监控日志长期累积占满磁盘。
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     30,
		Compress:   true,
	})
	encoder := zapcore.NewJSONEncoder(cfg.EncoderConfig)
	if !prod {
		encoder = zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	}
	l := zap.New(
		zapcore.NewCore(encoder, writer, cfg.Level),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	logger = l
	sugar = l.Sugar()
	return nil
}
