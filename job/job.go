package job

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/rfaulhaber/forcedata/auth"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	DefaultWatchTime = 5 * time.Second
	latestVersion    = "43.0"
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
	Message   string   `json:"message"`
	ErrorCode string   `json:"errorCode"`
	Fields    []string `json:"fields"`
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
	Error  chan error

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

// Creates a job on the server with the specified config for the job.
func (j *Job) Create() error {
	endpoint := j.ingestURL()

	reqBody, _ := json.Marshal(j.config)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.session.AccessToken)

	client := http.DefaultClient

	response, err := client.Do(req)

	if err != nil {
		return errors.Wrap(err, "creating job returned error")
	}

	var responseBody bytes.Buffer

	bodyReader := io.TeeReader(response.Body, &responseBody)

	info, err := getJobInfo(bodyReader)

	if err != nil {
		if jobErr, ok := checkJobError(err, &responseBody); ok {
			return jobErr
		} else {
			return errors.Wrap(err, "server returned error creating job")
		}
	}

	j.info = info
	return nil
}

// Uploads files to the job created. Must call Create() first. Sets the job to "Closed" when finished. Takes in CSV
// content as a Reader.
func (j *Job) Upload(content []byte) error {
	endpoint := j.batchURL()

	req, err := http.NewRequest("PUT", endpoint, bytes.NewReader(content))

	if err != nil {
		return errors.Wrap(err, "could not generate upload request")
	}

	req.Header.Add("Content-Type", "text/csv")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.session.AccessToken)

	client := http.DefaultClient

	resp, err := client.Do(req)

	if err != nil {
		return errors.Wrap(err, "upload response error")
	}

	if resp.StatusCode != 201 {
		jobError, err := getJobError(resp.Body)

		if err != nil {
			return errors.Wrap(err, "could not read upload response body")
		}

		return errors.Errorf("upload: server responded with %d, error: code: %s, message: %s", resp.StatusCode, jobError[0].ErrorCode, jobError[0].Message)
	}

	return j.uploadComplete()
}

// Continuously makes request to get job info from the server.  Writes progress to the Status channel. If server reports
// state as "JobComplete" or "Failed", Status channel closes. Writes any error encountered to Error.
func (j *Job) Watch(d time.Duration) {
	for {
		time.Sleep(d)
		info, err := j.GetInfo()

		if err != nil {
			j.Error <- err
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
	endpoint := j.ingestURLWithID()

	resp, err := j.jsonRequest("GET", endpoint, nil)

	if err != nil {
		return JobInfo{}, err
	}

	info, err := getJobInfo(resp.Body)

	if err != nil {
		return JobInfo{}, err
	}

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

	resp, err := j.jsonRequest("DELETE", endpoint, nil)

	if err != nil {
		return errors.Wrap(err, "delete request failed")
	}

	if resp.StatusCode != 204 {
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

func (j *Job) SetInfo(info JobInfo) {
	j.info = info
}

func (j *Job) ID() string {
	return j.info.ID
}

func (j *Job) uploadComplete() error {
	return j.setState("UploadComplete")
}

func (j *Job) setState(state string) error {
	endpoint := j.ingestURLWithID()

	content, err := json.Marshal(struct {
		State string `json:"state"`
	}{
		state,
	})

	resp, err := j.jsonRequest("PATCH", endpoint, content)

	if err != nil {
		return err
	}

	var responseBody bytes.Buffer

	bodyReader := io.TeeReader(resp.Body, &responseBody)

	_, err = getJobInfo(bodyReader)

	if err != nil {
		if jobErr, ok := checkJobError(err, &responseBody); ok {
			return jobErr
		} else {
			return errors.Wrap(err, "response error from setting state to "+state)
		}
	}

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

func (j *Job) jsonRequest(method string, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))

	if err != nil {
		return nil, errors.Wrap(err, "request generation failed")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.session.AccessToken)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, method+" response returned error")
	}

	return resp, nil
}

func getJobInfo(b io.Reader) (JobInfo, error) {
	var info JobInfo

	err := readJSONBody(b, &info)

	if err != nil {
		return info, errors.Wrap(err, "tried parsing job info")
	}

	return info, nil
}

func checkJobError(err error, reader io.Reader) (JobError, bool) {
	if _, ok := errors.Cause(err).(*json.UnmarshalTypeError); ok {
		jobErr, err := getJobError(reader)

		if err != nil {
			return JobError{}, false
		}

		return jobErr[0], true
	} else {
		return JobError{}, false
	}
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
