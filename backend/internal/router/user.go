package router

import (
	"wsai/backend/internal/handler/user"

	"github.com/gin-gonic/gin"
)

func RegisterUserRouter(r *gin.RouterGroup) {
	r.POST("/register", user.RegisterHandler)
	r.POST("/login", user.Login)
	r.POST("/captcha", user.HandleCaptcha)
}
