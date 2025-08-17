package utils

import (
	"runtime"
	"suppercommand/internal/types"
)

// GetCurrentPlatform returns the current platform
func GetCurrentPlatform() types.Platform {
	switch runtime.GOOS {
	case "windows":
		return types.PlatformWindows
	case "linux":
		return types.PlatformLinux
	case "darwin":
		return types.PlatformDarwin
	default:
		return types.PlatformLinux // Default fallback
	}
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return GetCurrentPlatform() == types.PlatformWindows
}

// IsLinux returns true if running on Linux
func IsLinux() bool {
	return GetCurrentPlatform() == types.PlatformLinux
}

// IsDarwin returns true if running on macOS
func IsDarwin() bool {
	return GetCurrentPlatform() == types.PlatformDarwin
}

// IsUnix returns true if running on a Unix-like system
func IsUnix() bool {
	platform := GetCurrentPlatform()
	return platform == types.PlatformLinux || platform == types.PlatformDarwin
}

// GetPlatformString returns the platform as a string
func GetPlatformString() string {
	return string(GetCurrentPlatform())
}

// SupportsPlatform checks if a command supports the current platform
func SupportsPlatform(supportedPlatforms []string) bool {
	currentPlatform := GetPlatformString()
	for _, platform := range supportedPlatforms {
		if platform == currentPlatform {
			return true
		}
	}
	return false
}
