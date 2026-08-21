package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client is a thin HTTP wrapper around resty for internal service calls.
type Client struct {
	r       *resty.Client
	token   string
}

// UpstreamError captures the HTTP status code returned by a downstream
// service so callers can propagate domain-specific statuses (e.g. 404).
type UpstreamError struct {
	StatusCode int
	Method     string
	Path       string
	Body       string
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("%s %s returned %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

// New creates a Client pointing to baseURL with an optional internal auth token.
func New(baseURL, token string) *Client {
	r := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(5 * time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	if token != "" {
		r.SetHeader("X-Internal-Token", token)
	}

	return &Client{r: r, token: token}
}

// Get performs a GET request and decodes JSON into out.
func (c *Client) Get(ctx context.Context, path string, out any) (*http.Response, error) {
	resp, err := c.r.R().SetContext(ctx).SetResult(out).Get(path)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return resp.RawResponse, &UpstreamError{StatusCode: resp.StatusCode(), Method: http.MethodGet, Path: path, Body: string(resp.Body())}
	}
	return resp.RawResponse, nil
}

// GetQuery performs a GET request with query params and decodes JSON into out.
func (c *Client) GetQuery(ctx context.Context, path string, query map[string]string, out any) (*http.Response, error) {
	resp, err := c.r.R().SetContext(ctx).SetQueryParams(query).SetResult(out).Get(path)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return resp.RawResponse, &UpstreamError{StatusCode: resp.StatusCode(), Method: http.MethodGet, Path: path, Body: string(resp.Body())}
	}
	return resp.RawResponse, nil
}

// Post performs a POST request with a JSON body and decodes the response into out.
func (c *Client) Post(ctx context.Context, path string, body, out any) (*http.Response, error) {
	resp, err := c.r.R().SetContext(ctx).SetBody(body).SetResult(out).Post(path)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return resp.RawResponse, &UpstreamError{StatusCode: resp.StatusCode(), Method: http.MethodPost, Path: path, Body: string(resp.Body())}
	}
	return resp.RawResponse, nil
}

// DecodeError is a helper to format upstream error bodies.
func DecodeError(body []byte) string {
	var m map[string]any
	if err := json.Unmarshal(body, &m); err == nil {
		if msg, ok := m["error"].(string); ok && msg != "" {
			return msg
		}
		if msg, ok := m["message"].(string); ok && msg != "" {
			return msg
		}
	}
	return string(body)
}
