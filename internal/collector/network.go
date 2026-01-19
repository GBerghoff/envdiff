package collector

import (
	"bufio"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// NetworkCollector gathers network-related information
type NetworkCollector struct{}

// Collect gathers network information
func (c *NetworkCollector) Collect(s *snapshot.Snapshot) error {
	s.Network = &snapshot.NetworkInfo{
		Hosts:          c.getHosts(),
		ListeningPorts: c.getListeningPorts(),
	}
	return nil
}

func (c *NetworkCollector) getHosts() map[string]string {
	hosts := make(map[string]string)

	file, err := os.Open("/etc/hosts")
	if err != nil {
		return hosts
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			ip := fields[0]
			for _, hostname := range fields[1:] {
				// Skip comments at end of line
				if strings.HasPrefix(hostname, "#") {
					break
				}
				hosts[hostname] = ip
			}
		}
	}

	return hosts
}

func (c *NetworkCollector) getListeningPorts() []int {
	switch runtime.GOOS {
	case "linux":
		return c.getLinuxListeningPorts()
	case "darwin":
		return c.getMacListeningPorts()
	default:
		return []int{}
	}
}

func (c *NetworkCollector) getLinuxListeningPorts() []int {
	var ports []int

	// Try ss first, fall back to netstat
	out, err := exec.Command("ss", "-tlnp").Output()
	if err != nil {
		out, err = exec.Command("netstat", "-tlnp").Output()
		if err != nil {
			return ports
		}
	}

	// Parse output for listening ports
	// ss format: "LISTEN  0  128  0.0.0.0:22  ..."
	// netstat format: "tcp  0  0  0.0.0.0:22  0.0.0.0:*  LISTEN  ..."
	portRE := regexp.MustCompile(`:(\d+)\s`)

	seen := make(map[int]bool)
	for _, match := range portRE.FindAllStringSubmatch(string(out), -1) {
		if len(match) >= 2 {
			port, err := strconv.Atoi(match[1])
			if err == nil && !seen[port] {
				ports = append(ports, port)
				seen[port] = true
			}
		}
	}

	return ports
}

func (c *NetworkCollector) getMacListeningPorts() []int {
	var ports []int

	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-nP").Output()
	if err != nil {
		return ports
	}

	// Parse lsof output
	// Format: "node    12345  user  ...  TCP *:3000 (LISTEN)"
	portRE := regexp.MustCompile(`:(\d+)\s+\(LISTEN\)`)

	seen := make(map[int]bool)
	for _, match := range portRE.FindAllStringSubmatch(string(out), -1) {
		if len(match) >= 2 {
			port, err := strconv.Atoi(match[1])
			if err == nil && !seen[port] {
				ports = append(ports, port)
				seen[port] = true
			}
		}
	}

	return ports
}
