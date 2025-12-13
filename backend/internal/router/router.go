package router

import "github.com/gin-gonic/gin"

func InitRouter() *gin.Engine {
	r := gin.Default()
	enterRouter := r.Group("api/v1")
	RegisterUserRouter(enterRouter.Group("user"))
	return r
}
