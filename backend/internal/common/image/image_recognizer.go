package image

import (
	"bufio"
	"bytes"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
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
	inputW       int
	labels       []string
	inputTensor  *ort.Tensor[float32]
	outputTensor *ort.Tensor[float32]
}

const (
	defaultInputName  = "data"
	defaultOutputName = "mobilenetv20_output_flatten0_reshape0"
)

var (
	initOnce sync.Once
	initErr  error
)

func init() {
	if runtime.GOOS != "windows" {
		return
	}
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("无法获取当前工作目录: %v", err)
	}

	dllPath := filepath.Join(cwd, "/backend/", "onnxruntime.dll")

	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		log.Fatalf("onnxruntime.dll 文件不存在，路径: %s\n请检查文件位置", dllPath)
	}
	ort.SetSharedLibraryPath(dllPath)
	log.Printf("[ORT INIT] 成功设置 ONNX Runtime DLL 路径: %s", dllPath)
}

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
	if err != nil {
		inTensor.Destroy()
		outTensor.Destroy()
		logger.L().Error("NewImageRecognize failed to create onnx session",
			zap.Error(err),
			zap.String("modelPath", modelPath),
			zap.String("inputName", defaultInputName),
			zap.String("outputName", defaultOutputName),
			zap.Int("inputH", inputH),
			zap.Int("inputW", inputW))
		return nil, err
	}
	// 读取 label 文件
	labels, err := loadLabels(labelPath)
	if err != nil {
		session.Destroy()
		inTensor.Destroy()
		outTensor.Destroy()
		logger.L().Error("NewImageRecognize failed to load labels",
			zap.Error(err),
			zap.String("labelPath", labelPath),
		)
		return nil, err
	}

	return &ImageRecognizer{
		session:      session,
		inputName:    defaultInputName,
		outputName:   defaultOutputName,
		inputH:       inputH,
		inputW:       inputW,
		labels:       labels,
		inputTensor:  inTensor,
		outputTensor: outTensor,
	}, nil
}

func (r *ImageRecognizer) Close() {
	if r.session != nil {
		_ = r.session.Destroy()
		r.session = nil
	}
	if r.inputTensor != nil {
		_ = r.inputTensor.Destroy()
		r.inputTensor = nil
	}
	if r.outputTensor != nil {
		_ = r.outputTensor.Destroy()
		r.outputTensor = nil
	}
}

func (r *ImageRecognizer) PredictFromFile(imagePath string) (string, error) {
	file, err := os.Open(filepath.Clean(imagePath))
	if err != nil {
		logger.L().Error("PredictFromFile failed to open image",
			zap.Error(err),
			zap.String("imagePath", imagePath),
		)
		return "", err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		logger.L().Error("PredictFromFile failed to decode image",
			zap.Error(err),
			zap.String("imagePath", imagePath))
		return "", err
	}
	return r.PredictFromImage(img)
}

func (r *ImageRecognizer) PredictFromBuffer(buf []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		logger.L().Error("PredictFromBiffer failed to decode image",
			zap.Error(err),
			zap.Int("bufferSizeBytes", len(buf)))
		return "", err
	}
	return r.PredictFromImage(img)
}
func (r *ImageRecognizer) PredictFromImage(img image.Image) (string, error) {
	resizedImg := image.NewRGBA(image.Rect(0, 0, r.inputW, r.inputH))
	draw.CatmullRom.Scale(resizedImg, resizedImg.Bounds(), img, img.Bounds(), draw.Over, nil)
	h, w := r.inputH, r.inputW
	ch := 3
	data := make([]float32, h*w*ch)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := resizedImg.At(x, y)
			rVal, gVal, bVal, _ := c.RGBA()

			rf := float32(rVal>>8) / 255.0
			gf := float32(gVal>>8) / 255.0
			bf := float32(bVal>>8) / 255.0
			offset := y*w + x
			data[offset] = rf
			data[h*w+offset] = gf
			data[2*h*w+offset] = bf
		}
	}

	//拷贝到预先分配的input tensor
	inData := r.inputTensor.GetData()
	copy(inData, data)
	//执行推理
	if err := r.session.Run(); err != nil {
		logger.L().Error("PredictFromImage failed to run onnx inference",
			zap.Error(err),
			zap.Int("inputH", r.inputH),
			zap.Int("inputW", r.inputW))
		return "", err
	}
	outData := r.outputTensor.GetData()
	if len(outData) == 0 {
		logger.L().Error("PredictFromImage onnx output is empty",
			zap.Int("expectedSize", 1000))
		return "", errors.New("PredictFromImage onnx output is empty")
	}
	maxIdx := 0
	MaxVal := outData[0]
	for i := 1; i < len(outData); i++ {
		if outData[i] > MaxVal {
			maxIdx = i
			MaxVal = outData[i]
		}
	}
	if maxIdx >= 0 && maxIdx < len(r.labels) {
		return r.labels[maxIdx], nil
	}
	logger.L().Warn("PredictFromImage index out of  label range",
		zap.Int("maxIdx", maxIdx),
		zap.Int("labelsLen", len(r.labels)))
	return "Unknown", nil

}

func loadLabels(path string) ([]string, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		logger.L().Error("PredictFromFile failed to open labels",
			zap.Error(err),
			zap.String("labelPath", path))
		return nil, err
	}
	defer f.Close()

	var labels []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line != "" {
			labels = append(labels, line)
		}
	}
	if err := sc.Err(); err != nil {
		logger.L().Error("PredictFromFile failed to read labels file",
			zap.Error(err),
			zap.String("labelPath", path))
		return nil, err
	}
	if len(labels) == 0 {
		logger.L().Error("PredictFromFile no labels found in file",
			zap.String("labelPath", path))
		return nil, errors.New("PredictFromFile no labels found in file")
	}
	return labels, nil
}
