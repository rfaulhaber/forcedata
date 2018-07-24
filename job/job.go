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

var delimMap = map[string]string{
	"`":   "BACKQUOTE",
	"^":   "CARET",
	",":   "COMMA",
	"|":   "PIPE",
	";":   "SEMICOLON",
	"\\t": "TAB",
}

// Returns true if valid delim, and returns name. Otherwise returns false.
func GetDelimName(delim string) (string, bool) {
	name, ok := delimMap[delim]
	return name, ok
}

type JobInfo struct {
	ApexProcessingTime      uint    `json:"apexProcessingTime"`
	APIActiveProcessingTime int     `json:"apiActiveProcessingTime"`
	APIVersion              float32 `json:"apiVersion"`
	ColumnDelimiter         string  `json:"columnDelimiter"`
	ConcurrencyMode         string  `json:"concurrencyMode"`
	ContentType             string  `json:"contentType"`
	ContentURL              string  `json:"contentUrl"`
	CreatedByID             string  `json:"createdById"`
	CreatedDate             string  `json:"createdDate"`
	ExternalIdFieldName     string  `json:"externalIdFieldName"`
	ID                      string  `json:"id"`
	JobType                 string  `json:"jobType"`
	LineEnding              string  `json:"lineEnding"`
	RecordsFailed           uint    `json:"numberRecordsFailed"`
	RecordsProcessed        uint    `json:"numberRecordsProcessed"`
	Retries                 uint    `json:"retries"`
	Object                  string  `json:"object"`
	Operation               string  `json:"operation"`
	State                   string  `json:"state"`
	SystemModstamp          string  `json:"SystemModstamp"`
	TotalProcessingTime     uint    `json:"totalProcessingTime"`
}

type JobConfig struct {
	Object      string `json:"object"`
	Operation   string `json:"operation"`
	ContentType string `json:"contentType"`
	Delim       string `json:"columnDelimiter"`
}

type ServerError struct {
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
}

func (s ServerError) Error() string {
	return s.Message
}

type Job struct {
	Status chan JobInfo
	Done   chan bool

	session auth.Session
	config  JobConfig
	info    JobInfo
}

func NewJob(config JobConfig, session auth.Session) *Job {
	return &Job{
		make(chan JobInfo),
		make(chan bool), session,
		config,
		JobInfo{},
	}
}

// Creates a job on the server with the specified config for the job
func (j *Job) Create() error {
	endpoint := j.session.InstanceURL + "/services/data/v" + latestVersion + "/jobs/ingest"

	log.Println("attempting to hit endpoint: ", endpoint)

	reqBody, _ := json.Marshal(j.config)

	log.Println("job", string(reqBody))

	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.session.AccessToken)

	client := http.DefaultClient

	response, err := client.Do(req)

	if err != nil {
		log.Println("error in response", err)
		return err
	}

	respBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Println("error in reading response", err)
		return err
	}

	var info JobInfo

	err = json.Unmarshal(respBody, &info)

	log.Println("response", string(respBody))

	if err != nil {
		log.Println("error in unmarshalling", err)
		return err
	}

	j.info = info
	log.Println("info from server", info)

	return nil
}

// Uploads files to the job created in Create()
func (j *Job) Upload(files ...string) error {
	endpoint := j.jobURL()

	log.Println("attemping to hit endpoint: ", endpoint)

	readFiles := make([][]byte, len(files))

	for i, path := range files {
		log.Println("reading file: ", path)
		content, err := ioutil.ReadFile(path)

		if err != nil {
			log.Println("couldn't read file: ", path, err)
			return err
		}

		readFiles[i] = content
	}

	for i := range readFiles {
		content := readFiles[i]
		log.Println("attempting to upload file: ", i)
		req, err := http.NewRequest("PUT", endpoint, bytes.NewReader(content))
		req.Header.Add("Content-Type", "text/csv")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", "Bearer "+j.session.AccessToken)

		if err != nil {
			log.Println("couldn't create request: ", err)
			return err
		}

		client := http.DefaultClient

		resp, err := client.Do(req)

		if err != nil {
			log.Println("response err", err)
			return err
		}

		if resp.StatusCode != 201 {
			respBody, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				log.Println("resp body err", err)
				return err
			}

			// TODO return custom error
			log.Fatalln("server responded with ", resp.StatusCode, "with file: ", files[i], string(respBody))
		}
	}

	return nil
}

// Writes progress to the Status channel, and Done when finished
func (j *Job) Watch() {
	panic("not implemented")
}

// Named "CloseJob" to avoid confusion that it implements any io type. Closes a job on the server.
func (j *Job) CloseJob() error {
	panic("not implemented")
}

// Aborts a job on the server
func (j *Job) Abort() error {
	panic("not implemented")
}

// Deletes a job on the server
func (j *Job) Delete() error {
	panic("not implemented")
}

// TODO handle getting successful, failed, and unprocessed jobs

func (j *Job) jobURL() string {
	return j.session.InstanceURL + "/" + j.info.ContentURL
}
