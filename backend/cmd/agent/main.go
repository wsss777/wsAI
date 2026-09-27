package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	agentapp "wsai/backend/application/agent"
	sessionapp "wsai/backend/application/session"
	"wsai/backend/config"
	"wsai/backend/infra/logger"
	"wsai/backend/infra/mysql"
	"wsai/backend/infra/rabbitmq"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "agent" {
		args = args[1:]
	}
	if len(args) < 1 || args[0] != "run" {
		fmt.Fprintln(os.Stderr, "用法: wsai agent run --user <开发用户名> [--model openai|zhipu|ollama]")
		os.Exit(2)
	}
	flags := flag.NewFlagSet("run", flag.ExitOnError)
	user := flags.String("user", "", "开发环境用户标识")
	modelType := flags.String("model", "openai", "模型: openai、zhipu 或 ollama")
	_ = flags.Parse(args[1:])
	config.InitConfig()
	if config.C.App.Env == "prod" {
		fmt.Fprintln(os.Stderr, "CLI 开发身份不允许在生产环境使用")
		os.Exit(1)
	}
	if strings.TrimSpace(*user) == "" {
		fmt.Fprintln(os.Stderr, "必须通过 --user 指定开发身份")
		os.Exit(2)
	}
	if err := logger.Init(false); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer logger.L().Sync()
	if err := mysql.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "MySQL 初始化失败:", err)
		os.Exit(1)
	}
	defer mysql.Close()
	rabbitmq.InitRabbitMQ()
	defer rabbitmq.DestroyRabbitMQ()
	run(*user, *modelType)
}

func run(user, initialModel string) {
	sessionID, code := sessionapp.CreateStreamSessionOnly(user, "CLI Agent 会话")
	if code != 1000 {
		fmt.Fprintln(os.Stderr, "创建会话失败")
		return
	}
	modelType := initialModel
	reader := bufio.NewScanner(os.Stdin)
	reader.Buffer(make([]byte, 4096), 128*1024)
	fmt.Printf("WsAI Agent · 模型 %s · 工作区 %s\n输入 /help 查看命令，/exit 退出。\n\n", modelType, config.C.AgentConfig.WorkspaceRoot)
	for {
		fmt.Print("> ")
		if !reader.Scan() {
			fmt.Println()
			return
		}
		input := strings.TrimSpace(reader.Text())
		if input == "" {
			continue
		}
		if strings.HasPrefix(input, "/") {
			if handleCommand(input, &sessionID, &modelType, &user) {
				return
			}
			continue
		}
		fmt.Println("[Agent] 开始执行")
		_, err := agentapp.Run(context.Background(), agentapp.Input{UserName: user, SessionID: sessionID, Question: input, ModelType: modelType}, printEvent)
		if err != nil {
			fmt.Fprintln(os.Stderr, "[错误]", err)
			continue
		}
	}
}

func handleCommand(input string, sessionID, modelType, user *string) bool {
	parts := strings.Fields(input)
	switch parts[0] {
	case "/exit", "/quit":
		return true
	case "/help":
		fmt.Println("/model [openai|zhipu|ollama]  切换模型\n/new  新建会话\n/status  查看当前会话\n/exit  退出")
	case "/status":
		fmt.Printf("会话 %s · 模型 %s\n", *sessionID, *modelType)
	case "/new":
		id, code := sessionapp.CreateStreamSessionOnly(*user, "CLI Agent 会话")
		if code != 1000 {
			fmt.Println("创建会话失败")
			return false
		}
		*sessionID = id
		fmt.Println("已新建会话", id)
	case "/model":
		if len(parts) == 1 {
			fmt.Printf("当前模型: %s；可选: openai, zhipu, ollama\n", *modelType)
			return false
		}
		candidate := strings.ToLower(parts[1])
		if candidate != "openai" && candidate != "zhipu" && candidate != "ollama" {
			fmt.Println("不支持的模型，可选: openai, zhipu, ollama")
			return false
		}
		*modelType = candidate
		fmt.Println("后续任务将使用模型", candidate)
	default:
		fmt.Println("未知命令，输入 /help 查看可用命令")
	}
	return false
}

func printEvent(event string, data any) {
	switch event {
	case "status":
		if value, ok := data.(map[string]string); ok && value["summary"] != "" {
			fmt.Println("[决策]", value["summary"])
		}
	case "tool_call":
		if value, ok := data.(map[string]string); ok {
			fmt.Println("[工具]", value["name"], "开始")
		}
	case "tool_result":
		if value, ok := data.(map[string]string); ok {
			fmt.Println("[工具]", value["name"], value["status"])
		}
	case "answer_delta":
		if value, ok := data.(map[string]string); ok {
			fmt.Println("\n" + value["content"] + "\n")
		}
	}
}
