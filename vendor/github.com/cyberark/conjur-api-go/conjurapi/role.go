package conjurapi

import (
	"encoding/json"
	"fmt"

	"github.com/cyberark/conjur-api-go/conjurapi/response"
)

// RoleExists checks whether or not a role exists
func (c *Client) RoleExists(roleID string) (bool, error) {
	req, err := c.RoleRequest(roleID)
	if err != nil {
		return false, err
	}

	resp, err := c.SubmitRequest(req)
	if err != nil {
		return false, err
	}

	if (resp.StatusCode >= 200 && resp.StatusCode < 300) || resp.StatusCode == 403 {
		return true, nil
	} else if resp.StatusCode == 404 {
		return false, nil
	} else {
		return false, fmt.Errorf("Role exists check failed with HTTP status %d", resp.StatusCode)
	}
}

// Role fetches detailed information about a specific role, including
// the role members
func (c *Client) Role(roleID string) (role map[string]interface{}, err error) {
	req, err := c.RoleRequest(roleID)
	if err != nil {
		return
	}

	resp, err := c.SubmitRequest(req)
	if err != nil {
		return
	}

	data, err := response.DataResponse(resp)
	if err != nil {
		return
	}

	role = make(map[string]interface{})
	err = json.Unmarshal(data, &role)
	return
}

// RoleMembers fetches members within a role
func (c *Client) RoleMembers(roleID string) (members []map[string]interface{}, err error) {
	req, err := c.RoleMembersRequest(roleID)
	if err != nil {
		return
	}

	resp, err := c.SubmitRequest(req)
	if err != nil {
		return
	}

	data, err := response.DataResponse(resp)
	if err != nil {
		return
	}

	members = make([]map[string]interface{}, 0)
	err = json.Unmarshal(data, &members)
	return
}

// RoleMemberships fetches memberships of a role, including
// a list of groups of which a specific host or user is a member
func (c *Client) RoleMemberships(roleID string) (memberships []map[string]interface{}, err error) {
	req, err := c.RoleMembershipsRequest(roleID)
	if err != nil {
		return
	}

	resp, err := c.SubmitRequest(req)
	if err != nil {
		return
	}

	data, err := response.DataResponse(resp)
	if err != nil {
		return
	}

	memberships = make([]map[string]interface{}, 0)
	err = json.Unmarshal(data, &memberships)
	return
}
