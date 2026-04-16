package prompt

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"

	"simple-agent/utils"
)

//go:embed prompt.json
var defaultPromptJSON []byte

var (
	corePath   = utils.MemoryDir + "/" + "core.md"
	memoryPath = utils.MemoryDir + "/" + "memory.md"
)

type system struct {
	core   prompt // 核心身份和行为说明
	memory prompt
}

type SystemPrompt interface {
	GetPrompt() string
}

func (s *system) buildSystemPrompt() prompt {
	return prompt(s.core.structPrompt("核心身份和行文说明") +
		s.memory.structPrompt("用户记忆"))
}

func (s *system) GetPrompt() string {
	return string(s.buildSystemPrompt())
}

type promptJSON struct {
	Identity string `json:"identity"`
	Memory   string `json:"memory"`
}

// LoadStaticSystemPrompt 加载 core.md 和 memory.md 到 system。
// 若文件不存在，从 prompt.json 读取默认值并创建文件。
func New() (SystemPrompt, error) {
	core, err := loadOrCreate(corePath, func(j promptJSON) string { return j.Identity })
	if err != nil {
		return &system{}, err
	}

	memory, err := loadOrCreate(memoryPath, func(j promptJSON) string { return j.Memory })
	if err != nil {
		return &system{}, err
	}

	return &system{core: prompt(core), memory: prompt(memory)}, nil
}

// loadOrCreate 读取 path 文件内容；若文件不存在则从 prompt.json 取默认值并写入文件。
func loadOrCreate(path string, pick func(promptJSON) string) (string, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return string(data), nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	// 文件不存在，从 prompt.json 读取默认值
	content, err := defaultFromJSON(pick)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return content, nil
}

// defaultFromJSON 从编译时嵌入的 prompt.json 中取出所需字段。
func defaultFromJSON(pick func(promptJSON) string) (string, error) {
	var j promptJSON
	if err := json.Unmarshal(defaultPromptJSON, &j); err != nil {
		return "", err
	}
	return pick(j), nil
}
