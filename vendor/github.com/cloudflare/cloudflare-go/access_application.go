package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

// AccessApplicationType represents the application type.
type AccessApplicationType string

// These constants represent all valid application types.
const (
	SelfHosted  AccessApplicationType = "self_hosted"
	SSH         AccessApplicationType = "ssh"
	VNC         AccessApplicationType = "vnc"
	Biso        AccessApplicationType = "biso"
	AppLauncher AccessApplicationType = "app_launcher"
	Warp        AccessApplicationType = "warp"
	Bookmark    AccessApplicationType = "bookmark"
	Saas        AccessApplicationType = "saas"
)

// AccessApplication represents an Access application.
type AccessApplication struct {
	GatewayRules             []AccessApplicationGatewayRule `json:"gateway_rules,omitempty"`
	AllowedIdps              []string                       `json:"allowed_idps,omitempty"`
	CustomDenyMessage        string                         `json:"custom_deny_message,omitempty"`
	LogoURL                  string                         `json:"logo_url,omitempty"`
	AUD                      string                         `json:"aud,omitempty"`
	Domain                   string                         `json:"domain"`
	SelfHostedDomains        []string                       `json:"self_hosted_domains"`
	Type                     AccessApplicationType          `json:"type,omitempty"`
	SessionDuration          string                         `json:"session_duration,omitempty"`
	SameSiteCookieAttribute  string                         `json:"same_site_cookie_attribute,omitempty"`
	CustomDenyURL            string                         `json:"custom_deny_url,omitempty"`
	CustomNonIdentityDenyURL string                         `json:"custom_non_identity_deny_url,omitempty"`
	Name                     string                         `json:"name"`
	ID                       string                         `json:"id,omitempty"`
	PrivateAddress           string                         `json:"private_address"`
	CorsHeaders              *AccessApplicationCorsHeaders  `json:"cors_headers,omitempty"`
	CreatedAt                *time.Time                     `json:"created_at,omitempty"`
	UpdatedAt                *time.Time                     `json:"updated_at,omitempty"`
	SaasApplication          *SaasApplication               `json:"saas_app,omitempty"`
	AutoRedirectToIdentity   *bool                          `json:"auto_redirect_to_identity,omitempty"`
	SkipInterstitial         *bool                          `json:"skip_interstitial,omitempty"`
	AppLauncherVisible       *bool                          `json:"app_launcher_visible,omitempty"`
	EnableBindingCookie      *bool                          `json:"enable_binding_cookie,omitempty"`
	HttpOnlyCookieAttribute  *bool                          `json:"http_only_cookie_attribute,omitempty"`
	ServiceAuth401Redirect   *bool                          `json:"service_auth_401_redirect,omitempty"`
	PathCookieAttribute      *bool                          `json:"path_cookie_attribute,omitempty"`
	CustomPages              []string                       `json:"custom_pages,omitempty"`
	Tags                     []string                       `json:"tags,omitempty"`
	AccessAppLauncherCustomization
}

type AccessApplicationGatewayRule struct {
	ID string `json:"id,omitempty"`
}

// AccessApplicationCorsHeaders represents the CORS HTTP headers for an Access
// Application.
type AccessApplicationCorsHeaders struct {
	AllowedMethods   []string `json:"allowed_methods,omitempty"`
	AllowedOrigins   []string `json:"allowed_origins,omitempty"`
	AllowedHeaders   []string `json:"allowed_headers,omitempty"`
	AllowAllMethods  bool     `json:"allow_all_methods,omitempty"`
	AllowAllHeaders  bool     `json:"allow_all_headers,omitempty"`
	AllowAllOrigins  bool     `json:"allow_all_origins,omitempty"`
	AllowCredentials bool     `json:"allow_credentials,omitempty"`
	MaxAge           int      `json:"max_age,omitempty"`
}

