package AppStoreConnect

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCustomerReviews_ListForApp(t *testing.T) {
	var captured struct {
		method string
		path   string
		query  string
	}
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		captured.query = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "customer_reviews_page1.json"))
	}))

	q := NewQuery().Sort("-createdDate").Limit(2)
	resp, err := svc.CustomerReviews().ListForApp(context.Background(), "1234567890", q)
	if err != nil {
		t.Fatalf("ListForApp: %v", err)
	}
	if captured.method != "GET" {
		t.Errorf("method = %q", captured.method)
	}
	if captured.path != "/v1/apps/1234567890/customerReviews" {
		t.Errorf("path = %q", captured.path)
	}
	if !strings.Contains(captured.query, "sort=-createdDate") {
		t.Errorf("missing sort: %q", captured.query)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(resp.Data))
	}
	if resp.Data[0].Attributes.Rating != 5 {
		t.Errorf("rating = %d", resp.Data[0].Attributes.Rating)
	}
	if resp.Data[1].Attributes.Title != "Bug" {
		t.Errorf("title = %q", resp.Data[1].Attributes.Title)
	}
	if resp.Links == nil || !strings.Contains(resp.Links.Next, "CURSOR_PAGE_2") {
		t.Errorf("missing next link: %+v", resp.Links)
	}
}

func TestCustomerReviews_ListForApp_EmptyAppIDFails(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	if _, err := svc.CustomerReviews().ListForApp(context.Background(), "", nil); err == nil {
		t.Error("expected error")
	}
}

func TestCustomerReviews_Respond(t *testing.T) {
	var captured struct {
		method      string
		path        string
		contentType string
		body        map[string]any
	}
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		captured.contentType = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &captured.body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(loadFixture(t, "customer_review_response_created.json"))
	}))

	resp, err := svc.CustomerReviews().Respond(context.Background(), "review-2", "Thanks for the feedback!")
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if captured.method != "POST" {
		t.Errorf("method = %q", captured.method)
	}
	if captured.path != "/v1/customerReviewResponses" {
		t.Errorf("path = %q", captured.path)
	}
	if !strings.Contains(captured.contentType, "application/json") {
		t.Errorf("Content-Type = %q", captured.contentType)
	}

	// Verify the JSON:API create document shape: data.type,
	// data.attributes.responseBody, data.relationships.review.data.
	data, ok := captured.body["data"].(map[string]any)
	if !ok {
		t.Fatalf("body.data missing: %+v", captured.body)
	}
	if data["type"] != "customerReviewResponses" {
		t.Errorf("data.type = %v", data["type"])
	}
	attrs, _ := data["attributes"].(map[string]any)
	if attrs["responseBody"] != "Thanks for the feedback!" {
		t.Errorf("attributes.responseBody = %v", attrs["responseBody"])
	}
	rels, _ := data["relationships"].(map[string]any)
	review, _ := rels["review"].(map[string]any)
	reviewData, _ := review["data"].(map[string]any)
	if reviewData["type"] != "customerReviews" || reviewData["id"] != "review-2" {
		t.Errorf("relationships.review.data = %v", reviewData)
	}

	if resp.Id != "resp-1" {
		t.Errorf("resp.Id = %q", resp.Id)
	}
	if resp.Attributes.State != "PENDING_PUBLISH" {
		t.Errorf("resp.State = %q", resp.Attributes.State)
	}
}

func TestCustomerReviews_Respond_MissingArgs(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	if _, err := svc.CustomerReviews().Respond(context.Background(), "", "body"); err == nil {
		t.Error("expected error for empty reviewID")
	}
	if _, err := svc.CustomerReviews().Respond(context.Background(), "review-1", ""); err == nil {
		t.Error("expected error for empty responseBody")
	}
}

func TestCustomerReviews_Get(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/customerReviews/review-1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"type":"customerReviews","id":"review-1","attributes":{"rating":5,"title":"Great"}}}`))
	}))
	rev, err := svc.CustomerReviews().Get(context.Background(), "review-1", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rev.Id != "review-1" || rev.Attributes.Rating != 5 {
		t.Errorf("unexpected review: %+v", rev)
	}
	if _, err := svc.CustomerReviews().Get(context.Background(), "", nil); err == nil {
		t.Error("expected error for empty id")
	}
}

func TestCustomerReviews_ListForAppIterator(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("cursor") == "CURSOR_PAGE_2" {
			// Terminal page: no next link.
			_, _ = w.Write([]byte(`{"data":[{"type":"customerReviews","id":"review-3","attributes":{"rating":4}}]}`))
			return
		}
		body := loadFixture(t, "customer_reviews_page1.json")
		rewritten := strings.ReplaceAll(string(body),
			"https://api.appstoreconnect.apple.com",
			"http://"+r.Host)
		_, _ = w.Write([]byte(rewritten))
	}))

	it := svc.CustomerReviews().ListForAppIterator("1234567890", NewQuery().Limit(2))
	var all []CustomerReview
	for it.Next(context.Background()) {
		all = append(all, it.Page().Data...)
	}
	if err := it.Err(); err != nil {
		t.Fatalf("iter err: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("total = %d, want 3", len(all))
	}
}

func TestCustomerReviews_DeleteResponse(t *testing.T) {
	var captured struct {
		method string
		path   string
	}
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	if err := svc.CustomerReviews().DeleteResponse(context.Background(), "resp-1"); err != nil {
		t.Fatalf("DeleteResponse: %v", err)
	}
	if captured.method != "DELETE" {
		t.Errorf("method = %q", captured.method)
	}
	if captured.path != "/v1/customerReviewResponses/resp-1" {
		t.Errorf("path = %q", captured.path)
	}
	if err := svc.CustomerReviews().DeleteResponse(context.Background(), ""); err == nil {
		t.Error("expected error for empty responseID")
	}
}
