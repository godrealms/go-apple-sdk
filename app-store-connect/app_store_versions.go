package AppStoreConnect

import (
	"context"
	"time"
)

// AppStoreVersionsService provides access to /v1/appStoreVersions,
// Apple's per-release catalog of App Store versions (not the binary
// builds — those live under /v1/builds).
//
// A version pairs a semantic version string with a platform and is
// linked to exactly one app. Once created it holds the release
// metadata, screenshots, and review submissions for that release.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/app_store_versions
type AppStoreVersionsService struct {
	svc *Service
}

// AppStoreVersion is a typed alias for a JSON:API appStoreVersions resource.
type AppStoreVersion = Resource[AppStoreVersionAttributes]

// AppStoreVersionAttributes models the attributes of an appStoreVersions
// resource. Apple evolves this schema regularly; unknown fields are
// tolerated.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/appstoreversion/attributes
type AppStoreVersionAttributes struct {
	Platform            string     `json:"platform,omitempty"`
	VersionString       string     `json:"versionString,omitempty"`
	AppStoreState       string     `json:"appStoreState,omitempty"`
	Copyright           string     `json:"copyright,omitempty"`
	ReleaseType         string     `json:"releaseType,omitempty"`
	EarliestReleaseDate *time.Time `json:"earliestReleaseDate,omitempty"`
	Downloadable        *bool      `json:"downloadable,omitempty"`
	CreatedDate         *time.Time `json:"createdDate,omitempty"`
}

// ListAppStoreVersionsResponse is the decoded response for
// [AppStoreVersionsService.ListForApp].
type ListAppStoreVersionsResponse struct {
	Data     []AppStoreVersion `json:"data"`
	Included []Resource[any]   `json:"included,omitempty"`
	Links    *Links            `json:"links,omitempty"`
}

