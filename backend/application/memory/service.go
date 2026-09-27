package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"wsai/backend/infra/mysql/repository"
	"wsai/backend/model"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

const (
	DefaultMaxContextTokens = 120000 //120k
	DefaultRecentTurns      = 4
	maxLongTermMemories     = 5
)

type Summarizer func(ctx context.Context, previousSummary, transcript string) (string, error)

type LoadOptions struct {
	MaxTokens       int
	KeepRecentTurns int
	Question        string
	Summarize       Summarizer
}

func AppendUser(sessionID, username, content string, runID uint) (*model.Message, error) {
	item, err := appendItem(&model.Message{MessageID: uuid.NewString(), SessionID: sessionID, UserName: username, Content: content, Role: model.MessageRoleUser, Visibility: model.MessageVisibilityUser, AgentRunID: runID})
	if err == nil {
		_ = CaptureExplicitLongTermMemory(username, sessionID, item.MessageID, content)
	}
	return item, err
}

func AppendAssistant(sessionID, username, content string, runID uint) (*model.Message, error) {
	return appendItem(&model.Message{MessageID: uuid.NewString(), SessionID: sessionID, UserName: username, Content: content, Role: model.MessageRoleAssistant, Visibility: model.MessageVisibilityUser, AgentRunID: runID})
}

func AppendToolCall(sessionID, username string, runID uint, content string, calls []schema.ToolCall) (*model.Message, error) {
	b, err := json.Marshal(calls)
	if err != nil {
		return nil, err
	}
	return appendItem(&model.Message{MessageID: uuid.NewString(), SessionID: sessionID, UserName: username, Content: content, Role: model.MessageRoleAssistant, ToolCallsJSON: string(b), Visibility: model.MessageVisibilityTrace, AgentRunID: runID})
}

func AppendToolResult(sessionID, username, name, callID, content, artifactID string, runID uint) (*model.Message, error) {
	return appendItem(&model.Message{MessageID: uuid.NewString(), SessionID: sessionID, UserName: username, Content: content, Role: model.MessageRoleTool, ToolName: name, ToolCallID: callID, ArtifactID: artifactID, Visibility: model.MessageVisibilityTrace, AgentRunID: runID})
}

func appendItem(item *model.Message) (*model.Message, error) {
	return repository.AppendConversationItem(item)
}

// Load 保留完整的最近轮次，并在超过词元预算时压缩更早的完整轮次。
func Load(ctx context.Context, sessionID, username string, options LoadOptions) ([]*schema.Message, error) {
	if options.MaxTokens <= 0 {
		options.MaxTokens = DefaultMaxContextTokens
	}
	if options.KeepRecentTurns <= 0 {
		options.KeepRecentTurns = DefaultRecentTurns
	}
	summary, err := repository.GetSessionMemory(sessionID)
	if err != nil {
		return nil, err
	}
	items, err := repository.ListMessagesAfter(sessionID, summary.CoveredSequenceNo)
	if err != nil {
		return nil, err
	}
	start := chooseTailStart(items, summary.Summary, options.MaxTokens, options.KeepRecentTurns)
	if start > 0 {
		if options.Summarize != nil {
			prefix := items[:start]
			updated, summarizeErr := options.Summarize(ctx, summary.Summary, transcript(prefix))
			if updated = strings.TrimSpace(updated); summarizeErr == nil && updated != "" {
				summary.Summary = truncateToTokens(updated, summaryTokenBudget(options.MaxTokens))
				summary.CoveredSequenceNo = prefix[len(prefix)-1].SequenceNo
				if err := repository.SaveSessionMemory(summary); err != nil {
					return nil, err
				}
			}
		}
		// 即使摘要暂时失败，本次请求仍须遵守严格的输入预算。
		// CoveredSequenceNo 不会推进，被省略的前缀会在下次重试。
		items = items[start:]
	}
	result := make([]*schema.Message, 0, len(items)+2)
	if summary.Summary != "" {
		result = append(result, schema.SystemMessage("以下是更早会话的工作摘要，仅用于保持上下文连续：\n"+summary.Summary))
	}
	memories, err := relevantLongTermMemories(username, options.Question)
	if err != nil {
		return nil, err
	}
	if len(memories) > 0 {
		lines, ids := make([]string, 0, len(memories)), make([]string, 0, len(memories))
		for _, item := range memories {
			lines = append(lines, "- "+item.Content)
			ids = append(ids, item.ID)
		}
		_ = repository.TouchLongTermMemories(ids)
		result = append(result, schema.SystemMessage("用户明确要求跨会话保留的记忆：\n"+strings.Join(lines, "\n")))
	}
	for _, item := range items {
		result = append(result, toSchemaMessage(item))
	}
	return fitMessagesToBudget(result, options.MaxTokens), nil
}

