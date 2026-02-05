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
	"wsai/backend/internal/ai"
	"wsai/backend/internal/common/mysql"
	"wsai/backend/internal/common/rabbitmq"
	"wsai/backend/internal/common/redis"
	"wsai/backend/internal/logger"
	"wsai/backend/internal/repository/message"
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

func readDataFromDB() error {
	manager := ai.GetGlobalManager()

	msgs, err := message.GetAllMessages()
	if err != nil {
		logger.L().Error("从数据库加载所有信息失败",
			zap.Error(err))
		return err
	}
	if len(msgs) == 0 {
		logger.L().Info("数据库无历史消息，无需恢复")
		return nil
	}
	logger.L().Info("开始从数据库恢复会话消息",
		zap.Int("total_messages", len(msgs)),
	)
	for i := range msgs {
		msg := &msgs[i]
		modelType := ai.ModelTypeOpenAI
		config := make(map[string]interface{})

		helper, err := manager.GetOrCreateAIHelper(msg.UserName, msg.SessionID, modelType, config)
		if err != nil {
			logger.L().Error("创建获取AIHelper失败",
				zap.String("username", msg.UserName),
				zap.String("sessionID", msg.SessionID),
				zap.Error(err),
			)
			continue
		}
		helper.AddMessage(msg.Content, msg.UserName, msg.IsUser, false)
		logger.L().Debug("成功恢复会话消息",
			zap.String("username", msg.UserName),
			zap.String("session_id", msg.SessionID),
		)
	}
	return nil
}

func main() {
	config.InitConfig()
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
		logger.L().Fatal(
			"MySQL 初始化失败，无法继续运行", zap.Error(err))
	}

	readDataFromDB()
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
