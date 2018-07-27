package cmd

import (
	"fmt"
	"github.com/rfaulhaber/forcedata/auth"
	"github.com/rfaulhaber/forcedata/job"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"time"
	"io"
	"os"
	"github.com/pkg/errors"
	"io/ioutil"
)

var (
	objFlag       string
	delimFlag     string
	watchFlag     bool
	watchTimeFlag time.Duration
)

// insertCmd represents the insert command
var insertCmd = &cobra.Command{
	Use:   "insert",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run:  runInsert,
	Args: func(cmd *cobra.Command, args []string) error {
		info, _ := os.Stdin.Stat()

		if info.Size() > 0  || len(args) == 1 {
			return nil
		} else {
			return errors.New("must either read in CSV content from stdin or specify one file to upload")
		}
	},
}

func init() {
	rootCmd.AddCommand(insertCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// insertCmd.PersistentFlags().String("foo", "", "A help for foo")
	insertCmd.Flags().StringVar(&delimFlag, "delim", ",", "Delimiter used in all specified fiels (defaults to \",\")")
	insertCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "Continuously checks server for job progress.")
	insertCmd.Flags().DurationVarP(&watchTimeFlag, "time", "t", job.DefaultWatchTime, "Requires --watch to be set. Frequency with which the job will check status on server. Defaults to 5s.")
	insertCmd.Flags().StringVar(&objFlag, "object", "", "Object being inserted.")
	insertCmd.MarkFlagRequired("object")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// insertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runInsert(cmd *cobra.Command, args []string) {
	// TODO validate config
	session := auth.Session{}
	err := viper.Unmarshal(&session)

	if !isValidSession(session) {
		log.Fatalln("Session info is not valid")
	}

	if err != nil {
		log.Fatalln("viper could not unmarshal config", err)
	}

	// TODO do better!
	delimName, _ := job.GetDelimName(delimFlag)

	config := job.JobConfig{
		Object:      objFlag,
		Operation:   "insert",
		Delim:       delimName,
		ContentType: "CSV", // TODO determine based on MIME type? ending?
	}

	j := job.NewJob(config, session)

	log.Println("creating job...")
	err = j.Create()

	// TODO create job for each arg

	if err != nil {
		log.Fatalln("error creating job:", err)
	}

	log.Println("uploading files...")
	var content []byte

	if isPipe() {
		content, err = ReadSource(os.Stdin)
	} else {
		content, err = ioutil.ReadFile(args[0])
	}

	if err != nil {
		log.Fatalln("could not read source", err)
	}

	err = j.Upload(content)

	if err != nil {
		log.Fatalln("upload error", err)
	}

	if watchFlag {
		go j.Watch(watchTimeFlag)

		for {
			select {
			case status, ok := <- j.Status:
				if ok {
					PrintStatus(status, os.Stdout)
				} else {
					return
				}
			case err := <- j.Error:
				log.Fatalln("watch error: ", err)
				return
			}
		}
	}
}

func PrintStatus(status job.JobInfo, out io.Writer) {
	io.WriteString(out, "")
	fmt.Fprintf(out, "Records processed: %d\tRecords failed: %d", status.RecordsProcessed, status.RecordsFailed)
}

// reads content from source
func ReadSource(source io.ReadCloser) ([]byte, error) {
	content, err := ioutil.ReadAll(source)

	if err != nil {
		return nil, errors.Wrap(err, "attempting to read file source")
	}

	return content, nil
}

func isPipe() bool {
	stat, err := os.Stdin.Stat()
	return err == nil && stat.Size() > 0
}

func isValidSession(session auth.Session) bool {
	return session.AccessToken != "" && session.InstanceURL != "" && session.ID != ""
}