package authn

type OidcAuthenticator struct {
	Code         string
	Nonce        string
	CodeVerifier string
	Authenticate func(code, noce, code_verifier string) ([]byte, error)
}

func (a *OidcAuthenticator) RefreshToken() ([]byte, error) {
	return a.Authenticate(a.Code, a.Nonce, a.CodeVerifier)
}

func (a *OidcAuthenticator) NeedsTokenRefresh() bool {
	return false
}
