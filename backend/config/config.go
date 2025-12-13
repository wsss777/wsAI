package config

import (
	"log"
	"os"
	"sync"

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
	} `mapstructuer:"jwt"`

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
}

func InitConfig() {
	once.Do(func() {
		v := viper.New()
		v.SetConfigType("toml")
		// 自动选择配置文件
		env := os.Getenv("APP_ENV")
		switch env {
		case "production", "prod":
			v.SetConfigName("config.prod")
		case "test":
			v.SetConfigName("config.test")
		default:
			v.SetConfigName("config.dev") // 本地开发默认
		}

		v.AddConfigPath("./backend/config")

		v.SetEnvPrefix("APP")
		v.AutomaticEnv()
		//默认值（防止空值崩溃）
		v.SetDefault("app.name", "wsai")
		v.SetDefault("app.env", "dev")
		v.SetDefault("app.port", "9091")
		v.SetDefault("jwt.access_ttl", "2h")
		v.SetDefault("jwt.refresh_ttl", "30d")

		if err := v.ReadInConfig(); err != nil {
			log.Printf("警告: 未找到配置文件，使用默认值+环境变量: %v", err)
		} else {
			log.Printf("配置文件加载成功: %s", v.ConfigFileUsed())
		}

		if err := v.Unmarshal(&C); err != nil {
			log.Fatalf("解析配置失败: %v", err)
		}
		log.Printf("配置初始化完成 | 环境: %s | 应用: %s", C.App.Env, C.App.Name)
	})

}
