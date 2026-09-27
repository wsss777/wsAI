package ai

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type StreamCallback func(msg string)

type AIModel interface {
	GenerateResponse(ctx context.Context, messages []*schema.Message) (string, error)
	GenerateWithTools(ctx context.Context, messages []*schema.Message, tools []*schema.ToolInfo) (*schema.Message, error)
	StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error)
	GetModelType() string
	GetModelName() string
}

func (o *OpenAIModel) GenerateResponse(ctx context.Context, messages []*schema.Message) (string, error) {
	response, err := o.llm.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("%s generate failed: %w", o.modelType, err)
	}
	return response.Content, nil
}

// GenerateWithTools 为单次 Agent 决策创建绑定工具的模型副本，不改变普通聊天模型的状态。
func (o *OpenAIModel) GenerateWithTools(ctx context.Context, messages []*schema.Message, tools []*schema.ToolInfo) (*schema.Message, error) {
	bound, err := o.llm.WithTools(tools)
	if err != nil {
		return nil, fmt.Errorf("%s bind tools failed: %w", o.modelType, err)
	}
	response, err := bound.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("%s agent generate failed: %w", o.modelType, err)
	}
	return response, nil
}

// OpenAI 模型。

type OpenAIModel struct {
	llm       model.ToolCallingChatModel
	modelType string
	modelName string
}

// NewOpenAIModel 创建兼容 OpenAI Chat Completions 协议的对话模型。
// ChatGPT 与智谱 GLM 均通过该协议客户端调用。
func NewOpenAIModel(ctx context.Context, config ChatProviderConfig) (*OpenAIModel, error) {
	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: config.BaseURL,
		Model:   config.Model,
		APIKey:  config.APIKey,
		Timeout: 300 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("create %s chat model failed: %w", config.Provider, err)
	}
	return &OpenAIModel{llm: llm, modelType: config.Provider, modelName: config.Model}, nil
}

func (o *OpenAIModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	stream, err := o.llm.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("%s stream failed: %w", o.modelType, err)
	}
	defer stream.Close()

	var fullResp strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("%s stream recv failed: %w", o.modelType, err)
		}
		if len(msg.Content) > 0 {
			fullResp.WriteString(msg.Content)
			cb(msg.Content)
		}

	}
	return fullResp.String(), nil
}
func (o *OpenAIModel) GetModelType() string {
	return o.modelType
}

func (o *OpenAIModel) GetModelName() string {
	return o.modelName
}

// Ollama 模型。

type OllamaModel struct {
	llm       model.ToolCallingChatModel
	modelName string
}

func NewOllamaModel(ctx context.Context, baseURL, modelName string) (*OllamaModel, error) {
	llm, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, fmt.Errorf("create ollama model failed: %v", err)
	}
	return &OllamaModel{llm: llm, modelName: modelName}, nil
}
func (o *OllamaModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	stream, err := o.llm.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("ollama stream failed: %v", err)
	}
	defer stream.Close()
	var fullResp strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("ollama stream recv failed: %v", err)
		}
		if len(msg.Content) > 0 {
			fullResp.WriteString(msg.Content)
			cb(msg.Content)
		}

	}
	return fullResp.String(), nil
}
func (o *OllamaModel) GenerateResponse(ctx context.Context, messages []*schema.Message) (string, error) {
	response, err := o.llm.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("ollama generate failed: %w", err)
	}
	return response.Content, nil
}
func (o *OllamaModel) GenerateWithTools(ctx context.Context, messages []*schema.Message, tools []*schema.ToolInfo) (*schema.Message, error) {
	bound, err := o.llm.WithTools(tools)
	if err != nil {
		return nil, fmt.Errorf("ollama bind tools failed: %w", err)
	}
	response, err := bound.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("ollama agent generate failed: %w", err)
	}
	return response, nil
}
func (o *OllamaModel) GetModelType() string {
	return "Ollama"
}

func (o *OllamaModel) GetModelName() string {
	return o.modelName
}
