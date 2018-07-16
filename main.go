package main

import (
	"flag"
	"fmt"
	// use when drawing progress bars!
	//"gopkg.in/cheggaaa/pb.v1"
	"encoding/json"
	"io/ioutil"
)

type SFConfig struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	LoginURL     string `json:"loginUrl"`
	InstanceName string `json:"instanceId"`
}

type SFAuth struct {
	SessionID string
	ServerURL string
}

var actionMap = map[string]struct{}{
	"insert": {},
	"update": {},
	"delete": {},
	"upsert": {},
}

type Command struct {
	Action string
	Files  []string
}

type Process struct {
	Commands []Command

	// if false, serial
	IsParallel bool
}

func (p Process) OptimizeCommands() []Command {
	m := p.CommandMap()
	cmds := make([]Command, 0)

	for action, files := range m {
		cmds = append(cmds, Command{action, files})
	}

	return cmds
}

// when not parallel, this returns an optimized mapping of actions to files. for instance, if the user enters the
// command:
// data insert file1 file2 update file3 insert file4
// this command will create a mapping where file1, file2, and file4 are all under "insert"
func (p Process) CommandMap() map[string][]string {
	m := make(map[string][]string)

	for _, cmd := range p.Commands {
		for _, file := range cmd.Files {
			m[cmd.Action] = append(m[cmd.Action], file)
		}
	}

	return m
}

type Job struct {
	C      chan JobInfo
	Done   chan bool

	auth SFAuth
}

type JobRequest struct {
	Operation   string `json:"operation"`
	Object      string `json:"object"`
	ContentType string `json:"contentType"`
}

type JobInfo struct {
	ApexProcessingTime      int     `json:"apexProcessingTime"`
	APIActiveProcessingTime int     `json:"apiActiveProcessingTime"`
	APIVersion              float32 `json:"apiVersion"`
	ConcurrencyMode         string  `json:"concurrencyMode"`
	CreatedByID             string  `json:"createdById"`
	CreatedDate             string  `json:"createdDate"`
	ID                      string  `json:"id"`
	BatchesCompleted        int     `json:"numberBatchesCompleted"`
	BatchesFailed           int     `json:"numberBatchesFailed"`
	BatchesInProgress       int     `json:"numberBatchesInProgress"`
	BatchesQueued           int     `json:"numberBatchesQueued"`
	BatchesTotal            int     `json:"numberBatchesTotal"`
	RecordsFailed           int     `json:"numberRecordsFailed"`
	RecordsProcessed        int     `json:"numberRecordsProcessed"`
	Retries                 int     `json:"numberRetries"`
	Object                  string  `json:"object"`
	Operation               string  `json:"operation"`
	State                   string  `json:"state"`
	SystemModstamp          string  `json:"SystemModstamp"`
	TotalProcessingTime     int     `json:"totalProcessingTime"`
}

func (j *Job) Create() {

}

func (j *Job) Run() {
}

func NewJob(auth SFAuth) *Job {
	return &Job{
		make(chan JobInfo),
		make(chan bool),
		auth,
	}
}

// example usage
// data --config config.json insert test1.csv test2.json update test3.xml

func main() {
	/*
	 * other flags to implement:
	 * - config flag
	 * - out flag (for writing / piping output)
	 * - flag to indicate program should convert non-CSV to CSV
	 */
	syncFlag := flag.Bool("s", false, "If true, does data load in sync mode")

	flag.Parse()

	cmds := parseArgs(flag.Args())

	process := Process{cmds, *syncFlag}

	fmt.Println(process)
	fmt.Println(process.CommandMap())
}

func parseArgs(args []string) []Command {
	// TODO handle invalid syntax
	cmds := make([]Command, 0)

	for len(args) > 0 {
		cmd := Command{}

		cmd.Action = args[0]

		i := 1

		for i < len(args) && !isCommand(args[i]) {
			cmd.Files = append(cmd.Files, args[i])
			i++
		}

		args = args[i:]
		cmds = append(cmds, cmd)
	}

	return cmds
}

func loadConfigFile(path string) (config SFConfig, err error) {
	file, err := ioutil.ReadFile(path)

	if err != nil {
		return SFConfig{}, err
	}

	err = json.Unmarshal(file, &config)

	return
}

func isCommand(str string) bool {
	_, ok := actionMap[str]
	return ok
}
