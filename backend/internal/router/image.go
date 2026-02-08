package router

import (
	"wsai/backend/internal/handler/image"

	"github.com/gin-gonic/gin"
)

func ImageRouter(r *gin.RouterGroup) {
	r.POST("/recognize", image.RecognizeImage)
}
