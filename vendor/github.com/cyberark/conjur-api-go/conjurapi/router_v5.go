package conjurapi

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cyberark/conjur-api-go/conjurapi/authn"
)

type RouterV5 struct {
	Config *Config
}

func (r RouterV5) AuthenticateRequest(loginPair authn.LoginPair) (*http.Request, error) {
	authenticateURL := makeRouterURL(r.authnURL(), url.QueryEscape(loginPair.Login), "authenticate").String()

	req, err := http.NewRequest("POST", authenticateURL, strings.NewReader(loginPair.APIKey))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")

	return req, nil
}

func (r RouterV5) RotateAPIKeyRequest(roleID string) (*http.Request, error) {
	account, _, _, err := parseID(roleID)
	if err != nil {
		return nil, err
	}
	if account != r.Config.Account {
		return nil, fmt.Errorf("Account of '%s' must match the configured account '%s'", roleID, r.Config.Account)
	}

	rotateURL := makeRouterURL(r.authnURL(), "api_key").withFormattedQuery("role=%s", roleID).String()

	return http.NewRequest(
		"PUT",
		rotateURL,
		nil,
	)
}

func (r RouterV5) CheckPermissionRequest(resourceID, privilege string) (*http.Request, error) {
	account, kind, id, err := parseID(resourceID)
	if err != nil {
		return nil, err
	}
	checkURL := makeRouterURL(r.resourcesURL(account), kind, url.QueryEscape(id)).withFormattedQuery("check=true&privilege=%s", url.QueryEscape(privilege)).String()

	return http.NewRequest(
		"GET",
		checkURL,
		nil,
	)
}

func (r RouterV5) ResourceRequest(resourceID string) (*http.Request, error) {
	account, kind, id, err := parseID(resourceID)
	if err != nil {
		return nil, err
	}

	requestURL := makeRouterURL(r.resourcesURL(account), kind, url.QueryEscape(id))

	return http.NewRequest(
		"GET",
		requestURL.String(),
		nil,
	)
}

func (r RouterV5) ResourcesRequest(filter *ResourceFilter) (*http.Request, error) {
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
	}

	requestURL := makeRouterURL(r.resourcesURL(r.Config.Account)).withQuery(query.Encode())

	return http.NewRequest(
		"GET",
		requestURL.String(),
		nil,
	)
}

func (r RouterV5) LoadPolicyRequest(mode PolicyMode, policyID string, policy io.Reader) (*http.Request, error) {
	fullPolicyID := makeFullId(r.Config.Account, "policy", policyID)

	account, kind, id, err := parseID(fullPolicyID)
	if err != nil {
		return nil, err
	}
	policyURL := makeRouterURL(r.policiesURL(account), kind, url.QueryEscape(id)).String()

	var method string
	switch mode {
	case PolicyModePost:
		method = "POST"
	case PolicyModePatch:
		method = "PATCH"
	case PolicyModePut:
		method = "PUT"
	default:
		return nil, fmt.Errorf("Invalid PolicyMode : %d", mode)
	}

	return http.NewRequest(
		method,
		policyURL,
		policy,
	)
}

func (r RouterV5) RetrieveBatchSecretsRequest(variableIDs []string, base64Flag bool) (*http.Request, error) {
	fullVariableIDs := []string{}
	for _, variableID := range variableIDs {
		fullVariableID := makeFullId(r.Config.Account, "variable", variableID)
		fullVariableIDs = append(fullVariableIDs, fullVariableID)
	}

	request, err := http.NewRequest(
		"GET",
		r.batchVariableURL(fullVariableIDs),
		nil,
	)

	if err != nil {
		return nil, err
	}

	if base64Flag {
		request.Header.Add("Accept", "base64")
	}

	return request, nil
}

func (r RouterV5) RetrieveSecretRequest(variableID string) (*http.Request, error) {
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

func (r RouterV5) AddSecretRequest(variableID, secretValue string) (*http.Request, error) {
	fullVariableID := makeFullId(r.Config.Account, "variable", variableID)

	variableURL, err := r.variableURL(fullVariableID)
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

func (r RouterV5) variableURL(variableID string) (string, error) {
	account, kind, id, err := parseID(variableID)
	if err != nil {
		return "", err
	}
	return makeRouterURL(r.secretsURL(account), kind, url.PathEscape(id)).String(), nil
}

func (r RouterV5) batchVariableURL(variableIDs []string) string {
	queryString := url.QueryEscape(strings.Join(variableIDs, ","))
	return makeRouterURL(r.globalSecretsURL()).withFormattedQuery("variable_ids=%s", queryString).String()
}

func (r RouterV5) authnURL() string {
	return makeRouterURL(r.Config.ApplianceURL, "authn", r.Config.Account).String()
}

func (r RouterV5) resourcesURL(account string) string {
	return makeRouterURL(r.Config.ApplianceURL, "resources", account).String()
}

func (r RouterV5) secretsURL(account string) string {
	return makeRouterURL(r.Config.ApplianceURL, "secrets", account).String()
}

func (r RouterV5) globalSecretsURL() string {
	return makeRouterURL(r.Config.ApplianceURL, "secrets").String()
}

func (r RouterV5) policiesURL(account string) string {
	return makeRouterURL(r.Config.ApplianceURL, "policies", account).String()
}
