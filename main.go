package main

import (
	"fmt"
	"os"

	"simple-agent/agent"
	"simple-agent/config"
	"simple-agent/model"
	"simple-agent/prompt"
	"simple-agent/tea"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误："+err.Error())
		os.Exit(1)
	}

	opts := []option.RequestOption{option.WithAPIKey(cfg.ApiKey)}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}
	client := anthropic.NewClient(opts...)

	endpoint := "api.anthropic.com"
	if cfg.BaseURL != "" {
		endpoint = cfg.BaseURL
	}

	// 用闭包打破 agent ↔ tea 的初始化循环依赖：
	// tea.New 需要 onSubmit 回调，agent.New 需要 ui 接口，
	// 两者通过闭包捕获指针在 main 层完成绑定。
	var ag *agent.Agent

	sysPrompt, err := prompt.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "system prompt 读取出错: %v\n", err)
		os.Exit(1)
	}

	prog, ui := tea.New(
		tea.Config{
			ModelName: cfg.Model,
			Endpoint:  endpoint,
		},
		func(question string) {
			ag.OnSubmit(question)
		},
	)

	modelCli := model.New(&client, cfg.Model, sysPrompt)

	ag = agent.New(modelCli, ui)

	if err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "程序出错: %v\n", err)
		os.Exit(1)
	}
}
