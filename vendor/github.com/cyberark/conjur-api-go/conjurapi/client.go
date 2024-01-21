package conjurapi

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-api-go/conjurapi/logging"
	"github.com/cyberark/conjur-api-go/conjurapi/response"
)

type Authenticator interface {
	RefreshToken() ([]byte, error)
	NeedsTokenRefresh() bool
}

type CredentialStorageProvider interface {
	StoreCredentials(login string, password string) error
	ReadCredentials() (login string, password string, err error)
	ReadAuthnToken() ([]byte, error)
	StoreAuthnToken(token []byte) error
	PurgeCredentials() error
}

type Client struct {
	config        Config
	authToken     *authn.AuthnToken
	httpClient    *http.Client
	authenticator Authenticator
	storage       CredentialStorageProvider
}

func NewClientFromKey(config Config, loginPair authn.LoginPair) (*Client, error) {
	authenticator := &authn.APIKeyAuthenticator{
		LoginPair: loginPair,
	}
	client, err := newClientWithAuthenticator(
		config,
		authenticator,
	)
	authenticator.Authenticate = client.Authenticate
	return client, err
}

func NewClientFromOidcCode(config Config, code, nonce, code_verifier string) (*Client, error) {
	authenticator := &authn.OidcAuthenticator{
		Code:         code,
		Nonce:        nonce,
		CodeVerifier: code_verifier,
	}
	client, err := newClientWithAuthenticator(
		config,
		authenticator,
	)
	if err == nil {
		authenticator.Authenticate = client.OidcAuthenticate
	}
	return client, err
}

// ReadResponseBody fully reads a response and closes it.
func ReadResponseBody(response io.ReadCloser) ([]byte, error) {
	defer response.Close()
	return io.ReadAll(response)
}

func NewClientFromToken(config Config, token string) (*Client, error) {
	return newClientWithAuthenticator(
		config,
		&authn.TokenAuthenticator{token},
	)
}

func NewClientFromTokenFile(config Config, tokenFile string) (*Client, error) {
	return newClientWithAuthenticator(
		config,
		&authn.TokenFileAuthenticator{
			TokenFile:   tokenFile,
			MaxWaitTime: -1,
		},
	)
}

func LoginPairFromEnv() (*authn.LoginPair, error) {
	return &authn.LoginPair{
		Login:  os.Getenv("CONJUR_AUTHN_LOGIN"),
		APIKey: os.Getenv("CONJUR_AUTHN_API_KEY"),
	}, nil
}

// TODO: Create a version of this function for creating an authenticator from environment
func NewClientFromEnvironment(config Config) (*Client, error) {
	err := config.Validate()

	if err != nil {
		return nil, err
	}

	authnTokenFile := os.Getenv("CONJUR_AUTHN_TOKEN_FILE")
	if authnTokenFile != "" {
		return NewClientFromTokenFile(config, authnTokenFile)
	}

	authnToken := os.Getenv("CONJUR_AUTHN_TOKEN")
	if authnToken != "" {
		return NewClientFromToken(config, authnToken)
	}

	authnJwtServiceID := os.Getenv("CONJUR_AUTHN_JWT_SERVICE_ID")
	if authnJwtServiceID != "" {
		return NewClientFromJwt(config, authnJwtServiceID)
	}

	loginPair, err := LoginPairFromEnv()
	if err == nil && loginPair.Login != "" && loginPair.APIKey != "" {
		return NewClientFromKey(config, *loginPair)
	}

	return newClientFromStoredCredentials(config)
}

