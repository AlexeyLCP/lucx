package ssh

import "strings"

// DistroInfo holds the detected OS, architecture, init system, and package manager.
type DistroInfo struct {
	OS     string
	Arch   string
	Init   string
	PkgMgr string
}

// DetectDistro probes the remote host to determine its OS distribution,
// package manager, init system, and CPU architecture.
func DetectDistro(client *Client) (*DistroInfo, error) {
	release, err := client.ReadFile("/etc/os-release")
	if err != nil {
		release, err = client.Exec("lsb_release -a 2>/dev/null")
		if err != nil {
			return &DistroInfo{OS: "unknown", Arch: "unknown", Init: "unknown", PkgMgr: "unknown"}, nil
		}
	}

	info := &DistroInfo{}
	lower := strings.ToLower(release)

	switch {
	case strings.Contains(lower, "openwrt"):
		info.OS = "openwrt"
		info.PkgMgr = "opkg"
	case strings.Contains(lower, "keenetic"):
		info.OS = "keenetic"
		info.PkgMgr = "entware"
	case strings.Contains(lower, "ubuntu"):
		info.OS = "ubuntu"
		info.PkgMgr = "apt"
	case strings.Contains(lower, "debian"):
		info.OS = "debian"
		info.PkgMgr = "apt"
	case strings.Contains(lower, "almalinux"), strings.Contains(lower, "centos"), strings.Contains(lower, "rhel"):
		info.OS = "almalinux"
		info.PkgMgr = "dnf"
	case strings.Contains(lower, "arch"):
		info.OS = "arch"
		info.PkgMgr = "pacman"
	default:
		info.OS = "unknown"
	}

	out, _ := client.Exec("cat /proc/1/comm 2>/dev/null")
	if strings.Contains(out, "systemd") {
		info.Init = "systemd"
	} else if strings.Contains(out, "procd") {
		info.Init = "procd"
	} else {
		info.Init = "openrc"
	}

	arch, _ := client.Exec("uname -m")
	arch = strings.TrimSpace(arch)
	switch arch {
	case "x86_64":
		info.Arch = "amd64"
	case "aarch64":
		info.Arch = "arm64"
	case "mips":
		info.Arch = "mips"
	case "mipsel":
		info.Arch = "mipsle"
	case "armv7l":
		info.Arch = "armv7"
	default:
		info.Arch = arch
	}

	return info, nil
}
