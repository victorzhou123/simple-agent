package config

import (
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

type Config struct {
	ApiKey  string
	BaseURL string // 留空则使用官方地址
	Model   string // 模型名，支持自定义
}

// loadConfig 从项目根目录的 .env 文件读取配置，不存在时回退到环境变量
//
//	ANTHROPIC_API_KEY   必填
//	ANTHROPIC_BASE_URL  可选，第三方站点地址，例如 https://api.example.com
//	ANTHROPIC_MODEL     可选，默认 claude-opus-4-6
func LoadConfig() (Config, error) {
	// 加载 .env（文件不存在时忽略错误，不影响已有环境变量）
	_ = godotenv.Load()

	cfg := Config{
		ApiKey:  os.Getenv("ANTHROPIC_API_KEY"),
		BaseURL: os.Getenv("ANTHROPIC_BASE_URL"),
		Model:   os.Getenv("ANTHROPIC_MODEL"),
	}
	if cfg.ApiKey == "" {
		return cfg, fmt.Errorf("未找到 ANTHROPIC_API_KEY，请在 .env 文件中配置")
	}
	if cfg.Model == "" {
		cfg.Model = string(anthropic.ModelClaudeOpus4_6)
	}
	return cfg, nil
}
