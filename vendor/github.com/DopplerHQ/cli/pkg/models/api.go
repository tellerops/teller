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
package models

// ComputedSecret holds all info about a secret
type ComputedSecret struct {
	Name          string `json:"name"`
	RawValue      string `json:"raw"`
	ComputedValue string `json:"computed"`
}

// WorkplaceSettings workplace settings
type WorkplaceSettings struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	BillingEmail string `json:"billing_email"`
}

// ProjectInfo project info
type ProjectInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

// EnvironmentInfo environment info
type EnvironmentInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	CreatedAt      string `json:"created_at"`
	InitialFetchAt string `json:"initial_fetch_at"`
	Project        string `json:"project"`
}

// ConfigInfo project info
type ConfigInfo struct {
	Name           string `json:"name"`
	Root           bool   `json:"root"`
	Locked         bool   `json:"locked"`
	Environment    string `json:"environment"`
	Project        string `json:"project"`
	CreatedAt      string `json:"created_at"`
	InitialFetchAt string `json:"initial_fetch_at"`
	LastFetchAt    string `json:"last_fetch_at"`
}

// ConfigLog a log
type ConfigLog struct {
	ID          string    `json:"id"`
	Text        string    `json:"text"`
	HTML        string    `json:"html"`
	CreatedAt   string    `json:"created_at"`
	Config      string    `json:"config"`
	Environment string    `json:"environment"`
	Project     string    `json:"project"`
	User        User      `json:"user"`
	Diff        []LogDiff `json:"diff"`
}

// ActivityLog an activity log
type ActivityLog struct {
	ID                 string `json:"id"`
	Text               string `json:"text"`
	HTML               string `json:"html"`
	CreatedAt          string `json:"created_at"`
	EnclaveConfig      string `json:"enclave_config"`
	EnclaveEnvironment string `json:"enclave_environment"`
	EnclaveProject     string `json:"enclave_project"`
	User               User   `json:"user"`
}

// User user profile
type User struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	Username     string `json:"username"`
	ProfileImage string `json:"profile_image_url"`
}

// LogDiff diff of log entries
type LogDiff struct {
	Name    string `json:"name"`
	Added   string `json:"added"`
	Removed string `json:"removed"`
}

// ConfigServiceToken a service token
type ConfigServiceToken struct {
	Name        string `json:"name"`
	Token       string `json:"token"`
	Slug        string `json:"slug"`
	CreatedAt   string `json:"created_at"`
	Project     string `json:"project"`
	Environment string `json:"environment"`
	Config      string `json:"config"`
}
