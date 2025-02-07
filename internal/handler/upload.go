package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"tfile/internal/utils"
)


func HandleUpload(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 获取上传路径参数
		uploadPath := r.FormValue("path")
		targetDir := filepath.Join(baseDir, uploadPath)

		// 路径验证
		if !utils.IsSafePath(targetDir, baseDir) {
			http.Error(w, "Invalid upload path", http.StatusBadRequest)
			return
		}

		// 创建目录（如果不存在）
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			http.Error(w, "Cannot create directory", http.StatusInternalServerError)
			return
		}

		// 解析表单
		if err := r.ParseMultipartForm(100 << 20); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["files"]
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Error retrieving file", http.StatusBadRequest)
				return
			}
			defer file.Close()

			// 构建安全路径
			targetPath := filepath.Join(targetDir, filepath.Base(fileHeader.Filename))
			if !utils.IsSafePath(targetPath, baseDir) {
				http.Error(w, "Invalid file path", http.StatusBadRequest)
				return
			}

			// 创建文件
			dst, err := os.Create(targetPath)
			if err != nil {
				http.Error(w, "Error creating file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Error saving file", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
	}
}