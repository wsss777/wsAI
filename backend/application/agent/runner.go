package agent

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	memory "wsai/backend/application/memory"
	ragapp "wsai/backend/application/rag"
	sessionapp "wsai/backend/application/session"
	"wsai/backend/config"
	workspace "wsai/backend/infra/agent"
	ai "wsai/backend/infra/llm"
	agentrepo "wsai/backend/infra/mysql/repository"
	"wsai/backend/model"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

type Input struct {
	UserName, SessionID, Question, ModelType string
}
type Emit func(string, any)

func Run(ctx context.Context, in Input, emit Emit) (string, error) {

	if strings.TrimSpace(in.Question) == "" {
		return "", errors.New("question is required")
	}

	owned, e := sessionapp.SessionBelongsToUser(in.UserName, in.SessionID)
	if e != nil || !owned {
		return "", errors.New("session is not available")
	}
	provider, e := ai.NormalizeModelType(in.ModelType)
	if e != nil {
		return "", e
	}

	run := &model.AgentRun{UserName: in.UserName, SessionID: in.SessionID, ModelType: provider, Status: model.AgentRunRunning, Input: in.Question, StartedAt: time.Now()}
	if e = agentrepo.CreateAgentRun(run); e != nil {
		return "", e
	}
	input, e := memory.AppendUser(in.SessionID, in.UserName, in.Question, run.ID)
	if e != nil {
		return fail(run, emit, e)
	}
	_ = agentrepo.SetAgentInputMessageID(run.ID, input.MessageID)
	emit("status", map[string]any{"run_id": run.ID, "status": run.Status, "session_id": in.SessionID})

	root, e := workspace.NewWorkspace(config.C.AgentConfig.WorkspaceRoot)
	if e != nil {
		return fail(run, emit, e)
	}

	m, e := ai.GetGlobalFactory().CreateAIModel(ctx, provider, nil)
	if e != nil {
		return fail(run, emit, e)
	}
	chat := m

	msgs := []*schema.Message{{Role: schema.System, Content: "你是 WsAI 的只读开发 Agent。必须先通过工具确认代码事实；禁止声称修改文件、执行命令或访问工作区外内容。最终以简洁中文回答。"}}
	history, e := memory.Load(ctx, in.SessionID, in.UserName, memory.LoadOptions{
		Question: in.Question, MaxTokens: config.C.MemoryConfig.ContextMaxTokens, KeepRecentTurns: config.C.MemoryConfig.RecentTurns,
		Summarize: func(summaryCtx context.Context, previous, transcript string) (string, error) {
			return chat.GenerateResponse(summaryCtx, []*schema.Message{
				schema.SystemMessage("你是会话工作摘要器。合并旧摘要与新增对话，只保留目标、约束、决定、已验证结果、未完成事项和必要引用。不要保留闲聊，不要编造。使用简洁中文。"),
				schema.UserMessage("旧摘要：\n" + previous + "\n\n新增对话：\n" + transcript),
			})
		},
	})
	if e != nil {
		return fail(run, emit, e)
	}
	msgs = append(msgs, history...)

	max := config.C.AgentConfig.MaxSteps
	if max <= 0 {
		max = 8
	}
	tools := toolInfos()

	for i := 1; i <= max; i++ {
		msgs = memory.FitMessagesToBudget(msgs, config.C.MemoryConfig.ContextMaxTokens, 1)
		out, err := chat.GenerateWithTools(ctx, msgs, tools)
		if err != nil {
			return fail(run, emit, err)
		}

		if len(out.ToolCalls) == 0 {
			answer := strings.TrimSpace(out.Content)
			if answer == "" {
				return fail(run, emit, errors.New("model returned neither answer nor tool call"))
			}
			output, err := memory.AppendAssistant(in.SessionID, in.UserName, answer, run.ID)
			if err != nil {
				return fail(run, emit, err)
			}
			_ = agentrepo.FinishAgentRun(run.ID, model.AgentRunCompleted, answer, output.MessageID, "")
			emit("answer_delta", map[string]string{"content": answer})
			emit("done", map[string]any{"run_id": run.ID})
			return answer, nil
		}

		msgs = append(msgs, out)
		if _, err := memory.AppendToolCall(in.SessionID, in.UserName, run.ID, toolCallSummary(out.ToolCalls), out.ToolCalls); err != nil {
			return fail(run, emit, err)
		}
		for _, call := range out.ToolCalls {
			// 解析工具调用原因，用于记录和展示
			var callArgs struct {
				Reason string `json:"reason"`
			}
			_ = json.Unmarshal([]byte(call.Function.Arguments), &callArgs)
			if strings.TrimSpace(callArgs.Reason) == "" {
				callArgs.Reason = "模型请求调用 " + call.Function.Name
			}
			// 记录决策步骤
			decisionAt := time.Now()
			decision := &model.AgentStep{RunID: run.ID, Sequence: i, Type: "model_decision", Status: "completed", Summary: callArgs.Reason, StartedAt: decisionAt, FinishedAt: &decisionAt}
			_ = agentrepo.CreateAgentStep(decision)
			emit("status", map[string]any{"run_id": run.ID, "status": "model_decision", "summary": callArgs.Reason})
			// 记录工具开始执行步骤
			startedAt := time.Now()
			started := &model.AgentStep{RunID: run.ID, Sequence: i, Type: "tool_started", Status: "running", Summary: "正在调用 " + call.Function.Name, StartedAt: startedAt}
			_ = agentrepo.CreateAgentStep(started)
			toolCall := &model.AgentToolCall{RunID: run.ID, StepID: started.ID, ToolName: call.Function.Name, ArgumentsJSON: call.Function.Arguments, ResultSummary: "", StartedAt: startedAt}
			_ = agentrepo.CreateAgentToolCall(toolCall)
			emit("tool_call", map[string]string{"name": call.Function.Name, "summary": callArgs.Reason, "status": "running"})
			// 执行工具
			result, err := execute(ctx, root, in, call.Function.Name, call.Function.Arguments)
			summary := summary(result, err)
			now := time.Now()
			// 记录工具执行结果步骤
			step := &model.AgentStep{RunID: run.ID, Sequence: i, Type: "tool_finished", Status: "completed", Summary: summary, StartedAt: now, FinishedAt: &now}
			if err != nil {
				step.Status = "failed"
			}
			_ = agentrepo.CreateAgentStep(step)
			_ = agentrepo.FinishAgentToolCall(toolCall.ID, summary, errorString(err))
			// 将工具结果以 ToolMessage 形式追加到对话，供下一轮模型参考
			if err != nil {
				result = map[string]string{"error": err.Error()}
			}
			b, _ := json.Marshal(result)
			toolContent := string(b)
			memoryContent, artifactID, artifactErr := persistToolResult(run.ID, toolCall.ID, summary, toolContent)
			if artifactErr != nil {
				return fail(run, emit, artifactErr)
			}
			if _, err := memory.AppendToolResult(in.SessionID, in.UserName, call.Function.Name, call.ID, memoryContent, artifactID, run.ID); err != nil {
				return fail(run, emit, err)
			}
			emit("tool_result", map[string]string{"name": call.Function.Name, "summary": summary, "status": step.Status, "artifact_id": artifactID})
			msgs = append(msgs, schema.ToolMessage(memoryContent, call.ID, schema.WithToolName(call.Function.Name)))
		}
	}
	// 超过最大步数仍未给出最终答案，标记失败
	return fail(run, emit, errors.New("agent exceeded maximum steps"))
}
func fail(r *model.AgentRun, e Emit, err error) (string, error) {
	_ = agentrepo.FinishAgentRun(r.ID, model.AgentRunFailed, "", "", err.Error())
	e("error", map[string]any{"run_id": r.ID, "message": "Agent 执行失败：" + err.Error()})
	return "", err
}
func toolInfos() []*schema.ToolInfo {
	return []*schema.ToolInfo{
		{Name: "list_files", Desc: "列出工作区目录", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"reason": {Type: schema.String, Desc: "本次调用目的", Required: true}, "path": {Type: schema.String, Desc: "相对路径"}, "depth": {Type: schema.Integer, Desc: "深度"}})},
		{Name: "search_files", Desc: "搜索工作区文本文件", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"reason": {Type: schema.String, Desc: "本次调用目的", Required: true}, "query": {Type: schema.String, Desc: "搜索词", Required: true}, "path": {Type: schema.String, Desc: "相对路径"}})},
		{Name: "read_file", Desc: "读取 UTF-8 文本文件", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"reason": {Type: schema.String, Desc: "本次调用目的", Required: true}, "path": {Type: schema.String, Desc: "相对文件路径", Required: true}, "start_line": {Type: schema.Integer}, "end_line": {Type: schema.Integer}})},
		{Name: "search_knowledge", Desc: "检索当前会话绑定资料", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"reason": {Type: schema.String, Desc: "本次调用目的", Required: true}, "question": {Type: schema.String, Desc: "问题", Required: true}})},
		{Name: "read_artifact", Desc: "按片段读取之前工具调用归档的完整结果", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"reason": {Type: schema.String, Desc: "本次调用目的", Required: true}, "artifact_id": {Type: schema.String, Desc: "产物 ID", Required: true}, "offset": {Type: schema.Integer, Desc: "从第几个字符开始，默认 0"}})},
	}
}
func execute(ctx context.Context, w *workspace.Workspace, in Input, name, raw string) (any, error) {
	var a struct {
		Path, Query, Question, ArtifactID string
		Depth, StartLine, EndLine, Offset int
	}
	if e := json.Unmarshal([]byte(raw), &a); e != nil {
		return nil, errors.New("tool arguments must be valid JSON")
	}
	switch name {
	case "list_files":
		return w.List(a.Path, a.Depth)
	case "search_files":
		return w.Search(a.Query, a.Path)
	case "read_file":
		return w.Read(a.Path, a.StartLine, a.EndLine)
	case "search_knowledge":
		r, e := ragapp.RetrieveContext(ctx, in.UserName, in.SessionID, a.Question, in.ModelType)
		return map[string]any{"knowledge": r.Knowledge, "citations": r.Citations}, e
	case "read_artifact":
		artifact, e := agentrepo.GetAgentArtifact(in.UserName, a.ArtifactID)
		if e != nil {
			return nil, e
		}
		content := []rune(artifact.Content)
		if a.Offset < 0 || a.Offset > len(content) {
			return nil, errors.New("artifact offset is out of range")
		}
		maxRunes := config.C.MemoryConfig.MaxInlineToolResultBytes / 4
		if maxRunes <= 0 {
			maxRunes = 2048
		}
		end := a.Offset + maxRunes
		if end > len(content) {
			end = len(content)
		}
		return map[string]any{"artifact_id": artifact.ID, "kind": artifact.Kind, "content": string(content[a.Offset:end]), "offset": a.Offset, "next_offset": end, "has_more": end < len(content)}, nil
	}
	return nil, errors.New("tool is not allowed")
}

