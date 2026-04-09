// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>
// See LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jcroots/devops-eye/config"
	"github.com/jcroots/devops-eye/internal/modules/hostinfo"
	"github.com/jcroots/devops-eye/internal/modules/netinfo"
	"github.com/jcroots/devops-eye/internal/scripts"
	"github.com/jcroots/devops-eye/internal/server"
)

func main() {
	cfg := config.Load()

	srv := server.New(cfg)
	srv.Register(hostinfo.New())
	srv.Register(netinfo.New())

	scriptsMod, err := scripts.New(cfg.ScriptsDir)
	if err != nil {
		log.Fatalf("scripts module: %v", err)
	}
	srv.Register(scriptsMod)

	fmt.Fprintf(os.Stderr, "devops-eye listening on http://%s\n", cfg.ListenAddr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
