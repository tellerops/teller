# ansible-vault-go

Go package to read/write Ansible Vault secrets

[![GoDoc](https://godoc.org/github.com/sosedoff/ansible-vault-go?status.svg)](https://godoc.org/github.com/sosedoff/ansible-vault-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/sosedoff/ansible-vault-go)](https://goreportcard.com/report/github.com/sosedoff/ansible-vault-go)

## Installation

```
go get github.com/sosedoff/ansible-vault-go
```

## Usage

```go
package main

import(
  "log"

  "github.com/sosedoff/ansible-vault-go"
)

func main() {
  // Encrypt secret data
  str, err := vault.Encrypt("secret", "password")

  // Decrypt secret data
  str, err := vault.Decrypt("secret", "password")

  // Write secret data to file
  err := vault.EncryptFile("path/to/secret/file", "secret", "password")

  // Read existing secret
  str, err := vault.DecryptFile("path/to/secret/file", "password")
}
```

## Doc

Check out the Ansible documentation regarding the Vault file format:

- https://docs.ansible.com/ansible/2.4/vault.html#vault-format

## License

MIT