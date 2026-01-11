package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type EnvironmentContext struct {
	CurrentDirectory string
	OSInfo          string
	Shell           string
	User            string
	OS              string
}

func GatherContext() (*EnvironmentContext, error) {
	ctx := &EnvironmentContext{}

	// Get OS type
	ctx.OS = runtime.GOOS

	// Get current working directory
	if wd, err := os.Getwd(); err == nil {
		ctx.CurrentDirectory = wd
	}

	// Get OS info (platform-specific)
	switch runtime.GOOS {
	case "windows":
		// On Windows, try to get system info
		if out, err := exec.Command("systeminfo").Output(); err == nil {
			// Extract OS name from systeminfo
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, "OS Name:") {
					ctx.OSInfo = strings.TrimSpace(strings.TrimPrefix(line, "OS Name:"))
					break
				}
			}
		}
		// Fallback to just Windows
		if ctx.OSInfo == "" {
			ctx.OSInfo = "Windows"
		}
		// Try to get PowerShell or cmd.exe
		if shell := os.Getenv("PSModulePath"); shell != "" {
			ctx.Shell = "PowerShell"
		} else {
			ctx.Shell = "cmd.exe"
		}
	case "darwin", "linux", "freebsd", "openbsd":
		// Unix-like systems
		if out, err := exec.Command("uname", "-a").Output(); err == nil {
			ctx.OSInfo = strings.TrimSpace(string(out))
		}
		// Get shell
		if shell := os.Getenv("SHELL"); shell != "" {
			ctx.Shell = shell
		} else {
			ctx.Shell = "/bin/bash" // fallback
		}
	default:
		ctx.OSInfo = runtime.GOOS
		ctx.Shell = "unknown"
	}

	// Get user (platform-specific)
	if runtime.GOOS == "windows" {
		if user := os.Getenv("USERNAME"); user != "" {
			ctx.User = user
		}
	} else {
		if user := os.Getenv("USER"); user != "" {
			ctx.User = user
		} else if user := os.Getenv("USERNAME"); user != "" {
			ctx.User = user
		}
	}

	return ctx, nil
}

func (ctx *EnvironmentContext) Format() string {
	var parts []string

	// Always include OS type
	parts = append(parts, fmt.Sprintf("Platform: %s", ctx.OS))

	if ctx.CurrentDirectory != "" {
		parts = append(parts, fmt.Sprintf("Current directory: %s", ctx.CurrentDirectory))
	}

	if ctx.OSInfo != "" {
		parts = append(parts, fmt.Sprintf("OS: %s", ctx.OSInfo))
	}

	if ctx.Shell != "" {
		parts = append(parts, fmt.Sprintf("Shell: %s", ctx.Shell))
	}

	if ctx.User != "" {
		parts = append(parts, fmt.Sprintf("User: %s", ctx.User))
	}

	if len(parts) == 0 {
		return ""
	}

	return "Environment context:\n" + strings.Join(parts, "\n")
}
