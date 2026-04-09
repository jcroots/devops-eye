// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>
// See LICENSE file.

package hostinfo

import (
	"bufio"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/jcroots/devops-eye/internal/render"
)

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Name() string        { return "Host Info" }
func (m *Module) Path() string        { return "/hostinfo" }
func (m *Module) Handler() http.Handler {
	return http.HandlerFunc(m.handle)
}

func (m *Module) handle(w http.ResponseWriter, r *http.Request) {
	var b strings.Builder
	b.WriteString("<table>\n")

	writeRow := func(key, val string) {
		b.WriteString(fmt.Sprintf("<tr><th>%s</th><td>%s</td></tr>\n", key, template.HTMLEscapeString(val)))
	}

	// Hostname
	hostname, _ := os.Hostname()
	writeRow("Hostname", hostname)

	// OS / Distro
	writeRow("OS", readOSRelease())

	// Kernel
	writeRow("Kernel", readFileOneLine("/proc/version"))

	// Architecture
	writeRow("Architecture", runtime.GOARCH)

	// Uptime
	writeRow("Uptime", readUptime())

	// CPU
	writeRow("CPU Count", fmt.Sprintf("%d", runtime.NumCPU()))

	// Memory
	memTotal, memAvail := readMemInfo()
	writeRow("Memory Total", memTotal)
	writeRow("Memory Available", memAvail)

	// Container detection
	writeRow("Container", detectContainer())

	b.WriteString("</table>\n")
	render.Page(w, "Host Info", template.HTML(b.String()))
}

func readFileOneLine(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	line, _, _ := strings.Cut(string(data), "\n")
	return line
}

func readOSRelease() string {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	defer f.Close()

	var name, version string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if k, v, ok := strings.Cut(line, "="); ok {
			v = strings.Trim(v, `"`)
			switch k {
			case "NAME":
				name = v
			case "VERSION":
				version = v
			}
		}
	}
	if name == "" {
		return "unknown"
	}
	if version != "" {
		return name + " " + version
	}
	return name
}

func readUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return "unknown"
	}
	var secs float64
	fmt.Sscanf(fields[0], "%f", &secs)
	d := time.Duration(secs) * time.Second
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
}

func readMemInfo() (total, avail string) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return "error", "error"
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if k, v, ok := strings.Cut(line, ":"); ok {
			v = strings.TrimSpace(v)
			switch k {
			case "MemTotal":
				total = v
			case "MemAvailable":
				avail = v
			}
		}
	}
	if total == "" {
		total = "unknown"
	}
	if avail == "" {
		avail = "unknown"
	}
	return total, avail
}

func detectContainer() string {
	// Check common container indicators.
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "Docker"
	}
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return "Podman"
	}
	// Check cgroup for container hints.
	data, err := os.ReadFile("/proc/1/cgroup")
	if err == nil {
		content := string(data)
		if strings.Contains(content, "docker") {
			return "Docker (cgroup)"
		}
		if strings.Contains(content, "kubepods") {
			return "Kubernetes"
		}
		if strings.Contains(content, "lxc") {
			return "LXC"
		}
	}
	return "none detected"
}
