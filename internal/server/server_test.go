// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>
// See LICENSE file.

package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jcroots/devops-eye/config"
)

type testModule struct {
	name    string
	path    string
	handler http.Handler
}

func (m *testModule) Name() string        { return m.name }
func (m *testModule) Path() string        { return m.path }
func (m *testModule) Handler() http.Handler { return m.handler }

func TestIndexEmpty(t *testing.T) {
	srv := New(&config.Config{})
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Modules") {
		t.Error("expected Modules heading")
	}
}

func TestIndexWithModule(t *testing.T) {
	srv := New(&config.Config{})
	srv.Register(&testModule{
		name: "Test",
		path: "/test",
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		}),
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `<a href="/test">Test</a>`) {
		t.Error("expected module link in index")
	}
}

func TestNotFound(t *testing.T) {
	srv := New(&config.Config{})
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
