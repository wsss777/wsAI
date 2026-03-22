package router

import (
	"wsai/backend/internal/handler/user"
	jwtmiddleware "wsai/backend/internal/middleware/jwt"

	"github.com/gin-gonic/gin"
)

func RegisterUserRouter(r *gin.RouterGroup) {
	r.POST("/users", user.Register)
	r.POST("/login", user.Login)
	r.POST("/captcha", user.HandleCaptcha)
	r.POST("/logout", jwtmiddleware.AuthMiddleware(), user.Logout)
}
