package handler

import (
	"html/template"
	"net/http"

	"tfile/web"
)

func HandleHome(baseDir string) http.HandlerFunc {
	tmpl := template.Must(template.ParseFS(web.TemplateFS, "templates/index.html"))

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