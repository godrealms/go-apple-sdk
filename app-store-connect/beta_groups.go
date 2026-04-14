package AppStoreConnect

import (
	"context"
	"time"
)

// BetaGroupsService provides access to /v1/betaGroups, Apple's
// TestFlight tester-group management API.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/beta_groups
type BetaGroupsService struct {
	svc *Service
}

// BetaGroup is a typed alias for a JSON:API betaGroups resource.
type BetaGroup = Resource[BetaGroupAttributes]

// BetaGroupAttributes models the attributes of a betaGroups resource.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/betagroup/attributes
type BetaGroupAttributes struct {
	Name                      string     `json:"name,omitempty"`
	CreatedDate               *time.Time `json:"createdDate,omitempty"`
	IsInternalGroup           *bool      `json:"isInternalGroup,omitempty"`
	HasAccessToAllBuilds      *bool      `json:"hasAccessToAllBuilds,omitempty"`
	PublicLinkEnabled         *bool      `json:"publicLinkEnabled,omitempty"`
	PublicLink                string     `json:"publicLink,omitempty"`
	PublicLinkLimitEnabled    *bool      `json:"publicLinkLimitEnabled,omitempty"`
	PublicLinkLimit           *int       `json:"publicLinkLimit,omitempty"`
	FeedbackEnabled           *bool      `json:"feedbackEnabled,omitempty"`
	IosBuildsAvailableForAppleSiliconMac *bool `json:"iosBuildsAvailableForAppleSiliconMac,omitempty"`
}

// ListBetaGroupsResponse is the decoded response for [BetaGroupsService.List].
type ListBetaGroupsResponse struct {
	Data     []BetaGroup     `json:"data"`
	Included []Resource[any] `json:"included,omitempty"`
	Links    *Links          `json:"links,omitempty"`
}

// List returns a page of beta groups matching the query.
// See https://developer.apple.com/documentation/appstoreconnectapi/list_beta_groups
func (s *BetaGroupsService) List(ctx context.Context, query *Query) (*ListBetaGroupsResponse, error) {
	var doc Document[BetaGroupAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/betaGroups", query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListBetaGroupsResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// ListIterator returns a paginator that walks every page of beta groups.
func (s *BetaGroupsService) ListIterator(query *Query) *Paginator[BetaGroupAttributes] {
	return newPaginator[BetaGroupAttributes](s.svc, "/v1/betaGroups", query)
}

// Get fetches a single beta group by id.
func (s *BetaGroupsService) Get(ctx context.Context, id string, query *Query) (*BetaGroup, error) {
	if id == "" {
		return nil, &ClientError{Message: "BetaGroups.Get: id is required"}
	}
	var doc Document[BetaGroupAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/betaGroups/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// CreateBetaGroupRequest describes the attributes of a new beta group,
// plus the parent app relationship. AppID is required.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/create_a_beta_group
type CreateBetaGroupRequest struct {
	AppID                  string // required
	Name                   string // required
	PublicLinkEnabled      *bool
	PublicLinkLimitEnabled *bool
	PublicLinkLimit        *int
	FeedbackEnabled        *bool
}

// Create creates a new beta group under the given app.
func (s *BetaGroupsService) Create(ctx context.Context, req CreateBetaGroupRequest) (*BetaGroup, error) {
	if req.AppID == "" {
		return nil, &ClientError{Message: "BetaGroups.Create: AppID is required"}
	}
	if req.Name == "" {
		return nil, &ClientError{Message: "BetaGroups.Create: Name is required"}
	}
	attrs := map[string]any{"name": req.Name}
	if req.PublicLinkEnabled != nil {
		attrs["publicLinkEnabled"] = *req.PublicLinkEnabled
	}
	if req.PublicLinkLimitEnabled != nil {
		attrs["publicLinkLimitEnabled"] = *req.PublicLinkLimitEnabled
	}
	if req.PublicLinkLimit != nil {
		attrs["publicLinkLimit"] = *req.PublicLinkLimit
	}
	if req.FeedbackEnabled != nil {
		attrs["feedbackEnabled"] = *req.FeedbackEnabled
	}
	body := map[string]any{
		"data": map[string]any{
			"type":       "betaGroups",
			"attributes": attrs,
			"relationships": map[string]any{
				"app": map[string]any{
					"data": map[string]any{"type": "apps", "id": req.AppID},
				},
			},
		},
	}
	var doc Document[BetaGroupAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/betaGroups", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete removes a beta group by id.
func (s *BetaGroupsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "BetaGroups.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/betaGroups/"+id, nil, nil, nil)
	return err
}

// AddBuilds attaches one or more builds to the given beta group.
// Apple will start distributing the build to the group's testers.
// See https://developer.apple.com/documentation/appstoreconnectapi/add_builds_to_a_beta_group
func (s *BetaGroupsService) AddBuilds(ctx context.Context, groupID string, buildIDs []string) error {
	if groupID == "" {
		return &ClientError{Message: "BetaGroups.AddBuilds: groupID is required"}
	}
	if len(buildIDs) == 0 {
		return &ClientError{Message: "BetaGroups.AddBuilds: buildIDs is empty"}
	}
	body := map[string]any{"data": buildIdentifiers("builds", buildIDs)}
	_, err := s.svc.do(ctx, "POST", "/v1/betaGroups/"+groupID+"/relationships/builds", nil, body, nil)
	return err
}

// AddBetaTesters attaches existing beta tester resources to the given group.
// See https://developer.apple.com/documentation/appstoreconnectapi/add_beta_testers_to_a_beta_group
func (s *BetaGroupsService) AddBetaTesters(ctx context.Context, groupID string, testerIDs []string) error {
	if groupID == "" {
		return &ClientError{Message: "BetaGroups.AddBetaTesters: groupID is required"}
	}
	if len(testerIDs) == 0 {
		return &ClientError{Message: "BetaGroups.AddBetaTesters: testerIDs is empty"}
	}
	body := map[string]any{"data": buildIdentifiers("betaTesters", testerIDs)}
	_, err := s.svc.do(ctx, "POST", "/v1/betaGroups/"+groupID+"/relationships/betaTesters", nil, body, nil)
	return err
}

// buildIdentifiers is a small helper to construct the `{type,id}` array
// that JSON:API relationship endpoints expect.
func buildIdentifiers(resourceType string, ids []string) []map[string]any {
	out := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		out = append(out, map[string]any{"type": resourceType, "id": id})
	}
	return out
}
