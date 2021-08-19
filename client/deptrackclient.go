package client

import (
	"bytes"
	"encoding/json"

	"io/ioutil"
	"net/url"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	packageurl "github.com/package-url/packageurl-go"
	log "github.com/sirupsen/logrus"
)

type JSON map[string]interface{}
type JSON_LIST []map[string]interface{}

type References []string

type PurlQualifiers struct {
	AdditionalProp1 string `json:"additionalProp1,omitempty"`
	AdditionalProp2 string `json:"additionalProp2,omitempty"`
	AdditionalProp3 string `json:"additionalProp3,omitempty"`
}

type Purl struct {
	Scheme     string         `json:"scheme,omitempty"`
	Type       string         `json:"type,omitempty"`
	Namespace  string         `json:"namespace,omitempty"`
	Name       string         `json:"name,omitempty"`
	Version    string         `json:"version,omitempty"`
	Qualifiers PurlQualifiers `json:"qualifiers,omitempty"`
	Subpath    string         `json:"subpath,omitempty"`
}

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

type LatestVersionParams struct {
	Purl string `json:"purl,omitempty"`
}

type GetComponentsByPURLParams struct {
	Purl string `json:"purl,omitempty"`
}

type GetVulnerabilityByUUIDParams struct {
	Suppressed bool `json:"suppressed,omitempty"`
}

