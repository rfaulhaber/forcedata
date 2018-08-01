package auth

import (
	"bytes"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestCredential_Encode(t *testing.T) {
	c := Credential{
		Username:     "test@example.com",
		Password:     "MyPassword123!!!",
		ClientID:     "SomeReallyLongClientId123456",
		ClientSecret: "somethingVerySecret",
		URL:          "https://login.salesforce.com",
	}

	result := c.Encode()

	u, err := url.Parse(result)

	if err != nil {
		t.Error("expected URL parsing to not throw an error", err)
	}

	q := u.Query()

	if q.Get("username") != c.Username {
		t.Error("Expected", c.Username, "\tReceived: ", q.Get("username"))
	}

	if q.Get("password") != c.Password {
		t.Error("Expected", c.Password, "\tReceived: ", q.Get("password"))
	}

	if q.Get("client_id") != c.ClientID {
		t.Error("Expected", c.ClientID, "\tReceived: ", q.Get("client_id"))
	}

	if q.Get("client_secret") != c.ClientSecret {
		t.Error("Expected", c.ClientSecret, "\tReceived: ", q.Get("client_secret"))
	}

	if q.Get("grant_type") != "password" {
		t.Error("Expected", "password", "\tReceived: ", q.Get("grant_type"))
	}
}

func TestAuthenticateFromFile(t *testing.T) {
	c := Credential{
		Username:     "test@example.com",
		Password:     "MyPassword123!!!",
		ClientID:     "SomeReallyLongClientId123456",
		ClientSecret: "somethingVerySecret",
		URL:          "",
	}

	s := Session{
		AccessToken: "token123",
		InstanceURL: "https://login.salesforce.com",
		ID:          "123",
		IssuedAt:    "12345",
		Signature:   "QWERTY",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(s)
		w.Write(data)
	}))

	c.URL = ts.URL

	b, err := json.Marshal(c)

	if err != nil {
		t.Error("could not marshal credential", err)
	}

	result, err := AuthenticateFromFile(b)

	if !cmp.Equal(result, s) {
		t.Error("expected: ", s, "received: ", result)
	}
}

func TestWriteSession(t *testing.T) {
	s := Session{
		AccessToken: "token123",
		InstanceURL: "https://login.salesforce.com",
		ID:          "123",
		IssuedAt:    "12345",
		Signature:   "QWERTY",
	}

	var buf bytes.Buffer

	WriteSession(s, &buf)

	var result Session

	err := json.Unmarshal(buf.Bytes(), &result)

	if err != nil {
		t.Error("unmarshal should not throw error", err)
	}

	if !cmp.Equal(s, result) {
		t.Error("expected: ", s, "received: ", result)
	}
}

func makeTempCredFile(c Credential) (*os.File, error) {
	currentDir, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	tmp, err := ioutil.TempFile(currentDir, "test-tmp")

	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(c)

	if err != nil {
		return nil, err
	}

	tmp.Write(data)

	return tmp, nil
}

func cleanTempFile(file *os.File) {
	file.Close()
	os.Remove(file.Name())
}
