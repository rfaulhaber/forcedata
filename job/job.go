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
	"`": "BACKQUOTE",
	"^": "CARET",
	",": "COMMA",
	"|": "PIPE",
	";": "SEMICOLON",
	"\\t": "TAB",
}

type JobInfo struct {
	ApexProcessingTime      uint   `json:"apexProcessingTime"`
	APIActiveProcessingTime int    `json:"apiActiveProcessingTime"`
	APIVersion              float32 `json:"apiVersion"`
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
	ContentType string `json:"contentType"`
	Delim string `json:"columnDelimiter"`
}

type ServerError struct {
	ErrorCode string `json:"errorCode"`
	Message string `json:"message"`
}

func (s ServerError) Error() string {
	return s.Message
}

// returns true if valid delim, and returns name. otherwise returns false
func GetDelimName(delim string) (string, bool) {
	name, ok := delimMap[delim]
	return name, ok
}

type Job struct {
	Status chan JobInfo
	Done   chan bool

	session auth.Session
	config  JobConfig
	info JobInfo
}

func NewJob(config JobConfig, session auth.Session) *Job {
	return &Job{
		make(chan JobInfo),
		make(chan bool), session,
		config,
		JobInfo{},
	}
}

func (j *Job) Create() {
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
		log.Fatalln("error in response", err)
	}

	respBody, err := ioutil.ReadAll(response.Body)

	var info JobInfo

	err = json.Unmarshal(respBody, &info)

	log.Println("response", string(respBody))

	if err != nil {
		log.Println("err: ", err)
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			var resp []ServerError

			err := json.Unmarshal(respBody, &resp)

			if err != nil {
				log.Fatalln("some other horrible error has occurred", err)
			}

			log.Fatalln("message from server: ", resp[0].Message)
		} else {
			log.Fatalln("error in unmarshal", err)
		}
	}

	j.info = info
	log.Println("info from server", info)
}

func (j *Job) Upload(files ...string) {
	endpoint := j.jobURL()

	log.Println("attemping to hit endpoint: ", endpoint)

	readFiles := make([][]byte, len(files))

	for i, path := range files {
		log.Println("reading file: ", path)
		content, err := ioutil.ReadFile(path)

		if err != nil {
			// TODO handle
			log.Fatalln("couldn't read file: ", path, err)
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
			// TODO handle
			log.Fatalln("couldn't create request: ", err)
		}

		client := http.DefaultClient

		resp, err := client.Do(req)

		if err != nil {
			log.Fatalln("response err", err)
		}

		if resp.StatusCode != 201 {
			respBody, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				log.Println("resp body err", err)
			}

			log.Fatalln("server responded with ", resp.StatusCode, "with file: ", files[i], string(respBody))
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

func (j *Job) jobURL() string {
	return j.session.InstanceURL + "/" + j.info.ContentURL
}
