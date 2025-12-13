package main

import (
	"fmt"
	"log"
	"wsai/backend/config"
	"wsai/backend/internal/router"
	"wsai/backend/utils/mysql"
	"wsai/backend/utils/redis"

	"github.com/gin-gonic/gin"
)

func StartServer(addr string, port int) error {
	r := router.InitRouter()

	return r.Run(fmt.Sprintf("%s:%d", addr, port))

}

func main() {
	config.InitConfig()
	if err := mysql.InitMysql(); err != nil {
		log.Println("MySQL 初始化失败:", err)
	}
	if err := redis.Init(); err != nil {
		log.Printf("Redis 配置加载结果: host=%s, port=%d, password=%s, db=%d",
			config.C.RedisConfig.Host, config.C.RedisConfig.Port, config.C.RedisConfig.Password, config.C.RedisConfig.DB)
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
		panic(err)
	}
}
