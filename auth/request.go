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

type AuthRequest struct {
	URL string
	ResponseType string
	ClientId string
	RedirectURI string
	Username string
	Password string
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

func Encode(rq AuthRequest) string {
	var u *url.URL

	if rq.Username != "" && rq.Password != "" {
		u, _ = url.Parse(strings.TrimSuffix(rq.URL, "/") + credEndpoint)
	} else {
		u, _ = url.Parse(strings.TrimSuffix(rq.URL, "/") + authEndpoint)
	}

	q := u.Query()
	q.Set("response_type", "token")
	q.Set("client_id", rq.ClientId)
	q.Set("redirect_uri", rq.RedirectURI)

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

func SendAuthRequest(authReq AuthRequest) (AuthResponse, error) {
	// if username, send request and wait for response
	if isCred(authReq) {
		req, _ := http.NewRequest("POST", Encode(authReq), nil)

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
	} else {
		s := Server{
			Port: "42111",
			C: make(chan string),
		}

		open.Start(Encode(authReq))

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

}

func isCred(req AuthRequest) bool {
	return req.Username != "" && req.Password != ""
}
