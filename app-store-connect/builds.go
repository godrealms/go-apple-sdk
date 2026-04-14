package AppStoreConnect

import (
	"context"
	"time"
)

// BuildsService provides access to /v1/builds, Apple's TestFlight
// build catalog. A "build" is a single uploaded binary (.ipa) that may
// or may not be processed, distributed to testers, or submitted for
// review.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/builds
type BuildsService struct {
	svc *Service
}

// Build is a typed alias for a JSON:API builds resource.
type Build = Resource[BuildAttributes]

// BuildAttributes models the attributes of a builds resource. Apple
// evolves this schema over time; unknown fields are tolerated.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/build/attributes
type BuildAttributes struct {
	Version                  string     `json:"version,omitempty"`
	UploadedDate             *time.Time `json:"uploadedDate,omitempty"`
	ExpirationDate           *time.Time `json:"expirationDate,omitempty"`
	Expired                  *bool      `json:"expired,omitempty"`
	MinOsVersion             string     `json:"minOsVersion,omitempty"`
	ProcessingState          string     `json:"processingState,omitempty"`
	BuildAudienceType        string     `json:"buildAudienceType,omitempty"`
	UsesNonExemptEncryption  *bool      `json:"usesNonExemptEncryption,omitempty"`
	IconAssetToken           any        `json:"iconAssetToken,omitempty"`
}

// ListBuildsResponse is the decoded response for [BuildsService.List].
type ListBuildsResponse struct {
	Data     []Build         `json:"data"`
	Included []Resource[any] `json:"included,omitempty"`
	Links    *Links          `json:"links,omitempty"`
}

// List returns a page of builds matching the query.
// See https://developer.apple.com/documentation/appstoreconnectapi/list_builds
func (s *BuildsService) List(ctx context.Context, query *Query) (*ListBuildsResponse, error) {
	var doc Document[BuildAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/builds", query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListBuildsResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// ListIterator returns a [Paginator] that walks every page of builds.
func (s *BuildsService) ListIterator(query *Query) *Paginator[BuildAttributes] {
	return newPaginator[BuildAttributes](s.svc, "/v1/builds", query)
}

// ListAll fetches every page matching the query and returns the
// concatenated result.
func (s *BuildsService) ListAll(ctx context.Context, query *Query) ([]Build, error) {
	return s.ListIterator(query).All(ctx)
}

// Get fetches a single build by id.
// See https://developer.apple.com/documentation/appstoreconnectapi/read_build_information
func (s *BuildsService) Get(ctx context.Context, id string, query *Query) (*Build, error) {
	if id == "" {
		return nil, &ClientError{Message: "Builds.Get: id is required"}
	}
	var doc Document[BuildAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/builds/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// BuildUpdate collects the mutable fields on a build resource.
// Only the fields set via setters are sent — nil-valued keys are
// omitted so Apple interprets them as "leave unchanged".
type BuildUpdate struct {
	attrs map[string]any
}

// NewBuildUpdate returns an empty [BuildUpdate].
func NewBuildUpdate() *BuildUpdate { return &BuildUpdate{attrs: make(map[string]any)} }

// Expired marks the build as expired, blocking further TestFlight
// installs.
func (u *BuildUpdate) Expired(v bool) *BuildUpdate {
	u.attrs["expired"] = v
	return u
}

// UsesNonExemptEncryption toggles the ITSAppUsesNonExemptEncryption
// declaration for the build.
func (u *BuildUpdate) UsesNonExemptEncryption(v bool) *BuildUpdate {
	u.attrs["usesNonExemptEncryption"] = v
	return u
}

// Set applies an arbitrary attribute key/value pair.
func (u *BuildUpdate) Set(key string, value any) *BuildUpdate {
	u.attrs[key] = value
	return u
}

// IsEmpty reports whether any attribute change is pending.
func (u *BuildUpdate) IsEmpty() bool { return len(u.attrs) == 0 }

// Update modifies the mutable fields on the given build.
// See https://developer.apple.com/documentation/appstoreconnectapi/modify_a_build
func (s *BuildsService) Update(ctx context.Context, id string, update *BuildUpdate) (*Build, error) {
	if id == "" {
		return nil, &ClientError{Message: "Builds.Update: id is required"}
	}
	if update == nil || update.IsEmpty() {
		return nil, &ClientError{Message: "Builds.Update: update has no attribute changes"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type":       "builds",
			"id":         id,
			"attributes": update.attrs,
		},
	}
	var doc Document[BuildAttributes]
	if _, err := s.svc.do(ctx, "PATCH", "/v1/builds/"+id, nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}