// ListForApp returns a page of App Store versions under the given app.
// Apple exposes version listings only under /v1/apps/{id}/appStoreVersions —
// there is no root collection.
func (s *AppStoreVersionsService) ListForApp(ctx context.Context, appID string, query *Query) (*ListAppStoreVersionsResponse, error) {
	if appID == "" {
		return nil, &ClientError{Message: "AppStoreVersions.ListForApp: appID is required"}
	}
	var doc Document[AppStoreVersionAttributes]
	path := "/v1/apps/" + appID + "/appStoreVersions"
	if _, err := s.svc.do(ctx, "GET", path, query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListAppStoreVersionsResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// ListForAppIterator returns a paginator that walks every page of
// versions for an app.
func (s *AppStoreVersionsService) ListForAppIterator(appID string, query *Query) *Paginator[AppStoreVersionAttributes] {
	return newPaginator[AppStoreVersionAttributes](s.svc, "/v1/apps/"+appID+"/appStoreVersions", query)
}

// Get fetches a single version by id.
func (s *AppStoreVersionsService) Get(ctx context.Context, id string, query *Query) (*AppStoreVersion, error) {
	if id == "" {
		return nil, &ClientError{Message: "AppStoreVersions.Get: id is required"}
	}
	var doc Document[AppStoreVersionAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/appStoreVersions/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// CreateAppStoreVersionRequest describes the attributes + required
// parent-app relationship for [AppStoreVersionsService.Create].
// Platform and VersionString are required; Copyright and ReleaseType
// are optional but commonly set at creation time.
type CreateAppStoreVersionRequest struct {
	AppID         string // required
	Platform      string // required — e.g. "IOS", "MAC_OS", "TV_OS"
	VersionString string // required — semantic version "1.2.3"
	Copyright     string
	ReleaseType   string // optional: "MANUAL", "AFTER_APPROVAL", "SCHEDULED"
}

// Create registers a new App Store version under the given app.
// See https://developer.apple.com/documentation/appstoreconnectapi/create_an_app_store_version
func (s *AppStoreVersionsService) Create(ctx context.Context, req CreateAppStoreVersionRequest) (*AppStoreVersion, error) {
	if req.AppID == "" || req.Platform == "" || req.VersionString == "" {
		return nil, &ClientError{Message: "AppStoreVersions.Create: AppID, Platform, and VersionString are required"}
	}
	attrs := map[string]any{
		"platform":      req.Platform,
		"versionString": req.VersionString,
	}
	if req.Copyright != "" {
		attrs["copyright"] = req.Copyright
	}
	if req.ReleaseType != "" {
		attrs["releaseType"] = req.ReleaseType
	}
	body := map[string]any{
		"data": map[string]any{
			"type":       "appStoreVersions",
			"attributes": attrs,
			"relationships": map[string]any{
				"app": map[string]any{
					"data": map[string]any{"type": "apps", "id": req.AppID},
				},
			},
		},
	}
	var doc Document[AppStoreVersionAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/appStoreVersions", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// AppStoreVersionUpdate collects the mutable attributes on an
// appStoreVersions resource. Nil-valued fields are omitted.
type AppStoreVersionUpdate struct {
	attrs map[string]any
}

// NewAppStoreVersionUpdate returns an empty update.
func NewAppStoreVersionUpdate() *AppStoreVersionUpdate {
	return &AppStoreVersionUpdate{attrs: make(map[string]any)}
}

// VersionString sets the semantic version string.
func (u *AppStoreVersionUpdate) VersionString(v string) *AppStoreVersionUpdate {
	u.attrs["versionString"] = v
	return u
}

// Copyright sets the copyright notice shown on the App Store listing.
func (u *AppStoreVersionUpdate) Copyright(v string) *AppStoreVersionUpdate {
	u.attrs["copyright"] = v
	return u
}

// ReleaseType sets the release strategy.
func (u *AppStoreVersionUpdate) ReleaseType(v string) *AppStoreVersionUpdate {
	u.attrs["releaseType"] = v
	return u
}

// EarliestReleaseDate sets the earliest date Apple may ship the
// version (used with ReleaseType "SCHEDULED").
func (u *AppStoreVersionUpdate) EarliestReleaseDate(v time.Time) *AppStoreVersionUpdate {
	u.attrs["earliestReleaseDate"] = v.UTC().Format(time.RFC3339)
	return u
}

// Downloadable toggles whether the version is downloadable after release.
func (u *AppStoreVersionUpdate) Downloadable(v bool) *AppStoreVersionUpdate {
	u.attrs["downloadable"] = v
	return u
}

// Set applies an arbitrary attribute key/value pair.
func (u *AppStoreVersionUpdate) Set(key string, value any) *AppStoreVersionUpdate {
	u.attrs[key] = value
	return u
}

// IsEmpty reports whether any attribute change is pending.
func (u *AppStoreVersionUpdate) IsEmpty() bool { return len(u.attrs) == 0 }

// Update modifies the given App Store version's attributes.
func (s *AppStoreVersionsService) Update(ctx context.Context, id string, update *AppStoreVersionUpdate) (*AppStoreVersion, error) {
	if id == "" {
		return nil, &ClientError{Message: "AppStoreVersions.Update: id is required"}
	}
	if update == nil || update.IsEmpty() {
		return nil, &ClientError{Message: "AppStoreVersions.Update: update has no attribute changes"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type":       "appStoreVersions",
			"id":         id,
			"attributes": update.attrs,
		},
	}
	var doc Document[AppStoreVersionAttributes]
	if _, err := s.svc.do(ctx, "PATCH", "/v1/appStoreVersions/"+id, nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete removes an App Store version.
func (s *AppStoreVersionsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "AppStoreVersions.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/appStoreVersions/"+id, nil, nil, nil)
	return err
}

// SelectBuild attaches the given build to an App Store version,
// linking the binary that will ship under that version.
// See https://developer.apple.com/documentation/appstoreconnectapi/modify_the_build_for_an_app_store_version
func (s *AppStoreVersionsService) SelectBuild(ctx context.Context, versionID, buildID string) error {
	if versionID == "" || buildID == "" {
		return &ClientError{Message: "AppStoreVersions.SelectBuild: versionID and buildID are required"}
	}
	body := map[string]any{
		"data": map[string]any{"type": "builds", "id": buildID},
	}
	_, err := s.svc.do(ctx, "PATCH", "/v1/appStoreVersions/"+versionID+"/relationships/build", nil, body, nil)
	return err
}
