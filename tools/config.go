package tools

import (
	"simple-agent/tools/base"
	"simple-agent/tools/subagent"
)

type Config struct {
	Tools          []base.ToolConfig `json:"tools"`
	SubagentConfig subagent.Config   `json:"subagent"`
}
