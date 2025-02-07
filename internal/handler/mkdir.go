package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"tfile/internal/utils"
)
func HandleMkdir(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var data struct {
			Name string `json:"name"`
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// 构建完整路径
		targetPath := filepath.Join(baseDir, data.Path, data.Name)
		if !utils.IsSafePath(targetPath, baseDir) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// 创建目录
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			http.Error(w, "Create directory failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}