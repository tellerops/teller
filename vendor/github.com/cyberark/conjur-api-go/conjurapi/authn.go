package conjurapi

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-api-go/conjurapi/logging"
	"github.com/cyberark/conjur-api-go/conjurapi/response"
)

// OidcProvider contains information about an OIDC provider.
type OidcProvider struct {
	ServiceID    string `json:"service_id"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Nonce        string `json:"nonce"`
	CodeVerifier string `json:"code_verifier"`
	RedirectURI  string `json:"redirect_uri"`
}

func (c *Client) RefreshToken() (err error) {
	// Fetch cached conjur access token if using OIDC
	if c.GetConfig().AuthnType == "oidc" {
		token := c.readCachedAccessToken()
		if token != nil {
			c.authToken = token
		}
	}

	if c.NeedsTokenRefresh() {
		return c.refreshToken()
	}

	return nil
}

func (c *Client) ForceRefreshToken() error {
	return c.refreshToken()
}

func (c *Client) refreshToken() error {
	var tokenBytes []byte
	tokenBytes, err := c.authenticator.RefreshToken()
	if err != nil {
		return err
	}

	token, err := authn.NewToken(tokenBytes)
	if err != nil {
		return err
	}

	token.FromJSON(tokenBytes)
	c.authToken = token
	return nil
}

func (c *Client) NeedsTokenRefresh() bool {
	return c.authToken == nil ||
		c.authToken.ShouldRefresh() ||
		c.authenticator.NeedsTokenRefresh()
}

func (c *Client) readCachedAccessToken() *authn.AuthnToken {
	tokenBytes, err := c.storage.ReadAuthnToken()
	if err != nil {
		return nil
	}

	token, err := authn.NewToken(tokenBytes)
	if err != nil {
		return nil
	}

	token.FromJSON(token.Raw())
	return token
}

func (c *Client) createAuthRequest(req *http.Request) error {
	if err := c.RefreshToken(); err != nil {
		return err
	}

	req.Header.Set(
		"Authorization",
		fmt.Sprintf("Token token=\"%s\"", base64.StdEncoding.EncodeToString(c.authToken.Raw())),
	)

	return nil
}

func (c *Client) ChangeUserPassword(username string, password string, newPassword string) ([]byte, error) {
	req, err := c.ChangeUserPasswordRequest(username, password, newPassword)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return response.DataResponse(res)
}

func (c *Client) ChangeCurrentUserPassword(newPassword string) ([]byte, error) {
	username, password, err := c.storage.ReadCredentials()
	if err != nil {
		return nil, err
	}

	return c.ChangeUserPassword(username, password, newPassword)
}

// Login exchanges a user's password for an API key.
func (c *Client) Login(login string, password string) ([]byte, error) {
	req, err := c.LoginRequest(login, password)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	apiKey, err := response.DataResponse(res)
	if err != nil {
		return nil, err
	}

	// Store the API key in the credentials store
	if c.storage != nil {
		err = c.storage.StoreCredentials(login, string(apiKey))
	}
	return apiKey, err
}

// PurgeCredentials purges credentials from the client's credential storage.
func (c *Client) PurgeCredentials() error {
	if c.storage == nil {
		return nil
	}

	return c.storage.PurgeCredentials()
}

// PurgeCredentials purges credentials from the credential storage indicated by the
// configuration.
func PurgeCredentials(config Config) error {
	storage, err := createStorageProvider(config)
	if err != nil {
		return err
	}

	if storage == nil {
		logging.ApiLog.Debugf("Not storing credentials, so nothing to purge")
		return nil
	}

	return storage.PurgeCredentials()
}

// Authenticate obtains a new access token using the internal authenticator.
func (c *Client) InternalAuthenticate() ([]byte, error) {
	if c.authenticator == nil {
		return nil, errors.New("unable to authenticate using client without authenticator")
	}

	// If using OIDC, check if we have a cached access token
	if c.GetConfig().AuthnType == "oidc" {
		token := c.readCachedAccessToken()
		if token != nil && !token.ShouldRefresh() {
			return token.Raw(), nil
		} else {
			// We can't simply refresh the token because it'll require user input. Instead,
			// we return an error and inform the client/user to login again.
			return nil, errors.New("No valid OIDC token found. Please login again.")
		}
	}

	// Otherwise refresh the token
	return c.authenticator.RefreshToken()
}

// WhoAmI obtains information on the current user.
func (c *Client) WhoAmI() ([]byte, error) {
	req, err := c.WhoAmIRequest()
	if err != nil {
		return nil, err
	}

	res, err := c.SubmitRequest(req)
	if err != nil {
		return nil, err
	}

	return response.DataResponse(res)
}

// Authenticate obtains a new access token.
func (c *Client) Authenticate(loginPair authn.LoginPair) ([]byte, error) {
	resp, err := c.authenticate(loginPair)
	if err != nil {
		return nil, err
	}

	return response.DataResponse(resp)
}

// AuthenticateReader obtains a new access token and returns it as a data stream.
func (c *Client) AuthenticateReader(loginPair authn.LoginPair) (io.ReadCloser, error) {
	resp, err := c.authenticate(loginPair)
	if err != nil {
		return nil, err
	}

	return response.SecretDataResponse(resp)
}

func (c *Client) authenticate(loginPair authn.LoginPair) (*http.Response, error) {
	req, err := c.AuthenticateRequest(loginPair)
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}

func (c *Client) OidcAuthenticate(code, nonce, code_verifier string) ([]byte, error) {
	req, err := c.OidcAuthenticateRequest(code, nonce, code_verifier)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	resp, err := response.DataResponse(res)

	if err == nil && c.storage != nil {
		c.storage.StoreAuthnToken(resp)
	}

	return resp, err
}

func (c *Client) ListOidcProviders() ([]OidcProvider, error) {
	req, err := c.ListOidcProvidersRequest()
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	providers := []OidcProvider{}
	err = response.JSONResponse(resp, &providers)

	return providers, err
}

// RotateAPIKey replaces the API key of a role on the server with a new
// random secret. Given that a fully-qualified resource id resembles
// '<account>:<kind>:<identifier>', argument roleID must be at least partially-qualified.
//
// The authenticated user must have update privilege on the role.
func (c *Client) RotateAPIKey(roleID string) ([]byte, error) {
	resp, err := c.rotateAPIKey(roleID)
	if err != nil {
		return nil, err
	}

	return response.DataResponse(resp)
}

func (c *Client) RotateCurrentUserAPIKey() ([]byte, error) {
	username, password, err := c.storage.ReadCredentials()
	if err != nil {
		return nil, err
	}

	resp, err := c.rotateCurrentUserAPIKey(username, password)
	if err != nil {
		return nil, err
	}

	return response.DataResponse(resp)
}

// RotateUserAPIKey constructs a role ID from a given user ID then replaces the
// API key of the role with a new random secret. Given that a fully-qualified
// resource ID resembles '<account>:<kind>:<identifier>', argument userID will
// be accepted as either fully- or partially-qualified, but the provided role
// must be a user.
//
// The authenticated user must have update privilege on the role.
func (c *Client) RotateUserAPIKey(userID string) ([]byte, error) {
	return c.rotateApiKeyAndEnforceKind(userID, "user")
}

// RotateHostAPIKey constructs a role ID from a given host ID then replaces the
// API key of the role with a new random secret. Given that a fully-qualified
// resource ID resembles '<account>:<kind>:<identifier>', argument hostID will
// be accepted as either fully- or partially-qualified, but the provided role
// must be a host.
//
// The authenticated user must have update privilege on the role.
func (c *Client) RotateHostAPIKey(hostID string) ([]byte, error) {
	return c.rotateApiKeyAndEnforceKind(hostID, "host")
}

func (c *Client) rotateApiKeyAndEnforceKind(roleID, kind string) ([]byte, error) {
	account, kind, identifier, err := c.parseIDandEnforceKind(roleID, kind)
	if err != nil {
		return nil, err
	}

	roleID = fmt.Sprintf("%s:%s:%s", account, kind, identifier)
	return c.RotateAPIKey(roleID)
}

// RotateAPIKeyReader replaces the API key of a role on the server with a new
// random secret and returns it as a data stream.
//
// The authenticated user must have update privilege on the role.
func (c *Client) RotateAPIKeyReader(roleID string) (io.ReadCloser, error) {
	resp, err := c.rotateAPIKey(roleID)
	if err != nil {
		return nil, err
	}

	return response.SecretDataResponse(resp)
}

func (c *Client) rotateAPIKey(roleID string) (*http.Response, error) {
	req, err := c.RotateAPIKeyRequest(roleID)
	if err != nil {
		return nil, err
	}

	return c.SubmitRequest(req)
}

func (c *Client) rotateCurrentUserAPIKey(username string, password string) (*http.Response, error) {
	req, err := c.RotateCurrentUserAPIKeyRequest(username, password)
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}

func (c *Client) PublicKeys(kind string, identifier string) ([]byte, error) {
	req, err := c.PublicKeysRequest(kind, identifier)
	if err != nil {
		return nil, err
	}

	res, err := c.SubmitRequest(req)
	if err != nil {
		return nil, err
	}

	return response.DataResponse(res)
}
