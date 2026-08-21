package image

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"wsai/backend/infra/image"
	"wsai/backend/infra/logger"

	"go.uber.org/zap"
)

func RecognizeImage(file *multipart.FileHeader) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		logger.L().Error("无法获取当前工作目录", zap.Error(err))
		return "", err
	}

	// 从当前工作目录向上两级定位到后端根目录。
	rootDir := filepath.Join(cwd, "..", "..")

	// 模型和标签文件路径
	modelPath := filepath.Join(rootDir, "data", "models", "mobilenetv2-7.onnx")
	labelPath := filepath.Join(rootDir, "data", "imagenet_classes.txt")
	inputH, inputW := 224, 224

	logger.L().Info("图像识别服务启动",
		zap.String("modelPath", modelPath),
		zap.String("labelPath", labelPath),
		zap.String("rootDir", rootDir))

	recognizer, err := image.NewImageRecognizer(modelPath, labelPath, inputH, inputW)
	if err != nil {
		logger.L().Error("RecognizeImage failed to create image recognizer",
			zap.Error(err),
			zap.String("modelPath", modelPath),
			zap.String("labelPath", labelPath),
			zap.String("filename", file.Filename))
		return "", err
	}
	defer recognizer.Close()

	src, err := file.Open()
	if err != nil {
		logger.L().Error("RecognizeImage failed to open file",
			zap.Error(err),
			zap.String("filename", file.Filename))
		return "", err
	}
	defer src.Close()

	buf, err := io.ReadAll(src)
	if err != nil {
		logger.L().Error("RecognizeImage failed to read file content into buffer",
			zap.Error(err),
			zap.String("filename", file.Filename))
		return "", err
	}
	return recognizer.PredictFromBuffer(buf)
}
