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

// AppUpdate collects the attributes to change in an [AppsService.Update]
// call. Nil-valued fields are omitted from the request body so that
// Apple interprets them as "leave unchanged" rather than "clear".
//
// Construct one with [NewAppUpdate] and chain setter methods:
//
//	upd := NewAppUpdate().
//	    PrimaryLocale("en-US").
//	    AvailableInNewTerritories(true)
type AppUpdate struct {
	attrs map[string]any
}

// NewAppUpdate returns an empty [AppUpdate] ready for chained setters.
func NewAppUpdate() *AppUpdate { return &AppUpdate{attrs: make(map[string]any)} }

// PrimaryLocale sets the app's primary locale (e.g. "en-US").
func (u *AppUpdate) PrimaryLocale(v string) *AppUpdate {
	u.attrs["primaryLocale"] = v
	return u
}

// BundleId changes the Bundle ID. Apple only permits this in limited
// circumstances; most bundle-id edits must go through the Bundle IDs
// endpoint instead.
func (u *AppUpdate) BundleId(v string) *AppUpdate {
	u.attrs["bundleId"] = v
	return u
}

// AvailableInNewTerritories toggles automatic availability in newly
// launched App Store territories.
func (u *AppUpdate) AvailableInNewTerritories(v bool) *AppUpdate {
	u.attrs["availableInNewTerritories"] = v
	return u
}

// ContentRightsDeclaration sets the app's content-rights declaration,
// e.g. "USES_THIRD_PARTY_CONTENT" or "DOES_NOT_USE_THIRD_PARTY_CONTENT".
func (u *AppUpdate) ContentRightsDeclaration(v string) *AppUpdate {
	u.attrs["contentRightsDeclaration"] = v
	return u
}

// SubscriptionStatusUrl sets the server-notifications v2 URL for
// production subscription status updates.
func (u *AppUpdate) SubscriptionStatusUrl(v string) *AppUpdate {
	u.attrs["subscriptionStatusUrl"] = v
	return u
}

// SubscriptionStatusUrlVersion sets the version field paired with the
// production subscription-status URL, e.g. "V2".
func (u *AppUpdate) SubscriptionStatusUrlVersion(v string) *AppUpdate {
	u.attrs["subscriptionStatusUrlVersion"] = v
	return u
}

// SubscriptionStatusUrlForSandbox sets the sandbox subscription-status URL.
func (u *AppUpdate) SubscriptionStatusUrlForSandbox(v string) *AppUpdate {
	u.attrs["subscriptionStatusUrlForSandbox"] = v
	return u
}

// SubscriptionStatusUrlVersionForSandbox sets the version paired with
// the sandbox subscription-status URL.
func (u *AppUpdate) SubscriptionStatusUrlVersionForSandbox(v string) *AppUpdate {
	u.attrs["subscriptionStatusUrlVersionForSandbox"] = v
	return u
}

// Set applies an arbitrary attribute key/value pair. Use this for
// fields not yet covered by a dedicated setter.
func (u *AppUpdate) Set(key string, value any) *AppUpdate {
	u.attrs[key] = value
	return u
}

// IsEmpty reports whether the update contains no attribute changes.
func (u *AppUpdate) IsEmpty() bool { return len(u.attrs) == 0 }

// Update applies the given attribute changes to the app with the given
// resource id. Apple returns the updated resource on success.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/modify_an_app
func (s *AppsService) Update(ctx context.Context, id string, update *AppUpdate) (*App, error) {
	if id == "" {
		return nil, &ClientError{Message: "Apps.Update: id is required"}
	}
	if update == nil || update.IsEmpty() {
		return nil, &ClientError{Message: "Apps.Update: update has no attribute changes"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type":       "apps",
			"id":         id,
			"attributes": update.attrs,
		},
	}
	var doc Document[AppAttributes]
	if _, err := s.svc.do(ctx, "PATCH", "/v1/apps/"+id, nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}