func NewClientFromJwt(config Config, authnJwtServiceID string) (*Client, error) {
	var jwtTokenString string
	jwtToken := os.Getenv("CONJUR_AUTHN_JWT_TOKEN")
	jwtTokenString = fmt.Sprintf("jwt=%s", jwtToken)
	if jwtToken == "" {
		jwtTokenPath := os.Getenv("JWT_TOKEN_PATH")
		if jwtTokenPath == "" {
			jwtTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		}

		jwtToken, err := os.ReadFile(jwtTokenPath)
		if err != nil {
			return nil, err
		}
		jwtTokenString = fmt.Sprintf("jwt=%s", string(jwtToken))
	}

	var httpClient *http.Client
	if config.IsHttps() {
		cert, err := config.ReadSSLCert()
		if err != nil {
			return nil, err
		}
		httpClient, err = newHTTPSClient(cert)
		if err != nil {
			return nil, err
		}

	} else {
		httpClient = &http.Client{Timeout: time.Second * 10}
	}

	authnJwtHostID := os.Getenv("CONJUR_AUTHN_JWT_HOST_ID")
	var authnJwtUrl string
	if authnJwtHostID != "" {
		authnJwtUrl = makeRouterURL(config.ApplianceURL, "authn-jwt", authnJwtServiceID, config.Account, url.PathEscape(authnJwtHostID), "authenticate").String()
	} else {
		authnJwtUrl = makeRouterURL(config.ApplianceURL, "authn-jwt", authnJwtServiceID, config.Account, "authenticate").String()
	}

	req, err := http.NewRequest("POST", authnJwtUrl, strings.NewReader(jwtTokenString))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	token, err := response.DataResponse(resp)
	if err != nil {
		return nil, err
	}

	return NewClientFromToken(config, string(token))
}

func newClientFromStoredCredentials(config Config) (*Client, error) {
	if config.AuthnType == "oidc" {
		return newClientFromStoredOidcCredentials(config)
	}

	// Attempt to load credentials from whatever storage provider is configured
	if storageProvider, _ := createStorageProvider(config); storageProvider != nil {
		login, password, err := storageProvider.ReadCredentials()
		if err != nil {
			return nil, err
		}
		if login != "" && password != "" {
			return NewClientFromKey(config, authn.LoginPair{Login: login, APIKey: password})
		}
	}

	return nil, fmt.Errorf("No valid credentials found. Please login again.")
}

func newClientFromStoredOidcCredentials(config Config) (*Client, error) {
	client, err := NewClientFromOidcCode(config, "", "", "")
	if err != nil {
		return nil, err
	}
	token := client.readCachedAccessToken()
	if token != nil && !token.ShouldRefresh() {
		return client, nil
	}
	return nil, fmt.Errorf("No valid OIDC token found. Please login again.")
}

func (c *Client) GetAuthenticator() Authenticator {
	return c.authenticator
}

func (c *Client) SetAuthenticator(authenticator Authenticator) {
	c.authenticator = authenticator
}

func (c *Client) GetHttpClient() *http.Client {
	return c.httpClient
}

func (c *Client) SetHttpClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

func (c *Client) GetConfig() Config {
	return c.config
}

func (c *Client) SubmitRequest(req *http.Request) (resp *http.Response, err error) {
	err = c.createAuthRequest(req)
	if err != nil {
		return
	}

	return c.submitRequestWithCustomAuth(req)
}

func (c *Client) submitRequestWithCustomAuth(req *http.Request) (resp *http.Response, err error) {
	logging.ApiLog.Debugf("req: %+v\n", req)
	resp, err = c.httpClient.Do(req)
	if err != nil {
		return
	}

	return
}

func (c *Client) WhoAmIRequest() (*http.Request, error) {
	return http.NewRequest("GET", makeRouterURL(c.config.ApplianceURL, "whoami").String(), nil)
}