func toSchemaMessage(item model.Message) *schema.Message {
	role := schema.RoleType(item.Role)
	if role == "" {
		if item.IsUser {
			role = schema.User
		} else {
			role = schema.Assistant
		}
	}
	message := &schema.Message{Role: role, Content: item.Content, ToolCallID: item.ToolCallID, ToolName: item.ToolName}
	if item.ToolCallsJSON != "" {
		_ = json.Unmarshal([]byte(item.ToolCallsJSON), &message.ToolCalls)
	}
	return message
}

func chooseTailStart(items []model.Message, summary string, budget, keepTurns int) int {
	if estimateTokens(summary)+estimateItems(items) <= budget {
		return 0
	}
	// 开始压缩后，为滚动摘要预留稳定的上下文份额，避免摘要增长挤占完整轮次。
	tailBudget := budget - summaryTokenBudget(budget)
	if existing := estimateTokens(summary); existing > summaryTokenBudget(budget) {
		tailBudget = budget - existing
	}
	if tailBudget < 1 {
		tailBudget = 1
	}
	starts := make([]int, 0)
	for i, item := range items {
		if roleOf(item) == model.MessageRoleUser {
			starts = append(starts, i)
		}
	}
	if len(starts) <= 1 {
		return 0
	}
	startIndex := len(starts) - keepTurns
	if startIndex < 1 {
		startIndex = 1
	}
	for startIndex < len(starts)-1 && estimateItems(items[starts[startIndex]:]) > tailBudget {
		startIndex++
	}
	return starts[startIndex]
}

func summaryTokenBudget(contextBudget int) int {
	if contextBudget < 4 {
		return 1
	}
	return contextBudget / 4
}

func fitMessagesToBudget(messages []*schema.Message, budget int) []*schema.Message {
	return FitMessagesToBudget(messages, budget, 0)
}

// FitMessagesToBudget 为执行中的 Agent 循环应用相同的严格预算。
// protectedPrefix 会优先保留开头的系统指令，直到普通会话内容都已裁剪。
func FitMessagesToBudget(messages []*schema.Message, budget, protectedPrefix int) []*schema.Message {
	if budget <= 0 {
		return messages
	}
	if protectedPrefix < 0 {
		protectedPrefix = 0
	}
	if protectedPrefix > len(messages) {
		protectedPrefix = len(messages)
	}
	order := make([]int, 0, len(messages))
	for index := protectedPrefix; index < len(messages); index++ {
		order = append(order, index)
	}
	for index := 0; index < protectedPrefix; index++ {
		order = append(order, index)
	}
	total := estimateSchemaMessages(messages)
	for _, messageIndex := range order {
		if total <= budget {
			break
		}
		message := messages[messageIndex]
		before := estimateTokens(message.Content)
		if before == 0 {
			continue
		}
		keep := before - (total - budget)
		if keep < 0 {
			keep = 0
		}
		message.Content = truncateToTokens(message.Content, keep)
		total -= before - estimateTokens(message.Content)
	}
	// 工具参数通常较小。若异常调用仍超出严格窗口，则保留调用标识和配对关系，仅替换参数。
	for _, messageIndex := range order {
		if total <= budget {
			break
		}
		message := messages[messageIndex]
		for index := range message.ToolCalls {
			before := estimateTokens(message.ToolCalls[index].Function.Arguments)
			if before == 0 {
				continue
			}
			message.ToolCalls[index].Function.Arguments = `{"context_trimmed":true}`
			total -= before - estimateTokens(message.ToolCalls[index].Function.Arguments)
		}
	}
	return messages
}

func estimateSchemaMessages(messages []*schema.Message) int {
	total := 0
	for _, message := range messages {
		total += estimateTokens(message.Content) + 8
		for _, call := range message.ToolCalls {
			total += estimateTokens(call.Function.Name) + estimateTokens(call.Function.Arguments) + 4
		}
	}
	return total
}

