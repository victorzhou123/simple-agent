package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadJSONFile 从本地文件加载 JSON 数据到指定的结构体
func LoadJSONFile(filepath string, v interface{}) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func LoadFile(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// 获取某个路径下的所有文件路径(支持递归)
// dpt 递归深度限制：-1表示无限递归(最大5层)，0表示只获取当前目录，1-5表示指定深度
func GetAllFilePaths(dirPath string, dpt int) ([]string, error) {
	if dpt > 5 {
		return nil, fmt.Errorf("递归深度不能超过5层，当前值: %d", dpt)
	}

	maxDepth := dpt
	if dpt == -1 {
		maxDepth = 5 // 无限递归时最大深度为5层
	}

	var files []string
	cleanBase := filepath.Clean(dirPath)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算当前路径相对于基础路径的深度
		relPath, err := filepath.Rel(cleanBase, path)
		if err != nil {
			return err
		}

		// 计算深度：统计路径分隔符数量
		depth := 0
		if relPath != "." {
			depth = strings.Count(relPath, string(filepath.Separator))
		}

		if depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
