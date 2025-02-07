package handler

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"tfile/internal/utils"
)

func HandleDownload(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodedPath := strings.TrimPrefix(r.URL.Path, "/api/download/")
		if encodedPath == "" {
			http.Error(w, "Filename required", http.StatusBadRequest)
			return
		}

		// 解码路径
		decodedPath, err := url.PathUnescape(encodedPath)
		if err != nil {
			http.Error(w, "Invalid filename", http.StatusBadRequest)
			return
		}

		// 构建完整路径
		targetPath := filepath.Join(baseDir, decodedPath)
		if !utils.IsSafePath(targetPath, baseDir) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// 检查文件存在
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// 设置下载头
        fileName := filepath.Base(decodedPath)
        escapedFilename := url.QueryEscape(fileName) // 用于 filename*
        w.Header().Set("Content-Disposition",
            fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
                html.EscapeString(fileName), // 保留原有兼容性
                escapedFilename))
        http.ServeFile(w, r, targetPath)
	}
}