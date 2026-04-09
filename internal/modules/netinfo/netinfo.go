// Copyright (c) Jeremías Casteglione <jeremias.rootstrap@gmail.com>
// See LICENSE file.

package netinfo

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/jcroots/devops-eye/internal/render"
)

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Name() string        { return "Network Info" }
func (m *Module) Path() string        { return "/netinfo" }
func (m *Module) Handler() http.Handler {
	return http.HandlerFunc(m.handle)
}

func (m *Module) handle(w http.ResponseWriter, r *http.Request) {
	var b strings.Builder

	writeInterfaces(&b)
	writeRoutes(&b)
	writeDNS(&b)
	writeListeningPorts(&b)

	render.Page(w, "Network Info", template.HTML(b.String()))
}

func writeInterfaces(b *strings.Builder) {
	b.WriteString("<h2>Interfaces</h2>\n<table>\n")
	b.WriteString("<tr><th>Name</th><th>Flags</th><th>MTU</th><th>Addresses</th></tr>\n")

	ifaces, err := net.Interfaces()
	if err != nil {
		b.WriteString(fmt.Sprintf("<tr><td colspan=\"4\">error: %v</td></tr>\n", err))
		b.WriteString("</table>\n")
		return
	}

	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		addrStrs := make([]string, 0, len(addrs))
		for _, a := range addrs {
			addrStrs = append(addrStrs, a.String())
		}
		b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%d</td><td>%s</td></tr>\n",
			template.HTMLEscapeString(iface.Name),
			template.HTMLEscapeString(iface.Flags.String()),
			iface.MTU,
			template.HTMLEscapeString(strings.Join(addrStrs, ", ")),
		))
	}
	b.WriteString("</table>\n")
}

func writeRoutes(b *strings.Builder) {
	b.WriteString("<h2>Routes</h2>\n<table>\n")
	b.WriteString("<tr><th>Destination</th><th>Gateway</th><th>Mask</th><th>Iface</th></tr>\n")

	f, err := os.Open("/proc/net/route")
	if err != nil {
		b.WriteString(fmt.Sprintf("<tr><td colspan=\"4\">error: %v</td></tr>\n", err))
		b.WriteString("</table>\n")
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip header
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 8 {
			continue
		}
		iface := fields[0]
		dest := hexToIP(fields[1])
		gw := hexToIP(fields[2])
		mask := hexToIP(fields[7])
		b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
			dest, gw, mask, template.HTMLEscapeString(iface)))
	}
	b.WriteString("</table>\n")
}

func hexToIP(s string) string {
	if len(s) != 8 {
		return s
	}
	decoded, err := hex.DecodeString(s)
	if err != nil || len(decoded) != 4 {
		return s
	}
	// /proc/net/route uses little-endian.
	return fmt.Sprintf("%d.%d.%d.%d", decoded[3], decoded[2], decoded[1], decoded[0])
}

func writeDNS(b *strings.Builder) {
	b.WriteString("<h2>DNS Resolvers</h2>\n<table>\n")
	b.WriteString("<tr><th>Resolver</th></tr>\n")

	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		b.WriteString(fmt.Sprintf("<tr><td>error: %v</td></tr>\n", err))
		b.WriteString("</table>\n")
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				b.WriteString(fmt.Sprintf("<tr><td>%s</td></tr>\n",
					template.HTMLEscapeString(fields[1])))
			}
		}
	}
	b.WriteString("</table>\n")
}

func writeListeningPorts(b *strings.Builder) {
	b.WriteString("<h2>Listening Ports</h2>\n<table>\n")
	b.WriteString("<tr><th>Proto</th><th>Local Address</th></tr>\n")

	parseProcNet(b, "/proc/net/tcp", "tcp")
	parseProcNet(b, "/proc/net/tcp6", "tcp6")

	b.WriteString("</table>\n")
}

func parseProcNet(b *strings.Builder, path, proto string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip header
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}
		// st field (connection state): 0A = LISTEN
		if fields[3] != "0A" {
			continue
		}
		localAddr := parseHexAddr(fields[1], proto)
		b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>\n", proto, localAddr))
	}
}

func parseHexAddr(s, proto string) string {
	host, port, ok := strings.Cut(s, ":")
	if !ok {
		return s
	}
	var portNum uint64
	fmt.Sscanf(port, "%X", &portNum)

	if proto == "tcp6" {
		return fmt.Sprintf("[::]:%d", portNum)
	}

	ip := hexToIP(host)
	return fmt.Sprintf("%s:%d", ip, portNum)
}
