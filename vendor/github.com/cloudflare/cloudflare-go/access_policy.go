package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

var (
	ErrMissingApplicationID = errors.New("missing required application ID")
)

type AccessApprovalGroup struct {
	EmailListUuid   string   `json:"email_list_uuid,omitempty"`
	EmailAddresses  []string `json:"email_addresses,omitempty"`
	ApprovalsNeeded int      `json:"approvals_needed,omitempty"`
}

// AccessPolicy defines a policy for allowing or disallowing access to
// one or more Access applications.
type AccessPolicy struct {
	ID         string     `json:"id,omitempty"`
	Precedence int        `json:"precedence"`
	Decision   string     `json:"decision"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
	Name       string     `json:"name"`

	IsolationRequired            *bool                 `json:"isolation_required,omitempty"`
	SessionDuration              *string               `json:"session_duration,omitempty"`
	PurposeJustificationRequired *bool                 `json:"purpose_justification_required,omitempty"`
	PurposeJustificationPrompt   *string               `json:"purpose_justification_prompt,omitempty"`
	ApprovalRequired             *bool                 `json:"approval_required,omitempty"`
	ApprovalGroups               []AccessApprovalGroup `json:"approval_groups"`

	// The include policy works like an OR logical operator. The user must
	// satisfy one of the rules.
	Include []interface{} `json:"include"`

	// The exclude policy works like a NOT logical operator. The user must
	// not satisfy all the rules in exclude.
	Exclude []interface{} `json:"exclude"`

	// The require policy works like a AND logical operator. The user must
	// satisfy all the rules in require.
	Require []interface{} `json:"require"`
}

// AccessPolicyListResponse represents the response from the list
// access policies endpoint.
type AccessPolicyListResponse struct {
	Result []AccessPolicy `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// AccessPolicyDetailResponse is the API response, containing a single
// access policy.
type AccessPolicyDetailResponse struct {
	Success  bool         `json:"success"`
	Errors   []string     `json:"errors"`
	Messages []string     `json:"messages"`
	Result   AccessPolicy `json:"result"`
}

type ListAccessPoliciesParams struct {
	ApplicationID string `json:"-"`
	ResultInfo
}

type GetAccessPolicyParams struct {
	ApplicationID string `json:"-"`
	PolicyID      string `json:"-"`
}

type CreateAccessPolicyParams struct {
	ApplicationID string `json:"-"`

	Precedence int    `json:"precedence"`
	Decision   string `json:"decision"`
	Name       string `json:"name"`

	IsolationRequired            *bool                 `json:"isolation_required,omitempty"`
	SessionDuration              *string               `json:"session_duration,omitempty"`
	PurposeJustificationRequired *bool                 `json:"purpose_justification_required,omitempty"`
	PurposeJustificationPrompt   *string               `json:"purpose_justification_prompt,omitempty"`
	ApprovalRequired             *bool                 `json:"approval_required,omitempty"`
	ApprovalGroups               []AccessApprovalGroup `json:"approval_groups"`

	// The include policy works like an OR logical operator. The user must
	// satisfy one of the rules.
	Include []interface{} `json:"include"`

	// The exclude policy works like a NOT logical operator. The user must
	// not satisfy all the rules in exclude.
	Exclude []interface{} `json:"exclude"`

	// The require policy works like a AND logical operator. The user must
	// satisfy all the rules in require.
	Require []interface{} `json:"require"`
}

type UpdateAccessPolicyParams struct {
	ApplicationID string `json:"-"`
	PolicyID      string `json:"-"`

	Precedence int    `json:"precedence"`
	Decision   string `json:"decision"`
	Name       string `json:"name"`

	IsolationRequired            *bool                 `json:"isolation_required,omitempty"`
	SessionDuration              *string               `json:"session_duration,omitempty"`
	PurposeJustificationRequired *bool                 `json:"purpose_justification_required,omitempty"`
	PurposeJustificationPrompt   *string               `json:"purpose_justification_prompt,omitempty"`
	ApprovalRequired             *bool                 `json:"approval_required,omitempty"`
	ApprovalGroups               []AccessApprovalGroup `json:"approval_groups"`

	// The include policy works like an OR logical operator. The user must
	// satisfy one of the rules.
	Include []interface{} `json:"include"`

	// The exclude policy works like a NOT logical operator. The user must
	// not satisfy all the rules in exclude.
	Exclude []interface{} `json:"exclude"`

	// The require policy works like a AND logical operator. The user must
	// satisfy all the rules in require.
	Require []interface{} `json:"require"`
}

type DeleteAccessPolicyParams struct {
	ApplicationID string `json:"-"`
	PolicyID      string `json:"-"`
}

// ListAccessPolicies returns all access policies for an access application.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-policies-list-access-policies
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-policies-list-access-policies
func (api *API) ListAccessPolicies(ctx context.Context, rc *ResourceContainer, params ListAccessPoliciesParams) ([]AccessPolicy, *ResultInfo, error) {
	if params.ApplicationID == "" {
		return []AccessPolicy{}, &ResultInfo{}, ErrMissingApplicationID
	}

	baseURL := fmt.Sprintf(
		"/%s/%s/access/apps/%s/policies",
		rc.Level,
		rc.Identifier,
		params.ApplicationID,
	)

	autoPaginate := true
	if params.PerPage >= 1 || params.Page >= 1 {
		autoPaginate = false
	}

	if params.PerPage < 1 {
		params.PerPage = 25
	}

	if params.Page < 1 {
		params.Page = 1
	}

	var accessPolicies []AccessPolicy
	var r AccessPolicyListResponse
	for {
		uri := buildURI(baseURL, params)
		res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return []AccessPolicy{}, &ResultInfo{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
		}

		err = json.Unmarshal(res, &r)
		if err != nil {
			return []AccessPolicy{}, &ResultInfo{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
		}
		accessPolicies = append(accessPolicies, r.Result...)
		params.ResultInfo = r.ResultInfo.Next()
		if params.ResultInfo.Done() || !autoPaginate {
			break
		}
	}

	return accessPolicies, &r.ResultInfo, nil
}

// GetAccessPolicy returns a single policy based on the policy ID.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-policies-get-an-access-policy
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-policies-get-an-access-policy
func (api *API) GetAccessPolicy(ctx context.Context, rc *ResourceContainer, params GetAccessPolicyParams) (AccessPolicy, error) {
	uri := fmt.Sprintf(
		"/%s/%s/access/apps/%s/policies/%s",
		rc.Level,
		rc.Identifier,
		params.ApplicationID,
		params.PolicyID,
	)

	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessPolicyDetailResponse AccessPolicyDetailResponse
	err = json.Unmarshal(res, &accessPolicyDetailResponse)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessPolicyDetailResponse.Result, nil
}

// CreateAccessPolicy creates a new access policy.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-policies-create-an-access-policy
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-policies-create-an-access-policy
func (api *API) CreateAccessPolicy(ctx context.Context, rc *ResourceContainer, params CreateAccessPolicyParams) (AccessPolicy, error) {
	uri := fmt.Sprintf(
		"/%s/%s/access/apps/%s/policies",
		rc.Level,
		rc.Identifier,
		params.ApplicationID,
	)

	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessPolicyDetailResponse AccessPolicyDetailResponse
	err = json.Unmarshal(res, &accessPolicyDetailResponse)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessPolicyDetailResponse.Result, nil
}

// UpdateAccessPolicy updates an existing access policy.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-policies-update-an-access-policy
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-policies-update-an-access-policy
func (api *API) UpdateAccessPolicy(ctx context.Context, rc *ResourceContainer, params UpdateAccessPolicyParams) (AccessPolicy, error) {
	if params.PolicyID == "" {
		return AccessPolicy{}, fmt.Errorf("access policy ID cannot be empty")
	}

	uri := fmt.Sprintf(
		"/%s/%s/access/apps/%s/policies/%s",
		rc.Level,
		rc.Identifier,
		params.ApplicationID,
		params.PolicyID,
	)

	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessPolicyDetailResponse AccessPolicyDetailResponse
	err = json.Unmarshal(res, &accessPolicyDetailResponse)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessPolicyDetailResponse.Result, nil
}

// DeleteAccessPolicy deletes an access policy.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-policies-delete-an-access-policy
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-policies-delete-an-access-policy
func (api *API) DeleteAccessPolicy(ctx context.Context, rc *ResourceContainer, params DeleteAccessPolicyParams) error {
	uri := fmt.Sprintf(
		"/%s/%s/access/apps/%s/policies/%s",
		rc.Level,
		rc.Identifier,
		params.ApplicationID,
		params.PolicyID,
	)

	_, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	return nil
}
