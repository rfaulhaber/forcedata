package cmd

import (
	"errors"
	"fmt"
	"github.com/rfaulhaber/force-data/auth"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// authenticateCmd represents the authenticate command
var authenticateCmd = &cobra.Command{
	Use: "authenticate",
	// TODO write this
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: validateArgs,
	Run:  runAuthenticate,
}

var fileFlag bool
var outFlag string

func init() {
	rootCmd.AddCommand(authenticateCmd)

	authenticateCmd.Flags().BoolVar(&fileFlag, "file", false, "if set, loads in username, password, and login URL from a configuration file")
	authenticateCmd.Flags().StringVar(&outFlag, "out", "", "directory to write session file to. if blank, writes to stdout")
}

func runAuthenticate(cmd *cobra.Command, args []string) {
	var config auth.SFConfig
	if fileFlag {
		conf, err := auth.AuthenticateFromFile(args[0])

		if err != nil {
			log.Println("from file error")
			log.Fatalln(err)
		}

		config = conf
	} else {
		config = auth.AuthenticatePrompt(os.Stdin, os.Stdout)
	}

	log.Println("config", config)

	session, err := auth.GetSessionInfo(config)

	if err != nil {
		fmt.Println("something went wrong with getting your session info")
		os.Exit(1)
	}

	writeOut(session)

	fmt.Println(outFlag)
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if fileFlag && len(args) != 1 {
		return errors.New("if the --file flag is specified, the only argument should be the path to the authentication file")
	} else if len(args) > 3 {
		return errors.New("args should only be username, password, and the login URL")
	}

	return nil
}

func writeOut(session auth.SFSession) {
	if len(outFlag) > 0 {
		// TODO if outflag is dir, write to new file in dir
		// TODO if outflag is file, write to file
		outFile, err := os.Create(outFlag + "/session.json")

		if err != nil {
			// only for development
			// TODO implement better error handling here
			log.Println("os create error")
			log.Fatalln(err)
		}

		auth.WriteSession(session, outFile)
	} else {
		auth.WriteSession(session, os.Stdout)
	}
}
