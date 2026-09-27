package ai

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	memory "wsai/backend/application/memory"
	"wsai/backend/config"
	"wsai/backend/infra/logger"
	"wsai/backend/model"

	"github.com/cloudwego/eino/schema"

	"go.uber.org/zap"
)

// AIHelper 表示与单个会话绑定的助手。
type AIHelper struct {
	model     AIModel
	streamMu  sync.Mutex
	SessionID string
}

func (a *AIHelper) StreamResponseWithContext(username string, ctx context.Context, cb StreamCallback, question, knowledge string) (*model.Message, error) {
	a.streamMu.Lock()
	defer a.streamMu.Unlock()

	if _, err := memory.AppendUser(a.SessionID, username, question, 0); err != nil {
		return nil, err
	}
	messages, err := a.loadConversation(ctx, username, question)
	if err != nil {
		return nil, err
	}
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
	messages = memory.FitMessagesToBudget(messages, config.C.MemoryConfig.ContextMaxTokens, 1)

	modelStart := time.Now()
	var firstChunkOnce sync.Once
	var ttft time.Duration
	measuredCallback := func(msg string) {
		if msg != "" {
			firstChunkOnce.Do(func() {
				ttft = time.Since(modelStart)
			})
		}
		cb(msg)
	}
	content, err := a.model.StreamResponse(ctx, messages, measuredCallback)
	ttftMS := int64(-1)
	if ttft > 0 {
		ttftMS = ttft.Milliseconds()
	}
	metrics := []zap.Field{
		zap.String("session_id", a.SessionID),
		zap.String("username", username),
		zap.String("model_type", a.model.GetModelType()),
		zap.String("model_name", a.model.GetModelName()),
		zap.Int64("ttft_ms", ttftMS),
		zap.Int64("completion_ms", time.Since(modelStart).Milliseconds()),
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			logger.L().Info("AI 流式响应指标", append(metrics,
				zap.String("result", "canceled"),
			)...)
			return nil, err
		}
		logger.L().Warn("AI 流式响应指标", append(metrics,
			zap.String("result", "error"),
			zap.Error(err),
		)...)
		return nil, err
	}
	logger.L().Info("AI 流式响应指标", append(metrics, zap.String("result", "success"))...)
	if _, err := memory.AppendAssistant(a.SessionID, username, content, 0); err != nil {
		return nil, err
	}
	return &model.Message{SessionID: a.SessionID, Content: content, UserName: username, IsUser: false}, nil
}

// NewAIHelper 创建新的 AIHelper 实例。
func NewAIHelper(model_ AIModel, SessionID string) *AIHelper {
	return &AIHelper{
		model:     model_,
		SessionID: SessionID,
	}
}

// 流式生成
func (a *AIHelper) StreamResponse(username string,
	ctx context.Context, cb StreamCallback,
	userQuestion string) (*model.Message, error) {
	a.streamMu.Lock()
	defer a.streamMu.Unlock()

	if _, err := memory.AppendUser(a.SessionID, username, userQuestion, 0); err != nil {
		return nil, err
	}
	messages, err := a.loadConversation(ctx, username, userQuestion)
	if err != nil {
		return nil, err
	}
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
	if _, err := memory.AppendAssistant(a.SessionID, username, content, 0); err != nil {
		return nil, err
	}

	return modelMsg, nil

}

func (a *AIHelper) loadConversation(ctx context.Context, username, question string) ([]*schema.Message, error) {
	return memory.Load(ctx, a.SessionID, username, memory.LoadOptions{
		Question: question, MaxTokens: config.C.MemoryConfig.ContextMaxTokens, KeepRecentTurns: config.C.MemoryConfig.RecentTurns,
		Summarize: func(summaryCtx context.Context, previous, transcript string) (string, error) {
			return a.model.GenerateResponse(summaryCtx, []*schema.Message{
				schema.SystemMessage("你是会话工作摘要器。合并旧摘要与新增对话，只保留目标、约束、决定、已验证结果、未完成事项和必要引用。不要保留闲聊，不要编造。使用简洁中文。"),
				schema.UserMessage("旧摘要：\n" + previous + "\n\n新增对话：\n" + transcript),
			})
		},
	})
}

// 获取当前使用的Ai模型
func (a *AIHelper) GetModelType() string {
	return a.model.GetModelType()
}
