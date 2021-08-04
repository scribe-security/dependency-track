package client

import (
	"io/ioutil"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

type JSON map[string]interface{}
type JSON_LIST []map[string]interface{}

type DepTrackPermission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type DepTrackPermissionsList []DepTrackPermission

type DepTrackTeamPut struct {
	Name        string                  `json:"name"`
	Permissions DepTrackPermissionsList `json:"permissions"`
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

// func (depClient *DepTrackClient) sendRequestForm(req *http.Request, response interface{}) error {

func (depClient *DepTrackClient) Login(username string, password string) error {

	login_values := url.Values{
		"username": {username}, //Read from env DEPEND_TRACK_USER
		"password": {password}, //Read from env DEPEND_TRACK_PASS
	}

	resp, err := depClient.Post("user/login", "application/x-www-form-urlencoded", strings.NewReader(login_values.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	login_response_bytes, err := ioutil.ReadAll(resp.Body)

	log.Info("Access token: ", string(login_response_bytes))
	depClient.Cfg.Token = string(login_response_bytes)
	return nil
}

func (depClient *DepTrackClient) GetJsonList(api string) error {
	var dst JSON_LIST
	err := depClient.GetJson(api, &dst)
	if err != nil {
		return err
	}
	return nil
}

func (depClient *DepTrackClient) GetTeam() error {
	return depClient.GetJsonList("team")
}

// func (depClient *DepTrackClient) PutTeam() error {
// 	var dst JSON
// 	body := DepTrackTeamPut{Name: "scribe_backend2", Permissions: DepTrackPermissionsList{DepTrackPermission{Name: "ACCESS_MANAGEMENT"}}}
// 	v, err := json.Marshal(body)
// 	if err != nil {
// 		return err
// 	}
// 	err = depClient.PutJson("team", bytes.NewBuffer(v), &dst)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(dst)
// 	return nil
// 	// 	depClient.PutJson("http://localhost:8081/api/v1/team", data)
// }
