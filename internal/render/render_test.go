// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>
// See LICENSE file.

package render

import (
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPage(t *testing.T) {
	w := httptest.NewRecorder()
	Page(w, "Test Page", template.HTML("<p>hello</p>"))

	body := w.Body.String()

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html, got %s", ct)
	}
	if !strings.Contains(body, "<title>Test Page - devops-eye</title>") {
		t.Error("title not found in output")
	}
	if !strings.Contains(body, "<p>hello</p>") {
		t.Error("content not found in output")
	}
	if !strings.Contains(body, `<a href="/">devops-eye</a>`) {
		t.Error("nav link not found in output")
	}
}
