package client

import "net/http"

type Interceptor func(http.RoundTripper) http.RoundTripper

type InterceptorChain struct {
	transport    http.RoundTripper
	interceptors []Interceptor
}

func NewInterceptorChain(interceptors []Interceptor) *InterceptorChain {
	return &InterceptorChain{
		transport:    http.DefaultTransport,
		interceptors: interceptors,
	}
}

func (c *InterceptorChain) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := c.transport
	len := len(c.interceptors)

	// Применяем интерсепторы в обратном порядке
	for i := len - 1; i >= 0; i-- {
		transport = c.interceptors[i](transport)
	}

	return transport.RoundTrip(req)
}
