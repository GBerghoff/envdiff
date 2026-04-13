package collector

import (
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// PackageCollector gathers information about installed system packages
type PackageCollector struct {
	PackageNames []string
}

// Collect gathers package information
func (c *PackageCollector) Collect(snap *snapshot.Snapshot) error {
	if len(c.PackageNames) == 0 {
		return nil
	}

	manager := c.detectManager()
	if manager == "" {
		return nil
	}

	snap.Packages = &snapshot.PackageInfo{
		Manager: manager,
		Items:   make(map[string]string),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, name := range c.PackageNames {
		wg.Add(1)
		go func(pkgName string) {
			defer wg.Done()
			version := c.checkPackage(manager, pkgName)
			if version != "" {
				mu.Lock()
				snap.Packages.Items[pkgName] = version
				mu.Unlock()
			}
		}(name)
	}

	wg.Wait()
	return nil
}

func (c *PackageCollector) detectManager() string {
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("brew"); err == nil {
			return "brew"
		}
	case "linux":
		if _, err := exec.LookPath("apt-get"); err == nil {
			return "apt"
		}
		if _, err := exec.LookPath("dnf"); err == nil {
			return "dnf"
		}
		if _, err := exec.LookPath("yum"); err == nil {
			return "yum"
		}
	}
	return ""
}

func (c *PackageCollector) checkPackage(manager, name string) string {
	var cmd *exec.Cmd
	switch manager {
	case "brew":
		cmd = exec.Command("brew", "list", "--versions", name)
	case "apt":
		cmd = exec.Command("dpkg-query", "-W", "-f=${Version}", name)
	case "dnf", "yum":
		cmd = exec.Command("rpm", "-q", "--queryformat", "%{VERSION}", name)
	default:
		return ""
	}

	out, err := cmd.Output()
	if err != nil {
		return "" // Not installed or error
	}

	version := strings.TrimSpace(string(out))
	if manager == "brew" && version != "" {
		// brew list --versions returns "pkgname version"
		parts := strings.Fields(version)
		if len(parts) >= 2 {
			version = parts[1]
		}
	}

	if version == "" {
		return "installed" // Fallback if we can't get version but command succeeded
	}

	return version
}
