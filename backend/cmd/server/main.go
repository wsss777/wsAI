package main

import (
	"log"
	"wsai/backend/config"
	"wsai/backend/utils/mysql"

	"github.com/gin-gonic/gin"
)

var aiModel any

func main() {
	config.InitConfig()
	if err := mysql.InitMysql(); err != nil {
		log.Fatal("MySQL 初始化失败:", err)
	}
	if config.C.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode) // 默认
	}

	r := gin.Default()

	//r.GET("/api/chat", handler.ChatHandler)

	r.Run(":9091")
}
