package types

import "fmt"

func BuildReminder(content string) Message {
	return Message{
		Type:    TYPE_TEXT,
		Content: fmt.Sprintf("<reminder>%s</reminder>", content),
	}
}
