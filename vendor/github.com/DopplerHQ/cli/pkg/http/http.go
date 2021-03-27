/*
Copyright © 2019 Doppler <support@doppler.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DopplerHQ/cli/pkg/utils"
	"github.com/DopplerHQ/cli/pkg/version"
)

type queryParam struct {
	Key   string
	Value string
}

type errorResponse struct {
	Messages []string
	Success  bool
}

// GetRequest perform HTTP GET
func GetRequest(host string, verifyTLS bool, headers map[string]string, uri string, params []queryParam) (int, http.Header, []byte, error) {
	url := fmt.Sprintf("%s%s", host, uri)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	statusCode, respHeaders, body, err := performRequest(req, verifyTLS, params)
	if err != nil {
		return statusCode, respHeaders, body, err
	}

	return statusCode, respHeaders, body, nil
}

// PostRequest perform HTTP POST
func PostRequest(host string, verifyTLS bool, headers map[string]string, uri string, params []queryParam, body []byte) (int, http.Header, []byte, error) {
	url := fmt.Sprintf("%s%s", host, uri)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	statusCode, respHeaders, body, err := performRequest(req, verifyTLS, params)
	if err != nil {
		return statusCode, respHeaders, body, err
	}

	return statusCode, respHeaders, body, nil
}

// DeleteRequest perform HTTP DELETE
func DeleteRequest(host string, verifyTLS bool, headers map[string]string, uri string, params []queryParam) (int, http.Header, []byte, error) {
	url := fmt.Sprintf("%s%s", host, uri)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return 0, nil, nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	statusCode, respHeaders, body, err := performRequest(req, verifyTLS, params)
	if err != nil {
		return statusCode, respHeaders, body, err
	}

	return statusCode, respHeaders, body, nil
}

func performRequest(req *http.Request, verifyTLS bool, params []queryParam) (int, http.Header, []byte, error) {
	// set headers
	req.Header.Set("client-sdk", "go-cli")
	req.Header.Set("client-version", version.ProgramVersion)
	req.Header.Set("user-agent", "doppler-go-cli-"+version.ProgramVersion)
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
	req.Header.Set("Content-Type", "application/json")

	// set url query parameters
	query := req.URL.Query()
	for _, param := range params {
		query.Add(param.Key, param.Value)
	}
	req.URL.RawQuery = query.Encode()

	// close the connection after reading the response, to help prevent socket exhaustion
	req.Close = true

	client := &http.Client{}
	// set http timeout
	if UseTimeout {
		client.Timeout = TimeoutDuration
	}

	transport := &http.Transport{
		// disable keep alives to prevent multiple CLI instances from exhausting the
		// OS's available network sockets. this adds a negligible performance penalty
		DisableKeepAlives: true,
	}
	// set TLS config
	// #nosec G402
	if !verifyTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client.Transport = transport

	startTime := time.Now()
	var response *http.Response
	response = nil

	requestErr := retry(5, 100*time.Millisecond, func() error {
		resp, err := client.Do(req)
		if err != nil {
			if resp != nil {
				defer resp.Body.Close()
			}

			utils.LogDebug(err.Error())

			if isTimeout(err) {
				// retry request
				return err
			}

			return StopRetry{err}
		}

		response = resp

		utils.LogDebug(fmt.Sprintf("Performing HTTP %s to %s", req.Method, req.URL))
		if requestID := resp.Header.Get("x-request-id"); requestID != "" {
			utils.LogDebug(fmt.Sprintf("Request ID %s", requestID))
		}

		if isSuccess(resp.StatusCode) {
			return nil
		}

		contentType := resp.Header.Get("content-type")
		if isRetry(resp.StatusCode, contentType) {
			// start logging retries after 10 seconds so it doesn't feel like we've frozen
			// we subtract 1 millisecond so that we always win the race against a request that exhausts its full 10 second time out
			if time.Now().After(startTime.Add(10 * time.Second).Add(-1 * time.Millisecond)) {
				utils.Log(fmt.Sprintf("Request failed with HTTP %d, retrying", resp.StatusCode))
			}
			return errors.New("Request failed")
		}

		// we cannot recover from this error code; accept defeat
		return StopRetry{errors.New("Request failed")}
	})

	if response != nil {
		defer response.Body.Close()
	}

	if requestErr != nil && response == nil {
		return 0, nil, nil, requestErr
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, nil, nil, err
	}

	headers := response.Header.Clone()

	// success
	if requestErr == nil {
		return response.StatusCode, headers, body, nil
	}

	// print the response body error messages
	if contentType := response.Header.Get("content-type"); strings.HasPrefix(contentType, "application/json") {
		var errResponse errorResponse
		err = json.Unmarshal(body, &errResponse)
		if err != nil {
			utils.LogDebug(fmt.Sprintf("Unable to parse response body: \n%s", string(body)))
			return response.StatusCode, headers, nil, err
		}

		return response.StatusCode, headers, body, errors.New(strings.Join(errResponse.Messages, "\n"))
	}

	return response.StatusCode, headers, nil, fmt.Errorf("Request failed with HTTP %d", response.StatusCode)
}

func isSuccess(statusCode int) bool {
	return (statusCode >= 200 && statusCode <= 299) || (statusCode >= 300 && statusCode <= 399)
}

func isRetry(statusCode int, contentType string) bool {
	return (statusCode == 429) ||
		(statusCode >= 100 && statusCode <= 199) ||
		// don't retry 5xx errors w/ a JSON body
		(statusCode >= 500 && statusCode <= 599 && !strings.HasPrefix(contentType, "application/json"))
}

func isTimeout(err error) bool {
	if urlErr, ok := err.(*url.Error); ok {
		if netErr, ok := urlErr.Err.(net.Error); ok {
			return netErr.Timeout()
		}
	}

	return false
}
