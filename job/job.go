package job

import (
	"bytes"
	"encoding/json"
	"github.com/rfaulhaber/force-data/auth"
	"io/ioutil"
	"log"
	"net/http"
)

const latestVersion = "43.0"

type Job struct {
	Status chan JobInfo
	Done   chan bool

	jobID   string
	jobURL  string
	session auth.Session
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

func NewJob(config JobConfig, session auth.Session) *Job {
	return &Job{
		make(chan JobInfo),
		make(chan bool),
		"",
		"",
		session,
		config,
	}
}

func (j *Job) Create() {
	endpoint := j.session.InstanceURL + "/services/data/v" + latestVersion + "/jobs/ingest"

	reqBody, _ := json.Marshal(j.config)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authentication", "Bearer "+j.session.AccessToken)

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

	j.setJobID(info.ID)
	j.setJobID(info.ContentURL)
}

func (j *Job) Upload(files ...string) {
	// TODO make async?
	endpoint := j.jobURL + "/services/data/v" + latestVersion + "/jobs/ingest/" + j.jobID + "/batches"

	readFiles := make([][]byte, len(files))

	for _, path := range files {
		content, err := ioutil.ReadFile(path)

		if err != nil {
			// TODO handle
			log.Fatalln("couldn't read file: ", path, err)
		}

		readFiles = append(readFiles, content)
	}

	for i := range readFiles {
		content := readFiles[i]
		req, err := http.NewRequest("POST", endpoint, bytes.NewReader(content))

		if err != nil {
			// TODO handle
			log.Fatalln("couldn't create request: ", err)
		}

		client := http.DefaultClient

		resp, err := client.Do(req)

		if err != nil {
			log.Fatalln("response err", err)
		}

		if resp.StatusCode != 201 {
			log.Fatalln("server responded with ", resp.StatusCode, "with file: ")
		}
	}
}

func (j *Job) Watch() {

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

func (j *Job) setJobURL(url string) {
	j.jobURL = url
}

//type endpointType int
//
//const (
//	createJob = iota
//	uploadJob
//	closeJob
//	deleteJob
//	jobInfo
//	success
//	failure
//	unprocessed
//)
