package authn

type TokenAuthenticator struct {
	Token string `env:"CONJUR_AUTHN_TOKEN"`
}

func (a *TokenAuthenticator) RefreshToken() ([]byte, error) {
	return []byte(a.Token), nil
}

func (a *TokenAuthenticator) NeedsTokenRefresh() bool {
	return false
}
