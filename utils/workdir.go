package utils

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
)

var (
	HomeDir           = homeDir()
	WorkDir           = workdir()
	ConfigDir         = HomeDir + "/" + ".simple-agent"
	ProjectDir        = ConfigDir + "/" + "projects"
	CurrentProjectDir = ProjectDir + "/" + getProjectName() + "-" + getProjectSlug()
	MemoryDir         = CurrentProjectDir + "/" + "memory"
	SkillDir          = CurrentProjectDir + "/" + "skills"
)

// homeDir 返回当前用户的 home 目录。
func homeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic("unable to get home directory: " + err.Error())
	}
	return dir
}

// workdir 返回程序启动时的工作目录。
func workdir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic("unable to get working directory: " + err.Error())
	}
	return dir
}

// getProjectName 返回当前工作目录的目录名。
func getProjectName() string {
	return filepath.Base(WorkDir)
}

// getProjectSlug 用 MD5 对工作目录的绝对路径编码，返回唯一标识。
func getProjectSlug() string {
	sum := md5.Sum([]byte(WorkDir))
	return fmt.Sprintf("%x", sum)
}
