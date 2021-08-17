package conjurapi

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/sirupsen/logrus"
)

type RouterV4 struct {
	Config *Config
}

func (r RouterV4) AuthenticateRequest(loginPair authn.LoginPair) (*http.Request, error) {
	authenticateURL := fmt.Sprintf("%s/authn/users/%s/authenticate", r.Config.ApplianceURL, url.QueryEscape(loginPair.Login))

	req, err := http.NewRequest("POST", authenticateURL, strings.NewReader(loginPair.APIKey))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")

	return req, nil
}

func (r RouterV4) RotateAPIKeyRequest(roleID string) (*http.Request, error) {
	account, kind, id, err := parseID(roleID)
	if err != nil {
		return nil, err
	}
	if account != r.Config.Account {
		return nil, fmt.Errorf("Account of '%s' must match the configured account '%s'", roleID, r.Config.Account)
	}

	var username string
	switch kind {
	case "user":
		username = id
	default:
		username = strings.Join([]string{kind, id}, "/")
	}

	rotateURL := fmt.Sprintf("%s/authn/users/api_key?id=%s", r.Config.ApplianceURL, url.QueryEscape(username))

	return http.NewRequest(
		"PUT",
		rotateURL,
		nil,
	)
}

func (r RouterV4) LoadPolicyRequest(mode PolicyMode, policyID string, policy io.Reader) (*http.Request, error) {
	return nil, fmt.Errorf("LoadPolicy is not supported for Conjur V4")
}

func (r RouterV4) ResourceRequest(resourceID string) (*http.Request, error) {
	logrus.Panic("ResourceRequest not implemented yet")
	return nil, nil
}

func (r RouterV4) ResourcesRequest(filter *ResourceFilter) (*http.Request, error) {
	logrus.Panic("ResourcesRequest not implemented yet")
	return nil, nil
}

func (r RouterV4) CheckPermissionRequest(resourceID, privilege string) (*http.Request, error) {
	account, kind, id, err := parseID(resourceID)
	if err != nil {
		return nil, err
	}

	checkURL := fmt.Sprintf("%s/authz/%s/resources/%s/%s?check=true&privilege=%s", r.Config.ApplianceURL, account, kind, url.QueryEscape(id), url.QueryEscape(privilege))

	return http.NewRequest(
		"GET",
		checkURL,
		nil,
	)
}

func (r RouterV4) AddSecretRequest(variableID, secretValue string) (*http.Request, error) {
	return nil, fmt.Errorf("AddSecret is not supported for Conjur V4")
}

func (r RouterV4) RetrieveBatchSecretsRequest(variableIDs []string, base64Flag bool) (*http.Request, error) {
	if base64Flag {
		return nil, fmt.Errorf("Batch retrieving Base64-encoded secrets is not supported in Conjur V4")
	}

	return http.NewRequest(
		"GET",
		r.batchVariableURL(variableIDs),
		nil,
	)
}

func (r RouterV4) RetrieveSecretRequest(variableID string) (*http.Request, error) {
	fullVariableID := makeFullId(r.Config.Account, "variable", variableID)

	variableURL, err := r.variableURL(fullVariableID)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(
		"GET",
		variableURL,
		nil,
	)
}

func (r RouterV4) variableURL(variableID string) (string, error) {
	_, _, id, err := parseID(variableID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/variables/%s/value", r.Config.ApplianceURL, url.PathEscape(id)), nil
}

func (r RouterV4) batchVariableURL(variableIDs []string) string {
	queryString := url.QueryEscape(strings.Join(variableIDs, ","))
	return fmt.Sprintf("%s/variables/values?vars=%s", r.Config.ApplianceURL, queryString)
}
