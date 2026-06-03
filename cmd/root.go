package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Set via ldflags at build time
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "dbbackup",
	Short: "Database backup tool with cloud storage support",
	Long: `dbbackup - A CLI tool for backing up databases to cloud storage.

Currently supports MySQL → Cloudflare R2, designed to be extensible
for additional databases and storage backends.

Commands:
  init      Generate a config file
  backup    Run database backup
  update    Check for new version on GitHub`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.toml", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose/debug logging")
	rootCmd.SetVersionTemplate(fmt.Sprintf("dbbackup %s (commit: %s, built: %s)\n", Version, Commit, Date))
}
