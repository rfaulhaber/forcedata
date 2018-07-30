package job

import (
	"encoding/json"
	"github.com/rfaulhaber/forcedata/auth"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetDelimName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"`", "BACKQUOTE"},
		{"^", "CARET"},
		{",", "COMMA"},
		{"|", "PIPE"},
		{";", "SEMICOLON"},
		{"\\t", "TAB"},
	}

	for _, tc := range testCases {
		output, ok := GetDelimName(tc.input)
		assert.Equal(t, tc.expected, output)
		assert.True(t, ok)
	}
}

func TestNewJob(t *testing.T) {
	testSession := makeSession("http://some.url")

	testConfig := JobConfig{
		Object:      "Contact",
		Operation:   "insert",
		ContentType: "CSV",
		Delim:       "COMMA",
	}

	expected := &Job{
		make(chan JobInfo),
		make(chan error),
		testSession,
		testConfig,
		JobInfo{},
	}

	result := NewJob(testConfig, testSession)

	assert.Equal(t, expected.session, result.session)
	assert.Equal(t, expected.config, result.config)
}

func TestJob_Create(t *testing.T) {
	testCases := []struct {
		handler http.HandlerFunc
	}{
		{
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := ioutil.ReadAll(r.Body)

				if err != nil {
					panic(err)
				}

				resp, _ := json.Marshal(JobInfo{
					ID: "123ID321",
				})

				w.WriteHeader(201)
				w.Write(resp)
			}),
		},
		{
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := ioutil.ReadAll(r.Body)

				if err != nil {
					panic(err)
				}

				resp, _ := json.Marshal([]JobError{{Message: "test message", ErrorCode: "TEST_CASE"}})

				w.WriteHeader(401)
				w.Write(resp)
			}),
		},
	}

	for i, tc := range testCases {
		server := httptest.NewServer(tc.handler)

		job := NewJob(JobConfig{"Contact", "insert", "CSV", "COMMA"}, makeSession(server.URL))

		err := job.Create()

		if i == 1 {
			assert.Error(t, err)
			assert.IsType(t, err, JobError{})
		} else {
			assert.NoError(t, err)
			assert.True(t, job.info.ID == "123ID321")
		}
	}
}

func TestJob_Upload(t *testing.T) {
	var actualBody []byte
	var actualEndpoint string
	var actualMethod string

	var actualCloseBody []byte
	var actualCloseURL string

	testBody := []byte("FirstName,LastName\nPerson,One")

	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			b, err := ioutil.ReadAll(r.Body)

			actualEndpoint = r.URL.String()
			actualMethod = r.Method
			actualBody = b

			if err != nil {
				panic(err)
			}

			w.WriteHeader(201)
		} else {
			b, err := ioutil.ReadAll(r.Body)
			actualCloseURL = r.URL.String()

			if err != nil {
				panic(err)
			}

			actualCloseBody = b

			w.WriteHeader(201)

			resp, _ := json.Marshal(JobInfo{})

			w.Write(resp)
		}
	}))

	job := NewJob(JobConfig{"Contact", "insert", "CSV", "COMMA"}, makeSession(server.URL))
	job.info.ID = "123ID321"
	job.info.ContentURL = "services/data/v43.0/jobs/batches"

	err := job.Upload(testBody)

	assert.NoError(t, err)
	assert.Equal(t, job.batchURL(), server.URL+actualEndpoint)
	assert.Equal(t, 2, callCount)
	assert.Equal(t, testBody, actualBody)
	assert.Equal(t, []byte(`{"state":"UploadComplete"}`), actualCloseBody)
	assert.Equal(t, "PUT", actualMethod)
}

func TestJob_UploadError(t *testing.T) {
	var actualBody []byte
	var actualEndpoint string

	testBody := []byte("FirstName,LastName\nPerson,One")

	callCount := 0

	testErrorCode := "ERROR_CODE"
	testErrorMessage := "test message"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			b, err := ioutil.ReadAll(r.Body)

			actualEndpoint = r.URL.String()

			actualBody = b

			if err != nil {
				panic(err)
			}

			w.WriteHeader(401)

			resp, _ := json.Marshal([]JobError{{ErrorCode: testErrorCode, Message: testErrorMessage}})

			w.Write(resp)
		}
	}))

	job := NewJob(JobConfig{"Contact", "insert", "CSV", "COMMA"}, makeSession(server.URL))
	job.info.ID = "123ID321"
	job.info.ContentURL = "services/data/v43.0/jobs/batches"

	err := job.Upload(testBody)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "upload: server responded with 401, error: code: ERROR_CODE, message: test message")
}

func TestJob_UploadCloseError(t *testing.T) {

}

func TestJob_Abort(t *testing.T) {

}

func TestJob_Complete(t *testing.T) {

}

func TestJob_Delete(t *testing.T) {

}

func TestJob_SetJobInfo(t *testing.T) {

}

func TestJob_GetInfo(t *testing.T) {

}

func TestJob_GetSuccess(t *testing.T) {

}

func TestJob_GetFailure(t *testing.T) {

}

func TestJob_GetUnprocessed(t *testing.T) {

}

func makeSession(instanceURL string) auth.Session {
	return auth.Session{
		AccessToken: "token123",
		InstanceURL: instanceURL,
		ID:          "ID123",
		IssuedAt:    time.Now().String(),
		Signature:   "123SIG321",
	}
}
