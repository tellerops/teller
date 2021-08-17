package authn

type APIKeyAuthenticator struct {
	Authenticate func(loginPair LoginPair) ([]byte, error)
	LoginPair
}

type LoginPair struct {
	Login  string
	APIKey string
}

func (a *APIKeyAuthenticator) RefreshToken() ([]byte, error) {
	return a.Authenticate(a.LoginPair)
}

func (a *APIKeyAuthenticator) NeedsTokenRefresh() bool {
	return false
}
