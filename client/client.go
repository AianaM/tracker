package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var client http.Client

func New(interceptors []Interceptor) {
	client = http.Client{
		Timeout:   time.Second * 10,
		Transport: NewInterceptorChain(interceptors),
	}

}

func RequestNew(ctx context.Context, httpMethod, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, httpMethod, url, body)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return req, err
	}

	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func ExecRequest(request *http.Request, bodyInterface interface{}) (status int, err error) {
	resp, err := client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("error making request to %s: %w", request.URL.Path, err)
	}
	defer resp.Body.Close()

	status = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		return status, fmt.Errorf("error: received status code %d", status)
	}

	if bodyInterface != nil {
		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(bodyInterface); err != nil {
			return status, fmt.Errorf("error decoding response body: %w", err)
		}
	}

	return status, nil
}
