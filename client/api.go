package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	JSON_MEDIA_TYPE     = "application/json"
	TEXT_MEDIA_TYPE     = "text/plain"
	CONTENT_TYPE_FORMAT = "%s; charset=utf-8"
)

type ServiceCfg struct {
	Url      string `yaml:"url" json:"url" mapstructure:"url"`
	Token    string `yaml:"token" json:"token" mapstructure:"token"`
	ApiToken string `yaml:"apitoken" json:"apitoken" mapstructure:"apitoken"`
	Username string `yaml:"username" json:"username" mapstructure:"username"`
	Password string `yaml:"password" json:"password" mapstructure:"password"`
	Enable   bool   `yaml:"enable" json:"enable" mapstructure:"enable"`
}

type ApiClient struct {
	Cfg        *ServiceCfg
	HTTPClient *http.Client
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type successResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func JoinURL(base string, paths ...string) string {
	p := path.Join(paths...)
	return fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(p, "/"))
}

//2DO add specific tls.Config
func NewApiClient(cfg *ServiceCfg) (*ApiClient, error) {

	if !cfg.Enable {
		return nil, errors.New("Disabled")
	}

	if cfg.Url == "" {
		return nil, errors.New("URL not set")
	}

	return &ApiClient{
		Cfg: cfg,
		HTTPClient: &http.Client{
			Timeout: time.Second * 2,
		},
	}, nil
}

func (c ApiClient) Is_enabled() bool {
	if c.Cfg.Enable && c.Cfg.Url == "" {
		log.Warn("Url not set")
		return false
	}

	return c.Cfg.Enable
}

func (c *ApiClient) NewRequest(method, api string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, JoinURL(c.Cfg.Url, api), body)

}

func (c *ApiClient) sendRequestJson(req *http.Request, response interface{}) error {
	resp, err := c.sendRequest(req, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	v, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(v, &response)
	if err != nil {
		return err
	}
	// if response != nil {
	// 	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
	// 		return err
	// 	}

	// }
	return nil
}

func (c *ApiClient) Post(api, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := c.NewRequest("POST", api, body)
	if err != nil {
		return nil, err
	}
	return c.sendRequest(req, contentType)
}

func (c *ApiClient) PutJson(api string, body io.Reader, response interface{}) error {
	req, err := c.NewRequest("PUT", api, body)
	if err != nil {
		return err
	}
	return c.sendRequestJson(req, response)
}

func (c *ApiClient) Put(api, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := c.NewRequest("PUT", api, body)
	if err != nil {
		return nil, err
	}
	return c.sendRequest(req, contentType)
}

func (c *ApiClient) GetJson(api string, response interface{}) (err error) {
	req, err := c.NewRequest("GET", api, nil)
	if err != nil {
		return err
	}
	return c.sendRequestJson(req, response)
}

func (c *ApiClient) Get(api string) (resp *http.Response, err error) {
	req, err := c.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	return c.sendRequest(req, "")
}

func (c *ApiClient) SetAuthorization(req *http.Request) bool {
	if c.Cfg.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Cfg.Token))
		return true
	}
	return false
}

func (c *ApiClient) SetApiKey(req *http.Request) bool {
	if c.Cfg.ApiToken != "" {
		req.Header.Set("X-Api-Key", c.Cfg.ApiToken)
		return true
	}
	return false
}

func (c *ApiClient) SetBasicAuth(req *http.Request) bool {
	if c.Cfg.Username != "" && c.Cfg.Password != "" {
		req.SetBasicAuth(c.Cfg.Username, c.Cfg.Password)
		return true
	}
	return false
}

func (c *ApiClient) sendRequest(req *http.Request, Content_type string) (*http.Response, error) {
	if c.Cfg != nil && c.Cfg.Url == "" {
		return nil, errors.New("URL not set")
	}

	if Content_type != "" {
		req.Header.Set("Content-Type", fmt.Sprintf(CONTENT_TYPE_FORMAT, Content_type))
	}

	if c.SetAuthorization(req) {
		log.Debug("Api client - Authorization")
	} else if c.SetApiKey(req) {
		log.Debug("Api client - Api key auth")
	} else if c.SetBasicAuth(req) {
		log.Debug("Api client - Basic auth")
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("Api client - response, url: %s, status_code: %d, Text: %s", req.URL, res.StatusCode, http.StatusText(res.StatusCode))
	}

	log.Infof("Api client - Post URL: %s,  Status: %d, Text: %s", req.URL, res.StatusCode, http.StatusText(res.StatusCode))

	return res, nil
}
