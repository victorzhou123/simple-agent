package tools

import (
	"simple-agent/tools/bash"
	editfile "simple-agent/tools/edit_file"
	readfile "simple-agent/tools/read_file"
	"simple-agent/tools/todo"
	writefile "simple-agent/tools/write_file"
)

const (
	TOOL_NAME_BASH       = "bash"
	TOOL_NAME_READ_FILE  = "read_file"
	TOOL_NAME_WRITE_FILE = "write_file"
	TOOL_NAME_EDIT_FILE  = "edit_file"
	TOOL_NAME_TODO       = "todo"
)

func init() {
	// Auto-register all built-in tools
	Register(TOOL_NAME_BASH, bash.New)
	Register(TOOL_NAME_READ_FILE, readfile.New)
	Register(TOOL_NAME_WRITE_FILE, writefile.New)
	Register(TOOL_NAME_EDIT_FILE, editfile.New)
	Register(TOOL_NAME_TODO, todo.New)
}
