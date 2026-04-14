package AppStoreConnect

import "context"

// AppStoreVersionLocalizationsService provides access to
// /v1/appStoreVersionLocalizations, Apple's per-locale store listing
// metadata (what's new, marketing text, keywords, support URL, etc.).
//
// Each appStoreVersions resource owns a set of localizations — one per
// locale the app ships in. The list for a version is returned from
// /v1/appStoreVersions/{id}/appStoreVersionLocalizations.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/app_store_version_localizations
type AppStoreVersionLocalizationsService struct {
	svc *Service
}

// AppStoreVersionLocalization is a typed alias for a JSON:API
// appStoreVersionLocalizations resource.
type AppStoreVersionLocalization = Resource[AppStoreVersionLocalizationAttributes]

// AppStoreVersionLocalizationAttributes models the mutable per-locale
// fields of an App Store version.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/appstoreversionlocalization/attributes
type AppStoreVersionLocalizationAttributes struct {
	Locale          string `json:"locale,omitempty"`
	Description     string `json:"description,omitempty"`
	Keywords        string `json:"keywords,omitempty"`
	MarketingURL    string `json:"marketingUrl,omitempty"`
	PromotionalText string `json:"promotionalText,omitempty"`
	SupportURL      string `json:"supportUrl,omitempty"`
	WhatsNew        string `json:"whatsNew,omitempty"`
}

// ListAppStoreVersionLocalizationsResponse is the decoded response for
// [AppStoreVersionLocalizationsService.ListForVersion].
type ListAppStoreVersionLocalizationsResponse struct {
	Data     []AppStoreVersionLocalization `json:"data"`
	Included []Resource[any]               `json:"included,omitempty"`
	Links    *Links                        `json:"links,omitempty"`
}

