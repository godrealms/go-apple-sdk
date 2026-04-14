package AppStoreConnect

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAppsService_List(t *testing.T) {
	var captured struct {
		method string
		path   string
		query  string
		auth   string
	}
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		captured.query = r.URL.RawQuery
		captured.auth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "apps_list_page1.json"))
	}))

	q := NewQuery().
		Filter("bundleId", "com.acme.widgets").
		Fields("apps", "name", "bundleId").
		Limit(2)
	resp, err := svc.Apps().List(context.Background(), q)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if captured.method != "GET" {
		t.Errorf("method = %q", captured.method)
	}
	if captured.path != "/v1/apps" {
		t.Errorf("path = %q", captured.path)
	}
	if !strings.Contains(captured.query, "filter%5BbundleId%5D=com.acme.widgets") {
		t.Errorf("missing filter param: %q", captured.query)
	}
	if !strings.Contains(captured.query, "limit=2") {
		t.Errorf("missing limit: %q", captured.query)
	}
	if captured.auth != "Bearer test-token" {
		t.Errorf("auth = %q", captured.auth)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("resp.Data len = %d, want 2", len(resp.Data))
	}
	if resp.Data[0].Attributes.Name != "Acme Widgets" {
		t.Errorf("Data[0].Name = %q", resp.Data[0].Attributes.Name)
	}
	if resp.Links == nil || !strings.Contains(resp.Links.Next, "CURSOR_PAGE_2") {
		t.Errorf("missing next link: %+v", resp.Links)
	}
}

func TestAppsService_Get(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/apps/1234567890" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "app_get.json"))
	}))

	app, err := svc.Apps().Get(context.Background(), "1234567890", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if app.Attributes.BundleId != "com.acme.widgets" {
		t.Errorf("BundleId = %q", app.Attributes.BundleId)
	}
	// Verify relationships decoded: appStoreVersions is to-many with 2 entries.
	versions, err := app.Relationships["appStoreVersions"].AsMany()
	if err != nil {
		t.Fatalf("AsMany: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("versions len = %d, want 2", len(versions))
	}
}

func TestAppsService_GetEmptyIdFails(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	if _, err := svc.Apps().Get(context.Background(), "", nil); err == nil {
		t.Error("expected error for empty id")
	}
}

func TestAppsService_ErrorResponse(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write(loadFixture(t, "error_403.json"))
	}))

	_, err := svc.Apps().List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
	if !apiErr.HasCode("FORBIDDEN_ERROR") {
		t.Error("expected FORBIDDEN_ERROR code")
	}
}

func TestAppsService_ListIterator_AutoPagination(t *testing.T) {
	// Two-page scenario: page 1 has links.next pointing to page 2,
	// page 2 has no next link. The paginator should return exactly 2
	// pages and stop.
	var calls int
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Query().Get("cursor") == "CURSOR_PAGE_2":
			_, _ = w.Write(loadFixture(t, "apps_list_page2.json"))
		default:
			// Rewrite the fixture's hard-coded next URL to point at our
			// httptest server so the paginator can actually follow it.
			// Use r.Host (set by httptest) instead of referencing the
			// server variable to avoid a forward-reference on srv.
			body := loadFixture(t, "apps_list_page1.json")
			rewritten := strings.ReplaceAll(
				string(body),
				"https://api.appstoreconnect.apple.com",
				"http://"+r.Host,
			)
			_, _ = w.Write([]byte(rewritten))
		}
	}))

	it := svc.Apps().ListIterator(NewQuery().Limit(2))
	var all []App
	for it.Next(context.Background()) {
		all = append(all, it.Page().Data...)
	}
	if err := it.Err(); err != nil {
		t.Fatalf("iterator err: %v", err)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
	if len(all) != 3 {
		t.Fatalf("total apps = %d, want 3", len(all))
	}
	if all[0].Id != "1234567890" || all[2].Id != "5555555555" {
		t.Errorf("unexpected ids: %q %q %q", all[0].Id, all[1].Id, all[2].Id)
	}
}

func TestAppsService_ListAll(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("cursor") == "CURSOR_PAGE_2" {
			_, _ = w.Write(loadFixture(t, "apps_list_page2.json"))
			return
		}
		body := loadFixture(t, "apps_list_page1.json")
		rewritten := strings.ReplaceAll(
			string(body),
			"https://api.appstoreconnect.apple.com",
			"http://"+r.Host,
		)
		_, _ = w.Write([]byte(rewritten))
	}))

	all, err := svc.Apps().ListAll(context.Background(), NewQuery().Limit(2))
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("len = %d, want 3", len(all))
	}
}

