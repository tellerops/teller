package auth

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"os"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// NewAuthorizerFromEnvironment creates a keyvault dataplane Authorizer configured from environment variables in the order:
// 1. Client credentials
// 2. Client certificate
// 3. Username password
// 4. MSI
func NewAuthorizerFromEnvironment() (autorest.Authorizer, error) {
	res, err := getResource()
	if err != nil {
		return nil, err
	}
	return auth.NewAuthorizerFromEnvironmentWithResource(res)
}

// NewAuthorizerFromFile creates a keyvault dataplane Authorizer configured from a configuration file.
// The path to the configuration file must be specified in the AZURE_AUTH_LOCATION environment variable.
func NewAuthorizerFromFile() (autorest.Authorizer, error) {
	res, err := getResource()
	if err != nil {
		return nil, err
	}
	return auth.NewAuthorizerFromFileWithResource(res)
}

// NewAuthorizerFromCLI creates a keyvault dataplane Authorizer configured from Azure CLI 2.0 for local development scenarios.
func NewAuthorizerFromCLI() (autorest.Authorizer, error) {
	res, err := getResource()
	if err != nil {
		return nil, err
	}
	return auth.NewAuthorizerFromCLIWithResource(res)
}

func getResource() (string, error) {
	var env azure.Environment

	if envName := os.Getenv("AZURE_ENVIRONMENT"); envName == "" {
		env = azure.PublicCloud
	} else {
		var err error
		env, err = azure.EnvironmentFromName(envName)
		if err != nil {
			return "", err
		}
	}

	resource := os.Getenv("AZURE_KEYVAULT_RESOURCE")
	if resource == "" {
		resource = env.ResourceIdentifiers.KeyVault
	}

	return resource, nil
}
