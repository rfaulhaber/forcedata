package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
)

var (
	cfgFile     string
	quietFlag   bool
	verboseFlag bool

	verbose   = log.New(ioutil.Discard, "", 0)
	stdWriter = log.New(os.Stdout, "", 0)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "data [OPTIONS] COMMAND",
	Short: "CLI tool for the Salesforce Bulk API",
	Long:  `A CLI tool that allows Salesforce developers to do data loads from the terminal.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if verboseFlag {
			verbose = log.New(os.Stderr, "", 0)
		}

		if quietFlag {
			stdWriter = log.New(ioutil.Discard, "", 0)
		}

		log.SetFlags(0)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is ./config.json)")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Suppresses all output to stdout")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Prints debug logs to stderr.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(".")
		viper.AddConfigPath(home + "/.config/forcedata")
		viper.AddConfigPath("/etc/forcedata")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	viper.ReadInConfig()
}
