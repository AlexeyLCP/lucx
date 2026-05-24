package scanner

import (
	"fmt"
	"strings"

	"github.com/alexeylcp/lucx-core/internal/ssh"
)

// DetectionResult holds the output of scanning a server for existing proxy software.
type DetectionResult struct {
	HasXray          bool   `json:"has_xray"`
	XrayManaged      string `json:"xray_managed"`
	HasOtherProxy    bool   `json:"has_other_proxy"`
	OtherProxyName   string `json:"other_proxy_name"`
	ConflictingPorts []int  `json:"conflicting_ports"`
}

// Detect scans a remote server for existing Xray installations, other proxy
// software, and port conflicts.
func Detect(client *ssh.Client) (*DetectionResult, error) {
	result := &DetectionResult{}

	out, _ := client.Exec("which xray 2>/dev/null && xray version 2>/dev/null | head -1")
	if strings.Contains(out, "Xray") {
		result.HasXray = true
		result.XrayManaged = "standalone"
	}

	out, _ = client.Exec("systemctl is-active x-ui 2>/dev/null || echo not-found")
	if strings.Contains(out, "active") {
		result.XrayManaged = "3x-ui"
	}

	for _, bin := range []string{"sing-box", "amneziawg", "hysteria", "tuic-server"} {
		out, _ := client.Exec("which " + bin + " 2>/dev/null")
		if strings.TrimSpace(out) != "" {
			result.HasOtherProxy = true
			result.OtherProxyName = bin
			break
		}
	}

	for _, port := range []int{443, 8443, 10085, 8080} {
		out, _ := client.Exec(fmt.Sprintf("ss -tlnp | grep ':%d ' 2>/dev/null", port))
		if strings.TrimSpace(out) != "" {
			result.ConflictingPorts = append(result.ConflictingPorts, port)
		}
	}

	return result, nil
}
