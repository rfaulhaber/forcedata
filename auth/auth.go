package auth

import (
	"strings"
	"io/ioutil"
	"log"
	"encoding/json"
	"io"
	"errors"
)

// TODO write custom errors for Salesforce errors!

const defaultLoginURL = "https://login.salesforce.com"

type Session struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
	InstanceURL  string `json:"instance_url"`
	ID           string `json:"id"`
	IssuedAt     string `json:"issued_at"`
	Signature    string `json:"signature"`
}

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

	resp, err := SendAuthRequest(creds, DefaultOpener)

	if err != nil {
		return Session{}, err
	}

	return resp, nil
}

func WriteSession(session Session, writer io.Writer) {
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

	creds.RedirectURI = "http://localhost:42111/callback"

	if creds.URL == "" {
		creds.URL = defaultLoginURL
	}

	return creds, nil
}

// TODO make custom error types!

func validateCreds(creds Credential) error {
	if creds.ClientID == "" {
		return errors.New("missing required field ClientID")
	} else {
		return nil
	}
}
