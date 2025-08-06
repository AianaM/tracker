package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	client "example.com/tracker/internal/client"
)

func TestRequestNew(t *testing.T) {
	type args struct {
		ctx        context.Context
		httpMethod string
		url        string
		body       io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *http.Request
		wantErr bool
	}{
		{
			name: "Valid GET request",
			args: args{
				ctx:        context.Background(),
				httpMethod: http.MethodGet,
				url:        "http://example.com",
				body:       nil,
			},
			want: &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Scheme: "http", Host: "example.com"},
				Header: http.Header{"Content-Type": []string{"application/json"}},
			},
			wantErr: false,
		},
		{
			name: "Invalid URL",
			args: args{
				ctx:        context.Background(),
				httpMethod: http.MethodGet,
				url:        "http://invalid-url\\",
				body:       nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Valid POST request with body",
			args: args{
				ctx:        context.Background(),
				httpMethod: http.MethodPost,
				url:        "http://example.com",
				body:       strings.NewReader(`{"key":"value"}`),
			},
			want: &http.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Scheme: "http", Host: "example.com"},
				Header: http.Header{"Content-Type": []string{"application/json"}},
			},
			wantErr: false,
		},
		{
			name: "Valid PUT request",
			args: args{
				ctx:        context.Background(),
				httpMethod: http.MethodPut,
				url:        "https://api.example.com/users/1",
				body:       bytes.NewReader([]byte(`{"name":"test"}`)),
			},
			want: &http.Request{
				Method: http.MethodPut,
				URL:    &url.URL{Scheme: "https", Host: "api.example.com", Path: "/users/1"},
				Header: http.Header{"Content-Type": []string{"application/json"}},
			},
			wantErr: false,
		},
		{
			name: "Valid DELETE request",
			args: args{
				ctx:        context.Background(),
				httpMethod: http.MethodDelete,
				url:        "https://api.example.com/users/1",
				body:       nil,
			},
			want: &http.Request{
				Method: http.MethodDelete,
				URL:    &url.URL{Scheme: "https", Host: "api.example.com", Path: "/users/1"},
				Header: http.Header{"Content-Type": []string{"application/json"}},
			},
			wantErr: false,
		},
		{
			name: "Cancelled context",
			args: args{
				ctx:        func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
				httpMethod: http.MethodGet,
				url:        "http://example.com",
				body:       nil,
			},
			want: &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Scheme: "http", Host: "example.com"},
				Header: http.Header{"Content-Type": []string{"application/json"}},
			},
			wantErr: false,
		},
	}

	httpClient := client.New(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := httpClient.NewRequest(tt.args.ctx, tt.args.httpMethod, tt.args.url, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestNew() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got != nil {
				if got.Method != tt.want.Method {
					t.Errorf("RequestNew() Method = %v, want %v", got.Method, tt.want.Method)
				}
				if got.URL.String() != tt.want.URL.String() {
					t.Errorf("RequestNew() URL = %v, want %v", got.URL.String(), tt.want.URL.String())
				}
				if got.Header.Get("Content-Type") != "application/json" {
					t.Errorf("RequestNew() Content-Type header not set correctly")
				}
			}
		})
	}
}

func TestExecRequest(t *testing.T) {
	// Создаем тестовый HTTP сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		case "/error":
			w.WriteHeader(http.StatusBadRequest)
		case "/timeout":
			time.Sleep(15 * time.Second) // Превышаем таймаут клиента
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name          string
		path          string
		bodyInterface interface{}
		wantStatus    int
		wantErr       bool
	}{
		{
			name:          "Successful request",
			path:          "/success",
			bodyInterface: &map[string]string{},
			wantStatus:    http.StatusOK,
			wantErr:       false,
		},
		{
			name:          "Error request",
			path:          "/error",
			bodyInterface: nil,
			wantStatus:    http.StatusBadRequest,
			wantErr:       true,
		},
		{
			name:          "Not found request",
			path:          "/notfound",
			bodyInterface: nil,
			wantStatus:    http.StatusNotFound,
			wantErr:       true,
		},
	}

	httpClient := client.New(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, server.URL+tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			gotStatus, err := httpClient.Do(req, tt.bodyInterface)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("ExecRequest() = %v, want %v", gotStatus, tt.wantStatus)
			}
		})
	}
}
