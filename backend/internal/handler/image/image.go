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

// RecognizeImage 图像分类/识别
// @Summary      上传图片进行分类/识别
// @Description  上传单张图片，服务端返回识别出的主要类别名称
// @Description
// @Description  成功时：code=0，class_name 会有值
// @Description  失败时：code 为错误码（1001 参数错误 / 1002 服务器忙等），class_name 为空
// @Tags         图像识别
// @Accept       multipart/form-data
// @Produce      json
// @Param        image  formData  file  true  "要识别的图片文件（支持常见图片格式：jpg,png,gif等）"
// @Success      200   {object}  image.RecognizeImageResponse{code=common.Response.Code,class_name=string}  "成功返回识别结果"
// @Failure      200   {object}  image.RecognizeImageResponse  "业务错误（code≠0）也返回 200，这是项目规范"
// @Router       /api/v1/image/recognize [post]
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
