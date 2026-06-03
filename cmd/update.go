package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "Check for new version",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Current version: %s\n", Version)
		fmt.Println("Checking for updates...")

		latest, url, err := getLatestRelease()
		if err != nil {
			return fmt.Errorf("checking updates: %w", err)
		}

		if latest == Version || latest == "v"+Version || "v"+latest == Version {
			fmt.Println("✅ You are on the latest version.")
			return nil
		}

		fmt.Printf("🆕 New version available: %s\n", latest)
		fmt.Printf("   Download: %s\n", url)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func getLatestRelease() (string, string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/thanhtaivtt/dbbackup/releases/latest", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "dbbackup/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	return release.TagName, release.HTMLURL, nil
}
