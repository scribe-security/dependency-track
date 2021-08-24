package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path"
	"strings"
	"time"

	// "github.com/scribe-security/scribe/pkg/log_wrapper"
	"github.com/sirupsen/logrus"
)

const (
	JSON_MEDIA_TYPE     = "application/json"
	TEXT_MEDIA_TYPE     = "text/plain"
	CONTENT_TYPE_FORMAT = "%s; charset=utf-8"
)

type ServiceCfg struct {
	Url                  string `yaml:"url" json:"url" mapstructure:"url"`
	Token                string `yaml:"token" json:"token" mapstructure:"token"`
	ApiToken             string `yaml:"apitoken" json:"apitoken" mapstructure:"apitoken"`
	Username             string `yaml:"username" json:"username" mapstructure:"username"`
	Password             string `yaml:"password" json:"password" mapstructure:"password"`
	Enable               bool   `yaml:"enable" json:"enable" mapstructure:"enable"`
	Insecure_skip_verify bool   `yaml:"-" json:"-" mapstructure:"host_match_validation"`
}

type ApiClient struct {
	Cfg        *ServiceCfg
	HTTPClient *http.Client
	Log        Logger
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type successResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func JoinURL(base string, paths ...string) string {
	p := path.Join(paths...)
	return fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(p, "/"))
}

func DefaultTlsTransport(server_name string, insecure_skip_verify bool, min_version uint16) *http.Transport {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	config := &tls.Config{
		InsecureSkipVerify: insecure_skip_verify,
		RootCAs:            rootCAs,
		ServerName:         server_name,
		MinVersion:         min_version,
	}

	tr := &http.Transport{TLSClientConfig: config}
	return tr
}

func DefaultHttpClient(server_name string, host_match_validation bool, min_version uint16) *http.Client {
	tr := DefaultTlsTransport(server_name, host_match_validation, min_version)
	return &http.Client{
		Timeout:   time.Second * 5,
		Transport: tr,
	}
}

// 2DO add specific tls.Config
func NewApiClient(cfg *ServiceCfg, https_only bool, min_version uint16) (*ApiClient, error) {
	if !cfg.Enable {
		return nil, errors.New("Disabled")
	}

	if cfg.Url == "" {
		return nil, errors.New("URL not set")
	}

	url, err := url.Parse(cfg.Url)
	if err != nil {
		return nil, errors.New("URL not set")
	}
	if https_only && url.Scheme != "https" {
		return nil, errors.New("Unsecure url")
	}

	if (https_only || url.Scheme == "https") && min_version != tls.VersionTLS12 && min_version != tls.VersionTLS13 {
		return nil, errors.New("Unknown tls type")
	}

	return &ApiClient{
		Cfg:        cfg,
		HTTPClient: DefaultHttpClient(url.Hostname(), cfg.Insecure_skip_verify, min_version),
		Log:        logrus.New(),
	}, nil
}

func (c *ApiClient) SetLogger(log Logger) {
	c.Log = log
}

func (c ApiClient) Is_enabled() bool {
	if c.Cfg.Enable && c.Cfg.Url == "" {
		c.Log.Warn("Url not set")
		return false
	}

	return c.Cfg.Enable
}

func (c *ApiClient) NewRequest(method, api string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, JoinURL(c.Cfg.Url, api), body)
}

func (c *ApiClient) ReadJson(resp *http.Response, response interface{}) error {
	defer resp.Body.Close()
	if response == nil {
		c.Log.Debug("Api client - skipping, parse response")
		return nil
	}

	v, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(v, &response)
	if err != nil {
		return err
	}
	return nil
}

func (c *ApiClient) SendRequestReadJson(req *http.Request, Content_type string, response interface{}) error {
	resp, err := c.SendRequest(req, Content_type)
	if err != nil {
		return err
	}

	if err := c.ReadJson(resp, response); err != nil {
		return err
	}
	return nil
}

func (c *ApiClient) SendRequestJson(req *http.Request, response interface{}) error {
	return c.SendRequestReadJson(req, "application/json", response)
}

