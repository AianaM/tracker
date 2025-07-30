package tracker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"example.com/tracker/client"
)

const baseUrl = "https://api.tracker.yandex.net/v2/"

var (
	ctx     = context.Background()
	timeout = 10 * time.Second
)

type keyValue struct {
	key, value string
}
type Request struct {
	Path, Method string
	headers      map[string]string
	params       []keyValue
}
type Response[T any] struct {
	StatusCode int
	Body       T
}
type RequestData[T any] struct {
	Request  Request
	Response Response[T]
}

func (r RequestData[T]) requestNew() (T, error) {
	url := baseUrl + r.Request.Path

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if req, err := client.RequestNew(ctx, r.Request.Method, url, nil); err != nil {
		return r.Response.Body, fmt.Errorf("creating request: %w", err)
	} else {
		query := req.URL.Query()
		for _, param := range r.Request.params {
			query.Add(param.key, param.value)
		}
		req.URL.RawQuery = query.Encode()
		for key, value := range r.Request.headers {
			req.Header.Add(key, value)
		}
		req.Header.Add("Accept", "application/json")

		status, err := client.ExecRequest(req, &r.Response.Body)

		if err != nil {
			return r.Response.Body, fmt.Errorf("executing request: %w", err)
		} else if status != http.StatusOK {
			return r.Response.Body, fmt.Errorf("received status code %d", status)
		}
		return r.Response.Body, nil
	}
}
