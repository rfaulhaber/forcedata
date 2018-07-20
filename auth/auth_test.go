package auth

import (
	"testing"
	"bytes"
	"strings"
	"io/ioutil"
	"os"
	"log"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
)

func TestAuthenticatePromptSavesInfo(t *testing.T) {
	username := "testusername@example.com"
	password := "testPassword!verySecure!"
	loginURL := "http://localhost:3030"

	readerInput := strings.Join([]string{username, password, loginURL}, "\n")

	testReader := bytes.NewReader([]byte(readerInput))
	testWriter := bytes.NewBufferString("")

	config := AuthenticatePrompt(testReader, testWriter)

	if config.Username != username {
		t.Error("Expected", username, "got", config.Username)
	}

	if config.Password != password {
		t.Error("Expected", password, "got", config.Password)
	}

	if config.LoginURL != loginURL {
		t.Error("Expected", loginURL, "got", config.LoginURL)
	}

	resultOutput := strings.Split(testWriter.String(), ": ")
	expectedOutput := []string{"Username", "Password + security token", "Login URL (" + defaultLoginURL + "):"}

	for i := range resultOutput {
		result := resultOutput[i]
		expected := expectedOutput[i]
		if result != expected {
			t.Error("Expected", expected, "got", result)
		}
	}
}

func TestAuthenticatePromptDefaultsURL(t *testing.T) {
	username := "testusername@example.com"
	password := "testPassword!verySecure!"

	readerInput := strings.Join([]string{username, password, " "}, "\n")

	testReader := bytes.NewReader([]byte(readerInput))
	testWriter := bytes.NewBufferString("")

	config := AuthenticatePrompt(testReader, testWriter)

	if config.Username != username {
		t.Error("Expected", username, "got", config.Username)
	}

	if config.Password != password {
		t.Error("Expected", password, "got", config.Password)
	}

	if config.LoginURL != defaultLoginURL {
		t.Error("Expected", defaultLoginURL, "got", config.LoginURL)
	}
}

func TestAuthenticateFromJSONFile(t *testing.T) {
	testConfig := SFConfig{
		Username: "myUsername",
		Password: "!myPassword1!",
		LoginURL: "http://localhost:6060",
	}

	wd, err := os.Getwd()

	if err != nil {
		log.Fatalln("getwd error", err)
	}

	tmpFile, err := ioutil.TempFile(wd, "tmpconfig")

	if err != nil {
		log.Fatalln("tempfile error", err)
	}

	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	b, _ := json.Marshal(testConfig)

	tmpFile.Write(b)

	config, err := AuthenticateFromFile(tmpFile.Name())

	if !cmp.Equal(testConfig, config) {
		t.Error("Expected", testConfig, "got", config)
	}
}

//func TestAuthenticateFromYAMLFile(t *testing.T) {
//	testConfig := SFConfig{
//		Username: "myUsername",
//		Password: "!myPassword1!",
//		LoginURL: "http://localhost:6060",
//	}
//
//	wd, err := os.Getwd()
//
//	if err != nil {
//		log.Fatalln("getwd error", err)
//	}
//
//	tmpFile, err := ioutil.TempFile(wd, "tmpconfig")
//
//	if err != nil {
//		log.Fatalln("tempfile error", err)
//	}
//
//	defer func() {
//		tmpFile.Close()
//		os.Remove(tmpFile.Name())
//	}()
//
//	b, _ := yaml.Marshal(testConfig)
//
//	tmpFile.Write(b)
//
//	config, err := AuthenticateFromFile(tmpFile.Name())
//
//	if !cmp.Equal(testConfig, config) {
//		t.Error("Expected", testConfig, "got", config)
//	}
//}

func TestGetSessionInfoSuccess(t *testing.T) {

}

func TestGetSessionInfoFailure(t *testing.T) {

}
