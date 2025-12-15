// @title WsAI Backend API
// @version 1.0
// @description WsAI 项目 Swagger 文档
// @host localhost:9091
// @BasePath /
package main

import (
	"fmt"
	"log"
	"wsai/backend/config"
	_ "wsai/backend/docs"

	"wsai/backend/internal/router"
	"wsai/backend/utils/mysql"
	"wsai/backend/utils/redis"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func StartServer(addr string, port int) error {
	r := router.InitRouter()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r.Run(fmt.Sprintf("%s:%d", addr, port))

}

func main() {
	config.InitConfig()
	if err := mysql.InitMysql(); err != nil {
		log.Println("MySQL 初始化失败:", err)
	}
	if err := redis.Init(); err != nil {
		log.Printf("Redis 初始化失败: %v", err)

	}

	//rabbitmq.Init()
	if config.C.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode) // 默认
	}
	host := config.C.App.Host
	port := config.C.App.Port

	err := StartServer(host, port)
	if err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
