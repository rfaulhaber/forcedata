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
	Args: cobra.MaximumNArgs(1),
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
	fmt.Println("attempting to insert: ", args)

	// TODO validate config
	session := auth.Session{}
	err := viper.Unmarshal(&session)

	if err != nil {
		log.Fatalln("viper could not unmarshal config", err)
	}

	log.Println("session url", session.InstanceURL)

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
		// TODO handle error, print message
		log.Fatalln("create error", err)
	}

	log.Println("uploading files...")
	err = j.Upload(args[0])

	if err != nil {
		log.Fatalln("upload error", err)
	}

	if watchFlag {
		go j.Watch(watchTimeFlag)

		for {
			if status, ok := <- j.Status; ok {
				PrintStatus(status, os.Stdout)
			} else {
				return
			}
		}
	}
}

func PrintStatus(status job.JobInfo, out io.Writer) {
	io.WriteString(out, "")
	fmt.Fprintf(out, "Records processed: %d\tRecords failed: %d", status.RecordsProcessed, status.RecordsFailed)
}
