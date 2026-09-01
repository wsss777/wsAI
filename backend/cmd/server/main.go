// @title WsAI 后端接口
// @version 1.0
// @description WsAI 项目 Swagger 文档
// @host localhost:9091
// @BasePath /
package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"wsai/backend/config"
	_ "wsai/backend/docs"
	"wsai/backend/handler"
	"wsai/backend/infra/logger"
	"wsai/backend/infra/mysql"
	"wsai/backend/infra/rabbitmq"
	"wsai/backend/infra/redis"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func StartServer(addr string, port int) error {
	r := handler.InitRouter()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r.Run(fmt.Sprintf("%s:%d", addr, port))

}

func main() {
	config.InitConfig()
	if len(strings.TrimSpace(config.C.JWTConfig.Secret)) < 32 {
		log.Fatal("JWT_SECRET 未配置或长度不足 32 字节")
	}
	if ttl, err := time.ParseDuration(config.C.JWTConfig.AccessTTL); err != nil || ttl <= 0 {
		log.Fatal("jwt.access_ttl 必须是正的 Go duration，例如 30m")
	}
	isProd := config.C.App.Env == "prod"
	if isProd {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode) // 默认
	}
	if err := logger.Init(isProd); err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.L().Sync()
	}()
	logger.L().Info("服务启动",
		zap.String("version", "v1"),
		zap.String("env", config.C.App.Env),
		zap.String("host", config.C.App.Host),
		zap.Int("port", config.C.App.Port),
	)

	if err := mysql.Init(); err != nil {
		logger.L().Fatal(
			"MySQL 初始化失败，无法继续运行", zap.Error(err))
	}

	if err := redis.Init(); err != nil {
		logger.L().Error("Redis 初始化失败，将影响相关功能", zap.Error(err))
	}

	rabbitmq.InitRabbitMQ()

	host := config.C.App.Host
	port := config.C.App.Port

	if err := StartServer(host, port); err != nil {
		logger.L().Fatal("服务器启动失败，无法继续运行",
			zap.Error(err),
			zap.String("listen_addr", fmt.Sprintf("%s:%d", host, port)),
		)
	}
}
