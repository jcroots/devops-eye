// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>
// See LICENSE file.

package config

import (
	"flag"
	"os"
)

type Config struct {
	ListenAddr string
	ScriptsDir string
}

func Load() *Config {
	cfg := &Config{
		ListenAddr: "127.0.0.1:8080",
		ScriptsDir: "./scripts.d",
	}

	// Env vars override defaults.
	if v := os.Getenv("DEVOPS_EYE_LISTEN"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("DEVOPS_EYE_SCRIPTS_DIR"); v != "" {
		cfg.ScriptsDir = v
	}

	// CLI flags override env vars.
	flag.StringVar(&cfg.ListenAddr, "listen", cfg.ListenAddr, "listen address (host:port)")
	flag.StringVar(&cfg.ScriptsDir, "scripts-dir", cfg.ScriptsDir, "path to scripts directory")
	flag.Parse()

	return cfg
}
