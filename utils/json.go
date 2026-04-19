package utils

import (
	"encoding/json"
	"os"
)

// LoadJSONFile 从本地文件加载 JSON 数据到指定的结构体
func LoadJSONFile(filepath string, v interface{}) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
