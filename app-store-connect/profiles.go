package AppStoreConnect

import (
	"context"
	"time"
)

// ProfilesService provides access to /v1/profiles, Apple's
// provisioning-profile catalog.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/profiles
type ProfilesService struct {
	svc *Service
}

// Profile is a typed alias for a JSON:API profiles resource.
type Profile = Resource[ProfileAttributes]

// ProfileAttributes models the attributes of a profiles resource.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/profile/attributes
type ProfileAttributes struct {
	Name           string     `json:"name,omitempty"`
	Platform       string     `json:"platform,omitempty"`
	ProfileContent string     `json:"profileContent,omitempty"` // base64
	UUID           string     `json:"uuid,omitempty"`
	CreatedDate    *time.Time `json:"createdDate,omitempty"`
	ProfileState   string     `json:"profileState,omitempty"`
	ProfileType    string     `json:"profileType,omitempty"`
	ExpirationDate *time.Time `json:"expirationDate,omitempty"`
}

// ListProfilesResponse is the decoded response for [ProfilesService.List].
type ListProfilesResponse struct {
	Data     []Profile       `json:"data"`
	Included []Resource[any] `json:"included,omitempty"`
	Links    *Links          `json:"links,omitempty"`
}

// List returns a page of profiles matching the query.
func (s *ProfilesService) List(ctx context.Context, query *Query) (*ListProfilesResponse, error) {
	var doc Document[ProfileAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/profiles", query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListProfilesResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// ListIterator returns a paginator that walks every page of profiles.
func (s *ProfilesService) ListIterator(query *Query) *Paginator[ProfileAttributes] {
	return newPaginator[ProfileAttributes](s.svc, "/v1/profiles", query)
}

// Get fetches a single profile by resource id.
func (s *ProfilesService) Get(ctx context.Context, id string, query *Query) (*Profile, error) {
	if id == "" {
		return nil, &ClientError{Message: "Profiles.Get: id is required"}
	}
	var doc Document[ProfileAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/profiles/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// CreateProfileRequest describes a new profile. All four fields are
// required: Name, ProfileType (e.g. IOS_APP_DEVELOPMENT), BundleID
// (resource id), and at least one CertificateID. DeviceIDs are
// required for development profiles but optional for distribution.
type CreateProfileRequest struct {
	Name           string
	ProfileType    string
	BundleID       string
	CertificateIDs []string
	DeviceIDs      []string
}

// Create registers a new provisioning profile.
// See https://developer.apple.com/documentation/appstoreconnectapi/create_a_profile
func (s *ProfilesService) Create(ctx context.Context, req CreateProfileRequest) (*Profile, error) {
	if req.Name == "" || req.ProfileType == "" || req.BundleID == "" {
		return nil, &ClientError{Message: "Profiles.Create: Name, ProfileType, and BundleID are required"}
	}
	if len(req.CertificateIDs) == 0 {
		return nil, &ClientError{Message: "Profiles.Create: at least one CertificateID is required"}
	}
	relationships := map[string]any{
		"bundleId": map[string]any{
			"data": map[string]any{"type": "bundleIds", "id": req.BundleID},
		},
		"certificates": map[string]any{
			"data": buildIdentifiers("certificates", req.CertificateIDs),
		},
	}
	if len(req.DeviceIDs) > 0 {
		relationships["devices"] = map[string]any{
			"data": buildIdentifiers("devices", req.DeviceIDs),
		}
	}
	body := map[string]any{
		"data": map[string]any{
			"type": "profiles",
			"attributes": map[string]any{
				"name":        req.Name,
				"profileType": req.ProfileType,
			},
			"relationships": relationships,
		},
	}
	var doc Document[ProfileAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/profiles", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete removes a profile by resource id.
func (s *ProfilesService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "Profiles.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/profiles/"+id, nil, nil, nil)
	return err
}
