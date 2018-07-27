package job

import (
	"bytes"
	"encoding/json"
	"github.com/rfaulhaber/forcedata/auth"
	"io/ioutil"
	"log"
	"net/http"
	"io"
	"time"
	"github.com/pkg/errors"
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
	ErrorCode string `json:"errorCode"`
	Fields []string `json:"fields"`
}

func (e JobError) Error() string {
	return e.Message
}

type JobConfig struct {
	Object      string `json:"object"`
	Operation   string `json:"operation"`
	ContentType string `json:"contentType"`
	Delim       string `json:"columnDelimiter"`
}

type Job struct {
	Status chan JobInfo
	Error chan error

	session auth.Session
	config  JobConfig
	info    JobInfo
}

func NewJob(config JobConfig, session auth.Session) *Job {
	return &Job{
		make(chan JobInfo),
		make(chan error),
		session,
		config,
		JobInfo{},
	}
}

// Creates a job on the server with the specified config for the job
func (j *Job) Create() error {
	endpoint := j.ingestURL()

	log.Println("attempting to hit endpoint: ", endpoint)

	reqBody, _ := json.Marshal(j.config)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.session.AccessToken)

	client := http.DefaultClient

	response, err := client.Do(req)

	if err != nil {
		return errors.Wrap(err, "create: error in response")
	}

	var responseBody bytes.Buffer

	bodyReader := io.TeeReader(response.Body, &responseBody)

	info, err := getJobInfo(bodyReader)

	if err != nil {
		if _, ok := errors.Cause(err).(*json.UnmarshalTypeError); ok {
			jobErr, err := getJobError(&responseBody)

			if err != nil {
				return errors.Wrap(err, "attempted to parse job error")
			}

			return jobErr[0]
		} else {
			return errors.Wrap(err, "create: get job info")
		}
	}

	j.info = info
	log.Println("info from server", info)

	return nil
}

// Uploads files to the job created. Must call Create() first. Sets the job to "Closed" when finished. Takes in CSV
// content as a Reader.
func (j *Job) Upload(content []byte) error {
	endpoint := j.batchURL()

	// TODO should I validate content?

	log.Println("attempting to hit endpoint: ", endpoint)

	req, err := http.NewRequest("PUT", endpoint, bytes.NewReader(content))

	if err != nil {
		return errors.Wrap(err, "upload: could not create request")
	}

	req.Header.Add("Content-Type", "text/csv")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.session.AccessToken)

	client := http.DefaultClient

	log.Println("uploading content")
	resp, err := client.Do(req)

	if err != nil {
		return errors.Wrap(err, "upload: response error")
	}

	// TODO return custom error
	if resp.StatusCode != 201 {
		log.Println("upload: server responded with non-OK status code")
		jobError, err := getJobError(resp.Body)

		if err != nil {
			return errors.Wrap(err, "could not read response body")
		}

		return errors.Errorf("upload: server responded with %d, error: code: %s, message: %s", resp.StatusCode, jobError[0].ErrorCode, jobError[0].Message)
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

	log.Println("attempting to hit endpoint: ", endpoint)

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
	endpoint := j.ingestURLWithID()

	log.Println("attempting to hit endpoint: ", endpoint)

	req, err := http.NewRequest("DELETE", endpoint, nil)

	if err != nil {
		return errors.Wrap(err, "delete request failed")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.session.AccessToken)

	client := http.DefaultClient

	resp, err := client.Do(req)

	if err != nil {
		return errors.Wrap(err, "delete callout failed")
	}

	if resp.StatusCode != 204 {
		// TODO parse error?
		return errors.New("something went wrong with deleting job")
	}

	return nil
}

func (j *Job) GetSuccess() ([]byte, error) {
	panic("not implemented")
}

func (j *Job) GetFailure() ([]byte, error) {
	panic("not implemented")
}

func (j *Job) GetUnprocessed() ([]byte, error) {
	panic("not implemented")
}

func (j *Job) SetJobInfo(info JobInfo) {
	j.info = info
}

// TODO generalize callouts, create cleaner mechanism for it


func (j *Job) uploadComplete() error {
	return j.setState("UploadComplete")
}

func (j *Job) setState(state string) error {
	endpoint := j.ingestURLWithID()

	log.Println("attempting to hit endpoint: ", endpoint)

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

func (j *Job) batchURL() string {
	return j.session.InstanceURL + "/" + j.info.ContentURL
}

func (j *Job) ingestURL() string {
	return j.session.InstanceURL + "/services/data/v" + latestVersion + "/jobs/ingest/"
}

func (j *Job) ingestURLWithID() string {
	return j.ingestURL() + j.info.ID
}

func getJobInfo(b io.Reader) (JobInfo, error) {
	var info JobInfo

	err := readJSONBody(b, &info)

	if err != nil {
		return info, errors.Wrap(err, "tried parsing job info")
	}

	return info, nil
}

func getJobError(b io.Reader) ([]JobError, error) {
	var e []JobError

	err := readJSONBody(b, &e)

	if err != nil {
		return e, errors.Wrap(err, "tried parsing job error")
	}

	return e, nil
}

func readJSONBody(b io.Reader, v interface{}) error {
	body, err := ioutil.ReadAll(b)

	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &v)

	if err != nil {
		return errors.Wrap(err, "json parse error")
	}

	return nil
}
