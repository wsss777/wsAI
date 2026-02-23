package image

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"wsai/backend/internal/common/image"
	"wsai/backend/internal/logger"

	"go.uber.org/zap"
)

func RecognizeImage(file *multipart.FileHeader) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		logger.L().Error("无法获取当前工作目录", zap.Error(err))
		return "", err
	}

	// 从 cwd 向上两级到 backend 根目录（因为 cwd 通常是 cmd/server）
	rootDir := filepath.Join(cwd, "..", "..")

	// 模型和标签文件路径（假设放在 backend/data/ 下）
	modelPath := filepath.Join(rootDir, "data", "models", "mobilenetv2-7.onnx")
	labelPath := filepath.Join(rootDir, "data", "imagenet_classes.txt")
	inputH, inputW := 224, 224

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
