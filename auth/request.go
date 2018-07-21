package auth

import (
	"net/url"
	"strings"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"time"
	"errors"
	"github.com/skratchdot/open-golang/open"
)

const (
	oauthEndpoint = "/services/oauth2"
	authEndpoint = "/authorize"
	credEndpoint = "/token"
)

// TODO implement YAML

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
	ClientID string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	URL string `json:"URL"`
	RedirectURI string
}

// if true, these credentials are used for username / password authentication
func (c Credential) IsCredFlow() bool {
	return c.Username != "" && c.Password != "" && c.ClientID != "" && c.ClientSecret != ""
}

// if true, these credentials are used for user flow authentication
func (c Credential) IsUserFlow() bool {
	return c.Username == "" && c.Password == "" && c.ClientID != "" && c.ClientSecret == ""
}

func (c Credential) Encode(redirectURI string) string {
	var u *url.URL
	q := u.Query()

	if c.IsUserFlow() {
		u, _ = url.Parse(strings.TrimSuffix(c.URL, "/") + oauthEndpoint + credEndpoint)
		q.Set("response_type", "token")
		q.Set("client_id", c.ClientID)
		q.Set("redirect_uri", redirectURI)
	} else {
		u, _ = url.Parse(strings.TrimSuffix(c.URL, "/") + oauthEndpoint + authEndpoint)
		q.Set("grant_type", "password")
		q.Set("client_id", c.ClientID)
		q.Set("client_secret", c.ClientSecret)
		q.Set("username", c.Username)
		q.Set("password", c.Password)
	}

	u.RawQuery = q.Encode()

	return u.String()
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType string
	RefreshToken string
	Scope string
	State string
	InstanceURL string `json:"instance_url"`
	ID string `json:"id"`
	IssuedAt string
	Signature string `json:"signature"`
}

func EncodeURL(c Credential) string {
	var u *url.URL

	if c.IsUserFlow() {
		u, _ = url.Parse(strings.TrimSuffix(c.URL, "/") + oauthEndpoint + credEndpoint)
	} else {
		u, _ = url.Parse(strings.TrimSuffix(c.URL, "/") + oauthEndpoint + authEndpoint)
	}

	q := u.Query()
	q.Set("response_type", "token")
	q.Set("client_id", c.ClientID)
	q.Set("redirect_uri", c.RedirectURI)

	u.RawQuery = q.Encode()

	return u.String()
}

func DecodeURL(url string) AuthResponse {
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

	return AuthResponse{
		AccessToken: queryMap["access_token"],
		TokenType: queryMap["token_type"],
		RefreshToken: queryMap["refresh_token"],
		Scope: queryMap["scope"],
		State: queryMap["state"],
		InstanceURL: queryMap["instance_url"],
		ID: queryMap["id"],
		IssuedAt: queryMap["issued_at"],
		Signature: queryMap["signature"],
	}
}

func DecodeJSON(data []byte) AuthResponse {
	var resp AuthResponse

	json.Unmarshal(data, &resp)

	return resp
}

type Opener interface {
	OpenPath(path string)
}

type BrowserOpener struct {
	path string
}

func (b BrowserOpener) OpenPath(path string) {
	open.Start(b.path)
}

func SendAuthRequest(c Credential) (AuthResponse, error) {
	// if username, send request and wait for response
	if c.IsUserFlow() {
		return sendUserFlow(c)
	} else if c.IsCredFlow() {
		return sendCredFlow(c)
	} else {
		return AuthResponse{}, errors.New("invalid credential format")
	}
}

func sendCredFlow(c Credential) (AuthResponse, error) {
	req, _ := http.NewRequest("POST", EncodeURL(c), nil)

	client := http.DefaultClient

	resp, err := client.Do(req)

	if err != nil {
		return AuthResponse{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return AuthResponse{}, err
	}

	var authResp AuthResponse

	err = json.Unmarshal(respBody, &authResp)

	return authResp, err
}

func sendUserFlow(c Credential) (AuthResponse, error) {
	s := Server{
		Port: "42111",
		C: make(chan string),
	}

	open.Start(EncodeURL(c))

	go s.Start()

	ticker := time.NewTicker(3 * time.Minute)

	for {
		select {
		case uri := <- s.C:
			return DecodeURL(uri), nil
		case <- ticker.C:
			close(s.C)
			// TODO make custom error
			return AuthResponse{}, errors.New("server response timed out")

		}
	}
}

