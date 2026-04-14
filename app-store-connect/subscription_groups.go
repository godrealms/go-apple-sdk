package AppStoreConnect

import "context"

// SubscriptionGroupsService provides access to /v1/subscriptionGroups,
// the container that ties auto-renewable subscriptions to an app.
//
// A subscription group holds one or more subscriptions at different
// tiers; customers can upgrade and downgrade within a group. Each app
// may have multiple groups (e.g. "Pro" vs "Family Plan").
//
// See https://developer.apple.com/documentation/appstoreconnectapi/subscription_groups
type SubscriptionGroupsService struct {
	svc *Service
}

// SubscriptionGroup is a typed alias for a JSON:API subscriptionGroups
// resource.
type SubscriptionGroup = Resource[SubscriptionGroupAttributes]

// SubscriptionGroupAttributes models the attributes of a
// subscriptionGroups resource.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/subscriptiongroup/attributes
type SubscriptionGroupAttributes struct {
	ReferenceName string `json:"referenceName,omitempty"`
}

// ListSubscriptionGroupsResponse is the decoded response for
// [SubscriptionGroupsService.ListForApp].
type ListSubscriptionGroupsResponse struct {
	Data     []SubscriptionGroup `json:"data"`
	Included []Resource[any]     `json:"included,omitempty"`
	Links    *Links              `json:"links,omitempty"`
}

// ListForApp returns every subscription group configured under the
// given app.
func (s *SubscriptionGroupsService) ListForApp(ctx context.Context, appID string, query *Query) (*ListSubscriptionGroupsResponse, error) {
	if appID == "" {
		return nil, &ClientError{Message: "SubscriptionGroups.ListForApp: appID is required"}
	}
	var doc Document[SubscriptionGroupAttributes]
	path := "/v1/apps/" + appID + "/subscriptionGroups"
	if _, err := s.svc.do(ctx, "GET", path, query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListSubscriptionGroupsResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// ListForAppIterator returns a paginator that walks every page of
// subscription groups for an app.
func (s *SubscriptionGroupsService) ListForAppIterator(appID string, query *Query) *Paginator[SubscriptionGroupAttributes] {
	return newPaginator[SubscriptionGroupAttributes](s.svc, "/v1/apps/"+appID+"/subscriptionGroups", query)
}

// Get fetches a single subscription group by resource id.
func (s *SubscriptionGroupsService) Get(ctx context.Context, id string, query *Query) (*SubscriptionGroup, error) {
	if id == "" {
		return nil, &ClientError{Message: "SubscriptionGroups.Get: id is required"}
	}
	var doc Document[SubscriptionGroupAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/subscriptionGroups/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// CreateSubscriptionGroupRequest describes a new subscription group.
// Both fields are required — Apple uses ReferenceName for internal
// bookkeeping only; customer-facing names live on per-locale
// subscription group localizations.
type CreateSubscriptionGroupRequest struct {
	AppID         string // required
	ReferenceName string // required — internal name (not customer-facing)
}

// Create registers a new subscription group under the given app.
// See https://developer.apple.com/documentation/appstoreconnectapi/create_a_subscription_group
func (s *SubscriptionGroupsService) Create(ctx context.Context, req CreateSubscriptionGroupRequest) (*SubscriptionGroup, error) {
	if req.AppID == "" || req.ReferenceName == "" {
		return nil, &ClientError{Message: "SubscriptionGroups.Create: AppID and ReferenceName are required"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type": "subscriptionGroups",
			"attributes": map[string]any{
				"referenceName": req.ReferenceName,
			},
			"relationships": map[string]any{
				"app": map[string]any{
					"data": map[string]any{"type": "apps", "id": req.AppID},
				},
			},
		},
	}
	var doc Document[SubscriptionGroupAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/subscriptionGroups", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Update renames a subscription group's reference name.
func (s *SubscriptionGroupsService) Update(ctx context.Context, id, referenceName string) (*SubscriptionGroup, error) {
	if id == "" {
		return nil, &ClientError{Message: "SubscriptionGroups.Update: id is required"}
	}
	if referenceName == "" {
		return nil, &ClientError{Message: "SubscriptionGroups.Update: referenceName is required"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type": "subscriptionGroups",
			"id":   id,
			"attributes": map[string]any{
				"referenceName": referenceName,
			},
		},
	}
	var doc Document[SubscriptionGroupAttributes]
	if _, err := s.svc.do(ctx, "PATCH", "/v1/subscriptionGroups/"+id, nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete removes a subscription group. Apple refuses deletion while
// the group still contains approved subscriptions.
func (s *SubscriptionGroupsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "SubscriptionGroups.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/subscriptionGroups/"+id, nil, nil, nil)
	return err
}
