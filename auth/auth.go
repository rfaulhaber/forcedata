package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
)

// TODO rewrite using connected app, oauth flow

const defaultLoginURL = "https://login.salesforce.com"

const authTemplate = `<?xml version="1.0" encoding="utf-8" ?>
		<env:Envelope xmlns:xsd="http://www.w3.org/2001/XMLSchema"
					  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
					  xmlns:env="http://schemas.xmlsoap.org/soap/envelope/">
		<env:Body>
			<n1:login xmlns:n1="urn:partner.soap.sforce.com">
				<n1:username>{{.Username}}</n1:username>
				<n1:password>{{.Password}}</n1:password>
			</n1:login>
		</env:Body>
	</env:Envelope>`

// TODO implement other formats?
type SFConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	LoginURL string `json:"loginUrl"`
}

type SFSession struct {
	ServerURL string `xml:"serverUrl" json:"serverUrl"`
	SessionID string `xml:"sessionId" json:"sessionId"`
}

type Prompter interface {
	Prompt(prompt string) string
}

func AuthenticatePrompt(in io.Reader, out io.Writer) SFConfig {
	reader := bufio.NewReader(in)
	io.WriteString(out, "Username: ")
	username, _ := reader.ReadString('\n')

	io.WriteString(out, "Password + security token: ")
	password, _ := reader.ReadString('\n')

	io.WriteString(out, "Login URL (" + defaultLoginURL + "):")
	url, _ := reader.ReadString('\n')

	if len(url) <= 1 {
		url = defaultLoginURL
	}

	username, password, url = cleanInput(username, password, url)

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
		Fault soapFault `xml:"Fault"`
	} `xml:"Body"`
	ServerURL string `xml:"serverUrl"`
}

type soapFault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}

// TODO write generic error handler?

func GetSessionInfo(config SFConfig) (SFSession, error) {
	t, _ := template.New("login").Parse(authTemplate)

	var buf bytes.Buffer

	t.Execute(&buf, config)

	req, _ := http.NewRequest("POST", config.LoginURL+"/services/Soap/u/43.0", &buf)
	req.Header.Add("Content-Type", "text/xml")
	req.Header.Add("SOAPAction", "login")

	client := http.DefaultClient

	resp, err := client.Do(req)

	if err != nil {
		return SFSession{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	log.Println("server response", string(respBody))

	if err != nil {
		return SFSession{}, err
	}

	// TODO handle auth error
	// TODO custom errors?

	var sResult soapResult

	err = xml.Unmarshal(respBody, &sResult)

	if err != nil {
		return SFSession{}, err
	}

	log.Println("fault", sResult.Body.Fault)

	if sResult.Body.Fault != (soapFault{}) {
		return SFSession{}, errors.New("server responded with: " + sResult.Body.Fault.FaultString)
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

	// TODO append http if not present?
	//if !strings.HasPrefix(config.LoginURL, "https://") {
	//	config.LoginURL = "https://" + config.LoginURL
	//}

	return config, nil
}

func WriteSession(session SFSession, writer io.Writer) {
	sessionJSON, _ := json.MarshalIndent(session, "", "\t")
	io.WriteString(writer, string(sessionJSON))
}

func cleanInput(username, password, url string) (string, string, string) {
	username = trimString(username)
	password = trimString(password)
	url = strings.TrimSuffix(trimString(url), "/")

	return username, password, url
}

func trimString(str string) string {
	return strings.TrimSpace(strings.TrimSuffix(str, "\n"))
}
