package Apple

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	AppStoreConnect "github.com/godrealms/go-apple-sdk/v2/app-store-connect"
	"github.com/godrealms/go-apple-sdk/v2/types"
	"github.com/golang-jwt/jwt/v5"
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

// Client represents the main client structure for Apple services.
//
// A Client holds one pre-built, immutable resty client per Apple service
// (App Store Connect / Server / Server Notifications), each with its own base
// URL and JWT auth middleware. SetService only records which service the
// legacy Request path should target; it never rebuilds transports, so the
// connection pools are never thrown away.
//
// The per-service clients are safe for concurrent use. SetService and Request
// synchronize access to the current-service selector, so calling them from
// multiple goroutines will not corrupt state. Note, however, that SetService
// mutates a single shared selector: if one goroutine needs service A while
// another needs service B, give each goroutine its own Client (or use the
// stateless AppStoreConnect service via [Client.AppStoreConnect]).
type Client struct {
	sandbox          bool
	config           *Config
	mu               sync.RWMutex
	service          AppleClient
	clients          map[AppleClient]*resty.Client // built once in NewClient; read-only thereafter
	baseURLOverrides map[AppleClient]string        // optional per-service base URL overrides
	middlewares      []Middleware
}

// WithServiceBaseURL overrides the base URL used for a specific service. This
// is useful for pointing the client at a mock/proxy server or a non-default
// Apple host in tests. Apply it via [NewClient].
func WithServiceBaseURL(service AppleClient, baseURL string) ClientOption {
	return func(c *Client) {
		if c.baseURLOverrides == nil {
			c.baseURLOverrides = make(map[AppleClient]string)
		}
		c.baseURLOverrides[service] = baseURL
	}
}