// AccessApplicationListResponse represents the response from the list
// access applications endpoint.
type AccessApplicationListResponse struct {
	Result []AccessApplication `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// AccessApplicationDetailResponse is the API response, containing a single
// access application.
type AccessApplicationDetailResponse struct {
	Success  bool              `json:"success"`
	Errors   []string          `json:"errors"`
	Messages []string          `json:"messages"`
	Result   AccessApplication `json:"result"`
}

type SourceConfig struct {
	Name      string            `json:"name,omitempty"`
	NameByIDP map[string]string `json:"name_by_idp,omitempty"`
}

type SAMLAttributeConfig struct {
	Name         string       `json:"name,omitempty"`
	NameFormat   string       `json:"name_format,omitempty"`
	FriendlyName string       `json:"friendly_name,omitempty"`
	Required     bool         `json:"required,omitempty"`
	Source       SourceConfig `json:"source"`
}

type SaasApplication struct {
	AppID              string                `json:"app_id,omitempty"`
	ConsumerServiceUrl string                `json:"consumer_service_url,omitempty"`
	SPEntityID         string                `json:"sp_entity_id,omitempty"`
	PublicKey          string                `json:"public_key,omitempty"`
	IDPEntityID        string                `json:"idp_entity_id,omitempty"`
	NameIDFormat       string                `json:"name_id_format,omitempty"`
	SSOEndpoint        string                `json:"sso_endpoint,omitempty"`
	DefaultRelayState  string                `json:"default_relay_state,omitempty"`
	UpdatedAt          *time.Time            `json:"updated_at,omitempty"`
	CreatedAt          *time.Time            `json:"created_at,omitempty"`
	CustomAttributes   []SAMLAttributeConfig `json:"custom_attributes,omitempty"`
}

type AccessAppLauncherCustomization struct {
	LandingPageDesign     AccessLandingPageDesign `json:"landing_page_design"`
	LogoURL               string                  `json:"app_launcher_logo_url"`
	HeaderBackgroundColor string                  `json:"header_bg_color"`
	BackgroundColor       string                  `json:"bg_color"`
	FooterLinks           []AccessFooterLink      `json:"footer_links"`
}

type AccessFooterLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type AccessLandingPageDesign struct {
	Title           string `json:"title"`
	Message         string `json:"message"`
	ImageURL        string `json:"image_url"`
	ButtonColor     string `json:"button_color"`
	ButtonTextColor string `json:"button_text_color"`
}
type ListAccessApplicationsParams struct {
	ResultInfo
}

type CreateAccessApplicationParams struct {
	AllowedIdps              []string                       `json:"allowed_idps,omitempty"`
	AppLauncherVisible       *bool                          `json:"app_launcher_visible,omitempty"`
	AUD                      string                         `json:"aud,omitempty"`
	AutoRedirectToIdentity   *bool                          `json:"auto_redirect_to_identity,omitempty"`
	CorsHeaders              *AccessApplicationCorsHeaders  `json:"cors_headers,omitempty"`
	CustomDenyMessage        string                         `json:"custom_deny_message,omitempty"`
	CustomDenyURL            string                         `json:"custom_deny_url,omitempty"`
	CustomNonIdentityDenyURL string                         `json:"custom_non_identity_deny_url,omitempty"`
	Domain                   string                         `json:"domain"`
	EnableBindingCookie      *bool                          `json:"enable_binding_cookie,omitempty"`
	GatewayRules             []AccessApplicationGatewayRule `json:"gateway_rules,omitempty"`
	HttpOnlyCookieAttribute  *bool                          `json:"http_only_cookie_attribute,omitempty"`
	LogoURL                  string                         `json:"logo_url,omitempty"`
	Name                     string                         `json:"name"`
	PathCookieAttribute      *bool                          `json:"path_cookie_attribute,omitempty"`
	PrivateAddress           string                         `json:"private_address"`
	SaasApplication          *SaasApplication               `json:"saas_app,omitempty"`
	SameSiteCookieAttribute  string                         `json:"same_site_cookie_attribute,omitempty"`
	SelfHostedDomains        []string                       `json:"self_hosted_domains"`
	ServiceAuth401Redirect   *bool                          `json:"service_auth_401_redirect,omitempty"`
	SessionDuration          string                         `json:"session_duration,omitempty"`
	SkipInterstitial         *bool                          `json:"skip_interstitial,omitempty"`
	Type                     AccessApplicationType          `json:"type,omitempty"`
	CustomPages              []string                       `json:"custom_pages,omitempty"`
	Tags                     []string                       `json:"tags,omitempty"`
	AccessAppLauncherCustomization
}

type UpdateAccessApplicationParams struct {
	ID                       string                         `json:"id,omitempty"`
	AllowedIdps              []string                       `json:"allowed_idps,omitempty"`
	AppLauncherVisible       *bool                          `json:"app_launcher_visible,omitempty"`
	AUD                      string                         `json:"aud,omitempty"`
	AutoRedirectToIdentity   *bool                          `json:"auto_redirect_to_identity,omitempty"`
	CorsHeaders              *AccessApplicationCorsHeaders  `json:"cors_headers,omitempty"`
	CustomDenyMessage        string                         `json:"custom_deny_message,omitempty"`
	CustomDenyURL            string                         `json:"custom_deny_url,omitempty"`
	CustomNonIdentityDenyURL string                         `json:"custom_non_identity_deny_url,omitempty"`
	Domain                   string                         `json:"domain"`
	EnableBindingCookie      *bool                          `json:"enable_binding_cookie,omitempty"`
	GatewayRules             []AccessApplicationGatewayRule `json:"gateway_rules,omitempty"`
	HttpOnlyCookieAttribute  *bool                          `json:"http_only_cookie_attribute,omitempty"`
	LogoURL                  string                         `json:"logo_url,omitempty"`
	Name                     string                         `json:"name"`
	PathCookieAttribute      *bool                          `json:"path_cookie_attribute,omitempty"`
	PrivateAddress           string                         `json:"private_address"`
	SaasApplication          *SaasApplication               `json:"saas_app,omitempty"`
	SameSiteCookieAttribute  string                         `json:"same_site_cookie_attribute,omitempty"`
	SelfHostedDomains        []string                       `json:"self_hosted_domains"`
	ServiceAuth401Redirect   *bool                          `json:"service_auth_401_redirect,omitempty"`
	SessionDuration          string                         `json:"session_duration,omitempty"`
	SkipInterstitial         *bool                          `json:"skip_interstitial,omitempty"`
	Type                     AccessApplicationType          `json:"type,omitempty"`
	CustomPages              []string                       `json:"custom_pages,omitempty"`
	Tags                     []string                       `json:"tags,omitempty"`
	AccessAppLauncherCustomization
}

// ListAccessApplications returns all applications within an account or zone.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-applications-list-access-applications
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-applications-list-access-applications
func (api *API) ListAccessApplications(ctx context.Context, rc *ResourceContainer, params ListAccessApplicationsParams) ([]AccessApplication, *ResultInfo, error) {
	baseURL := fmt.Sprintf("/%s/%s/access/apps", rc.Level, rc.Identifier)

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

	var applications []AccessApplication
	var r AccessApplicationListResponse

	for {
		uri := buildURI(baseURL, params)

		res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return []AccessApplication{}, &ResultInfo{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
		}

		err = json.Unmarshal(res, &r)
		if err != nil {
			return []AccessApplication{}, &ResultInfo{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
		}
		applications = append(applications, r.Result...)
		params.ResultInfo = r.ResultInfo.Next()
		if params.ResultInfo.Done() || !autoPaginate {
			break
		}
	}

	return applications, &r.ResultInfo, nil
}

// GetAccessApplication returns a single application based on the application
// ID for either account or zone.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-applications-get-an-access-application
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-applications-get-an-access-application
func (api *API) GetAccessApplication(ctx context.Context, rc *ResourceContainer, applicationID string) (AccessApplication, error) {
	uri := fmt.Sprintf(
		"/%s/%s/access/apps/%s",
		rc.Level,
		rc.Identifier,
		applicationID,
	)

	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return AccessApplication{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessApplicationDetailResponse AccessApplicationDetailResponse
	err = json.Unmarshal(res, &accessApplicationDetailResponse)
	if err != nil {
		return AccessApplication{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessApplicationDetailResponse.Result, nil
}

// CreateAccessApplication creates a new access application.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-applications-add-an-application
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-applications-add-a-bookmark-application
func (api *API) CreateAccessApplication(ctx context.Context, rc *ResourceContainer, params CreateAccessApplicationParams) (AccessApplication, error) {
	uri := fmt.Sprintf("/%s/%s/access/apps", rc.Level, rc.Identifier)

	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return AccessApplication{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessApplicationDetailResponse AccessApplicationDetailResponse
	err = json.Unmarshal(res, &accessApplicationDetailResponse)
	if err != nil {
		return AccessApplication{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessApplicationDetailResponse.Result, nil
}

// UpdateAccessApplication updates an existing access application.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-applications-update-a-bookmark-application
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-applications-update-a-bookmark-application
func (api *API) UpdateAccessApplication(ctx context.Context, rc *ResourceContainer, params UpdateAccessApplicationParams) (AccessApplication, error) {
	if params.ID == "" {
		return AccessApplication{}, fmt.Errorf("access application ID cannot be empty")
	}

	uri := fmt.Sprintf(
		"/%s/%s/access/apps/%s",
		rc.Level,
		rc.Identifier,
		params.ID,
	)

	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return AccessApplication{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessApplicationDetailResponse AccessApplicationDetailResponse
	err = json.Unmarshal(res, &accessApplicationDetailResponse)
	if err != nil {
		return AccessApplication{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessApplicationDetailResponse.Result, nil
}

// DeleteAccessApplication deletes an access application.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-applications-delete-an-access-application
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-applications-delete-an-access-application
func (api *API) DeleteAccessApplication(ctx context.Context, rc *ResourceContainer, applicationID string) error {
	uri := fmt.Sprintf(
		"/%s/%s/access/apps/%s",
		rc.Level,
		rc.Identifier,
		applicationID,
	)

	_, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	return nil
}

// RevokeAccessApplicationTokens revokes tokens associated with an
// access application.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-applications-revoke-service-tokens
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-applications-revoke-service-tokens
func (api *API) RevokeAccessApplicationTokens(ctx context.Context, rc *ResourceContainer, applicationID string) error {
	uri := fmt.Sprintf(
		"/%s/%s/access/apps/%s/revoke-tokens",
		rc.Level,
		rc.Identifier,
		applicationID,
	)

	_, err := api.makeRequestContext(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	return nil
}
