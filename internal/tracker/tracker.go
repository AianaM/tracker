package tracker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"example.com/tracker/internal/client"
)

const baseUrl = "https://api.tracker.yandex.net/v3/"

type Config struct {
	Ctx     context.Context
	Timeout time.Duration
	Client  *client.Client
	HostURL string
}

type TrackerClient struct {
	Config
}

func NewTrackerClient(config Config) *TrackerClient {
	return &TrackerClient{
		Config: config,
	}
}

type keyValue struct {
	key, value string
}
type request struct {
	path, method string
	headers      map[string]string
	params       []keyValue
}
type response[T any] struct {
	statusCode int
	body       T
}
type requestData[T any] struct {
	client   *TrackerClient
	request  request
	response response[T]
}

func (r requestData[T]) requestNew() (T, error) {
	url := baseUrl + r.request.path

	ctx, cancel := context.WithTimeout(r.client.Config.Ctx, r.client.Config.Timeout)
	defer cancel()

	if req, err := r.client.Config.Client.NewRequest(ctx, r.request.method, url, nil); err != nil {
		return r.response.body, fmt.Errorf("creating request: %w", err)
	} else {
		query := req.URL.Query()
		for _, param := range r.request.params {
			query.Add(param.key, param.value)
		}
		req.URL.RawQuery = query.Encode()
		for key, value := range r.request.headers {
			req.Header.Add(key, value)
		}
		req.Header.Add("Accept", "application/json")

		status, err := r.client.Config.Client.Do(req, &r.response.body)

		if err != nil {
			return r.response.body, fmt.Errorf("executing request: %w", err)
		} else if status != http.StatusOK {
			return r.response.body, fmt.Errorf("received status code %d", status)
		}
		return r.response.body, nil
	}
}
