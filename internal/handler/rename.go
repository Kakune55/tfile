package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"tfile/internal/utils"
)

func HandleRename(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var data struct {
			Old string `json:"old"`
			New string `json:"new"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// 构建完整路径
		oldPath := filepath.Join(baseDir, data.Old)
		newPath := filepath.Join(baseDir, data.New)

		// 双重验证
		if !utils.IsSafePath(oldPath, baseDir) || !utils.IsSafePath(newPath, baseDir) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// 检查源文件存在
		if _, err := os.Stat(oldPath); os.IsNotExist(err) {
			http.Error(w, "Source not found", http.StatusNotFound)
			return
		}

		// 执行重命名
		if err := os.Rename(oldPath, newPath); err != nil {
			log.Printf("Rename error: %v", err)
			http.Error(w, "Rename failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
