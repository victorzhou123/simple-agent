package editfile

import (
	"context"
	"fmt"
	"os"
	"strings"

	"simple-agent/tools/base"
)

func New(cfg base.ToolConfig) base.Tool {
	return &editFileTool{base.NewBaseTool(cfg.Name, cfg.Description, cfg.InputSchema)}
}

type editFileTool struct{ base.BaseTool }

func (t *editFileTool) Call(ctx context.Context, args map[string]any) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("edit_file: path must be a non-empty string")
	}
	oldText, ok := args["old_text"].(string)
	if !ok {
		return "", fmt.Errorf("edit_file: old_text must be a string")
	}
	newText, ok := args["new_text"].(string)
	if !ok {
		return "", fmt.Errorf("edit_file: new_text must be a string")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	if !strings.Contains(content, oldText) {
		return "", fmt.Errorf("edit_file: old_text not found in %s", path)
	}
	content = strings.Replace(content, oldText, newText, 1)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return "ok", nil
}
