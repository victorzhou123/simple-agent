package bash

import (
	"context"
	"fmt"
	"os/exec"

	"simple-agent/tools/base"
)

func New(cfg base.ToolConfig) base.Tool {
	return &bashTool{base.NewBaseTool(cfg.Name, cfg.Description, cfg.InputSchema)}
}

type bashTool struct{ base.BaseTool }

func (t *bashTool) Call(ctx context.Context, args map[string]any) (string, error) {
	cmd, ok := args["command"].(string)
	if !ok || cmd == "" {
		return "", fmt.Errorf("bash: command must be a non-empty string")
	}
	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	out, err := c.CombinedOutput()
	return string(out), err
}
