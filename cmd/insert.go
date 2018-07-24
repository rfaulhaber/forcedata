package cmd

import (
	"fmt"
	"github.com/rfaulhaber/force-data/auth"
	"github.com/rfaulhaber/force-data/job"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var (
	objFlag   string
	delimFlag string
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
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(insertCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// insertCmd.PersistentFlags().String("foo", "", "A help for foo")
	insertCmd.Flags().StringVar(&delimFlag, "delim", ",", "Delimiter used in all specified fiels (defaults to \",\")")
	insertCmd.Flags().StringVar(&objFlag, "object", "", "Object being inserted.")
	insertCmd.MarkFlagRequired("object")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// insertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runInsert(cmd *cobra.Command, args []string) {
	fmt.Println("attempting to insert: ", args)

	session := auth.Session{}
	err := viper.Unmarshal(&session)

	if err != nil {
		log.Fatalln("viper could not unmarshal", err)
	}

	log.Println("session url", session.InstanceURL)

	// TODO do better!
	delimName, _ := job.GetDelimName(delimFlag)

	config := job.JobConfig{
		Object:    objFlag,
		Operation: "insert",
		Delim:     delimName,
		// TODO dynamically populate
		ContentType: "CSV",
	}

	j := job.NewJob(config, session)

	log.Println("creating job...")
	j.Create()

	log.Println("uploading files...")
	j.Upload(args...)
}
