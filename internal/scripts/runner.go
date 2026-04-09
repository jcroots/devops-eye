// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>
// See LICENSE file.

package scripts

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jcroots/devops-eye/internal/render"
)

const defaultTimeout = 30 * time.Second

type Script struct {
	Name string
	Path string
}

type Module struct {
	dir     string
	scripts []Script
}

func New(dir string) (*Module, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("scripts: resolve dir: %w", err)
	}
	absDir, err = filepath.EvalSymlinks(absDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("scripts: resolve dir symlinks: %w", err)
	}
	m := &Module{dir: absDir}
	m.scripts, err = discover(absDir)
	if err != nil {
		return nil, fmt.Errorf("scripts: discover: %w", err)
	}
	return m, nil
}

func (m *Module) Name() string { return "Scripts" }
func (m *Module) Path() string { return "/scripts/" }
func (m *Module) Handler() http.Handler {
	return http.HandlerFunc(m.handle)
}

func (m *Module) handle(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/scripts/")
	if name == "" {
		m.list(w)
		return
	}
	m.run(w, name)
}

func (m *Module) list(w http.ResponseWriter) {
	var b strings.Builder
	if len(m.scripts) == 0 {
		b.WriteString("<p>No scripts found in <code>" + template.HTMLEscapeString(m.dir) + "</code></p>\n")
	} else {
		b.WriteString("<table>\n<tr><th>Script</th></tr>\n")
		for _, s := range m.scripts {
			b.WriteString(fmt.Sprintf(`<tr><td><a href="/scripts/%s">%s</a></td></tr>`+"\n",
				template.HTMLEscapeString(s.Name),
				template.HTMLEscapeString(s.Name)))
		}
		b.WriteString("</table>\n")
	}
	render.Page(w, "Scripts", template.HTML(b.String()))
}

func (m *Module) run(w http.ResponseWriter, name string) {
	// Validate: name must match a discovered script.
	var script *Script
	for i := range m.scripts {
		if m.scripts[i].Name == name {
			script = &m.scripts[i]
			break
		}
	}
	if script == nil {
		http.NotFound(w, nil)
		return
	}

	// Double-check the path is still inside the scripts directory.
	if !isInsideDir(script.Path, m.dir) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Execute with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, script.Path)
	cmd.Dir = m.dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<h2>%s</h2>\n", template.HTMLEscapeString(script.Name)))

	if err != nil {
		b.WriteString(fmt.Sprintf("<p><strong>Error:</strong> %s</p>\n", template.HTMLEscapeString(err.Error())))
	}
	if stdout.Len() > 0 {
		b.WriteString("<h3>Output</h3>\n<pre>")
		b.WriteString(template.HTMLEscapeString(stdout.String()))
		b.WriteString("</pre>\n")
	}
	if stderr.Len() > 0 {
		b.WriteString("<h3>Stderr</h3>\n<pre>")
		b.WriteString(template.HTMLEscapeString(stderr.String()))
		b.WriteString("</pre>\n")
	}
	if stdout.Len() == 0 && stderr.Len() == 0 && err == nil {
		b.WriteString("<p>(no output)</p>\n")
	}

	render.Page(w, "Script: "+script.Name, template.HTML(b.String()))
}

func discover(dir string) ([]Script, error) {
	// Resolve symlinks on dir so isInsideDir checks are consistent.
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	entries, err := os.ReadDir(realDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var scripts []Script
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Use Lstat to avoid following symlinks.
		fullPath := filepath.Join(realDir, entry.Name())
		info, err := os.Lstat(fullPath)
		if err != nil {
			continue
		}
		// Skip symlinks.
		if info.Mode()&fs.ModeSymlink != 0 {
			continue
		}
		// Must be a regular file with executable bit.
		if !info.Mode().IsRegular() {
			continue
		}
		if info.Mode().Perm()&0111 == 0 {
			continue
		}
		// Ensure the resolved path is inside the scripts directory.
		resolved, err := filepath.EvalSymlinks(fullPath)
		if err != nil {
			continue
		}
		if !isInsideDir(resolved, realDir) {
			continue
		}
		scripts = append(scripts, Script{
			Name: entry.Name(),
			Path: resolved,
		})
	}
	return scripts, nil
}

func isInsideDir(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}
