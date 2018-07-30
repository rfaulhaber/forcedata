package cmd

import (
	"errors"
	"github.com/rfaulhaber/forcedata/auth"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// authenticateCmd represents the authenticate command
var authenticateCmd = &cobra.Command{
	Use: "authenticate [OPTIONS]",
	// TODO write this
	Short: "Generate REST API credentials",
	Long: `Generates a JSON file containing an access token via the Salesforce REST API based on a username, 
password, client ID and client secret. 

If no arguments are specified, the user will be prompted to input a username, password, client ID, 
and client secret.

If the --file flag is specified with a path to a JSON file, that file will be read and used 
as credentials. You may also pass in the file to stdin without specifying the --stdin flag.

If the --username, --password, --client-id, or --client-secret flags are specified, those 
will be used as credentials and the user will be prompted for anything missing.

If the --stdin flag is specified, the program will attempt to read the password (and only 
the password) from stdin.`,
	Args: validateArgs,
	Run:  runAuthenticate,
}

var (
	fileFlag         bool
	usernameFlag     string
	passFlag         string
	stdinFlag        bool
	clientIDFlag     string
	clientSecretFlag string
	outFlag          string
)

func init() {
	rootCmd.AddCommand(authenticateCmd)

	authenticateCmd.Flags().BoolVar(&fileFlag, "file", false, "Load user credentials from file")
	authenticateCmd.Flags().StringVar(&usernameFlag, "username", "", "Username")
	authenticateCmd.Flags().StringVar(&passFlag, "password", "", "Password")
	authenticateCmd.Flags().StringVar(&clientIDFlag, "client-id", "", "Client ID, the Consumer Key field to the connected app.")
	authenticateCmd.Flags().StringVar(&clientSecretFlag, "client-secret", "", "Client Secret, the Consumer Secret field to the connected app.")
	authenticateCmd.Flags().BoolVar(&stdinFlag, "stdin", false, "Read password from stdin")

	// TODO should only be specified with prompt?
	authenticateCmd.Flags().StringVar(&outFlag, "out", "", "Writes saved session info to specified file instead of stdout")
}

func runAuthenticate(cmd *cobra.Command, args []string) {
	// if file specified, authenticate from file. file must either have client key and url or username, pass, and url
	if fileFlag {
		session, err := auth.AuthenticateFromFile(args[0])

		if err != nil {
			switch err.(type) {
			case auth.MissingFieldError:
				log.Println("A required field for authentication is missing from your file. ", err.Error())
			default:
				log.Println("something went wrong, worhthwile error messages aren't implemented yet!")
				log.Fatalln("error message:", err)
			}
		}

		writeOut(session)
	} else {
		stdWriter.Println("This should trigger authentication via prompt, but it isn't implemented yet!")
		os.Exit(1)
	}
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if fileFlag && len(args) != 1 {
		return errors.New("if the --file flag is specified, the only argument should be the path to the authentication file")
	} else if len(args) > 3 {
		return errors.New("args should only be username, password, and the login URL")
	}

	return nil
}

func writeOut(session auth.Session) {
	if outFlag != "" {
		outFile, err := os.Create(outFlag)

		if err != nil {
			verbose.Println("os.Create encountered error")
			log.Fatalln(err)
		}

		auth.WriteSession(session, outFile)
	} else {
		auth.WriteSession(session, os.Stdout)
	}
}
