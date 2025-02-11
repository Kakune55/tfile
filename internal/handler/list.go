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


func HandlerGetIndex(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		index, err := generateIndex(baseDir)
		if err != nil {
			http.Error(w, "Failed to generate index", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(index)
	}
}


// 扫描所有目录并生成索引
func generateIndex(directory string) ([]model.FileInfo, error) {
	var index []model.FileInfo
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		path = filepath.ToSlash(path)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			index = append(index, model.FileInfo{
				Name: info.Name(),
				IsDir: info.IsDir(),
				Size: info.Size(),
				ModTime: info.ModTime().Format(time.RFC3339),
				Path: path,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return index, nil
}
