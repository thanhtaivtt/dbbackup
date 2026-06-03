package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const configTemplate = `[backup]
compress = true
# dump_method = "go" | "binary"
dump_method = "go"

[database.mysql]
host = "127.0.0.1"
port = 3306
user = "root"
password = ""
databases = ["mydb"]

[storage.r2]
account_id = ""
access_key_id = ""
access_key_secret = ""
bucket = ""
path_prefix = "mysql/"

[retention]
# strategy = "count" | "days"
strategy = "count"
count = 7
# days = 30

[notification.telegram]
enabled = false
bot_token = ""
chat_id = ""
`

var initCmd = &cobra.Command{
	Use:          "init",
	Short:        "Generate a config file",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(cfgFile); err == nil {
			return fmt.Errorf("config file already exists: %s", cfgFile)
		}

		if err := os.WriteFile(cfgFile, []byte(configTemplate), 0600); err != nil {
			return fmt.Errorf("writing config file: %w", err)
		}

		fmt.Printf("✅ Config file created: %s\n", cfgFile)
		fmt.Println("   Edit it with your database and storage credentials.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
