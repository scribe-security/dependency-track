package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

type DepTrackTeamPut struct {
	Name string `json:"name"`
}

type DepTrackClient struct {
	*ApiClient
	token       string
	accessToken string
}

func NewDepTrackClient(access_token string) (*DepTrackClient, error) {
	cfg := ServiceCfg{Token: access_token, Url: "http://localhost:8081/api/v1", Enable: true}
	client, err := NewApiClient(&cfg)
	if err != nil {
		return nil, err
	}
	return &DepTrackClient{ApiClient: client}, nil
}

func (depClient *DepTrackClient) Login(username string, password string) error {

	login_values := url.Values{
		"username": {username}, //Read from env DEPEND_TRACK_USER
		"password": {password}, //Read from env DEPEND_TRACK_PASS
	}

	req, err := http.NewRequest("POST", JoinURL(depClient.Cfg.Url, "user/login"), strings.NewReader(login_values.Encode()))
	if err != nil {
		return err
	}

	resp, err := depClient.sendRequest(req, "application/x-www-form-urlencoded")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	login_response_bytes, err := ioutil.ReadAll(resp.Body)

	log.Info("Access token: ", string(login_response_bytes))
	depClient.Cfg.Token = string(login_response_bytes)
	return nil
}

//2DO move to api client
func (depClient *DepTrackClient) Get(api string) error {
	req, _ := http.NewRequest("GET", JoinURL(depClient.Cfg.Url, api), nil)
	resp, err := depClient.sendRequest(req, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)

	dst := &bytes.Buffer{}
	if err := json.Indent(dst, b, "", "  "); err != nil {
		return err
	}

	fmt.Println(dst.String())
	return nil
}

// func (depClient *DepTrackClient) Put(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
// 	req, err := http.NewRequest("PUT", url, body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if depClient.token != "" {
// 		req.Header.Add("Authorization", "Bearer "+depClient.token)
// 	}
// 	req.Header.Set("Content-Type", contentType)
// 	return depClient.Client.Do(req)
// }

// func (depClient *DepTrackClient) PutJson(url string, body interface{}) {
// 	v, _ := json.Marshal(body)
// 	resp, err := depClient.Put(url, "application/json", bytes.NewBuffer(v))
// 	if err != nil {
// 		os.Exit(1)
// 	}
// 	b, _ := ioutil.ReadAll(resp.Body)

// 	dst := &bytes.Buffer{}
// 	if err := json.Indent(dst, b, "", "  "); err != nil {
// 		panic(err)
// 	}

// 	fmt.Println(dst.String())
// }

func (depClient *DepTrackClient) GetTeam() error {
	return depClient.Get("team")
}

// func (depClient *DepTrackClient) NewTeam() {
// 	data := DepTrackTeamPut{Name: "scribe_backend2"}
// 	depClient.PutJson("http://localhost:8081/api/v1/team", data)
// }