// RequestParams contains all possible parameters for making a request
type RequestParams struct {
	Ctx         context.Context   // Optional request context (timeout / cancellation). nil → context.Background.
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

// NewClient creates a new instance of the Apple service client.
//
// The returned client is fully initialized: it has a working
// resty.Client with sane defaults (timeout, retries from
// [Config]) and any caller-supplied [ClientOption]s applied.
// Service-specific configuration (base URL, JWT middleware) is
// installed on demand via [Client.SetService].
func NewClient(Sandbox bool, kid, iss, bid, privateKey string, opts ...ClientOption) *Client {
	client := &Client{
		sandbox:     Sandbox,
		config:      NewConfig(kid, iss, bid, privateKey),
		middlewares: make([]Middleware, 0),
	}
	// Apply caller options first (they may adjust config), then build the
	// per-service resty clients from the final config exactly once.
	for _, opt := range opts {
		opt(client)
	}
	client.clients = client.buildClients()
	return client
}

// baseURLFor returns the correct base URL for the given service, honoring the
// client's sandbox flag for the storekit-hosted services.
func (client *Client) baseURLFor(service AppleClient) string {
	if u, ok := client.baseURLOverrides[service]; ok && u != "" {
		return u
	}
	switch service {
	case AppStoreConnectClient:
		// App Store Connect API shares a single host for production and sandbox.
		// See https://developer.apple.com/documentation/appstoreconnectapi
		return "https://api.appstoreconnect.apple.com"
	case AppStoreServerClient, AppStoreServerNotificationsClient:
		// App Store Server API and Server Notifications V2 use the same host
		// family (storekit.itunes.apple.com / storekit-sandbox).
		if client.sandbox {
			return "https://api.storekit-sandbox.itunes.apple.com"
		}
		return "https://api.storekit.itunes.apple.com"
	default:
		return ""
	}
}

// buildClients constructs one resty client per service, each pinned to its
// own base URL and JWT auth middleware. Called once from NewClient; the
// returned map is treated as read-only afterwards so it needs no locking.
func (client *Client) buildClients() map[AppleClient]*resty.Client {
	newServiceClient := func(service AppleClient, handler func(*resty.Request) error) *resty.Client {
		rc := resty.New().
			SetBaseURL(client.baseURLFor(service)).
			SetTimeout(client.config.Timeout).
			SetRetryCount(client.config.RetryCount).
			SetRetryWaitTime(client.config.RetryWaitTime).
			SetRetryMaxWaitTime(client.config.RetryMaxWaitTime)
		rc.OnBeforeRequest(func(_ *resty.Client, req *resty.Request) error {
			return handler(req)
		})
		return rc
	}
	return map[AppleClient]*resty.Client{
		AppStoreConnectClient:             newServiceClient(AppStoreConnectClient, client.handleAppStoreConnect),
		AppStoreServerClient:              newServiceClient(AppStoreServerClient, client.handleAppStoreServer),
		AppStoreServerNotificationsClient: newServiceClient(AppStoreServerNotificationsClient, client.handleAppStoreNotifications),
	}
}

// SetService records which service the legacy [Client.Request] path targets.
// It no longer rebuilds transports (the per-service clients are pre-built in
// NewClient); it only updates a mutex-guarded selector, so it is safe to call
// concurrently. See the [Client] doc comment for the cross-service caveat.
func (client *Client) SetService(service AppleClient) *Client {
	client.mu.Lock()
	client.service = service
	client.mu.Unlock()
	return client
}

// Service-specific handlers
func (client *Client) handleAppStoreConnect(req *resty.Request) error {
	auth, err := client.GenerateAppStoreConnectAuthorizationJWT(req.Method, req.URL)
	if err != nil {
		return fmt.Errorf("appstoreconnect authorization: %w", err)
	}
	req.SetHeader("Authorization", auth)
	return nil
}

func (client *Client) handleAppStoreServer(req *resty.Request) error {
	auth, err := client.GenerateAppStoreServerAuthorizationJWT()
	if err != nil {
		return fmt.Errorf("appstoreserver authorization: %w", err)
	}
	req.SetHeader("Authorization", auth)
	return nil
}

func (client *Client) handleAppStoreNotifications(req *resty.Request) error {
	auth, err := client.GenerateAppStoreServerAuthorizationJWT()
	if err != nil {
		return fmt.Errorf("appstoreservernotifications authorization: %w", err)
	}
	req.SetHeader("Authorization", auth)
	return nil
}

// GenerateAppStoreServerAuthorizationJWT builds an "Bearer …" JWT
// header value for App Store Server API + Server Notifications
// requests. Returns a wrapped error on private-key parse failure
// or signing failure rather than silently producing an empty
// header (the previous behavior, which led to mysterious 401s
// from Apple instead of clear local errors).
func (client *Client) GenerateAppStoreServerAuthorizationJWT() (string, error) {
	privateKey, err := types.ParsePrivateKey(client.config.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": client.config.Iss,
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"aud": "appstoreconnect-v1",
		"bid": client.config.Bid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = client.config.Kid
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return fmt.Sprintf("Bearer %s", signedToken), nil
}

// GenerateAppStoreConnectAuthorizationJWT builds a "Bearer …" JWT
// header value scoped to a single (method, endpoint) pair. Returns
// a wrapped error rather than an empty string on failure.
//
// Most callers should use [Client.AppStoreConnect] instead, which
// goes through the new App Store Connect [Service] with an
// unscoped JWT — Apple permits 20-minute unscoped tokens, and the
// scoped variant produced here can cause unexpected 401s when
// callers pass the URL with query parameters that Apple's
// authoriser does not normalise the same way the SDK does.
func (client *Client) GenerateAppStoreConnectAuthorizationJWT(method string, endpoint string) (string, error) {
	privateKey, err := types.ParsePrivateKey(client.config.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": client.config.Iss,
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"aud": "appstoreconnect-v1",
		"scope": []string{
			fmt.Sprintf("%s %s", method, endpoint),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = client.config.Kid
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return fmt.Sprintf("Bearer %s", signedToken), nil
}

// Request is the main method for making HTTP requests
func (client *Client) Request(params RequestParams, opts ...RequestOption) error {
	client.mu.RLock()
	httpclient := client.clients[client.service]
	client.mu.RUnlock()
	if httpclient == nil {
		return fmt.Errorf("no service selected: call SetService before Request")
	}
	req := httpclient.R()

	// Attach context for timeout / cancellation propagation.
	// Defaults to context.Background when callers leave Ctx nil.
	if params.Ctx != nil {
		req.SetContext(params.Ctx)
	}

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
				// Append each value under the same key. SetQueryParam
				// uses Set (overwrite), which would keep only the last
				// value; Add preserves all of them (status=1&status=2…).
				for _, item := range val {
					req.QueryParam.Add(k, item)
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
		{"Request Headers", redactSensitiveHeaders(req.Header)},
		{"Response Time", resp.Time()},
		{"Response Headers", resp.Header()},
		{"Response Size", len(resp.Body())},
	}

	// 使用 strings.Builder 构建日志信息
	var logMsg strings.Builder
	logMsg.WriteString("\n=== Error Response Details ===\n")

	for _, info := range logInfo {
		fmt.Fprintf(&logMsg, "%-20s: %v\n", info.Key, info.Value)
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

	// Any non-2xx status that is not a >=400 error (e.g. an unexpected 3xx
	// from a proxy or misconfigured redirect) still reached here because
	// Request only calls handleError when IsSuccess() is false. Returning
	// nil would make the caller believe the request succeeded while Result
	// was never populated, so surface an explicit error instead.
	return fmt.Errorf("unexpected non-success HTTP %d: %s",
		resp.StatusCode(), strings.TrimSpace(string(resp.Body())))
}

// redactSensitiveHeaders returns a copy of h with credential-bearing headers
// masked, so diagnostic logging never writes bearer tokens or cookies to the
// process log.
func redactSensitiveHeaders(h http.Header) http.Header {
	clone := h.Clone()
	if clone == nil {
		return h
	}
	for _, key := range []string{"Authorization", "Cookie", "Proxy-Authorization", "Set-Cookie"} {
		if clone.Get(key) != "" {
			clone.Set(key, "[REDACTED]")
		}
	}
	return clone
}

// AppStoreConnect returns a service for calling the App Store Connect API.
// It is a thin factory that injects this client's JWT signer as the
// authorizer on every outgoing request. Each call returns a fresh
// service; the service is safe for concurrent use.
//
// Example:
//
//	svc := client.AppStoreConnect()
//	apps, err := svc.Apps().List(ctx, AppStoreConnect.NewQuery().Limit(200))
//
// See https://developer.apple.com/documentation/appstoreconnectapi for
// a full catalog of available endpoints.
func (client *Client) AppStoreConnect() *AppStoreConnect.Service {
	return AppStoreConnect.New(AppStoreConnect.Config{
		BaseURL:   "https://api.appstoreconnect.apple.com",
		UserAgent: "go-apple-sdk",
		Authorizer: AppStoreConnect.AuthorizerFunc(func(req *http.Request) error {
			token, err := client.signAppStoreConnectToken()
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+token)
			return nil
		}),
	})
}

// signAppStoreConnectToken builds an unscoped App Store Connect JWT that
// is valid for any endpoint. Apple permits tokens up to 20 minutes old;
// we mint a fresh one on every request rather than caching, which keeps
// the signer simple and stateless at the cost of a few extra signatures
// per call.
//
// Distinct from [Client.GenerateAppStoreConnectAuthorizationJWT], which
// includes a scoped "scope" claim tied to a specific method+endpoint.
// Scoped tokens are stricter than needed for most endpoints and can
// cause unexpected 401s, so the new AppStoreConnect.Service path uses
// this unscoped variant instead.
func (client *Client) signAppStoreConnectToken() (string, error) {
	privateKey, err := types.ParsePrivateKey(client.config.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": client.config.Iss,
		"iat": now.Unix(),
		"exp": now.Add(20 * time.Minute).Unix(),
		"aud": "appstoreconnect-v1",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = client.config.Kid
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}
