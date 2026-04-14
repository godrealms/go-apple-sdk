package AppStoreConnect

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the production host for the App Store Connect API.
// Sandbox shares the same host; there is no separate sandbox endpoint.
const DefaultBaseURL = "https://api.appstoreconnect.apple.com"

// Config configures a [Service].
type Config struct {
	// BaseURL defaults to [DefaultBaseURL] when empty.
	BaseURL string
	// Authorizer signs outgoing requests. Required.
	Authorizer Authorizer
	// HTTPClient is used for transport. Defaults to a [http.Client]
	// with a 30 second timeout when nil.
	HTTPClient *http.Client
	// UserAgent, when non-empty, is set as the User-Agent request header.
	UserAgent string
}

// Service is the entry point into App Store Connect API endpoints.
// Create one with [New], or obtain one from the root SDK client via
// Client.AppStoreConnect().
//
// Service is safe for concurrent use by multiple goroutines.
type Service struct {
	baseURL    string
	httpClient *http.Client
	authorizer Authorizer
	userAgent  string

	apps *AppsService
}

// New constructs a [Service] with the given configuration.
// Authorizer must be non-nil.
func New(cfg Config) *Service {
	if cfg.Authorizer == nil {
		panic("AppStoreConnect.New: Authorizer is required")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	s := &Service{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
		authorizer: cfg.Authorizer,
		userAgent:  cfg.UserAgent,
	}
	s.apps = &AppsService{svc: s}
	return s
}

// Apps returns the Apps sub-service for managing apps on App Store Connect.
// See https://developer.apple.com/documentation/appstoreconnectapi/apps
func (s *Service) Apps() *AppsService { return s.apps }

// BaseURL returns the service's base URL (without trailing slash).
func (s *Service) BaseURL() string { return s.baseURL }

// do performs an HTTP request and decodes a JSON response into out.
// If the server returns a non-2xx status, do returns an [*APIError] parsed
// from the response body.
//
// path may be either a path relative to the base URL (e.g. "/v1/apps")
// or an absolute URL (e.g. an Apple pagination "next" link); both are
// supported so paginators can follow cursor links verbatim.
func (s *Service) do(ctx context.Context, method, path string, query *Query, body any, out any) (*http.Response, error) {
	reqURL, err := s.resolveURL(path, query)
	if err != nil {
		return nil, &ClientError{Message: "invalid request URL", Cause: err}
	}

	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, &ClientError{Message: "marshal request body", Cause: err}
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, &ClientError{Message: "build request", Cause: err}
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if s.userAgent != "" {
		req.Header.Set("User-Agent", s.userAgent)
	}
	if err := s.authorizer.Authorize(req); err != nil {
		return nil, &ClientError{Message: "authorize request", Cause: err}
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, &ClientError{Message: "execute request", Cause: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, &ClientError{Message: "read response body", Cause: err}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, parseErrorBody(resp.StatusCode, respBody)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return resp, &ClientError{Message: "decode response body", Cause: err}
		}
	}
	return resp, nil
}

// resolveURL converts a path plus optional query into an absolute URL.
// Accepts absolute URLs unchanged (for pagination link-following).
func (s *Service) resolveURL(path string, query *Query) (string, error) {
	var (
		u   *url.URL
		err error
	)
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		u, err = url.Parse(path)
	} else {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		u, err = url.Parse(s.baseURL + path)
	}
	if err != nil {
		return "", err
	}

	encoded := query.Encode()
	if encoded != "" {
		// Merge with any query already present on an absolute URL.
		if u.RawQuery == "" {
			u.RawQuery = encoded
		} else {
			u.RawQuery = u.RawQuery + "&" + encoded
		}
	}
	return u.String(), nil
}
