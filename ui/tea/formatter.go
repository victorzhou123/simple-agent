package tea

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// 工具输出样式
var (
	toolHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#3B82F6")).
			Padding(0, 1).
			Width(50)

	toolContentStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#6B7280")).
				Padding(0, 1).
				Width(50)

	toolSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				Bold(true)

	toolFailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)
)

type Formatter interface {
	FormatToolCall(toolName string, input map[string]any) string
	FormatToolResult(toolName string, output string, err error) string
}

type formatter struct {
	toolIcons map[string]string
}

func newFormatter(cfg FormatterConfig) Formatter {
	return &formatter{
		toolIcons: cfg.Icons,
	}
}

// FormatToolCall 格式化工具调用标题
func (f *formatter) FormatToolCall(toolName string, input map[string]any) string {
	icon, ok := f.toolIcons[toolName]
	if !ok {
		icon = "⚙️"
	}

	header := fmt.Sprintf("%s %s", icon, toolName)

	// 如果有重要参数，显示在标题中
	var params []string
	if cmd, ok := input["command"].(string); ok && cmd != "" {
		if len(cmd) > 30 {
			cmd = cmd[:30] + "..."
		}
		params = append(params, cmd)
	}
	if path, ok := input["file_path"].(string); ok && path != "" {
		if len(path) > 30 {
			path = "..." + path[len(path)-27:]
		}
		params = append(params, path)
	}

	if len(params) > 0 {
		header += " • " + strings.Join(params, " ")
	}

	return "\n\n" + toolHeaderStyle.Render(header)
}

// FormatToolResult 格式化工具执行结果
func (f *formatter) FormatToolResult(toolName string, output string, err error) string {
	var result strings.Builder

	// 输出内容
	if output != "" {
		result.WriteString("\n")

		// 特殊处理 todo 工具，使用表格展示
		if toolName == "todo" {
			formatted := f.formatTodoOutput(output)
			result.WriteString(formatted)
		} else {
			// 限制输出长度，避免界面过长
			lines := strings.Split(output, "\n")
			if len(lines) > 20 {
				output = strings.Join(lines[:20], "\n") + "\n... (输出已截断)"
			}
			result.WriteString(toolContentStyle.Render(output))
		}
	}

	// 状态标识
	result.WriteString("\n")
	if err != nil {
		result.WriteString(toolFailStyle.Render("✗ 执行失败: " + err.Error()))
	} else {
		result.WriteString(toolSuccessStyle.Render("✓ 执行成功"))
	}
	result.WriteString("\n")

	return result.String()
}

// formatTodoOutput 解析 todo 工具的输出并格式化为表格
func (f *formatter) formatTodoOutput(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return "无任务"
	}

	headers := []string{"#", "任务", "状态"}
	rows := make([][]string, 0, len(lines))

	for i, line := range lines {
		// 解析格式: [x] 任务内容 (id)
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var statusIcon string
		var text string

		if strings.HasPrefix(line, "[x]") {
			statusIcon = "✓ 完成"
			text = strings.TrimSpace(line[3:])
		} else if strings.HasPrefix(line, "[-]") {
			statusIcon = "⏳ 进行中"
			text = strings.TrimSpace(line[3:])
		} else if strings.HasPrefix(line, "[ ]") {
			statusIcon = "⭕ 待办"
			text = strings.TrimSpace(line[3:])
		} else {
			continue
		}

		// 移除 ID 部分 (xxx)
		if idx := strings.LastIndex(text, "("); idx > 0 {
			text = strings.TrimSpace(text[:idx])
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			text,
			statusIcon,
		})
	}

	if len(rows) == 0 {
		return "无任务"
	}

	return f.formatTable(headers, rows)
}

// FormatTable 创建表格样式（用于 todo 等结构化数据）
func (f *formatter) formatTable(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	purple := lipgloss.Color("#7C3AED")
	gray := lipgloss.Color("#9CA3AF")

	headerStyle := lipgloss.NewStyle().
		Foreground(purple).
		Bold(true).
		Align(lipgloss.Center)

	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	evenRow := cellStyle.Foreground(lipgloss.Color("#E5E7EB"))
	oddRow := cellStyle.Foreground(gray)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		Headers(headers...).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return headerStyle
			case row%2 == 0:
				return evenRow
			default:
				return oddRow
			}
		})

	for _, row := range rows {
		t.Row(row...)
	}

	return t.String()
}
