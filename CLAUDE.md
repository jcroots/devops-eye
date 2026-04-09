# devops-eye

Observability toolkit for DevOps/SysAdmin daily work. A lightweight HTTP server
written in Go that exposes debug and diagnostic information from the host it runs
on, designed to be deployed on cloud VMs (EC2, etc.), containers, or any Linux
server.

## Project Goals

- **Simple HTTP server** listening on a configurable TCP port, serving HTML pages
  with diagnostic information that helps understand what is working, what is not,
  and why.
- **Built-in modules** that are always available (networking, host/container info,
  etc.) providing real, live data from the running system.
- **Script runner** that can execute user-provided scripts from a well-known,
  configured directory and present their output as HTML. Security is critical:
  only scripts located in the designated directory are allowed to run; arbitrary
  command execution is never permitted.
- **Minimal and clear HTML output**: no JavaScript frameworks, no complex CSS.
  Pages should be easy to read and understand at a glance.

## Architecture

```
cmd/devops-eye/       # main entrypoint
internal/
  server/             # HTTP server setup, routing, module registry
  modules/            # built-in diagnostic modules
    hostinfo/         # hostname, OS, kernel, uptime, container detection
    netinfo/          # interfaces, IPs, routes, DNS, connectivity checks
    ...               # future modules
  scripts/            # script runner: discovery, validation, execution
  render/             # HTML rendering/templating utilities
config/               # configuration loading (CLI flags, env vars)
scripts.d/            # default well-known directory for user scripts
```

### HTTP Server

- Uses Go standard library `net/http` (no external frameworks).
- Configurable listen address/port via CLI flags or environment variables
  (flags take precedence over env vars, env vars over defaults).
- Each module registers its own HTTP handler(s) under a path prefix.
- The index page (`/`) lists all available modules and scripts with links.

### Built-in Modules

Each module is a self-contained package that:
1. Gathers live data from the system (reading /proc, /sys, running safe
   commands, calling OS APIs).
2. Returns structured data that the render layer turns into HTML.
3. Registers its routes with the server at startup.

Current modules:
- **hostinfo**: hostname, OS/distro, kernel version, uptime, CPU/memory summary,
  container runtime detection (Docker, Podman, Kubernetes, LXC).
- **netinfo**: network interfaces, IP addresses, routing table, DNS resolvers,
  listening ports (TCP/TCP6).

### Script Runner

- Scripts live in a **single, well-known directory** (default: `./scripts.d/`,
  configurable).
- Only regular files with the executable bit set inside that directory are
  discovered and made available.
- No path traversal, no symlink following outside the scripts directory, no
  shell expansion of user input.
- Scripts are executed with a timeout and their stdout/stderr is captured and
  rendered as HTML.
- The script list is discovered at startup (or on reload), not on every request.

### HTML Rendering

- Go `html/template` for all output.
- A shared base layout with a simple, clean style (minimal inline CSS).
- Each module/script result is rendered into the base layout.
- No JavaScript required for core functionality.

## Tech Stack & Conventions

- **Language**: Go (minimum 1.22).
- **License**: BSD 3-Clause.
- **Dependencies**: prefer the Go standard library. External dependencies only
  when they provide significant value and are well-maintained.
- **Build**: `make build` compiles the binary. `make test` runs all tests.
  `make lint` runs static analysis.
- **Code style**: standard `gofmt`/`goimports`. Keep functions short. Avoid
  global state. Pass dependencies explicitly.
- **Error handling**: return errors, don't panic. Log errors with context.
- **Testing**: table-driven tests using the standard `testing` package.
- **Copyright header**: all `.go` files must start with the copyright header:
  `// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>`
  followed by `// See LICENSE file.`
- **Configuration precedence**: CLI flags > environment variables > defaults.

## Maintaining This File

Keep CLAUDE.md up to date whenever changes affect the project's architecture,
conventions, modules, configuration, or any other aspect documented here. Update
proactively — don't wait to be asked.

## Security Considerations

- The script runner is the primary attack surface. Defense in depth:
  - Allowlist by directory (no arbitrary paths).
  - No shell interpretation of URL parameters.
  - Execution timeout enforced.
  - Scripts run as the same user as the server (no privilege escalation).
- The server is intended for **internal/debug use**, not public internet
  exposure. Document this clearly.
- No authentication built-in initially (keep it simple), but the architecture
  should not prevent adding it later.
