// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>
// See LICENSE file.

package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/jcroots/devops-eye/config"
	"github.com/jcroots/devops-eye/internal/render"
)

type Module interface {
	Name() string
	Path() string
	Handler() http.Handler
}

type entry struct {
	Name string
	Path string
}

type Server struct {
	cfg     *config.Config
	mux     *http.ServeMux
	modules []entry
}

func New(cfg *config.Config) *Server {
	s := &Server{
		cfg: cfg,
		mux: http.NewServeMux(),
	}
	s.mux.HandleFunc("/", s.index)
	return s
}

func (s *Server) Register(m Module) {
	s.modules = append(s.modules, entry{Name: m.Name(), Path: m.Path()})
	s.mux.Handle(m.Path(), m.Handler())
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.cfg.ListenAddr, s.mux)
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var b strings.Builder
	b.WriteString("<h2>Modules</h2>\n<ul>\n")
	for _, m := range s.modules {
		b.WriteString(fmt.Sprintf(`<li><a href="%s">%s</a></li>`+"\n", m.Path, m.Name))
	}
	b.WriteString("</ul>\n")

	render.Page(w, "Index", template.HTML(b.String()))
}
