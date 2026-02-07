package router

import (
	"wsai/backend/internal/middleware/jwt"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	enterRouter := r.Group("api/v1")
	{
		RegisterUserRouter(enterRouter.Group("/user"))
	}
	{
		AIGroup := enterRouter.Group("/AI")
		AIGroup.Use(jwt.AuthMiddleware())
		AIRouter(AIGroup)
	}
	{
		ImageGroup := enterRouter.Group("/image")
		ImageRouter(ImageGroup)
		ImageRouter(ImageGroup)
	}

	return r
}
