package tea

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UI 是 tea 包对外暴露的接口，Agent 调用它来驱动显示。
type UI interface {
	AppendChunk(text string)                                      // 推送一个流式文本块
	Done(finalContent string)                                     // 流式结束，传入完整内容
	Fail(err error)                                               // 出错
	ShowToolCall(toolName string, input map[string]any)           // 显示工具调用
	ShowToolResult(toolName string, output string, err error)     // 显示工具结果
}

// SubmitFunc 是用户提交问题时的回调，由外部（Agent）注册。
type SubmitFunc func(question string)

// Config 持有标题栏需要的展示信息。
type Config struct {
	ModelName string
	Endpoint  string
}

// Program 包装 bubbletea.Program，对外只暴露 Run。
type Program struct {
	p *bubbletea.Program
}

func (p *Program) Run() error {
	_, err := p.p.Run()
	return err
}

// New 创建 tea Program 和 UI 接口。
// onSubmit 在用户按下 Ctrl+J 时被调用，应当非阻塞（内部自行启动 goroutine）。
func New(cfg Config, onSubmit SubmitFunc) (*Program, UI) {
	ta := textarea.New()
	ta.Placeholder = "输入问题，Enter 发送，Ctrl+C 退出..."
	ta.Focus()
	ta.CharLimit = 2000
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	sp := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))),
	)

	vp := viewport.New(80, 20)

	m := teaModel{
		cfg:      cfg,
		onSubmit: onSubmit,
		textarea: ta,
		viewport: vp,
		spinner:  sp,
	}

	ui := &uiImpl{}
	p := bubbletea.NewProgram(m, bubbletea.WithAltScreen(), bubbletea.WithMouseCellMotion())
	ui.send = p.Send

	return &Program{p: p}, ui
}

// ── 内部 UI 实现 ───────────────────────────────────────────────────────────────

type uiImpl struct {
	send func(bubbletea.Msg)
}

func (u *uiImpl) AppendChunk(text string)  { u.send(chunkMsg(text)) }
func (u *uiImpl) Done(final string)        { u.send(doneMsg{content: final}) }
func (u *uiImpl) Fail(err error)           { u.send(failMsg{err: err}) }

func (u *uiImpl) ShowToolCall(toolName string, input map[string]any) {
	u.send(chunkMsg(FormatToolCall(toolName, input)))
}

func (u *uiImpl) ShowToolResult(toolName string, output string, err error) {
	u.send(chunkMsg(FormatToolResult(toolName, output, err)))
}

// ── 内部消息类型 ───────────────────────────────────────────────────────────────

type chunkMsg string
type doneMsg struct{ content string }
type failMsg struct{ err error }

// ── 状态 ──────────────────────────────────────────────────────────────────────

type state int

const (
	stateIdle      state = iota
	stateStreaming
)

// ── 显示消息（已完成的对话条目）────────────────────────────────────────────────

type displayMsg struct {
	role    string // "user" | "assistant"
	content string
}

// ── Tea Model ─────────────────────────────────────────────────────────────────

type teaModel struct {
	cfg      Config
	onSubmit SubmitFunc

	// 已完成的对话历史（仅用于渲染）
	displayMessages []displayMsg
	// 当前正在流式输出的内容（流结束后清空）
	currentResponse string

	// UI 组件
	textarea textarea.Model
	viewport viewport.Model
	spinner  spinner.Model

	// 状态
	state      state
	windowSize bubbletea.WindowSizeMsg
	errMsg     string
}

func (m teaModel) Init() bubbletea.Cmd {
	return textarea.Blink
}

func (m teaModel) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	var cmds []bubbletea.Cmd

	switch msg := msg.(type) {

	case bubbletea.WindowSizeMsg:
		m.windowSize = msg
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10
		m.textarea.SetWidth(msg.Width - 4)
		m.refreshViewport()

	case bubbletea.KeyMsg:
		switch msg.Type {
		case bubbletea.KeyCtrlC:
			return m, bubbletea.Quit
		case bubbletea.KeyEnter:
			if m.state == stateIdle {
				return m.handleSubmit()
			}
		default:
			if m.state == stateIdle {
				var cmd bubbletea.Cmd
				m.textarea, cmd = m.textarea.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	// Agent 推送流式文本块
	case chunkMsg:
		m.currentResponse += string(msg)
		m.errMsg = ""
		m.refreshViewport()

	// Agent 通知流式结束
	case doneMsg:
		if msg.content != "" {
			m.displayMessages = append(m.displayMessages, displayMsg{
				role:    "assistant",
				content: msg.content,
			})
		}
		m.currentResponse = ""
		m.state = stateIdle
		m.refreshViewport()
		m.textarea.Focus()

	// Agent 通知出错
	case failMsg:
		m.errMsg = msg.err.Error()
		m.currentResponse = ""
		m.state = stateIdle
		m.refreshViewport()
		m.textarea.Focus()

	case spinner.TickMsg:
		if m.state == stateStreaming {
			var cmd bubbletea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	var vpCmd bubbletea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, bubbletea.Batch(cmds...)
}

func (m teaModel) handleSubmit() (teaModel, bubbletea.Cmd) {
	question := strings.TrimSpace(m.textarea.Value())
	if question == "" {
		return m, nil
	}

	m.displayMessages = append(m.displayMessages, displayMsg{role: "user", content: question})
	m.currentResponse = ""
	m.errMsg = ""
	m.state = stateStreaming

	m.textarea.Reset()
	m.textarea.Blur()
	m.refreshViewport()

	// 在 bubbletea 管理的 goroutine 里调用 onSubmit，避免阻塞事件循环
	submitCmd := func() bubbletea.Msg {
		m.onSubmit(question)
		return nil
	}

	return m, bubbletea.Batch(submitCmd, m.spinner.Tick)
}

func (m *teaModel) refreshViewport() {
	var sb strings.Builder

	for _, msg := range m.displayMessages {
		if msg.role == "user" {
			sb.WriteString(userStyle.Render("你") + "\n")
		} else {
			sb.WriteString(assistantStyle.Render("Claude") + "\n")
		}
		sb.WriteString(msg.content + "\n\n")
	}

	if m.currentResponse != "" {
		sb.WriteString(assistantStyle.Render("Claude") + "\n")
		sb.WriteString(m.currentResponse)
		if m.state == stateStreaming {
			sb.WriteString(m.spinner.View())
		}
		sb.WriteString("\n\n")
	}

	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

func (m teaModel) View() string {
	var sb strings.Builder

	title := fmt.Sprintf(" %s @ %s ", m.cfg.ModelName, m.cfg.Endpoint)
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n\n")

	sb.WriteString(borderStyle.Render(m.viewport.View()))
	sb.WriteString("\n")

	if m.state == stateStreaming {
		sb.WriteString(dimStyle.Render("AI 正在输出中 " + m.spinner.View()))
	} else if m.errMsg != "" {
		sb.WriteString(errorStyle.Render("错误: " + m.errMsg))
	} else {
		sb.WriteString(dimStyle.Render("Enter 发送  Ctrl+C 退出  ↑/↓ 滚动历史"))
	}
	sb.WriteString("\n\n")

	sb.WriteString(borderStyle.Render(m.textarea.View()))

	return sb.String()
}

// ── 样式 ──────────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 2)

	userStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10B981"))

	assistantStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED"))
)
