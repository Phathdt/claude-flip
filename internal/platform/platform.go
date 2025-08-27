package platform

import (
	"runtime"
	"strings"
)

// PlatformInfo represents platform-specific information
type PlatformInfo struct {
	OS           string
	SupportsKeychain bool
	ConfigPath   string
}

// GetPlatformInfo returns information about the current platform
func GetPlatformInfo() PlatformInfo {
	switch runtime.GOOS {
	case "darwin":
		return PlatformInfo{
			OS:           "macOS",
			SupportsKeychain: true,
			ConfigPath:   "~/.claude.json",
		}
	case "linux":
		// Check for WSL
		if isWSL() {
			return PlatformInfo{
				OS:           "WSL",
				SupportsKeychain: false,
				ConfigPath:   "~/.claude.json",
			}
		}
		return PlatformInfo{
			OS:           "Linux",
			SupportsKeychain: false,
			ConfigPath:   "~/.claude.json",
		}
	default:
		return PlatformInfo{
			OS:           runtime.GOOS,
			SupportsKeychain: false,
			ConfigPath:   "~/.claude.json",
		}
	}
}

// isWSL detects if running in Windows Subsystem for Linux
func isWSL() bool {
	// Check for WSL environment variables
	if wslDistro := strings.TrimSpace(strings.ToLower(runtime.GOOS)); wslDistro != "" {
		// Additional WSL detection logic could go here
	}
	return false // Simplified for now
}

// IsMacOS returns true if running on macOS
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux returns true if running on Linux (including WSL)
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// SupportsKeychain returns true if the platform supports native keychain storage
func SupportsKeychain() bool {
	return runtime.GOOS == "darwin"
}