package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "automate",
	Short: "AutoMate - Security tool automation CLI",
	Long: `AutoMate is a CLI tool that syncs with the AutoTest knowledge site,
detects installed tools, installs missing ones, and executes
tools with parameterized commands.

  automate list        List all available tools
  automate status      Check which tools are installed
  automate sync        Sync tool definitions from AutoTest site
  automate install     Install tools
  automate run         Execute a tool with parameters
  automate search      Search tools by name/capability/domain
  automate exec        Open an interactive terminal session
  automate mcp         Start MCP server for AI integration
  automate tui         Launch the interactive TUI
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
}
