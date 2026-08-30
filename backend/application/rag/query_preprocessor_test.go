package rag

import "testing"

func TestParseQueryPlan(t *testing.T) {
	plan, err := parseQueryPlan("```json\n{\"need_retrieval\":true,\"search_query\":\"JWT Refresh Token 轮换方案\"}\n```", "原问题")
	if err != nil {
		t.Fatalf("parseQueryPlan returned error: %v", err)
	}
	if !plan.NeedRetrieval || plan.SearchQuery != "JWT Refresh Token 轮换方案" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}

func TestPrepareQuerySkipsCasualQuestion(t *testing.T) {
	plan, err := PrepareQuery(t.Context(), "openai", "你好", nil)
	if err != nil {
		t.Fatalf("PrepareQuery returned error: %v", err)
	}
	if plan.NeedRetrieval || plan.Planner != "rule" {
		t.Fatalf("casual question should skip retrieval: %#v", plan)
	}
}
