package job

import (
	"bytes"
	"encoding/json"
	"github.com/rfaulhaber/forcedata/auth"
	"io/ioutil"
	"log"
	"net/http"
	"io"
	"errors"
	"time"
)

// TODO enforce job state?

const (
	DefaultWatchTime = 5 * time.Second
	latestVersion = "43.0"
)

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

type JobError struct {
	Message string `json:"message"`
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
	Error chan JobError

	session auth.Session
	config  JobConfig
	info    JobInfo
}

func NewJob(config JobConfig, session auth.Session) *Job {
	return &Job{
		make(chan JobInfo),
		make(chan JobError),
		session,
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

	var info JobInfo

	err = readJSONBody(response.Body, &info)

	if err != nil {
		return err
	}

	j.info = info
	log.Println("info from server", info)

	return nil
}

// Uploads files to the job created in Create(). Sets the job to "Closed" when finished.
func (j *Job) Upload(file string) error {
	endpoint := j.jobURL()

	log.Println("attemping to hit endpoint: ", endpoint)

	log.Println("reading file: ", file)
	content, err := ioutil.ReadFile(file)

	if err != nil {
		log.Println("couldn't read file: ", file, err)
		return err
	}

	log.Println("attempting to upload file: ", file)
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

	// TODO return custom error
	if resp.StatusCode != 201 {
		respBody, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Println("resp body err", err)
			return err
		}

		log.Fatalln("server responded with ", resp.StatusCode, "with file: ", file, string(respBody))
	}

	return j.uploadComplete()
}

// Writes progress to the Status channel, and Done when job finishes.
func (j *Job) Watch(d time.Duration) {
	for {
		time.Sleep(d)
		info, err := j.GetInfo()

		// TODO do something more constructive here
		if err != nil {
			return
		}

		if info.State == "JobComplete" || info.State == "Failed" {
			close(j.Status)
			return
		}

		j.Status <- info
	}
}

func (j *Job) GetInfo() (JobInfo, error) {
	endpoint := j.session.InstanceURL + "/services/data/v" + latestVersion + "/jobs/ingest/" + j.info.ID

	log.Println("attemping to hit endpoint: ", endpoint)

	req, err := http.NewRequest("GET", endpoint, nil)

	if err != nil {
		return JobInfo{}, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.session.AccessToken)

	client := http.DefaultClient

	resp, err := client.Do(req)

	if err != nil {
		return JobInfo{}, err
	}

	info, err := getJobInfo(resp.Body)

	if err != nil {
		return JobInfo{}, err
	}

	log.Println("raw info", info)

	return info, err
}

// Sets job to "UploadComplete" state on the server.
func (j *Job) Complete() error {
	return j.uploadComplete()
}

// Aborts a job on the server
func (j *Job) Abort() error {
	return j.setState("Aborted")
}

// Deletes a job on the server
func (j *Job) Delete() error {
	endpoint := j.session.InstanceURL + "/services/data/v" + latestVersion + "/jobs/ingest/" + j.info.ID

	log.Println("attemping to hit endpoint: ", endpoint)

	req, err := http.NewRequest("DELETE", endpoint, nil)

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.session.AccessToken)

	client := http.DefaultClient

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		// TODO parse error?
		return errors.New("something went wrong with deleting job")
	}

	return nil
}

// TODO handle getting successful, failed, and unprocessed jobs
// TODO generalize callouts, create cleaner mechanism for it


func (j *Job) uploadComplete() error {
	return j.setState("UploadComplete")
}

func (j *Job) setState(state string) error {
	endpoint := j.session.InstanceURL + "/services/data/v" + latestVersion + "/jobs/ingest/" + j.info.ID

	log.Println("attemping to hit endpoint: ", endpoint)

	content, err := json.Marshal(struct {
		State string `json:"state"`
	}{
		state,
	})

	log.Println("raw request content", string(content))

	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", endpoint, bytes.NewReader(content))

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.session.AccessToken)

	client := http.DefaultClient

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	info, err := getJobInfo(resp.Body)

	if err != nil {
		return err
	}

	log.Println("raw info", info)

	return nil
}

func (j *Job) jobURL() string {
	return j.session.InstanceURL + "/" + j.info.ContentURL
}

func getJobInfo(b io.ReadCloser) (JobInfo, error) {
	var info JobInfo

	err := readJSONBody(b, &info)

	if err != nil {
		return JobInfo{}, err
	}

	return info, nil
}

func readJSONBody(b io.ReadCloser, v interface{}) error {
	body, err := ioutil.ReadAll(b)

	if err != nil {
		return err
	}

	log.Println("raw body", string(body))

	err = json.Unmarshal(body, &v)

	if err != nil {
		log.Println("info err", err)
	}

	return nil
}
