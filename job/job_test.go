package job

import (
	"github.com/rfaulhaber/forcedata/auth"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const testInstanceURL = "http://localhost:3030"

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
	testSession := makeSession()

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
}

func TestJob_Upload(t *testing.T) {

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

func makeSession() auth.Session {
	return auth.Session{
		AccessToken: "token123",
		InstanceURL: testInstanceURL,
		ID:          "ID123",
		IssuedAt:    time.Now().String(),
		Signature:   "123SIG321",
	}
}
