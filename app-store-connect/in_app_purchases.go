package AppStoreConnect

import "context"

// InAppPurchasesService provides access to /v1/inAppPurchasesV2,
// Apple's managed catalog of non-subscription in-app purchases (the
// "V2" endpoints replaced the legacy /v1/inAppPurchases surface).
//
// Subscriptions live under /v1/subscriptionGroups and
// /v1/subscriptions, not here — see [SubscriptionGroupsService].
//
// See https://developer.apple.com/documentation/appstoreconnectapi/in-app_purchases_v2
type InAppPurchasesService struct {
	svc *Service
}

// InAppPurchase is a typed alias for a JSON:API inAppPurchases
// (v2) resource.
type InAppPurchase = Resource[InAppPurchaseAttributes]

// InAppPurchaseAttributes models the attributes of an inAppPurchases
// v2 resource.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/inapppurchasev2/attributes
type InAppPurchaseAttributes struct {
	Name                 string `json:"name,omitempty"`
	ProductID            string `json:"productId,omitempty"`
	InAppPurchaseType    string `json:"inAppPurchaseType,omitempty"` // CONSUMABLE / NON_CONSUMABLE / NON_RENEWING_SUBSCRIPTION
	State                string `json:"state,omitempty"`
	ReviewNote           string `json:"reviewNote,omitempty"`
	FamilySharable       *bool  `json:"familySharable,omitempty"`
	ContentHosting       *bool  `json:"contentHosting,omitempty"`
	AvailableInAllTerris *bool  `json:"availableInAllTerritories,omitempty"`
}

// ListInAppPurchasesResponse is the decoded response for
// [InAppPurchasesService.ListForApp].
type ListInAppPurchasesResponse struct {
	Data     []InAppPurchase `json:"data"`
	Included []Resource[any] `json:"included,omitempty"`
	Links    *Links          `json:"links,omitempty"`
}

// ListForApp returns every in-app purchase configured for the given app.
// Use the query builder to filter by state or project fields.
func (s *InAppPurchasesService) ListForApp(ctx context.Context, appID string, query *Query) (*ListInAppPurchasesResponse, error) {
	if appID == "" {
		return nil, &ClientError{Message: "InAppPurchases.ListForApp: appID is required"}
	}
	var doc Document[InAppPurchaseAttributes]
	path := "/v1/apps/" + appID + "/inAppPurchasesV2"
	if _, err := s.svc.do(ctx, "GET", path, query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListInAppPurchasesResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// ListForAppIterator returns a paginator that walks every page of IAPs
// for an app.
func (s *InAppPurchasesService) ListForAppIterator(appID string, query *Query) *Paginator[InAppPurchaseAttributes] {
	return newPaginator[InAppPurchaseAttributes](s.svc, "/v1/apps/"+appID+"/inAppPurchasesV2", query)
}

// Get fetches a single in-app purchase by resource id.
func (s *InAppPurchasesService) Get(ctx context.Context, id string, query *Query) (*InAppPurchase, error) {
	if id == "" {
		return nil, &ClientError{Message: "InAppPurchases.Get: id is required"}
	}
	var doc Document[InAppPurchaseAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/inAppPurchases/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// CreateInAppPurchaseRequest describes a new IAP. Apple requires
// AppID, Name, ProductID, and InAppPurchaseType; the rest are optional
// defaults the developer can flip later.
type CreateInAppPurchaseRequest struct {
	AppID             string // required
	Name              string // required — reference name (not shown to customers)
	ProductID         string // required — must be globally unique under the app
	InAppPurchaseType string // required — CONSUMABLE / NON_CONSUMABLE / NON_RENEWING_SUBSCRIPTION
	ReviewNote        string
	FamilySharable    *bool
}

// Create registers a new non-subscription in-app purchase under the
// given app. For renewable subscriptions use [SubscriptionGroupsService.Create]
// and related subscription endpoints instead.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/create_an_in-app_purchase_v2
func (s *InAppPurchasesService) Create(ctx context.Context, req CreateInAppPurchaseRequest) (*InAppPurchase, error) {
	if req.AppID == "" || req.Name == "" || req.ProductID == "" || req.InAppPurchaseType == "" {
		return nil, &ClientError{Message: "InAppPurchases.Create: AppID, Name, ProductID, and InAppPurchaseType are required"}
	}
	attrs := map[string]any{
		"name":              req.Name,
		"productId":         req.ProductID,
		"inAppPurchaseType": req.InAppPurchaseType,
	}
	if req.ReviewNote != "" {
		attrs["reviewNote"] = req.ReviewNote
	}
	if req.FamilySharable != nil {
		attrs["familySharable"] = *req.FamilySharable
	}
	body := map[string]any{
		"data": map[string]any{
			"type":       "inAppPurchases",
			"attributes": attrs,
			"relationships": map[string]any{
				"app": map[string]any{
					"data": map[string]any{"type": "apps", "id": req.AppID},
				},
			},
		},
	}
	var doc Document[InAppPurchaseAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/inAppPurchases", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// InAppPurchaseUpdate collects the mutable attributes on an IAP
// resource. Nil fields are omitted from the PATCH body.
type InAppPurchaseUpdate struct {
	attrs map[string]any
}

// NewInAppPurchaseUpdate returns an empty update.
func NewInAppPurchaseUpdate() *InAppPurchaseUpdate {
	return &InAppPurchaseUpdate{attrs: make(map[string]any)}
}

// Name sets the developer-facing reference name.
func (u *InAppPurchaseUpdate) Name(v string) *InAppPurchaseUpdate {
	u.attrs["name"] = v
	return u
}

// ReviewNote sets the App Review reviewer note.
func (u *InAppPurchaseUpdate) ReviewNote(v string) *InAppPurchaseUpdate {
	u.attrs["reviewNote"] = v
	return u
}

// FamilySharable toggles whether the purchase is family-shareable.
func (u *InAppPurchaseUpdate) FamilySharable(v bool) *InAppPurchaseUpdate {
	u.attrs["familySharable"] = v
	return u
}

// Set applies an arbitrary attribute key/value pair.
func (u *InAppPurchaseUpdate) Set(key string, value any) *InAppPurchaseUpdate {
	u.attrs[key] = value
	return u
}

// IsEmpty reports whether any attribute change is pending.
func (u *InAppPurchaseUpdate) IsEmpty() bool { return len(u.attrs) == 0 }

// Update modifies the given IAP's attributes.
func (s *InAppPurchasesService) Update(ctx context.Context, id string, update *InAppPurchaseUpdate) (*InAppPurchase, error) {
	if id == "" {
		return nil, &ClientError{Message: "InAppPurchases.Update: id is required"}
	}
	if update == nil || update.IsEmpty() {
		return nil, &ClientError{Message: "InAppPurchases.Update: update has no attribute changes"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type":       "inAppPurchases",
			"id":         id,
			"attributes": update.attrs,
		},
	}
	var doc Document[InAppPurchaseAttributes]
	if _, err := s.svc.do(ctx, "PATCH", "/v1/inAppPurchases/"+id, nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete removes an in-app purchase.
func (s *InAppPurchasesService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "InAppPurchases.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/inAppPurchases/"+id, nil, nil, nil)
	return err
}
