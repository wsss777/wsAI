package ai

import (
	"fmt"
	"strings"
	"wsai/backend/config"
)

// ChatProviderConfig 描述当前启用的对话模型配置。
type ChatProviderConfig struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
}

// NormalizeModelType 规范化前端传入的模型类型，并限制为已支持的提供方。
func NormalizeModelType(provider string) (string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return DefaultChatModelType()
	}

	switch provider {
	case ModelTypeOpenAI, "chatgpt":
		return ModelTypeOpenAI, nil
	case ModelTypeZhipu, "glm":
		return ModelTypeZhipu, nil
	default:
		return "", fmt.Errorf("unsupported CHAT_MODEL_PROVIDER: %s", provider)
	}
}

// DefaultChatModelType 读取 .env 中配置的默认模型。
func DefaultChatModelType() (string, error) {
	provider := config.C.ChatConfig.ModelProvider
	if strings.TrimSpace(provider) == "" {
		provider = ModelTypeOpenAI
	}
	return NormalizeModelType(provider)
}

// LoadChatProviderConfig 加载指定提供方的模型、地址和密钥。
func LoadChatProviderConfig(modelType string) (ChatProviderConfig, error) {
	provider, err := NormalizeModelType(modelType)
	if err != nil {
		return ChatProviderConfig{}, err
	}

	providerConfig := ChatProviderConfig{Provider: provider}
	switch provider {
	case ModelTypeOpenAI:
		providerConfig.APIKey = strings.TrimSpace(config.C.OpenAIConfig.APIKey)
		providerConfig.BaseURL = strings.TrimSpace(config.C.OpenAIConfig.BaseURL)
		providerConfig.Model = strings.TrimSpace(config.C.OpenAIConfig.Model)
	case ModelTypeZhipu:
		providerConfig.APIKey = strings.TrimSpace(config.C.ZhipuConfig.APIKey)
		providerConfig.BaseURL = strings.TrimSpace(config.C.ZhipuConfig.BaseURL)
		providerConfig.Model = strings.TrimSpace(config.C.ZhipuConfig.ChatModel)
	}

	if providerConfig.APIKey == "" || providerConfig.BaseURL == "" || providerConfig.Model == "" {
		return ChatProviderConfig{}, fmt.Errorf("%s chat model configuration is incomplete", provider)
	}
	return providerConfig, nil
}
