package aihelper

import (
	"context"
	"sync"
	"wsai/backend/internal/ai"
	"wsai/backend/internal/common/rabbitmq"
	"wsai/backend/internal/logger"
	"wsai/backend/internal/model"
	"wsai/backend/internal/service/chatMessage"

	"go.uber.org/zap"
)

// 一个会话绑定一个AIHelper
type AIHelper struct {
	model     ai.AIModel
	message   []*model.Message
	muRW      sync.RWMutex
	SessionID string
	saveFunc  func(*model.Message) (*model.Message, error)
}

// NewAIHelper 创建新的AIHelper实例
func NewAIHelper(model_ ai.AIModel, SessionID string) *AIHelper {
	return &AIHelper{
		model:   model_,
		message: make([]*model.Message, 0, 20),
		saveFunc: func(msg *model.Message) (*model.Message, error) {
			data, genErr := chatMessage.GenerateMessageMQPara(msg.SessionID, msg.Content, msg.UserName, msg.IsUser)
			if genErr != nil {
				logger.L().Error("Generate RabbitMQ message param failed",
					zap.Error(genErr),
					zap.String("session_id", msg.SessionID),
					zap.String("username", msg.UserName),
					zap.Bool("is_user", msg.IsUser))
				return msg, genErr
			}

			err := rabbitmq.RMQMessage.PublishWork(data)
			if err != nil {
				logger.L().Error("Async publish message to RabbitMQ failed",
					zap.Error(err),
					zap.String("session_id", msg.SessionID),
					zap.String("username", msg.UserName),
					zap.Bool("is_user", msg.IsUser),
					zap.Int("content_length", len(msg.Content)))
			}
			return msg, err

		},
		SessionID: SessionID,
	}
}

// addMessage 添加消息到内存中并调用自定义存储函数
func (a *AIHelper) AddMessage(Content string, UserName string, IsUser bool, Save bool) {
	userMsg := model.Message{
		SessionID: a.SessionID,
		Content:   Content,
		UserName:  UserName,
		IsUser:    IsUser,
	}
	a.message = append(a.message, &userMsg)
	if Save {
		if _, err := a.saveFunc(&userMsg); err != nil {
			logger.L().Warn("Call saveFunc failed ",
				zap.Error(err),
				zap.String("session_id", a.SessionID),
				zap.String("username", UserName),
				zap.Bool("is_user", IsUser),
			)
		}
	}
}

// SaveMessage 保存消息到数据库（通过回调函数避免循环依赖）
// 通过传入func，自己调用外部的保存函数，即可支持同步异步等多种策略
func (a *AIHelper) SetSaveFunc(saveFunc func(*model.Message) (*model.Message, error)) {
	a.saveFunc = saveFunc
}

// GetMessages 获取所有消息历史
func (a *AIHelper) GetAllMessage() []*model.Message {
	a.muRW.RLock()
	defer a.muRW.RUnlock()
	out := make([]*model.Message, len(a.message))
	copy(out, a.message)
	return out
}

// 流式生成
func (a *AIHelper) StreamResponse(username string,
	ctx context.Context, cb ai.StreamCallback,
	userQuestion string) (*model.Message, error) {

	a.AddMessage(userQuestion, username, true, true)
	a.muRW.RLock()
	message := ut
}
