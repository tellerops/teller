package conjurapi

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-api-go/conjurapi/logging"
)

type Authenticator interface {
	RefreshToken() ([]byte, error)
	NeedsTokenRefresh() bool
}

type Client struct {
	config        Config
	authToken     authn.AuthnToken
	httpClient    *http.Client
	authenticator Authenticator
	router        Router
}

type Router interface {
	AddSecretRequest(variableID, secretValue string) (*http.Request, error)
	AuthenticateRequest(loginPair authn.LoginPair) (*http.Request, error)
	CheckPermissionRequest(resourceID, privilege string) (*http.Request, error)
	LoadPolicyRequest(mode PolicyMode, policyID string, policy io.Reader) (*http.Request, error)
	ResourceRequest(resourceID string) (*http.Request, error)
	ResourcesRequest(filter *ResourceFilter) (*http.Request, error)
	RetrieveBatchSecretsRequest(variableIDs []string, base64Flag bool) (*http.Request, error)
	RetrieveSecretRequest(variableID string) (*http.Request, error)
	RotateAPIKeyRequest(roleID string) (*http.Request, error)
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

// ReadResponseBody fully reads a response and closes it.
func ReadResponseBody(response io.ReadCloser) ([]byte, error) {
	defer response.Close()
	return ioutil.ReadAll(response)
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

func LoginPairFromNetRC(config Config) (*authn.LoginPair, error) {
	if config.NetRCPath == "" {
		config.NetRCPath = os.ExpandEnv("$HOME/.netrc")
	}

	rc, err := netrc.ParseFile(config.NetRCPath)
	if err != nil {
		return nil, err
	}

	m := rc.FindMachine(config.ApplianceURL + "/authn")

	if m == nil {
		return nil, fmt.Errorf("No credentials found in NetRCPath")
	}

	return &authn.LoginPair{Login: m.Login, APIKey: m.Password}, nil
}

func NewClientFromEnvironment(config Config) (*Client, error) {
	err := config.validate()

	if err != nil {
		return nil, err
	}

	authnTokenFile := os.Getenv("CONJUR_AUTHN_TOKEN_FILE")
	if authnTokenFile != "" {
		return NewClientFromTokenFile(config, authnTokenFile)
	}

	loginPair, err := LoginPairFromEnv()
	if err == nil && loginPair.Login != "" && loginPair.APIKey != "" {
		return NewClientFromKey(config, *loginPair)
	}

	loginPair, err = LoginPairFromNetRC(config)
	if err == nil && loginPair.Login != "" && loginPair.APIKey != "" {
		return NewClientFromKey(config, *loginPair)
	}

	return nil, fmt.Errorf("Environment variables and machine identity files satisfying at least one authentication strategy must be present!")
}

func (c *Client) GetHttpClient() (*http.Client) {
	return c.httpClient
}

func (c *Client) SetHttpClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

func (c *Client) GetConfig() (Config) {
	return c.config
}

func (c *Client) SubmitRequest(req *http.Request) (resp *http.Response, err error) {
	err = c.createAuthRequest(req)
	if err != nil {
		return
	}

	logging.ApiLog.Debugf("req: %+v\n", req)
	resp, err = c.httpClient.Do(req)
	if err != nil {
		return
	}

	return
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

func parseID(fullID string) (account, kind, id string, err error) {
	tokens := strings.SplitN(fullID, ":", 3)
	if len(tokens) != 3 {
		err = fmt.Errorf("Id '%s' must be fully qualified", fullID)
		return
	}
	return tokens[0], tokens[1], tokens[2], nil
}

func newClientWithAuthenticator(config Config, authenticator Authenticator) (*Client, error) {
	var (
		err error
	)

	err = config.validate()

	if err != nil {
		return nil, err
	}

	var httpClient *http.Client
	var router Router

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

	if config.V4 {
		router = RouterV4{&config}
	} else {
		router = RouterV5{&config}
	}

	return &Client{
		config:        config,
		authenticator: authenticator,
		httpClient:    httpClient,
		router:        router,
	}, nil
}

func newHTTPSClient(cert []byte) (*http.Client, error) {
	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(cert)
	if !ok {
		return nil, fmt.Errorf("Can't append Conjur SSL cert")
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: pool},
	}
	return &http.Client{Transport: tr, Timeout: time.Second * 10}, nil
}
