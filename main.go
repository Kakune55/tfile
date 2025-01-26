package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type FileInfo struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
	Path    string `json:"path"`
}

func main() {
	var dir string
	flag.StringVar(&dir, "path", ".", "Directory to share")
	flag.Parse()

	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/static/", http.FileServer(http.FS(staticFS)))
	http.HandleFunc("/", handleHome(absDir))
	http.HandleFunc("/api/list", handleList(absDir))
	http.HandleFunc("/api/upload", handleUpload(absDir))
	http.HandleFunc("/api/rename", handleRename(absDir))
	http.HandleFunc("/api/delete", handleDelete(absDir))
	http.HandleFunc("/api/mkdir", handleMkdir(absDir))
	http.HandleFunc("/api/download/", handleDownload(absDir))

	log.Printf("Serving directory %s on :8080", absDir)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHome(baseDir string) http.HandlerFunc {
	tmpl := template.Must(template.ParseFS(templateFS, "templates/index.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, struct{ BaseDir string }{baseDir}); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
		}
	}
}

func handleList(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取请求路径参数
		reqPath := r.URL.Query().Get("path")
		targetDir := filepath.Join(baseDir, reqPath)
		
		// 安全检查
		if !isSafePath(targetDir, baseDir) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		
		entries, err := os.ReadDir(targetDir)
		if err != nil {
			http.Error(w, "Unable to read directory", http.StatusInternalServerError)
			return
		}

		files := make([]FileInfo, 0, len(entries))
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			files = append(files, FileInfo{
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

// 新增创建目录处理函数
func handleMkdir(baseDir string) http.HandlerFunc {
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
		if !isSafePath(targetPath, baseDir) {
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

func handleUpload(baseDir string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        // 获取上传路径参数
        uploadPath := r.FormValue("path")
        targetDir := filepath.Join(baseDir, uploadPath)

        // 路径验证
        if !isSafePath(targetDir, baseDir) {
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
            if !isSafePath(targetPath, baseDir) {
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

func handleDownload(baseDir string) http.HandlerFunc {
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
        if !isSafePath(targetPath, baseDir) {
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
        w.Header().Set("Content-Disposition", 
            fmt.Sprintf("attachment; filename=\"%s\"", html.EscapeString(fileName)))
        http.ServeFile(w, r, targetPath)
    }
}

func handleRename(baseDir string) http.HandlerFunc {
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
        if !isSafePath(oldPath, baseDir) || !isSafePath(newPath, baseDir) {
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

func handleDelete(baseDir string) http.HandlerFunc {
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
        if !isSafePath(targetPath, baseDir) {
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

func isSafePath(target, baseDir string) bool {
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