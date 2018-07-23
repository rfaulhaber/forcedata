package auth

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// TODO write custom errors for Salesforce errors!

const (
	defaultLoginURL = "https://login.salesforce.com"
	authEndpoint    = "/services/oauth2/token"
)

type Credential struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	URL          string `json:"url"`
}

func (c Credential) Encode() string {
	var u *url.URL

	u, _ = url.Parse(strings.TrimSuffix(c.URL, "/") + authEndpoint)
	q := u.Query()
	q.Set("grant_type", "password")
	q.Set("client_id", c.ClientID)
	q.Set("client_secret", c.ClientSecret)
	q.Set("username", c.Username)
	q.Set("password", c.Password)

	u.RawQuery = q.Encode()

	return u.String()
}

type Session struct {
	AccessToken string `json:"access_token" mapstructure:"access_token"`
	InstanceURL string `json:"instance_url" mapstructure:"instance_url"`
	ID          string `json:"id" mapstructure:"id"`
	IssuedAt    string `json:"issued_at" mapstructure:"issued_at"`
	Signature   string `json:"signature" mapstructure:"signature"`
}

// TODO implement!
// func AuthenticateFromPrompt(in io.Reader, out io.Writer) (Session, err)

func AuthenticateFromFile(path string) (Session, error) {
	creds, err := getCredsFromFile(path)

	if err != nil {
		return Session{}, err
	}

	err = validateCreds(creds)

	if err != nil {
		return Session{}, err
	}

	resp, err := SendAuthRequest(creds)

	if err != nil {
		return Session{}, err
	}

	return resp, nil
}

func WriteSession(session Session, writer io.Writer) {
	sessionJSON, _ := json.MarshalIndent(session, "", "\t")
	io.WriteString(writer, string(sessionJSON))
}

func SendAuthRequest(c Credential) (Session, error) {
	req, _ := http.NewRequest("POST", c.Encode(), nil)

	client := http.DefaultClient

	log.Println("making request for cred flow:", c.Encode())

	resp, err := client.Do(req)

	if err != nil {
		return Session{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	log.Println("raw response", string(respBody))

	if err != nil {
		return Session{}, err
	}

	return decodeJSON(respBody)
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

func getCredsFromFile(path string) (Credential, error) {
	fileBytes, err := ioutil.ReadFile(path)

	log.Println("path", path)

	if err != nil {
		log.Println("authentication from file error")
		return Credential{}, err
	}

	var creds Credential

	err = json.Unmarshal(fileBytes, &creds)

	if err != nil {
		log.Println("json unmarshal error")
		return Credential{}, err
	}

	if creds.URL == "" {
		creds.URL = defaultLoginURL
	}

	return creds, nil
}

type MissingFieldError struct {
	field string
}

func (e MissingFieldError) Error() string {
	return "Missing required field: " + e.field
}

func (e MissingFieldError) Field() string {
	return e.field
}

func validateCreds(creds Credential) error {
	if creds.ClientID == "" {
		return MissingFieldError{"client_id"}
	} else if creds.ClientSecret == "" {
		return MissingFieldError{"client_secret"}
	} else if creds.Username == "" {
		return MissingFieldError{"username"}
	} else if creds.Password == "" {
		return MissingFieldError{"password"}
	} else {
		return nil
	}
}

func decodeJSON(data []byte) (Session, error) {
	var resp Session

	err := json.Unmarshal(data, &resp)

	return resp, err
}
