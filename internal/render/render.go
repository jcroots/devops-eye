// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>
// See LICENSE file.

package render

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*.html
var templateFS embed.FS

var baseTmpl = template.Must(template.ParseFS(templateFS, "templates/base.html"))

type pageData struct {
	Title   string
	Content template.HTML
}

func Page(w http.ResponseWriter, title string, content template.HTML) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := baseTmpl.Execute(w, pageData{
		Title:   title,
		Content: content,
	})
	if err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
	}
}
