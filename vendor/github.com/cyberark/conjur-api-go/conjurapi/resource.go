package conjurapi

import (
	"encoding/json"
	"fmt"

	"github.com/cyberark/conjur-api-go/conjurapi/response"
)

type ResourceFilter struct {
	Kind   string
	Search string
	Limit  int
	Offset int
}

// CheckPermission determines whether the authenticated user has a specified privilege
// on a resource.
func (c *Client) CheckPermission(resourceID, privilege string) (bool, error) {
	req, err := c.router.CheckPermissionRequest(resourceID, privilege)
	if err != nil {
		return false, err
	}

	resp, err := c.SubmitRequest(req)
	if err != nil {
		return false, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, nil
	} else if resp.StatusCode == 404 || resp.StatusCode == 403 {
		return false, nil
	} else {
		return false, fmt.Errorf("Permission check failed with HTTP status %d", resp.StatusCode)
	}
}

// Resource fetches a single user-visible resource by id.
func (c *Client) Resource(resourceID string) (resource map[string]interface{}, err error) {
	req, err := c.router.ResourceRequest(resourceID)
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

	resource = make(map[string]interface{})
	err = json.Unmarshal(data, &resource)
	return
}

// Resources fetches user-visible resources. The set of resources can
// be limited by the given ResourceFilter. If filter is non-nil, only
// non-zero-valued members of the filter will be applied.
func (c *Client) Resources(filter *ResourceFilter) (resources []map[string]interface{}, err error) {
	req, err := c.router.ResourcesRequest(filter)
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

	resources = make([]map[string]interface{}, 1)

	err = json.Unmarshal(data, &resources)

	return
}
