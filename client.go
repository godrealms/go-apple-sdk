package Apple

import (
	"github.com/go-resty/resty/v2"
)

// AppleClient defines the type of Apple service client
type AppleClient string

const (
	AppStoreConnectClient             AppleClient = "AppStoreConnectAPI"
	AppStoreServerClient              AppleClient = "AppStoreServerAPI"
	AppStoreServerNotificationsClient AppleClient = "AppStoreServerNotifications"
)

// ClientOption defines function type for client configuration
type ClientOption func(*Client)

// Middleware defines function type for request middleware
type Middleware func(*resty.Request) error

// RequestOption defines function type for request configuration
type RequestOption func(*resty.Request)

// Client represents the main client structure for Apple services
type Client struct {
	config      *Config
	service     AppleClient
	httpclient  *resty.Client
	middlewares []Middleware
}

// RequestParams contains all possible parameters for making a request
type RequestParams struct {
	Method      string            // HTTP method (GET, POST, etc.)
	Path        string            // Request path
	Body        any               // Request body
	Result      any               // Response result
	QueryParams map[string]string // URL query parameters
	Headers     map[string]string // HTTP headers
	PathParams  map[string]string // URL path parameters
	FormData    map[string]string // Form data
	Files       map[string]string // Files to upload (key: field name, value: file path)
}

// NewClient creates a new instance of the Apple service client
func NewClient(config *Config, opts ...ClientOption) *Client {
	client := &Client{
		config:      config,
		middlewares: make([]Middleware, 0),
	}

	// Initialize base HTTP client
	client.resetHttpClient()

	// Apply custom configurations
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// resetHttpClient reinitializes the HTTP client with base configuration
func (client *Client) resetHttpClient() {
	client.httpclient = resty.New().
		SetBaseURL(client.config.BaseUrl).
		SetTimeout(client.config.Timeout).
		SetRetryCount(client.config.RetryCount).
		SetRetryWaitTime(client.config.RetryWaitTime).
		SetRetryMaxWaitTime(client.config.RetryMaxWaitTime)
}

// setupServiceHandlers configures service-specific request handlers
func (client *Client) setupServiceHandlers(service AppleClient) {
	handler := client.getServiceHandler(service)
	if handler != nil {
		client.httpclient.OnBeforeRequest(handler)
	}
}

// getServiceHandler returns the appropriate handler for the specified service
func (client *Client) getServiceHandler(service AppleClient) resty.RequestMiddleware {
	switch service {
	case AppStoreConnectClient:
		return func(c *resty.Client, req *resty.Request) error {
			return client.handleAppStoreConnect(req)
		}
	case AppStoreServerClient:
		return func(c *resty.Client, req *resty.Request) error {
			return client.handleAppStoreServer(req)
		}
	case AppStoreServerNotificationsClient:
		return func(c *resty.Client, req *resty.Request) error {
			return client.handleAppStoreNotifications(req)
		}
	default:
		return nil
	}
}

// SetService sets the current service type and configures appropriate handlers
func (client *Client) SetService(service AppleClient) *Client {
	if client.service == service {
		return client
	}

	client.service = service
	client.resetHttpClient()
	client.setupServiceHandlers(service)

	return client
}

// Service-specific handlers
func (client *Client) handleAppStoreConnect(req *resty.Request) error {
	// TODO: Implement AppStoreConnect authentication and request handling
	return nil
}

func (client *Client) handleAppStoreServer(req *resty.Request) error {
	// TODO: Implement AppStoreServer authentication and request handling
	return nil
}

func (client *Client) handleAppStoreNotifications(req *resty.Request) error {
	// TODO: Implement AppStoreNotifications authentication and request handling
	return nil
}

// Request is the main method for making HTTP requests
func (client *Client) Request(params RequestParams, opts ...RequestOption) error {
	req := client.httpclient.R()

	// Set request body and response result
	if params.Body != nil {
		req.SetBody(params.Body)
	}
	if params.Result != nil {
		req.SetResult(params.Result)
	}

	// Set query parameters
	if len(params.QueryParams) > 0 {
		req.SetQueryParams(params.QueryParams)
	}

	// Set headers
	if len(params.Headers) > 0 {
		req.SetHeaders(params.Headers)
	}

	// Set path parameters
	if len(params.PathParams) > 0 {
		req.SetPathParams(params.PathParams)
	}

	// Set form data
	if len(params.FormData) > 0 {
		req.SetFormData(params.FormData)
	}

	// Set files for upload
	if len(params.Files) > 0 {
		for field, filePath := range params.Files {
			req.SetFile(field, filePath)
		}
	}

	// Apply custom request options
	for _, opt := range opts {
		opt(req)
	}

	// Execute middlewares
	for _, middleware := range client.middlewares {
		if err := middleware(req); err != nil {
			return err
		}
	}

	// Execute request
	resp, err := req.Execute(params.Method, params.Path)
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return client.handleError(resp)
	}

	return nil
}

// Request option helpers
func WithHeader(key, value string) RequestOption {
	return func(req *resty.Request) {
		req.SetHeader(key, value)
	}
}

func WithQueryParam(key, value string) RequestOption {
	return func(req *resty.Request) {
		req.SetQueryParam(key, value)
	}
}

func WithPathParam(key, value string) RequestOption {
	return func(req *resty.Request) {
		req.SetPathParam(key, value)
	}
}

// Convenience methods for common HTTP operations
func (client *Client) Get(path string, result interface{}, queryParams map[string]string) error {
	return client.Request(RequestParams{
		Method:      "GET",
		Path:        path,
		Result:      result,
		QueryParams: queryParams,
	})
}

func (client *Client) Post(path string, body interface{}, result interface{}) error {
	return client.Request(RequestParams{
		Method: "POST",
		Path:   path,
		Body:   body,
		Result: result,
	})
}

func (client *Client) Put(path string, body interface{}, result interface{}) error {
	return client.Request(RequestParams{
		Method: "PUT",
		Path:   path,
		Body:   body,
		Result: result,
	})
}

func (client *Client) Delete(path string) error {
	return client.Request(RequestParams{
		Method: "DELETE",
		Path:   path,
	})
}

// handleError processes request errors
func (client *Client) handleError(resp *resty.Response) error {
	// TODO: Implement error handling logic
	return nil
}

// Usage examples
func Example() {
	// Initialize client
	config := NewConfig("https://api.apple.com", "kid", "iss", "bid")
	client := NewClient(config)

	// Example GET request with query parameters
	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"data"`
	}

	err := client.Request(RequestParams{
		Method: "GET",
		Path:   "/v1/apps",
		QueryParams: map[string]string{
			"limit":  "20",
			"fields": "bundleId,name",
		},
		Result: &result,
	})

	// Example POST request with file upload
	err = client.Request(RequestParams{
		Method: "POST",
		Path:   "/v1/apps/{app_id}/screenshots",
		PathParams: map[string]string{
			"app_id": "123456789",
		},
		FormData: map[string]string{
			"type": "APP_SCREENSHOT",
		},
		Files: map[string]string{
			"file": "/path/to/screenshot.png",
		},
	})

	if err != nil {
		// Handle error
	}
}
