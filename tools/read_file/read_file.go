package readfile

import (
	"context"
	"fmt"
	"os"
	"strings"

	"simple-agent/tools/base"
)

func New(cfg base.ToolConfig) base.Tool {
	return &readFileTool{base.NewBaseTool(cfg.Name, cfg.Description, cfg.InputSchema)}
}

type readFileTool struct{ base.BaseTool }

func (t *readFileTool) Call(ctx context.Context, args map[string]any) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("read_file: path must be a non-empty string")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)

	var truncated bool
	if raw, ok := args["limit"]; ok {
		if limit := toInt(raw); limit > 0 {
			lines := strings.SplitN(content, "\n", limit+1)
			if len(lines) > limit {
				lines = lines[:limit]
				truncated = true
			}
			content = strings.Join(lines, "\n")
		}
	}

	// 格式化输出
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📄 %s\n\n", path))

	// 添加行号
	lines := strings.Split(content, "\n")
	maxLineNum := len(lines)
	numWidth := len(fmt.Sprintf("%d", maxLineNum))

	for i, line := range lines {
		sb.WriteString(fmt.Sprintf("%*d │ %s\n", numWidth, i+1, line))
	}

	if truncated {
		sb.WriteString("\n... (内容已截断)")
	}

	return sb.String(), nil
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
