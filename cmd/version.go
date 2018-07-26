package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

const version = "0.1-rc1"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the current version and exits",
	Long:  `Prints the current version and exits.`,
	Run:   printVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func printVersion(cmd *cobra.Command, args []string) {
	fmt.Println("ForceData v" + version)
}
