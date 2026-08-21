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
