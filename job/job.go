package job

import "github.com/rfaulhaber/force-data/auth"

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

var operationMap = map[string]Operation{
	"insert": Insert,
	"update": Update,
	"delete": Delete,
	"upsert": Upsert,
}

func (o Operation) String() string {
	return operations[o]
}

type Job struct {
	Status chan JobInfo
	Done   chan bool

	jobID   string
	session auth.SFSession
	config JobConfig
}

type JobInfo struct {
	ApexProcessingTime      uint   `json:"apexProcessingTime"`
	APIActiveProcessingTime int    `json:"apiActiveProcessingTime"`
	APIVersion              string `json:"apiVersion"`
	ColumnDelimiter         string `json:"columnDelimiter"`
	ConcurrencyMode         string `json:"concurrencyMode"`
	ContentType             string `json:"contentType"`
	ContentURL              string `json:"contentUrl"`
	CreatedByID             string `json:"createdById"`
	CreatedDate             string `json:"createdDate"`
	ExternalIdFieldName     string `json:"externalIdFieldName"`
	ID                      string `json:"id"`
	JobType                 string `json:"jobType"`
	LineEnding              string `json:"lineEnding"`
	RecordsFailed           uint   `json:"numberRecordsFailed"`
	RecordsProcessed        uint   `json:"numberRecordsProcessed"`
	Retries                 uint   `json:"retries"`
	Object                  string `json:"object"`
	Operation               string `json:"operation"`
	State                   string `json:"state"`
	SystemModstamp          string `json:"SystemModstamp"`
	TotalProcessingTime     uint   `json:"totalProcessingTime"`
}

type JobConfig struct {
	Object    string
	Operation Operation
}

func NewJob(config JobConfig, session auth.SFSession) *Job {
	return &Job{
		make(chan JobInfo),
		make(chan bool),
		"",
		session,
		config,
	}
}

func (j *Job) Create() {

}

func (j *Job) Run() {

}

func (j *Job) Close() {

}

func (j *Job) Abort() {

}

func (j *Job) Delete() {
}

// TODO handle getting successful, failed, and unprocessed jobs

func (j *Job) setJobID(id string) {
	j.jobID = id
}
