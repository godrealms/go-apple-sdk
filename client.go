package Apple

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/godrealms/go-apple-sdk/types"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"strings"
	"time"
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
	sandbox     bool
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
	QueryParams map[string]any    // URL query parameters
	Headers     map[string]string // HTTP headers
	PathParams  map[string]string // URL path parameters
	FormData    map[string]string // Form data
	Files       map[string]string // Files to upload (key: field name, value: file path)
}

// NewClient creates a new instance of the Apple service client
func NewClient(Sandbox bool, kid, iss, bid, privateKey string, opts ...ClientOption) *Client {
	client := &Client{
		sandbox:     Sandbox,
		config:      NewConfig(kid, iss, bid, privateKey),
		middlewares: make([]Middleware, 0),
	}

	if client.service != "" {
		// Initialize base HTTP client
		client.resetHttpClient()

		// Apply custom configurations
		for _, opt := range opts {
			opt(client)
		}
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
	switch service {
	case AppStoreConnectClient:
		client.config.BaseUrl = ""
		if client.sandbox {
			client.config.BaseUrl = ""
		}
	case AppStoreServerClient:
		client.config.BaseUrl = "https://api.appstoreconnect.apple.com"
		if client.sandbox {
			client.config.BaseUrl = "https://api.storekit-sandbox.itunes.apple.com"
		}
	case AppStoreServerNotificationsClient:
		client.config.BaseUrl = "https://api.appstoreconnect.apple.com"
		if client.sandbox {
			client.config.BaseUrl = "https://api.storekit-sandbox.itunes.apple.com"
		}
	}

	client.resetHttpClient()
	client.setupServiceHandlers(service)

	return client
}

// Service-specific handlers
func (client *Client) handleAppStoreConnect(req *resty.Request) error {
	// Implement AppStoreConnect authentication and request handling
	req.SetHeader("Authorization", client.GenerateAppStoreConnectAuthorizationJWT(req.Method, req.URL))
	return nil
}

func (client *Client) handleAppStoreServer(req *resty.Request) error {
	// Implement AppStoreServer authentication and request handling
	req.SetHeader("Authorization", client.GenerateAppStoreServerAuthorizationJWT())
	return nil
}

func (client *Client) handleAppStoreNotifications(req *resty.Request) error {
	// Implement AppStoreNotifications authentication and request handling
	req.SetHeader("Authorization", client.GenerateAppStoreServerAuthorizationJWT())
	return nil
}

func (client *Client) GenerateAppStoreServerAuthorizationJWT() string {
	privateKey, err := types.ParsePrivateKey(client.config.PrivateKey)
	if err != nil {
		log.Printf("failed to parse private key: %v", err)
		return ""
	}
	// 创建 JWT 的 Header 和 Claims
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": client.config.Iss,               // Apple Team ID
		"iat": now.Unix(),                      // CURRENT TIMESTAMP
		"exp": now.Add(5 * time.Minute).Unix(), // Expiration time (30 minutes)
		"aud": "appstoreconnect-v1",            // Fixed value appstoreconnect-v1
		"bid": client.config.Bid,
	}
	// 创建 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = client.config.Kid // Set Header's kid (key ID)

	// 使用私钥签名
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		log.Println("failed to sign token: ", err.Error())
		return ""
	}
	return fmt.Sprintf("Bearer %s", signedToken)
}

func (client *Client) GenerateAppStoreConnectAuthorizationJWT(method string, endpoint string) string {
	privateKey, err := types.ParsePrivateKey(client.config.PrivateKey)
	if err != nil {
		log.Printf("failed to parse private key: %v", err)
		return ""
	}
	// 创建 JWT 的 Header 和 Claims
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": client.config.Iss,               // Apple Team ID
		"iat": now.Unix(),                      // CURRENT TIMESTAMP
		"exp": now.Add(5 * time.Minute).Unix(), // Expiration time (30 minutes)
		"aud": "appstoreconnect-v1",            // Fixed value appstoreconnect-v1
		"scope": []string{
			fmt.Sprintf("%s %s", method, endpoint),
		},
	}
	// 创建 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = client.config.Kid // Set Header's kid (key ID)

	// 使用私钥签名
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		log.Println("failed to sign token: ", err.Error())
		return ""
	}
	return fmt.Sprintf("Bearer %s", signedToken)
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
		for k, v := range params.QueryParams {
			switch val := v.(type) {
			case string:
				req.SetQueryParam(k, val)
			case bool:
				req.SetQueryParam(k, fmt.Sprintf("%v", val))
			case int, int8, int16, int32, int64:
				req.SetQueryParam(k, fmt.Sprintf("%d", val))
			case float32, float64:
				req.SetQueryParam(k, fmt.Sprintf("%v", val))
			case []string:
				for _, item := range val {
					req.SetQueryParam(k, item)
				}
			default:
				// 对于其他类型，尝试使用 json.Marshal
				if jsonStr, err := json.Marshal(val); err == nil {
					req.SetQueryParam(k, string(jsonStr))
				}
			}
		}
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
func (client *Client) Get(path string, result interface{}, queryParams map[string]any) error {
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
	// 获取请求信息
	req := resp.Request.RawRequest

	// 构建详细的错误日志
	logInfo := []struct {
		Key   string
		Value interface{}
	}{
		{"Status Code", resp.StatusCode()},
		{"Request URL", req.URL.String()},
		{"Request Method", req.Method},
		{"Request Headers", req.Header},
		{"Response Time", resp.Time()},
		{"Response Headers", resp.Header()},
		{"Response Size", len(resp.Body())},
	}

	// 使用 strings.Builder 构建日志信息
	var logMsg strings.Builder
	logMsg.WriteString("\n=== Error Response Details ===\n")

	for _, info := range logInfo {
		logMsg.WriteString(fmt.Sprintf("%-20s: %v\n", info.Key, info.Value))
	}

	// 尝试解析响应体为 JSON 并格式化
	var prettyJSON bytes.Buffer
	if json.Valid(resp.Body()) {
		if err := json.Indent(&prettyJSON, resp.Body(), "", "  "); err == nil {
			logMsg.WriteString("\nResponse Body (JSON):\n")
			logMsg.Write(prettyJSON.Bytes())
		} else {
			// 如果不是有效的 JSON，直接打印原始响应
			logMsg.WriteString("\nResponse Body (Raw):\n")
			logMsg.Write(resp.Body())
		}
	} else {
		logMsg.WriteString("\nResponse Body (Raw):\n")
		logMsg.WriteString(string(resp.Body()))
	}

	logMsg.WriteString("\n=== End Error Response ===\n")

	// 打印完整的日志信息
	log.Println(logMsg.String())

	// 如果响应包含错误信息，尝试解析并返回结构化错误
	if resp.IsError() {
		var apiError struct {
			ErrorCode    string `json:"ErrorCode"`
			ErrorMessage string `json:"errorMessage"`
			Details      any    `json:"details,omitempty"`
		}

		if err := json.Unmarshal(resp.Body(), &apiError); err == nil {
			return fmt.Errorf("API Error - Code: %s, Message: %s, Details: %+v",
				apiError.ErrorCode, apiError.ErrorMessage, apiError.Details)
		}

		// 如果无法解析为标准错误格式，返回HTTP状态码和原始响应
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}
