package AppStoreConnect

import (
	"bytes"
	"compress/gzip"
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

	apps            *AppsService
	reports         *ReportsService
	customerReviews *CustomerReviewsService

	builds                *BuildsService
	betaGroups            *BetaGroupsService
	betaTesterInvitations *BetaTesterInvitationsService
	bundleIDs             *BundleIDsService
	certificates          *CertificatesService
	profiles              *ProfilesService
	users                 *UsersService
	userInvitations       *UserInvitationsService

	appStoreVersions             *AppStoreVersionsService
	appStoreVersionSubmissions   *AppStoreVersionSubmissionsService
	appStoreVersionLocalizations *AppStoreVersionLocalizationsService
	appScreenshotSets            *AppScreenshotSetsService
	appScreenshots               *AppScreenshotsService
	inAppPurchases               *InAppPurchasesService
	subscriptionGroups           *SubscriptionGroupsService
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
	s.reports = &ReportsService{svc: s}
	s.customerReviews = &CustomerReviewsService{svc: s}
	s.builds = &BuildsService{svc: s}
	s.betaGroups = &BetaGroupsService{svc: s}
	s.betaTesterInvitations = &BetaTesterInvitationsService{svc: s}
	s.bundleIDs = &BundleIDsService{svc: s}
	s.certificates = &CertificatesService{svc: s}
	s.profiles = &ProfilesService{svc: s}
	s.users = &UsersService{svc: s}
	s.userInvitations = &UserInvitationsService{svc: s}
	s.appStoreVersions = &AppStoreVersionsService{svc: s}
	s.appStoreVersionSubmissions = &AppStoreVersionSubmissionsService{svc: s}
	s.appStoreVersionLocalizations = &AppStoreVersionLocalizationsService{svc: s}
	s.appScreenshotSets = &AppScreenshotSetsService{svc: s}
	s.appScreenshots = &AppScreenshotsService{svc: s}
	s.inAppPurchases = &InAppPurchasesService{svc: s}
	s.subscriptionGroups = &SubscriptionGroupsService{svc: s}
	return s
}

// Apps returns the Apps sub-service for managing apps on App Store Connect.
// See https://developer.apple.com/documentation/appstoreconnectapi/apps
func (s *Service) Apps() *AppsService { return s.apps }

// Reports returns the Reports sub-service for downloading sales and
// finance reports.
// See https://developer.apple.com/documentation/appstoreconnectapi/download_sales_and_trends_reports
func (s *Service) Reports() *ReportsService { return s.reports }

// CustomerReviews returns the CustomerReviews sub-service for reading
// customer reviews and posting developer responses.
// See https://developer.apple.com/documentation/appstoreconnectapi/customer_reviews
func (s *Service) CustomerReviews() *CustomerReviewsService { return s.customerReviews }

// Builds returns the Builds sub-service for TestFlight build management.
// See https://developer.apple.com/documentation/appstoreconnectapi/builds
func (s *Service) Builds() *BuildsService { return s.builds }

// BetaGroups returns the BetaGroups sub-service for managing TestFlight
// tester groups.
// See https://developer.apple.com/documentation/appstoreconnectapi/beta_groups
func (s *Service) BetaGroups() *BetaGroupsService { return s.betaGroups }

// BetaTesterInvitations returns the sub-service for sending TestFlight
// invitation emails.
// See https://developer.apple.com/documentation/appstoreconnectapi/beta_tester_invitations
func (s *Service) BetaTesterInvitations() *BetaTesterInvitationsService {
	return s.betaTesterInvitations
}

// BundleIDs returns the Bundle IDs sub-service.
// See https://developer.apple.com/documentation/appstoreconnectapi/bundle_ids
func (s *Service) BundleIDs() *BundleIDsService { return s.bundleIDs }

// Certificates returns the Certificates sub-service.
// See https://developer.apple.com/documentation/appstoreconnectapi/certificates
func (s *Service) Certificates() *CertificatesService { return s.certificates }

// Profiles returns the provisioning Profiles sub-service.
// See https://developer.apple.com/documentation/appstoreconnectapi/profiles
func (s *Service) Profiles() *ProfilesService { return s.profiles }

// Users returns the Users sub-service for managing team members.
// See https://developer.apple.com/documentation/appstoreconnectapi/users
func (s *Service) Users() *UsersService { return s.users }

// UserInvitations returns the UserInvitations sub-service for
// sending team invitations.
// See https://developer.apple.com/documentation/appstoreconnectapi/user_invitations
func (s *Service) UserInvitations() *UserInvitationsService { return s.userInvitations }

// AppStoreVersions returns the sub-service for App Store release
// catalog entries (one per shipping version of an app).
// See https://developer.apple.com/documentation/appstoreconnectapi/app_store_versions
func (s *Service) AppStoreVersions() *AppStoreVersionsService { return s.appStoreVersions }

