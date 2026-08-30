package ai

import (
	"context"
	"fmt"
	"sync"
	"wsai/backend/infra/logger"
	"wsai/backend/infra/rabbitmq"
	"wsai/backend/model"
	"wsai/backend/utils"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"go.uber.org/zap"
)

// AIHelper 表示与单个会话绑定的助手。
type AIHelper struct {
	model     AIModel
	messages  []*model.Message
	muRW      sync.RWMutex
	streamMu  sync.Mutex
	SessionID string
	saveFunc  func(*model.Message) (*model.Message, error)
}

func (a *AIHelper) StreamResponseWithContext(username string, ctx context.Context, cb StreamCallback, question, knowledge string) (*model.Message, error) {
	a.streamMu.Lock()
	defer a.streamMu.Unlock()

	a.AddMessage(question, username, true, true)
	a.muRW.RLock()
	messages := utils.ConvertToSchemaMessages(a.messages)
	a.muRW.RUnlock()
	providerName := a.model.GetModelType()
	if providerName == ModelTypeZhipu {
		providerName = "智谱 GLM"
	} else if providerName == ModelTypeOpenAI {
		providerName = "ChatGPT / OpenAI 兼容服务"
	}
	formatInstruction := fmt.Sprintf("当前本次回答由“%s”提供，模型标识为“%s”。当用户询问你是什么模型、是否为 ChatGPT 或 GLM 时，必须按此配置如实回答；不得声称不知道、无法查看，或说自己是其他提供方。请使用清晰、简洁的 Markdown 回复；仅在确有多层结构时使用标题，普通问答不要使用 # 标题。所有标题、列表项和代码块必须各自从新行开始，标题标记 # 后必须有空格；不要把 ##、- 或数字列表接在正文或公式末尾。行内数学公式使用 $...$，独立公式使用 $$...$$。", providerName, a.model.GetModelName())
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
			data, genErr := rabbitmq.GenerateMessageMQPara(msg.MessageID, msg.SessionID, msg.Content, msg.UserName, msg.IsUser)
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
		MessageID: uuid.NewString(),
		SessionID: a.SessionID,
		Content:   Content,
		UserName:  UserName,
		IsUser:    IsUser,
	}
	a.muRW.Lock()
	a.messages = append(a.messages, &userMsg)
	a.messages = trimContextMessages(a.messages)
	a.muRW.Unlock()

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

func trimContextMessages(messages []*model.Message) []*model.Message {
	if len(messages) > MaxContextMessages {
		messages = messages[len(messages)-MaxContextMessages:]
	}
	// 上下文从用户提问开始，避免窗口截断后以助手回答开头。
	if len(messages) > 0 && !messages[0].IsUser {
		messages = messages[1:]
	}
	return messages
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

// RestoreMessages 在切换模型时保留已有会话上下文。
func (a *AIHelper) RestoreMessages(messages []*model.Message) {
	a.muRW.Lock()
	defer a.muRW.Unlock()
	a.messages = trimContextMessages(append([]*model.Message(nil), messages...))
}

// 流式生成
func (a *AIHelper) StreamResponse(username string,
	ctx context.Context, cb StreamCallback,
	userQuestion string) (*model.Message, error) {
	a.streamMu.Lock()
	defer a.streamMu.Unlock()

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
