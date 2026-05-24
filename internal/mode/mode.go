package mode

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

type RunMode string

const (
	ModeDesktop RunMode = "desktop"
	ModeServer  RunMode = "server"
	ModeRouter  RunMode = "router"
)

type Config struct {
	Mode           RunMode
	MonitorEnabled bool
	MaxSSHConns    int
	GOMEMLIMIT     string
	GOGC           int
}

var current *Config

func Init(forced RunMode) *Config {
	cfg := &Config{Mode: forced}
	if forced == "" {
		cfg.Mode = detect()
	}

	switch cfg.Mode {
	case ModeDesktop:
		cfg.MonitorEnabled = true
		cfg.MaxSSHConns = 10
		cfg.GOMEMLIMIT = ""
		cfg.GOGC = 100
	case ModeServer:
		cfg.MonitorEnabled = true
		cfg.MaxSSHConns = 5
		cfg.GOMEMLIMIT = "64MiB"
		cfg.GOGC = 50
	case ModeRouter:
		cfg.MonitorEnabled = false
		cfg.MaxSSHConns = 1
		cfg.GOMEMLIMIT = "32MiB"
		cfg.GOGC = 50
	}

	applyLimits(cfg)
	current = cfg
	return cfg
}

func Current() *Config { return current }

func detect() RunMode {
	// Check for router indicators
	if _, err := os.Stat("/etc/openwrt_release"); err == nil {
		return ModeRouter
	}
	if _, err := os.Stat("/etc/keenetic_version"); err == nil {
		return ModeRouter
	}

	// Check total RAM — if < 512 MB, treat as router
	mem := totalMemoryMB()
	if mem > 0 && mem < 512 {
		return ModeRouter
	}

	// Check if running in a headless server environment
	if os.Getenv("LUCX_MODE") == "server" {
		return ModeServer
	}

	return ModeDesktop
}

func totalMemoryMB() int {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return -1
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, _ := strconv.Atoi(fields[1])
				return kb / 1024
			}
		}
	}
	return -1
}

func applyLimits(cfg *Config) {
	if cfg.GOMEMLIMIT != "" {
		os.Setenv("GOMEMLIMIT", cfg.GOMEMLIMIT)
	}
	if cfg.GOGC != 100 {
		os.Setenv("GOGC", strconv.Itoa(cfg.GOGC))
	}
	runtime.GC()
}
