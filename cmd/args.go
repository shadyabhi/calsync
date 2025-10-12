package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type cmdArgs struct {
	config    string
	deleteDst string
}

var rootCmd = &cobra.Command{
	Use:   "calsync",
	Short: "Synchronize calendar events between different calendar sources",
	Long: `CalSync is a Go application that synchronizes calendar events between different
calendar sources (Mac Calendar, Google Calendar, ICS feeds). It reads events
from source calendars and syncs them to target calendars, helping users maintain
a unified view across different calendar systems.`,
	Run: func(cmd *cobra.Command, args []string) {
		deleteDst, err := cmd.Flags().GetString("delete-dst")
		if err != nil {
			fmt.Println("Invalid argument delete-dst", err)
			os.Exit(1)
		}
		config, err := cmd.Flags().GetString("config")
		if err != nil {
			fmt.Println("Invalid argument config", err)
			os.Exit(1)
		}

		cmdArgs := cmdArgs{
			config:    config,
			deleteDst: deleteDst,
		}
		run(cmdArgs)
	},
}

func Execute() {
	rootCmd.Flags().StringP("delete-dst", "", "", "Delete all calsync-managed events from the specified destination calendar (e.g., 'Google')")
	rootCmd.Flags().StringP("config", "c", "", "Location to config.toml. (Default: ~/.config/calsync/config.toml)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
