package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"tfile/internal/utils"
)

func HandleDelete(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var data struct {
			Path string `json:"path"` // 完整相对路径
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// 构建完整路径
		targetPath := filepath.Join(baseDir, data.Path)

		// 安全验证
		if !utils.IsSafePath(targetPath, baseDir) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// 检查存在性
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// 执行删除
		if err := os.RemoveAll(targetPath); err != nil {
			log.Printf("Delete error: %v", err)
			http.Error(w, "Delete failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}