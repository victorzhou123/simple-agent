package claude

import (
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

type Config struct {
	ApiKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

func (cfg *Config) SetDefault() {
	if cfg.Model == "" {
		cfg.Model = anthropic.ModelClaudeOpus4_6
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com"
	}
}

func (cfg *Config) Validate() error {
	if cfg.ApiKey == "" {
		return fmt.Errorf("未找到 ANTHROPIC_API_KEY，请在 .env 文件中配置")
	}
	return nil
}

func (cfg *Config) LoadFromEnv() {
	if cfg.ApiKey == "" {
		cfg.ApiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = os.Getenv("ANTHROPIC_BASE_URL")
	}

	if cfg.Model == "" {
		cfg.Model = os.Getenv("ANTHROPIC_MODEL")
	}
}
