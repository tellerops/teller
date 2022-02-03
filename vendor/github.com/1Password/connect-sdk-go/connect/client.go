package connect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaegerClientConfig "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"

	"github.com/1Password/connect-sdk-go/onepassword"
)

const (
	defaultUserAgent = "connect-sdk-go/%s"
)

// Client Represents an available 1Password Connect API to connect to
type Client interface {
	GetVaults() ([]onepassword.Vault, error)
	GetVault(uuid string) (*onepassword.Vault, error)
	GetVaultsByTitle(uuid string) ([]onepassword.Vault, error)
	GetItem(uuid string, vaultUUID string) (*onepassword.Item, error)
	GetItems(vaultUUID string) ([]onepassword.Item, error)
	GetItemsByTitle(title string, vaultUUID string) ([]onepassword.Item, error)
	GetItemByTitle(title string, vaultUUID string) (*onepassword.Item, error)
	CreateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error)
	UpdateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error)
	DeleteItem(item *onepassword.Item, vaultUUID string) error
	GetFile(fileUUID string, itemUUID string, vaultUUID string) (*onepassword.File, error)
	GetFileContent(file *onepassword.File) ([]byte, error)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

const (
	envHostVariable  = "OP_CONNECT_HOST"
	envTokenVariable = "OP_CONNECT_TOKEN"
)

// NewClientFromEnvironment Returns a Secret Service client assuming that your
// jwt is set in the OP_TOKEN environment variable
func NewClientFromEnvironment() (Client, error) {
	host, found := os.LookupEnv(envHostVariable)
	if !found {
		return nil, fmt.Errorf("There is no hostname available in the %q variable", envHostVariable)
	}

	token, found := os.LookupEnv(envTokenVariable)
	if !found {
		return nil, fmt.Errorf("There is no token available in the %q variable", envTokenVariable)
	}

	return NewClient(host, token), nil
}

// NewClient Returns a Secret Service client for a given url and jwt
func NewClient(url string, token string) Client {
	return NewClientWithUserAgent(url, token, fmt.Sprintf(defaultUserAgent, SDKVersion))
}

// NewClientWithUserAgent Returns a Secret Service client for a given url and jwt and identifies with userAgent
func NewClientWithUserAgent(url string, token string, userAgent string) Client {
	if !opentracing.IsGlobalTracerRegistered() {
		cfg := jaegerClientConfig.Configuration{}
		zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
		cfg.InitGlobalTracer(
			userAgent,
			jaegerClientConfig.Injector(opentracing.HTTPHeaders, zipkinPropagator),
			jaegerClientConfig.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
			jaegerClientConfig.ZipkinSharedRPCSpan(true),
		)
	}

	return &restClient{
		URL:   url,
		Token: token,

		userAgent: userAgent,
		tracer:    opentracing.GlobalTracer(),

		client: http.DefaultClient,
	}
}

type restClient struct {
	URL       string
	Token     string
	userAgent string
	tracer    opentracing.Tracer
	client    httpClient
}

// GetVaults Get a list of all available vaults
func (rs *restClient) GetVaults() ([]onepassword.Vault, error) {
	span := rs.tracer.StartSpan("GetVaults")
	defer span.Finish()

	vaultURL := fmt.Sprintf("/v1/vaults")
	request, err := rs.buildRequest(http.MethodGet, vaultURL, http.NoBody, span)
	if err != nil {
		return nil, err
	}

	response, err := rs.client.Do(request)
	if err != nil {
		return nil, err
	}

	var vaults []onepassword.Vault
	if err := parseResponse(response, http.StatusOK, &vaults); err != nil {
		return nil, err
	}

	return vaults, nil
}

// GetVaults Get a list of all available vaults
func (rs *restClient) GetVault(uuid string) (*onepassword.Vault, error) {
	if uuid == "" {
		return nil, errors.New("no uuid provided")
	}

	span := rs.tracer.StartSpan("GetVault")
	defer span.Finish()

	vaultURL := fmt.Sprintf("/v1/vaults/%s", uuid)
	request, err := rs.buildRequest(http.MethodGet, vaultURL, http.NoBody, span)
	if err != nil {
		return nil, err
	}

	response, err := rs.client.Do(request)
	if err != nil {
		return nil, err
	}
	var vault onepassword.Vault
	if err := parseResponse(response, http.StatusOK, &vault); err != nil {
		return nil, err
	}

	return &vault, nil
}

