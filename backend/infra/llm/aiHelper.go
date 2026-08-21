package ai

import (
	"context"
	"github.com/cloudwego/eino/schema"
	"sync"
	"wsai/backend/infra/logger"
	"wsai/backend/infra/rabbitmq"
	"wsai/backend/model"
	"wsai/backend/utils"

	"go.uber.org/zap"
)

// AIHelper 表示与单个会话绑定的助手。
type AIHelper struct {
	model     AIModel
	messages  []*model.Message
	muRW      sync.RWMutex
	SessionID string
	saveFunc  func(*model.Message) (*model.Message, error)
}

func (a *AIHelper) StreamResponseWithContext(username string, ctx context.Context, cb StreamCallback, question, knowledge string) (*model.Message, error) {
	a.AddMessage(question, username, true, true)
	a.muRW.RLock()
	messages := utils.ConvertToSchemaMessages(a.messages)
	a.muRW.RUnlock()
	formatInstruction := "请使用清晰、简洁的 Markdown 回复：标题必须单独成行且后面留空行；用列表表达并列信息；代码使用代码块；避免把多个段落、标题或列表挤在同一行。"
	systemMessages := []*schema.Message{{Role: schema.System, Content: formatInstruction}}
	if knowledge != "" {
		systemMessages = append(systemMessages, &schema.Message{Role: schema.System, Content: "请仅依据以下资料回答；资料不足时明确说明。\n\n" + knowledge})
	}
	messages = append(systemMessages, messages...)
	content, err := a.model.StreamResponse(ctx, messages, cb)
	if err != nil {
		return nil, err
	}
	a.AddMessage(content, username, false, true)
	return &model.Message{SessionID: a.SessionID, Content: content, UserName: username, IsUser: false}, nil
}

// NewAIHelper 创建新的 AIHelper 实例。
func NewAIHelper(model_ AIModel, SessionID string) *AIHelper {
	return &AIHelper{
		model:    model_,
		messages: make([]*model.Message, 0, 20),
		saveFunc: func(msg *model.Message) (*model.Message, error) {
			data, genErr := rabbitmq.GenerateMessageMQPara(msg.SessionID, msg.Content, msg.UserName, msg.IsUser)
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

// addMessage 将消息添加到内存中，并调用自定义存储函数。
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

// SetSaveFunc 通过回调函数保存消息到数据库，以避免循环依赖。
// 传入外部保存函数后，可支持同步、异步等不同策略。
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