// ListForVersion returns every localization attached to the given
// App Store version. Apple caps the collection well below the pagination
// limit (at most a few dozen locales), so this is almost always a
// single-page request — we still honor the query parameter for
// `fields[appStoreVersionLocalizations]` projection.
func (s *AppStoreVersionLocalizationsService) ListForVersion(ctx context.Context, versionID string, query *Query) (*ListAppStoreVersionLocalizationsResponse, error) {
	if versionID == "" {
		return nil, &ClientError{Message: "AppStoreVersionLocalizations.ListForVersion: versionID is required"}
	}
	var doc Document[AppStoreVersionLocalizationAttributes]
	path := "/v1/appStoreVersions/" + versionID + "/appStoreVersionLocalizations"
	if _, err := s.svc.do(ctx, "GET", path, query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListAppStoreVersionLocalizationsResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// Get fetches a single localization by resource id.
func (s *AppStoreVersionLocalizationsService) Get(ctx context.Context, id string, query *Query) (*AppStoreVersionLocalization, error) {
	if id == "" {
		return nil, &ClientError{Message: "AppStoreVersionLocalizations.Get: id is required"}
	}
	var doc Document[AppStoreVersionLocalizationAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/appStoreVersionLocalizations/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// CreateAppStoreVersionLocalizationRequest describes a new localization.
// VersionID and Locale are required; everything else is optional but
// each nonempty attribute travels in the POST body.
type CreateAppStoreVersionLocalizationRequest struct {
	VersionID       string // required — appStoreVersions resource id
	Locale          string // required — e.g. "en-US", "ja"
	Description     string
	Keywords        string
	MarketingURL    string
	PromotionalText string
	SupportURL      string
	WhatsNew        string
}

// Create registers a new localization under the given App Store version.
// See https://developer.apple.com/documentation/appstoreconnectapi/create_an_app_store_version_localization
func (s *AppStoreVersionLocalizationsService) Create(ctx context.Context, req CreateAppStoreVersionLocalizationRequest) (*AppStoreVersionLocalization, error) {
	if req.VersionID == "" || req.Locale == "" {
		return nil, &ClientError{Message: "AppStoreVersionLocalizations.Create: VersionID and Locale are required"}
	}
	attrs := map[string]any{"locale": req.Locale}
	if req.Description != "" {
		attrs["description"] = req.Description
	}
	if req.Keywords != "" {
		attrs["keywords"] = req.Keywords
	}
	if req.MarketingURL != "" {
		attrs["marketingUrl"] = req.MarketingURL
	}
	if req.PromotionalText != "" {
		attrs["promotionalText"] = req.PromotionalText
	}
	if req.SupportURL != "" {
		attrs["supportUrl"] = req.SupportURL
	}
	if req.WhatsNew != "" {
		attrs["whatsNew"] = req.WhatsNew
	}
	body := map[string]any{
		"data": map[string]any{
			"type":       "appStoreVersionLocalizations",
			"attributes": attrs,
			"relationships": map[string]any{
				"appStoreVersion": map[string]any{
					"data": map[string]any{"type": "appStoreVersions", "id": req.VersionID},
				},
			},
		},
	}
	var doc Document[AppStoreVersionLocalizationAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/appStoreVersionLocalizations", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// AppStoreVersionLocalizationUpdate collects the mutable attributes on
// an appStoreVersionLocalizations resource. Only set fields travel in
// the PATCH body, so callers can tweak a single locale field without
// overwriting the others.
type AppStoreVersionLocalizationUpdate struct {
	attrs map[string]any
}

// NewAppStoreVersionLocalizationUpdate returns an empty update.
func NewAppStoreVersionLocalizationUpdate() *AppStoreVersionLocalizationUpdate {
	return &AppStoreVersionLocalizationUpdate{attrs: make(map[string]any)}
}

// Description sets the long App Store description for this locale.
func (u *AppStoreVersionLocalizationUpdate) Description(v string) *AppStoreVersionLocalizationUpdate {
	u.attrs["description"] = v
	return u
}

// Keywords sets the comma-separated keyword string.
func (u *AppStoreVersionLocalizationUpdate) Keywords(v string) *AppStoreVersionLocalizationUpdate {
	u.attrs["keywords"] = v
	return u
}

// MarketingURL sets the marketing URL shown on the store listing.
func (u *AppStoreVersionLocalizationUpdate) MarketingURL(v string) *AppStoreVersionLocalizationUpdate {
	u.attrs["marketingUrl"] = v
	return u
}

// PromotionalText sets the promotional text banner that Apple allows
// editing without re-submission.
func (u *AppStoreVersionLocalizationUpdate) PromotionalText(v string) *AppStoreVersionLocalizationUpdate {
	u.attrs["promotionalText"] = v
	return u
}

// SupportURL sets the support URL for customers.
func (u *AppStoreVersionLocalizationUpdate) SupportURL(v string) *AppStoreVersionLocalizationUpdate {
	u.attrs["supportUrl"] = v
	return u
}

// WhatsNew sets the release notes ("What's New in This Version").
func (u *AppStoreVersionLocalizationUpdate) WhatsNew(v string) *AppStoreVersionLocalizationUpdate {
	u.attrs["whatsNew"] = v
	return u
}

// Set applies an arbitrary attribute key/value pair for fields this
// builder has not modelled yet.
func (u *AppStoreVersionLocalizationUpdate) Set(key string, value any) *AppStoreVersionLocalizationUpdate {
	u.attrs[key] = value
	return u
}

// IsEmpty reports whether any attribute change is pending.
func (u *AppStoreVersionLocalizationUpdate) IsEmpty() bool { return len(u.attrs) == 0 }

// Update modifies the given localization's attributes.
// See https://developer.apple.com/documentation/appstoreconnectapi/modify_an_app_store_version_localization
func (s *AppStoreVersionLocalizationsService) Update(ctx context.Context, id string, update *AppStoreVersionLocalizationUpdate) (*AppStoreVersionLocalization, error) {
	if id == "" {
		return nil, &ClientError{Message: "AppStoreVersionLocalizations.Update: id is required"}
	}
	if update == nil || update.IsEmpty() {
		return nil, &ClientError{Message: "AppStoreVersionLocalizations.Update: update has no attribute changes"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type":       "appStoreVersionLocalizations",
			"id":         id,
			"attributes": update.attrs,
		},
	}
	var doc Document[AppStoreVersionLocalizationAttributes]
	if _, err := s.svc.do(ctx, "PATCH", "/v1/appStoreVersionLocalizations/"+id, nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete removes a localization. Apple refuses to delete the
// primary locale of a version.
func (s *AppStoreVersionLocalizationsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "AppStoreVersionLocalizations.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/appStoreVersionLocalizations/"+id, nil, nil, nil)
	return err
}
