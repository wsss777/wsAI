package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	ai "wsai/backend/infra/llm"
	"wsai/backend/model"

	"github.com/cloudwego/eino/schema"
)

const queryPlannerTimeout = 20 * time.Second

// QueryPlan 描述一次资料库检索的决策。SearchQuery 仅用于召回，不替代用户原始问题。
type QueryPlan struct {
	NeedRetrieval bool   `json:"need_retrieval"`
	SearchQuery   string `json:"search_query"`
	Planner       string `json:"-"`
}

type queryPlanPayload struct {
	NeedRetrieval bool   `json:"need_retrieval"`
	SearchQuery   string `json:"search_query"`
}

// PrepareQuery 基于当前问题和有限会话上下文，决定是否检索并生成用于向量召回的 Query。
// 模型规划失败时返回原问题检索，避免预处理故障使知识库能力失效。
func PrepareQuery(ctx context.Context, modelType, question string, history []model.Message) (QueryPlan, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return QueryPlan{NeedRetrieval: false, Planner: "rule"}, nil
	}
	if isCasualQuestion(question) {
		return QueryPlan{NeedRetrieval: false, SearchQuery: question, Planner: "rule"}, nil
	}

	planCtx, cancel := context.WithTimeout(ctx, queryPlannerTimeout)
	defer cancel()
	model, err := ai.GetGlobalFactory().CreateAIModel(planCtx, modelType, nil)
	if err != nil {
		return fallbackPlan(question), fmt.Errorf("create query planner model: %w", err)
	}
	content, err := model.GenerateResponse(planCtx, []*schema.Message{
		{Role: schema.System, Content: `你是知识库检索规划器。根据用户当前问题和最近对话，判断是否必须查询已绑定的资料。
只在问题需要资料事实、用户要求依据资料回答、或需要解析上文指代时检索。闲聊、问候、表达感谢、询问模型身份不检索。
若检索，将“这/它/第二种方案”等指代补全为独立、具体、适合向量检索的一句话；不得编造资料中不存在的内容。
只能输出 JSON：{"need_retrieval":true,"search_query":"..."}。不要 Markdown、解释或其他字段。`},
		{Role: schema.User, Content: buildPlannerInput(question, history)},
	})
	if err != nil {
		return fallbackPlan(question), fmt.Errorf("generate query plan: %w", err)
	}
	plan, err := parseQueryPlan(content, question)
	if err != nil {
		return fallbackPlan(question), err
	}
	plan.Planner = "model"
	return plan, nil
}

func fallbackPlan(question string) QueryPlan {
	return QueryPlan{NeedRetrieval: true, SearchQuery: question, Planner: "fallback"}
}

func parseQueryPlan(content, originalQuestion string) (QueryPlan, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(strings.TrimSpace(content), "```")
	var payload queryPlanPayload
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return QueryPlan{}, fmt.Errorf("parse query plan JSON: %w", err)
	}
	payload.SearchQuery = strings.TrimSpace(payload.SearchQuery)
	if payload.NeedRetrieval && payload.SearchQuery == "" {
		payload.SearchQuery = originalQuestion
	}
	if len([]rune(payload.SearchQuery)) > 500 {
		payload.SearchQuery = string([]rune(payload.SearchQuery)[:500])
	}
	return QueryPlan{NeedRetrieval: payload.NeedRetrieval, SearchQuery: payload.SearchQuery}, nil
}

func buildPlannerInput(question string, history []model.Message) string {
	const maxHistoryMessages = 6
	const maxMessageRunes = 500
	start := len(history) - maxHistoryMessages
	if start < 0 {
		start = 0
	}
	var builder strings.Builder
	builder.WriteString("最近对话：\n")
	for _, message := range history[start:] {
		role := "助手"
		if message.IsUser {
			role = "用户"
		}
		content := []rune(strings.TrimSpace(message.Content))
		if len(content) > maxMessageRunes {
			content = content[:maxMessageRunes]
		}
		fmt.Fprintf(&builder, "%s：%s\n", role, string(content))
	}
	fmt.Fprintf(&builder, "当前问题：%s", question)
	return builder.String()
}

func isCasualQuestion(question string) bool {
	normalized := strings.ToLower(strings.TrimSpace(question))
	for _, casual := range []string{"你好", "您好", "嗨", "hi", "hello", "谢谢", "感谢", "再见", "你是谁", "你是什么模型"} {
		if normalized == casual {
			return true
		}
	}
	return false
}
