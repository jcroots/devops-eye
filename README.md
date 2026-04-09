# devops-eye

Lightweight HTTP server written in Go that exposes live diagnostic and debug
information from the host it runs on. Designed for deployment on cloud VMs (EC2,
etc.), containers, or any Linux server.

## Features

- **Built-in modules**: host/container info, networking details (interfaces, IPs,
  routes, DNS), and more.
- **Script runner**: execute diagnostic scripts from a secure, well-known
  directory and view their output in the browser.
- **Simple HTML output**: no JS frameworks, no complex CSS — just clear,
  readable pages with real data.
- **Configurable**: listen address/port via CLI flags, environment variables, or
  config file.

## License

BSD 3-Clause. See [LICENSE](LICENSE).
