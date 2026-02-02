package ai

import (
	"context"
	"sync"
	"wsai/backend/internal/common/rabbitmq"
	"wsai/backend/internal/logger"
	"wsai/backend/internal/model"
	"wsai/backend/internal/service/chatMessage"
	"wsai/backend/utils"

	"go.uber.org/zap"
)

// AIHelper 一个会话绑定一个AIHelper
type AIHelper struct {
	model     AIModel
	messages  []*model.Message
	muRW      sync.RWMutex
	SessionID string
	saveFunc  func(*model.Message) (*model.Message, error)
}

// NewAIHelper 创建新的AIHelper实例
func NewAIHelper(model_ AIModel, SessionID string) *AIHelper {
	return &AIHelper{
		model:    model_,
		messages: make([]*model.Message, 0, 20),
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
	a.messages = append(a.messages, &userMsg)
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

// SetSaveFunc SaveMessage 保存消息到数据库（通过回调函数避免循环依赖）
// 通过传入func，自己调用外部的保存函数，即可支持同步异步等多种策略
func (a *AIHelper) SetSaveFunc(saveFunc func(*model.Message) (*model.Message, error)) {
	a.saveFunc = saveFunc
}

// GetMessages 获取所有消息历史
func (a *AIHelper) GetAllMessage() []*model.Message {
	a.muRW.RLock()
	defer a.muRW.RUnlock()
	out := make([]*model.Message, len(a.messages))
	copy(out, a.messages)
	return out
}

// 流式生成
func (a *AIHelper) StreamResponse(username string,
	ctx context.Context, cb StreamCallback,
	userQuestion string) (*model.Message, error) {

	a.AddMessage(userQuestion, username, true, true)
	a.muRW.RLock()
	messages := utils.ConvertToSchemaMessages(a.messages)
	a.muRW.RUnlock()
	content, err := a.model.StreamResponse(ctx, messages, cb)
	if err != nil {
		logger.L().Error("AI model StreamResponse failed",
			zap.Error(err),
			zap.String("session_id", a.SessionID),
			zap.String("username", username))
		return nil, err
	}

	//构造保存完整AI回复

	modelMsg := &model.Message{
		SessionID: a.SessionID,
		Content:   content,
		UserName:  username,
		IsUser:    false,
	}
	a.AddMessage(content, username, false, true)

	return modelMsg, nil

}

// 获取当前使用的Ai模型
func (a *AIHelper) GetModelType() string {
	return a.model.GetModelType()
}
