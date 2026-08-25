package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	once sync.Once
	C    *Config
)

type Config struct {
	App struct {
		Name           string `mapstructure:"name"`
		Host           string `mapstructure:"host"`
		Port           int    `mapstructure:"port"`
		Env            string `mapstructure:"env"`
		EnableRegister bool   `mapstructure:"enable_register"`
	} `mapstructure:"app"`

	JWTConfig struct {
		Secret     string `mapstructure:"secret"`
		AccessTTL  string `mapstructure:"access_ttl"`
		RefreshTTL string `mapstructure:"refresh_ttl"`
		Issuer     string `mapstructure:"issuer"`
		Subject    string `mapstructure:"subject"`
	} `mapstructure:"jwt"`

	MysqlConfig struct {
		Host         string `mapstructure:"host"`
		Port         int    `mapstructure:"port"`
		User         string `mapstructure:"user"`
		Password     string `mapstructure:"password"`
		Database     string `mapstructure:"databaseName"`
		Charset      string `mapstructure:"charset"`
		MaxIdleConns int    `mapstructure:"maxIdleConns"`
		MaxOpenConns int    `mapstructure:"maxOpenConns"`
	} `mapstructure:"mysql"`

	RedisConfig struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`

	RabbitmqConfig struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Vhost    string `mapstructure:"vhost"`
	} `mapstructure:"rabbitmq"`

	EmailConfig struct {
		Email    string `mapstructure:"email"`
		Authcode string `mapstructure:"authcode"`
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		IsSSL    bool   `mapstructure:"is_ssl"`
	} `mapstructure:"email"`
	RAGConfig struct {
		QdrantURL      string `mapstructure:"qdrant_url"`
		Collection     string `mapstructure:"collection"`
		EmbeddingModel string `mapstructure:"embedding_model"`
		EmbeddingSize  int    `mapstructure:"embedding_size"`
		TopK           int    `mapstructure:"top_k"`
	} `mapstructure:"rag"`
	ChatConfig struct {
		ModelProvider string `mapstructure:"model_provider"`
	} `mapstructure:"chat"`
	OpenAIConfig struct {
		APIKey  string `mapstructure:"api_key"`
		BaseURL string `mapstructure:"base_url"`
		Model   string `mapstructure:"model_name"`
	} `mapstructure:"openai"`
	ZhipuConfig struct {
		APIKey    string `mapstructure:"api_key"`
		BaseURL   string `mapstructure:"base_url"`
		ChatModel string `mapstructure:"chat_model"`
	} `mapstructure:"zhipu"`
}

func InitConfig() {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Printf("加载 .env 文件失败: %v\n", err)
		} else {
			log.Println("成功加载 .env 文件")
		}

		v := viper.New()
		v.SetConfigFile("./backend/config/config.default.toml")
		if err := v.ReadInConfig(); err != nil {
			log.Fatalf("加载默认配置失败: %v", err)
		}

		// 代码默认值只作为 config.default.toml 缺失字段时的兜底。
		v.SetDefault("app.name", "wsai")
		v.SetDefault("app.env", "dev")
		v.SetDefault("app.port", "9091")
		v.SetDefault("jwt.access_ttl", "30m")
		v.SetDefault("jwt.refresh_ttl", "720h")
		v.SetDefault("rag.qdrant_url", "http://127.0.0.1:6333")
		v.SetDefault("rag.collection", "wsai_doc_chunks")
		v.SetDefault("rag.embedding_model", "embedding-3")
		v.SetDefault("rag.embedding_size", 1024)
		v.SetDefault("rag.top_k", 5)

		C = new(Config)
		if err := v.Unmarshal(C); err != nil {
			log.Fatalf("解析配置失败: %v", err)
		}
		applyEnvironment(C)
		log.Printf("配置初始化完成 | 环境: %s | 应用: %s", C.App.Env, C.App.Name)
	})

}

// applyEnvironment 仅覆盖 .env/进程环境中的敏感项和聊天提供方配置。
// 先反序列化完整 TOML，再逐字段覆盖，可避免 Viper 绑定嵌套字段时覆盖整组配置。
func applyEnvironment(c *Config) {
	override := func(key string, set func(string)) {
		if value, ok := os.LookupEnv(key); ok {
			set(value)
		}
	}

	override("JWT_SECRET", func(value string) { c.JWTConfig.Secret = value })
	override("MYSQL_PASSWORD", func(value string) { c.MysqlConfig.Password = value })
	override("REDIS_PASSWORD", func(value string) { c.RedisConfig.Password = value })
	override("RABBITMQ_PASSWORD", func(value string) { c.RabbitmqConfig.Password = value })
	override("EMAIL_EMAIL", func(value string) { c.EmailConfig.Email = value })
	override("EMAIL_AUTHCODE", func(value string) { c.EmailConfig.Authcode = value })

	override("CHAT_MODEL_PROVIDER", func(value string) { c.ChatConfig.ModelProvider = value })
	override("OPENAI_API_KEY", func(value string) { c.OpenAIConfig.APIKey = value })
	override("OPENAI_BASE_URL", func(value string) { c.OpenAIConfig.BaseURL = value })
	override("OPENAI_MODEL_NAME", func(value string) { c.OpenAIConfig.Model = value })
	override("ZHIPU_API_KEY", func(value string) { c.ZhipuConfig.APIKey = value })
	override("ZHIPU_BASE_URL", func(value string) { c.ZhipuConfig.BaseURL = value })
	override("ZHIPU_CHAT_MODEL", func(value string) { c.ZhipuConfig.ChatModel = value })
}
