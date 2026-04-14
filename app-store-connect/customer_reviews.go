package AppStoreConnect

import (
	"context"
	"time"
)

// CustomerReviewsService provides access to the customer reviews and
// customer review responses endpoints.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/customer_reviews
type CustomerReviewsService struct {
	svc *Service
}

// CustomerReview is a typed alias for a JSON:API customerReviews resource.
type CustomerReview = Resource[CustomerReviewAttributes]

// CustomerReviewAttributes models the attributes of a customerReviews
// resource. Unknown fields are tolerated — Apple adds fields to this
// schema over time.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/customerreview/attributes
type CustomerReviewAttributes struct {
	Rating           int        `json:"rating,omitempty"`
	Title            string     `json:"title,omitempty"`
	Body             string     `json:"body,omitempty"`
	ReviewerNickname string     `json:"reviewerNickname,omitempty"`
	CreatedDate      *time.Time `json:"createdDate,omitempty"`
	Territory        string     `json:"territory,omitempty"`
}

// CustomerReviewResponse is a typed alias for a JSON:API
// customerReviewResponses resource.
type CustomerReviewResponse = Resource[CustomerReviewResponseAttributes]

// CustomerReviewResponseAttributes models the attributes of a
// customerReviewResponses resource.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/customerreviewresponsev1/attributes
type CustomerReviewResponseAttributes struct {
	ResponseBody     string     `json:"responseBody,omitempty"`
	LastModifiedDate *time.Time `json:"lastModifiedDate,omitempty"`
	State            string     `json:"state,omitempty"`
}

// ListCustomerReviewsResponse is the decoded response for
// [CustomerReviewsService.ListForApp].
type ListCustomerReviewsResponse struct {
	Data     []CustomerReview `json:"data"`
	Included []Resource[any]  `json:"included,omitempty"`
	Links    *Links           `json:"links,omitempty"`
}

// ListForApp fetches one page of customer reviews for the given app.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/list_all_customer_reviews_for_an_app
func (s *CustomerReviewsService) ListForApp(ctx context.Context, appID string, query *Query) (*ListCustomerReviewsResponse, error) {
	if appID == "" {
		return nil, &ClientError{Message: "CustomerReviews.ListForApp: appID is required"}
	}
	var doc Document[CustomerReviewAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/apps/"+appID+"/customerReviews", query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListCustomerReviewsResponse{
		Data:     data,
		Included: doc.Included,
		Links:    doc.Links,
	}, nil
}

// ListForAppIterator returns a paginator that walks every page of
// customer reviews for the given app, auto-following links.next.
func (s *CustomerReviewsService) ListForAppIterator(appID string, query *Query) *Paginator[CustomerReviewAttributes] {
	return newPaginator[CustomerReviewAttributes](s.svc, "/v1/apps/"+appID+"/customerReviews", query)
}

// Get returns a single customer review by its resource id.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/read_customer_review_information
func (s *CustomerReviewsService) Get(ctx context.Context, id string, query *Query) (*CustomerReview, error) {
	if id == "" {
		return nil, &ClientError{Message: "CustomerReviews.Get: id is required"}
	}
	var doc Document[CustomerReviewAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/customerReviews/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Respond creates a developer response to the given customer review.
//
// Apple returns the created customerReviewResponses resource on
// success. Note that responses are moderated server-side and may sit
// in a "PENDING_PUBLISH" state until Apple approves them.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/create_a_customer_review_response
func (s *CustomerReviewsService) Respond(ctx context.Context, reviewID, responseBody string) (*CustomerReviewResponse, error) {
	if reviewID == "" {
		return nil, &ClientError{Message: "CustomerReviews.Respond: reviewID is required"}
	}
	if responseBody == "" {
		return nil, &ClientError{Message: "CustomerReviews.Respond: responseBody is required"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type": "customerReviewResponses",
			"attributes": map[string]any{
				"responseBody": responseBody,
			},
			"relationships": map[string]any{
				"review": map[string]any{
					"data": map[string]any{
						"type": "customerReviews",
						"id":   reviewID,
					},
				},
			},
		},
	}
	var doc Document[CustomerReviewResponseAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/customerReviewResponses", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// DeleteResponse removes a previously-submitted developer response.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/delete_a_customer_review_response
func (s *CustomerReviewsService) DeleteResponse(ctx context.Context, responseID string) error {
	if responseID == "" {
		return &ClientError{Message: "CustomerReviews.DeleteResponse: responseID is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/customerReviewResponses/"+responseID, nil, nil, nil)
	return err
}
