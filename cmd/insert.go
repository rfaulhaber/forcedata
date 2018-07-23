package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/rfaulhaber/force-data/job"
	"github.com/rfaulhaber/force-data/auth"
	"log"
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

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// insertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runInsert(cmd *cobra.Command, args []string) {
	fmt.Println("attempting to insert: ", args)
	fmt.Println("server", viper.GetString("instance_url"))

	session := auth.Session{}
	err := viper.Unmarshal(&session)

	log.Println("session", session)

	if err != nil {
		log.Fatalln("viper could not unmarshal", err)
	}

	log.Println("session url", session.InstanceURL)

	config := job.JobConfig{
		Object: "Contact",
		Operation: "insert",
	}

	j := job.NewJob(config, session)

	j.Create()
}
