package tracker

import (
	"net/http"

	"example.com/tracker/client"
)

// Интерсептор для добавления токена
func AuthTokenInterceptor(token string) client.Interceptor {
	return func(next http.RoundTripper) http.RoundTripper {
		return &authRoundTripper{next: next, token: token}
	}
}

type authRoundTripper struct {
	next  http.RoundTripper
	token string
}

func (a *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", "Bearer "+a.token)
	return a.next.RoundTrip(reqClone)
}
