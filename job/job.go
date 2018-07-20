package job

import (
	"github.com/rfaulhaber/force-data/auth"
	"net/http"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"log"
)

const latestVersion = "43.0"

type Job struct {
	Status chan JobInfo
	Done   chan bool

	jobID   string
	session auth.SFSession
	config  JobConfig
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
	Object    string `json:"object"`
	Operation string `json:"operation"`
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
	endpoint := j.session.ServerURL + "/services/data/v" + latestVersion + "/jobs/ingest"

	reqBody, _ := json.Marshal(j.config)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")

	client := http.DefaultClient

	response, err := client.Do(req)

	if err != nil {
		log.Fatalln("error in response", err)
	}

	respBody, err := ioutil.ReadAll(response.Body)

	var info JobInfo

	err = json.Unmarshal(respBody, &info)

	if err != nil {
		log.Fatalln("error in unmarshal", err)
	}

	j.jobID = info.ID
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

type endpointType int

const (
	createJob = iota
	uploadJob
	closeJob
	deleteJob
	jobInfo
	success
	failure
	unprocessed
)
