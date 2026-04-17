package bash

import (
	"fmt"
	"os/exec"

	"simple-agent/tools/base"
)

func New(cfg base.ToolConfig) base.Tool {
	return &bashTool{base.NewBaseTool(cfg.Name, cfg.Description, cfg.InputSchema)}
}

type bashTool struct{ base.BaseTool }

func (t *bashTool) Call(args map[string]any) (string, error) {
	cmd, ok := args["command"].(string)
	if !ok || cmd == "" {
		return "", fmt.Errorf("bash: command must be a non-empty string")
	}
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	return string(out), err
}
