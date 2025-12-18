package ai

import (
	"sync"
	"wsai/backend/internal/model"
)

// 一个会话绑定一个AIHelper
type AIHelper struct {
	model     AIModel
	message   []*model.Message
	muRW      sync.RWMutex
	SessionID string
	saveFunc  func(*model.Message) (*model.Message, error)
}

func NewAIHelper(model_ AIModel, SessionID string) *AIHelper {
	return &AIHelper{
		model:   model_,
		message: make([]*model.Message, 0),
		saveFunc: func(msg *model.Message) (*model.Message, error) {
			//data:=rabbitmq.
		},
	}
}
