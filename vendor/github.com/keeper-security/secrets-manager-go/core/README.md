# Secrets Management Go SDK

![Go](https://github.com/keeper-security/secrets-manager-go/actions/workflows/test.go.yml/badge.svg)

<p align="center">
  <a href="https://docs.keeper.io/secrets-manager/secrets-manager/developer-sdk-library/golang-sdk">View docs</a>
</p>

This library provides interface to Keeper® Secrets Manager and can be used to access your Keeper vault, read and update existing records, rotate passwords and more. Keeper Secrets Manager is an open source project with contributions from Keeper's engineering team and partners.

## Features:

## Obtain a One-Time Access Token
Keeper Secrets Manager authenticates your API requests using advanced encryption that uses locally stored private key, device id and client id.
To register your device and generate private key you will need to generate a One-Time Access Token via Web Vault or Keeper Commander CLI.

### Via Web Vault
**Secrets Manager > Applications > Create Application** - will let you chose application name, shared folder(s) and permissions and generate One-Time Access Token. _Note: Keeper does not store One-Time Access Tokens - save or copy the token offline for later use._

One-Time Access Tokens can be generated as needed: **Secrets Manager > Applications > Application Name > Devices Tab > Edit > Add Device button** - will let you create new Device and generate its One-Time Access Token.

[What is an application?](https://docs.keeper.io/secrets-manager/secrets-manager/overview/terminology)

### Via Keeper Commander CLI
Login to Keeper with Commander CLI and perform following:
1. Create Application
    ```bash
   $ sm app create [NAME]
    ```

2. Share Secrets to the Application
    ```bash
   $ sm share add --app [NAME] --secret [UID] --editable
    ```
    - `--app` - Name of the Application.
    - `--secret` - Record UID or Shared Folder UID
    - `--editable` - if omitted defaults to false

3. Create client
    ```bash
   $ sm client add --app [NAME] --unlock-ip --count 1
    ```

### Install
```bash
go get github.com/keeper-security/secrets-manager-go/core
```

### Quick Start

```golang
package main

// Import Secrets Manager
import ksm "github.com/keeper-security/secrets-manager-go/core"

func main() {
	// Establish connection
	// One time secrets generated via Web Vault or Commander CLI
	clientOptions := &ksm.ClientOptions{
		Token:  "US:ONE_TIME_TOKEN_BASE64",
		Config: ksm.NewFileKeyValueStorage("ksm-config.json")}
	sm := ksm.NewSecretsManager(clientOptions)
	// One time tokens can be used only once - afterwards use the generated config file
	// sm := ksm.NewSecretsManager(&ksm.ClientOptions{Config: ksm.NewFileKeyValueStorage("ksm-config.json")})

	// Retrieve all records
	allRecords, _ := sm.GetSecrets([]string{})

	// Get password from first record:
	password := allRecords[0].Password()

	// WARNING: Avoid logging sensitive data
	print("My password from Keeper: ", password)
}
```

## Samples
### File Download
```golang
sm := ksm.NewSecretsManager(&ksm.ClientOptions{Config: ksm.NewFileKeyValueStorage("ksm-config.json")})

if records, err := sm.GetSecrets([]string{}); err == nil {
	for _, r := range records {
		fmt.Println("\tTitle: " + r.Title())
		for i, f := range r.Files {
			fmt.Printf("\t\tfile #%d -> name: %s", i, f.Name)
			f.SaveFile("/tmp/"+f.Name, true)
		}
	}
}
```

### Update record
```golang
sm := ksm.NewSecretsManager(&ksm.ClientOptions{Config: ksm.NewFileKeyValueStorage("ksm-config.json")})

if records, err := sm.GetSecrets([]string{}); err == nil && len(records) > 0 {
	record := records[0]
	newPassword := fmt.Sprintf("Test Password - " + time.Now().Format(time.RFC850))
	record.SetPassword(newPassword)

	if err := sm.Save(record); err != nil {
		fmt.Println("Error saving record: " + err.Error())
	}
}
```

## Configuration

### Types

Listed in priority order
1. Environment variable
1. Configuration store
1. Code

### Available configurations:

- `clientKey` - One Time Access Token used during initialization
- `hostname` - Keeper Backend host. Available values:
    - `keepersecurity.com`
    - `keepersecurity.eu`
    - `keepersecurity.com.au`
    - `govcloud.keepersecurity.us`

## Adding more records or shared folders to the Application

### Via Web Vault
Drag&Drop records into the shared folder or select from the record menu any of the options to CreateDuplicate/Move or create new records straight into the shared folder. As an alternative use: **Secrets Manager > Application > Application Name > Folders & Records > Edit** and use search field to add any folders or records then click Save.

### Via Commander CLI
```bash
sm share add --app [NAME] --secret [UID2]
sm share add --app [NAME] --secret [UID3] --editable
```

### Retrieve secret(s)
```golang
sm := ksm.NewSecretsManager(&ksm.ClientOptions{Config: ksm.NewFileKeyValueStorage("ksm-config.json")})
allSecrets, _ := sm.GetSecrets([]string{})
```

### Update secret
```golang
secretToUpdate = allSecrets[0]
secretToUpdate.SetPassword("NewPassword123$")
secretsManager.Save(secretToUpdate)
```

# Change Log

## 1.6.2

* KSM-467 - Fixed ExpiresOn conversion from UnixTimeMilliseconds.

## 1.6.1

* KSM-450 - Added `folderUid` and `innerFolderUid` to Record
* KSM-451 - Fix `subFolderUid` crash on empty string value

## 1.6.0

* KSM-414 - Added support for Folders
* KSM-435 - Improved Passkey field type support

## 1.5.2

* KSM-409 New field type: Passkey
* KSM-404 New filed type: script and modification to some record types
* KSM-384 Support for record Transactions


## 1.5.0

* KSM-317 - Notation improvements
* KSM-356 - Create custom fields
* KSM-365 - Fixed KEY_CLINET_KEY is missing error
* KSM-366 - Avoid exceptions/panics and return errors instead
* KSM-367 - Fixed license not shown on pkg.go.dev

## 1.4.0

* KSM-288 - Record removal
* KSM-306 - Added support for Japan and Canada data centers
* KSM-312 - Improve password generation entropy

For additional information please check our detailed [Go SDK docs](https://docs.keeper.io/secrets-manager/secrets-manager/developer-sdk-library/golang-sdk) for Keeper Secrets Manager.

### Documentation
[Secrets Manager Guide](https://docs.keeper.io/secrets-manager/secrets-manager/overview)

[Enterprise Admin Guide](https://docs.keeper.io/enterprise-guide/)

[Keeper Commander Guide](https://docs.keeper.io/secrets-manager/commander-cli/overview)

[Keeper Security Website](https://www.keepersecurity.com/secrets-manager.html)
