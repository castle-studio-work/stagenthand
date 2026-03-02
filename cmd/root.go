// Package cmd implements the shand CLI using cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/baochen10luo/stagenthand/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	dryRun  bool
	cfg     *config.Config
)

// rootCmd is the base command for shand.
var rootCmd = &cobra.Command{
	Use:   "shand",
	Short: "StagentHand — CLI-first AI short drama pipeline",
	Long: `StagentHand (shand) is a CLI tool for generating AI-powered short dramas.
Each subcommand reads JSON from stdin and writes JSON to stdout.
Use --dry-run to validate without calling external APIs.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return nil
	},
}

// Execute runs the root command. Called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.shand/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "validate without calling external APIs")

	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(checkpointCmd)
}
