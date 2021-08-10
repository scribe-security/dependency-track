package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
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

type DepTrackSbomPost struct {
	Project        string `json:"project,omitempty"`
	AutoCreate     string `json:"autoCreate,omitempty"`
	ProjectName    string `json:"projectName,omitempty"`
	ProjectVersion string `json:"projectVersion,omitempty"`
}

type DepTrackSbomPostResponse struct {
	Token string `json:"token"`
	Test  string
}

type DepTrackClient struct {
	*ApiClient
	token       string
	accessToken string
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func NewDepTrackClient(access_token string) (*DepTrackClient, error) {
	cfg := ServiceCfg{ApiToken: access_token, Url: "http://localhost:8081/api/v1", Enable: true}
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

func (depClient *DepTrackClient) GetJsonList(api string) (JSON_LIST, error) {
	var dst JSON_LIST
	err := depClient.GetJson(api, &dst)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

func (depClient *DepTrackClient) GetTeam() (JSON_LIST, error) {
	return depClient.GetJsonList("team")
}

func (depClient *DepTrackClient) PostSbom(api string, deptrack_params *DepTrackSbomPost, bom *cdx.BOM, response *DepTrackSbomPostResponse) error {
	buf := new(bytes.Buffer)

	var extraParams map[string]string
	v, err := json.Marshal(deptrack_params)
	if err != nil {
		return err
	}
	json.Unmarshal([]byte(v), &extraParams)

	multipart_writer, part, err := depClient.MultipartWriter(api, deptrack_params.ProjectName, extraParams, buf)
	if err != nil {
		return err
	}

	encoder := cdx.NewBOMEncoder(part, cdx.BOMFileFormatJSON)
	encoder.SetPretty(true)
	err = encoder.Encode(bom)
	if err != nil {
		return err
	}
	multipart_writer.Close()
	resp, err := depClient.Post("bom", multipart_writer.FormDataContentType(), buf)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	v, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(v, &response)
	if err != nil {
		return err
	}

	return err
}
