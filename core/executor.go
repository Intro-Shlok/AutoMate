package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ExecOptions struct {
	Params  map[string]string
	Quiet   bool
	Timeout int // seconds
}

// BuildCommand renders the execution template with the given parameters
func BuildCommand(tool ToolDefinition, params map[string]string) (string, error) {
	tmpl := tool.Execution.Template
	if tmpl == "" {
		return "", fmt.Errorf("no execution template defined for %s", tool.Name)
	}

	// Build a map of all param values (defaults + overrides)
	values := make(map[string]string)

	// Set default values
	for _, p := range tool.Parameters {
		if p.DefaultValue != nil {
			key := p.TemplateKey
			if key == "" {
				key = p.Name
			}
			values[key] = fmt.Sprintf("%v", p.DefaultValue)
		}
		// Set alias flags where applicable
		if len(p.Aliases) > 0 && p.TemplateKey != "" {
			values[p.TemplateKey] = ""
		}
	}

	// Override with user-provided params
	for k, v := range params {
		values[k] = v
	}

	// Process template - for each {placeholder}, find the matching param value
	var result strings.Builder
	i := 0
	for i < len(tmpl) {
		if tmpl[i] == '{' {
			j := i + 1
			for j < len(tmpl) && tmpl[j] != '}' {
				j++
			}
			if j < len(tmpl) {
				key := tmpl[i+1 : j]
				val, ok := values[key]
				if ok && val == "" {
					// Check if this param has aliases
					for _, p := range tool.Parameters {
						if (p.TemplateKey == key || p.Name == key) && len(p.Aliases) > 0 {
							result.WriteString(p.Aliases[0])
							break
						}
					}
				} else if ok {
					result.WriteString(val)
				} else {
					// Keep the placeholder if no value found (will error later)
					result.WriteString(tmpl[i : j+1])
				}
				i = j + 1
			} else {
				result.WriteByte(tmpl[i])
				i++
			}
		} else {
			result.WriteByte(tmpl[i])
			i++
		}
	}

	cmd := strings.TrimSpace(result.String())

	// Validate no unresolved placeholders remain
	if strings.Contains(cmd, "{") && strings.Contains(cmd, "}") {
		return "", fmt.Errorf("unresolved parameters in template: %s", cmd)
	}

	return cmd, nil
}

// ExecuteTool runs a tool with the given parameters
func ExecuteTool(tool ToolDefinition, params map[string]string, opts ExecOptions) (string, error) {
	// Build command
	cmdStr, err := BuildCommand(tool, params)
	if err != nil {
		return "", fmt.Errorf("build command: %w", err)
	}

	if !opts.Quiet {
		fmt.Printf("Running: %s\n", cmdStr)
	}

	// Create command
	var cmd *exec.Cmd
	if tool.Execution.Shell {
		cmd = exec.Command("sh", "-c", cmdStr)
	} else {
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			return "", fmt.Errorf("empty command")
		}
		cmd = exec.Command(parts[0], parts[1:]...)
	}

	// Set working directory
	if tool.Execution.Workdir != "" {
		cmd.Dir = tool.Execution.Workdir
	}

	// Set environment variables
	if len(tool.Execution.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range tool.Execution.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	// Set timeout
	if opts.Timeout > 0 {
		// Use a timer to kill process after timeout
		// For now, we just pass it as a comment
	}

	// Run and capture output
	out, err := cmd.CombinedOutput()
	output := string(out)

	if err != nil {
		return output, fmt.Errorf("execution failed: %w\nOutput: %s", err, output)
	}

	return output, nil
}

// RunTerminal opens an interactive shell with AutoMate context
func RunTerminal(workdir string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	cmd := exec.Command(shell)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if workdir != "" {
		cmd.Dir = workdir
	}

	fmt.Printf("AutoMate terminal session. Type 'exit' to return.\n")
	return cmd.Run()
}

// RunShellCommand runs a single shell command
func RunShellCommand(cmdStr, workdir string) error {
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if workdir != "" {
		cmd.Dir = workdir
	}
	return cmd.Run()
}