func (c *Client) LoginRequest(login string, password string) (*http.Request, error) {
	authenticateURL := makeRouterURL(c.authnURL(), "login").String()

	req, err := http.NewRequest("GET", authenticateURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(login, password)
	req.Header.Set("Content-Type", "text/plain")

	return req, nil
}

func (c *Client) AuthenticateRequest(loginPair authn.LoginPair) (*http.Request, error) {
	authenticateURL := makeRouterURL(c.authnURL(), url.QueryEscape(loginPair.Login), "authenticate").String()

	req, err := http.NewRequest("POST", authenticateURL, strings.NewReader(loginPair.APIKey))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")

	return req, nil
}

func (c *Client) ListOidcProvidersRequest() (*http.Request, error) {
	return http.NewRequest("GET", c.oidcProvidersUrl(), nil)
}

func (c *Client) OidcAuthenticateRequest(code, nonce, code_verifier string) (*http.Request, error) {
	authenticateURL := makeRouterURL(c.authnURL(), "authenticate").withFormattedQuery("code=%s&nonce=%s&code_verifier=%s", code, nonce, code_verifier).String()

	req, err := http.NewRequest("GET", authenticateURL, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// RotateAPIKeyRequest requires roleID argument to be at least partially-qualified
// ID of from [<account>:]<kind>:<identifier>.
func (c *Client) RotateAPIKeyRequest(roleID string) (*http.Request, error) {
	account, kind, identifier, err := c.parseID(roleID)
	if err != nil {
		return nil, err
	}
	roleID = fmt.Sprintf("%s:%s:%s", account, kind, identifier)

	rotateURL := makeRouterURL(c.authnURL(), "api_key").withFormattedQuery("role=%s", roleID).String()

	return http.NewRequest(
		"PUT",
		rotateURL,
		nil,
	)
}

func (c *Client) RotateCurrentUserAPIKeyRequest(login string, password string) (*http.Request, error) {
	rotateUrl := makeRouterURL(c.authnURL(), "api_key")

	req, err := http.NewRequest(
		"PUT",
		rotateUrl.String(),
		nil,
	)

	if err != nil {
		return nil, err
	}

	// API key can only be rotated via basic auth, NOT using bearer token
	req.SetBasicAuth(login, password)

	return req, nil
}

func (c *Client) ChangeUserPasswordRequest(username string, password string, newPassword string) (*http.Request, error) {
	passwordURL := makeRouterURL(c.config.ApplianceURL, "authn", c.config.Account, "password")

	req, err := http.NewRequest(
		"PUT",
		passwordURL.String(),
		strings.NewReader(newPassword),
	)

	if err != nil {
		return nil, err
	}

	// Password can only be updated via basic auth, NOT using bearer token
	req.SetBasicAuth(username, password)

	return req, nil
}

// CheckPermissionRequest crafts an HTTP request to Conjur's /resource endpoint
// to check if the authenticated user has the given privilege on the given resourceID.
func (c *Client) CheckPermissionRequest(resourceID, privilege string) (*http.Request, error) {
	account, kind, id, err := c.parseID(resourceID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("check=true&privilege=%s", url.QueryEscape(privilege))

	checkURL := makeRouterURL(c.resourcesURL(account), kind, url.QueryEscape(id)).withQuery(query).String()

	return http.NewRequest(
		"GET",
		checkURL,
		nil,
	)
}

// CheckPermissionForRoleRequest crafts an HTTP request to Conjur's /resource endpoint
// to check if a given role has the given privilege on the given resourceID.
func (c *Client) CheckPermissionForRoleRequest(resourceID, roleID, privilege string) (*http.Request, error) {
	account, kind, id, err := c.parseID(resourceID)
	if err != nil {
		return nil, err
	}

	roleAccount, roleKind, roleIdentifier, err := c.parseID(roleID)
	if err != nil {
		return nil, err
	}
	fullyQualifiedRoleID := strings.Join([]string{roleAccount, roleKind, roleIdentifier}, ":")

	query := fmt.Sprintf("check=true&privilege=%s&role=%s", url.QueryEscape(privilege), url.QueryEscape(fullyQualifiedRoleID))

	checkURL := makeRouterURL(c.resourcesURL(account), kind, url.QueryEscape(id)).withQuery(query).String()

	return http.NewRequest(
		"GET",
		checkURL,
		nil,
	)
}

func (c *Client) ResourceRequest(resourceID string) (*http.Request, error) {
	account, kind, id, err := c.parseID(resourceID)
	if err != nil {
		return nil, err
	}

	requestURL := makeRouterURL(c.resourcesURL(account), kind, url.QueryEscape(id))

	return http.NewRequest(
		"GET",
		requestURL.String(),
		nil,
	)
}

func (c *Client) ResourcesRequest(filter *ResourceFilter) (*http.Request, error) {
	query := url.Values{}

	if filter != nil {
		if filter.Kind != "" {
			query.Add("kind", filter.Kind)
		}
		if filter.Search != "" {
			query.Add("search", filter.Search)
		}

		if filter.Limit != 0 {
			query.Add("limit", strconv.Itoa(filter.Limit))
		}

		if filter.Offset != 0 {
			query.Add("offset", strconv.Itoa(filter.Offset))
		}

		if filter.Role != "" {
			query.Add("acting_as", filter.Role)
		}
	}

	requestURL := makeRouterURL(c.resourcesURL(c.config.Account)).withQuery(query.Encode())

	return http.NewRequest(
		"GET",
		requestURL.String(),
		nil,
	)
}

func (c *Client) PermittedRolesRequest(resourceID string, privilege string) (*http.Request, error) {
	account, kind, id, err := c.parseID(resourceID)
	if err != nil {
		return nil, err
	}
	permittedRolesURL := makeRouterURL(c.resourcesURL(account), kind, url.QueryEscape(id)).withFormattedQuery("permitted_roles=true&privilege=%s", url.QueryEscape(privilege)).String()

	return http.NewRequest(
		"GET",
		permittedRolesURL,
		nil,
	)
}

func (c *Client) RoleRequest(roleID string) (*http.Request, error) {
	account, kind, id, err := c.parseID(roleID)
	if err != nil {
		return nil, err
	}
	roleURL := makeRouterURL(c.rolesURL(account), kind, url.QueryEscape(id))

	return http.NewRequest(
		"GET",
		roleURL.String(),
		nil,
	)
}

func (c *Client) RoleMembersRequest(roleID string) (*http.Request, error) {
	account, kind, id, err := c.parseID(roleID)
	if err != nil {
		return nil, err
	}
	roleMembersURL := makeRouterURL(c.rolesURL(account), kind, url.QueryEscape(id)).withFormattedQuery("members")

	return http.NewRequest(
		"GET",
		roleMembersURL.String(),
		nil,
	)
}

func (c *Client) RoleMembershipsRequest(roleID string) (*http.Request, error) {
	account, kind, id, err := c.parseID(roleID)
	if err != nil {
		return nil, err
	}
	roleMembershipsURL := makeRouterURL(c.rolesURL(account), kind, url.QueryEscape(id)).withFormattedQuery("memberships")

	return http.NewRequest(
		"GET",
		roleMembershipsURL.String(),
		nil,
	)
}

func (c *Client) LoadPolicyRequest(mode PolicyMode, policyID string, policy io.Reader) (*http.Request, error) {
	fullPolicyID := makeFullId(c.config.Account, "policy", policyID)

	account, kind, id, err := c.parseID(fullPolicyID)
	if err != nil {
		return nil, err
	}
	policyURL := makeRouterURL(c.policiesURL(account), kind, url.QueryEscape(id)).String()

	var method string
	switch mode {
	case PolicyModePost:
		method = "POST"
	case PolicyModePatch:
		method = "PATCH"
	case PolicyModePut:
		method = "PUT"
	default:
		return nil, fmt.Errorf("Invalid PolicyMode: %d", mode)
	}

	return http.NewRequest(
		method,
		policyURL,
		policy,
	)
}

func (c *Client) RetrieveBatchSecretsRequest(variableIDs []string, base64Flag bool) (*http.Request, error) {
	fullVariableIDs := []string{}
	for _, variableID := range variableIDs {
		fullVariableID := makeFullId(c.config.Account, "variable", variableID)
		fullVariableIDs = append(fullVariableIDs, fullVariableID)
	}

	request, err := http.NewRequest(
		"GET",
		c.batchVariableURL(fullVariableIDs),
		nil,
	)

	if err != nil {
		return nil, err
	}

	if base64Flag {
		request.Header.Add("Accept-Encoding", "base64")
	}

	return request, nil
}

func (c *Client) RetrieveSecretRequest(variableID string) (*http.Request, error) {
	fullVariableID := makeFullId(c.config.Account, "variable", variableID)

	variableURL, err := c.variableURL(fullVariableID)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(
		"GET",
		variableURL,
		nil,
	)
}

func (c *Client) RetrieveSecretWithVersionRequest(variableID string, version int) (*http.Request, error) {
	fullVariableID := makeFullId(c.config.Account, "variable", variableID)

	variableURL, err := c.variableWithVersionURL(fullVariableID, version)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(
		"GET",
		variableURL,
		nil,
	)
}

func (c *Client) AddSecretRequest(variableID, secretValue string) (*http.Request, error) {
	fullVariableID := makeFullId(c.config.Account, "variable", variableID)

	variableURL, err := c.variableURL(fullVariableID)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(
		"POST",
		variableURL,
		strings.NewReader(secretValue),
	)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return request, nil
}

func (c *Client) CreateTokenRequest(body string) (*http.Request, error) {

	tokenURL := c.createTokenURL()
	request, err := http.NewRequest(
		"POST",
		tokenURL,
		strings.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return request, nil

}

func (c *Client) DeleteTokenRequest(token string) (*http.Request, error) {
	tokenURL := c.createTokenURL() + "/" + token

	request, err := http.NewRequest(
		"DELETE",
		tokenURL,
		nil,
	)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return request, nil
}

func (c *Client) CreateHostRequest(body string, token string) (*http.Request, error) {
	hostURL := c.createHostURL()
	request, err := http.NewRequest(
		"POST",
		hostURL,
		strings.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", fmt.Sprintf("Token token=\"%s\"", token))

	return request, nil
}

func (c *Client) PublicKeysRequest(kind string, identifier string) (*http.Request, error) {
	publicKeysURL := makeRouterURL(c.config.ApplianceURL, "public_keys", c.config.Account, kind, identifier)
	return http.NewRequest("GET", publicKeysURL.String(), nil)
}

func (c *Client) createTokenURL() string {
	return makeRouterURL(c.config.ApplianceURL, "host_factory_tokens").String()
}

func (c *Client) createHostURL() string {
	return makeRouterURL(c.config.ApplianceURL, "host_factories/hosts").String()
}

func (c *Client) variableURL(variableID string) (string, error) {
	account, kind, id, err := c.parseID(variableID)
	if err != nil {
		return "", err
	}
	return makeRouterURL(c.secretsURL(account), kind, url.PathEscape(id)).String(), nil
}

func (c *Client) variableWithVersionURL(variableID string, version int) (string, error) {
	account, kind, id, err := c.parseID(variableID)
	if err != nil {
		return "", err
	}
	return makeRouterURL(c.secretsURL(account), kind, url.PathEscape(id)).
		withFormattedQuery("version=%d", version).String(), nil
}

func (c *Client) batchVariableURL(variableIDs []string) string {
	queryString := url.QueryEscape(strings.Join(variableIDs, ","))
	return makeRouterURL(c.globalSecretsURL()).withFormattedQuery("variable_ids=%s", queryString).String()
}

func (c *Client) authnURL() string {
	if c.config.AuthnType != "" && c.config.AuthnType != "authn" {
		// If using an alternate authn service, such as authn-oidc, the URL will be
		// '/authn-<type>/<service-id>/<account>'
		authnType := fmt.Sprintf("authn-%s", c.config.AuthnType)
		return makeRouterURL(c.config.ApplianceURL, authnType, c.config.ServiceID, c.config.Account).String()
	}
	// For the default authn service, the URL will be '/authn/<account>'
	return makeRouterURL(c.config.ApplianceURL, "authn", c.config.Account).String()
}

func (c *Client) oidcProvidersUrl() string {
	return makeRouterURL(c.config.ApplianceURL, "authn-oidc", c.config.Account, "providers").String()
}

func (c *Client) resourcesURL(account string) string {
	return makeRouterURL(c.config.ApplianceURL, "resources", account).String()
}

func (c *Client) rolesURL(account string) string {
	return makeRouterURL(c.config.ApplianceURL, "roles", account).String()
}

func (c *Client) secretsURL(account string) string {
	return makeRouterURL(c.config.ApplianceURL, "secrets", account).String()
}

func (c *Client) globalSecretsURL() string {
	return makeRouterURL(c.config.ApplianceURL, "secrets").String()
}

func (c *Client) policiesURL(account string) string {
	return makeRouterURL(c.config.ApplianceURL, "policies", account).String()
}

func makeFullId(account, kind, id string) string {
	tokens := strings.SplitN(id, ":", 3)
	switch len(tokens) {
	case 1:
		tokens = []string{account, kind, tokens[0]}
	case 2:
		tokens = []string{account, tokens[0], tokens[1]}
	}
	return strings.Join(tokens, ":")
}

// parseID accepts as argument a resource ID and returns its components - account,
// resource kind, and identifier. The provided ID can either be fully- or
// partially-qualified. If the ID is only partially-qualified, the configured
// account will be returned.
//
// Examples:
// c.parseID("dev:user:alice")  =>  "dev", "user", "alice", nil
// c.parseID("user:alice")      =>  "dev", "user", "alice", nil
// c.parseID("prod:user:alice") => "prod", "user", "alice", nil
// c.parseID("malformed")       =>     "",     "",      "". error
func (c *Client) parseID(id string) (account, kind, identifier string, err error) {
	account, kind, identifier = c.unopinionatedParseID(id)
	if identifier == "" || kind == "" {
		return "", "", "", fmt.Errorf("Malformed ID '%s': must be fully- or partially-qualified, of form [<account>:]<kind>:<identifier>", id)
	}
	if account == "" {
		account = c.config.Account
	}
	return account, kind, identifier, nil
}

// parseIDandEnforceKind accepts as argument a resource ID and a kind, and returns
// the components - account, resource kind, and identifier - only if the provided
// resource matches the expected kind. If the ID is only partially-qualified, the
// configured account will be returned, and if the ID consists only of the
// identifier, the expected kind will be returned.
//
// Examples:
// c.parseID("dev:user:alice", "user")  =>  "dev", "user", "alice", nil
// c.parseID("user:alice", "user")      =>  "dev", "user", "alice", nil
// c.parseID("alice", "user")           =>  "dev", "user", "alice", nil
// c.parseID("prod:user:alice", "user") => "prod", "user", "alice", nil
// c.parseID("host:alice", "user")      =>     "",     "",      "", error
func (c *Client) parseIDandEnforceKind(id, enforcedKind string) (account, kind, identifier string, err error) {
	account, kind, identifier = c.unopinionatedParseID(id)
	if (identifier == "") || (kind != "" && kind != enforcedKind) {
		return "", "", "", fmt.Errorf("Malformed ID '%s', must represent a %s, of form [[<account>:]%s:]<identifier>", id, enforcedKind, enforcedKind)
	}
	if kind == "" {
		kind = enforcedKind
	}
	if account == "" {
		account = c.config.Account
	}
	return account, kind, identifier, nil
}

// unopinionatedParseID returns the components of the provided ID - account,
// resource kind, and identifier - without expectation on resource kind or
// account inclusion.
func (c *Client) unopinionatedParseID(id string) (account, kind, identifier string) {
	tokens := strings.SplitN(id, ":", 3)
	for len(tokens) < 3 {
		tokens = append([]string{""}, tokens...)
	}
	return tokens[0], tokens[1], tokens[2]
}

func NewClient(config Config) (*Client, error) {
	var err error

	err = config.Validate()

	if err != nil {
		return nil, err
	}

	httpClient, err := createHttpClient(config)
	if err != nil {
		return nil, err
	}

	storageProvider, err := createStorageProvider(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
		storage:    storageProvider,
	}, nil
}

func createHttpClient(config Config) (*http.Client, error) {
	var httpClient *http.Client

	if config.IsHttps() {
		cert, err := config.ReadSSLCert()
		if err != nil {
			return nil, err
		}
		httpClient, err = newHTTPSClient(cert)
		if err != nil {
			return nil, err
		}
	} else {
		httpClient = &http.Client{Timeout: time.Second * 10}
	}
	return httpClient, nil
}

func newClientWithAuthenticator(config Config, authenticator Authenticator) (*Client, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	client.authenticator = authenticator
	return client, nil
}

func newHTTPSClient(cert []byte) (*http.Client, error) {
	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(cert)
	if !ok {
		return nil, fmt.Errorf("Can't append Conjur SSL cert")
	}
	//TODO: Test what happens if this cert is expired
	//TODO: What if server cert is rotated
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: pool},
	}
	return &http.Client{Transport: tr, Timeout: time.Second * 10}, nil
}
