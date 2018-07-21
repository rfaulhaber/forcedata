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
var usernameFlag string
var passFlag string
var stdinFlag bool
var clientIDFlag string
var clientSecretFlag string

func init() {
	rootCmd.AddCommand(authenticateCmd)

	authenticateCmd.Flags().BoolVar(&fileFlag, "file", false, "Load user credentials from file")
	authenticateCmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username")
	authenticateCmd.Flags().StringVarP(&passFlag, "password", "p", "", "Password")
	authenticateCmd.Flags().StringVar(&clientIDFlag, "client-id", "", "Client ID, the Consumer Key field to the connected app.")
	authenticateCmd.Flags().StringVar(&clientSecretFlag, "client-secret", "", "Client Secret, the Consumer Secret field to the connected app.")
	authenticateCmd.Flags().BoolVar(&stdinFlag, "stdin", false, "Read password from stdin")
}

func runAuthenticate(cmd *cobra.Command, args []string) {
	// if file specified, authenticate from file. file must either have client key and url or username, pass, and url
	if fileFlag {
		// if file contains username, password, authenticate with credentials
		// else, attempt to authenticate using user agent
	} else {
		// if client id specified, use that
		// else if username, pass, and url flag specified, use those
		// else if only some flags specified, prompt for missing
	}

	// once authenticated, attempt to save default config file
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