type VersionResponse struct {
	RepositoryType string `json:"repositoryType,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	Name           string `json:"name,omitempty"`
	LatestVersion  string `json:"latestVersion,omitempty"`
	// Published      time.Time `json:"published,omitempty"`
	// LastCheck      time.Time `json:"lastCheck,omitempty"`
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

type Cwe struct {
	cweId int `json:"vulnId,omitempty"`
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
type SbomProcessingState struct {
	processing bool `json:"processing,omitempty"`
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

type PurlVersionStruct struct {
	CurrentVersion *packageurl.PackageURL
	LatestVersion  *packageurl.PackageURL
	isVersionEquel bool
}

type VulnraibilityListStruct struct {
	VulnraibilityList VulnraibilityList
}

type VulnraibilityListMap map[string]VulnraibilityListStruct
type PurlVersionStructMap map[string]PurlVersionStruct
type ComponentList []Component
type VulnraibilityList []Vulnraibility

const GetAllVulnerabilities string = "/vulnerability/component"
const GetAllComponent string = "/component/identity"

func NewDepTrackClient(access_token string) (*DepTrackClient, error) {
	cfg := ServiceCfg{ApiToken: access_token, Url: "http://localhost:8081/api/v1", Enable: true}
	client, err := NewApiClient(&cfg, false, 0)
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
	if err != nil {
		return err
	}

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

func (depClient *DepTrackClient) removeComponents(bom *cdx.BOM) {
	var filtered_componenets []cdx.Component
	for _, component := range *bom.Components {
		if component.Type != "file" {
			filtered_componenets = append(filtered_componenets, component)
		}
	}
	bom.Components = &filtered_componenets
}

func (depClient *DepTrackClient) PostSbom(api string, deptrack_params *DepTrackSbomPost, bom *cdx.BOM, response *DepTrackSbomPostResponse) error {
	depClient.removeComponents(bom)

	buf := new(bytes.Buffer)

	// 2DO removing FILE and DEP graphs (to much work for deptrack)
	bom.Dependencies = nil
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

func (depClient *DepTrackClient) GetRepositoryLatest(PURL string) (*VersionResponse, error) {
	var latestVersion VersionResponse
	params := LatestVersionParams{Purl: PURL}

	err := depClient.GetJsonWithParams("/repository/latest", params, &latestVersion)
	if err != nil {
		return nil, err
	}

	return &latestVersion, nil
}

func (depClient *DepTrackClient) GetComponentsByPURL(PURL string) (ComponentList, error) {
	var component_list ComponentList
	params := GetComponentsByPURLParams{Purl: PURL}

	if err := depClient.GetJsonWithParams(GetAllComponent, params, &component_list); err != nil {
		return nil, err
	}

	return component_list, nil
}

func (depClient *DepTrackClient) GetVulnerabilityComponenetByUUID(uuid string, isSupported bool) (VulnraibilityList, error) {

	var vulnraibilityList VulnraibilityList
	FullVulnrabilityPath := GetAllVulnerabilities + "/" + uuid

	if err := depClient.GetJson(FullVulnrabilityPath, &vulnraibilityList); err != nil {
		return nil, err
	}

	return vulnraibilityList, nil
}

func (depClient *DepTrackClient) GetLatestVersion(PURL string) (*packageurl.PackageURL, *packageurl.PackageURL, bool, error) {
	latest_version_response, err := depClient.GetRepositoryLatest(PURL)
	if err != nil {
		return nil, nil, false, err
	}

	parsed_purl, err := packageurl.FromString(PURL)
	if err != nil {
		return nil, nil, false, err
	}

	latest_parsed_purl, _ := depClient.LatestToPurl(parsed_purl, latest_version_response)

	return latest_parsed_purl, &parsed_purl, CmpPurl(&parsed_purl, latest_parsed_purl), err
}

func (depClient *DepTrackClient) LatestToPurl(base packageurl.PackageURL, resp *VersionResponse) (*packageurl.PackageURL, error) {
	new_purl := packageurl.NewPackageURL(strings.ToLower(resp.RepositoryType), resp.Namespace, resp.Name, resp.LatestVersion, base.Qualifiers, base.Subpath)
	normlized_new_purl, err := packageurl.FromString(new_purl.ToString())
	if err != nil {
		return nil, err
	}
	return &normlized_new_purl, nil
}

func CmpPurl(a *packageurl.PackageURL, b *packageurl.PackageURL) bool {
	return a.ToString() == b.ToString()
}

func (depClient *DepTrackClient) GetVulnraibilityList(PURL string) (VulnraibilityList, error) {
	component_list, err := depClient.GetComponentsByPURL(PURL)
	var final_vulnraibility_list VulnraibilityList
	if err != nil {
		return nil, err
	}
	for _, componenet := range component_list {
		vulnraibility_list, err := depClient.GetVulnerabilityComponenetByUUID(componenet.UUID, true)
		if err != nil {
			return nil, err
		}
		final_vulnraibility_list = append(final_vulnraibility_list, vulnraibility_list...)
	}

	return final_vulnraibility_list, err
}

func (depClient *DepTrackClient) GetLatestVersionBySbom(bom *cdx.BOM) (PurlVersionStructMap, error) {
	components_map := make(PurlVersionStructMap)

	for _, component := range *bom.Components {

		if component.Type != "library" {
			continue
		}
		if component.PackageURL == "" {
			log.Debugf("PURL is empty skippimg, Name: %s", component.Name)
			continue
		}
		current_version, latest_version, is_version_equel, err := depClient.GetLatestVersion(component.PackageURL)
		if err != nil {
			log.Debugf("Get Latest version error skipping, Purl: %s Err: %+v", component.PackageURL, err)
			continue
		}
		components_map[component.Name] = PurlVersionStruct{current_version, latest_version, is_version_equel}
	}

	return components_map, nil
}

func (depClient *DepTrackClient) GetVulnraibilityListBySbom(bom *cdx.BOM) (VulnraibilityListMap, error) {
	components_map := make(VulnraibilityListMap)

	for _, component := range *bom.Components {
		if component.Type != "library" {
			continue
		}

		if component.PackageURL == "" {
			log.Debugf("PURL is empty skippimg, Name: %s", component.Name)
			continue
		}
		vulnraibility_list, err := depClient.GetVulnraibilityList(component.PackageURL)
		if err != nil {
			log.Debugf("Get vulnraibility error skipping, Purl: %s Err: %+v", component.PackageURL, err)
			return nil, err
		}
		if len(vulnraibility_list) == 0 {
			continue
		}
		components_map[component.Name] = VulnraibilityListStruct{vulnraibility_list}
	}

	return components_map, nil
}

func (depClient *DepTrackClient) IsSbomFinishedToUpload(sbom_uuid string) (bool, error) {
	var sbomProcessingState SbomProcessingState
	sbom_token_query := "/bom/token/" + sbom_uuid

	if err := depClient.GetJson(sbom_token_query, &sbomProcessingState); err != nil {
		return false, err
	}
	// If still processing return the false
	return !sbomProcessingState.processing, nil
}
