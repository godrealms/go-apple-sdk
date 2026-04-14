package AppStoreConnect

import "context"

// AppScreenshotSetsService manages /v1/appScreenshotSets, the
// per-device-class containers that group screenshots for a given
// App Store version localization.
//
// A screenshot set is keyed by screenshotDisplayType (e.g.
// "APP_IPHONE_67", "APP_IPAD_PRO_129"). Each set belongs to one
// appStoreVersionLocalizations resource.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/app_screenshot_sets
type AppScreenshotSetsService struct {
	svc *Service
}

// AppScreenshotSet is a typed alias for a JSON:API appScreenshotSets resource.
type AppScreenshotSet = Resource[AppScreenshotSetAttributes]

// AppScreenshotSetAttributes models the attributes of an
// appScreenshotSets resource.
type AppScreenshotSetAttributes struct {
	ScreenshotDisplayType string `json:"screenshotDisplayType,omitempty"`
}

// ListAppScreenshotSetsResponse is the decoded response for
// [AppScreenshotSetsService.ListForLocalization].
type ListAppScreenshotSetsResponse struct {
	Data     []AppScreenshotSet `json:"data"`
	Included []Resource[any]    `json:"included,omitempty"`
	Links    *Links             `json:"links,omitempty"`
}

// ListForLocalization returns every screenshot set attached to a given
// version localization.
func (s *AppScreenshotSetsService) ListForLocalization(ctx context.Context, localizationID string, query *Query) (*ListAppScreenshotSetsResponse, error) {
	if localizationID == "" {
		return nil, &ClientError{Message: "AppScreenshotSets.ListForLocalization: localizationID is required"}
	}
	var doc Document[AppScreenshotSetAttributes]
	path := "/v1/appStoreVersionLocalizations/" + localizationID + "/appScreenshotSets"
	if _, err := s.svc.do(ctx, "GET", path, query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListAppScreenshotSetsResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// Get fetches a single screenshot set by resource id.
func (s *AppScreenshotSetsService) Get(ctx context.Context, id string, query *Query) (*AppScreenshotSet, error) {
	if id == "" {
		return nil, &ClientError{Message: "AppScreenshotSets.Get: id is required"}
	}
	var doc Document[AppScreenshotSetAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/appScreenshotSets/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// CreateAppScreenshotSetRequest describes the screenshot set to create.
type CreateAppScreenshotSetRequest struct {
	LocalizationID        string // required — parent appStoreVersionLocalizations id
	ScreenshotDisplayType string // required — e.g. "APP_IPHONE_67"
}

// Create registers a new screenshot set under the given version
// localization. Returns the empty set — upload screenshots via
// [AppScreenshotsService.Reserve].
func (s *AppScreenshotSetsService) Create(ctx context.Context, req CreateAppScreenshotSetRequest) (*AppScreenshotSet, error) {
	if req.LocalizationID == "" || req.ScreenshotDisplayType == "" {
		return nil, &ClientError{Message: "AppScreenshotSets.Create: LocalizationID and ScreenshotDisplayType are required"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type": "appScreenshotSets",
			"attributes": map[string]any{
				"screenshotDisplayType": req.ScreenshotDisplayType,
			},
			"relationships": map[string]any{
				"appStoreVersionLocalization": map[string]any{
					"data": map[string]any{"type": "appStoreVersionLocalizations", "id": req.LocalizationID},
				},
			},
		},
	}
	var doc Document[AppScreenshotSetAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/appScreenshotSets", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete removes a screenshot set (and every screenshot inside it).
func (s *AppScreenshotSetsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "AppScreenshotSets.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/appScreenshotSets/"+id, nil, nil, nil)
	return err
}
