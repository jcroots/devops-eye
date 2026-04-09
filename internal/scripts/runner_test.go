// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>
// See LICENSE file.

package scripts

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverEmpty(t *testing.T) {
	dir := t.TempDir()
	scripts, err := discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) != 0 {
		t.Errorf("expected 0 scripts, got %d", len(scripts))
	}
}

func TestDiscoverNonexistent(t *testing.T) {
	scripts, err := discover("/nonexistent/path")
	if err != nil {
		t.Fatal(err)
	}
	if scripts != nil {
		t.Errorf("expected nil, got %v", scripts)
	}
}

func TestDiscoverExecutable(t *testing.T) {
	dir := t.TempDir()

	// Create an executable script.
	scriptPath := filepath.Join(dir, "test.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/sh\necho ok\n"), 0755)

	// Create a non-executable file.
	os.WriteFile(filepath.Join(dir, "notexec.txt"), []byte("data"), 0644)

	// Create a directory (should be skipped).
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)

	scripts, err := discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) != 1 {
		t.Fatalf("expected 1 script, got %d", len(scripts))
	}
	if scripts[0].Name != "test.sh" {
		t.Errorf("expected test.sh, got %s", scripts[0].Name)
	}
}

func TestDiscoverSkipsSymlinks(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()

	// Create a script outside the dir.
	outsideScript := filepath.Join(outside, "evil.sh")
	os.WriteFile(outsideScript, []byte("#!/bin/sh\necho evil\n"), 0755)

	// Symlink it into the scripts dir.
	os.Symlink(outsideScript, filepath.Join(dir, "evil.sh"))

	scripts, err := discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range scripts {
		if s.Name == "evil.sh" {
			t.Error("symlink to outside directory should be rejected")
		}
	}
}

func TestIsInsideDir(t *testing.T) {
	tests := []struct {
		path, dir string
		want      bool
	}{
		{"/a/b/c", "/a/b", true},
		{"/a/b", "/a/b", true},
		{"/a/x", "/a/b", false},
		{"/other/path", "/a/b", false},
	}
	for _, tt := range tests {
		got := isInsideDir(tt.path, tt.dir)
		if got != tt.want {
			t.Errorf("isInsideDir(%q, %q) = %v, want %v", tt.path, tt.dir, got, tt.want)
		}
	}
}

func TestModuleList(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "hello.sh"), []byte("#!/bin/sh\necho hello\n"), 0755)

	m, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/scripts/", nil)
	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "hello.sh") {
		t.Error("expected hello.sh in listing")
	}
}

func TestModuleRun(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "greet.sh"), []byte("#!/bin/sh\necho hello world\n"), 0755)

	m, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/scripts/greet.sh", nil)
	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "hello world") {
		t.Error("expected script output in response")
	}
}

func TestModuleRunNotFound(t *testing.T) {
	dir := t.TempDir()
	m, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/scripts/nope.sh", nil)
	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
