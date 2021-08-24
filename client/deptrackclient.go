package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"io/ioutil"
	"net/url"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	retry "github.com/avast/retry-go"
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

type GetProjectLookupParams struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type GetProjectParams struct {
	Name string `json:"name,omitempty"`
}

type GetComponentsIdentityParams struct {
	Group     string `json:"group,omitempty"`
	Name      string `json:"name,omitempty"`
	Version   string `json:"version,omitempty"`
	Purl      string `json:"purl,omitempty"`
	Cpe       string `json:"cpe,omitempty"`
	SwidTagId string `json:"swidTagId,omitempty"`
	PaginationParams
}

type GetVulnerabilityByUUIDParams struct {
	Suppressed bool `json:"suppressed,omitempty"`
}

type VersionResponse struct {
	RepositoryType string `json:"repositoryType,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	Name           string `json:"name,omitempty"`
	LatestVersion  string `json:"latestVersion,omitempty"`
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
	Processing bool `json:"processing,omitempty"`
}

type PaginationParams struct {
	Offset string `json:"offset,omitempty"`
	Limit  string `json:"limit,omitempty"`
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

type Project struct {
	Name                   string       `json:"name,omitempty"`
	UUID                   string       `json:"uuid,omitempty"`
	LastInheritedRiskScore float64      `json:"lastInheritedRiskScore,omitempty"`
	LastBomImportFormat    string       `json:"lastBomImportFormat,omitempty"`
	Active                 bool         `json:"active,omitempty"`
	Metrics                Metrics_stat `json:"metrics,omitempty"`
}

type Metrics_stat struct {
	Vulnerabilities      int `json:"vulnerabilities,omitempty"`
	VulnerableComponents int `json:"vulnerableComponents,omitempty"`
	Component            int `json:"component,omitempty"`
}

type ProjectList []Project

type PurlVersionStruct struct {
	CurrentVersion *packageurl.PackageURL
	LatestVersion  *packageurl.PackageURL
	IsVersionEquel bool
}

type VulnraibilityListMap map[string]VulnraibilityList
type PurlVersionStructMap map[string]PurlVersionStruct
type ComponentList []Component
type VulnraibilityList []Vulnraibility

const (
	ApiComponentProject      = "/component/project"
	ApiVulnrabilityComponent = "/vulnerability/component"
	ApiComponentIdentity     = "/component/identity"
	ApiProjectLookup         = "/project/lookup"
	ApiProject               = "/project"
	ApiSbomTokenQuery        = "/bom/token"
	ApiRepositoryLatest      = "/repository/latest"
	ApiUserLoginPath         = "user/login"
	ApiServerPath            = "http://localhost:8081/api/v1"
)

var DefaultPagination = PaginationParams{Offset: "0", Limit: "10000"}

func NewDepTrackClient(access_token string) (*DepTrackClient, error) {
	cfg := ServiceCfg{ApiToken: access_token, Url: ApiServerPath, Enable: true}
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

	resp, err := depClient.Post(ApiUserLoginPath, "application/x-www-form-urlencoded", strings.NewReader(login_values.Encode()))
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

	err := depClient.GetJsonWithParams(ApiRepositoryLatest, params, &latestVersion)
	if err != nil {
		return nil, err
	}

	return &latestVersion, nil
}

func (depClient *DepTrackClient) GetComponentsIdentity(params GetComponentsIdentityParams) (ComponentList, error) {
	var component_list ComponentList
	if err := depClient.GetJsonWithParams(ApiComponentIdentity, params, &component_list); err != nil {
		return nil, err
	}

	return component_list, nil
}

func (depClient *DepTrackClient) GetProjectLookup(params GetProjectLookupParams) (*Project, error) {

	var project Project
	if err := depClient.GetJsonWithParams(ApiProjectLookup, params, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

func (depClient *DepTrackClient) GetProject(params GetProjectParams) (ProjectList, error) {
	var project_list ProjectList
	if err := depClient.GetJsonWithParams(ApiProject, params, &project_list); err != nil {
		return nil, err
	}
	return project_list, nil
}

func (depClient *DepTrackClient) GetComponentsByProjectUUID(uuid string, pagination_param *PaginationParams) (ComponentList, error) {
	var component_list ComponentList
	full_api := ApiComponentProject + "/" + uuid
	if pagination_param != nil {
		if err := depClient.GetJsonWithParams(full_api, pagination_param, &component_list); err != nil {
			return nil, err
		}
	} else {
		if err := depClient.GetJson(full_api, &component_list); err != nil {
			return nil, err
		}
	}

	return component_list, nil
}

func (depClient *DepTrackClient) GetVulnerabilityComponenetByUUID(uuid string, isSupported bool, pagination_param *PaginationParams) (VulnraibilityList, error) {

	var vulnraibilityList VulnraibilityList
	full_api := ApiVulnrabilityComponent + "/" + uuid
	if pagination_param != nil {
		if err := depClient.GetJsonWithParams(full_api, pagination_param, &vulnraibilityList); err != nil {
			return nil, err
		}
	} else {
		if err := depClient.GetJson(full_api, &vulnraibilityList); err != nil {
			return nil, err
		}
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
	component_list, err := depClient.GetComponentsIdentity(GetComponentsIdentityParams{Purl: PURL})
	var final_vulnraibility_list VulnraibilityList
	if err != nil {
		return nil, err
	}
	for _, componenet := range component_list {
		vulnraibility_list, err := depClient.GetVulnerabilityComponenetByUUID(componenet.UUID, true, &DefaultPagination)
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
		// log.Debugf("Old version found, Name: %s, current: %s, latest: %s\n", component.Name, current_version, latest_version)
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

		// log.Debugf("Vulnrability found, Name: %s, Purl: %s, Num: %d\n", component.Name, component.PackageURL, len(vulnraibility_list))
		components_map[component.Name] = vulnraibility_list
	}

	return components_map, nil
}

func (depClient *DepTrackClient) GetSbomDetails(sbom_uuid string) (bool, error) {
	var sbomProcessingState SbomProcessingState
	sbom_token_query := ApiSbomTokenQuery + "/" + sbom_uuid
	if err := depClient.GetJson(sbom_token_query, &sbomProcessingState); err != nil {
		return false, err
	}

	return sbomProcessingState.Processing, nil
}

func (depClient *DepTrackClient) WaitforSbomFinishUpload(sbom_uuid string) (bool, error) {
	// DefaultAttempts := uint(30)
	DefaultDelay := 50 * time.Millisecond
	// DefaultMaxJitter := 10 * time.Millisecond
	// DefaultOnRetry := func(n uint, err error) {}
	// DefaultRetryIf := IsRecoverable
	// DefaultDelayType := retry.CombineDelay(retry.BackOffDelay, RandomDelay)
	// DefaultLastErrorOnly := false
	// DefaultContext := context.Background()

	err := retry.Do(
		func() error {
			is_finished, err := depClient.GetSbomDetails(sbom_uuid)
			if err != nil {
				return err
			}
			if !is_finished {
				return nil
			}
			return errors.New("Processing")
		},
		retry.Delay(DefaultDelay),
	)
	if err != nil {
		return false, err
	}
	// If not processing return true
	return true, nil
}
