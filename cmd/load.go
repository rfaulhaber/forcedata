package cmd

import (
	"github.com/pkg/errors"
	"github.com/rfaulhaber/forcedata/auth"
	"github.com/rfaulhaber/forcedata/job"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

var flags CtxFlags

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load [FILES...]",
	Short: "Generic data loading operation, for inserting, updating, upserting, and deleting records.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRunE: preRunLoad,
	Run:     runLoad,
	Args:    validateCmdArgs,
}

func init() {
	rootCmd.AddCommand(loadCmd)
	loadCmd.Flags().StringVar(&flags.DelimFlag, "delim", ",", "Delimiter used in files.")
	loadCmd.Flags().DurationVar(&flags.WatchFlag, "watch", job.DefaultWatchTime, "Continuously checks server on job progress.")
	loadCmd.Flags().StringVar(&flags.ObjFlag, "object", "", "Object being inserted.")
	loadCmd.Flags().BoolVarP(&flags.InsertFlag, "insert", "i", false, "Operation flag. Specifies insert job.")
	loadCmd.Flags().BoolVarP(&flags.UpdateFlag, "update", "u", false, "Operation flag. Specifies update job.")
	loadCmd.Flags().BoolVar(&flags.UpsertFlag, "upsert", false, "Operation flag. Specifies upsert job.")
	loadCmd.Flags().BoolVarP(&flags.DeleteFlag, "delete", "d", false, "Operation flag. Specifies delete job.")

	loadCmd.MarkFlagRequired("object")
	loadCmd.Flags().Lookup("watch").NoOptDefVal = job.DefaultWatchTime.String()
}

func preRunLoad(cmd *cobra.Command, args []string) error {
	_, err := validateFlags(flags)

	return err
}

func runLoad(cmd *cobra.Command, args []string) {
	session, err := getSession()

	op, _ := validateFlags(flags)

	if err != nil {
		log.Fatalln(err)
	}

	ctx, err := NewRunContext(op, flags, session)

	if cmd.Flags().Changed("watch") {
		ctx.SetWatch()
	}

	if err != nil {
		log.Fatalln(err)
	}

	if err := ctx.Run(args); err != nil {
		log.Fatalln(err)
	}
}

func validateCmdArgs(cmd *cobra.Command, args []string) error {
	info, _ := os.Stdin.Stat()

	if info.Size() > 0 || len(args) == 1 {
		return nil
	} else {
		return errors.New("must either read in CSV content from stdin or specify one file to upload")
	}
}

func validateFlags(flags CtxFlags) (string, error) {
	opMap := map[string]bool{
		"insert": flags.InsertFlag,
		"update": flags.UpdateFlag,
		"upsert": flags.UpsertFlag,
		"delete": flags.DeleteFlag,
	}

	log.Println("opmap", opMap)

	count := 0

	var op string

	for k, v := range opMap {
		if v {
			count++
			op = k
		}
	}

	if count > 1 {
		return "", errors.New("You can only specify one operation flag. See help for details.")
	}

	if count == 0 {
		return "", errors.New("You must specify an operation flag. See help for details.")
	}

	return op, nil
}

func validSession(session auth.Session) (missing []string, ok bool) {
	if session.AccessToken == "" {
		missing = append(missing, "access_token")
	}

	if session.InstanceURL == "" {
		missing = append(missing, "instance_url")
	}

	if session.ID == "" {
		missing = append(missing, "id")
	}

	return missing, len(missing) == 0
}

func getSession() (auth.Session, error) {
	session := auth.Session{}

	if err := viper.Unmarshal(&session); err != nil {
		return session, errors.Wrap(err, "Attempted to parse config file, received following error. It may be missing or improperly formatted.")
	}

	if missing, ok := validSession(session); !ok {
		return session, errors.New("Session info not valid. Missing the following fields: " + strings.Join(missing, ", "))
	}

	return session, nil
}