func truncateToTokens(value string, budget int) string {
	if budget <= 0 {
		return ""
	}
	if estimateTokens(value) <= budget {
		return value
	}
	runes := []rune(value)
	low, high := 0, len(runes)
	for low < high {
		mid := (low + high + 1) / 2
		if estimateTokens(string(runes[:mid])) <= budget {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return string(runes[:low])
}

func estimateItems(items []model.Message) int {
	total := 0
	for _, item := range items {
		total += estimateTokens(item.Content) + estimateTokens(item.ToolCallsJSON) + 8
	}
	return total
}

func estimateTokens(value string) int {
	ascii, other := 0, 0
	for _, r := range value {
		if r <= unicode.MaxASCII {
			ascii++
		} else {
			other++
		}
	}
	return (ascii+3)/4 + other
}

func roleOf(item model.Message) string {
	if item.Role != "" {
		return item.Role
	}
	if item.IsUser {
		return model.MessageRoleUser
	}
	return model.MessageRoleAssistant
}

func transcript(items []model.Message) string {
	var b strings.Builder
	for _, item := range items {
		fmt.Fprint(&b, roleOf(item))
		if item.ToolName != "" {
			fmt.Fprintf(&b, "(%s)", item.ToolName)
		}
		fmt.Fprintf(&b, ": %s\n", item.Content)
	}
	return b.String()
}

// CaptureExplicitLongTermMemory 只接受明确的“记住”或“忘记”指令。
func CaptureExplicitLongTermMemory(username, sessionID, messageID, input string) error {
	text := strings.TrimSpace(input)
	for _, prefix := range []string{"请记住：", "请记住:", "记住：", "记住:", "remember:"} {
		if strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix)) {
			content := strings.TrimSpace(text[len(prefix):])
			if content == "" {
				return nil
			}
			existing, err := repository.ListActiveLongTermMemories(username)
			if err != nil {
				return err
			}
			for _, item := range existing {
				if strings.EqualFold(item.Content, content) {
					return nil
				}
			}
			category := "fact"
			if strings.Contains(content, "偏好") || strings.Contains(content, "喜欢") || strings.Contains(content, "以后") {
				category = "preference"
			}
			memoryKey, normalized := parseMemoryKey(content)
			if err := repository.SupersedeLongTermMemories(username, category, memoryKey); err != nil {
				return err
			}
			return repository.CreateLongTermMemory(&model.LongTermMemory{ID: uuid.NewString(), UserName: username, Category: category, MemoryKey: memoryKey, Content: normalized, Status: model.LongTermMemoryActive, SourceSessionID: sessionID, SourceMessageID: messageID})
		}
	}
	for _, prefix := range []string{"请忘记：", "请忘记:", "忘记：", "忘记:", "forget:"} {
		if strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix)) {
			content := strings.TrimSpace(text[len(prefix):])
			if content != "" {
				_, err := repository.DeleteLongTermMemories(username, content)
				return err
			}
		}
	}
	return nil
}

func parseMemoryKey(content string) (string, string) {
	for _, separator := range []string{"=", "：", ":"} {
		if parts := strings.SplitN(content, separator, 2); len(parts) == 2 {
			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if key != "" && value != "" {
				return key, key + "：" + value
			}
		}
	}
	return "", content
}

func relevantLongTermMemories(username, question string) ([]model.LongTermMemory, error) {
	items, err := repository.ListActiveLongTermMemories(username)
	if err != nil {
		return nil, err
	}
	type scored struct {
		item  model.LongTermMemory
		score int
	}
	words := searchTerms(strings.ToLower(question))
	scores := make([]scored, 0, len(items))
	for _, item := range items {
		score, content := 0, strings.ToLower(item.Content)
		for _, word := range words {
			if len([]rune(word)) >= 2 && strings.Contains(content, word) {
				score++
			}
		}
		if item.Category == "preference" {
			score++
		}
		if score > 0 || strings.TrimSpace(question) == "" {
			scores = append(scores, scored{item, score})
		}
	}
	sort.SliceStable(scores, func(i, j int) bool { return scores[i].score > scores[j].score })
	if len(scores) > maxLongTermMemories {
		scores = scores[:maxLongTermMemories]
	}
	result := make([]model.LongTermMemory, len(scores))
	for i := range scores {
		result[i] = scores[i].item
	}
	return result, nil
}

func searchTerms(question string) []string {
	terms := strings.FieldsFunc(question, func(r rune) bool { return unicode.IsSpace(r) || strings.ContainsRune("，。！？,.!?;；:/", r) })
	runes := []rune(strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Han, r) {
			return r
		}
		return ' '
	}, question))
	for _, segment := range strings.Fields(string(runes)) {
		chars := []rune(segment)
		for i := 0; i+1 < len(chars); i++ {
			terms = append(terms, string(chars[i:i+2]))
		}
	}
	return terms
}
