package router

import (
	"net/http"

	"tfile/internal/handler"
	"tfile/web"
)

// NewRouter 创建并返回一个路由器
func NewRouter(baseDir string) http.Handler {
	mux := http.NewServeMux()

	// 路由注册
	mux.Handle("/static/", http.FileServer(http.FS(web.StaticFS)))
	mux.HandleFunc("/", handler.HandleHome(baseDir))
	mux.HandleFunc("/api/list", handler.HandleList(baseDir))
	mux.HandleFunc("/api/upload", handler.HandleUpload(baseDir))
	mux.HandleFunc("/api/rename", handler.HandleRename(baseDir))
	mux.HandleFunc("/api/delete", handler.HandleDelete(baseDir))
	mux.HandleFunc("/api/mkdir", handler.HandleMkdir(baseDir))
	mux.HandleFunc("/api/download/", handler.HandleDownload(baseDir))
	mux.HandleFunc("/api/getIndex", handler.HandlerGetIndex(baseDir))

	return mux
}
