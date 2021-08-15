package client

import (
	"bytes"
	"encoding/json"
<<<<<<< Updated upstream
=======
	"fmt"
>>>>>>> Stashed changes
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

<<<<<<< Updated upstream
=======
type LatestVersionParams struct {
	Purl string `json:"purl,omitempty"`
}

type GetComponentByPurlParams struct {
	Purl string `json:"purl,omitempty"`
}

type GetVulnerabilityByUUIDParams struct {
	Suppressed bool `json:"suppressed,omitempty"`
}

type LatestVersionResponse struct {
	RepositoryType string `json:"repositoryType,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	Name           string `json:"name,omitempty"`
	LatestVersion  string `json:"latestVersion,omitempty"`
	// TBD: Add support to time parsing
	// Published      time.Time `json:"published,omitempty"`
	// LastCheck      time.Time `json:"lastCheck,omitempty"`
}

>>>>>>> Stashed changes
type DepTrackSbomPostResponse struct {
	Token string `json:"token"`
	Test  string
}

type DepTrackClient struct {
	*ApiClient
	token       string
	accessToken string
}

<<<<<<< Updated upstream
=======
type Cwe struct {
	cweId int    `json:"vulnId,omitempty"`
	name  string `json:"name,omitempty"`
}

type Vulnraibility struct {
	VulnId          string  `json:"vulnId,omitempty"`
	Source          string  `json:"source,omitempty"`
	Description     string  `json:"description,omitempty"`
	CvssV2BaseScore float64 `json:"cvssV2BaseScore,omitempty"`
	CvssV3BaseScore float64 `json:"cvssV3BaseScore,omitempty"`
	Cwe             Cwe     `json:"cwe,omitempty"`
	References      string  `json:"references,omitempty"`
	Severity        string  `json:"severity,omitempty"`
	// TBD: Add support to time parsing
	//Published       string  `json:"published,omitempty"`
	//Updated         time.Time `json:"updated,omitempty"`
}

type Component struct {
	Author    string `json:"author,omitempty"`
	Publisher string `json:"publisher,omitempty"`
	Group     string `json:"group,omitempty"`
	Name      string `json:"name,omitempty"`
	Version   string `json:"version,omitempty"`
	Filename  string `json:"filename,omitempty"`
	Extension string `json:"extension,omitempty"`
	Md5       string `json:"md5,omitempty"`
	Sha1      string `json:"sha1,omitempty"`
	Sha256    string `json:"sha256,omitempty"`
	Cpe       string `json:"cpe,omitempty"`
	Purl      Purl   `json:"-,omitempty"`
	UUID      string `json:"uuid,omitempty"`
}

type ComponentList []Component
type VulnraibilityList []Vulnraibility

const GetAllVulnerabilities string = "/vulnerability/component"
const GetAllComponent string = "/component/identity"

>>>>>>> Stashed changes
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
<<<<<<< Updated upstream
=======

func (depClient *DepTrackClient) GetRepositoryLatest(PURL string) (*LatestVersionResponse, error) {
	const LatestApiPath string = "/repository/latest"
	var latestVersion LatestVersionResponse
	params := LatestVersionParams{Purl: PURL}

	err := depClient.GetJsonWithParams(LatestApiPath, params, &latestVersion)
	if err != nil {
		return nil, err
	}

	return &latestVersion, nil
}

func (depClient *DepTrackClient) GetComponentByPURL(PURL string) (ComponentList, error) {
	var component_list ComponentList
	params := GetComponentByPurlParams{Purl: PURL}

	if err := depClient.GetJsonWithParams(GetAllComponent, params, &component_list); err != nil {
		return nil, err
	}

	for _, com := range component_list {
		fmt.Printf("%+v\n", com)
	}

	return component_list, nil
}

func (depClient *DepTrackClient) GetVulnerabilityComponenetByUUID(uuid string, isSupported bool) (VulnraibilityList, error) {

	var vulnraibilityList VulnraibilityList
	//params := GetVulnerabilityByUUIDParams{Suppressed: true}
	FullVulnrabilityPath := GetAllVulnerabilities + "/" + uuid

	if err := depClient.GetJson(FullVulnrabilityPath, &vulnraibilityList); err != nil {
		return nil, err
	}

	for _, com := range vulnraibilityList {
		fmt.Printf("%+v\n", com)
	}

	return vulnraibilityList, nil
}

func (depClient *DepTrackClient) GetLatestVersion(PURL string) (float64, float64, error) {

	latest, err := depClient.GetRepositoryLatest(PURL)
	if err != nil {
		return 0, 0, err
	}
	LatestVersion := latest.LatestVersion
	PURL.
		purl = NewPurl(PURL)
	resp = GetRepositoryLatest(purl.raw)
	return purl.version, LatestVersion
}
>>>>>>> Stashed changes