func (rs *restClient) GetVaultsByTitle(title string) ([]onepassword.Vault, error) {
	span := rs.tracer.StartSpan("GetVaultsByTitle")
	defer span.Finish()

	filter := url.QueryEscape(fmt.Sprintf("title eq \"%s\"", title))
	itemURL := fmt.Sprintf("/v1/vaults?filter=%s", filter)
	request, err := rs.buildRequest(http.MethodGet, itemURL, http.NoBody, span)
	if err != nil {
		return nil, err
	}

	response, err := rs.client.Do(request)
	if err != nil {
		return nil, err
	}

	var vaults []onepassword.Vault
	if err := parseResponse(response, http.StatusOK, &vaults); err != nil {
		return nil, err
	}

	return vaults, nil
}

// GetItem Get a specific Item from the 1Password Connect API
func (rs *restClient) GetItem(uuid string, vaultUUID string) (*onepassword.Item, error) {
	span := rs.tracer.StartSpan("GetItem")
	defer span.Finish()

	itemURL := fmt.Sprintf("/v1/vaults/%s/items/%s", vaultUUID, uuid)
	request, err := rs.buildRequest(http.MethodGet, itemURL, http.NoBody, span)
	if err != nil {
		return nil, err
	}

	response, err := rs.client.Do(request)
	if err != nil {
		return nil, err
	}
	var item onepassword.Item
	if err := parseResponse(response, http.StatusOK, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

func (rs *restClient) GetItemByTitle(title string, vaultUUID string) (*onepassword.Item, error) {
	span := rs.tracer.StartSpan("GetItemByTitle")
	defer span.Finish()
	items, err := rs.GetItemsByTitle(title, vaultUUID)
	if err != nil {
		return nil, err
	}

	if len(items) != 1 {
		return nil, fmt.Errorf("Found %d item(s) in vault %q with title %q", len(items), vaultUUID, title)
	}

	return rs.GetItem(items[0].ID, items[0].Vault.ID)
}

func (rs *restClient) GetItemsByTitle(title string, vaultUUID string) ([]onepassword.Item, error) {
	span := rs.tracer.StartSpan("GetItemsByTitle")
	defer span.Finish()

	filter := url.QueryEscape(fmt.Sprintf("title eq \"%s\"", title))
	itemURL := fmt.Sprintf("/v1/vaults/%s/items?filter=%s", vaultUUID, filter)
	request, err := rs.buildRequest(http.MethodGet, itemURL, http.NoBody, span)
	if err != nil {
		return nil, err
	}

	response, err := rs.client.Do(request)
	if err != nil {
		return nil, err
	}

	var items []onepassword.Item
	if err := parseResponse(response, http.StatusOK, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (rs *restClient) GetItems(vaultUUID string) ([]onepassword.Item, error) {
	span := rs.tracer.StartSpan("GetItems")
	defer span.Finish()

	itemURL := fmt.Sprintf("/v1/vaults/%s/items", vaultUUID)
	request, err := rs.buildRequest(http.MethodGet, itemURL, http.NoBody, span)
	if err != nil {
		return nil, err
	}

	response, err := rs.client.Do(request)
	if err != nil {
		return nil, err
	}

	var items []onepassword.Item
	if err := parseResponse(response, http.StatusOK, &items); err != nil {
		return nil, err
	}

	return items, nil
}

// CreateItem Create a new item in a specified vault
func (rs *restClient) CreateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	span := rs.tracer.StartSpan("CreateItem")
	defer span.Finish()

	itemURL := fmt.Sprintf("/v1/vaults/%s/items", vaultUUID)
	itemBody, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	request, err := rs.buildRequest(http.MethodPost, itemURL, bytes.NewBuffer(itemBody), span)
	if err != nil {
		return nil, err
	}

	response, err := rs.client.Do(request)
	if err != nil {
		return nil, err
	}

	var newItem onepassword.Item
	if err := parseResponse(response, http.StatusOK, &newItem); err != nil {
		return nil, err
	}

	return &newItem, nil
}

// UpdateItem Update a new item in a specified vault
func (rs *restClient) UpdateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	span := rs.tracer.StartSpan("UpdateItem")
	defer span.Finish()

	itemURL := fmt.Sprintf("/v1/vaults/%s/items/%s", item.Vault.ID, item.ID)
	itemBody, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	request, err := rs.buildRequest(http.MethodPut, itemURL, bytes.NewBuffer(itemBody), span)
	if err != nil {
		return nil, err
	}

	response, err := rs.client.Do(request)
	if err != nil {
		return nil, err
	}

	var newItem onepassword.Item
	if err := parseResponse(response, http.StatusOK, &newItem); err != nil {
		return nil, err
	}

	return &newItem, nil
}

// DeleteItem Delete a new item in a specified vault
func (rs *restClient) DeleteItem(item *onepassword.Item, vaultUUID string) error {
	span := rs.tracer.StartSpan("DeleteItem")
	defer span.Finish()

	itemURL := fmt.Sprintf("/v1/vaults/%s/items/%s", item.Vault.ID, item.ID)
	request, err := rs.buildRequest(http.MethodDelete, itemURL, http.NoBody, span)
	if err != nil {
		return err
	}

	response, err := rs.client.Do(request)
	if err != nil {
		return err
	}

	if err := parseResponse(response, http.StatusNoContent, nil); err != nil {
		return err
	}

	return nil
}

// GetFile Get a specific File in a specified item.
// This does not include the file contents. Call GetFileContent() to load the file's content.
func (rs *restClient) GetFile(uuid string, itemUUID string, vaultUUID string) (*onepassword.File, error) {
	span := rs.tracer.StartSpan("GetFile")
	defer span.Finish()

	itemURL := fmt.Sprintf("/v1/vaults/%s/items/%s/files/%s", vaultUUID, itemUUID, uuid)
	request, err := rs.buildRequest(http.MethodGet, itemURL, http.NoBody, span)
	if err != nil {
		return nil, err
	}

	response, err := rs.client.Do(request)
	if err != nil {
		return nil, err
	}
	if err := expectMinimumConnectVersion(response, version{1, 3, 0}); err != nil {
		return nil, err
	}

	var file onepassword.File
	if err := parseResponse(response, http.StatusOK, &file); err != nil {
		return nil, err
	}

	return &file, nil
}

// GetFileContent retrieves the file's content.
// If the file's content have previously been fetched, those contents are returned without making another request.
func (rs *restClient) GetFileContent(file *onepassword.File) ([]byte, error) {
	if content, err := file.Content(); err == nil {
		return content, nil
	}

	span := rs.tracer.StartSpan("GetFileContent")
	defer span.Finish()

	request, err := rs.buildRequest(http.MethodGet, file.ContentPath, http.NoBody, span)
	if err != nil {
		return nil, err
	}

	response, err := rs.client.Do(request)
	if err != nil {
		return nil, err
	}
	if err := expectMinimumConnectVersion(response, version{1, 3, 0}); err != nil {
		return nil, err
	}

	content, err := readResponseBody(response, http.StatusOK)
	if err != nil {
		return nil, err
	}

	file.SetContent(content)
	return content, nil
}

func (rs *restClient) buildRequest(method string, path string, body io.Reader, span opentracing.Span) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", rs.URL, path)

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rs.Token))
	request.Header.Set("User-Agent", rs.userAgent)

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, path)
	ext.HTTPMethod.Set(span, method)

	rs.tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(request.Header))

	return request, nil
}

func parseResponse(resp *http.Response, expectedStatusCode int, result interface{}) error {
	body, err := readResponseBody(resp, expectedStatusCode)
	if err != nil {
		return err
	}
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("decoding response: %s", err)
		}
	}
	return nil
}

func readResponseBody(resp *http.Response, expectedStatusCode int) ([]byte, error) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != expectedStatusCode {
		var errResp *onepassword.Error
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("decoding error response: %s", err)
		}
		return nil, errResp
	}
	return body, nil
}
