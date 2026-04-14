package AppStoreConnect

import "context"

// BundleIDsService provides access to /v1/bundleIds, the developer
// portal's Bundle ID catalog used by provisioning profiles.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/bundle_ids
type BundleIDsService struct {
	svc *Service
}

// BundleID is a typed alias for a JSON:API bundleIds resource.
type BundleID = Resource[BundleIDAttributes]

// BundleIDAttributes models the attributes of a bundleIds resource.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/bundleid/attributes
type BundleIDAttributes struct {
	Identifier string `json:"identifier,omitempty"`
	Name       string `json:"name,omitempty"`
	Platform   string `json:"platform,omitempty"` // IOS, MAC_OS, UNIVERSAL
	SeedID     string `json:"seedId,omitempty"`
}

// ListBundleIDsResponse is the decoded response for [BundleIDsService.List].
type ListBundleIDsResponse struct {
	Data     []BundleID      `json:"data"`
	Included []Resource[any] `json:"included,omitempty"`
	Links    *Links          `json:"links,omitempty"`
}

// List returns a page of bundle IDs matching the query.
func (s *BundleIDsService) List(ctx context.Context, query *Query) (*ListBundleIDsResponse, error) {
	var doc Document[BundleIDAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/bundleIds", query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListBundleIDsResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// ListIterator returns a paginator that walks every page of bundle IDs.
func (s *BundleIDsService) ListIterator(query *Query) *Paginator[BundleIDAttributes] {
	return newPaginator[BundleIDAttributes](s.svc, "/v1/bundleIds", query)
}

// Get fetches a single bundle ID by resource id.
func (s *BundleIDsService) Get(ctx context.Context, id string, query *Query) (*BundleID, error) {
	if id == "" {
		return nil, &ClientError{Message: "BundleIDs.Get: id is required"}
	}
	var doc Document[BundleIDAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/bundleIds/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// CreateBundleIDRequest describes a new bundle ID registration.
// Identifier, Name, and Platform are all required. SeedID is optional
// and almost always left blank.
type CreateBundleIDRequest struct {
	Identifier string // e.g. "com.acme.widgets"
	Name       string // human-readable label
	Platform   string // "IOS", "MAC_OS", or "UNIVERSAL"
	SeedID     string // optional team seed id
}

// Create registers a new bundle ID with Apple.
func (s *BundleIDsService) Create(ctx context.Context, req CreateBundleIDRequest) (*BundleID, error) {
	if req.Identifier == "" || req.Name == "" || req.Platform == "" {
		return nil, &ClientError{Message: "BundleIDs.Create: Identifier, Name, and Platform are required"}
	}
	attrs := map[string]any{
		"identifier": req.Identifier,
		"name":       req.Name,
		"platform":   req.Platform,
	}
	if req.SeedID != "" {
		attrs["seedId"] = req.SeedID
	}
	body := map[string]any{
		"data": map[string]any{
			"type":       "bundleIds",
			"attributes": attrs,
		},
	}
	var doc Document[BundleIDAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/bundleIds", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete removes a bundle ID by resource id. Apple will refuse this
// if any active provisioning profile references it.
func (s *BundleIDsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "BundleIDs.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/bundleIds/"+id, nil, nil, nil)
	return err
}
