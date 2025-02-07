package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"

	"tfile/internal/router"
	"tfile/internal/utils"
)



var ServerPort = "8080"

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

	// 初始化路由
	router := router.NewRouter(absDir)

	log.Printf("Serving directory %s on Port %s\n\n", absDir, ServerPort)
	utils.ShowIPinfo()
	log.Fatal(http.ListenAndServe(":"+ServerPort, router))
}