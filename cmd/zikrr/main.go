package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sachin-duhan/zikrr/internal/auth"
	"github.com/sachin-duhan/zikrr/internal/cli/tui"
	"github.com/sachin-duhan/zikrr/internal/github"
	"github.com/sachin-duhan/zikrr/pkg/util"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zikrr",
	Short: "Zikrr - GitHub Organization Repository Cloner",
	Long: `Zikrr is a powerful command-line tool for cloning GitHub organization repositories.
It provides interactive selection, concurrent cloning, and multi-branch support.
Complete documentation is available at https://github.com/sachin-duhan/zikrr`,
	Version: "0.1.0",
	RunE:    run,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.zikrr.yaml)")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringP("output", "o", "", "output format for summary (json, yaml)")
	rootCmd.PersistentFlags().StringP("token", "t", "", "GitHub personal access token (can also be set via GITHUB_TOKEN env)")
	rootCmd.PersistentFlags().StringP("org", "g", "", "GitHub organization name")
}

func run(cmd *cobra.Command, args []string) error {
	// Initialize logger
	logLevel, _ := cmd.Flags().GetString("log-level")
	if err := util.InitLogger(logLevel, "text", ""); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Get GitHub token
	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		token = auth.GetTokenFromEnv()
	}
	if token == "" {
		return fmt.Errorf("GitHub token not provided. Use --token flag or set GITHUB_TOKEN environment variable")
	}

	// Validate token
	ctx := context.Background()
	authToken, err := auth.ValidateToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid GitHub token: %w", err)
	}

	// Create GitHub client
	client := github.NewClient(ctx, authToken)

	// Create and run TUI
	model := tui.NewModel(ctx, client)

	// If organization is provided via flag, pre-fill it
	if org, _ := cmd.Flags().GetString("org"); org != "" {
		model.SetOrganization(org)
	}

	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
