package client

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

// Интерсептор для логирования
func LoggingInterceptor() Interceptor {
	return func(next http.RoundTripper) http.RoundTripper {
		return &LoggingTransport{transport: next}
	}
}

type LoggingTransport struct {
	transport http.RoundTripper
}

func (s *LoggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	bytes, _ := httputil.DumpRequestOut(r, true)

	resp, err := s.transport.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	respBytes, _ := httputil.DumpResponse(resp, true)
	bytes = append(bytes, respBytes...)

	fmt.Printf("%s\n", bytes)

	return resp, err
}
