package prompt

import "fmt"

type prompt string

// 输入标签获取结构化的prompt
func (p prompt) structPrompt(t string) prompt {
	return prompt(fmt.Sprintf("# %s\n%s\n", t, string(p)))
}
