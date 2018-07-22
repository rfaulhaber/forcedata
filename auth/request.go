package auth

import (
	"encoding/json"
	"errors"
	"github.com/skratchdot/open-golang/open"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
	"log"
)

const (
	oauthEndpoint = "/services/oauth2"
	userFlowEndpoint = "/authorize"
	credEndpoint  = "/token"
	defaultRedirectURI = "http://localhost:42111/callback"
)

// TODO implement YAML

type Credential struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	URL          string `json:"url"`
	RedirectURI  string
}

// if true, these credentials are used for username / password authentication
func (c Credential) IsCredFlow() bool {
	return c.Username != "" && c.Password != "" && c.ClientID != "" && c.ClientSecret != ""
}

// if true, these credentials are used for user flow authentication
func (c Credential) IsUserFlow() bool {
	return c.Username == "" && c.Password == "" && c.ClientID != "" && c.ClientSecret == ""
}

func (c Credential) Encode() string {
	var u *url.URL
	var q url.Values

	if c.IsUserFlow() {
		u, _ = url.Parse(strings.TrimSuffix(c.URL, "/") + oauthEndpoint + userFlowEndpoint)
		q = u.Query()
		q.Set("response_type", "token")
		q.Set("client_id", c.ClientID)
		q.Set("redirect_uri", defaultRedirectURI)
	} else {
		u, _ = url.Parse(strings.TrimSuffix(c.URL, "/") + oauthEndpoint + credEndpoint)
		q = u.Query()
		q.Set("grant_type", "password")
		q.Set("client_id", c.ClientID)
		q.Set("client_secret", c.ClientSecret)
		q.Set("username", c.Username)
		q.Set("password", c.Password)
	}

	u.RawQuery = q.Encode()

	return u.String()
}

func DecodeURL(url string) Session {
	startIndex := strings.Index(url, "#")
	queryStr := url[startIndex+1:]

	kv := strings.Split(queryStr, "&")

	queryMap := make(map[string]string)

	for i := range kv {
		q := strings.Split(kv[i], "=")

		key := q[0]
		val := q[1]
		queryMap[key] = val
	}

	return Session{
		AccessToken:  queryMap["access_token"],
		TokenType:    queryMap["token_type"],
		RefreshToken: queryMap["refresh_token"],
		Scope:        queryMap["scope"],
		State:        queryMap["state"],
		InstanceURL:  queryMap["instance_url"],
		ID:           queryMap["id"],
		IssuedAt:     queryMap["issued_at"],
		Signature:    queryMap["signature"],
	}
}

func DecodeJSON(data []byte) Session {
	var resp Session

	json.Unmarshal(data, &resp)

	return resp
}

type Opener interface {
	OpenPath(path string)
}

type BrowserOpener struct{}

func (b BrowserOpener) OpenPath(path string) {
	open.Start(path)
}

var DefaultOpener = BrowserOpener{}

func SendAuthRequest(c Credential, o Opener) (Session, error) {
	// if username, send request and wait for response
	if c.IsUserFlow() {
		return sendUserFlow(c, o)
	} else if c.IsCredFlow() {
		return sendCredFlow(c)
	} else {
		return Session{}, errors.New("invalid credential format")
	}
}

func sendCredFlow(c Credential) (Session, error) {
	req, _ := http.NewRequest("POST", c.Encode(), nil)

	client := http.DefaultClient

	log.Println("making request for cred flow")

	resp, err := client.Do(req)

	if err != nil {
		return Session{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return Session{}, err
	}

	var authResp Session

	err = json.Unmarshal(respBody, &authResp)

	return authResp, err
}

func sendUserFlow(c Credential, o Opener) (Session, error) {
	s := Server{
		Port: "42111",
		C:    make(chan string),
	}

	log.Println("user flow: attemping to open encoded url", c.Encode())
	o.OpenPath(c.Encode())

	log.Println("starting response server")
	go s.Start()

	log.Println("starting ticker")
	ticker := time.NewTicker(3 * time.Minute)

	for {
		select {
		case uri := <-s.C:
			log.Println("received response", uri)
			return DecodeURL(uri), nil
		case <-ticker.C:
			close(s.C)
			return Session{}, errors.New("server response timed out")

		}
	}
}
