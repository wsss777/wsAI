package ai

import (
	"context"
	"sync"
)
//定义模型创建函数类型
type ModelCreator func(ctx context.Context , config map[string]interface{}) (AIModel ,error)
//AIModelFactory AI模型工厂
type AIModelFactory struct{
	creators map[string]ModelCreator
}

var(
	globalFactory *AIModelFactory
	factoryOnce sync.Once
)

//获取全局单例
func GetGlobalFactory() *AIModelFactory{
	factoryOnce.Do(func() {
		globalFactory = &AIModelFactory{
			creators:make(map[string]ModelCreator)
		}
		globalFactory.registerCreators()

	})
	return globalFactory
}

//注册模型
func (f *AIModelFactory) registerCreators() {
	//openAI
	f.creators["1"] = func(ctx context.Context , config map[string]interface{}) (AIModel ,error) {
		return NewOpenAIModel(ctx)
	}

	//ollama
	f.creators["2"] = func(ctx context.Context , config map[string]interface{}) (AIModel ,error) {
		baseURL:=config["base"]
	}
}