// AppStoreVersionSubmissions returns the sub-service for submitting
// App Store versions for review (the programmatic "Submit for Review"
// button).
// See https://developer.apple.com/documentation/appstoreconnectapi/app_store_version_submissions
func (s *Service) AppStoreVersionSubmissions() *AppStoreVersionSubmissionsService {
	return s.appStoreVersionSubmissions
}

// AppStoreVersionLocalizations returns the sub-service for managing
// per-locale App Store metadata (description, keywords, what's new,
// marketing/support URLs, promotional text).
// See https://developer.apple.com/documentation/appstoreconnectapi/app_store_version_localizations
func (s *Service) AppStoreVersionLocalizations() *AppStoreVersionLocalizationsService {
	return s.appStoreVersionLocalizations
}

// AppScreenshotSets returns the sub-service for screenshot set
// containers (one per screenshotDisplayType per localization).
// See https://developer.apple.com/documentation/appstoreconnectapi/app_screenshot_sets
func (s *Service) AppScreenshotSets() *AppScreenshotSetsService { return s.appScreenshotSets }

// AppScreenshots returns the sub-service for uploading individual
// screenshots. Use [AppScreenshotsService.Upload] for the full
// reserve → PUT → commit flow from an in-memory buffer.
// See https://developer.apple.com/documentation/appstoreconnectapi/app_screenshots
func (s *Service) AppScreenshots() *AppScreenshotsService { return s.appScreenshots }

// InAppPurchases returns the sub-service for non-subscription in-app
// purchases under /v1/inAppPurchasesV2.
// See https://developer.apple.com/documentation/appstoreconnectapi/in-app_purchases_v2
func (s *Service) InAppPurchases() *InAppPurchasesService { return s.inAppPurchases }

// SubscriptionGroups returns the sub-service for auto-renewable
// subscription group management.
// See https://developer.apple.com/documentation/appstoreconnectapi/subscription_groups
func (s *Service) SubscriptionGroups() *SubscriptionGroupsService { return s.subscriptionGroups }

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

// doRaw performs an HTTP request and returns the raw (optionally
// gunzipped) response body. It is used by endpoints that return binary
// or non-JSON payloads — notably the sales and finance report
// endpoints, which serve gzipped TSV under Content-Type
// "application/a-gzip".
//
// If the server responds with Content-Encoding: gzip, Go's http package
// already transparently decodes the stream. If the response is instead
// a gzip *payload* (Apple's reports) — identified by Content-Type —
// doRaw gunzips it before returning. Callers can therefore treat the
// returned bytes as plain TSV.
//
// Non-2xx responses are parsed via [parseErrorBody] into an [APIError].
// Report endpoints serve JSON error bodies on failure, not gzipped
// ones, so this treatment is correct.
func (s *Service) doRaw(ctx context.Context, method, path string, query *Query) (*http.Response, []byte, error) {
	reqURL, err := s.resolveURL(path, query)
	if err != nil {
		return nil, nil, &ClientError{Message: "invalid request URL", Cause: err}
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, nil, &ClientError{Message: "build request", Cause: err}
	}
	// Apple's report endpoints require this Accept header to serve
	// the gzipped TSV payload.
	req.Header.Set("Accept", "application/a-gzip, application/json")
	if s.userAgent != "" {
		req.Header.Set("User-Agent", s.userAgent)
	}
	if err := s.authorizer.Authorize(req); err != nil {
		return nil, nil, &ClientError{Message: "authorize request", Cause: err}
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, nil, &ClientError{Message: "execute request", Cause: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, &ClientError{Message: "read response body", Cause: err}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, nil, parseErrorBody(resp.StatusCode, respBody)
	}

	// Apple reports ship the gzip as the payload itself — not as
	// transport encoding — so http.Client will not decode it. Detect
	// via Content-Type (application/a-gzip) or a magic-number sniff
	// so we can transparently hand the caller plain TSV bytes.
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "gzip") || isGzipped(respBody) {
		zr, err := gzip.NewReader(bytes.NewReader(respBody))
		if err != nil {
			return resp, nil, &ClientError{Message: "open gzip stream", Cause: err}
		}
		defer zr.Close()
		decoded, err := io.ReadAll(zr)
		if err != nil {
			return resp, nil, &ClientError{Message: "decompress gzip stream", Cause: err}
		}
		respBody = decoded
	}
	return resp, respBody, nil
}

// isGzipped returns true if the given buffer begins with the gzip magic
// number (0x1f 0x8b). Used as a fallback when Content-Type is missing
// or lies about the payload format.
func isGzipped(b []byte) bool {
	return len(b) >= 2 && b[0] == 0x1f && b[1] == 0x8b
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
