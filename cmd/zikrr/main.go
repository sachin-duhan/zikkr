package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zikrr",
	Short: "Zikrr - GitHub Organization Repository Cloner",
	Long: `Zikrr is a powerful command-line tool for cloning GitHub organization repositories.
It provides interactive selection, concurrent cloning, and multi-branch support.
Complete documentation is available at https://github.com/sachin-duhan/zikrr`,
	Version: "0.1.0",
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.zikrr.yaml)")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringP("output", "o", "", "output format for summary (json, yaml)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
