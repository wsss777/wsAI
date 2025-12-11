// model/model.go
package model

import (
	"context"
	"os"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type ChatStream = *schema.StreamReader[*schema.Message]

type AI = *openai.ChatModel

var (
	once     sync.Once
	instance *openai.ChatModel
)

func Get() AI {
	once.Do(func() {
		model, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			BaseURL: os.Getenv("OPENAI_BASE_URL"),
			Model:   os.Getenv("OPENAI_MODEL_NAME"),
		})
		if err != nil {
			panic("初始化大模型失败: " + err.Error())
		}
		instance = model
	})
	return instance
}
