package ai

import (
	"context"
	"fmt"
	"sync"
	"wsai/backend/internal/logger"

	"go.uber.org/zap"
)

const (
	ModelTypeOpenAI = "openai"
	ModelTypeOllama = "ollama"
	// 新增模型时在这里加常量
)

// ModelCreator 定义模型创建函数类型
type ModelCreator func(ctx context.Context, config map[string]interface{}) (AIModel, error)

// AIModelFactory AI模型工厂
type AIModelFactory struct {
	creators map[string]ModelCreator
	mu       sync.Mutex
}

var (
	globalFactory *AIModelFactory
	factoryOnce   sync.Once
)

// GetGlobalFactory 获取全局单例
func GetGlobalFactory() *AIModelFactory {
	factoryOnce.Do(func() {
		globalFactory = &AIModelFactory{
			creators: make(map[string]ModelCreator),
		}
		globalFactory.registerCreators()

	})
	return globalFactory
}

// 注册模型
func (f *AIModelFactory) registerCreators() {
	//openAI
	f.creators[ModelTypeOpenAI] = func(ctx context.Context, config map[string]interface{}) (AIModel, error) {
		return NewOpenAIModel(ctx)
	}

	//ollama
	f.creators[ModelTypeOllama] = func(ctx context.Context, config map[string]interface{}) (AIModel, error) {
		baseURL := config["baseURL"].(string)
		modelName, ok := config["modelName"].(string)
		if !ok {
			err := fmt.Errorf("ollama requires non-empty modelName")
			logger.L().Error("Failed to create Ollama model: missing modelName",
				zap.String("model_type", ModelTypeOllama),
				zap.Error(err),
			)
			return nil, err
		}
		return NewOllamaModel(ctx, baseURL, modelName)

	}
}

// CreateAIModel 根据 modelType 创建具体 AIModel 实例
func (f *AIModelFactory) CreateAIModel(ctx context.Context, modelType string, config map[string]interface{}) (AIModel, error) {

	creator, ok := f.creators[modelType]
	if !ok {
		return nil, fmt.Errorf("unsupported model type: %s", modelType)
	}
	return creator(ctx, config)
}

// CreateAIHelper 一键创建aihelper
func (f *AIModelFactory) CreateAIHelper(ctx context.Context, modelType string, sessionID string, config map[string]interface{}) (*AIHelper, error) {
	model, err := f.CreateAIModel(ctx, modelType, config)
	if err != nil {
		return nil, err
	}
	helper := NewAIHelper(model, sessionID)
	logger.L().Info("AIHelper created",
		zap.String("session_id", sessionID),
		zap.String("model_type", modelType),
	)
	return helper, nil

}

// RegisterModel 可扩展注册
func (f *AIModelFactory) RegisterModel(modelType string, creator ModelCreator) {
	if modelType == "" || creator == nil {
		logger.L().Warn("model type or creator is nil")
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.creators[modelType] = creator
	logger.L().Info("New AI model registered",
		zap.String("model_type", modelType))

}
