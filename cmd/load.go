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
	"io/ioutil"
	"io"
	"time"
	"fmt"
)

type flagStr struct {
	objFlag    string
	delimFlag  string
	watchFlag  time.Duration
	insertFlag bool
	updateFlag bool
	upsertFlag bool
	deleteFlag bool
}

var flags flagStr

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:     "load [FILES...]",
	Short:   "Load data from a CSV file.",
	Long:    `Generic data loading operation, for inserting, updating, upserting, and deleting records.`,
	PreRunE: preRunLoad,
	Run:     runLoad,
	Args:    validateCmdArgs,
}

func init() {
	rootCmd.AddCommand(loadCmd)
	loadCmd.Flags().StringVar(&flags.delimFlag, "delim", ",", "Delimiter used in files.")
	loadCmd.Flags().DurationVar(&flags.watchFlag, "watch", job.DefaultWatchTime, "Continuously checks server on job progress.")
	loadCmd.Flags().StringVar(&flags.objFlag, "object", "", "Object being inserted.")
	loadCmd.Flags().BoolVarP(&flags.insertFlag, "insert", "i", false, "Operation flag. Specifies insert job.")
	loadCmd.Flags().BoolVarP(&flags.updateFlag, "update", "u", false, "Operation flag. Specifies update job.")
	loadCmd.Flags().BoolVar(&flags.upsertFlag, "upsert", false, "Operation flag. Specifies upsert job.")
	loadCmd.Flags().BoolVarP(&flags.deleteFlag, "delete", "d", false, "Operation flag. Specifies delete job.")

	loadCmd.MarkFlagRequired("object")
	loadCmd.Flags().Lookup("watch").NoOptDefVal = job.DefaultWatchTime.String()
}

func preRunLoad(cmd *cobra.Command, args []string) error {
	_, err := validateFlags(flags)
	return err
}

func runLoad(cmd *cobra.Command, args []string) {
	session, err := getSession()

	if err != nil {
		log.Fatalln(err)
	}

	op, _ := validateFlags(flags)

	delim, ok := job.GetDelimName(flags.delimFlag)

	if !ok {
		log.Fatalln(errors.Errorf("Invalid delimiter: %s", flags.delimFlag))
	}

	config := job.JobConfig{
		Object:      flags.objFlag,
		Operation: op,
		Delim: delim,
		ContentType: "CSV",
	}

	verbose.Println("creating job...")

	j := job.New(config, session)

	if err := j.Create(); err != nil {
		log.Fatalln("could not create job:", err)
	}

	verbose.Println("uploading content...")
	var content []byte

	if isPipeInput() {
		content, err = readSource(os.Stdin)
	} else {
		// we only support uploading one file at a time at the moment!
		content, err = ioutil.ReadFile(args[0])
	}

	if err != nil {
		log.Fatalln("could not read source of ")
	}

	if err = j.Upload(content); err != nil {
		log.Fatalln("could not upload content to job")
	}

	if cmd.Flags().Changed("watch") {
		go j.Watch(flags.watchFlag)

		for {
			select {
			case status, ok := <-j.Status:
				if ok {
					printStatus(status)
				} else {
					return
				}
			case err := <-j.Error:
				log.Fatalln("watching job reported error: ", err)
			}
		}
	}
}

func validateCmdArgs(cmd *cobra.Command, args []string) error {
	if !isPipeInput() && len(args) != 1 {
		fmt.Println("is pipe", isPipeInput())
		return errors.New("must either read in CSV content from stdin or specify one file to upload")
	}

	return nil
}

func validateFlags(flags flagStr) (string, error) {
	opMap := map[string]bool{
		"insert": flags.insertFlag,
		"update": flags.updateFlag,
		"upsert": flags.upsertFlag,
		"delete": flags.deleteFlag,
	}

	count := 0

	var op string

	for k, v := range opMap {
		if v {
			count++
			op = k
		}
	}

	if count > 1 {
		return "", errors.New("You can only specify one operation flag.")
	}

	if count == 0 {
		return "", errors.New("You must specify an operation flag.")
	}

	return op, nil
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

func printStatus(status job.JobInfo) {
	stdWriter.Printf("Records processed: %d\tRecords failed: %d", status.RecordsProcessed, status.RecordsFailed)
}

// reads content from source
func readSource(source io.ReadCloser) ([]byte, error) {
	content, err := ioutil.ReadAll(source)

	if err != nil {
		return nil, errors.Wrap(err, "attempting to read file source")
	}

	return content, nil
}

func isPipeInput() bool {
	stat, err := os.Stdin.Stat()
	return stat.Mode()&os.ModeCharDevice != 0 || stat.Size() <= 0 && err == nil
}