func toolCallSummary(calls []schema.ToolCall) string {
	names := make([]string, 0, len(calls))
	for _, call := range calls {
		names = append(names, call.Function.Name)
	}
	return "准备调用工具：" + strings.Join(names, "、")
}

func persistToolResult(runID, toolCallID uint, resultSummary, content string) (string, string, error) {
	limit := config.C.MemoryConfig.MaxInlineToolResultBytes
	if limit <= 0 {
		limit = 8 * 1024
	}
	if len(content) <= limit {
		return content, "", nil
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	artifact := &model.AgentArtifact{ID: uuid.NewString(), RunID: runID, Kind: "tool_result", ContentHash: hash, Content: content}
	if err := agentrepo.CreateAgentArtifact(artifact); err != nil {
		return "", "", err
	}
	if err := agentrepo.SetAgentToolCallArtifact(toolCallID, artifact.ID); err != nil {
		return "", "", err
	}
	return resultSummary + "\n完整结果已归档，artifact_id=" + artifact.ID, artifact.ID, nil
}

func summary(v any, e error) string {
	if e != nil {
		return "工具执行失败：" + e.Error()
	}
	b, _ := json.Marshal(v)
	s := string(b)
	if len(s) > 240 {
		s = s[:240] + "…"
	}
	return "工具执行完成：" + s
}
func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
