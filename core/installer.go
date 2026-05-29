package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InstallTool installs a tool using its defined install methods
func InstallTool(tool ToolDefinition, opts InstallOptions) error {
	if len(tool.Install) == 0 {
		return fmt.Errorf("no install methods defined for %s", tool.Name)
	}

	fmt.Printf("Installing %s...\n", tool.Name)

	// Try each install method in order, use first that works
	for _, inst := range tool.Install {
		if opts.Method != "" && inst.Method != opts.Method {
			continue
		}

		fmt.Printf("  Method: %s\n", inst.Method)

		if err := runInstallMethod(inst, tool.Name); err != nil {
			fmt.Printf("  ! Failed: %v\n", err)
			if opts.Method != "" {
				return err
			}
			continue
		}

		fmt.Printf("  ✓ %s installed successfully\n", tool.Name)
		return nil
	}

	return fmt.Errorf("all install methods failed for %s", tool.Name)
}

// InstallAllTools installs all tools that are not already installed
func InstallAllTools(tools []ToolDefinition, opts InstallOptions) (int, error) {
	var installed int
	for _, t := range tools {
		status := DetectTool(t)
		if status.OnPath || status.DockerImage || status.PackageManager != "" {
			fmt.Printf("  • %s already installed\n", t.Name)
			continue
		}
		if err := InstallTool(t, opts); err != nil {
			fmt.Printf("  ✗ %s: %v\n", t.Name, err)
			continue
		}
		installed++
	}
	return installed, nil
}

type InstallOptions struct {
	Method string // If set, only use this install method
	DryRun bool
	Yes    bool // Auto-confirm
}

func runInstallMethod(inst Install, toolName string) error {
	for _, cmdStr := range inst.Commands {
		if cmdStr == "" {
			continue
		}

		fmt.Printf("    → %s\n", cmdStr)

		var cmd *exec.Cmd
		if strings.Contains(cmdStr, "&&") || strings.Contains(cmdStr, "|") || strings.Contains(cmdStr, ";") {
			cmd = exec.Command("sh", "-c", cmdStr)
		} else {
			parts := strings.Fields(cmdStr)
			if len(parts) == 0 {
				continue
			}
			cmd = exec.Command(parts[0], parts[1:]...)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}
	return nil
}
