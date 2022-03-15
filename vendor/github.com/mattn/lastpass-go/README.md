# LastPass Go API

**This is unofficial LastPass API.**

This is a port of [Ruby API](https://github.com/detunized/lastpass-ruby).

## Usage

```go
vault, _ := lastpass.CreateVault(username, password)
for _, account := range vault.Accounts {
	fmt.Println(account.Username, account.Password)
}
```

## Requirements

golang

## Installation

```
$ go get github.com/mattn/lastpass-go
```

## License

MIT

Note that this repository include code of `ecb` (Electronic Code Block) provided by Go Authors.

## Author

Yasuhiro Matsumoto (a.k.a mattn)
