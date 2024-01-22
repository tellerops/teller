package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaegerClientConfig "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"

	"github.com/1Password/connect-sdk-go/onepassword"
)

const (
	defaultUserAgent = "connect-sdk-go/%s"
)

var (
	vaultUUIDError = fmt.Errorf("malformed vault uuid provided")
	itemUUIDError  = fmt.Errorf("malformed item uuid provided")
	fileUUIDError  = fmt.Errorf("malformed file uuid provided")
)

// Client Represents an available 1Password Connect API to connect to
type Client interface {
	GetVaults() ([]onepassword.Vault, error)
	GetVault(uuid string) (*onepassword.Vault, error)
	GetVaultByUUID(uuid string) (*onepassword.Vault, error)
	GetVaultByTitle(title string) (*onepassword.Vault, error)
	GetVaultsByTitle(uuid string) ([]onepassword.Vault, error)
	GetItems(vaultQuery string) ([]onepassword.Item, error)
	GetItem(itemQuery, vaultQuery string) (*onepassword.Item, error)
	GetItemByUUID(uuid string, vaultQuery string) (*onepassword.Item, error)
	GetItemByTitle(title string, vaultQuery string) (*onepassword.Item, error)
	GetItemsByTitle(title string, vaultQuery string) ([]onepassword.Item, error)
	CreateItem(item *onepassword.Item, vaultQuery string) (*onepassword.Item, error)
	UpdateItem(item *onepassword.Item, vaultQuery string) (*onepassword.Item, error)
	DeleteItem(item *onepassword.Item, vaultQuery string) error
	DeleteItemByID(itemUUID string, vaultQuery string) error
	DeleteItemByTitle(title string, vaultQuery string) error
	GetFiles(itemQuery string, vaultQuery string) ([]onepassword.File, error)
	GetFile(uuid string, itemQuery string, vaultQuery string) (*onepassword.File, error)
	GetFileContent(file *onepassword.File) ([]byte, error)
	DownloadFile(file *onepassword.File, targetDirectory string, overwrite bool) (string, error)
	LoadStructFromItemByUUID(config interface{}, itemUUID string, vaultQuery string) error
	LoadStructFromItemByTitle(config interface{}, itemTitle string, vaultQuery string) error
	LoadStructFromItem(config interface{}, itemQuery string, vaultQuery string) error
	LoadStruct(config interface{}) error
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

// GetVault Get a vault based on its name or ID
func (rs *restClient) GetVault(vaultQuery string) (*onepassword.Vault, error) {
	span := rs.tracer.StartSpan("GetVault")
	defer span.Finish()

	if vaultQuery == "" {
		return nil, fmt.Errorf("Please provide either the vault name or its ID.")
	}
	if !isValidUUID(vaultQuery) {
		return rs.GetVaultByTitle(vaultQuery)
	}
	return rs.GetVaultByUUID(vaultQuery)
}

func (rs *restClient) GetVaultByUUID(uuid string) (*onepassword.Vault, error) {
	if !isValidUUID(uuid) {
		return nil, vaultUUIDError
	}

	span := rs.tracer.StartSpan("GetVaultByUUID")
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

func (rs *restClient) GetVaultByTitle(vaultName string) (*onepassword.Vault, error) {
	span := rs.tracer.StartSpan("GetVaultByTitle")
	defer span.Finish()

	vaults, err := rs.GetVaultsByTitle(vaultName)
	if err != nil {
		return nil, err
	}

	if len(vaults) != 1 {
		return nil, fmt.Errorf("Found %d vaults with title %q", len(vaults), vaultName)
	}

	return &vaults[0], nil
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

func (rs *restClient) getVaultUUID(vaultQuery string) (string, error) {
	if vaultQuery == "" {
		return "", fmt.Errorf("Please provide either the vault name or its ID.")
	}
	if isValidUUID(vaultQuery) {
		return vaultQuery, nil
	}
	vault, err := rs.GetVaultByTitle(vaultQuery)
	if err != nil {
		return "", err
	}
	return vault.ID, nil
}

// GetItem Get a specific Item from the 1Password Connect API by either title or UUID
func (rs *restClient) GetItem(itemQuery string, vaultQuery string) (*onepassword.Item, error) {
	span := rs.tracer.StartSpan("GetItem")
	defer span.Finish()

	if itemQuery == "" {
		return nil, fmt.Errorf("Please provide either the item name or its ID.")
	}

	if isValidUUID(itemQuery) {
		item, err := rs.GetItemByUUID(itemQuery, vaultQuery)
		if item != nil {
			return item, err
		}
	}
	return rs.GetItemByTitle(itemQuery, vaultQuery)
}

// GetItemByUUID Get a specific Item from the 1Password Connect API by its UUID
func (rs *restClient) GetItemByUUID(uuid string, vaultQuery string) (*onepassword.Item, error) {
	if !isValidUUID(uuid) {
		return nil, itemUUIDError
	}

	vaultUUID, err := rs.getVaultUUID(vaultQuery)
	if err != nil {
		return nil, err
	}

	span := rs.tracer.StartSpan("GetItemByUUID")
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

func (rs *restClient) GetItemByTitle(title string, vaultQuery string) (*onepassword.Item, error) {
	vaultUUID, err := rs.getVaultUUID(vaultQuery)
	if err != nil {
		return nil, err
	}

	span := rs.tracer.StartSpan("GetItemByTitle")
	defer span.Finish()
	items, err := rs.GetItemsByTitle(title, vaultUUID)
	if err != nil {
		return nil, err
	}

	if len(items) != 1 {
		return nil, fmt.Errorf("Found %d item(s) in vault %q with title %q", len(items), vaultUUID, title)
	}

	return &items[0], nil
}

func (rs *restClient) GetItemsByTitle(title string, vaultQuery string) ([]onepassword.Item, error) {
	vaultUUID, err := rs.getVaultUUID(vaultQuery)
	if err != nil {
		return nil, err
	}

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

	var itemSummaries []onepassword.Item
	if err := parseResponse(response, http.StatusOK, &itemSummaries); err != nil {
		return nil, err
	}

	items := make([]onepassword.Item, len(itemSummaries))
	for i, itemSummary := range itemSummaries {
		tempItem, err := rs.GetItem(itemSummary.ID, itemSummary.Vault.ID)
		if err != nil {
			return nil, err
		}
		items[i] = *tempItem
	}

	return items, nil
}

func (rs *restClient) GetItems(vaultQuery string) ([]onepassword.Item, error) {
	vaultUUID, err := rs.getVaultUUID(vaultQuery)
	if err != nil {
		return nil, err
	}

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

func (rs *restClient) getItemUUID(itemQuery, vaultQuery string) (string, error) {
	if itemQuery == "" {
		return "", fmt.Errorf("Please provide either the item name or its ID.")
	}
	if isValidUUID(itemQuery) {
		return itemQuery, nil
	}
	item, err := rs.GetItemByTitle(itemQuery, vaultQuery)
	if err != nil {
		return "", err
	}
	return item.ID, nil
}

// CreateItem Create a new item in a specified vault
func (rs *restClient) CreateItem(item *onepassword.Item, vaultQuery string) (*onepassword.Item, error) {
	vaultUUID, err := rs.getVaultUUID(vaultQuery)
	if err != nil {
		return nil, err
	}

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

// DeleteItemByID Delete a new item in a specified vault, specifying the item's uuid
func (rs *restClient) DeleteItemByID(itemUUID string, vaultQuery string) error {
	if !isValidUUID(itemUUID) {
		return itemUUIDError
	}
	vaultUUID, err := rs.getVaultUUID(vaultQuery)
	if err != nil {
		return err
	}

	span := rs.tracer.StartSpan("DeleteItemByID")
	defer span.Finish()

	itemURL := fmt.Sprintf("/v1/vaults/%s/items/%s", vaultUUID, itemUUID)
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

// DeleteItemByTitle Delete a new item in a specified vault, specifying the item's title
func (rs *restClient) DeleteItemByTitle(title string, vaultQuery string) error {
	span := rs.tracer.StartSpan("DeleteItemByTitle")
	defer span.Finish()

	item, err := rs.GetItemByTitle(title, vaultQuery)
	if err != nil {
		return err
	}

	return rs.DeleteItem(item, item.Vault.ID)
}

func (rs *restClient) GetFiles(itemQuery string, vaultQuery string) ([]onepassword.File, error) {
	vaultUUID, err := rs.getVaultUUID(vaultQuery)
	if err != nil {
		return nil, err
	}
	itemUUID, err := rs.getItemUUID(itemQuery, vaultQuery)
	if err != nil {
		return nil, err
	}

	span := rs.tracer.StartSpan("GetFiles")
	defer span.Finish()

	jsonURL := fmt.Sprintf("/v1/vaults/%s/items/%s/files", vaultUUID, itemUUID)
	request, err := rs.buildRequest(http.MethodGet, jsonURL, http.NoBody, span)
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
	var files []onepassword.File
	if err := parseResponse(response, http.StatusOK, &files); err != nil {
		return nil, err
	}

	return files, nil
}

// GetFile Get a specific File in a specified item.
// This does not include the file contents. Call GetFileContent() to load the file's content.
func (rs *restClient) GetFile(uuid string, itemQuery string, vaultQuery string) (*onepassword.File, error) {
	if !isValidUUID(uuid) {
		return nil, fileUUIDError
	}
	vaultUUID, err := rs.getVaultUUID(vaultQuery)
	if err != nil {
		return nil, err
	}
	itemUUID, err := rs.getItemUUID(itemQuery, vaultQuery)
	if err != nil {
		return nil, err
	}

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
	response, err := rs.retrieveDocumentContent(file)
	if err != nil {
		return nil, err
	}
	content, err := readResponseBody(response, http.StatusOK)
	if err != nil {
		return nil, err
	}
	file.SetContent(content)
	return content, nil
}

func (rs *restClient) DownloadFile(file *onepassword.File, targetDirectory string, overwriteIfExists bool) (string, error) {
	response, err := rs.retrieveDocumentContent(file)
	if err != nil {
		return "", err
	}

	path := filepath.Join(targetDirectory, filepath.Base(file.Name))

	var osFile *os.File

	if overwriteIfExists {
		osFile, err = createFile(path)
		if err != nil {
			return "", err
		}
	} else {
		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			osFile, err = createFile(path)
			if err != nil {
				return "", err
			}
		} else {
			return "", fmt.Errorf("a file already exists under the %s path. In order to overwrite it, set `overwriteIfExists` to true", path)
		}
	}
	defer osFile.Close()
	if _, err = io.Copy(osFile, response.Body); err != nil {
		return "", err
	}

	return path, nil
}

func (rs *restClient) retrieveDocumentContent(file *onepassword.File) (*http.Response, error) {
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
	return response, nil
}

func createFile(path string) (*os.File, error) {
	osFile, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	err = os.Chmod(path, 0600)
	if err != nil {
		return nil, err
	}
	return osFile, nil
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

func loadToStruct(item *parsedItem, config reflect.Value) error {
	t := config.Type()
	for i := 0; i < t.NumField(); i++ {
		value := config.Field(i)
		field := t.Field(i)

		if !value.CanSet() {
			return fmt.Errorf("cannot load config into private fields")
		}

		item.fields = append(item.fields, &field)
		item.values = append(item.values, &value)
	}
	return nil
}

// LoadStructFromItem Load configuration values based on struct tag from one 1P item.
// It accepts as parameters item title/UUID and vault title/UUID.
func (rs *restClient) LoadStructFromItem(i interface{}, itemQuery string, vaultQuery string) error {
	if itemQuery == "" {
		return fmt.Errorf("Please provide either the item name or its ID.")
	}
	if isValidUUID(itemQuery) {
		return rs.LoadStructFromItemByUUID(i, itemQuery, vaultQuery)
	}
	return rs.LoadStructFromItemByTitle(i, itemQuery, vaultQuery)
}

// LoadStructFromItemByUUID Load configuration values based on struct tag from one 1P item.
func (rs *restClient) LoadStructFromItemByUUID(i interface{}, itemUUID string, vaultQuery string) error {
	vaultUUID, err := rs.getVaultUUID(vaultQuery)
	if err != nil {
		return err
	}
	if !isValidUUID(itemUUID) {
		return itemUUIDError
	}
	config, err := checkStruct(i)
	if err != nil {
		return err
	}
	item := parsedItem{}
	item.itemUUID = itemUUID
	item.vaultUUID = vaultUUID

	if err := loadToStruct(&item, config); err != nil {
		return err
	}
	if err := setValuesForTag(rs, &item, false); err != nil {
		return err
	}

	return nil
}

// LoadStructFromItemByTitle Load configuration values based on struct tag from one 1P item
func (rs *restClient) LoadStructFromItemByTitle(i interface{}, itemTitle string, vaultQuery string) error {
	vaultUUID, err := rs.getVaultUUID(vaultQuery)
	if err != nil {
		return err
	}

	config, err := checkStruct(i)
	if err != nil {
		return err
	}
	item := parsedItem{}
	item.itemTitle = itemTitle
	item.vaultUUID = vaultUUID

	if err := loadToStruct(&item, config); err != nil {
		return err
	}
	if err := setValuesForTag(rs, &item, true); err != nil {
		return err
	}

	return nil
}

// LoadStruct Load configuration values based on struct tag
func (rs *restClient) LoadStruct(i interface{}) error {
	config, err := checkStruct(i)
	if err != nil {
		return err
	}

	t := config.Type()

	// Multiple fields may be from a single item so we will collect them
	items := map[string]parsedItem{}

	// Fetch the Vault from the environment
	vaultUUID, envVarFound := os.LookupEnv(envVaultVar)

	for i := 0; i < t.NumField(); i++ {
		value := config.Field(i)
		field := t.Field(i)
		tag := field.Tag.Get(itemTag)

		if tag == "" {
			continue
		}

		if !value.CanSet() {
			return fmt.Errorf("Cannot load config into private fields")
		}

		itemVault, err := vaultUUIDForField(&field, vaultUUID, envVarFound)
		if err != nil {
			return err
		}
		if !isValidUUID(itemVault) {
			return vaultUUIDError
		}

		key := fmt.Sprintf("%s/%s", itemVault, tag)
		parsed := items[key]
		parsed.vaultUUID = itemVault
		parsed.itemTitle = tag
		parsed.fields = append(parsed.fields, &field)
		parsed.values = append(parsed.values, &value)
		items[key] = parsed
	}

	for _, item := range items {
		if err := setValuesForTag(rs, &item, true); err != nil {
			return err
		}
	}

	return nil
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
		var errResp onepassword.Error
		if json.Valid(body) {
			if err := json.Unmarshal(body, &errResp); err != nil {
				return nil, fmt.Errorf("decoding error response: %s", err)
			}
		} else {
			errResp.StatusCode = resp.StatusCode
			errResp.Message = http.StatusText(resp.StatusCode)
		}
		return nil, &errResp
	}
	return body, nil
}

func isValidUUID(u string) bool {
	r := regexp.MustCompile("^[a-z0-9]{26}$")
	return r.MatchString(u)
}
