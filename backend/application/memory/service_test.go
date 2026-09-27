package memory

import (
	"strings"
	"testing"
	"wsai/backend/model"

	"github.com/cloudwego/eino/schema"
)

func TestChooseTailStartKeepsCompleteRecentTurns(t *testing.T) {
	items := []model.Message{
		{Role: model.MessageRoleUser, Content: strings.Repeat("旧", 80)},
		{Role: model.MessageRoleAssistant, ToolCallsJSON: strings.Repeat("x", 80)},
		{Role: model.MessageRoleTool, Content: strings.Repeat("结果", 40)},
		{Role: model.MessageRoleAssistant, Content: "旧回答"},
		{Role: model.MessageRoleUser, Content: strings.Repeat("中", 80)},
		{Role: model.MessageRoleAssistant, Content: "中回答"},
		{Role: model.MessageRoleUser, Content: "最新问题"},
	}
	start := chooseTailStart(items, "", 40, 2)
	if start == 0 || roleOf(items[start]) != model.MessageRoleUser {
		t.Fatalf("compression must start on a user-turn boundary, got %d", start)
	}
	for i := start; i < len(items); i++ {
		if roleOf(items[i]) == model.MessageRoleTool && i == start {
			t.Fatal("tool result was separated from its assistant tool call")
		}
	}
}

func TestChooseTailStartDoesNotCompressWithinBudget(t *testing.T) {
	items := []model.Message{{Role: model.MessageRoleUser, Content: "你好"}, {Role: model.MessageRoleAssistant, Content: "你好"}}
	if got := chooseTailStart(items, "", 100, 4); got != 0 {
		t.Fatalf("unexpected compression start: %d", got)
	}
}

func TestParseMemoryKey(t *testing.T) {
	key, content := parseMemoryKey("回复语言=简体中文")
	if key != "回复语言" || content != "回复语言：简体中文" {
		t.Fatalf("unexpected parsed memory: %q %q", key, content)
	}
	key, content = parseMemoryKey("我喜欢简洁回答")
	if key != "" || content != "我喜欢简洁回答" {
		t.Fatalf("unstructured memory should remain unchanged: %q %q", key, content)
	}
}

func TestSearchTermsIncludesChineseBigrams(t *testing.T) {
	terms := searchTerms("请用简体中文回答")
	found := false
	for _, term := range terms {
		if term == "中文" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Chinese bigram not generated: %#v", terms)
	}
}

func TestFitMessagesToBudgetIsHardLimit(t *testing.T) {
	messages := []*schema.Message{
		schema.SystemMessage(strings.Repeat("摘要", 50)),
		schema.UserMessage(strings.Repeat("问题", 50)),
	}
	got := fitMessagesToBudget(messages, 40)
	if tokens := estimateSchemaMessages(got); tokens > 40 {
		t.Fatalf("context exceeds hard budget: %d", tokens)
	}
}
