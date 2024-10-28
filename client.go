package main

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// описание запроса запроса
type requestData[T any] struct {
	url             string
	headers, params map[string]string
	body            T
}

func makeRequestData[T any](url string, headers, params map[string]string, body T) requestData[T] {
	return requestData[T]{
		url:     url,
		headers: map[string]string{},
		params:  map[string]string{},
		body:    body,
	}
}

// отправляем запрос
func (r *requestData[T]) get() error {
	client := &http.Client{}
	reqURL, err := url.Parse(r.url)
	if err != nil {
		return err
	}

	query := reqURL.Query()
	for key, value := range r.params {
		query.Add(key, value)
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
