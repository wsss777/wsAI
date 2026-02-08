package image

import (
	"net/http"
	"wsai/backend/internal/common"
	"wsai/backend/internal/common/code"
	"wsai/backend/internal/logger"
	"wsai/backend/internal/service/image"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RecognizeImageResponse struct {
	ClassName string `json:"class_name,omitempty"`
	common.Response
}

func RecognizeImage(c *gin.Context) {
	res := new(RecognizeImageResponse)
	file, err := c.FormFile("image")
	if err != nil {
		logger.L().Warn("FromFile failed to get from file",
			zap.Error(err))
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	className, err := image.RecognizeImage(file)
	if err != nil {
		logger.L().Error("RecognizeImage failed to recognize",
			zap.Error(err),
			zap.String("filename", file.Filename))
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
	res.Success()
	res.ClassName = className
	c.JSON(http.StatusOK, res)

}
