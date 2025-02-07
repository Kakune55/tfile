package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"tfile/internal/model"
	"tfile/internal/utils"
)

func HandleList(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取请求路径参数
		reqPath := r.URL.Query().Get("path")
		targetDir := filepath.Join(baseDir, reqPath)

		// 安全检查
		if !utils.IsSafePath(targetDir, baseDir) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		entries, err := os.ReadDir(targetDir)
		if err != nil {
			http.Error(w, "Unable to read directory", http.StatusInternalServerError)
			return
		}

		files := make([]model.FileInfo, 0, len(entries))
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			files = append(files, model.FileInfo{
				Name:    entry.Name(),
				IsDir:   entry.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime().Format(time.RFC3339),
				Path:    filepath.Join(reqPath, entry.Name()), // 记录相对路径
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
	}
}