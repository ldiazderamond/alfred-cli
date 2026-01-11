package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type ExecutionError struct {
	Command    string
	CommandNum int
	Err        error
	Stdout     string
	Stderr     string
	Combined   string
}

func (e *ExecutionError) Error() string {
	return fmt.Sprintf("command %d failed: %s\nError: %v\nOutput: %s", e.CommandNum, e.Command, e.Err, e.Combined)
}

func ParseCommands(command string) []string {
	// Split by newlines and filter out empty lines
	lines := strings.Split(command, "\n")
	var commands []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			commands = append(commands, line)
		}
	}
	// If no commands found after splitting, return the original as single command
	if len(commands) == 0 {
		return []string{command}
	}
	return commands
}

func ConfirmExecution(commands []string) (bool, error) {
	fmt.Printf("\nGenerated command(s):\n")
	if len(commands) == 1 {
		fmt.Printf("%s\n\n", commands[0])
	} else {
		for i, cmd := range commands {
			fmt.Printf("%d. %s\n", i+1, cmd)
		}
		fmt.Println()
	}
	fmt.Print("Execute these commands? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

func ExecuteCommands(commands []string) error {
	if len(commands) == 0 {
		return fmt.Errorf("no commands to execute")
	}

	// Get current working directory
	startDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("\nExecuting commands...")
	fmt.Println(strings.Repeat("-", 50))

	// Execute commands sequentially, maintaining working directory
	for i, command := range commands {
		if strings.TrimSpace(command) == "" {
			continue
		}

		if len(commands) > 1 {
			fmt.Printf("\n[%d/%d] Executing: %s\n", i+1, len(commands), command)
		}

		var cmd *exec.Cmd

		// Execute through appropriate shell based on OS
		switch runtime.GOOS {
		case "windows":
			// On Windows, prefer PowerShell, fallback to cmd.exe
			if _, err := exec.LookPath("powershell.exe"); err == nil {
				// PowerShell: use -Command flag
				cmd = exec.Command("powershell.exe", "-Command", command)
			} else {
				// cmd.exe: use /c flag
				cmd = exec.Command("cmd.exe", "/c", command)
			}
		case "darwin", "linux", "freebsd", "openbsd":
			// Unix-like systems: use bash or sh
			shell := "/bin/bash"
			if _, err := exec.LookPath("bash"); err != nil {
				shell = "/bin/sh"
			}
			cmd = exec.Command(shell, "-c", command)
		default:
			return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
		}

		// Set working directory to maintain context between commands
		cmd.Dir = startDir
		cmd.Stdin = os.Stdin

		// Capture stdout and stderr for error reporting
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

		if err := cmd.Run(); err != nil {
			stdoutStr := strings.TrimSpace(stdoutBuf.String())
			stderrStr := strings.TrimSpace(stderrBuf.String())
			combined := strings.TrimSpace(stdoutStr + "\n" + stderrStr)

			return &ExecutionError{
				Command:    command,
				CommandNum: i + 1,
				Err:        err,
				Stdout:     stdoutStr,
				Stderr:     stderrStr,
				Combined:   combined,
			}
		}

		// Update working directory for next command if this was a cd command
		cmdParts := strings.Fields(command)
		if len(cmdParts) > 0 && (cmdParts[0] == "cd" || strings.HasPrefix(command, "cd ")) {
			var newDir string
			if len(cmdParts) > 1 {
				newDir = cmdParts[1]
			} else {
				// cd without arguments goes to home directory
				newDir = os.Getenv("HOME")
				if newDir == "" && runtime.GOOS == "windows" {
					newDir = os.Getenv("USERPROFILE")
				}
			}

			if newDir != "" {
				// Handle ~ expansion
				if strings.HasPrefix(newDir, "~") {
					home := os.Getenv("HOME")
					if home == "" && runtime.GOOS == "windows" {
						home = os.Getenv("USERPROFILE")
					}
					if home != "" {
						newDir = filepath.Join(home, strings.TrimPrefix(newDir, "~"))
					}
				}

				// Resolve relative paths
				if !filepath.IsAbs(newDir) {
					newDir = filepath.Join(startDir, newDir)
				}

				// Clean and resolve the path
				if absPath, err := filepath.Abs(newDir); err == nil {
					if info, err := os.Stat(absPath); err == nil && info.IsDir() {
						startDir = absPath
					}
				}
			}
		}
	}

	fmt.Println(strings.Repeat("-", 50))
	if len(commands) > 1 {
		fmt.Printf("All %d commands completed successfully.\n", len(commands))
	} else {
		fmt.Println("Command completed.")
	}

	return nil
}
