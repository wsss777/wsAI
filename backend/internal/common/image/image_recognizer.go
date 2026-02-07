package image

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"
	"wsai/backend/internal/logger"

	ort "github.com/yalue/onnxruntime_go"
	"go.uber.org/zap"
	"golang.org/x/image/draw"
)

type ImageRecognizer struct {
	session      *ort.Session[float32]
	inputName    string
	outputName   string
	inputH       int
	outputH      int
	labels       []string
	inputTensor  *ort.Tensor[float32]
	outputTensor *ort.Tensor[float32]
}

const (
	defaultInputName  = "data"
	dafaultOutputName = "mobilenetv20_output_flatten0_reshape0"
)

var (
	initOnce sync.Once
	initErr  error
)

func NewImageRecognizer(modelPath, labelPath string, inputH, inputW int) (*ImageRecognizer, error) {
	initOnce.Do(func() {
		initErr = ort.InitializeEnvironment()
	})
	if initErr != nil {
		logger.L().Error("NewImageRecognize failed to initialize onnxruntime environment",
			zap.Error(initErr))
		return nil, initErr
	}

	//预先创建输入 Tensor
	inputShape := ort.NewShape(1, 3, int64(inputH), int64(inputW))
	inData := make([]float32, inputShape.FlattenedSize())
	inTensor, err := ort.NewTensor(inputShape, inData)
	if err != nil {
		logger.L().Error("NewImageRecognize failed to create input tensor",
			zap.Error(err),
			zap.String("modelPath", modelPath),
			zap.Int("inputH", inputH),
			zap.Int("inputW", inputW))
		return nil, err
	}
	//预先创建输出Tensor
	outShape := ort.NewShape(1, 100)
	outTensor, err := ort.NewEmptyTensor[float32](outShape)
	if err != nil {
		inTensor.Destroy()
		logger.L().Error("NewImageRecognize failed to create output tensor",
			zap.Error(err),
			zap.String("modelPath", modelPath))
		return nil, err
	}

	//创建 ONNX Session
	session, err := ort.NewSession[float32](
		modelPath,
		[]string{defaultInputName},
		[]string{defaultOutputName},
		[]*ort.Tensor[float32]{inTensor},
		[]*ort.Tensor[float32]{outTensor},
	)

}