func TestAppsService_Update(t *testing.T) {
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
		_, _ = w.Write(loadFixture(t, "app_updated.json"))
	}))

	upd := NewAppUpdate().
		PrimaryLocale("en-GB").
		AvailableInNewTerritories(true).
		ContentRightsDeclaration("USES_THIRD_PARTY_CONTENT")

	app, err := svc.Apps().Update(context.Background(), "1234567890", upd)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if captured.method != "PATCH" {
		t.Errorf("method = %q", captured.method)
	}
	if captured.path != "/v1/apps/1234567890" {
		t.Errorf("path = %q", captured.path)
	}
	if !strings.Contains(captured.contentType, "application/json") {
		t.Errorf("Content-Type = %q", captured.contentType)
	}

	data, ok := captured.body["data"].(map[string]any)
	if !ok {
		t.Fatalf("body.data missing: %+v", captured.body)
	}
	if data["type"] != "apps" {
		t.Errorf("data.type = %v", data["type"])
	}
	if data["id"] != "1234567890" {
		t.Errorf("data.id = %v", data["id"])
	}
	attrs, _ := data["attributes"].(map[string]any)
	if attrs["primaryLocale"] != "en-GB" {
		t.Errorf("attributes.primaryLocale = %v", attrs["primaryLocale"])
	}
	if attrs["availableInNewTerritories"] != true {
		t.Errorf("attributes.availableInNewTerritories = %v", attrs["availableInNewTerritories"])
	}
	if attrs["contentRightsDeclaration"] != "USES_THIRD_PARTY_CONTENT" {
		t.Errorf("attributes.contentRightsDeclaration = %v", attrs["contentRightsDeclaration"])
	}

	if app.Attributes.PrimaryLocale != "en-GB" {
		t.Errorf("app.PrimaryLocale = %q", app.Attributes.PrimaryLocale)
	}
}

func TestAppsService_Update_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	if _, err := svc.Apps().Update(context.Background(), "", NewAppUpdate().PrimaryLocale("en-US")); err == nil {
		t.Error("expected error for empty id")
	}
	if _, err := svc.Apps().Update(context.Background(), "1", nil); err == nil {
		t.Error("expected error for nil update")
	}
	if _, err := svc.Apps().Update(context.Background(), "1", NewAppUpdate()); err == nil {
		t.Error("expected error for empty update")
	}
}

func TestAppUpdate_SetArbitraryAttribute(t *testing.T) {
	u := NewAppUpdate().Set("customKey", "customValue")
	if u.IsEmpty() {
		t.Error("IsEmpty = true, want false after Set")
	}
	if u.attrs["customKey"] != "customValue" {
		t.Errorf("custom attr not recorded: %+v", u.attrs)
	}
}

func TestAppUpdate_BuilderCoverage(t *testing.T) {
	// Exercises every setter so future refactors don't silently drop
	// one. Each setter is a thin one-liner, so a smoke test here is
	// sufficient — behavior is covered by the Update HTTP test above.
	u := NewAppUpdate().
		PrimaryLocale("en-US").
		BundleId("com.acme.widgets").
		AvailableInNewTerritories(false).
		ContentRightsDeclaration("DOES_NOT_USE_THIRD_PARTY_CONTENT").
		SubscriptionStatusUrl("https://example.com/notify").
		SubscriptionStatusUrlVersion("V2").
		SubscriptionStatusUrlForSandbox("https://sandbox.example.com/notify").
		SubscriptionStatusUrlVersionForSandbox("V2")
	cases := map[string]any{
		"primaryLocale":                          "en-US",
		"bundleId":                               "com.acme.widgets",
		"availableInNewTerritories":              false,
		"contentRightsDeclaration":               "DOES_NOT_USE_THIRD_PARTY_CONTENT",
		"subscriptionStatusUrl":                  "https://example.com/notify",
		"subscriptionStatusUrlVersion":           "V2",
		"subscriptionStatusUrlForSandbox":        "https://sandbox.example.com/notify",
		"subscriptionStatusUrlVersionForSandbox": "V2",
	}
	for k, want := range cases {
		if got := u.attrs[k]; got != want {
			t.Errorf("attrs[%q] = %v, want %v", k, got, want)
		}
	}
}

func TestService_AuthorizerCanFail(t *testing.T) {
	badAuth := AuthorizerFunc(func(req *http.Request) error {
		return errors.New("boom")
	})
	svc := New(Config{
		BaseURL:    "http://127.0.0.1:1",
		Authorizer: badAuth,
	})
	_, err := svc.Apps().List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var ce *ClientError
	if !errors.As(err, &ce) {
		t.Errorf("expected *ClientError, got %T", err)
	}
}
