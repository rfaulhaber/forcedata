package main

import (
		// use when drawing progress bars!
	//"gopkg.in/cheggaaa/pb.v1"
	"encoding/json"
	"github.com/ogier/pflag"
	"io/ioutil"
	)

type SFConfig struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	LoginURL     string `json:"loginUrl"`
	InstanceName string `json:"instanceId"`
}

type SFSession struct {
	SessionID string
	ServerURL string
}

type Operation int

const (
	Insert = iota
	Update
	Delete
	Upsert
)

var operations = []string{
	"insert",
	"update",
	"delete",
	"upsert",
}

var operationMap = map[string]struct{}{
	"insert": {},
	"update": {},
	"delete": {},
	"upsert": {},
}

func (o Operation) String() string {
	return operations[o]
}

type Batch struct {
	Action string
	Files  []string
}

type Process struct {
	Batches []Batch

	// if false, serial
	IsParallel bool
}

func (p Process) OptimizeBatches() []Batch {
	m := p.CommandMap()
	cmds := make([]Batch, 0)

	for action, files := range m {
		cmds = append(cmds, Batch{action, files})
	}

	return cmds
}

// when not parallel, this returns an optimized mapping of actions to files. for instance, if the user enters the
// command:
// data insert file1 file2 update file3 insert file4
// this command will create a mapping where file1, file2, and file4 are all under "insert"
func (p Process) CommandMap() map[string][]string {
	m := make(map[string][]string)

	for _, b := range p.Batches {
		for _, file := range b.Files {
			m[b.Action] = append(m[b.Action], file)
		}
	}

	return m
}

type ConcurrencyMode int

const (
	Parallel = iota
	Serial
)

var concurrencyModes = []string{
	"Parallel",
	"Serial",
}

func (c ConcurrencyMode) String() string {
	return concurrencyModes[c]
}

func (c ConcurrencyMode) MarshalJSON() ([]byte, error) {
	str := concurrencyModes[c]
	return []byte(`"` + str + `"`), nil
}

type Job struct {
	C    chan JobInfo
	Done chan bool

	session         SFSession
	batches         []Batch
	concurrencyMode ConcurrencyMode
}

type JobRequest struct {
	Operation       string `json:"operation"`
	Object          string `json:"object"`
	ContentType     string `json:"contentType"`
	ConcurrencyMode string `json:"concurrencyMode"`
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

// sends request to server to create job
func (j *Job) Create() {
}

// adds a batch to the object, to be sent to the server later
func (j *Job) AddBatch(b Batch) {
	j.batches = append(j.batches, b)
}

func (j *Job) SetAuth(session SFSession) {
	j.session = session
}

// Job defaults to "Parallel", set this to make it "Serial" instead
func (j *Job) SetSerialMode() {
	j.concurrencyMode = Serial
}

// sends all batches to the server
func (j *Job) Run() {
}

func NewJob(object string, operation string) *Job {
	return &Job{
		make(chan JobInfo),
		make(chan bool),
		SFSession{},
		make([]Batch, 0),
		Parallel,
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
	//syncFlag := pflag.BoolP("sync", "s", false, "If true, does data load in sync mode")

	pflag.Parse()

	//cmds := parseArgs(pflag.Args())
	//
	//process := Process{cmds, *syncFlag}
	//
	//fmt.Println(process)
	//fmt.Println(process.CommandMap())

	// TODO add waitgroup while everything is processing
}

func parseArgs(args []string) []Batch {
	// TODO handle invalid syntax
	cmds := make([]Batch, 0)

	for len(args) > 0 {
		cmd := Batch{}

		cmd.Action = args[0]

		i := 1

		for i < len(args) && !isOperation(args[i]) {
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

func isOperation (str string) bool {
	_, ok := operationMap[str]
	return ok
}
