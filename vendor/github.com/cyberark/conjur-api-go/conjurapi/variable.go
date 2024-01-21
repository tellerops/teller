package conjurapi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/cyberark/conjur-api-go/conjurapi/response"
)

// RetrieveBatchSecrets fetches values for all variables in a slice using a
// single API call
//
// The authenticated user must have execute privilege on all variables.
func (c *Client) RetrieveBatchSecrets(variableIDs []string) (map[string][]byte, error) {
	jsonResponse, err := c.retrieveBatchSecrets(variableIDs, false)
	if err != nil {
		return nil, err
	}

	resolvedVariables := map[string][]byte{}
	for id, value := range jsonResponse {
		resolvedVariables[id] = []byte(value)
	}

	return resolvedVariables, nil
}

// RetrieveBatchSecretsSafe fetches values for all variables in a slice using a
// single API call. This version of the method will automatically base64-encode
// the secrets on the server side allowing the retrieval of binary values in
// batch requests. Secrets are NOT base64 encoded in the returned map.
//
// The authenticated user must have execute privilege on all variables.
func (c *Client) RetrieveBatchSecretsSafe(variableIDs []string) (map[string][]byte, error) {
	jsonResponse, err := c.retrieveBatchSecrets(variableIDs, true)
	if err != nil {
		return nil, err
	}

	resolvedVariables := map[string][]byte{}
	var decodedValue []byte
	for id, value := range jsonResponse {
		decodedValue, err = base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}
		resolvedVariables[id] = decodedValue
	}

	return resolvedVariables, nil
}

// RetrieveSecret fetches a secret from a variable.
//
// The authenticated user must have execute privilege on the variable.
func (c *Client) RetrieveSecret(variableID string) ([]byte, error) {
	resp, err := c.retrieveSecret(variableID)
	if err != nil {
		return nil, err
	}

	return response.DataResponse(resp)
}

// RetrieveSecretReader fetches a secret from a variable and returns it as a
// data stream.
//
// The authenticated user must have execute privilege on the variable.
func (c *Client) RetrieveSecretReader(variableID string) (io.ReadCloser, error) {
	resp, err := c.retrieveSecret(variableID)
	if err != nil {
		return nil, err
	}

	return response.SecretDataResponse(resp)
}

// RetrieveSecretWithVersion fetches a specific version of a secret from a
// variable.
//
// The authenticated user must have execute privilege on the variable.
func (c *Client) RetrieveSecretWithVersion(variableID string, version int) ([]byte, error) {
	resp, err := c.retrieveSecretWithVersion(variableID, version)
	if err != nil {
		return nil, err
	}

	return response.DataResponse(resp)
}

// RetrieveSecretWithVersionReader fetches a specific version of a secret from a
// variable and returns it as a data stream.
//
// The authenticated user must have execute privilege on the variable.
func (c *Client) RetrieveSecretWithVersionReader(variableID string, version int) (io.ReadCloser, error) {
	resp, err := c.retrieveSecretWithVersion(variableID, version)
	if err != nil {
		return nil, err
	}

	return response.SecretDataResponse(resp)
}

func (c *Client) retrieveBatchSecrets(variableIDs []string, base64Flag bool) (map[string]string, error) {
	req, err := c.RetrieveBatchSecretsRequest(variableIDs, base64Flag)
	if err != nil {
		return nil, err
	}

	resp, err := c.SubmitRequest(req)
	if err != nil {
		return nil, err
	}

	data, err := response.DataResponse(resp)
	if err != nil {
		return nil, err
	}

	if base64Flag && resp.Header.Get("Content-Encoding") != "base64" {
		return nil, errors.New(
			"Conjur response is not Base64-encoded. " +
			"The Conjur version may not be compatible with this function - " +
			"try using RetrieveBatchSecrets instead." )
	}

	jsonResponse := map[string]string{}
	err = json.Unmarshal(data, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return jsonResponse, nil
}

func (c *Client) retrieveSecret(variableID string) (*http.Response, error) {
	req, err := c.RetrieveSecretRequest(variableID)
	if err != nil {
		return nil, err
	}

	return c.SubmitRequest(req)
}

func (c *Client) retrieveSecretWithVersion(variableID string, version int) (*http.Response, error) {
	req, err := c.RetrieveSecretWithVersionRequest(variableID, version)
	if err != nil {
		return nil, err
	}

	return c.SubmitRequest(req)
}

// AddSecret adds a secret value to a variable.
//
// The authenticated user must have update privilege on the variable.
func (c *Client) AddSecret(variableID string, secretValue string) error {
	req, err := c.AddSecretRequest(variableID, secretValue)
	if err != nil {
		return err
	}

	resp, err := c.SubmitRequest(req)
	if err != nil {
		return err
	}

	return response.EmptyResponse(resp)
}
