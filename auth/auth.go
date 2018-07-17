package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

const defaultLoginURL = "https://login.salesforce.com"

type SFConfig struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password"`
	LoginURL string `json:"loginUrl"`
}

type SFSession struct {
	ServerURL string `xml:"serverUrl" json:"serverUrl"`
	SessionID string `xml:"sessionId" json:"sessionId"`
}

type client interface {
	Do(request *http.Request) (*http.Response, error)
}

type httpClient struct {
	client *http.Client
}

func (c *httpClient) Do(request *http.Request) (*http.Response, error) {
	return c.client.Do(request)
}

func AuthenticatePrompt() SFConfig {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Password + security token: ")
	password, _ := reader.ReadString('\n')

	fmt.Print("Login URL (" + defaultLoginURL + "):")
	url, _ := reader.ReadString('\n')

	if len(url) == 1 && []byte(url)[0] == 10 {
		url = defaultLoginURL
	}

	username = trimString(username)
	password = trimString(password)
	url = strings.TrimSuffix(trimString(url), "/")

	return SFConfig{
		username,
		password,
		url,
	}
}

type soapResult struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		LoginResponse struct {
			Result SFSession `xml:"result"`
		} `xml:"loginResponse"`
	} `xml:"Body"`
	ServerURL string `xml:"serverUrl"`
}

// TODO write generic error handler?

func GetSessionInfo(config SFConfig, c client) (SFSession, error) {
	loginFile, err := ioutil.ReadFile("./auth/login.xml")

	if err != nil {
		log.Println("file read err")
		panic(err)
	}

	t, _ := template.New("login").Parse(string(loginFile))

	var buf bytes.Buffer

	t.Execute(&buf, config)

	req, _ := http.NewRequest("POST", config.LoginURL+"/services/Soap/u/43.0", &buf)
	req.Header.Add("Content-Type", "text/xml")
	req.Header.Add("SOAPAction", "login")

	resp, err := c.Do(req)

	if err != nil {
		return SFSession{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return SFSession{}, err
	}

	var sResult soapResult

	err = xml.Unmarshal(respBody, &sResult)

	if err != nil {
		return SFSession{}, err
	}

	return sResult.Body.LoginResponse.Result, nil
}

// TODO should this be dealt with via --config?
func AuthenticateFromFile(path string) (SFConfig, error) {
	fileBytes, err := ioutil.ReadFile(path)

	log.Println("path", path)

	if err != nil {
		log.Println("authentication from file error")
		return SFConfig{}, err
	}

	var config SFConfig

	// TODO deal with different files
	err = json.Unmarshal(fileBytes, &config)

	if err != nil {
		log.Println("json unmarshal error")
		return SFConfig{}, err
	}

	if len(config.LoginURL) == 0 {
		config.LoginURL = defaultLoginURL
	}

	return config, nil
}

func WriteSession(session SFSession, writer io.Writer) {
	sessionJSON, _ := json.MarshalIndent(session, "", "\t")
	io.WriteString(writer, string(sessionJSON))
}

func trimString(str string) string {
	return strings.TrimSpace(strings.TrimSuffix(str, "\n"))
}
