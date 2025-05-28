package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

const (
	clientTimeout = 10 * time.Second
)

type keyValue struct {
	key, value string
}

// описание запроса
type requestData[T any] struct {
	url     string
	headers map[string]string
	params  []keyValue
	body    T
}

func newRequestData[T any](url string, headers map[string]string, params []keyValue, body T) requestData[T] {
	if headers == nil {
		headers = map[string]string{}
	}
	if params == nil {
		params = []keyValue{}
	}
	return requestData[T]{
		url:     url,
		headers: headers,
		params:  params,
		body:    body,
	}
}

// отправляем запрос
func (r *requestData[T]) get() error {
	client := &http.Client{Timeout: clientTimeout}
	reqURL, err := url.Parse(r.url)
	if err != nil {
		return err
	}

	query := reqURL.Query()
	for _, value := range r.params {
		query.Add(value.key, value.value)
	}
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+getToken())
	for key, value := range r.headers {
		req.Header.Add(key, value)
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&r.body)

	return nil
}
