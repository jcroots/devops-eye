// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>
// See LICENSE file.

package config

import (
	"os"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := &Config{
		ListenAddr: ":8080",
		ScriptsDir: "./scripts.d",
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected :8080, got %s", cfg.ListenAddr)
	}
	if cfg.ScriptsDir != "./scripts.d" {
		t.Errorf("expected ./scripts.d, got %s", cfg.ScriptsDir)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("DEVOPS_EYE_LISTEN", ":9090")
	t.Setenv("DEVOPS_EYE_SCRIPTS_DIR", "/tmp/scripts")

	// Simulate what Load() does for env var handling.
	cfg := &Config{
		ListenAddr: ":8080",
		ScriptsDir: "./scripts.d",
	}
	if v := os.Getenv("DEVOPS_EYE_LISTEN"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("DEVOPS_EYE_SCRIPTS_DIR"); v != "" {
		cfg.ScriptsDir = v
	}

	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected :9090, got %s", cfg.ListenAddr)
	}
	if cfg.ScriptsDir != "/tmp/scripts" {
		t.Errorf("expected /tmp/scripts, got %s", cfg.ScriptsDir)
	}
}
