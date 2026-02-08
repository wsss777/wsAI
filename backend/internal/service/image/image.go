package image

import (
	"io"
	"mime/multipart"
	"wsai/backend/internal/common/image"
	"wsai/backend/internal/logger"

	"go.uber.org/zap"
)

func RecognizeImage(file *multipart.FileHeader) (string, error) {
	modelPath := "/root/models/mobilenetv2/mobilenetv2-7.onnx"
	labelPath := "/root/imagenet_classes.txt"
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
