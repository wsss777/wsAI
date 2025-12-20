// @title WsAI Backend API
// @version 1.0
// @description WsAI 项目 Swagger 文档
// @host localhost:9091
// @BasePath /
package main

import (
	"fmt"
	"os"
	"wsai/backend/config"
	_ "wsai/backend/docs"
	"wsai/backend/internal/logger"
	"wsai/backend/internal/mysql"
	"wsai/backend/internal/redis"

	"wsai/backend/internal/router"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func StartServer(addr string, port int) error {
	r := router.InitRouter()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r.Run(fmt.Sprintf("%s:%d", addr, port))

}

func main() {
	config.InitConfig()
	if config.C.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode) // 默认
	}
	isProd := config.C.App.Env == "prod"
	if err := logger.Init(isProd); err != nil {
		panic(err)
	}
	defer func() {
		if err := logger.L().Sync(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "zap Logger.Sync() failed: %v\n", err)
		}
	}()
	logger.L().Info("服务启动",
		zap.String("version", "v1"),
		zap.String("env", config.C.App.Env),
		zap.String("host", config.C.App.Host),
		zap.Int("port", config.C.App.Port),
	)

	if err := mysql.Init(); err != nil {
		logger.L().Fatal("MySQL 初始化失败，无法继续运行", zap.Error(err))
	}
	if err := redis.Init(); err != nil {
		logger.L().Error("Redis 初始化失败，将影响相关功能", zap.Error(err))
	}

	//rabbitmq.Init()

	host := config.C.App.Host
	port := config.C.App.Port

	if err := StartServer(host, port); err != nil {
		logger.L().Fatal("服务器启动失败，无法继续运行",
			zap.Error(err),
			zap.String("listen_addr", fmt.Sprintf("%s:%d", host, port)),
		)
	}
}
