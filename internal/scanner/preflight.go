package scanner

import (
	"fmt"

	"github.com/alexeylcp/lucx-core/internal/ssh"
)

// SafetyCheck represents the result of a pre-installation safety scan.
type SafetyCheck struct {
	Safe       bool             `json:"safe"`
	CanInstall bool             `json:"can_install"`
	CanImport  bool             `json:"can_import"`
	Warning    string           `json:"warning,omitempty"`
	Detection  *DetectionResult `json:"detection"`
}

// PreInstallCheck runs Detect and returns a SafetyCheck that tells the caller
// whether it is safe to install, whether an import is possible, or whether
// a warning must be shown first.
func PreInstallCheck(client *ssh.Client) (*SafetyCheck, error) {
	det, err := Detect(client)
	if err != nil {
		return nil, err
	}
	check := &SafetyCheck{Detection: det}

	switch {
	case det.HasXray && det.XrayManaged == "standalone":
		check.CanImport = true
		check.Warning = "Standalone Xray detected. Import instead of reinstalling."

	case det.HasXray && det.XrayManaged == "3x-ui":
		check.Warning = "Xray managed by 3x-UI detected. LucX cannot manage 3x-UI-controlled Xray."

	case det.HasOtherProxy:
		check.Warning = fmt.Sprintf("Existing proxy detected: %s. Remove it first.", det.OtherProxyName)

	default:
		check.Safe = true
		check.CanInstall = true
	}

	return check, nil
}
