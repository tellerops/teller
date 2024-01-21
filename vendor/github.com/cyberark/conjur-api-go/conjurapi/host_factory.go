package conjurapi

import (
	"encoding/json"
	"fmt"
	"github.com/cyberark/conjur-api-go/conjurapi/response"
	"net/url"
	"time"
)

type HostFactoryTokenResponse struct {
	Expiration string   `json:"expiration"`
	Cidr       []string `json:"cidr"`
	Token      string   `json:"token"`
}

type HostFactoryHostResponse struct {
	CreatedAt    string   `json:"created_at"`
	Id           string   `json:"id"`
	Owner        string   `json:"owner"`
	Permissions  []string `json:"permissions"`
	Annotations  []string `json:"annotations"`
	RestrictedTo []string `json:"restricted_to"`
	ApiKey       string   `json:"api_key"`
}

func (c *Client) CreateToken(durationStr string, hostFactory string, cidrs []string, count int) ([]HostFactoryTokenResponse, error) {

	data := url.Values{}
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return nil, err
	}
	expiration := time.Now().Add(duration).Format(time.RFC3339)
	account, kind, identifier, err := c.parseIDandEnforceKind(hostFactory, "host_factory")
	if err != nil {
		return nil, err
	}
	hostFactory = fmt.Sprintf("%s:%s:%s", account, kind, identifier)
	data.Set("host_factory", hostFactory)
	data.Set("expiration", expiration)
	data.Set("count", fmt.Sprint(count))
	for _, cidr := range cidrs {
		data.Add("cidr[]", cidr)
	}
	return c.createToken(data)
}

func (c *Client) createToken(data url.Values) ([]HostFactoryTokenResponse, error) {

	encodedData := data.Encode()

	req, err := c.CreateTokenRequest(encodedData)
	if err != nil {
		return nil, err
	}

	resp, err := c.SubmitRequest(req)
	if err != nil {
		return nil, err
	}
	respData, err := response.DataResponse(resp)
	if err != nil {
		return nil, err
	}

	var jsonResponse []HostFactoryTokenResponse
	err = json.Unmarshal(respData, &jsonResponse)
	if err != nil {
		return nil, err
	}
	return jsonResponse, response.EmptyResponse(resp)
}

func (c *Client) DeleteToken(token string) error {

	req, err := c.DeleteTokenRequest(token)
	if err != nil {
		return err
	}

	resp, err := c.SubmitRequest(req)
	if err != nil {
		return err
	}
	return response.EmptyResponse(resp)
}

func (c *Client) CreateHost(id string, token string) (HostFactoryHostResponse, error) {
	data := url.Values{}
	data.Set("id", id)
	return c.createHost(data, token)
}

func (c *Client) createHost(data url.Values, token string) (HostFactoryHostResponse, error) {

	var jsonResponse HostFactoryHostResponse
	encodedData := data.Encode()
	req, err := c.CreateHostRequest(encodedData, token)
	if err != nil {
		return jsonResponse, err
	}

	resp, err := c.submitRequestWithCustomAuth(req)
	if err != nil {
		return jsonResponse, err
	}
	err = response.JSONResponse(resp, &jsonResponse)
	return jsonResponse, err
}
