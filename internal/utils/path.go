package utils

import (
	"path/filepath"
	"strings"
)


func IsSafePath(target, baseDir string) bool {
	// 获取相对路径
	rel, err := filepath.Rel(baseDir, target)
	if err != nil {
		return false
	}

	// 防止路径穿越
	if strings.Contains(rel, "..") {
		return false
	}

	// 标准化路径比较
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}

	return strings.HasPrefix(absTarget, absBase)
}