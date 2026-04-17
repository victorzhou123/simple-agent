package readfile

import (
	"fmt"
	"os"
	"strings"

	"simple-agent/tools/base"
)

func New(cfg base.ToolConfig) base.Tool {
	return &readFileTool{base.NewBaseTool(cfg.Name, cfg.Description, cfg.InputSchema)}
}

type readFileTool struct{ base.BaseTool }

func (t *readFileTool) Call(args map[string]any) (string, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("read_file: path must be a non-empty string")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	if raw, ok := args["limit"]; ok {
		if limit := toInt(raw); limit > 0 {
			lines := strings.SplitN(content, "\n", limit+1)
			if len(lines) > limit {
				lines = lines[:limit]
			}
			content = strings.Join(lines, "\n")
		}
	}
	return content, nil
}

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}
