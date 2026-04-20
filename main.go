package main

import (
	"fmt"
	"os"

	"simple-agent/agent"
	"simple-agent/config"
	"simple-agent/model/claude"
	"simple-agent/prompt"
	"simple-agent/tools"
	"simple-agent/ui/tea"
)

func main() {
	cfg := new(config.Config)
	if err := config.LoadConfig("./config/config.json", cfg); err != nil {
		fmt.Fprintf(os.Stderr, "配置文件加载出错: %v\n", err)
		os.Exit(1)
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

	ui := tea.New(
		cfg.UI.Tea,
		func(question string) {
			ag.OnSubmit(question)
		},
	)

	// tools init
	tools.Init(cfg.Tools.Tools)

	modelCli := claude.New(cfg.Model.Claude, sysPrompt)

	ag = agent.New(modelCli, ui, cfg.Agent)

	if err := ui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "程序出错: %v\n", err)
		os.Exit(1)
	}
}
