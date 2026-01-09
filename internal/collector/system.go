package collector

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// SystemCollector gathers OS and hardware information
type SystemCollector struct{}

// Collect gathers system information
func (c *SystemCollector) Collect(s *snapshot.Snapshot) error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	s.Hostname = hostname
	s.System.Hostname = hostname
	s.System.OS = runtime.GOOS
	s.System.Arch = runtime.GOARCH
	s.System.CPUCores = runtime.NumCPU()

	// Get OS version
	s.System.OSVersion = c.getOSVersion()

	// Get kernel version
	s.System.Kernel = c.getKernel()

	// Get memory
	s.System.MemoryGB = c.getMemoryGB()

	return nil
}

func (c *SystemCollector) getOSVersion() string {
	switch runtime.GOOS {
	case "linux":
		return c.getLinuxOSVersion()
	case "darwin":
		return c.getMacOSVersion()
	case "windows":
		return c.getWindowsVersion()
	default:
		return runtime.GOOS
	}
}

func (c *SystemCollector) getLinuxOSVersion() string {
	// Try /etc/os-release first
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "Linux"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var prettyName string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			prettyName = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			break
		}
	}
	if prettyName != "" {
		return prettyName
	}
	return "Linux"
}

func (c *SystemCollector) getMacOSVersion() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "macOS"
	}
	return "macOS " + strings.TrimSpace(string(out))
}

func (c *SystemCollector) getWindowsVersion() string {
	out, err := exec.Command("cmd", "/c", "ver").Output()
	if err != nil {
		return "Windows"
	}
	return strings.TrimSpace(string(out))
}

func (c *SystemCollector) getKernel() string {
	out, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (c *SystemCollector) getMemoryGB() int {
	switch runtime.GOOS {
	case "linux":
		return c.getLinuxMemory()
	case "darwin":
		return c.getMacMemory()
	default:
		return 0
	}
}

func (c *SystemCollector) getLinuxMemory() int {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			// Format: "MemTotal:       16384000 kB"
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					return int(kb / 1024 / 1024) // Convert KB to GB
				}
			}
		}
	}
	return 0
}

func (c *SystemCollector) getMacMemory() int {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 0
	}
	bytes, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0
	}
	return int(bytes / 1024 / 1024 / 1024) // Convert bytes to GB
}
