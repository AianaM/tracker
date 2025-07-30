package tracker

import (
	"net/http"

	"example.com/tracker/client"
)

// Интерсептор для добавления токена
func AuthTokenInterceptor(headers map[string]string) client.Interceptor {
	return func(next http.RoundTripper) http.RoundTripper {
		return &authRoundTripper{next: next, headers: headers}
	}
}

type authRoundTripper struct {
	next    http.RoundTripper
	headers map[string]string
}

func (a *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqClone := req.Clone(req.Context())
	for key, value := range a.headers {
		reqClone.Header.Set(key, value)
	}
	return a.next.RoundTrip(reqClone)
}
