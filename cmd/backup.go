package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thanhtaivtt/dbbackup/internal/config"
	"github.com/thanhtaivtt/dbbackup/internal/engine"
)

var (
	flagCompress       *bool
	flagRetentionCount *int
	flagRetentionDays  *int
)

var backupCmd = &cobra.Command{
	Use:          "backup",
	Short:        "Run database backup",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		// CLI flag overrides
		if cmd.Flags().Changed("compress") {
			cfg.Backup.Compress = *flagCompress
		}
		if cmd.Flags().Changed("retention-count") {
			cfg.Retention.Strategy = "count"
			cfg.Retention.Count = *flagRetentionCount
		}
		if cmd.Flags().Changed("retention-days") {
			cfg.Retention.Strategy = "days"
			cfg.Retention.Days = *flagRetentionDays
		}

		level := slog.LevelInfo
		if verbose {
			level = slog.LevelDebug
		}
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

		if verbose {
			logger.Debug("loaded config", "file", cfgFile, "databases", cfg.Database.MySQL.Databases, "compress", cfg.Backup.Compress, "dump_method", cfg.Backup.DumpMethod, "retention_strategy", cfg.Retention.Strategy)
		}

		eng, err := engine.New(cfg, logger)
		if err != nil {
			return err
		}

		return eng.Run(context.Background())
	},
}

func init() {
	flagCompress = backupCmd.Flags().Bool("compress", false, "enable gzip compression")
	flagRetentionCount = backupCmd.Flags().Int("retention-count", 0, "keep N latest backups")
	flagRetentionDays = backupCmd.Flags().Int("retention-days", 0, "keep backups for N days")
	rootCmd.AddCommand(backupCmd)
}
