package types

// streamChunkMsg 是每个流式文本块触发的消息
type StreamChunkMsg string

// streamDoneMsg 表示流式输出结束
type StreamDoneMsg struct{}

// streamErrMsg 表示流式输出出错
type StreamErrMsg struct{ err error }

type Message struct {
	Role    string // "user" | "assistant"
	Content string
}
