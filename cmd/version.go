package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

const version = "0.1"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the current version and exits",
	Long:  `Prints the current version and exits.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("data v" + version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
