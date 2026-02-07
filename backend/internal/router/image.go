package router

import "github.com/gin-gonic/gin"

func ImageRouter(r *gin.RouterGroup) {
	r.POST("/recognize", image.RecognizeImage)
}
