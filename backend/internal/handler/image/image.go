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

// RecognizeImage 图片分类/识别
// @Summary      上传图片进行分类/识别
// @Description  上传单张图片，服务端返回识别出的主要类别名称。
// @Description  成功时返回 class_name，失败时通过 status_code 和 status_msg 返回错误信息。
// @Tags         图像识别
// @Accept       multipart/form-data
// @Produce      json
// @Param        image  formData  file  true  "要识别的图片文件"
// @Success      200   {object}  image.RecognizeImageResponse  "成功返回识别结果"
// @Failure      200   {object}  image.RecognizeImageResponse  "业务错误时也统一返回 200，通过 status_code 区分"
// @Router       /api/v1/image/recognize [post]
func RecognizeImage(c *gin.Context) {
	res := new(RecognizeImageResponse)
	file, err := c.FormFile("image")
	if err != nil {
		logger.L().Warn("获取上传图片失败", zap.Error(err))
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	className, err := image.RecognizeImage(file)
	if err != nil {
		logger.L().Error("图片识别失败",
			zap.Error(err),
			zap.String("filename", file.Filename))
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	res.Success()
	res.ClassName = className
	c.JSON(http.StatusOK, res)
}
