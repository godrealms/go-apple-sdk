package AppStoreConnect

import (
	"context"
	"time"
)

// AppsService provides access to the /v1/apps endpoints.
// See https://developer.apple.com/documentation/appstoreconnectapi/apps
type AppsService struct {
	svc *Service
}

// App is a typed alias for a JSON:API App resource.
type App = Resource[AppAttributes]

// AppAttributes models the "attributes" object of an App resource.
// Apple evolves this schema over time; unknown fields are ignored.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/app/attributes
type AppAttributes struct {
	Name                          string     `json:"name,omitempty"`
	BundleId                      string     `json:"bundleId,omitempty"`
	Sku                           string     `json:"sku,omitempty"`
	PrimaryLocale                 string     `json:"primaryLocale,omitempty"`
	IsOrEverWasMadeForKids        bool       `json:"isOrEverWasMadeForKids,omitempty"`
	SubscriptionStatusUrl         string     `json:"subscriptionStatusUrl,omitempty"`
	SubscriptionStatusUrlVersion  string     `json:"subscriptionStatusUrlVersion,omitempty"`
	SubscriptionStatusUrlForSandbox        string `json:"subscriptionStatusUrlForSandbox,omitempty"`
	SubscriptionStatusUrlVersionForSandbox string `json:"subscriptionStatusUrlVersionForSandbox,omitempty"`
	ContentRightsDeclaration      string     `json:"contentRightsDeclaration,omitempty"`
	AvailableInNewTerritories     *bool      `json:"availableInNewTerritories,omitempty"`
	CreatedDate                   *time.Time `json:"createdDate,omitempty"`
}

// ListAppsResponse is the decoded response for [AppsService.List].
type ListAppsResponse struct {
	Data     []App           `json:"data"`
	Included []Resource[any] `json:"included,omitempty"`
	Links    *Links          `json:"links,omitempty"`
}

// List returns a page of apps matching the given query.
//
// Returned alongside the page is the raw [Document] envelope so the
// caller can inspect links.next if they want to drive pagination
// manually. For most callers [AppsService.ListIterator] or
// [AppsService.ListAll] is simpler.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/list_apps
func (s *AppsService) List(ctx context.Context, query *Query) (*ListAppsResponse, error) {
	var doc Document[AppAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/apps", query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListAppsResponse{
		Data:     data,
		Included: doc.Included,
		Links:    doc.Links,
	}, nil
}

// ListIterator returns a [Paginator] that walks through every page of
// apps matching the query, automatically following links.next.
func (s *AppsService) ListIterator(query *Query) *Paginator[AppAttributes] {
	return newPaginator[AppAttributes](s.svc, "/v1/apps", query)
}

// ListAll fetches every page matching the query and returns the
// concatenated result. For large accounts prefer [AppsService.ListIterator]
// to avoid loading everything into memory at once.
func (s *AppsService) ListAll(ctx context.Context, query *Query) ([]App, error) {
	return s.ListIterator(query).All(ctx)
}

// Get fetches a single app by its App Store Connect resource id.
// See https://developer.apple.com/documentation/appstoreconnectapi/read_app_information
func (s *AppsService) Get(ctx context.Context, id string, query *Query) (*App, error) {
	if id == "" {
		return nil, &ClientError{Message: "Apps.Get: id is required"}
	}
	var doc Document[AppAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/apps/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}