func (c *ApiClient) PostMultipart(api string, body io.Reader) (*http.Response, error) {
	req, err := c.NewRequest("POST", api, body)
	if err != nil {
		return nil, err
	}
	return c.SendRequest(req, "multipart/form-data")
}

func (c *ApiClient) PostJson(api string, body io.Reader, response interface{}) error {
	req, err := c.NewRequest("POST", api, body)
	if err != nil {
		return err
	}
	return c.SendRequestJson(req, response)
}

func (c *ApiClient) Post(api, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := c.NewRequest("POST", api, body)
	if err != nil {
		return nil, err
	}
	return c.SendRequest(req, contentType)
}

func (c *ApiClient) PutJson(api string, body io.Reader, response interface{}) error {
	req, err := c.NewRequest("PUT", api, body)
	if err != nil {
		return err
	}
	return c.SendRequestJson(req, response)
}

func (c *ApiClient) Put(api, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := c.NewRequest("PUT", api, body)
	if err != nil {
		return nil, err
	}
	return c.SendRequest(req, contentType)
}

func (c *ApiClient) AddParams(api string, params interface{}) (string, error) {
	var extraParams map[string]string
	v, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	json.Unmarshal([]byte(v), &extraParams)

	url_parm := url.Values{}
	for k, v := range extraParams {
		url_parm.Set(k, v)
	}

	api_with_params := api + "?" + url_parm.Encode()

	return api_with_params, nil
}

func (c *ApiClient) Get(api string) (resp *http.Response, err error) {
	req, err := c.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	return c.SendRequest(req, "")
}

func (c *ApiClient) GetJson(api string, response interface{}) (err error) {
	req, err := c.NewRequest("GET", api, nil)
	if err != nil {
		return err
	}
	return c.SendRequestJson(req, response)
}

func (c *ApiClient) GetWithParams(api string, params interface{}) (resp *http.Response, err error) {
	api_with_params, err := c.AddParams(api, params)
	if err != nil {
		return nil, err
	}
	return c.Get(api_with_params)
}

func (c *ApiClient) GetJsonWithParams(api string, params interface{}, response interface{}) (err error) {
	api_with_params, err := c.AddParams(api, params)
	if err != nil {
		return err
	}
	return c.GetJson(api_with_params, response)
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

func (c *ApiClient) SendRequest(req *http.Request, Content_type string) (*http.Response, error) {
	if c.Cfg != nil && c.Cfg.Url == "" {
		return nil, errors.New("URL not set")
	}

	if Content_type != "" {
		req.Header.Set("Content-Type", fmt.Sprintf(CONTENT_TYPE_FORMAT, Content_type))
	}

	if c.SetAuthorization(req) {
		c.Log.Debug("Api client - Authorization")
	} else if c.SetApiKey(req) {
		c.Log.Debug("Api client - Api key auth")
	} else if c.SetBasicAuth(req) {
		c.Log.Debug("Api client - Basic auth")
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("api client - response, url: %s, status_code: %d, Text: %s", req.URL, res.StatusCode, http.StatusText(res.StatusCode))
	}

	c.Log.Debugf("Api client - Post URL: %s,  Status: %d, Text: %s", req.URL, res.StatusCode, http.StatusText(res.StatusCode))

	return res, nil
}

func (c *ApiClient) MultipartWriter(api string, filename string, params map[string]string, w io.Writer) (*multipart.Writer, io.Writer, error) {
	multipart_writer := multipart.NewWriter(w)

	for key, val := range params {
		_ = multipart_writer.WriteField(key, val)
	}

	partHeaders := textproto.MIMEHeader{}
	partHeaders.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(api), escapeQuotes(filename)))
	partHeaders.Set("Content-Type", "application/octet-stream")
	writer, err := multipart_writer.CreatePart(partHeaders)
	return multipart_writer, writer, err
}

func (c *ApiClient) Close() {
	c.HTTPClient.CloseIdleConnections()
}
